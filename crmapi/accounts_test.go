package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestClient_AccountsCount: GET /api/accounts/count декодирует {total} и
// прокидывает removed_only (счётчик удалённых перед их загрузкой).
func TestClient_AccountsCount(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/accounts/count":
			gotQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"status":"success","data":{"total":30028}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	total, err := client.AccountsCount(context.Background(), 100, false, true)
	if err != nil {
		t.Fatalf("AccountsCount error: %v", err)
	}
	if total != 30028 {
		t.Fatalf("total = %d, want 30028", total)
	}
	if !strings.Contains(gotQuery, "user_id=100") || !strings.Contains(gotQuery, "removed_only=true") {
		t.Fatalf("query = %q, want user_id=100 & removed_only=true", gotQuery)
	}
}

func TestClient_AccountsCount_RejectsInvalidUserID(t *testing.T) {
	client := mustNewClient(t, "https://example.test", &http.Client{})
	if _, err := client.AccountsCount(context.Background(), 0, false, false); err == nil || !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("expected user_id ValidationError, got %v", err)
	}
}

// TestClient_AccountsList_RemovedOnlyDays: removed_only + days уходят в query,
// days=0 без removed_only - чистый query (штатный активный список).
func TestClient_AccountsList_RemovedOnlyDays(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/accounts/list":
			gotQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"status":"success","data":[]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())

	// Удалённые за период.
	if _, err := client.AccountsList(context.Background(), 100, false, true, 30); err != nil {
		t.Fatalf("AccountsList(removedOnly, days=30) error: %v", err)
	}
	if !strings.Contains(gotQuery, "removed_only=true") || !strings.Contains(gotQuery, "days=30") {
		t.Fatalf("query = %q, want removed_only=true & days=30", gotQuery)
	}

	// Штатный активный список: без removed_only и без days.
	if _, err := client.AccountsList(context.Background(), 100, false, false, 0); err != nil {
		t.Fatalf("AccountsList(active) error: %v", err)
	}
	if strings.Contains(gotQuery, "removed_only=") || strings.Contains(gotQuery, "days=") {
		t.Fatalf("query = %q, want no removed_only/days for active", gotQuery)
	}
}
