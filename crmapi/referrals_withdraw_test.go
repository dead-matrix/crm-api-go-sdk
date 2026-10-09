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

// Суммы заявки в центах: какие поля заполнены, решает статус, остальные nil.
func TestReferralsWithdrawRequest_MinorAmountsByStatus(t *testing.T) {
	response := ""
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/referrals/withdraw/request")
		writeSuccess(w, response)
	})
	client := mustNewClient(t, server.URL, server.Client())
	request := func(data string) *WithdrawRequestResult {
		t.Helper()
		response = data
		res, err := client.ReferralsWithdrawRequest(context.Background(), 9001, "wallet")
		if err != nil {
			t.Fatalf("ReferralsWithdrawRequest error: %v", err)
		}
		return res
	}

	res := request(`{"status":"below_min","available_minor":730,"min_minor":1000,"available_usd":7.3,"min_usd":10.0}`)
	if res.Status != "below_min" {
		t.Fatalf("Status = %s, want below_min", res.Status)
	}
	if res.AvailableMinor == nil || *res.AvailableMinor != 730 || res.MinMinor == nil || *res.MinMinor != 1000 {
		t.Fatalf("below_min minor = %v / %v", res.AvailableMinor, res.MinMinor)
	}
	if res.AvailableUSD == nil || *res.AvailableUSD != 7.3 || res.MinUSD == nil || *res.MinUSD != 10 {
		t.Fatalf("below_min usd = %v / %v", res.AvailableUSD, res.MinUSD)
	}
	if res.WithdrawalID != nil || res.AmountMinor != nil || res.AmountUSD != nil || res.Method != nil {
		t.Fatalf("below_min must not carry a withdrawal: %+v", res)
	}

	res = request(`{"status":"created","withdrawal_id":7,"amount_minor":1250,"amount_usd":12.5,"method":"wallet"}`)
	if res.Status != "created" || res.AmountMinor == nil || *res.AmountMinor != 1250 || res.WithdrawalID == nil || *res.WithdrawalID != 7 {
		t.Fatalf("created = %+v", res)
	}
	if res.AvailableMinor != nil || res.MinMinor != nil || res.MinUSD != nil {
		t.Fatalf("created must not carry balance fields: %+v", res)
	}

	res = request(`{"status":"already_pending","withdrawal_id":9,"amount_minor":3000,"amount_usd":30.0}`)
	if res.Status != "already_pending" || res.AmountMinor == nil || *res.AmountMinor != 3000 || res.Method != nil {
		t.Fatalf("already_pending = %+v", res)
	}

	// Ноль в no_balance - значение, а не отсутствие: указатель не nil.
	res = request(`{"status":"no_balance","available_minor":0,"available_usd":0.0}`)
	if res.Status != "no_balance" || res.AvailableMinor == nil || *res.AvailableMinor != 0 || res.MinMinor != nil {
		t.Fatalf("no_balance = %+v", res)
	}
}

// Повтор по закрытой заявке и неизвестный id приходят успехом со своим
// статусом: деньги не двигались, и по current_status видно, чем заявка кончилась.
func TestReferralsWithdrawSettle_AlreadySettledAndNotFound(t *testing.T) {
	response := ""
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodPost, "/api/referrals/withdraw/settle")
		writeSuccess(w, response)
	})
	client := mustNewClient(t, server.URL, server.Client())
	wid := int64(7)

	response = `{"status":"already_settled","withdrawal_id":7,"current_status":"rejected"}`
	res, err := client.ReferralsWithdrawSettle(context.Background(), 9001, 3000, "wallet", &wid)
	if err != nil {
		t.Fatalf("settle error: %v", err)
	}
	if res.Status != "already_settled" || res.WithdrawalID != 7 || res.CurrentStatus == nil || *res.CurrentStatus != "rejected" {
		t.Fatalf("already_settled = %+v", res)
	}
	if res.PaidUSD != 0 || res.Method != "" {
		t.Fatalf("already_settled must not report a payout: %+v", res)
	}

	response = `{"status":"not_found","withdrawal_id":7}`
	res, err = client.ReferralsWithdrawSettle(context.Background(), 9001, 3000, "wallet", &wid)
	if err != nil {
		t.Fatalf("settle error: %v", err)
	}
	if res.Status != "not_found" || res.CurrentStatus != nil {
		t.Fatalf("not_found = %+v", res)
	}

	response = `{"status":"settled","withdrawal_id":7,"paid_usd":30,"available_after_usd":0,"method":"wallet"}`
	res, err = client.ReferralsWithdrawSettle(context.Background(), 9001, 3000, "wallet", &wid)
	if err != nil {
		t.Fatalf("settle error: %v", err)
	}
	if res.Status != "settled" || res.CurrentStatus != nil || res.PaidUSD != 30 {
		t.Fatalf("settled = %+v", res)
	}
}
