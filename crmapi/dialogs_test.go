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
	if res.Status != "In progress" {
		t.Fatalf("status = %q, want %q", res.Status, "In progress")
	}
	if captured["user_id"] != float64(100) || captured["status_id"] != float64(20) {
		t.Fatalf("unexpected request body: %+v", captured)
	}
}

// ClearDialogStatus отправляет payload БЕЗ status_id и возвращает пустую
// строку, когда сервер отвечает status=null.
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
	if res.Status != "" {
		t.Fatalf("status = %q, want empty (cleared)", res.Status)
	}

	if _, ok := captured["status_id"]; ok {
		t.Fatalf("status_id must not be sent in clear payload: %+v", captured)
	}
	if captured["user_id"] != float64(42) {
		t.Fatalf("user_id mismatch: %+v", captured)
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
