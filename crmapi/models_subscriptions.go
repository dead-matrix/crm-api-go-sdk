package crmapi

import "time"

type AccessPaymentRef struct {
	ID          *int64     `json:"id,omitempty"`
	AmountMinor *int64     `json:"amount_minor,omitempty"`
	Currency    *string    `json:"currency,omitempty"`
	Status      *string    `json:"status,omitempty"`
	DatePaid    *time.Time `json:"date_paid,omitempty"`
}

type AccessStaffRef struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type AccessHistoryItem struct {
	Action     string            `json:"action"`
	BotID      int64             `json:"bot_id"`
	Access     any               `json:"access,omitempty"`
	ActionDate *time.Time        `json:"action_date,omitempty"`
	AccessEnd  *time.Time        `json:"access_end,omitempty"`
	Payment    *AccessPaymentRef `json:"payment,omitempty"`
	Staff      *AccessStaffRef   `json:"staff,omitempty"`
	Ref        *string           `json:"ref,omitempty"`
}

type SubscriptionsHistoryResult struct {
	UserID  int64               `json:"user_id"`
	History []AccessHistoryItem `json:"history"`
}

type AccessDefinitionsResult struct {
	Main   map[string]string `json:"main"`
	Poster map[string]string `json:"poster"`
}

type TransferLinkResult struct {
	TransferLink string `json:"transfer_link"`
}
