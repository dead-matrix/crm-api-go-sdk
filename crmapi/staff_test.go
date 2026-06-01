package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ----------------------------------------------------------------------------
// ListStaff (GET /api/staff/list) — сотрудники с user_id > 1000
// ----------------------------------------------------------------------------

func TestListStaffReturnsItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/staff/list":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			fmt.Fprint(w, `{"status":"success","data":[
				{"user_id":1001,"name":"Alice","role":"admin"},
				{"user_id":7014133383,"name":"Bob","role":"support"}
			]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ListStaff(context.Background())
	if err != nil {
		t.Fatalf("ListStaff() error = %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("len(res) = %d, want 2", len(res))
	}
	if res[0].UserID != 1001 || res[0].Name != "Alice" || res[0].Role != "admin" {
		t.Fatalf("res[0] = %+v, want {1001 Alice admin}", res[0])
	}
	if res[1].UserID != 7014133383 || res[1].Name != "Bob" || res[1].Role != "support" {
		t.Fatalf("res[1] = %+v, want {7014133383 Bob support}", res[1])
	}
}

func TestListStaffEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/staff/list":
			fmt.Fprint(w, `{"status":"success","data":[]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ListStaff(context.Background())
	if err != nil {
		t.Fatalf("ListStaff() error = %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("len(res) = %d, want 0", len(res))
	}
}
