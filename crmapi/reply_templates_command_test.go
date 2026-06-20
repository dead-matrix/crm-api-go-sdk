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
)

// ----------------------------------------------------------------------------
// normalizeReplyTemplateCommand (pure)
// ----------------------------------------------------------------------------

func TestNormalizeReplyTemplateCommand(t *testing.T) {
	mk := func(s string) *string { return &s }
	cases := []struct {
		name    string
		in      *string
		want    *string
		wantErr bool
	}{
		{"nil clears", nil, nil, false},
		{"empty clears", mk(""), nil, false},
		{"spaces clear", mk("   "), nil, false},
		{"slash only clears", mk("/"), nil, false},
		{"strips slash + lowercases", mk("/HI"), mk("hi"), false},
		{"trims padding", mk("  /Hi  "), mk("hi"), false},
		{"cyrillic allowed", mk("привет"), mk("привет"), false},
		{"underscore + digits", mk("at_work2"), mk("at_work2"), false},
		{"space rejected", mk("a b"), nil, true},
		{"hyphen rejected", mk("a-b"), nil, true},
		{"punct rejected", mk("hi!"), nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeReplyTemplateCommand(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %v", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (got == nil) != (tc.want == nil) {
				t.Fatalf("nilness mismatch: got %v want %v", got, tc.want)
			}
			if got != nil && *got != *tc.want {
				t.Fatalf("got %q want %q", *got, *tc.want)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// ReplyTemplatesSetCommand (PATCH /api/reply-templates/{id})
// ----------------------------------------------------------------------------

func TestReplyTemplatesSetCommandSetsAndReturns(t *testing.T) {
	var capturedMethod string
	var capturedBody struct {
		Command *string `json:"command"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42":
			capturedMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &capturedBody)
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 42, "publicId": "uuid-x", "title": "t", "kind": "single",
				"command": "hi",
				"creator": {"employeeId": 5, "name": "Olga"},
				"items": [],
				"createdAt": null, "updatedAt": null
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	cmd := "/HI"
	full, err := client.ReplyTemplatesSetCommand(context.Background(), 42, &cmd)
	if err != nil {
		t.Fatalf("ReplyTemplatesSetCommand() error = %v", err)
	}
	if capturedMethod != http.MethodPatch {
		t.Fatalf("method = %s, want PATCH", capturedMethod)
	}
	// Client normalises "/HI" → "hi" before the round-trip.
	if capturedBody.Command == nil || *capturedBody.Command != "hi" {
		t.Fatalf("body command = %v, want normalized \"hi\"", capturedBody.Command)
	}
	if full.Command == nil || *full.Command != "hi" {
		t.Fatalf("returned command = %v, want \"hi\"", full.Command)
	}
}

func TestReplyTemplatesSetCommandClearsWithNil(t *testing.T) {
	var bodyHasNullCommand bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42":
			body, _ := io.ReadAll(r.Body)
			bodyHasNullCommand = strings.Contains(strings.ReplaceAll(string(body), " ", ""), `"command":null`)
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 42, "publicId": "uuid-x", "title": "t", "kind": "single",
				"command": null, "creator": {"employeeId": 5, "name": "Olga"},
				"items": [], "createdAt": null, "updatedAt": null
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	full, err := client.ReplyTemplatesSetCommand(context.Background(), 42, nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !bodyHasNullCommand {
		t.Fatalf("expected body to carry \"command\":null")
	}
	if full.Command != nil {
		t.Fatalf("returned command = %v, want nil", full.Command)
	}
}

func TestReplyTemplatesSetCommandRejectsBadFormatBeforeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/reply-templates") {
			t.Fatalf("server must not be hit for an invalid command")
		}
		fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	bad := "/has space"
	if _, err := client.ReplyTemplatesSetCommand(context.Background(), 42, &bad); err == nil {
		t.Fatalf("expected validation error for an invalid command")
	}
}

func TestReplyTemplatesSetCommandRejectsNonPositiveID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server must not be hit for templateID=0")
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	if _, err := client.ReplyTemplatesSetCommand(context.Background(), 0, nil); err == nil {
		t.Fatalf("expected error for templateID=0")
	}
}
