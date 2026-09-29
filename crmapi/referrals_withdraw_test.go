package crmapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReferralsWithdrawRequestCreated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/referrals/withdraw/request":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"method":"wallet"`) {
				t.Fatalf("body must contain method=wallet, got %s", body)
			}
			fmt.Fprint(w, `{"status":"success","data":{"status":"created","withdrawal_id":7,"amount_usd":12.5,"method":"wallet"}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReferralsWithdrawRequest(context.Background(), 123, "wallet")
	if err != nil {
		t.Fatalf("ReferralsWithdrawRequest() error = %v", err)
	}
	if res.Status != "created" {
		t.Fatalf("Status = %s, want created", res.Status)
	}
	if res.WithdrawalID == nil || *res.WithdrawalID != 7 {
		t.Fatalf("WithdrawalID = %v, want 7", res.WithdrawalID)
	}
	if res.AmountUSD == nil || *res.AmountUSD != 12.5 {
		t.Fatalf("AmountUSD = %v, want 12.5", res.AmountUSD)
	}
	if res.AvailableUSD != nil {
		t.Fatalf("AvailableUSD = %v, want nil", res.AvailableUSD)
	}
}

func TestReferralsWithdrawRequestAlreadyPending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/referrals/withdraw/request":
			fmt.Fprint(w, `{"status":"success","data":{"status":"already_pending","withdrawal_id":9,"amount_usd":30}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReferralsWithdrawRequest(context.Background(), 123, "wallet")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if res.Status != "already_pending" || res.WithdrawalID == nil || *res.WithdrawalID != 9 {
		t.Fatalf("got %+v", res)
	}
}

func TestReferralsWithdrawSettlePartial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/referrals/withdraw/settle":
			body, _ := io.ReadAll(r.Body)
			s := string(body)
			if !strings.Contains(s, `"amount_minor":3000`) {
				t.Fatalf("body must contain amount_minor=3000, got %s", s)
			}
			if !strings.Contains(s, `"withdrawal_id":7`) {
				t.Fatalf("body must contain withdrawal_id=7, got %s", s)
			}
			fmt.Fprint(w, `{"status":"success","data":{"status":"settled","withdrawal_id":7,"paid_usd":30,"available_after_usd":30,"method":"subscription"}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	wid := int64(7)
	res, err := client.ReferralsWithdrawSettle(context.Background(), 123, 3000, "subscription", &wid)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if res.Status != "settled" || res.PaidUSD != 30 || res.AvailableAfterUSD != 30 || res.Method != "subscription" {
		t.Fatalf("got %+v", res)
	}
}

func TestReferralsWithdrawLocalValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("server must not be hit on local validation: %s", r.URL.Path)
	}))
	defer server.Close()
	client := mustNewClient(t, server.URL, server.Client())

	if _, err := client.ReferralsWithdrawRequest(context.Background(), 123, "bank"); err == nil {
		t.Fatalf("expected error for bad method")
	}
	_, err := client.ReferralsWithdrawRequest(context.Background(), 123, "subscription")
	var vErr *ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("subscription in request: err = %v, want *ValidationError", err)
	}
	if vErr.Message != "method must be 'wallet'" {
		t.Fatalf("ValidationError.Message = %q", vErr.Message)
	}
	if _, err := client.ReferralsWithdrawRequest(context.Background(), 0, "wallet"); err == nil {
		t.Fatalf("expected error for non-positive user_id")
	}
	if _, err := client.ReferralsWithdrawSettle(context.Background(), 123, 0, "wallet", nil); err == nil {
		t.Fatalf("expected error for non-positive amount")
	}
	if _, err := client.ReferralsWithdrawSettle(context.Background(), 123, 100, "paypal", nil); err == nil {
		t.Fatalf("expected error for bad method")
	}
}

func TestReferralsWithdrawSettleAcceptsWalletAndSubscription(t *testing.T) {
	var gotMethods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/referrals/withdraw/settle":
			var body struct {
				Method string `json:"method"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			gotMethods = append(gotMethods, body.Method)
			fmt.Fprintf(w, `{"status":"success","data":{"status":"settled","withdrawal_id":1,"paid_usd":1,"available_after_usd":0,"method":%q}}`, body.Method)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	for _, method := range []string{"wallet", "subscription"} {
		res, err := client.ReferralsWithdrawSettle(context.Background(), 123, 100, method, nil)
		if err != nil {
			t.Fatalf("settle %s: error = %v", method, err)
		}
		if res.Method != method {
			t.Fatalf("settle %s: Method = %s", method, res.Method)
		}
	}
	if len(gotMethods) != 2 || gotMethods[0] != "wallet" || gotMethods[1] != "subscription" {
		t.Fatalf("server got methods %v", gotMethods)
	}
}
