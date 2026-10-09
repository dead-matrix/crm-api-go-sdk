package crmapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

// Полный список: параметр kind не уходит вовсе, а в ответе он null.
func TestListBotBlocks_AllKinds(t *testing.T) {
	var query url.Values
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/bot-blocks")
		query = r.URL.Query()
		writeSuccess(w, `{"bot_id":10,"kind":null,"user_ids":[11,22,33],"count":3,
			"counts":{"blocked":2,"unreachable":1,"total":3}}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ListBotBlocks(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListBotBlocks error: %v", err)
	}
	if query.Get("bot_id") != "10" || len(query) != 1 {
		t.Fatalf("query = %v, want only bot_id", query)
	}
	if res.BotID != 10 || res.Kind != nil || len(res.UserIDs) != 3 || res.Count != 3 {
		t.Fatalf("result = %+v", res)
	}
	if res.Counts != (BotBlockCounts{Blocked: 2, Unreachable: 1, Total: 3}) {
		t.Fatalf("Counts = %+v", res.Counts)
	}
}

// Список одной категории: Count считает отфильтрованное, Counts - все флаги.
func TestListBotBlocksByKind_SendsKind(t *testing.T) {
	var query url.Values
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/bot-blocks")
		query = r.URL.Query()
		writeSuccess(w, `{"bot_id":10,"kind":"unreachable","user_ids":[33],"count":1,
			"counts":{"blocked":2,"unreachable":1,"total":3}}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	// Регистр и пробелы нормализуются: CRM сравнивает категорию буквально.
	res, err := client.ListBotBlocksByKind(context.Background(), 10, " Unreachable ")
	if err != nil {
		t.Fatalf("ListBotBlocksByKind error: %v", err)
	}
	if query.Get("bot_id") != "10" || query.Get("kind") != BotBlockKindUnreachable || len(query) != 2 {
		t.Fatalf("query = %v", query)
	}
	if res.Kind == nil || *res.Kind != "unreachable" || res.Count != 1 || len(res.UserIDs) != 1 || res.UserIDs[0] != 33 {
		t.Fatalf("result = %+v", res)
	}
	if res.Counts.Total != 3 || res.Counts.Blocked != 2 {
		t.Fatalf("Counts = %+v", res.Counts)
	}
}

// CRM без kind и counts в ответе: поля остаются nil и нулями.
func TestListBotBlocks_LegacyResponse(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"bot_id":1,"user_ids":[5],"count":1}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ListBotBlocks(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListBotBlocks error: %v", err)
	}
	if res.Kind != nil || res.Counts != (BotBlockCounts{}) || res.Count != 1 {
		t.Fatalf("result = %+v", res)
	}
}

// На неизвестную категорию CRM отвечает успехом с пустым списком, поэтому
// отсечь её обязан SDK.
func TestListBotBlocks_ValidationWithoutHTTP(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})
	client := mustNewClient(t, server.URL, server.Client())
	ctx := context.Background()

	calls := map[string]func() error{
		"bot":          func() error { _, err := client.ListBotBlocks(ctx, 0); return err },
		"kind bot":     func() error { _, err := client.ListBotBlocksByKind(ctx, 0, BotBlockKindBlocked); return err },
		"empty kind":   func() error { _, err := client.ListBotBlocksByKind(ctx, 10, ""); return err },
		"unknown kind": func() error { _, err := client.ListBotBlocksByKind(ctx, 10, "deactivated"); return err },
	}
	for name, call := range calls {
		var vErr *ValidationError
		if err := call(); !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}

func TestReportBotBlock_AddedAndIgnored(t *testing.T) {
	response := `{"added":true}`
	var body map[string]any
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/bot-blocks/report")
		body = readJSONBody(t, r)
		writeSuccess(w, response)
	})
	client := mustNewClient(t, server.URL, server.Client())

	res, err := client.ReportBotBlock(context.Background(), 10, 42, "", "Forbidden: bot was blocked by the user")
	if err != nil {
		t.Fatalf("ReportBotBlock error: %v", err)
	}
	if body["bot_id"] != float64(10) || body["user_id"] != float64(42) || body["reason"] != "blocked" {
		t.Fatalf("body = %v", body)
	}
	if !res.Added || res.Ignored {
		t.Fatalf("result = %+v, want added and not ignored", res)
	}

	// Бот, для которого CRM флаги не ведёт: сигнал отброшен.
	response = `{"added":false,"ignored":true}`
	res, err = client.ReportBotBlock(context.Background(), 999, 42, "blocked", "")
	if err != nil {
		t.Fatalf("ReportBotBlock error: %v", err)
	}
	if res.Added || !res.Ignored {
		t.Fatalf("result = %+v, want ignored", res)
	}
	if _, sent := body["error"]; sent {
		t.Fatalf("error must be omitted when empty, body = %v", body)
	}
}
