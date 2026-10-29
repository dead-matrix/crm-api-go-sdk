package crmapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Вызов только по аккаунту: user_id в теле быть не должно (CRM отвергает 0),
// а null в ответе превращается в UserID=0.
func TestAddAccess_AccountOnly(t *testing.T) {
	var body map[string]any
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/access/add")
		body = readJSONBody(t, r)
		writeSuccess(w, `{"created":true,"id":501,"account_id":77,"user_id":null,"bot_id":10,
			"action":"remove","action_date":"2026-10-09T12:00:00+03:00","access_end":null}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	accountID := int64(77)
	key := "op-2026-10-09-1"
	res, err := client.AddAccess(context.Background(), AddAccessInput{
		AccountID:      &accountID,
		BotID:          10,
		Action:         ActionRemove,
		Access:         map[string]bool{"cabinet.pro": true},
		IdempotencyKey: &key,
	})
	if err != nil {
		t.Fatalf("AddAccess error: %v", err)
	}

	if _, sent := body["user_id"]; sent {
		t.Fatalf("user_id must be omitted for account-only call, body = %v", body)
	}
	if body["account_id"] != float64(77) || body["action"] != "remove" || body["idempotency_key"] != key {
		t.Fatalf("body = %v", body)
	}
	if res.AccountID == nil || *res.AccountID != 77 {
		t.Fatalf("AccountID = %v, want 77", res.AccountID)
	}
	if res.UserID != 0 {
		t.Fatalf("UserID = %d, want 0 for null", res.UserID)
	}
	if res.ID == nil || *res.ID != 501 || res.Action != "remove" || res.AccessEnd != nil {
		t.Fatalf("result = %+v", res)
	}
	want := time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC)
	if res.ActionDate == nil || !res.ActionDate.Equal(want) {
		t.Fatalf("ActionDate = %v, want %v", res.ActionDate, want)
	}
}

// Прежний вызов по человеку не меняется на проводе: user_id есть, новых
// необязательных ключей нет.
func TestAddAccess_UserOnlyKeepsWireFormat(t *testing.T) {
	var body map[string]any
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/access/add")
		body = readJSONBody(t, r)
		writeSuccess(w, `{"created":true,"id":1,"account_id":5,"user_id":42,"bot_id":10,"action":"add"}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	end := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	res, err := client.AddAccess(context.Background(), AddAccessInput{
		UserID: 42, BotID: 10, Action: ActionAdd, AccessEnd: &end,
	})
	if err != nil {
		t.Fatalf("AddAccess error: %v", err)
	}
	if body["user_id"] != float64(42) {
		t.Fatalf("user_id = %v, want 42", body["user_id"])
	}
	for _, key := range []string{"account_id", "idempotency_key"} {
		if _, sent := body[key]; sent {
			t.Fatalf("%s must be omitted when not set, body = %v", key, body)
		}
	}
	if res.UserID != 42 || res.AccountID == nil || *res.AccountID != 5 {
		t.Fatalf("result = %+v", res)
	}
}

// Старая CRM ключ account_id не присылает: это не ошибка, поле остаётся nil.
func TestAddAccess_LegacyResponseWithoutAccountID(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"created":true,"id":1,"user_id":42,"bot_id":1,"action":"add"}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.AddAccess(context.Background(), AddAccessInput{UserID: 42, BotID: 1, Action: ActionAdd})
	if err != nil {
		t.Fatalf("AddAccess error: %v", err)
	}
	if res.AccountID != nil {
		t.Fatalf("AccountID = %d, want nil", *res.AccountID)
	}
}

func TestAddAccessInput_Validate(t *testing.T) {
	zero, negative, account := int64(0), int64(-1), int64(7)
	tooLong := strings.Repeat("k", 65)
	// 64 кириллических символа - это 128 байт: проверка идёт по символам.
	cyrillic := strings.Repeat("я", 64)

	invalid := map[string]AddAccessInput{
		"no target":        {BotID: 10, Action: ActionAdd},
		"negative user":    {UserID: -5, AccountID: &account, BotID: 10, Action: ActionAdd},
		"zero account":     {AccountID: &zero, BotID: 10, Action: ActionAdd},
		"negative account": {UserID: 1, AccountID: &negative, BotID: 10, Action: ActionAdd},
		"long key":         {UserID: 1, BotID: 10, Action: ActionAdd, IdempotencyKey: &tooLong},
		"unknown action":   {UserID: 1, BotID: 10, Action: "delete"},
	}
	for name, in := range invalid {
		var vErr *ValidationError
		if err := in.Validate(); !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}

	valid := map[string]AddAccessInput{
		"account only":     {AccountID: &account, BotID: 10, Action: ActionRemove},
		"user and account": {UserID: 1, AccountID: &account, BotID: 10, Action: ActionRevoke},
		"cyrillic key":     {UserID: 1, BotID: 10, Action: ActionAdd, IdempotencyKey: &cyrillic},
		"remove any case":  {UserID: 1, BotID: 10, Action: " Remove "},
	}
	for name, in := range valid {
		if err := in.Validate(); err != nil {
			t.Fatalf("%s: unexpected error %v", name, err)
		}
	}
}

func TestManageAccess_AccountOnly(t *testing.T) {
	var body map[string]any
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/access/manage")
		body = readJSONBody(t, r)
		writeSuccess(w, `{"account_id":77,"user_id":null,"bot_id":10,"op":"grant","action":"add",
			"access":{"cabinet.pro":true},"access_end":"2026-11-08T12:00:00+03:00","crm_access_id":900}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	accountID := int64(77)
	days := 30
	res, err := client.ManageAccess(context.Background(), AccessManageInput{
		AccountID: &accountID,
		BotID:     10,
		Op:        "grant",
		Features:  []string{"cabinet.pro"},
		Days:      &days,
	})
	if err != nil {
		t.Fatalf("ManageAccess error: %v", err)
	}
	if _, sent := body["user_id"]; sent {
		t.Fatalf("user_id must be omitted for account-only call, body = %v", body)
	}
	if body["account_id"] != float64(77) || body["op"] != "grant" || body["days"] != float64(30) {
		t.Fatalf("body = %v", body)
	}
	if res.AccountID == nil || *res.AccountID != 77 || res.UserID != 0 {
		t.Fatalf("AccountID/UserID = %v/%d, want 77/0", res.AccountID, res.UserID)
	}
	if res.CrmAccessID == nil || *res.CrmAccessID != 900 || res.Action != "add" || res.AccessEnd == nil {
		t.Fatalf("result = %+v", res)
	}
}

func TestManageAccess_UserOnlyKeepsWireFormat(t *testing.T) {
	var body map[string]any
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body = readJSONBody(t, r)
		writeSuccess(w, `{"account_id":5,"user_id":42,"bot_id":10,"op":"revoke_all","action":"revoke","access":null,"access_end":null,"crm_access_id":3}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ManageAccess(context.Background(), AccessManageInput{UserID: 42, BotID: 10, Op: "revoke_all"})
	if err != nil {
		t.Fatalf("ManageAccess error: %v", err)
	}
	if body["user_id"] != float64(42) {
		t.Fatalf("user_id = %v, want 42", body["user_id"])
	}
	if _, sent := body["account_id"]; sent {
		t.Fatalf("account_id must be omitted when not set, body = %v", body)
	}
	if res.UserID != 42 || res.AccountID == nil || *res.AccountID != 5 {
		t.Fatalf("result = %+v", res)
	}
}

func TestManageAccess_ValidationWithoutHTTP(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})

	client := mustNewClient(t, server.URL, server.Client())
	zero := int64(0)
	cases := map[string]AccessManageInput{
		"no target":     {BotID: 10, Op: "grant"},
		"zero account":  {AccountID: &zero, BotID: 10, Op: "grant"},
		"negative user": {UserID: -1, BotID: 10, Op: "grant"},
	}
	for name, in := range cases {
		_, err := client.ManageAccess(context.Background(), in)
		var vErr *ValidationError
		if !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}
