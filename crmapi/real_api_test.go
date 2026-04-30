package crmapi

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRealAPI_Smoke(t *testing.T) {
	if strings.TrimSpace(os.Getenv("RUN_REAL_API_TESTS")) != "1" {
		t.Skip("set RUN_REAL_API_TESTS=1 to run real API smoke tests")
	}

	loadDotEnvForTest(t)

	baseURL := requiredEnv(t, "CRM_API_BASE_URL")
	staffID := requiredEnvInt64(t, "CRM_API_STAFF_ID")
	serviceToken := requiredEnv(t, "CRM_API_SERVICE_TOKEN")
	userID := requiredEnvInt64(t, "CRM_API_TEST_USER_ID")
	botID := requiredEnvInt64(t, "CRM_API_TEST_BOT_ID")
	taskID := requiredEnvInt64(t, "CRM_API_TEST_TASK_ID")
	taskType := requiredEnv(t, "CRM_API_TEST_TASK_TYPE")
	paymentUUID := requiredEnv(t, "CRM_API_TEST_PAYMENT_UUID")

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

	t.Run("getUser", func(t *testing.T) {
		user, err := client.GetUser(ctx, userID)
		if err != nil {
			t.Fatalf("GetUser() error: %v", err)
		}
		if user == nil || user.UserID != userID {
			t.Fatalf("GetUser() returned unexpected user payload")
		}
	})

	t.Run("accountsList", func(t *testing.T) {
		accounts, err := client.AccountsList(ctx, userID)
		if err != nil {
			t.Fatalf("AccountsList() error: %v", err)
		}
		if accounts == nil {
			t.Fatalf("AccountsList() returned nil slice")
		}
	})

	t.Run("tasksLog", func(t *testing.T) {
		logRes, err := client.TasksLog(ctx, userID, taskType, taskID, botID)
		if err != nil {
			t.Fatalf("TasksLog() error: %v", err)
		}
		if logRes == nil {
			t.Fatalf("TasksLog() returned nil result")
		}
		if logRes.Content == nil {
			t.Fatalf("TasksLog() returned nil content")
		}
	})

	t.Run("getStaff", func(t *testing.T) {
		staff, err := client.GetStaff(ctx)
		if err != nil {
			t.Fatalf("GetStaff() error: %v", err)
		}
		if staff == nil {
			t.Fatalf("GetStaff() returned nil result")
		}
	})

	t.Run("productsActive", func(t *testing.T) {
		products, err := client.ProductsActive(ctx)
		if err != nil {
			t.Fatalf("ProductsActive() error: %v", err)
		}
		if products == nil {
			t.Fatalf("ProductsActive() returned nil map")
		}
	})

	t.Run("profileStatistics", func(t *testing.T) {
		stats, err := client.ProfileStatistics(ctx, userID)
		if err != nil {
			t.Fatalf("ProfileStatistics() error: %v", err)
		}
		if stats == nil {
			t.Fatalf("ProfileStatistics() returned nil result")
		}
	})

	t.Run("promptGet", func(t *testing.T) {
		if _, err := client.PromptGet(ctx, userID); err != nil {
			t.Fatalf("PromptGet() error: %v", err)
		}
	})

	t.Run("referralsInfo", func(t *testing.T) {
		info, err := client.ReferralsInfo(ctx, userID)
		if err != nil {
			t.Fatalf("ReferralsInfo() error: %v", err)
		}
		if info == nil {
			t.Fatalf("ReferralsInfo() returned nil result")
		}
	})

	t.Run("getPayments", func(t *testing.T) {
		payments, err := client.GetPayments(ctx, &userID, 100, 0)
		if err != nil {
			t.Fatalf("GetPayments() error: %v", err)
		}
		if payments == nil {
			t.Fatalf("GetPayments() returned nil result")
		}
	})

	t.Run("getInvoiceInfo", func(t *testing.T) {
		invoice, err := client.GetInvoiceInfo(ctx, paymentUUID)
		if err != nil {
			t.Fatalf("GetInvoiceInfo() error: %v", err)
		}
		if invoice == nil || invoice.UUID == "" {
			t.Fatalf("GetInvoiceInfo() returned unexpected payload")
		}
	})
}

func loadDotEnvForTest(t *testing.T) {
	t.Helper()
	for _, name := range []string{".env", filepath.Join("..", ".env")} {
		f, err := os.Open(name)
		if err != nil {
			continue
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, value)
			}
		}
		return
	}
}

func requiredEnv(t *testing.T, key string) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		t.Fatalf("required env %s is empty", key)
	}
	return v
}

func requiredEnvInt64(t *testing.T, key string) int64 {
	t.Helper()
	v := requiredEnv(t, key)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		t.Fatalf("env %s must be int64: %v", key, err)
	}
	return n
}
