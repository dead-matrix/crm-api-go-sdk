package crmapi

import (
	"context"
	"fmt"
)

// AIUsageFunctionSummary — итоги расхода AI по одной функции для клиента:
// текущий остаток (баланс/база/пакеты), расход (списанные токены + стоимость
// по ключу OpenRouter) и число генераций, плюс режим личного ключа.
type AIUsageFunctionSummary struct {
	BalanceTokens int64   `json:"balance_tokens"`
	BaseTokens    int64   `json:"base_tokens"`
	PackageTokens int64   `json:"package_tokens"`
	SpentTokens   int64   `json:"spent_tokens"`
	SpentUSD      float64 `json:"spent_usd"`
	RawTokens     int64   `json:"raw_tokens"`
	Generations   int64   `json:"generations"`
	KeyMode       string  `json:"key_mode,omitempty"`
}

// AIUsageBucket — расход за день (Date) или месяц (Month) по одной функции.
type AIUsageBucket struct {
	Date     string  `json:"date,omitempty"`
	Month    string  `json:"month,omitempty"`
	Function string  `json:"function"`
	Tokens   int64   `json:"tokens"`
	USD      float64 `json:"usd"`
	Count    int64   `json:"count"`
}

// AIUsageGeneration — одна генерация (списание) без текстов запроса/ответа.
type AIUsageGeneration struct {
	CreatedAt        string  `json:"created_at,omitempty"`
	Function         string  `json:"function"`
	Model            string  `json:"model,omitempty"`
	PromptTokens     *int64  `json:"prompt_tokens,omitempty"`
	CompletionTokens *int64  `json:"completion_tokens,omitempty"`
	TotalTokens      *int64  `json:"total_tokens,omitempty"`
	CostUSD          float64 `json:"cost_usd"`
	BilledTokens     int64   `json:"billed_tokens"`
}

// AIUsageResult — отчёт о расходе AI-кредитов клиента (леджер).
type AIUsageResult struct {
	BotID      int64                             `json:"bot_id"`
	ByFunction map[string]AIUsageFunctionSummary `json:"by_function"`
	Daily      []AIUsageBucket                   `json:"daily"`
	Monthly    []AIUsageBucket                   `json:"monthly"`
	Recent     []AIUsageGeneration               `json:"recent"`
}

// AIUsage возвращает отчёт о расходе AI-кредитов клиента: итоги/остаток по
// функциям, разбивку списаний по дням и месяцам и последние генерации. Нужен
// для разбора жалоб «слишком быстро ушли токены».
//
// GET /api/users/{user_id}/ai-usage?bot_id=1
func (c *Client) AIUsage(ctx context.Context, userID int64) (*AIUsageResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var res AIUsageResult
	if err := c.get(
		ctx,
		fmt.Sprintf("/api/users/%d/ai-usage", userID),
		map[string]string{"bot_id": "1"},
		true,
		&res,
	); err != nil {
		return nil, err
	}
	return &res, nil
}
