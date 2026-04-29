package crmapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// CreateUser — idempotent contract (POST /api/users)
// ----------------------------------------------------------------------------

func TestCreateUserFirstTimeReturnsCreatedTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"user_id":999`) {
				t.Fatalf("body must contain user_id=999, got %s", body)
			}
			fmt.Fprint(w, `{"status":"success","data":{
				"created":true,
				"user_id":999,
				"full_name":"Charlie",
				"username":"charlie",
				"bot_id":1,
				"refer":null,
				"date_reg":"2026-04-29T12:00:00Z"
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	username := "charlie"
	res, err := client.CreateUser(context.Background(), CreateUserInput{
		UserID:   999,
		FullName: "Charlie",
		Username: &username,
		BotID:    1,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if !res.Created {
		t.Fatalf("Created = false, want true")
	}
	if res.UserID != 999 {
		t.Fatalf("UserID = %d, want 999", res.UserID)
	}
	if res.FullName != "Charlie" {
		t.Fatalf("FullName = %q, want Charlie", res.FullName)
	}
	if res.BotID != 1 {
		t.Fatalf("BotID = %d, want 1", res.BotID)
	}
	if res.DateReg == nil {
		t.Fatalf("DateReg should be parsed, got nil")
	}
}

func TestCreateUserIdempotentReturnsExistingRecord(t *testing.T) {
	// Сервер моделирует случай: регистрация (12345, 1) уже существовала.
	// Возвращает старые значения, игнорируя присланный full_name="NewName".
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users":
			fmt.Fprint(w, `{"status":"success","data":{
				"created":false,
				"user_id":12345,
				"full_name":"OldName",
				"username":"oldun",
				"bot_id":1,
				"refer":"oldref",
				"date_reg":"2026-01-15T12:00:00Z"
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	newName := "NewName"
	newUsername := "newun"
	newRefer := "newref"
	res, err := client.CreateUser(context.Background(), CreateUserInput{
		UserID:   12345,
		FullName: newName,
		Username: &newUsername,
		BotID:    1,
		Refer:    &newRefer,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if res.Created {
		t.Fatalf("Created = true, want false (idempotent path)")
	}
	if res.FullName != "OldName" {
		t.Fatalf("FullName = %q, want OldName (server returns existing, not payload)", res.FullName)
	}
	if res.Refer == nil || *res.Refer != "oldref" {
		t.Fatalf("Refer = %v, want pointer to \"oldref\"", res.Refer)
	}
}

func TestCreateUserValidationFailsWithoutHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called when validation fails: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.CreateUser(context.Background(), CreateUserInput{
		UserID: 0, FullName: "X", BotID: 1,
	})
	var ve *ValidationError
	if !errorsAsValidation(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// ListUsers (GET /api/users?bot_id=...)
// ----------------------------------------------------------------------------

func TestListUsersReturnsPaginatedItems(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users":
			capturedQuery = r.URL.Query()
			fmt.Fprint(w, `{"status":"success","data":{
				"bot_id":1,"limit":50,"offset":10,"count":2,
				"items":[
					{"user_id":101,"full_name":"Alice","username":"alice","date_reg":"2026-04-01T10:00:00Z","refer":null,"restricted":false},
					{"user_id":102,"full_name":"Bob","username":null,"date_reg":"2026-04-02T11:00:00Z","refer":"ref-x","restricted":true}
				]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ListUsers(context.Background(), 1, 50, 10)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	// Query params correctly encoded.
	if got := capturedQuery.Get("bot_id"); got != "1" {
		t.Fatalf("bot_id query = %q, want 1", got)
	}
	if got := capturedQuery.Get("limit"); got != "50" {
		t.Fatalf("limit query = %q, want 50", got)
	}
	if got := capturedQuery.Get("offset"); got != "10" {
		t.Fatalf("offset query = %q, want 10", got)
	}

	// Response shape.
	if res.BotID != 1 || res.Limit != 50 || res.Offset != 10 || res.Count != 2 {
		t.Fatalf("envelope = %+v, want bot_id=1 limit=50 offset=10 count=2", res)
	}
	if len(res.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(res.Items))
	}

	alice, bob := res.Items[0], res.Items[1]
	if alice.UserID != 101 || alice.FullName != "Alice" || alice.Restricted {
		t.Fatalf("alice = %+v", alice)
	}
	if alice.Username == nil || *alice.Username != "alice" {
		t.Fatalf("alice.Username = %v, want pointer to \"alice\"", alice.Username)
	}
	if alice.DateReg == nil {
		t.Fatalf("alice.DateReg should be parsed")
	}

	if bob.UserID != 102 || bob.Username != nil || !bob.Restricted {
		t.Fatalf("bob = %+v (Username=%v)", bob, bob.Username)
	}
	if bob.Refer == nil || *bob.Refer != "ref-x" {
		t.Fatalf("bob.Refer = %v, want pointer to \"ref-x\"", bob.Refer)
	}
}

func TestListUsersValidationFailsWithoutHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	cases := []struct {
		name           string
		botID          int64
		limit          int64
		offset         int64
	}{
		{"bot_id zero", 0, 100, 0},
		{"bot_id negative", -1, 100, 0},
		{"limit zero", 1, 0, 0},
		{"offset negative", 1, 100, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ListUsers(context.Background(), tc.botID, tc.limit, tc.offset)
			var ve *ValidationError
			if !errorsAsValidation(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

func TestListUsersEmptyItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users":
			fmt.Fprint(w, `{"status":"success","data":{
				"bot_id":2,"limit":100000,"offset":0,"count":0,"items":[]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ListUsers(context.Background(), 2, 100_000, 0)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if res.Count != 0 || len(res.Items) != 0 {
		t.Fatalf("expected empty list, got %+v", res)
	}
}
