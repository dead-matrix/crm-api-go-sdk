package crmapi

import "strings"

type CreateUserInput struct {
	UserID   int64   `json:"user_id"`
	FullName string  `json:"full_name"`
	Username *string `json:"username,omitempty"`
	BotID    int64   `json:"bot_id"`
	Refer    *string `json:"refer,omitempty"`
}

func (in CreateUserInput) Validate() error {
	if in.UserID <= 0 {
		return &ValidationError{Message: "user_id must be a positive integer"}
	}
	if strings.TrimSpace(in.FullName) == "" {
		return &ValidationError{Message: "full_name must not be empty"}
	}
	if len(strings.TrimSpace(in.FullName)) > 128 {
		return &ValidationError{Message: "full_name must be at most 128 characters"}
	}
	if in.Username != nil && len(*in.Username) > 128 {
		return &ValidationError{Message: "username must be at most 128 characters"}
	}
	if in.BotID <= 0 {
		return &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if in.Refer != nil && len(*in.Refer) > 32 {
		return &ValidationError{Message: "refer must be at most 32 characters"}
	}
	return nil
}

func (in CreateUserInput) normalized() CreateUserInput {
	in.FullName = strings.TrimSpace(in.FullName)
	return in
}

type UpdateUserInput struct {
	FullName string  `json:"full_name"`
	Username *string `json:"username,omitempty"`
}

func (in UpdateUserInput) Validate() error {
	if strings.TrimSpace(in.FullName) == "" {
		return &ValidationError{Message: "full_name must not be empty"}
	}
	if len(strings.TrimSpace(in.FullName)) > 128 {
		return &ValidationError{Message: "full_name must be at most 128 characters"}
	}
	if in.Username != nil && len(*in.Username) > 128 {
		return &ValidationError{Message: "username must be at most 128 characters"}
	}
	return nil
}

func (in UpdateUserInput) normalized() UpdateUserInput {
	in.FullName = strings.TrimSpace(in.FullName)
	return in
}
