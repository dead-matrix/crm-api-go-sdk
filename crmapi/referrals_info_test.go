package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReferralsInfoParsesWithdrawnByMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/referrals/info":
			fmt.Fprint(w, `{"status":"success","data":{
				"ref_link":"https://traffsoft.com/?bot=7",
				"percent":10,"registrations":5,"ref_payments":3,"ref_total_sum":12345,
				"earned_usd":50.00,"available_usd":378.61,
				"withdrawn_wallet_usd":30.00,"withdrawn_subscription_usd":20.00,
				"referrees":[]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReferralsInfo(context.Background(), 7)
	if err != nil {
		t.Fatalf("ReferralsInfo() error = %v", err)
	}
	if res.EarnedUSD != 50.00 || res.AvailableUSD != 378.61 {
		t.Fatalf("earned/available = %v/%v", res.EarnedUSD, res.AvailableUSD)
	}
	// New fields: withdrawn split by method.
	if res.WithdrawnWalletUSD != 30.00 || res.WithdrawnSubscriptionUSD != 20.00 {
		t.Fatalf("withdrawn wallet/subscription = %v/%v, want 30/20", res.WithdrawnWalletUSD, res.WithdrawnSubscriptionUSD)
	}
}

func referralsInfoServer(t *testing.T, data string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/referrals/info":
			fmt.Fprintf(w, `{"status":"success","code":null,"message":null,"data":%s}`, data)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
}

const referralsInfoLegacyFields = `"ref_link":"https://socialtraff.com/?ref=OWNER1",
	"percent":20,"registrations":1,"ref_payments":0,"ref_total_sum":0,
	"earned_usd":0,"available_usd":0,
	"withdrawn_wallet_usd":0,"withdrawn_subscription_usd":0,
	"referrees":[]`

// Тело partner совпадает с data из CRM tests/contract/customer_referrals_active.json.
const referralsInfoPartnerActive = `{
	"available_usd_cents": 0,
	"earned_usd_cents": 0,
	"first_percent": 40,
	"hold_days": 14,
	"min_withdrawal_usd_cents": 2000,
	"on_hold_usd_cents": 0,
	"paid_referrals_count": 0,
	"pending_withdrawal": {
		"amount_usd_cents": 2500,
		"created_at": "2026-09-25T10:00:00+03:00",
		"method": "wallet",
		"withdrawal_id": "wdr_88504592762fafd9"
	},
	"percent": 20,
	"promo_activations_count": 0,
	"promo_codes": [{"code": "OWNER1", "kind": "partner_auto"}],
	"recent_accruals": [
		{"amount_usd_cents": 600, "available_at": "2026-09-20T12:00:00+03:00", "kind": "recurring", "reverted": true},
		{"amount_usd_cents": 600, "available_at": "2026-10-05T12:00:00+03:00", "kind": "recurring", "reverted": false},
		{"amount_usd_cents": 1200, "available_at": "2026-09-05T12:00:00+03:00", "kind": "first", "reverted": false}
	],
	"recurring_percent": 20,
	"ref_bot_link": "https://t.me/socialtraff_robot?start=ref_OWNER1",
	"ref_link": "https://socialtraff.com/?ref=OWNER1",
	"referred_payments_count": 0,
	"referred_turnover_rub_kopecks": 0,
	"registrations": 1,
	"spent_on_subscriptions_usd_cents": 0,
	"withdraw_methods": ["wallet"],
	"withdrawn_subscription_usd_cents": 0,
	"withdrawn_wallet_usd_cents": 0
}`

func TestReferralsInfoParsesPartner(t *testing.T) {
	server := referralsInfoServer(t, `{`+referralsInfoLegacyFields+`,"partner":`+referralsInfoPartnerActive+`}`)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReferralsInfo(context.Background(), 7)
	if err != nil {
		t.Fatalf("ReferralsInfo() error = %v", err)
	}
	p := res.Partner
	if p == nil {
		t.Fatalf("Partner = nil, want parsed object")
	}
	if p.RefLink != "https://socialtraff.com/?ref=OWNER1" || p.RefBotLink != "https://t.me/socialtraff_robot?start=ref_OWNER1" {
		t.Fatalf("links = %q / %q", p.RefLink, p.RefBotLink)
	}
	if p.Percent != 20 || p.FirstPercent != 40 || p.RecurringPercent != 20 || p.HoldDays != 14 {
		t.Fatalf("percents/hold = %d/%d/%d/%d", p.Percent, p.FirstPercent, p.RecurringPercent, p.HoldDays)
	}
	if p.Registrations != 1 || p.MinWithdrawalUSDCents != 2000 {
		t.Fatalf("registrations/min = %d/%d", p.Registrations, p.MinWithdrawalUSDCents)
	}
	if len(p.WithdrawMethods) != 1 || p.WithdrawMethods[0] != "wallet" {
		t.Fatalf("WithdrawMethods = %v", p.WithdrawMethods)
	}
	if len(p.PromoCodes) != 1 || p.PromoCodes[0] != (PartnerPromoCode{Code: "OWNER1", Kind: "partner_auto"}) {
		t.Fatalf("PromoCodes = %+v", p.PromoCodes)
	}

	pw := p.PendingWithdrawal
	if pw == nil {
		t.Fatalf("PendingWithdrawal = nil")
	}
	if pw.WithdrawalID != "wdr_88504592762fafd9" || pw.AmountUSDCents != 2500 || pw.Method != "wallet" {
		t.Fatalf("PendingWithdrawal = %+v", pw)
	}
	wantCreated := time.Date(2026, 9, 25, 7, 0, 0, 0, time.UTC)
	if pw.CreatedAt == nil || !pw.CreatedAt.Equal(wantCreated) {
		t.Fatalf("CreatedAt = %v, want %v", pw.CreatedAt, wantCreated)
	}

	if len(p.RecentAccruals) != 3 {
		t.Fatalf("RecentAccruals len = %d, want 3", len(p.RecentAccruals))
	}
	a0, a2 := p.RecentAccruals[0], p.RecentAccruals[2]
	if a0.AmountUSDCents != 600 || a0.Kind != "recurring" || !a0.Reverted {
		t.Fatalf("accrual[0] = %+v", a0)
	}
	if a2.AmountUSDCents != 1200 || a2.Kind != "first" || a2.Reverted {
		t.Fatalf("accrual[2] = %+v", a2)
	}
	wantAvail := time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)
	if a0.AvailableAt == nil || !a0.AvailableAt.Equal(wantAvail) {
		t.Fatalf("accrual[0].AvailableAt = %v, want %v", a0.AvailableAt, wantAvail)
	}
}

func TestReferralsInfoPartnerNegativeHoldAndNullPending(t *testing.T) {
	partner := `{"ref_link":"r","ref_bot_link":"b","percent":20,"first_percent":40,
		"recurring_percent":20,"hold_days":14,"registrations":0,"paid_referrals_count":0,
		"promo_activations_count":0,"referred_payments_count":0,"referred_turnover_rub_kopecks":150000,
		"earned_usd_cents":900,"available_usd_cents":0,"on_hold_usd_cents":-300,
		"spent_on_subscriptions_usd_cents":0,"withdrawn_wallet_usd_cents":0,
		"withdrawn_subscription_usd_cents":0,"min_withdrawal_usd_cents":2000,
		"withdraw_methods":["wallet"],"promo_codes":[],"pending_withdrawal":null,"recent_accruals":[]}`
	server := referralsInfoServer(t, `{`+referralsInfoLegacyFields+`,"partner":`+partner+`}`)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReferralsInfo(context.Background(), 7)
	if err != nil {
		t.Fatalf("ReferralsInfo() error = %v", err)
	}
	p := res.Partner
	if p == nil {
		t.Fatalf("Partner = nil")
	}
	if p.OnHoldUSDCents != -300 || p.EarnedUSDCents != 900 || p.ReferredTurnoverRUBKopecks != 150000 {
		t.Fatalf("partner sums = %+v", p)
	}
	if p.PendingWithdrawal != nil {
		t.Fatalf("PendingWithdrawal = %+v, want nil", p.PendingWithdrawal)
	}
	if p.PromoCodes == nil || len(p.PromoCodes) != 0 || p.RecentAccruals == nil || len(p.RecentAccruals) != 0 {
		t.Fatalf("empty lists must be non-nil and empty: %v / %v", p.PromoCodes, p.RecentAccruals)
	}
}

func TestReferralsInfoWithoutPartnerFromLegacyCRM(t *testing.T) {
	server := referralsInfoServer(t, `{`+referralsInfoLegacyFields+`}`)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReferralsInfo(context.Background(), 7)
	if err != nil {
		t.Fatalf("ReferralsInfo() error = %v", err)
	}
	if res.Partner != nil {
		t.Fatalf("Partner = %+v, want nil for legacy CRM", res.Partner)
	}
	if res.RefLink != "https://socialtraff.com/?ref=OWNER1" || res.Percent != 20 {
		t.Fatalf("legacy fields = %q / %d", res.RefLink, res.Percent)
	}
}
