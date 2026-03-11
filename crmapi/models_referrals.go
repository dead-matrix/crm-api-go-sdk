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
	RefLink       string         `json:"ref_link"`
	Percent       int64          `json:"percent"`
	Registrations int64          `json:"registrations"`
	RefPayments   int64          `json:"ref_payments"`
	RefTotalSum   int64          `json:"ref_total_sum"`
	EarnedUSD     float64        `json:"earned_usd"`
	AvailableUSD  float64        `json:"available_usd"`
	Referrees     []ReferreeInfo `json:"referrees"`
}
