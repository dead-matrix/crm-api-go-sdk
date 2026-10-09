package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// CreateUser creates a user via POST /api/users (idempotent).
//
// If a registration for (UserID, BotID) already exists on the server, the
// response carries Created=false plus the existing record — no side effects.
// Otherwise Created=true and the fields describe the freshly created record.
func (c *Client) CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		Created  bool    `json:"created"`
		UserID   int64   `json:"user_id"`
		FullName *string `json:"full_name"`
		Username *string `json:"username"`
		BotID    int64   `json:"bot_id"`
		Refer    *string `json:"refer"`
		DateReg  *string `json:"date_reg"`
	}

	if err := c.post(ctx, "/api/users", nil, true, input, &raw); err != nil {
		return nil, err
	}

	var dateReg *time.Time
	if raw.DateReg != nil {
		dateReg = utils.ParseTime(*raw.DateReg)
	}

	return &CreateUserResult{
		Created:  raw.Created,
		UserID:   raw.UserID,
		FullName: raw.FullName,
		Username: raw.Username,
		BotID:    raw.BotID,
		Refer:    raw.Refer,
		DateReg:  dateReg,
	}, nil
}

// ListUsers fetches a paginated list of users registered for the given bot.
//
// botID is required (>0). limit must be >0 (CRM allows up to 1_000_000;
// the recommended default mirrors the Python SDK at 100_000). offset must be >=0.
//
// Each ListUserItem includes a Restricted flag — clients that consume CRM
// messages should skip restricted users.
func (c *Client) ListUsers(ctx context.Context, botID int64, limit int64, offset int64) (*ListUsersResult, error) {
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if limit <= 0 {
		return nil, &ValidationError{Message: "limit must be a positive integer"}
	}
	if offset < 0 {
		return nil, &ValidationError{Message: "offset must be non-negative"}
	}

	query := map[string]string{
		"bot_id": fmt.Sprintf("%d", botID),
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}

	var raw struct {
		BotID  int64 `json:"bot_id"`
		Limit  int64 `json:"limit"`
		Offset int64 `json:"offset"`
		Count  int64 `json:"count"`
		Items  []struct {
			UserID     int64   `json:"user_id"`
			FullName   *string `json:"full_name"`
			Username   *string `json:"username"`
			DateReg    *string `json:"date_reg"`
			Refer      *string `json:"refer"`
			Restricted bool    `json:"restricted"`
			HasActive  bool    `json:"has_active_subscription"`
			Frozen     bool    `json:"frozen"`
		} `json:"items"`
	}

	if err := c.get(ctx, "/api/users", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]ListUserItem, 0, len(raw.Items))
	for _, it := range raw.Items {
		var dateReg *time.Time
		if it.DateReg != nil {
			dateReg = utils.ParseTime(*it.DateReg)
		}
		items = append(items, ListUserItem{
			UserID:     it.UserID,
			FullName:   it.FullName,
			Username:   it.Username,
			DateReg:    dateReg,
			Refer:      it.Refer,
			Restricted: it.Restricted,

			HasActiveSubscription: it.HasActive,
			Frozen:                it.Frozen,
		})
	}

	return &ListUsersResult{
		BotID:  raw.BotID,
		Limit:  raw.Limit,
		Offset: raw.Offset,
		Count:  raw.Count,
		Items:  items,
	}, nil
}

func (c *Client) UpdateUser(ctx context.Context, userID int64, input UpdateUserInput) (*UpdateUserResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		UserID   int64   `json:"user_id"`
		FullName string  `json:"full_name"`
		Username *string `json:"username"`
	}

	if err := c.put(ctx, fmt.Sprintf("/api/users/%d", userID), nil, true, input, &raw); err != nil {
		return nil, err
	}

	return &UpdateUserResult{
		UserID:   raw.UserID,
		FullName: raw.FullName,
		Username: raw.Username,
	}, nil
}

// GetUser возвращает карточку человека по Telegram id (GET /api/users/{user_id}).
//
// Кроме сведений о ботах (BotsInfo) карточка CRM SocialTraff несёт покупателя
// (Buyer) и все аккаунты человека с их доступом (Accounts). Подробности одного
// аккаунта - участники, чаты, боты, промокоды - отдаёт GetUserAccount.
func (c *Client) GetUser(ctx context.Context, userID int64) (*GetUserResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var raw struct {
		UserID    int64   `json:"user_id"`
		FullName  *string `json:"full_name"`
		Username  *string `json:"username"`
		Status    *string `json:"status"`
		HasActive bool    `json:"has_active_subscription"`
		Frozen    bool    `json:"frozen"`
		Buyer     *struct {
			BuyerID     int64   `json:"buyer_id"`
			Email       *string `json:"email"`
			DisplayName *string `json:"display_name"`
			RefCode     *string `json:"ref_code"`
			Refer       *string `json:"refer"`
			Landing     *string `json:"landing"`
			CreatedAt   *string `json:"created_at"`
		} `json:"buyer"`
		Accounts []struct {
			AccountID    int64             `json:"account_id"`
			Title        string            `json:"title"`
			Role         string            `json:"role"`
			Access       any               `json:"access"`
			AccessExpiry map[string]string `json:"access_expiry"`
			AccessEnd    *string           `json:"access_end"`
			IsPersonal   bool              `json:"is_personal"`
			IsPrimary    bool              `json:"is_primary"`
			OwnerBuyerID int64             `json:"owner_buyer_id"`
			CreatedAt    *string           `json:"created_at"`
			JoinedAt     *string           `json:"joined_at"`
			MembersCount int64             `json:"members_count"`
			ChatsCount   int64             `json:"chats_count"`
			BotsCount    int64             `json:"bots_count"`
		} `json:"accounts"`
		BotsInfo []struct {
			BotID        int64             `json:"bot_id"`
			BotName      string            `json:"bot_name"`
			AccountID    *int64            `json:"account_id"`
			Registered   *string           `json:"registered"`
			Refer        *string           `json:"refer"`
			Access       any               `json:"access"`
			AccessEnd    *string           `json:"access_end"`
			Frozen       *bool             `json:"frozen"`
			FrozenAt     *string           `json:"frozen_at"`
			FrozenExpiry map[string]string `json:"frozen_expiry"`
			AccessExpiry map[string]string `json:"access_expiry"`
		} `json:"bots_info"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/users/%d", userID), nil, true, &raw); err != nil {
		return nil, err
	}

	bots := make([]UserBotInfo, 0, len(raw.BotsInfo))
	for _, item := range raw.BotsInfo {
		var registered *time.Time
		var accessEnd *time.Time
		var frozenAt *time.Time

		if item.Registered != nil {
			registered = utils.ParseTime(*item.Registered)
		}
		if item.AccessEnd != nil {
			accessEnd = utils.ParseTime(*item.AccessEnd)
		}
		// frozen_at — наивный ISO без зоны (как registered/access_end), поэтому
		// тоже через utils.ParseTime, а не прямой *time.Time-анмаршал.
		if item.FrozenAt != nil {
			frozenAt = utils.ParseTime(*item.FrozenAt)
		}
		frozen := false
		if item.Frozen != nil {
			frozen = *item.Frozen
		}

		bots = append(bots, UserBotInfo{
			BotID:        item.BotID,
			BotName:      item.BotName,
			AccountID:    item.AccountID,
			Registered:   registered,
			Refer:        item.Refer,
			Access:       item.Access,
			AccessEnd:    accessEnd,
			Frozen:       frozen,
			FrozenAt:     frozenAt,
			FrozenExpiry: item.FrozenExpiry,
			AccessExpiry: item.AccessExpiry,
		})
	}

	var buyer *UserBuyer
	if b := raw.Buyer; b != nil {
		buyer = &UserBuyer{
			BuyerID:     b.BuyerID,
			Email:       b.Email,
			DisplayName: b.DisplayName,
			RefCode:     b.RefCode,
			Refer:       b.Refer,
			Landing:     b.Landing,
			CreatedAt:   parseOptionalTime(b.CreatedAt),
		}
	}

	accounts := make([]UserAccount, 0, len(raw.Accounts))
	for _, a := range raw.Accounts {
		accounts = append(accounts, UserAccount{
			AccountID:    a.AccountID,
			Title:        a.Title,
			Role:         a.Role,
			Access:       accessFlags(a.Access),
			AccessExpiry: a.AccessExpiry,
			AccessEnd:    parseOptionalTime(a.AccessEnd),
			IsPersonal:   a.IsPersonal,
			IsPrimary:    a.IsPrimary,
			OwnerBuyerID: a.OwnerBuyerID,
			CreatedAt:    parseOptionalTime(a.CreatedAt),
			JoinedAt:     parseOptionalTime(a.JoinedAt),
			MembersCount: a.MembersCount,
			ChatsCount:   a.ChatsCount,
			BotsCount:    a.BotsCount,
		})
	}

	return &GetUserResult{
		UserID:   raw.UserID,
		FullName: raw.FullName,
		Username: raw.Username,
		Status:   raw.Status,

		HasActiveSubscription: raw.HasActive,
		Frozen:                raw.Frozen,
		BotsInfo:              bots,
		Buyer:                 buyer,
		Accounts:              accounts,
	}, nil
}

// accessFlags приводит access аккаунта к набору фич {ключ: живая ли}. Штатно
// CRM отдаёт {ключ: true}, но у строк доступа без карты сроков уходит то, что
// лежит в базе, и значение там бывает не булевым. CRM считает ключ живым по
// истинности значения; здесь то же правило, иначе одна такая строка роняла бы
// разбор всей карточки. Не объект (в том числе null) даёт nil.
func accessFlags(raw any) map[string]bool {
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]bool, len(obj))
	for key, value := range obj {
		out[key] = jsonTruthy(value)
	}
	return out
}

// jsonTruthy повторяет истинность значения в Python для разобранного JSON:
// CRM написана на нём и именно так решает, включена ли фича.
func jsonTruthy(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return v != ""
	case map[string]any:
		return len(v) > 0
	case []any:
		return len(v) > 0
	default:
		return true
	}
}

// GetUserAccount возвращает полную карточку одного аккаунта человека:
// участников, подключённые чаты и ботов, ожидающие промокоды
// (GET /api/users/{user_id}/accounts/{account_id}).
//
// Список аккаунтов человека берётся из GetUser(...).Accounts. Если человек не
// является действующим участником аккаунта, CRM отвечает 404 и метод
// возвращает *APIError с Code == "not_found".
func (c *Client) GetUserAccount(ctx context.Context, userID, accountID int64) (*UserAccountCard, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if accountID <= 0 {
		return nil, &ValidationError{Message: "account_id must be a positive integer"}
	}

	var raw struct {
		AccountID    int64             `json:"account_id"`
		Title        string            `json:"title"`
		IsPersonal   bool              `json:"is_personal"`
		IsPrimary    bool              `json:"is_primary"`
		CreatedAt    *string           `json:"created_at"`
		OwnerBuyerID int64             `json:"owner_buyer_id"`
		Role         string            `json:"role"`
		Access       any               `json:"access"`
		AccessEnd    *string           `json:"access_end"`
		AccessExpiry map[string]string `json:"access_expiry"`
		Members      []struct {
			BuyerID     int64   `json:"buyer_id"`
			TgID        *int64  `json:"tg_id"`
			Email       *string `json:"email"`
			DisplayName *string `json:"display_name"`
			Role        string  `json:"role"`
			JoinedAt    *string `json:"joined_at"`
		} `json:"members"`
		Chats []struct {
			TgChatID int64   `json:"tg_chat_id"`
			Type     string  `json:"type"`
			LinkedAt *string `json:"linked_at"`
		} `json:"chats"`
		Bots []struct {
			BotTelegramID int64   `json:"bot_telegram_id"`
			LinkedAt      *string `json:"linked_at"`
		} `json:"bots"`
		PromoPending []struct {
			Code                  string  `json:"code"`
			Effect                string  `json:"effect"`
			Value                 int64   `json:"value"`
			FirstSubscriptionOnly bool    `json:"first_subscription_only"`
			Status                string  `json:"status"`
			CreatedAt             *string `json:"created_at"`
		} `json:"promo_pending"`
	}

	path := fmt.Sprintf("/api/users/%d/accounts/%d", userID, accountID)
	if err := c.get(ctx, path, nil, true, &raw); err != nil {
		return nil, err
	}

	members := make([]UserAccountMember, 0, len(raw.Members))
	for _, m := range raw.Members {
		members = append(members, UserAccountMember{
			BuyerID:     m.BuyerID,
			TgID:        m.TgID,
			Email:       m.Email,
			DisplayName: m.DisplayName,
			Role:        m.Role,
			JoinedAt:    parseOptionalTime(m.JoinedAt),
		})
	}

	chats := make([]UserAccountChat, 0, len(raw.Chats))
	for _, ch := range raw.Chats {
		chats = append(chats, UserAccountChat{
			TgChatID: ch.TgChatID,
			Type:     ch.Type,
			LinkedAt: parseOptionalTime(ch.LinkedAt),
		})
	}

	bots := make([]UserAccountBot, 0, len(raw.Bots))
	for _, b := range raw.Bots {
		bots = append(bots, UserAccountBot{
			BotTelegramID: b.BotTelegramID,
			LinkedAt:      parseOptionalTime(b.LinkedAt),
		})
	}

	promos := make([]UserAccountPromo, 0, len(raw.PromoPending))
	for _, pr := range raw.PromoPending {
		promos = append(promos, UserAccountPromo{
			Code:                  pr.Code,
			Effect:                pr.Effect,
			Value:                 pr.Value,
			FirstSubscriptionOnly: pr.FirstSubscriptionOnly,
			Status:                pr.Status,
			CreatedAt:             parseOptionalTime(pr.CreatedAt),
		})
	}

	return &UserAccountCard{
		AccountID:    raw.AccountID,
		Title:        raw.Title,
		IsPersonal:   raw.IsPersonal,
		IsPrimary:    raw.IsPrimary,
		CreatedAt:    parseOptionalTime(raw.CreatedAt),
		OwnerBuyerID: raw.OwnerBuyerID,
		Role:         raw.Role,
		Access:       accessFlags(raw.Access),
		AccessExpiry: raw.AccessExpiry,
		AccessEnd:    parseOptionalTime(raw.AccessEnd),
		Members:      members,
		Chats:        chats,
		Bots:         bots,
		PromoPending: promos,
	}, nil
}

// ExtendUserAccess продлевает доступ основного аккаунта человека на days дней
// (POST /api/users/{user_id}/access/extend).
func (c *Client) ExtendUserAccess(ctx context.Context, userID int64, botID int64, days int64) (*ExtendAccessResult, error) {
	return c.extendUserAccess(ctx, userID, botID, days, nil)
}

// ExtendUserAccessForAccount продлевает доступ конкретного аккаунта человека.
// Нужен, когда у человека несколько аккаунтов: без account_id CRM продлевает
// только основной. Человек обязан состоять в аккаунте, иначе вернётся
// *APIError с кодом not_found (404).
func (c *Client) ExtendUserAccessForAccount(ctx context.Context, userID, botID, days, accountID int64) (*ExtendAccessResult, error) {
	if accountID <= 0 {
		return nil, &ValidationError{Message: "account_id must be a positive integer"}
	}
	return c.extendUserAccess(ctx, userID, botID, days, &accountID)
}

func (c *Client) extendUserAccess(ctx context.Context, userID, botID, days int64, accountID *int64) (*ExtendAccessResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if days <= 0 {
		return nil, &ValidationError{Message: "days must be a positive integer"}
	}

	query := map[string]string{
		"bot_id": fmt.Sprintf("%d", botID),
		"days":   fmt.Sprintf("%d", days),
	}

	// Тело шлём только с аккаунтом: прежний вызов без тела остаётся
	// байт-в-байт тем же запросом, который понимает и старая CRM.
	var body any
	if accountID != nil {
		body = map[string]int64{"account_id": *accountID}
	}

	var raw struct {
		UserID    int64   `json:"user_id"`
		AccountID *int64  `json:"account_id"`
		AccessEnd *string `json:"access_end"`
	}

	if err := c.post(ctx, fmt.Sprintf("/api/users/%d/access/extend", userID), query, true, body, &raw); err != nil {
		return nil, err
	}

	var accessEnd *time.Time
	if raw.AccessEnd != nil {
		accessEnd = utils.ParseTime(*raw.AccessEnd)
	}

	return &ExtendAccessResult{
		UserID:    raw.UserID,
		AccountID: raw.AccountID,
		AccessEnd: accessEnd,
	}, nil
}

func (c *Client) ExtendAILimit(ctx context.Context, userID int64, millions int64) (*ExtendAILimitResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if millions <= 0 {
		return nil, &ValidationError{Message: "millions must be a positive integer"}
	}

	query := map[string]string{
		"millions": fmt.Sprintf("%d", millions),
	}

	var raw struct {
		PreviousAILimit int64 `json:"previous_ai_limit"`
		AILimit         int64 `json:"ai_limit"`
	}

	if err := c.post(ctx, fmt.Sprintf("/api/users/%d/ai-limit/extend", userID), query, true, nil, &raw); err != nil {
		return nil, err
	}

	return &ExtendAILimitResult{
		PreviousAILimit: raw.PreviousAILimit,
		AILimit:         raw.AILimit,
	}, nil
}

// GrantAITokens начисляет пакет токенов НОВОЙ системы: грант в AI-леджер под
// функцию + легаси-инкремент ai_limit (эквивалент покупки токен-тарифа).
// Допустимые пакеты валидирует CRM по активным токен-товарам каталога —
// клиент здесь проверяет только позитивность/непустоту, чтобы новый тариф
// не требовал правок SDK. ref — idempotency-ключ вызова (uuid): повтор с тем
// же ref безопасен и возвращает Granted=false. botID обязателен и должен быть
// положительным: значения по умолчанию нет, при botID <= 0 возвращается
// ConfigError без обращения к CRM.
func (c *Client) GrantAITokens(ctx context.Context, userID int64, tokens int64, function, ref string, botID int64) (*GrantAITokensResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if tokens <= 0 {
		return nil, &ValidationError{Message: "tokens must be a positive integer"}
	}
	if function == "" {
		return nil, &ValidationError{Message: "function must not be empty"}
	}
	if ref == "" {
		return nil, &ValidationError{Message: "ref must not be empty"}
	}
	if botID <= 0 {
		return nil, &ConfigError{Message: "bot_id must be a positive integer"}
	}

	query := map[string]string{
		"tokens":   fmt.Sprintf("%d", tokens),
		"function": function,
		"ref":      ref,
		"bot_id":   fmt.Sprintf("%d", botID),
	}

	var raw GrantAITokensResult
	if err := c.post(ctx, fmt.Sprintf("/api/users/%d/ai-tokens/grant", userID), query, true, nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}
