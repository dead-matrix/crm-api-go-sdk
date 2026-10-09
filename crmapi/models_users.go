package crmapi

import "time"

type UserBotInfo struct {
	BotID   int64  `json:"bot_id"`
	BotName string `json:"bot_name"`
	// AccountID - аккаунт, чей доступ показан в этом элементе: в CRM SocialTraff
	// это основной аккаунт человека. nil, если у человека нет ни одного аккаунта
	// или CRM старая и ключа не присылает.
	AccountID  *int64     `json:"account_id,omitempty"`
	Registered *time.Time `json:"registered,omitempty"`
	Refer      *string    `json:"refer,omitempty"`
	Access     any        `json:"access,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
	// AccessExpiry — пофичевые сроки живых фич {ключ: end_iso}: у каждого
	// модуля свой конец (в отличие от общего AccessEnd = максимума). Пусто
	// для legacy-строк без карты.
	AccessExpiry map[string]string `json:"access_expiry,omitempty"`
	// Subscription-freeze fields, populated when the latest access row for this
	// bot is action="freeze". Frozen flags the state; FrozenAt is the moment of
	// freezing; FrozenExpiry is the per-feature end-date snapshot {feature:
	// end_iso} captured at freeze time (remaining days per feature =
	// FrozenExpiry[feat] − FrozenAt, since the frozen period doesn't elapse).
	Frozen       bool              `json:"frozen,omitempty"`
	FrozenAt     *time.Time        `json:"frozen_at,omitempty"`
	FrozenExpiry map[string]string `json:"frozen_expiry,omitempty"`
}

// UserBuyer - покупатель продукта, привязанный к Telegram-пользователю
// (ключ buyer карточки GET /api/users/{user_id}).
type UserBuyer struct {
	// BuyerID - идентификатор покупателя. Именно его, а не Telegram id, CRM
	// ждёт в параметре user_id реферальных маршрутов: передавайте BuyerID в
	// ReferralsInfo, ReferralsWithdrawRequest и ReferralsWithdrawSettle.
	BuyerID     int64   `json:"buyer_id"`
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	// RefCode - собственный реферальный код покупателя (для ссылок ?ref=).
	// Refer и Landing - метка его регистрации: UTM или чужой реф-код и страница
	// входа; nil, если метки нет.
	RefCode   *string    `json:"ref_code,omitempty"`
	Refer     *string    `json:"refer,omitempty"`
	Landing   *string    `json:"landing,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// UserAccount - участие человека в аккаунте (элемент accounts карточки
// GET /api/users/{user_id}). Подписка в CRM SocialTraff принадлежит аккаунту,
// поэтому доступ и сроки лежат здесь, а не на человеке.
//
// Поля от AccessEnd и ниже появились позже остальных: CRM, которая их ещё не
// отдаёт, оставляет в них нулевые значения и nil.
type UserAccount struct {
	AccountID int64  `json:"account_id"`
	Title     string `json:"title"`
	// Role - роль человека в аккаунте: owner | admin | buyer | viewer.
	Role string `json:"role"`
	// Access - живые фичи аккаунта {ключ: true}; nil, если активного доступа
	// нет. AccessExpiry - конец каждой живой фичи {ключ: ISO-строка}; nil у
	// неактивного доступа и у старых строк без карты сроков. AccessEnd - самый
	// поздний из этих концов.
	Access       map[string]bool   `json:"access,omitempty"`
	AccessExpiry map[string]string `json:"access_expiry,omitempty"`
	AccessEnd    *time.Time        `json:"access_end,omitempty"`
	// IsPersonal - аккаунт, заведённый при регистрации покупателя. IsPrimary -
	// основной аккаунт человека: в него идут действия без явного account_id
	// (AddAccess, ManageAccess, ExtendUserAccess, SubscriptionsHistory).
	IsPersonal   bool       `json:"is_personal"`
	IsPrimary    bool       `json:"is_primary"`
	OwnerBuyerID int64      `json:"owner_buyer_id"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	// JoinedAt - когда человек стал участником аккаунта.
	JoinedAt     *time.Time `json:"joined_at,omitempty"`
	MembersCount int64      `json:"members_count"`
	ChatsCount   int64      `json:"chats_count"`
	BotsCount    int64      `json:"bots_count"`
}

type GetUserResult struct {
	UserID   int64   `json:"user_id"`
	FullName *string `json:"full_name,omitempty"`
	Username *string `json:"username,omitempty"`
	Status   *string `json:"status,omitempty"`
	// HasActiveSubscription / Frozen - сводные признаки подписки по всем ботам
	// (тот же предикат CRM, что у SubscriptionState); BotsInfo - детализация.
	HasActiveSubscription bool          `json:"has_active_subscription"`
	Frozen                bool          `json:"frozen"`
	BotsInfo              []UserBotInfo `json:"bots_info"`
	// Buyer - покупатель продукта за этим Telegram-пользователем. nil, если
	// человек писал боту, но покупателем не стал, либо CRM ключ не присылает.
	Buyer *UserBuyer `json:"buyer,omitempty"`
	// Accounts - все активные участия человека в аккаунтах. Всегда не nil:
	// у человека без аккаунтов и у CRM без этого ключа срез пустой.
	Accounts []UserAccount `json:"accounts"`
}

// UserAccountMember - участник аккаунта в карточке GetUserAccount. TgID равен
// nil у покупателя без привязанного Telegram.
type UserAccountMember struct {
	BuyerID     int64      `json:"buyer_id"`
	TgID        *int64     `json:"tg_id,omitempty"`
	Email       *string    `json:"email,omitempty"`
	DisplayName *string    `json:"display_name,omitempty"`
	Role        string     `json:"role"`
	JoinedAt    *time.Time `json:"joined_at,omitempty"`
}

// UserAccountChat - чат или канал Telegram, подключённый к аккаунту.
type UserAccountChat struct {
	TgChatID int64      `json:"tg_chat_id"`
	Type     string     `json:"type"`
	LinkedAt *time.Time `json:"linked_at,omitempty"`
}

// UserAccountBot - Telegram-бот, подключённый к аккаунту.
type UserAccountBot struct {
	BotTelegramID int64      `json:"bot_telegram_id"`
	LinkedAt      *time.Time `json:"linked_at,omitempty"`
}

// UserAccountPromo - промокод, который аккаунт активировал, но выгоду по нему
// ещё не получил: она ждёт платежа. Effect - вид выгоды (bonus_days |
// discount_percent), Value - её размер в днях или процентах. Значения
// зафиксированы на момент активации и не меняются при правке самого кода.
type UserAccountPromo struct {
	Code                  string     `json:"code"`
	Effect                string     `json:"effect"`
	Value                 int64      `json:"value"`
	FirstSubscriptionOnly bool       `json:"first_subscription_only"`
	Status                string     `json:"status"`
	CreatedAt             *time.Time `json:"created_at,omitempty"`
}

// UserAccountCard - полная карточка аккаунта глазами одного его участника
// (GET /api/users/{user_id}/accounts/{account_id}). В отличие от UserAccount
// несёт сами списки участников, чатов и ботов, а не только их количество.
//
// Role - роль запрошенного человека в этом аккаунте. Access, AccessExpiry и
// AccessEnd читаются так же, как у UserAccount. Срезы всегда не nil.
type UserAccountCard struct {
	AccountID    int64               `json:"account_id"`
	Title        string              `json:"title"`
	IsPersonal   bool                `json:"is_personal"`
	IsPrimary    bool                `json:"is_primary"`
	CreatedAt    *time.Time          `json:"created_at,omitempty"`
	OwnerBuyerID int64               `json:"owner_buyer_id"`
	Role         string              `json:"role"`
	Access       map[string]bool     `json:"access,omitempty"`
	AccessExpiry map[string]string   `json:"access_expiry,omitempty"`
	AccessEnd    *time.Time          `json:"access_end,omitempty"`
	Members      []UserAccountMember `json:"members"`
	Chats        []UserAccountChat   `json:"chats"`
	Bots         []UserAccountBot    `json:"bots"`
	PromoPending []UserAccountPromo  `json:"promo_pending"`
}

// CreateUserResult is the result of POST /api/users (idempotent).
//
// If a registration for (UserID, BotID) already existed, Created is false and
// the rest of the fields contain the existing record. Otherwise Created is
// true and the fields describe the newly created record.
//
// FullName is nullable: on the idempotent path the server returns
// `user.full_name` directly from the database row, which is technically a
// nullable column. New registrations always have a non-nil FullName.
type CreateUserResult struct {
	Created  bool       `json:"created"`
	UserID   int64      `json:"user_id"`
	FullName *string    `json:"full_name,omitempty"`
	Username *string    `json:"username,omitempty"`
	BotID    int64      `json:"bot_id"`
	Refer    *string    `json:"refer,omitempty"`
	DateReg  *time.Time `json:"date_reg,omitempty"`
}

// ListUserItem is one element of GET /api/users?bot_id=... response.
//
// FullName is nullable: the server returns `user.full_name` directly from
// the database row, which is technically a nullable column.
type ListUserItem struct {
	UserID     int64      `json:"user_id"`
	FullName   *string    `json:"full_name,omitempty"`
	Username   *string    `json:"username,omitempty"`
	DateReg    *time.Time `json:"date_reg,omitempty"`
	Refer      *string    `json:"refer,omitempty"`
	Restricted bool       `json:"restricted"`
	// HasActiveSubscription / Frozen - признаки подписки уровня пользователя
	// по всем ботам, а не только по BotID из запроса.
	HasActiveSubscription bool `json:"has_active_subscription"`
	Frozen                bool `json:"frozen"`
}

// ListUsersResult is the response of GET /api/users?bot_id=...&limit=...&offset=...
type ListUsersResult struct {
	BotID  int64          `json:"bot_id"`
	Limit  int64          `json:"limit"`
	Offset int64          `json:"offset"`
	Count  int64          `json:"count"`
	Items  []ListUserItem `json:"items"`
}

type UpdateUserResult struct {
	UserID   int64   `json:"user_id"`
	FullName string  `json:"full_name"`
	Username *string `json:"username,omitempty"`
}

// AddAccessResult - результат AddAccess.
//
// AccountID - аккаунт, в который записана строка доступа: по нему видно, какой
// аккаунт CRM выбрала основным, когда вызов шёл только по UserID. nil у старой
// CRM без этого ключа. UserID равен нулю, если вызов шёл только по AccountID.
type AddAccessResult struct {
	Created    bool       `json:"created"`
	ID         *int64     `json:"id,omitempty"`
	AccountID  *int64     `json:"account_id,omitempty"`
	UserID     int64      `json:"user_id"`
	BotID      int64      `json:"bot_id"`
	Action     string     `json:"action"`
	ActionDate *time.Time `json:"action_date,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
}

// ExtendAccessResult - результат ExtendUserAccess и ExtendUserAccessForAccount.
// AccountID - аккаунт, чей доступ продлён (nil у старой CRM без этого ключа).
type ExtendAccessResult struct {
	UserID    int64      `json:"user_id"`
	AccountID *int64     `json:"account_id,omitempty"`
	AccessEnd *time.Time `json:"access_end,omitempty"`
}

type ExtendAILimitResult struct {
	PreviousAILimit int64 `json:"previous_ai_limit"`
	AILimit         int64 `json:"ai_limit"`
}

// GrantAITokensResult - результат начисления пакета токенов новой системы
// (леджер + легаси ai_limit). Granted=false - повтор с тем же ref (идемпотентный
// no-op), previous/ai_limit тогда null у CRM и остаются нулями здесь. CRM
// SocialTraff легаси-лимит не ведёт и отдаёт в них null всегда.
type GrantAITokensResult struct {
	Granted bool `json:"granted"`
	// AccountID - аккаунт, на чей AI-баланс легло начисление: CRM SocialTraff
	// ведёт баланс по основному аккаунту человека. Ноль у CRM без этого ключа.
	AccountID       int64  `json:"account_id"`
	Function        string `json:"function"`
	Tokens          int64  `json:"tokens"`
	PreviousAILimit int64  `json:"previous_ai_limit"`
	AILimit         int64  `json:"ai_limit"`
	BalanceTokens   int64  `json:"balance_tokens"`
	// Unlimited - у аккаунта безлимит по этой функции: BalanceTokens остаётся
	// настоящим остатком, но расход им не ограничен.
	Unlimited bool `json:"unlimited"`
}
