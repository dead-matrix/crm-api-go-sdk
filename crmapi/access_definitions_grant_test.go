package crmapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func newAccessDefinitionsServer(t *testing.T, data string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/access/definitions":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			fmt.Fprintf(w, `{"status":"success","data":%s}`, data)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
}

func TestAccessDefinitions_Categories(t *testing.T) {
	server := newAccessDefinitionsServer(t, `{"main":{},"poster":{},"categories":{
		"cabinet":{"cabinet.pro":"Кабинет Pro","cabinet.agency":"Кабинет Agency"},
		"privetka":{"privetka.pro":"Приветка Pro"},
		"empty":null
	}}`)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.AccessDefinitions(context.Background())
	if err != nil {
		t.Fatalf("AccessDefinitions error: %v", err)
	}
	if res.Main == nil || len(res.Main) != 0 || res.Poster == nil || len(res.Poster) != 0 {
		t.Fatalf("main/poster = %v / %v, want empty non-nil", res.Main, res.Poster)
	}
	if len(res.Categories) != 3 {
		t.Fatalf("categories = %v, want 3 entries", res.Categories)
	}
	if got := res.Categories["cabinet"]["cabinet.agency"]; got != "Кабинет Agency" {
		t.Fatalf("cabinet.agency = %q", got)
	}
	if got := res.Categories["privetka"]["privetka.pro"]; got != "Приветка Pro" {
		t.Fatalf("privetka.pro = %q", got)
	}
	if inner, ok := res.Categories["empty"]; !ok || inner == nil {
		t.Fatalf("null category must become empty non-nil map, got %v (present=%v)", inner, ok)
	}
}

func TestAccessDefinitions_LegacyWithoutCategories(t *testing.T) {
	server := newAccessDefinitionsServer(t, `{"main":{"1":"Базовый"},"poster":{"2":"Постер"}}`)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.AccessDefinitions(context.Background())
	if err != nil {
		t.Fatalf("AccessDefinitions error: %v", err)
	}
	if res.Main["1"] != "Базовый" || res.Poster["2"] != "Постер" {
		t.Fatalf("main/poster = %v / %v", res.Main, res.Poster)
	}
	if res.Categories == nil || len(res.Categories) != 0 {
		t.Fatalf("categories = %v, want empty non-nil map", res.Categories)
	}
}

func TestGrantAITokens_BotIDRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	for _, botID := range []int64{0, -1} {
		_, err := client.GrantAITokens(context.Background(), 42, 1000, "chat", "ref-1", botID)
		var cErr *ConfigError
		if !errors.As(err, &cErr) {
			t.Fatalf("botID=%d: err = %v, want ConfigError", botID, err)
		}
	}
}

func TestGrantAITokens_SendsBotID(t *testing.T) {
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users/42/ai-tokens/grant":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			gotQuery = r.URL.Query()
			fmt.Fprint(w, `{"status":"success","data":{"granted":true,"function":"chat","tokens":1000,"previous_ai_limit":0,"ai_limit":1000,"balance_tokens":1000}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.GrantAITokens(context.Background(), 42, 1000, "chat", "ref-1", 10)
	if err != nil {
		t.Fatalf("GrantAITokens error: %v", err)
	}
	if got := gotQuery.Get("bot_id"); got != "10" {
		t.Fatalf("bot_id query = %q, want 10", got)
	}
	if !res.Granted || res.Tokens != 1000 {
		t.Fatalf("result = %+v", res)
	}
}
