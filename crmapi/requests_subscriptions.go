package crmapi

import (
	"strings"
	"time"
)

type ActionType string

const (
	ActionAdd    ActionType = "add"
	ActionExtend ActionType = "extend"
	ActionRevoke ActionType = "revoke"
	ActionRefund ActionType = "refund"
)

type AddAccessInput struct {
	UserID     int64      `json:"user_id"`
	BotID      int64      `json:"bot_id"`
	Action     ActionType `json:"action"`
	Access     any        `json:"access,omitempty"`
	ActionDate *time.Time `json:"action_date,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
	PaymentID  *int64     `json:"payment_id,omitempty"`
	Ref        *string    `json:"ref,omitempty"`
}

func (in AddAccessInput) Validate() error {
	if in.UserID <= 0 {
		return &ValidationError{Message: "user_id must be a positive integer"}
	}
	if in.BotID <= 0 {
		return &ValidationError{Message: "bot_id must be a positive integer"}
	}

	action := ActionType(strings.ToLower(strings.TrimSpace(string(in.Action))))
	switch action {
	case ActionAdd, ActionExtend, ActionRevoke, ActionRefund:
	default:
		return &ValidationError{Message: "action must be one of: add, extend, revoke, refund"}
	}

	if in.PaymentID != nil && *in.PaymentID <= 0 {
		return &ValidationError{Message: "payment_id must be a positive integer"}
	}
	if in.Ref != nil && len(*in.Ref) > 2048 {
		return &ValidationError{Message: "ref must be at most 2048 characters"}
	}

	return nil
}

func (in AddAccessInput) normalized() AddAccessInput {
	in.Action = ActionType(strings.ToLower(strings.TrimSpace(string(in.Action))))
	return in
}
