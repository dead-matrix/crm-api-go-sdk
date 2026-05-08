package crmapi

import (
	"context"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// ActivationRedeem performs the redeem of an ACT_<token> deep-link.
// Used by bots when handling /start ACT_<token> after a paid CRM order.
//
// Behaviour mirrors SubscriptionsTransferRedeem: CRM business errors
// (already_used / expired / invalid_token / not_found / wrong_bot /
// wrong_recipient) are returned as a Result with Success=false and
// ErrorCode populated; transport errors propagate as Go errors.
//
// The CRM endpoint is idempotent on the activation_code state - a
// repeated call after a successful redeem returns Success=false with
// ErrorCode="already_used".
func (c *Client) ActivationRedeem(ctx context.Context, input ActivationRedeemInput) (*ActivationRedeemResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var raw struct {
		UserID           int64   `json:"user_id"`
		BotID            int64   `json:"bot_id"`
		Action           string  `json:"action"`
		Access           any     `json:"access"`
		AccessEnd        *string `json:"access_end"`
		ActivationCodeID int64   `json:"activation_code_id"`
		PaymentID        int64   `json:"payment_id"`
	}

	if err := c.post(ctx, "/api/activation/redeem", nil, true, input, &raw); err != nil {
		if code, msg, ok := businessErrorCode(err); ok {
			return &ActivationRedeemResult{
				Success:      false,
				ErrorCode:    code,
				ErrorMessage: msg,
			}, nil
		}
		return nil, err
	}

	var accessEnd *time.Time
	if raw.AccessEnd != nil {
		accessEnd = utils.ParseTime(*raw.AccessEnd)
	}

	return &ActivationRedeemResult{
		Success:          true,
		UserID:           raw.UserID,
		BotID:            raw.BotID,
		Action:           raw.Action,
		Access:           raw.Access,
		AccessEnd:        accessEnd,
		ActivationCodeID: raw.ActivationCodeID,
		PaymentID:        raw.PaymentID,
	}, nil
}
