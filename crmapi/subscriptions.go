package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) AddAccess(ctx context.Context, input AddAccessInput) (*AddAccessResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		Created    bool    `json:"created"`
		ID         *int64  `json:"id"`
		UserID     int64   `json:"user_id"`
		BotID      int64   `json:"bot_id"`
		Action     string  `json:"action"`
		ActionDate *string `json:"action_date"`
		AccessEnd  *string `json:"access_end"`
	}

	if err := c.post(ctx, "/api/access/add", nil, true, input, &raw); err != nil {
		return nil, err
	}

	var actionDate *time.Time
	var accessEnd *time.Time

	if raw.ActionDate != nil {
		actionDate = utils.ParseTime(*raw.ActionDate)
	}
	if raw.AccessEnd != nil {
		accessEnd = utils.ParseTime(*raw.AccessEnd)
	}

	return &AddAccessResult{
		Created:    raw.Created,
		ID:         raw.ID,
		UserID:     raw.UserID,
		BotID:      raw.BotID,
		Action:     raw.Action,
		ActionDate: actionDate,
		AccessEnd:  accessEnd,
	}, nil
}

func (c *Client) SubscriptionsHistory(ctx context.Context, userID int64) (*SubscriptionsHistoryResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var raw struct {
		UserID  int64 `json:"user_id"`
		History []struct {
			Action     string  `json:"action"`
			BotID      int64   `json:"bot_id"`
			Access     any     `json:"access"`
			ActionDate *string `json:"action_date"`
			AccessEnd  *string `json:"access_end"`
			Payment    *struct {
				ID          *int64  `json:"id"`
				AmountMinor *int64  `json:"amount_minor"`
				Currency    *string `json:"currency"`
				Status      *string `json:"status"`
				DatePaid    *string `json:"date_paid"`
			} `json:"payment"`
			Staff *struct {
				ID   *int64  `json:"id"`
				Name *string `json:"name"`
			} `json:"staff"`
			Ref *string `json:"ref"`
		} `json:"history"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/users/%d/subscriptions/history", userID), nil, true, &raw); err != nil {
		return nil, err
	}

	history := make([]AccessHistoryItem, 0, len(raw.History))
	for _, h := range raw.History {
		var actionDate *time.Time
		var accessEnd *time.Time

		if h.ActionDate != nil {
			actionDate = utils.ParseTime(*h.ActionDate)
		}
		if h.AccessEnd != nil {
			accessEnd = utils.ParseTime(*h.AccessEnd)
		}

		var paymentRef *AccessPaymentRef
		if h.Payment != nil {
			var datePaid *time.Time
			if h.Payment.DatePaid != nil {
				datePaid = utils.ParseTime(*h.Payment.DatePaid)
			}

			paymentRef = &AccessPaymentRef{
				ID:          h.Payment.ID,
				AmountMinor: h.Payment.AmountMinor,
				Currency:    h.Payment.Currency,
				Status:      h.Payment.Status,
				DatePaid:    datePaid,
			}
		}

		var staffRef *AccessStaffRef
		if h.Staff != nil {
			staffRef = &AccessStaffRef{
				ID:   h.Staff.ID,
				Name: h.Staff.Name,
			}
		}

		history = append(history, AccessHistoryItem{
			Action:     h.Action,
			BotID:      h.BotID,
			Access:     h.Access,
			ActionDate: actionDate,
			AccessEnd:  accessEnd,
			Payment:    paymentRef,
			Staff:      staffRef,
			Ref:        h.Ref,
		})
	}

	return &SubscriptionsHistoryResult{
		UserID:  raw.UserID,
		History: history,
	}, nil
}

func (c *Client) AccessDefinitions(ctx context.Context) (*AccessDefinitionsResult, error) {
	var raw struct {
		Main   map[string]string `json:"main"`
		Poster map[string]string `json:"poster"`
	}

	if err := c.get(ctx, "/api/access/definitions", nil, true, &raw); err != nil {
		return nil, err
	}

	if raw.Main == nil {
		raw.Main = map[string]string{}
	}
	if raw.Poster == nil {
		raw.Poster = map[string]string{}
	}

	return &AccessDefinitionsResult{
		Main:   raw.Main,
		Poster: raw.Poster,
	}, nil
}

// SubscriptionsTransferLink requests a transfer link from the CRM for
// the given source user and bot.
//
// Behaviour:
//   - On a successful CRM response, all of TransferLink / Token / BotID /
//     ExpiresAt / TTLHours are populated.
//   - On a structured business error (CRM responded with error_code such
//     as not_supported / no_subscription / configuration_error), the
//     result is returned with ErrorCode and ErrorMessage set, and the
//     link fields left zero. This mirrors the Python SDK's contract and
//     lets callers branch on result.ErrorCode without unwrapping errors.
//   - Transport / system errors (network, decode, 5xx) are returned as a
//     non-nil error.
func (c *Client) SubscriptionsTransferLink(ctx context.Context, userID int64, botID int64) (*TransferLinkResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
		"bot_id":  fmt.Sprintf("%d", botID),
	}

	var raw struct {
		TransferLink string  `json:"transfer_link"`
		Token        string  `json:"token"`
		BotID        int64   `json:"bot_id"`
		ExpiresAt    *string `json:"expires_at"`
		TTLHours     int     `json:"ttl_hours"`
	}

	if err := c.post(ctx, "/api/subscriptions/transfer-link", query, true, nil, &raw); err != nil {
		if code, msg, ok := businessErrorCode(err); ok {
			return &TransferLinkResult{
				ErrorCode:    code,
				ErrorMessage: msg,
			}, nil
		}
		return nil, err
	}

	var expiresAt *time.Time
	if raw.ExpiresAt != nil {
		expiresAt = utils.ParseTime(*raw.ExpiresAt)
	}

	return &TransferLinkResult{
		TransferLink: raw.TransferLink,
		Token:        raw.Token,
		BotID:        raw.BotID,
		ExpiresAt:    expiresAt,
		TTLHours:     raw.TTLHours,
	}, nil
}

// SubscriptionsTransferRedeem performs the redeem of a TR_<token>
// link. Used by bots when handling a /start TR_<token> deep-link.
//
// Behaviour mirrors SubscriptionsTransferLink: CRM business errors are
// returned as a Result with Success=false and ErrorCode populated;
// transport errors propagate as Go errors. See TransferRedeemResult
// for the list of supported error codes.
func (c *Client) SubscriptionsTransferRedeem(ctx context.Context, input TransferRedeemInput) (*TransferRedeemResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var raw struct {
		SourceUserID    int64   `json:"source_user_id"`
		RecipientUserID int64   `json:"recipient_user_id"`
		BotID           int64   `json:"bot_id"`
		Access          any     `json:"access"`
		AccessEnd       *string `json:"access_end"`
	}

	if err := c.post(ctx, "/api/subscriptions/transfer/redeem", nil, true, input, &raw); err != nil {
		if code, msg, ok := businessErrorCode(err); ok {
			return &TransferRedeemResult{
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

	return &TransferRedeemResult{
		Success:         true,
		SourceUserID:    raw.SourceUserID,
		RecipientUserID: raw.RecipientUserID,
		BotID:           raw.BotID,
		Access:          raw.Access,
		AccessEnd:       accessEnd,
	}, nil
}

// businessErrorCode extracts the error_code from any of the SDK-typed
// errors that wrap a CRM-side structured response. CRM puts the
// error_code into the response envelope's `code` field; client.go
// dispatches by HTTP status into AuthError (401/403) /
// ValidationError (400/422) / APIError (everything else).
//
// For activation/transfer endpoints the same codes can come back with
// different statuses (e.g. wrong_recipient is 403, already_used is 409,
// invalid_token is 400) - we therefore probe all three error types and
// return the code regardless of class. Returns ok=false when err is
// transport-level and should be propagated to the caller.
func businessErrorCode(err error) (code string, message string, ok bool) {
	if err == nil {
		return "", "", false
	}
	switch e := err.(type) {
	case *APIError:
		if e.Code != "" {
			return e.Code, e.Message, true
		}
	case *AuthError:
		if e.Code != "" {
			return e.Code, e.Message, true
		}
	case *ValidationError:
		if e.Code != "" {
			return e.Code, e.Message, true
		}
	}
	return "", "", false
}
