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
	RefLink       string  `json:"ref_link"`
	Percent       int64   `json:"percent"`
	Registrations int64   `json:"registrations"`
	RefPayments   int64   `json:"ref_payments"`
	RefTotalSum   int64   `json:"ref_total_sum"`
	// EarnedUSD — Σ выплаченных комиссий = всего ВЫВЕДЕНО (USD). AvailableUSD
	// — текущий остаток к выводу. «Всего заработано» потребитель считает как
	// EarnedUSD + AvailableUSD. WithdrawnWalletUSD/WithdrawnSubscriptionUSD —
	// разбивка выведенного по методам (сумма ≈ EarnedUSD).
	EarnedUSD                float64        `json:"earned_usd"`
	AvailableUSD             float64        `json:"available_usd"`
	WithdrawnWalletUSD       float64        `json:"withdrawn_wallet_usd"`
	WithdrawnSubscriptionUSD float64        `json:"withdrawn_subscription_usd"`
	Referrees                []ReferreeInfo `json:"referrees"`
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
