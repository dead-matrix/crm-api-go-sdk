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
// прокидывает include_removed.
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
	total, err := client.AccountsCount(context.Background(), 100, true)
	if err != nil {
		t.Fatalf("AccountsCount error: %v", err)
	}
	if total != 30028 {
		t.Fatalf("total = %d, want 30028", total)
	}
	if !strings.Contains(gotQuery, "user_id=100") || !strings.Contains(gotQuery, "include_removed=true") {
		t.Fatalf("query = %q, want user_id=100 & include_removed=true", gotQuery)
	}
}

func TestClient_AccountsCount_RejectsInvalidUserID(t *testing.T) {
	client := mustNewClient(t, "https://example.test", &http.Client{})
	if _, err := client.AccountsCount(context.Background(), 0, false); err == nil || !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("expected user_id ValidationError, got %v", err)
	}
}

// TestClient_AccountsList_DaysParam: days>0 уходит в query, days=0 — нет.
func TestClient_AccountsList_DaysParam(t *testing.T) {
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

	if _, err := client.AccountsList(context.Background(), 100, true, 30); err != nil {
		t.Fatalf("AccountsList(days=30) error: %v", err)
	}
	if !strings.Contains(gotQuery, "days=30") {
		t.Fatalf("query = %q, want days=30", gotQuery)
	}

	if _, err := client.AccountsList(context.Background(), 100, true, 0); err != nil {
		t.Fatalf("AccountsList(days=0) error: %v", err)
	}
	if strings.Contains(gotQuery, "days=") {
		t.Fatalf("query = %q, want no days param for days=0", gotQuery)
	}
}
