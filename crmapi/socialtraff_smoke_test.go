package crmapi

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestSocialTraffSmoke проверяет на живом CRM SocialTraff, что ответы
// subscription-state, определений доступа и истории платежей декодируются
// текущими моделями SDK (nullable client_id, account_id, access).
func TestSocialTraffSmoke(t *testing.T) {
	if strings.TrimSpace(os.Getenv("RUN_REAL_API_TESTS")) != "1" {
		t.Skip("set RUN_REAL_API_TESTS=1 to run real API smoke tests")
	}

	loadDotEnvForTest(t)

	baseURL := requiredEnv(t, "CRM_API_BASE_URL")
	staffID := requiredEnvInt64(t, "CRM_API_STAFF_ID")
	serviceToken := requiredEnv(t, "CRM_API_SERVICE_TOKEN")
	userID := requiredEnvInt64(t, "CRM_API_TEST_USER_ID")

	client, err := NewClient(Config{
		BaseURL:        baseURL,
		StaffID:        staffID,
		ServiceToken:   serviceToken,
		Timeout:        20 * time.Second,
		RequestRetries: 3,
	})
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("issueJWT", func(t *testing.T) {
		if _, _, err := client.issueJWT(ctx); err != nil {
			t.Fatalf("issueJWT() error: %v", err)
		}
	})

	t.Run("subscriptionState", func(t *testing.T) {
		items, err := client.SubscriptionState(ctx, []int64{userID})
		if err != nil {
			t.Fatalf("SubscriptionState() error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("SubscriptionState() returned %d items, want 1", len(items))
		}
	})

	t.Run("accessDefinitions", func(t *testing.T) {
		defs, err := client.AccessDefinitions(ctx)
		if err != nil {
			t.Fatalf("AccessDefinitions() error: %v", err)
		}
		if defs == nil {
			t.Fatalf("AccessDefinitions() returned nil result")
		}
	})

	t.Run("getPayments", func(t *testing.T) {
		res, err := client.GetPayments(ctx, nil, 5, 0)
		if err != nil {
			t.Fatalf("GetPayments() error: %v", err)
		}
		if res == nil {
			t.Fatalf("GetPayments() returned nil result")
		}
	})
}
