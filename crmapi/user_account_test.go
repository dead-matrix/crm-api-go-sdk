package crmapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestGetUserAccount_FullCard(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42/accounts/77")
		if len(r.URL.Query()) != 0 {
			t.Fatalf("query = %v, want empty", r.URL.Query())
		}
		writeSuccess(w, `{
			"account_id":77,"title":"Команда","is_personal":false,"is_primary":true,
			"created_at":"2026-09-10T10:00:00+03:00","owner_buyer_id":9002,"role":"admin",
			"access":{"cabinet.agency":true,"privetka.pro":true},
			"access_end":"2026-12-01T00:00:00+03:00",
			"access_expiry":{"cabinet.agency":"2026-12-01T00:00:00+03:00","privetka.pro":"2026-11-01T00:00:00+03:00"},
			"members":[
				{"buyer_id":9002,"tg_id":555,"email":"owner@example.com","display_name":"Владелец","role":"owner","joined_at":"2026-09-10T10:00:00+03:00"},
				{"buyer_id":9001,"tg_id":null,"email":null,"display_name":null,"role":"admin","joined_at":"2026-09-15T18:30:00+03:00"}
			],
			"chats":[{"tg_chat_id":-1001234567890,"type":"channel","linked_at":"2026-09-11T09:00:00+03:00"}],
			"bots":[{"bot_telegram_id":7000000001,"linked_at":"2026-09-12T09:00:00+03:00"}],
			"promo_pending":[{"code":"WELCOME","effect":"discount_percent","value":20,
				"first_subscription_only":true,"status":"pending","created_at":"2026-10-01T12:00:00+03:00"}]
		}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	card, err := client.GetUserAccount(context.Background(), 42, 77)
	if err != nil {
		t.Fatalf("GetUserAccount error: %v", err)
	}

	if card.AccountID != 77 || card.Title != "Команда" || card.IsPersonal || !card.IsPrimary {
		t.Fatalf("card head = %+v", card)
	}
	if card.OwnerBuyerID != 9002 || card.Role != "admin" {
		t.Fatalf("owner/role = %d/%s", card.OwnerBuyerID, card.Role)
	}
	created := time.Date(2026, 9, 10, 7, 0, 0, 0, time.UTC)
	if card.CreatedAt == nil || !card.CreatedAt.Equal(created) {
		t.Fatalf("CreatedAt = %v, want %v", card.CreatedAt, created)
	}
	if len(card.Access) != 2 || !card.Access["cabinet.agency"] || !card.Access["privetka.pro"] {
		t.Fatalf("Access = %v", card.Access)
	}
	if card.AccessExpiry["privetka.pro"] != "2026-11-01T00:00:00+03:00" {
		t.Fatalf("AccessExpiry = %v", card.AccessExpiry)
	}
	end := time.Date(2026, 11, 30, 21, 0, 0, 0, time.UTC)
	if card.AccessEnd == nil || !card.AccessEnd.Equal(end) {
		t.Fatalf("AccessEnd = %v, want %v", card.AccessEnd, end)
	}

	if len(card.Members) != 2 {
		t.Fatalf("len(Members) = %d, want 2", len(card.Members))
	}
	owner, admin := card.Members[0], card.Members[1]
	if owner.BuyerID != 9002 || owner.TgID == nil || *owner.TgID != 555 || owner.Role != "owner" {
		t.Fatalf("owner = %+v", owner)
	}
	if owner.Email == nil || *owner.Email != "owner@example.com" || owner.DisplayName == nil || owner.JoinedAt == nil {
		t.Fatalf("owner contacts = %+v", owner)
	}
	if admin.BuyerID != 9001 || admin.TgID != nil || admin.Email != nil || admin.DisplayName != nil || admin.Role != "admin" {
		t.Fatalf("admin = %+v, nullable fields must stay nil", admin)
	}
	joined := time.Date(2026, 9, 15, 15, 30, 0, 0, time.UTC)
	if admin.JoinedAt == nil || !admin.JoinedAt.Equal(joined) {
		t.Fatalf("admin.JoinedAt = %v, want %v", admin.JoinedAt, joined)
	}

	if len(card.Chats) != 1 || card.Chats[0].TgChatID != -1001234567890 || card.Chats[0].Type != "channel" || card.Chats[0].LinkedAt == nil {
		t.Fatalf("Chats = %+v", card.Chats)
	}
	if len(card.Bots) != 1 || card.Bots[0].BotTelegramID != 7000000001 || card.Bots[0].LinkedAt == nil {
		t.Fatalf("Bots = %+v", card.Bots)
	}
	if len(card.PromoPending) != 1 {
		t.Fatalf("PromoPending = %+v", card.PromoPending)
	}
	promo := card.PromoPending[0]
	if promo.Code != "WELCOME" || promo.Effect != "discount_percent" || promo.Value != 20 ||
		!promo.FirstSubscriptionOnly || promo.Status != "pending" || promo.CreatedAt == nil {
		t.Fatalf("promo = %+v", promo)
	}
}

// Аккаунт без доступа и без связей: null и пропущенные массивы дают nil-карты
// и пустые, но не nil, срезы.
func TestGetUserAccount_EmptyCollections(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, `{"account_id":5,"title":"Личный","is_personal":true,"is_primary":true,
			"created_at":"2026-09-01T10:00:00+03:00","owner_buyer_id":9001,"role":"owner",
			"access":null,"access_end":null,"access_expiry":null,
			"members":[],"chats":null,"promo_pending":[]}`)
	})

	client := mustNewClient(t, server.URL, server.Client())
	card, err := client.GetUserAccount(context.Background(), 42, 5)
	if err != nil {
		t.Fatalf("GetUserAccount error: %v", err)
	}
	if card.Access != nil || card.AccessExpiry != nil || card.AccessEnd != nil {
		t.Fatalf("access must be nil for null: %v / %v / %v", card.Access, card.AccessExpiry, card.AccessEnd)
	}
	if card.Members == nil || card.Chats == nil || card.Bots == nil || card.PromoPending == nil {
		t.Fatalf("slices must never be nil: %#v %#v %#v %#v", card.Members, card.Chats, card.Bots, card.PromoPending)
	}
	if len(card.Members)+len(card.Chats)+len(card.Bots)+len(card.PromoPending) != 0 {
		t.Fatalf("slices must be empty: %+v", card)
	}
}

// Человек не участник аккаунта: 404 not_found приходит обычным *APIError.
func TestGetUserAccount_NotAMember(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42/accounts/999")
		writeAPIError(w, http.StatusNotFound, "not_found", "Аккаунт не найден")
	})

	client := mustNewClient(t, server.URL, server.Client())
	card, err := client.GetUserAccount(context.Background(), 42, 999)
	if card != nil {
		t.Fatalf("card = %+v, want nil on error", card)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v (%T), want *APIError", err, err)
	}
	if apiErr.Code != "not_found" || apiErr.Status != http.StatusNotFound {
		t.Fatalf("APIError = %+v, want not_found/404", apiErr)
	}
}

func TestGetUserAccount_ValidationWithoutHTTP(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})

	client := mustNewClient(t, server.URL, server.Client())
	for name, ids := range map[string][2]int64{
		"zero user":        {0, 77},
		"negative user":    {-1, 77},
		"zero account":     {42, 0},
		"negative account": {42, -5},
	} {
		_, err := client.GetUserAccount(context.Background(), ids[0], ids[1])
		var vErr *ValidationError
		if !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}
