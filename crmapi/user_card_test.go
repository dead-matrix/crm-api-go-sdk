package crmapi

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// Полная карточка CRM SocialTraff: покупатель, два аккаунта с новыми полями и
// account_id в bots_info.
func TestGetUser_BuyerAndAccounts(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42")
		writeSuccess(w, `{
			"user_id":42,"full_name":"Иван","username":"qz__vp","status":null,
			"has_active_subscription":true,"frozen":false,
			"buyer":{"buyer_id":9001,"email":"a@example.com","display_name":"Иван","ref_code":"OWNER1",
				"refer":"utm_source=tg","landing":"/pricing","created_at":"2026-09-01T10:00:00+03:00"},
			"bots_info":[{"bot_id":10,"bot_name":"SocialTraff","account_id":5,
				"registered":"2026-09-01T10:00:00+03:00","refer":null,
				"access":{"cabinet.pro":true},"access_end":"2026-11-01T00:00:00+03:00",
				"access_expiry":{"cabinet.pro":"2026-11-01T00:00:00+03:00"},
				"frozen":false,"frozen_at":null,"frozen_expiry":null}],
			"accounts":[
				{"account_id":5,"title":"Личный","role":"owner",
				 "access":{"cabinet.pro":true},
				 "access_expiry":{"cabinet.pro":"2026-11-01T00:00:00+03:00"},
				 "access_end":"2026-11-01T00:00:00+03:00",
				 "is_personal":true,"is_primary":true,"owner_buyer_id":9001,
				 "created_at":"2026-09-01T10:00:00+03:00","joined_at":"2026-09-01T10:00:00+03:00",
				 "members_count":3,"chats_count":2,"bots_count":1},
				{"account_id":77,"title":"Команда","role":"viewer",
				 "access":null,"access_expiry":null,"access_end":null,
				 "is_personal":false,"is_primary":false,"owner_buyer_id":9002,
				 "created_at":"2026-09-10T10:00:00+03:00","joined_at":"2026-09-15T18:30:00+03:00",
				 "members_count":8,"chats_count":0,"bots_count":0}
			]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}

	buyer := user.Buyer
	if buyer == nil || buyer.BuyerID != 9001 {
		t.Fatalf("Buyer = %+v, want buyer_id 9001", buyer)
	}
	if buyer.Email == nil || *buyer.Email != "a@example.com" || buyer.RefCode == nil || *buyer.RefCode != "OWNER1" {
		t.Fatalf("Buyer email/ref_code = %v/%v", buyer.Email, buyer.RefCode)
	}
	if buyer.Refer == nil || *buyer.Refer != "utm_source=tg" || buyer.Landing == nil || *buyer.Landing != "/pricing" {
		t.Fatalf("Buyer refer/landing = %v/%v", buyer.Refer, buyer.Landing)
	}
	// +03:00 обязан дойти до UTC, а не потеряться: 10:00 МСК = 07:00 UTC.
	created := time.Date(2026, 9, 1, 7, 0, 0, 0, time.UTC)
	if buyer.CreatedAt == nil || !buyer.CreatedAt.Equal(created) {
		t.Fatalf("Buyer.CreatedAt = %v, want %v", buyer.CreatedAt, created)
	}

	if len(user.BotsInfo) != 1 || user.BotsInfo[0].AccountID == nil || *user.BotsInfo[0].AccountID != 5 {
		t.Fatalf("BotsInfo = %+v, want account_id 5", user.BotsInfo)
	}

	if len(user.Accounts) != 2 {
		t.Fatalf("len(Accounts) = %d, want 2", len(user.Accounts))
	}
	personal, team := user.Accounts[0], user.Accounts[1]
	if personal.AccountID != 5 || personal.Title != "Личный" || personal.Role != "owner" {
		t.Fatalf("personal = %+v", personal)
	}
	if !personal.Access["cabinet.pro"] || personal.AccessExpiry["cabinet.pro"] != "2026-11-01T00:00:00+03:00" {
		t.Fatalf("personal access = %v / %v", personal.Access, personal.AccessExpiry)
	}
	end := time.Date(2026, 10, 31, 21, 0, 0, 0, time.UTC)
	if personal.AccessEnd == nil || !personal.AccessEnd.Equal(end) {
		t.Fatalf("personal.AccessEnd = %v, want %v", personal.AccessEnd, end)
	}
	if !personal.IsPersonal || !personal.IsPrimary || personal.OwnerBuyerID != 9001 {
		t.Fatalf("personal flags = %+v", personal)
	}
	if personal.CreatedAt == nil || personal.JoinedAt == nil || !personal.JoinedAt.Equal(created) {
		t.Fatalf("personal dates = %v / %v", personal.CreatedAt, personal.JoinedAt)
	}
	if personal.MembersCount != 3 || personal.ChatsCount != 2 || personal.BotsCount != 1 {
		t.Fatalf("personal counts = %d/%d/%d", personal.MembersCount, personal.ChatsCount, personal.BotsCount)
	}

	if team.AccountID != 77 || team.Role != "viewer" || team.IsPersonal || team.IsPrimary || team.OwnerBuyerID != 9002 {
		t.Fatalf("team = %+v", team)
	}
	if team.Access != nil || team.AccessExpiry != nil || team.AccessEnd != nil {
		t.Fatalf("team access must be nil for null, got %v / %v / %v", team.Access, team.AccessExpiry, team.AccessEnd)
	}
	joined := time.Date(2026, 9, 15, 15, 30, 0, 0, time.UTC)
	if team.JoinedAt == nil || !team.JoinedAt.Equal(joined) || team.MembersCount != 8 {
		t.Fatalf("team joined/members = %v/%d", team.JoinedAt, team.MembersCount)
	}
}

// Человек писал боту, но покупателем не стал: buyer=null и пустой accounts.
func TestGetUser_NotACustomer(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"user_id":42,"full_name":"Иван","username":null,"status":null,
			"has_active_subscription":false,"frozen":false,"buyer":null,
			"bots_info":[{"bot_id":10,"bot_name":"SocialTraff","account_id":null,"registered":null,"refer":null,
				"access":null,"access_end":null,"access_expiry":null,"frozen":false,"frozen_at":null,"frozen_expiry":null}],
			"accounts":[]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if user.Buyer != nil {
		t.Fatalf("Buyer = %+v, want nil", user.Buyer)
	}
	if user.Accounts == nil || len(user.Accounts) != 0 {
		t.Fatalf("Accounts = %#v, want empty non-nil", user.Accounts)
	}
	if user.BotsInfo[0].AccountID != nil {
		t.Fatalf("BotsInfo[0].AccountID = %d, want nil", *user.BotsInfo[0].AccountID)
	}
}

// CRM до этих правок: нет buyer, а у аккаунта только пять старых ключей.
// Отсутствие новых ключей не ошибка, они остаются нулевыми.
func TestGetUser_ServerWithoutNewAccountFields(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"user_id":42,"full_name":"Иван","username":null,"status":null,
			"has_active_subscription":true,"frozen":false,
			"bots_info":[{"bot_id":10,"bot_name":"SocialTraff","account_id":5}],
			"accounts":[{"account_id":5,"title":"Личный","role":"owner",
				"access":{"cabinet.pro":true},"access_expiry":{"cabinet.pro":"2026-11-01T00:00:00+03:00"}}]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if user.Buyer != nil || len(user.Accounts) != 1 {
		t.Fatalf("Buyer/Accounts = %+v / %+v", user.Buyer, user.Accounts)
	}
	acc := user.Accounts[0]
	if acc.AccountID != 5 || acc.Role != "owner" || !acc.Access["cabinet.pro"] {
		t.Fatalf("account = %+v", acc)
	}
	if acc.AccessEnd != nil || acc.CreatedAt != nil || acc.JoinedAt != nil {
		t.Fatalf("dates must be nil when absent: %v / %v / %v", acc.AccessEnd, acc.CreatedAt, acc.JoinedAt)
	}
	if acc.IsPersonal || acc.IsPrimary || acc.OwnerBuyerID != 0 || acc.MembersCount != 0 || acc.ChatsCount != 0 || acc.BotsCount != 0 {
		t.Fatalf("absent fields must stay zero: %+v", acc)
	}
}

// CRM совсем без ключа accounts (другая инсталляция): срез пустой, не nil.
func TestGetUser_ServerWithoutAccountsKey(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"user_id":42,"full_name":"Иван","status":"new","has_active_subscription":false,"frozen":false,
			"bots_info":[{"bot_id":1,"bot_name":"Main","registered":"2026-01-01T10:00:00","access":{"main":{"invite":true}}}]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if user.Accounts == nil || len(user.Accounts) != 0 {
		t.Fatalf("Accounts = %#v, want empty non-nil", user.Accounts)
	}
	if len(user.BotsInfo) != 1 || user.BotsInfo[0].AccountID != nil || user.BotsInfo[0].Registered == nil {
		t.Fatalf("BotsInfo = %+v", user.BotsInfo)
	}
}

// У строки доступа без карты сроков CRM отдаёт access как лежит в базе, и
// значения там бывают не булевыми. Карточка обязана разобраться, а живость
// фичи считается по истинности значения, как это делает сама CRM.
func TestGetUser_AccountAccessWithNonBooleanValues(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"user_id":42,"has_active_subscription":true,"frozen":false,"bots_info":[],
			"accounts":[{"account_id":5,"title":"Личный","role":"owner","access_expiry":null,
				"access":{"flag":true,"off":false,"one":1,"zero":0,"nested":{"invite":true},
					"empty":{},"text":"yes","blank":"","none":null,"list":[1],"nolist":[]}}]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	got := user.Accounts[0].Access
	want := map[string]bool{
		"flag": true, "off": false, "one": true, "zero": false, "nested": true,
		"empty": false, "text": true, "blank": false, "none": false, "list": true, "nolist": false,
	}
	if len(got) != len(want) {
		t.Fatalf("access = %v, want %v", got, want)
	}
	for key, live := range want {
		if value, ok := got[key]; !ok || value != live {
			t.Fatalf("access[%q] = %v (present=%v), want %v", key, value, ok, live)
		}
	}
}
