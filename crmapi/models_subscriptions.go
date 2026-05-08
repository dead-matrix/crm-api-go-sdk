package crmapi

import (
	"strings"
	"time"
)

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

// TransferLinkResult is the response of SubscriptionsTransferLink.
//
// Backward compatibility: TransferLink is the legacy field with the
// fully-formed t.me/...?start=TR_<token> URL and is still populated on
// success.
//
// New fields (Token, BotID, ExpiresAt, TTLHours) carry the structured
// representation introduced together with the CRM activation/transfer
// outbox flow. Token is the raw TR_<base32> string for callers that
// want to embed it elsewhere; ExpiresAt is the absolute UTC moment when
// the link stops working.
//
// On structured business errors (CRM responded with an error_code such
// as not_supported / no_subscription / configuration_error), the result
// is returned with ErrorCode and ErrorMessage populated and the link
// fields left zero. Transport / system errors are still returned as a
// non-nil error from SubscriptionsTransferLink itself.
type TransferLinkResult struct {
	TransferLink string     `json:"transfer_link,omitempty"`
	Token        string     `json:"token,omitempty"`
	BotID        int64      `json:"bot_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	TTLHours     int        `json:"ttl_hours,omitempty"`

	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// TransferRedeemInput is the request body for SubscriptionsTransferRedeem.
type TransferRedeemInput struct {
	Token           string `json:"token"`
	RecipientUserID int64  `json:"recipient_user_id"`
	BotID           int64  `json:"bot_id"`
}

// Validate checks that the input has all required fields populated.
// Network-level validation (HMAC, expiry, signature) is performed
// server-side; this only catches obviously malformed calls.
func (in TransferRedeemInput) Validate() error {
	if strings.TrimSpace(in.Token) == "" {
		return &ValidationError{Message: "token must be provided"}
	}
	if in.RecipientUserID <= 0 {
		return &ValidationError{Message: "recipient_user_id must be a positive integer"}
	}
	if in.BotID <= 0 {
		return &ValidationError{Message: "bot_id must be a positive integer"}
	}
	return nil
}

// TransferRedeemResult is the response of SubscriptionsTransferRedeem.
//
// Success=true means the source's subscription has been moved to the
// recipient atomically: the access JSON and access_end the recipient
// now holds are returned for downstream UI / local state sync.
//
// Success=false carries an ErrorCode from the CRM:
//   - no_subscription:        source has no active subscription (or
//     transfer was already redeemed, since the first redeem revokes it)
//   - recipient_has_access:   recipient already has an active subscription
//   - invalid_token:          token is malformed or HMAC mismatch
//   - expired:                token TTL elapsed
//   - wrong_bot:              token belongs to a different bot
//   - same_user:              source == recipient
type TransferRedeemResult struct {
	Success bool `json:"success"`

	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`

	SourceUserID    int64      `json:"source_user_id,omitempty"`
	RecipientUserID int64      `json:"recipient_user_id,omitempty"`
	BotID           int64      `json:"bot_id,omitempty"`
	Access          any        `json:"access,omitempty"`
	AccessEnd       *time.Time `json:"access_end,omitempty"`
}
