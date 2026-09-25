package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// ----------------------------------------------------------------------------
// GetPayments (GET /api/payments) — paginated envelope
// ----------------------------------------------------------------------------

func TestGetPaymentsReturnsEnvelopeWithPagination(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments":
			capturedQuery = r.URL.Query()
			fmt.Fprint(w, `{"status":"success","data":{
				"limit":50,"offset":100,"count":1,
				"items":[
					{
						"uuid":"pay-001","status":"paid","status_ru":"Оплачен",
						"client_id":42,"account_id":7,"client_email":"u@example.com",
						"amount_minor":99000,"currency":"RUB",
						"items":[{"id":1,"title":"Product"}],
						"provider":"yookassa",
						"provider_invoice_id":"platega-tx-7788",
						"date_create":"2024-01-10T10:00:00Z",
						"date_paid":"2024-01-10T10:10:00Z",
						"access":{"account_id":7,"plans":["pro","ai"],"access_end":"2024-02-10T10:10:00Z"}
					}
				]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	uid := int64(42)
	res, err := client.GetPayments(context.Background(), &uid, 50, 100)
	if err != nil {
		t.Fatalf("GetPayments() error = %v", err)
	}

	if got := capturedQuery.Get("user_id"); got != "42" {
		t.Fatalf("user_id query = %q, want 42", got)
	}
	if got := capturedQuery.Get("limit"); got != "50" {
		t.Fatalf("limit query = %q, want 50", got)
	}
	if got := capturedQuery.Get("offset"); got != "100" {
		t.Fatalf("offset query = %q, want 100", got)
	}

	if res.Limit != 50 || res.Offset != 100 || res.Count != 1 {
		t.Fatalf("envelope = %+v, want limit=50 offset=100 count=1", res)
	}
	if len(res.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(res.Items))
	}
	p := res.Items[0]
	if p.UUID != "pay-001" || p.Status != "paid" || p.AmountMinor != 99000 {
		t.Fatalf("payment = %+v", p)
	}
	if p.ProviderInvoiceID == nil || *p.ProviderInvoiceID != "platega-tx-7788" {
		t.Fatalf("ProviderInvoiceID = %v, want platega-tx-7788", p.ProviderInvoiceID)
	}
	if p.ClientID == nil || *p.ClientID != 42 {
		t.Fatalf("ClientID = %v, want 42", p.ClientID)
	}
	if p.AccountID == nil || *p.AccountID != 7 {
		t.Fatalf("AccountID = %v, want 7", p.AccountID)
	}
	if p.Access == nil {
		t.Fatalf("Access = nil, want decoded access")
	}
	if p.Access.AccountID == nil || *p.Access.AccountID != 7 {
		t.Fatalf("Access.AccountID = %v, want 7", p.Access.AccountID)
	}
	if len(p.Access.Plans) != 2 || p.Access.Plans[0] != "pro" || p.Access.Plans[1] != "ai" {
		t.Fatalf("Access.Plans = %v, want [pro ai]", p.Access.Plans)
	}
	if p.Access.AccessEnd == nil || p.Access.AccessEnd.Year() != 2024 || p.Access.AccessEnd.Month() != 2 {
		t.Fatalf("Access.AccessEnd = %v, want 2024-02-10", p.Access.AccessEnd)
	}
	if len(p.Activation) != 0 {
		t.Fatalf("Activation = %+v, want empty", p.Activation)
	}
}

func TestGetPaymentsNullClientIDAndAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments":
			fmt.Fprint(w, `{"status":"success","data":{
				"limit":10,"offset":0,"count":1,
				"items":[
					{
						"uuid":"pay-002","status":"draft","status_ru":"Черновик",
						"client_id":null,"account_id":9,
						"amount_minor":1000,"currency":"RUB","items":[],
						"access":null
					}
				]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.GetPayments(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("GetPayments() error = %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(res.Items))
	}
	p := res.Items[0]
	if p.ClientID != nil {
		t.Fatalf("ClientID = %v, want nil (null in response)", *p.ClientID)
	}
	if p.AccountID == nil || *p.AccountID != 9 {
		t.Fatalf("AccountID = %v, want 9", p.AccountID)
	}
	if p.Access != nil {
		t.Fatalf("Access = %+v, want nil (null in response)", p.Access)
	}
}

// ----------------------------------------------------------------------------
// GetInvoiceInfo (GET /api/payments/invoice/{uuid})
// ----------------------------------------------------------------------------

func TestGetInvoiceInfoNullClientID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments/invoice/inv-001":
			fmt.Fprint(w, `{"status":"success","data":{
				"uuid":"inv-001","status":"invoiced","status_ru":"Выставлен",
				"client_id":null,"account_id":15,
				"amount_minor":5000,"currency":"RUB","description":"d","items":[],
				"provider":"platega","web_return_url":"https://example.com/return"
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.GetInvoiceInfo(context.Background(), "inv-001")
	if err != nil {
		t.Fatalf("GetInvoiceInfo() error = %v", err)
	}
	if res.ClientID != nil {
		t.Fatalf("ClientID = %v, want nil (null in response)", *res.ClientID)
	}
	if res.AccountID == nil || *res.AccountID != 15 {
		t.Fatalf("AccountID = %v, want 15", res.AccountID)
	}
	if res.WebReturnURL == nil || *res.WebReturnURL != "https://example.com/return" {
		t.Fatalf("WebReturnURL = %v, want https://example.com/return", res.WebReturnURL)
	}
}

func TestGetPaymentsWithoutUserIDOmitsParam(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments":
			capturedQuery = r.URL.Query()
			fmt.Fprint(w, `{"status":"success","data":{"limit":100000,"offset":0,"count":0,"items":[]}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.GetPayments(context.Background(), nil, 100_000, 0)
	if err != nil {
		t.Fatalf("GetPayments() error = %v", err)
	}
	if capturedQuery.Has("user_id") {
		t.Fatalf("user_id should not be sent when nil, got %q", capturedQuery.Get("user_id"))
	}
	if capturedQuery.Get("limit") != "100000" || capturedQuery.Get("offset") != "0" {
		t.Fatalf("limit/offset query = %v", capturedQuery)
	}
}

func TestGetPaymentsValidationFailsWithoutHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	uidNeg := int64(-1)
	cases := []struct {
		name   string
		userID *int64
		limit  int64
		offset int64
	}{
		{"limit zero", nil, 0, 0},
		{"limit negative", nil, -1, 0},
		{"offset negative", nil, 100, -1},
		{"user_id negative", &uidNeg, 100, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetPayments(context.Background(), tc.userID, tc.limit, tc.offset)
			var ve *ValidationError
			if !errorsAsValidation(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// GetMonthlySales (GET /api/payments/sales)
// ----------------------------------------------------------------------------

func TestGetMonthlySalesEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments/sales":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if got := r.URL.RawQuery; got != "" {
				t.Fatalf("query string = %q, want empty (no params)", got)
			}
			fmt.Fprint(w, `{"status":"success","data":{
				"month_start":"2026-04-01T00:00:00Z",
				"payments":[]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.GetMonthlySales(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetMonthlySales() error = %v", err)
	}
	if res.MonthStart == nil {
		t.Fatalf("MonthStart should be parsed")
	}
	if len(res.Payments) != 0 {
		t.Fatalf("Payments = %d, want 0", len(res.Payments))
	}
}

func TestGetMonthlySalesReturnsCategorizedPayments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments/sales":
			fmt.Fprint(w, `{"status":"success","data":{
				"month_start":"2026-04-01T00:00:00Z",
				"payments":[
					{
						"uuid":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"user_id":100,"staff_id":200,
						"amount_minor":600000,
						"category":"main","repeat_purchase":true,"first_ever_purchase":false,
						"date_paid":"2026-04-10T12:00:00Z"
					},
					{
						"uuid":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						"user_id":null,"staff_id":null,
						"amount_minor":100000,
						"category":"other","repeat_purchase":false,"first_ever_purchase":true,
						"date_paid":"2026-04-11T12:00:00Z"
					},
					{
						"uuid":"cccccccccccccccccccccccccccccccc",
						"user_id":102,"staff_id":201,
						"amount_minor":300000,
						"category":"extra","repeat_purchase":false,"first_ever_purchase":false,
						"date_paid":null
					}
				]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.GetMonthlySales(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetMonthlySales() error = %v", err)
	}
	if len(res.Payments) != 3 {
		t.Fatalf("len(Payments) = %d, want 3", len(res.Payments))
	}

	mainRepeat := res.Payments[0]
	if mainRepeat.Category != "main" {
		t.Fatalf("Category = %q, want main", mainRepeat.Category)
	}
	if !mainRepeat.RepeatPurchase {
		t.Fatalf("RepeatPurchase = false, want true")
	}
	// Returning client (repeat) → not their first-ever purchase.
	if mainRepeat.FirstEverPurchase {
		t.Fatalf("FirstEverPurchase = true, want false")
	}
	if mainRepeat.StaffID == nil || *mainRepeat.StaffID != 200 {
		t.Fatalf("StaffID = %v, want pointer to 200", mainRepeat.StaffID)
	}
	if mainRepeat.AmountMinor != 600000 {
		t.Fatalf("AmountMinor = %d, want 600000", mainRepeat.AmountMinor)
	}

	otherNew := res.Payments[1]
	if otherNew.Category != "other" || otherNew.RepeatPurchase {
		t.Fatalf("otherNew = %+v", otherNew)
	}
	// This client's first-ever paid purchase → first_ever_purchase true.
	if !otherNew.FirstEverPurchase {
		t.Fatalf("FirstEverPurchase = false, want true")
	}
	if otherNew.StaffID != nil {
		t.Fatalf("StaffID = %v, want nil (null in response)", otherNew.StaffID)
	}
	if otherNew.UserID != nil {
		t.Fatalf("UserID = %v, want nil (null in response)", *otherNew.UserID)
	}
	if mainRepeat.UserID == nil || *mainRepeat.UserID != 100 {
		t.Fatalf("UserID = %v, want pointer to 100", mainRepeat.UserID)
	}

	extraNoDate := res.Payments[2]
	if extraNoDate.Category != "extra" {
		t.Fatalf("Category = %q, want extra", extraNoDate.Category)
	}
	if extraNoDate.DatePaid != nil {
		t.Fatalf("DatePaid = %v, want nil (null in response)", extraNoDate.DatePaid)
	}
}

// LastPurchaseDates (POST /api/payments/last-purchase)
func TestLastPurchaseDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/payments/last-purchase":
			fmt.Fprint(w, `{"status":"success","data":{"100":"2026-06-01T12:00:00+03:00","101":null,"102":"2026-05-01T09:00:00Z"}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.LastPurchaseDates(context.Background(), []int64{100, 101, 102})
	if err != nil {
		t.Fatalf("LastPurchaseDates() error = %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("len(res) = %d, want 3 (every requested id present)", len(res))
	}
	if res[100] == nil || res[100].Year() != 2026 {
		t.Fatalf("res[100] = %v, want a 2026 date", res[100])
	}
	if res[101] != nil {
		t.Fatalf("res[101] = %v, want nil (null in response)", res[101])
	}
	if res[102] == nil {
		t.Fatalf("res[102] = nil, want a date")
	}
}
