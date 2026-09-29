package crmapi

import "time"

type ReferreePayment struct {
	Date          *time.Time `json:"date,omitempty"`
	AmountMinor   int64      `json:"amount_minor"`
	CommissionUSD float64    `json:"commission_usd"`
	Status        string     `json:"status"`
}

type ReferreeInfo struct {
	UserID           int64             `json:"user_id"`
	FullName         *string           `json:"full_name,omitempty"`
	Username         *string           `json:"username,omitempty"`
	PaymentsCount    int64             `json:"payments_count"`
	PaymentsSumMinor int64             `json:"payments_sum_minor"`
	Payments         []ReferreePayment `json:"payments"`
}

type ReferralsInfoResult struct {
	RefLink       string `json:"ref_link"`
	Percent       int64  `json:"percent"`
	Registrations int64  `json:"registrations"`
	RefPayments   int64  `json:"ref_payments"`
	RefTotalSum   int64  `json:"ref_total_sum"`
	// EarnedUSD — Σ выплаченных комиссий = всего ВЫВЕДЕНО (USD). AvailableUSD
	// — текущий остаток к выводу. «Всего заработано» потребитель считает как
	// EarnedUSD + AvailableUSD. WithdrawnWalletUSD/WithdrawnSubscriptionUSD —
	// разбивка выведенного по методам (сумма ≈ EarnedUSD).
	EarnedUSD                float64        `json:"earned_usd"`
	AvailableUSD             float64        `json:"available_usd"`
	WithdrawnWalletUSD       float64        `json:"withdrawn_wallet_usd"`
	WithdrawnSubscriptionUSD float64        `json:"withdrawn_subscription_usd"`
	Referrees                []ReferreeInfo `json:"referrees"`
	// Partner: сводка партнёра (ключ partner в GET /referrals/info). nil,
	// если CRM его не прислала (старая CRM). Единицы в суффиксах JSON-ключей:
	// _usd_cents в USD-центах, _rub_kopecks в копейках. AvailableUSD верхнего
	// уровня теперь считается без холда: начисления в холде лежат в
	// Partner.OnHoldUSDCents.
	Partner *ReferralPartner `json:"partner,omitempty"`
}

// PartnerPromoCode: промокод партнёра. Kind, например "partner_auto".
type PartnerPromoCode struct {
	Code string `json:"code"`
	Kind string `json:"kind"`
}

// PartnerPendingWithdrawal: открытая заявка партнёра на вывод.
// WithdrawalID строковый, вида "wdr_...".
type PartnerPendingWithdrawal struct {
	WithdrawalID   string     `json:"withdrawal_id"`
	AmountUSDCents int64      `json:"amount_usd_cents"`
	Method         string     `json:"method"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
}

// PartnerAccrual: реферальное начисление. Kind: "first" | "recurring".
// AvailableAt: момент выхода из холда; Reverted: начисление отменено.
type PartnerAccrual struct {
	AmountUSDCents int64      `json:"amount_usd_cents"`
	Kind           string     `json:"kind"`
	AvailableAt    *time.Time `json:"available_at,omitempty"`
	Reverted       bool       `json:"reverted"`
}

// ReferralPartner: сводка партнёра из GET /referrals/info (ключ partner).
//
// Суммы в USD-центах (_usd_cents) и копейках (_rub_kopecks).
// AvailableUSDCents не бывает отрицательным; OnHoldUSDCents может быть
// меньше нуля (отмены начислений после выхода из холда).
// PendingWithdrawal nil, если открытой заявки нет.
type ReferralPartner struct {
	RefLink                       string                    `json:"ref_link"`
	RefBotLink                    string                    `json:"ref_bot_link"`
	Percent                       int64                     `json:"percent"`
	FirstPercent                  int64                     `json:"first_percent"`
	RecurringPercent              int64                     `json:"recurring_percent"`
	HoldDays                      int64                     `json:"hold_days"`
	Registrations                 int64                     `json:"registrations"`
	PaidReferralsCount            int64                     `json:"paid_referrals_count"`
	PromoActivationsCount         int64                     `json:"promo_activations_count"`
	ReferredPaymentsCount         int64                     `json:"referred_payments_count"`
	ReferredTurnoverRUBKopecks    int64                     `json:"referred_turnover_rub_kopecks"`
	EarnedUSDCents                int64                     `json:"earned_usd_cents"`
	AvailableUSDCents             int64                     `json:"available_usd_cents"`
	OnHoldUSDCents                int64                     `json:"on_hold_usd_cents"`
	SpentOnSubscriptionsUSDCents  int64                     `json:"spent_on_subscriptions_usd_cents"`
	WithdrawnWalletUSDCents       int64                     `json:"withdrawn_wallet_usd_cents"`
	WithdrawnSubscriptionUSDCents int64                     `json:"withdrawn_subscription_usd_cents"`
	MinWithdrawalUSDCents         int64                     `json:"min_withdrawal_usd_cents"`
	WithdrawMethods               []string                  `json:"withdraw_methods"`
	PromoCodes                    []PartnerPromoCode        `json:"promo_codes"`
	PendingWithdrawal             *PartnerPendingWithdrawal `json:"pending_withdrawal,omitempty"`
	RecentAccruals                []PartnerAccrual          `json:"recent_accruals"`
}

// WithdrawRequestResult — результат заявки на вывод
// (POST /referrals/withdraw/request).
//
// Status: "no_balance" | "already_pending" | "created". Поля заполняются по
// ветке (nullable): AmountUSD/WithdrawalID — для pending/created,
// AvailableUSD — для no_balance.
type WithdrawRequestResult struct {
	Status       string   `json:"status"`
	WithdrawalID *int64   `json:"withdrawal_id,omitempty"`
	AmountUSD    *float64 `json:"amount_usd,omitempty"`
	Method       *string  `json:"method,omitempty"`
	AvailableUSD *float64 `json:"available_usd,omitempty"`
}

// WithdrawSettleResult — результат проведения вывода
// (POST /referrals/withdraw/settle).
type WithdrawSettleResult struct {
	Status            string  `json:"status"`
	WithdrawalID      int64   `json:"withdrawal_id"`
	PaidUSD           float64 `json:"paid_usd"`
	AvailableAfterUSD float64 `json:"available_after_usd"`
	Method            string  `json:"method"`
}
