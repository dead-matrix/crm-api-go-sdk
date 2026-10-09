package crmapi

import (
	"context"
	"fmt"
	"strconv"
)

// AIUsageFunctionSummary — итоги расхода AI по одной функции для клиента:
// текущий остаток (баланс/база/пакеты), расход (списанные токены + стоимость
// по ключу OpenRouter) и число генераций, плюс режим личного ключа.
type AIUsageFunctionSummary struct {
	// Unlimited - у функции безлимит: остаток ниже настоящий, но расход им не
	// ограничен.
	Unlimited     bool    `json:"unlimited"`
	BalanceTokens int64   `json:"balance_tokens"`
	BaseTokens    int64   `json:"base_tokens"`
	PackageTokens int64   `json:"package_tokens"`
	SpentTokens   int64   `json:"spent_tokens"`
	SpentUSD      float64 `json:"spent_usd"`
	RawTokens     int64   `json:"raw_tokens"`
	Generations   int64   `json:"generations"`
	// Admin* - траты администратора основного бота. Остаток клиента они не
	// уменьшают, поэтому идут отдельно и в SpentTokens/SpentUSD/Generations не
	// входят.
	AdminSpentTokens int64   `json:"admin_spent_tokens"`
	AdminSpentUSD    float64 `json:"admin_spent_usd"`
	AdminGenerations int64   `json:"admin_generations"`
	KeyMode          string  `json:"key_mode,omitempty"`
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
	// Source - чья это трата: "service" (клиент) или "admin" (администратор
	// основного бота, остаток клиента не уменьшает). nil у CRM без этого ключа.
	Source *string `json:"source,omitempty"`
}

// AIKeyStats - ключ OpenRouter аккаунта по данным самого провайдера. Расход
// здесь считается всегда, даже когда лимиты в CRM выключены и леджер трат не
// ведёт, поэтому цифры могут не совпадать с ByFunction. Сам ключ не
// передаётся, только маска.
//
// LimitUSD и LimitRemainingUSD равны nil у ключа без лимита: ноль означал бы
// "лимит исчерпан".
type AIKeyStats struct {
	Mask              string   `json:"mask"`
	UsageUSD          float64  `json:"usage_usd"`
	LimitUSD          *float64 `json:"limit_usd,omitempty"`
	LimitRemainingUSD *float64 `json:"limit_remaining_usd,omitempty"`
	Disabled          bool     `json:"disabled"`
}

// AIUsageResult — отчёт о расходе AI-кредитов клиента (леджер).
type AIUsageResult struct {
	BotID int64 `json:"bot_id"`
	// AccountID - аккаунт, чей расход показан: CRM SocialTraff ведёт AI-баланс
	// по основному аккаунту человека. Ноль у CRM без этого ключа.
	AccountID  int64                             `json:"account_id"`
	ByFunction map[string]AIUsageFunctionSummary `json:"by_function"`
	Daily      []AIUsageBucket                   `json:"daily"`
	Monthly    []AIUsageBucket                   `json:"monthly"`
	Recent     []AIUsageGeneration               `json:"recent"`
	// KeyStats - статистика ключа аккаунта. nil, если ключа ещё нет, провайдер
	// не ответил или CRM этот ключ не присылает: отчёт по леджеру при этом цел.
	KeyStats *AIKeyStats `json:"key_stats,omitempty"`
}

// AIUsageOptions - необязательные параметры AIUsageWithOptions. Нулевое поле
// в запрос не попадает, и CRM берёт своё значение по умолчанию: свой бот,
// 30 дней, 12 месяцев, 100 последних генераций. Границы (дни до 180, месяцы до
// 24, генерации до 500) проверяет CRM и отвечает ValidationError.
type AIUsageOptions struct {
	BotID  int64
	Days   int
	Months int
	Recent int
}

// AIUsage возвращает отчёт о расходе AI-кредитов клиента: итоги/остаток по
// функциям, разбивку списаний по дням и месяцам и последние генерации. Нужен
// для разбора жалоб «слишком быстро ушли токены».
//
// GET /api/users/{user_id}/ai-usage
//
// bot_id не передаётся: бота выбирает CRM. Раньше метод всегда слал bot_id=1,
// и CRM с другим ботом считала отчёт не по тому боту. Конкретный бот и глубина
// отчёта задаются через AIUsageWithOptions.
func (c *Client) AIUsage(ctx context.Context, userID int64) (*AIUsageResult, error) {
	return c.AIUsageWithOptions(ctx, userID, AIUsageOptions{})
}

// AIUsageWithOptions - тот же отчёт, что AIUsage, с явным ботом и глубиной:
// opts.Days дней в Daily, opts.Months месяцев в Monthly, opts.Recent последних
// генераций в Recent.
func (c *Client) AIUsageWithOptions(ctx context.Context, userID int64, opts AIUsageOptions) (*AIUsageResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if opts.BotID < 0 || opts.Days < 0 || opts.Months < 0 || opts.Recent < 0 {
		return nil, &ValidationError{Message: "bot_id, days, months and recent must be non-negative"}
	}

	query := map[string]string{}
	if opts.BotID > 0 {
		query["bot_id"] = strconv.FormatInt(opts.BotID, 10)
	}
	if opts.Days > 0 {
		query["days"] = strconv.Itoa(opts.Days)
	}
	if opts.Months > 0 {
		query["months"] = strconv.Itoa(opts.Months)
	}
	if opts.Recent > 0 {
		query["recent"] = strconv.Itoa(opts.Recent)
	}

	var res AIUsageResult
	if err := c.get(ctx, fmt.Sprintf("/api/users/%d/ai-usage", userID), query, true, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
