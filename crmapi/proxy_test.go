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

// TestClient_ProxyBindings: GET /api/proxy/bindings декодирует сводку привязок.
func TestClient_ProxyBindings(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/proxy/bindings":
			gotQuery = r.URL.RawQuery
			_, _ = w.Write([]byte(`{"status":"success","data":{
				"total_accounts":4,"accounts_with_proxy":3,"accounts_without_proxy":1,
				"total_proxies":3,"proxies_with_accounts":2,"proxies_without_accounts":1,
				"avg_accounts_per_proxy":1.5
			}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ProxyBindings(context.Background(), 100)
	if err != nil {
		t.Fatalf("ProxyBindings error: %v", err)
	}
	if !strings.Contains(gotQuery, "user_id=100") {
		t.Fatalf("query = %q, want user_id=100", gotQuery)
	}
	if res.AccountsWithProxy != 3 || res.AccountsWithoutProxy != 1 {
		t.Fatalf("accounts with/without = %d/%d, want 3/1", res.AccountsWithProxy, res.AccountsWithoutProxy)
	}
	if res.ProxiesWithAccounts != 2 || res.ProxiesWithoutAccounts != 1 {
		t.Fatalf("proxies with/without = %d/%d, want 2/1", res.ProxiesWithAccounts, res.ProxiesWithoutAccounts)
	}
	if res.AvgAccountsPerProxy != 1.5 {
		t.Fatalf("avg = %v, want 1.5", res.AvgAccountsPerProxy)
	}
}

func TestClient_ProxyBindings_RejectsInvalidUserID(t *testing.T) {
	client := mustNewClient(t, "https://example.test", &http.Client{})
	if _, err := client.ProxyBindings(context.Background(), 0); err == nil || !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("expected user_id ValidationError, got %v", err)
	}
}

// TestClient_AccountsList_ProxyField: новое поле proxy (ip:port) декодируется,
// null остаётся nil.
func TestClient_AccountsList_ProxyField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/accounts/list":
			_, _ = w.Write([]byte(`{"status":"success","data":[
				{"session_name":"a.session","proxy":"1.2.3.4:1080"},
				{"session_name":"b.session","proxy":null}
			]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	items, err := client.AccountsList(context.Background(), 100, false)
	if err != nil {
		t.Fatalf("AccountsList error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Proxy == nil || *items[0].Proxy != "1.2.3.4:1080" {
		t.Fatalf("item0 proxy = %v, want 1.2.3.4:1080", items[0].Proxy)
	}
	if items[1].Proxy != nil {
		t.Fatalf("item1 proxy = %v, want nil", *items[1].Proxy)
	}
}
