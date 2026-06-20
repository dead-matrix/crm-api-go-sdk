package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
