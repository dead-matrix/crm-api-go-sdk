package crmapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ChangeDialogStatus отправляет {user_id, status_id} в POST /api/dialogs/status
// и распаковывает ответ в ChangeStatusResult.Status.
func TestClient_ChangeDialogStatus(t *testing.T) {
	var captured map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/dialogs/status":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &captured)
			_, _ = w.Write([]byte(`{"status":"success","data":{"status":"In progress"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ChangeDialogStatus(context.Background(), 100, 20)
	if err != nil {
		t.Fatalf("ChangeDialogStatus error: %v", err)
	}
	if res.Status == nil {
		t.Fatalf("status = nil, want %q", "In progress")
	}
	if *res.Status != "In progress" {
		t.Fatalf("status = %q, want %q", *res.Status, "In progress")
	}
	if captured["user_id"] != float64(100) || captured["status_id"] != float64(20) {
		t.Fatalf("unexpected request body: %+v", captured)
	}
}

// ClearDialogStatus отправляет payload с status_id=null (явно — паритет с
// Python SDK), и сохраняет распаковку ответа status=null как
// ChangeStatusResult.Status == nil (паритет с Python: str | None).
func TestClient_ClearDialogStatus(t *testing.T) {
	var captured map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/dialogs/status":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &captured)
			_, _ = w.Write([]byte(`{"status":"success","data":{"status":null}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ClearDialogStatus(context.Background(), 42)
	if err != nil {
		t.Fatalf("ClearDialogStatus error: %v", err)
	}
	if res.Status != nil {
		t.Fatalf("status = %q, want nil (cleared)", *res.Status)
	}

	if v, ok := captured["status_id"]; !ok || v != nil {
		t.Fatalf("status_id must be present and null: %+v", captured)
	}
	if captured["user_id"] != float64(42) {
		t.Fatalf("user_id mismatch: %+v", captured)
	}
}

// TestClient_SearchDialogs_NullStatusPreserved: сервер может вернуть
// status/status_color=null (диалог без статуса в департаменте без
// default_status). SDK сохраняет nil; не превращает в "".
func TestClient_SearchDialogs_NullStatusPreserved(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/dialogs/sales/search":
			_, _ = w.Write([]byte(`{"status":"success","data":{
				"dialogs":[
					{"user_id":1,"full_name":"User","has_active_subscription":false,"status":null,"status_color":null}
				],
				"limit":50,"offset":0
			}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.SearchDialogs(context.Background(), "sales", "user", 0)
	if err != nil {
		t.Fatalf("SearchDialogs error: %v", err)
	}
	if len(res.Dialogs) != 1 {
		t.Fatalf("expected 1 dialog, got %d", len(res.Dialogs))
	}
	if res.Dialogs[0].Status != nil {
		t.Fatalf("Status = %q, want nil", *res.Dialogs[0].Status)
	}
	if res.Dialogs[0].StatusColor != nil {
		t.Fatalf("StatusColor = %q, want nil", *res.Dialogs[0].StatusColor)
	}
}

// TestClient_ChangeDialogStatus_NullDistinct гарантирует, что nil status
// (статус снят) и пустая строка остаются разными значениями — это и есть
// цель перехода ChangeStatusResult.Status на *string.
func TestClient_ChangeDialogStatus_NullDistinct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/dialogs/status":
			_, _ = w.Write([]byte(`{"status":"success","data":{"status":""}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ChangeDialogStatus(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("ChangeDialogStatus error: %v", err)
	}
	if res.Status == nil {
		t.Fatalf("status = nil, expected empty-string pointer (distinct from cleared)")
	}
	if *res.Status != "" {
		t.Fatalf("status = %q, expected empty string", *res.Status)
	}
}

func TestClient_ClearDialogStatus_RejectsInvalidUserID(t *testing.T) {
	client := mustNewClient(t, "https://example.test", &http.Client{})
	_, err := client.ClearDialogStatus(context.Background(), 0)
	if err == nil {
		t.Fatal("expected ValidationError, got nil")
	}
	if !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("unexpected error: %v", err)
	}
}
