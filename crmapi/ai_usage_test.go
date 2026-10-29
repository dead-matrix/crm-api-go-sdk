package crmapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

const aiUsageFullReport = `{
	"bot_id":10,"account_id":5,
	"by_function":{
		"comment":{"unlimited":true,"balance_tokens":1500,"base_tokens":1000,"package_tokens":500,
			"spent_tokens":300,"spent_usd":0.0123,"raw_tokens":280,"generations":4,
			"admin_spent_tokens":70,"admin_spent_usd":0.0031,"admin_generations":2,"key_mode":"service"},
		"chatting":{"unlimited":false,"balance_tokens":0,"base_tokens":0,"package_tokens":0,
			"spent_tokens":0,"spent_usd":0.0,"raw_tokens":0,"generations":0,
			"admin_spent_tokens":0,"admin_spent_usd":0.0,"admin_generations":0,"key_mode":"service"}
	},
	"daily":[{"date":"2026-10-09","function":"comment","tokens":300,"usd":0.0123,"count":4}],
	"monthly":[{"month":"2026-10","function":"comment","tokens":300,"usd":0.0123,"count":4}],
	"recent":[
		{"created_at":"2026-10-09T12:00:00","function":"comment","model":"gpt","prompt_tokens":10,
		 "completion_tokens":20,"total_tokens":30,"cost_usd":0.001,"billed_tokens":40,"source":"admin"},
		{"created_at":"2026-10-09T11:00:00","function":"comment","model":null,"prompt_tokens":null,
		 "completion_tokens":null,"total_tokens":null,"cost_usd":0.002,"billed_tokens":50,"source":null}
	],
	"key_stats":{"mask":"sk-or-…abcd","usage_usd":1.25,"limit_usd":10.0,"limit_remaining_usd":8.75,"disabled":false}
}`

// Главное в правке: bot_id больше не подставляется. Раньше уходил bot_id=1, а
// у CRM SocialTraff единственный бот 10, и свой бот она выбирает сама.
func TestAIUsage_DoesNotSendBotID(t *testing.T) {
	var query url.Values
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42/ai-usage")
		query = r.URL.Query()
		writeSuccess(w, aiUsageFullReport)
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.AIUsage(context.Background(), 42)
	if err != nil {
		t.Fatalf("AIUsage error: %v", err)
	}
	if len(query) != 0 {
		t.Fatalf("query = %v, want no parameters (server defaults)", query)
	}

	if res.BotID != 10 || res.AccountID != 5 {
		t.Fatalf("bot/account = %d/%d, want 10/5", res.BotID, res.AccountID)
	}
	comment := res.ByFunction["comment"]
	if !comment.Unlimited || comment.BalanceTokens != 1500 || comment.SpentTokens != 300 || comment.Generations != 4 {
		t.Fatalf("comment = %+v", comment)
	}
	if comment.AdminSpentTokens != 70 || comment.AdminSpentUSD != 0.0031 || comment.AdminGenerations != 2 {
		t.Fatalf("comment admin = %+v", comment)
	}
	if res.ByFunction["chatting"].Unlimited {
		t.Fatalf("chatting.Unlimited = true, want false")
	}
	if len(res.Daily) != 1 || res.Daily[0].Date != "2026-10-09" || len(res.Monthly) != 1 || res.Monthly[0].Month != "2026-10" {
		t.Fatalf("daily/monthly = %+v / %+v", res.Daily, res.Monthly)
	}

	if len(res.Recent) != 2 {
		t.Fatalf("len(Recent) = %d, want 2", len(res.Recent))
	}
	if src := res.Recent[0].Source; src == nil || *src != "admin" {
		t.Fatalf("Recent[0].Source = %v, want admin", src)
	}
	if res.Recent[1].Source != nil || res.Recent[1].PromptTokens != nil || res.Recent[1].Model != "" {
		t.Fatalf("Recent[1] nullable fields = %+v", res.Recent[1])
	}

	ks := res.KeyStats
	if ks == nil || ks.Mask != "sk-or-…abcd" || ks.UsageUSD != 1.25 || ks.Disabled {
		t.Fatalf("KeyStats = %+v", ks)
	}
	if ks.LimitUSD == nil || *ks.LimitUSD != 10 || ks.LimitRemainingUSD == nil || *ks.LimitRemainingUSD != 8.75 {
		t.Fatalf("KeyStats limits = %v / %v", ks.LimitUSD, ks.LimitRemainingUSD)
	}
}

func TestAIUsageWithOptions_SendsOnlySetFields(t *testing.T) {
	cases := []struct {
		name string
		opts AIUsageOptions
		want map[string]string
	}{
		{"all", AIUsageOptions{BotID: 10, Days: 90, Months: 6, Recent: 250},
			map[string]string{"bot_id": "10", "days": "90", "months": "6", "recent": "250"}},
		{"days only", AIUsageOptions{Days: 7}, map[string]string{"days": "7"}},
		{"bot only", AIUsageOptions{BotID: 3}, map[string]string{"bot_id": "3"}},
		{"zero value", AIUsageOptions{}, map[string]string{}},
	}

	var query url.Values
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		wantRequest(t, r, http.MethodGet, "/api/users/42/ai-usage")
		query = r.URL.Query()
		writeSuccess(w, `{"bot_id":10,"account_id":5,"by_function":{},"daily":[],"monthly":[],"recent":[],"key_stats":null}`)
	})
	client := mustNewClient(t, server.URL, server.Client())

	for _, tc := range cases {
		res, err := client.AIUsageWithOptions(context.Background(), 42, tc.opts)
		if err != nil {
			t.Fatalf("%s: error %v", tc.name, err)
		}
		if len(query) != len(tc.want) {
			t.Fatalf("%s: query = %v, want %v", tc.name, query, tc.want)
		}
		for key, value := range tc.want {
			if query.Get(key) != value {
				t.Fatalf("%s: query[%s] = %q, want %q", tc.name, key, query.Get(key), value)
			}
		}
		if res.KeyStats != nil {
			t.Fatalf("%s: KeyStats = %+v, want nil for null", tc.name, res.KeyStats)
		}
	}
}

// У ключа без лимита OpenRouter отдаёт null: nil нельзя превращать в 0, это
// читалось бы как "лимит исчерпан". CRM без новых ключей тоже не ошибка.
func TestAIUsage_KeyWithoutLimitAndLegacyReport(t *testing.T) {
	payload := `{"bot_id":10,"account_id":5,"by_function":{},"daily":[],"monthly":[],"recent":[],
		"key_stats":{"mask":"sk-or-…wxyz","usage_usd":0.0,"limit_usd":null,"limit_remaining_usd":null,"disabled":true}}`
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, payload)
	})
	client := mustNewClient(t, server.URL, server.Client())

	res, err := client.AIUsage(context.Background(), 42)
	if err != nil {
		t.Fatalf("AIUsage error: %v", err)
	}
	if ks := res.KeyStats; ks == nil || ks.LimitUSD != nil || ks.LimitRemainingUSD != nil || !ks.Disabled {
		t.Fatalf("KeyStats = %+v, want nil limits and disabled", res.KeyStats)
	}

	payload = `{"bot_id":1,
		"by_function":{"comment":{"balance_tokens":5,"base_tokens":5,"package_tokens":0,"spent_tokens":1,"spent_usd":0.1,"raw_tokens":1,"generations":1}},
		"daily":[],"monthly":[],
		"recent":[{"created_at":"2026-01-01T10:00:00","function":"comment","model":"m","cost_usd":0.1,"billed_tokens":1}]}`
	res, err = client.AIUsage(context.Background(), 42)
	if err != nil {
		t.Fatalf("AIUsage legacy error: %v", err)
	}
	if res.AccountID != 0 || res.KeyStats != nil || res.Recent[0].Source != nil {
		t.Fatalf("legacy report = %+v", res)
	}
	if fn := res.ByFunction["comment"]; fn.Unlimited || fn.AdminGenerations != 0 || fn.BalanceTokens != 5 {
		t.Fatalf("legacy comment = %+v", fn)
	}
}

func TestAIUsage_ValidationWithoutHTTP(t *testing.T) {
	server := newCRMTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	})
	client := mustNewClient(t, server.URL, server.Client())
	ctx := context.Background()

	calls := map[string]func() error{
		"user":   func() error { _, err := client.AIUsage(ctx, 0); return err },
		"bot":    func() error { _, err := client.AIUsageWithOptions(ctx, 42, AIUsageOptions{BotID: -1}); return err },
		"days":   func() error { _, err := client.AIUsageWithOptions(ctx, 42, AIUsageOptions{Days: -1}); return err },
		"months": func() error { _, err := client.AIUsageWithOptions(ctx, 42, AIUsageOptions{Months: -1}); return err },
		"recent": func() error { _, err := client.AIUsageWithOptions(ctx, 42, AIUsageOptions{Recent: -1}); return err },
	}
	for name, call := range calls {
		var vErr *ValidationError
		if err := call(); !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}
