package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestActivationRedeem_NullPaymentID: сервер может вернуть payment_id=null
// (код активации не был привязан к платежу — например, технический код).
// До перехода PaymentID на *int64 SDK декодировал null в 0, что неотличимо
// от валидного payment_id.
func TestActivationRedeem_NullPaymentID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/activation/redeem":
			fmt.Fprint(w, `{"status":"success","data":{
				"user_id":42,
				"bot_id":1,
				"action":"add",
				"access":{"main":{"invite":true}},
				"access_end":"2030-01-01T00:00:00Z",
				"activation_code_id":7,
				"payment_id":null
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ActivationRedeem(context.Background(), ActivationRedeemInput{
		Token: "ACT_abc", RecipientUserID: 42, BotID: 1,
	})
	if err != nil {
		t.Fatalf("ActivationRedeem error: %v", err)
	}
	if !res.Success {
		t.Fatalf("Success = false, want true")
	}
	if res.PaymentID != nil {
		t.Fatalf("PaymentID = %d, want nil (server returned null)", *res.PaymentID)
	}
	if res.ActivationCodeID != 7 {
		t.Fatalf("ActivationCodeID = %d, want 7", res.ActivationCodeID)
	}
}

// TestActivationRedeem_NonNullPaymentID: на штатном пути payment_id !=
// null, и Result.PaymentID — указатель на корректное значение.
func TestActivationRedeem_NonNullPaymentID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/activation/redeem":
			fmt.Fprint(w, `{"status":"success","data":{
				"user_id":42,
				"bot_id":1,
				"action":"add",
				"access":null,
				"access_end":"2030-01-01T00:00:00Z",
				"activation_code_id":7,
				"payment_id":12345
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ActivationRedeem(context.Background(), ActivationRedeemInput{
		Token: "ACT_abc", RecipientUserID: 42, BotID: 1,
	})
	if err != nil {
		t.Fatalf("ActivationRedeem error: %v", err)
	}
	if res.PaymentID == nil || *res.PaymentID != 12345 {
		t.Fatalf("PaymentID = %v, want pointer to 12345", res.PaymentID)
	}
}
