package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	res, err := client.GetMonthlySales(context.Background())
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
						"category":"main","repeat_purchase":true,
						"date_paid":"2026-04-10T12:00:00Z"
					},
					{
						"uuid":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						"user_id":101,"staff_id":null,
						"amount_minor":100000,
						"category":"other","repeat_purchase":false,
						"date_paid":"2026-04-11T12:00:00Z"
					},
					{
						"uuid":"cccccccccccccccccccccccccccccccc",
						"user_id":102,"staff_id":201,
						"amount_minor":300000,
						"category":"extra","repeat_purchase":false,
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
	res, err := client.GetMonthlySales(context.Background())
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
	if otherNew.StaffID != nil {
		t.Fatalf("StaffID = %v, want nil (null in response)", otherNew.StaffID)
	}

	extraNoDate := res.Payments[2]
	if extraNoDate.Category != "extra" {
		t.Fatalf("Category = %q, want extra", extraNoDate.Category)
	}
	if extraNoDate.DatePaid != nil {
		t.Fatalf("DatePaid = %v, want nil (null in response)", extraNoDate.DatePaid)
	}
}
