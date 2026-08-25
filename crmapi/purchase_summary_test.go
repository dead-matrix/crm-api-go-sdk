package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestClient_PurchaseSummaries: агрегат оплат декодируется, отсутствующий в
// ответе id всё равно присутствует в карте как «не покупал».
func TestClient_PurchaseSummaries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/payments/purchase-summary":
			_, _ = w.Write([]byte(`{"status":"success","data":{
				"100":{"first_paid_at":"2026-05-01T09:00:00+03:00","last_paid_at":"2026-06-01T12:00:00+03:00","payments_count":2,"total_minor":500000},
				"101":{"first_paid_at":null,"last_paid_at":null,"payments_count":0,"total_minor":0}
			}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	// 102 сервер не вернул - защитный проход обязан добавить его как пустой.
	res, err := client.PurchaseSummaries(context.Background(), []int64{100, 101, 102})
	if err != nil {
		t.Fatalf("PurchaseSummaries error: %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("len = %d, want 3", len(res))
	}

	buyer := res[100]
	if !buyer.EverPurchased() || buyer.PaymentsCount != 2 || buyer.TotalMinor != 500000 {
		t.Fatalf("buyer = %+v, want count=2 total=500000", buyer)
	}
	if buyer.FirstPaidAt == nil || buyer.LastPaidAt == nil {
		t.Fatalf("buyer dates not parsed: %+v", buyer)
	}
	// Первая оплата строго раньше последней - именно её берёт обрезка диалога.
	if !buyer.FirstPaidAt.Before(*buyer.LastPaidAt) {
		t.Fatalf("first %v must precede last %v", buyer.FirstPaidAt, buyer.LastPaidAt)
	}

	for _, id := range []int64{101, 102} {
		if res[id].EverPurchased() || res[id].FirstPaidAt != nil {
			t.Fatalf("id %d must be a non-buyer, got %+v", id, res[id])
		}
	}
}

func TestClient_PurchaseSummaries_EmptyInput(t *testing.T) {
	client := mustNewClient(t, "https://example.test", &http.Client{})
	res, err := client.PurchaseSummaries(context.Background(), nil)
	if err != nil || len(res) != 0 {
		t.Fatalf("empty input: res=%v err=%v", res, err)
	}
}
