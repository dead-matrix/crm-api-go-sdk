package crmapi

import (
	"strings"
	"time"
)

// ActivationRedeemInput is the request body for ActivationRedeem.
type ActivationRedeemInput struct {
	Token           string `json:"token"`
	RecipientUserID int64  `json:"recipient_user_id"`
	BotID           int64  `json:"bot_id"`
}

// Validate checks that all required fields are populated. CRM performs
// HMAC / expiry / state validation server-side; this only catches the
// obviously empty calls.
func (in ActivationRedeemInput) Validate() error {
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

// ActivationRedeemResult is the response of ActivationRedeem.
//
// Success=true means the activation code has been consumed and the
// recipient now has an Access row. Action will be "add" (first ever
// access for this bot) or "extend" (recipient already had an active
// subscription that was prolonged).
//
// Success=false carries an ErrorCode from the CRM:
//   - already_used:      activation code is already consumed
//   - expired:           token TTL elapsed
//   - invalid_token:     malformed or HMAC mismatch
//   - not_found:         no activation_code row for this token
//   - wrong_bot:         token belongs to a different bot
//   - wrong_recipient:   token was issued to a different Telegram user
type ActivationRedeemResult struct {
	Success bool `json:"success"`

	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`

	UserID           int64      `json:"user_id,omitempty"`
	BotID            int64      `json:"bot_id,omitempty"`
	Action           string     `json:"action,omitempty"` // "add" | "extend"
	Access           any        `json:"access,omitempty"`
	AccessEnd        *time.Time `json:"access_end,omitempty"`
	ActivationCodeID int64      `json:"activation_code_id,omitempty"`
	PaymentID        int64      `json:"payment_id,omitempty"`
}
