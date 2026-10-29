package crmapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// Прежний метод обязан остаться запросом без тела: так его понимает и старая
// CRM, где параметра account_id нет.
func TestExtendUserAccess_NoBody(t *testing.T) {
	var rawBody []byte
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/users/42/access/extend")
		if q := r.URL.Query(); q.Get("bot_id") != "10" || q.Get("days") != "7" || len(q) != 2 {
			t.Fatalf("query = %v", q)
		}
		rawBody, _ = io.ReadAll(r.Body)
		writeSuccess(w, `{"user_id":42,"account_id":5,"access_end":"2026-10-16T12:00:00+03:00"}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ExtendUserAccess(context.Background(), 42, 10, 7)
	if err != nil {
		t.Fatalf("ExtendUserAccess error: %v", err)
	}
	if len(rawBody) != 0 {
		t.Fatalf("body = %q, want empty", rawBody)
	}
	if res.UserID != 42 || res.AccountID == nil || *res.AccountID != 5 {
		t.Fatalf("result = %+v", res)
	}
	want := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	if res.AccessEnd == nil || !res.AccessEnd.Equal(want) {
		t.Fatalf("AccessEnd = %v, want %v", res.AccessEnd, want)
	}
}

func TestExtendUserAccessForAccount_SendsAccountInBody(t *testing.T) {
	var body map[string]any
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/users/42/access/extend")
		if q := r.URL.Query(); q.Get("bot_id") != "10" || q.Get("days") != "30" || len(q) != 2 {
			t.Fatalf("query = %v, account_id belongs to the body", q)
		}
		body = readJSONBody(t, r)
		writeSuccess(w, `{"user_id":42,"account_id":77,"access_end":"2026-11-08T12:00:00+03:00"}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ExtendUserAccessForAccount(context.Background(), 42, 10, 30, 77)
	if err != nil {
		t.Fatalf("ExtendUserAccessForAccount error: %v", err)
	}
	if len(body) != 1 || body["account_id"] != float64(77) {
		t.Fatalf("body = %v, want only account_id=77", body)
	}
	if res.AccountID == nil || *res.AccountID != 77 || res.AccessEnd == nil {
		t.Fatalf("result = %+v", res)
	}
}

// Старая CRM account_id в ответе не присылает: поле остаётся nil.
func TestExtendUserAccess_LegacyResponseWithoutAccountID(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"user_id":42,"access_end":"2026-10-16T09:00:00"}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ExtendUserAccess(context.Background(), 42, 1, 7)
	if err != nil {
		t.Fatalf("ExtendUserAccess error: %v", err)
	}
	if res.AccountID != nil {
		t.Fatalf("AccountID = %d, want nil", *res.AccountID)
	}
}

// Человек не состоит в аккаунте: CRM отвечает 404 not_found, SDK отдаёт
// обычный *APIError с кодом.
func TestExtendUserAccessForAccount_NotFound(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(w, http.StatusNotFound, "not_found", "аккаунт не найден")
	})

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.ExtendUserAccessForAccount(context.Background(), 42, 10, 30, 999)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "not_found" || apiErr.Status != http.StatusNotFound {
		t.Fatalf("err = %v, want APIError not_found/404", err)
	}
}

func TestExtendUserAccess_ValidationWithoutHTTP(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})

	client := mustNewClient(t, server.URL, server.Client())
	ctx := context.Background()
	calls := map[string]func() error{
		"user":    func() error { _, err := client.ExtendUserAccess(ctx, 0, 10, 7); return err },
		"bot":     func() error { _, err := client.ExtendUserAccess(ctx, 42, 0, 7); return err },
		"days":    func() error { _, err := client.ExtendUserAccess(ctx, 42, 10, 0); return err },
		"account": func() error { _, err := client.ExtendUserAccessForAccount(ctx, 42, 10, 7, 0); return err },
		"account user": func() error {
			_, err := client.ExtendUserAccessForAccount(ctx, 0, 10, 7, 77)
			return err
		},
	}
	for name, call := range calls {
		var vErr *ValidationError
		if err := call(); !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}
