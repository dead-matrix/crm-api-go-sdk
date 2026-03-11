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
		TransferLink string `json:"transfer_link"`
	}

	if err := c.post(ctx, "/api/subscriptions/transfer-link", query, true, nil, &raw); err != nil {
		return nil, err
	}

	return &TransferLinkResult{
		TransferLink: raw.TransferLink,
	}, nil
}
