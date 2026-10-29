package crmapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestSubscriptionsHistory_PrimaryAccount(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42/subscriptions/history")
		if len(r.URL.Query()) != 0 {
			t.Fatalf("query = %v, want empty", r.URL.Query())
		}
		writeSuccess(w, `{"user_id":42,"account_id":5,"history":[
			{"action":"revoke","bot_id":10,"access":{"cabinet.pro":true},
			 "added":[],"removed":["privetka.pro"],
			 "action_date":"2026-10-09T12:00:00+03:00","access_end":"2026-11-01T00:00:00+03:00",
			 "payment":null,"staff":{"id":1001,"name":"Оператор"},"ref":"по просьбе"},
			{"action":"add","bot_id":10,"access":{"cabinet.pro":true,"privetka.pro":true},
			 "added":["cabinet.pro","privetka.pro"],"removed":[],
			 "action_date":"2026-10-01T12:00:00+03:00","access_end":"2026-11-01T00:00:00+03:00",
			 "payment":{"id":7,"amount_minor":99000,"currency":"RUB","status":"paid","date_paid":"2026-10-01T11:59:00+03:00"},
			 "staff":null,"ref":null}
		]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.SubscriptionsHistory(context.Background(), 42)
	if err != nil {
		t.Fatalf("SubscriptionsHistory error: %v", err)
	}
	if res.UserID != 42 || res.AccountID == nil || *res.AccountID != 5 || len(res.History) != 2 {
		t.Fatalf("result = %+v", res)
	}

	revoke, add := res.History[0], res.History[1]
	if len(revoke.Added) != 0 || len(revoke.Removed) != 1 || revoke.Removed[0] != "privetka.pro" {
		t.Fatalf("revoke delta = +%v -%v", revoke.Added, revoke.Removed)
	}
	if len(add.Added) != 2 || add.Added[0] != "cabinet.pro" || add.Added[1] != "privetka.pro" || len(add.Removed) != 0 {
		t.Fatalf("add delta = +%v -%v", add.Added, add.Removed)
	}
	if revoke.Staff == nil || revoke.Staff.Name == nil || *revoke.Staff.Name != "Оператор" || revoke.Payment != nil {
		t.Fatalf("revoke refs = %+v / %+v", revoke.Staff, revoke.Payment)
	}
	if add.Payment == nil || add.Payment.ID == nil || *add.Payment.ID != 7 || add.Payment.DatePaid == nil {
		t.Fatalf("add payment = %+v", add.Payment)
	}
}

func TestSubscriptionsHistoryForAccount_SendsAccountQuery(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42/subscriptions/history")
		if q := r.URL.Query(); q.Get("account_id") != "77" || len(q) != 1 {
			t.Fatalf("query = %v, want account_id=77", q)
		}
		writeSuccess(w, `{"user_id":42,"account_id":77,"history":[]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.SubscriptionsHistoryForAccount(context.Background(), 42, 77)
	if err != nil {
		t.Fatalf("SubscriptionsHistoryForAccount error: %v", err)
	}
	if res.AccountID == nil || *res.AccountID != 77 || res.History == nil || len(res.History) != 0 {
		t.Fatalf("result = %+v", res)
	}
}

// У человека без аккаунта account_id приходит null, а старая CRM не шлёт ни
// его, ни added/removed: всё это должно разобраться без ошибки.
func TestSubscriptionsHistory_NullAccountAndLegacyItems(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"user_id":42,"account_id":null,"history":[
			{"action":"add","bot_id":1,"access":{"main":{"invite":true}},"action_date":"2026-01-01T10:00:00","access_end":null}
		]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.SubscriptionsHistory(context.Background(), 42)
	if err != nil {
		t.Fatalf("SubscriptionsHistory error: %v", err)
	}
	if res.AccountID != nil {
		t.Fatalf("AccountID = %d, want nil", *res.AccountID)
	}
	item := res.History[0]
	if item.Added == nil || item.Removed == nil || len(item.Added) != 0 || len(item.Removed) != 0 {
		t.Fatalf("added/removed = %#v / %#v, want empty non-nil", item.Added, item.Removed)
	}
}

func TestSubscriptionsHistoryForAccount_NotFound(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeAPIError(w, http.StatusNotFound, "not_found", "Аккаунт не найден")
	})

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.SubscriptionsHistoryForAccount(context.Background(), 42, 999)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "not_found" {
		t.Fatalf("err = %v, want APIError not_found", err)
	}
}

func TestSubscriptionsHistory_ValidationWithoutHTTP(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})

	client := mustNewClient(t, server.URL, server.Client())
	ctx := context.Background()
	calls := map[string]func() error{
		"user":         func() error { _, err := client.SubscriptionsHistory(ctx, 0); return err },
		"account user": func() error { _, err := client.SubscriptionsHistoryForAccount(ctx, 0, 77); return err },
		"account":      func() error { _, err := client.SubscriptionsHistoryForAccount(ctx, 42, 0); return err },
	}
	for name, call := range calls {
		var vErr *ValidationError
		if err := call(); !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}
