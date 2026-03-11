package crmapi

import "time"

type CalcItem struct {
	ID                       int64  `json:"id"`
	Title                    string `json:"title"`
	UnitPriceMinor           int64  `json:"unit_price_minor"`
	DiscountPercent          int64  `json:"discount_percent"`
	UnitPriceDiscountedMinor int64  `json:"unit_price_discounted_minor"`
	Quantity                 int64  `json:"quantity"`
	LineTotalMinor           int64  `json:"line_total_minor"`
}

type PaymentsCalculateResult struct {
	AmountMinor int64      `json:"amount_minor"`
	AmountUSD   *float64   `json:"amount_usd,omitempty"`
	Currency    string     `json:"currency"`
	Items       []CalcItem `json:"items"`
}

type InvoiceDraftResult struct {
	UUID    string `json:"uuid"`
	PayLink string `json:"pay_link"`
	Status  string `json:"status"`
}

type InvoiceIssueResult struct {
	PayURL string `json:"pay_url"`
	Status string `json:"status"`
}

type ActivationLink struct {
	BotID  *int64 `json:"bot_id,omitempty"`
	Code   string `json:"code"`
	IsUsed bool   `json:"is_used"`
	URL    string `json:"url"`
}

type PaymentHistoryItem struct {
	UUID            string           `json:"uuid"`
	Status          string           `json:"status"`
	StatusRU        string           `json:"status_ru"`
	ClientID        int64            `json:"client_id"`
	ClientEmail     *string          `json:"client_email,omitempty"`
	RefererID       *int64           `json:"referer_id,omitempty"`
	StaffID         *int64           `json:"staff_id,omitempty"`
	AmountMinor     int64            `json:"amount_minor"`
	FXRateRUBUSD    *float64         `json:"fx_rate_rub_usd,omitempty"`
	Currency        string           `json:"currency"`
	DiscountPercent *int64           `json:"discount_percent,omitempty"`
	Description     *string          `json:"description,omitempty"`
	Items           []map[string]any `json:"items"`
	Provider        *string          `json:"provider,omitempty"`
	PayLink         *string          `json:"pay_link,omitempty"`
	PayURL          *string          `json:"pay_url,omitempty"`
	DateCreate      *time.Time       `json:"date_create,omitempty"`
	DateInvoiced    *time.Time       `json:"date_invoiced,omitempty"`
	DatePaid        *time.Time       `json:"date_paid,omitempty"`
	Activation      []ActivationLink `json:"activation"`
}

type ConfirmPaymentResult struct {
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}

type RefundResult struct {
	UUID     string  `json:"uuid"`
	Provider string  `json:"provider"`
	Allowed  bool    `json:"allowed"`
	Message  string  `json:"message"`
	Status   *string `json:"status,omitempty"`
}

type InvoiceInfoResult struct {
	UUID            string           `json:"uuid"`
	Status          string           `json:"status"`
	StatusRU        string           `json:"status_ru"`
	ClientID        int64            `json:"client_id"`
	ClientEmail     *string          `json:"client_email,omitempty"`
	RefererID       *int64           `json:"referer_id,omitempty"`
	StaffID         *int64           `json:"staff_id,omitempty"`
	AmountMinor     int64            `json:"amount_minor"`
	FXRateRUBUSD    *float64         `json:"fx_rate_rub_usd,omitempty"`
	Currency        string           `json:"currency"`
	DiscountPercent *int64           `json:"discount_percent,omitempty"`
	Description     string           `json:"description"`
	Items           []map[string]any `json:"items"`
	Provider        string           `json:"provider"`
	PayLink         *string          `json:"pay_link,omitempty"`
	PayURL          *string          `json:"pay_url,omitempty"`
	DateCreate      *time.Time       `json:"date_create,omitempty"`
	DateInvoiced    *time.Time       `json:"date_invoiced,omitempty"`
	DatePaid        *time.Time       `json:"date_paid,omitempty"`
}
