package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		Created bool `json:"created"`
	}

	if err := c.post(ctx, "/api/users", nil, true, input, &raw); err != nil {
		return nil, err
	}

	return &CreateUserResult{
		Created: raw.Created,
	}, nil
}

func (c *Client) UpdateUser(ctx context.Context, userID int64, input UpdateUserInput) (*UpdateUserResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		UserID   int64   `json:"user_id"`
		FullName string  `json:"full_name"`
		Username *string `json:"username"`
	}

	if err := c.put(ctx, fmt.Sprintf("/api/users/%d", userID), nil, true, input, &raw); err != nil {
		return nil, err
	}

	return &UpdateUserResult{
		UserID:   raw.UserID,
		FullName: raw.FullName,
		Username: raw.Username,
	}, nil
}

func (c *Client) GetUser(ctx context.Context, userID int64) (*GetUserResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var raw struct {
		UserID   int64   `json:"user_id"`
		FullName *string `json:"full_name"`
		Username *string `json:"username"`
		Status   *string `json:"status"`
		BotsInfo []struct {
			BotID      int64   `json:"bot_id"`
			BotName    string  `json:"bot_name"`
			Registered *string `json:"registered"`
			Refer      *string `json:"refer"`
			Access     any     `json:"access"`
			AccessEnd  *string `json:"access_end"`
		} `json:"bots_info"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/users/%d", userID), nil, true, &raw); err != nil {
		return nil, err
	}

	bots := make([]UserBotInfo, 0, len(raw.BotsInfo))
	for _, item := range raw.BotsInfo {
		var registered *time.Time
		var accessEnd *time.Time

		if item.Registered != nil {
			registered = utils.ParseTime(*item.Registered)
		}
		if item.AccessEnd != nil {
			accessEnd = utils.ParseTime(*item.AccessEnd)
		}

		bots = append(bots, UserBotInfo{
			BotID:      item.BotID,
			BotName:    item.BotName,
			Registered: registered,
			Refer:      item.Refer,
			Access:     item.Access,
			AccessEnd:  accessEnd,
		})
	}

	return &GetUserResult{
		UserID:   raw.UserID,
		FullName: raw.FullName,
		Username: raw.Username,
		Status:   raw.Status,
		BotsInfo: bots,
	}, nil
}

func (c *Client) ExtendUserAccess(ctx context.Context, userID int64, botID int64, days int64) (*ExtendAccessResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if days <= 0 {
		return nil, &ValidationError{Message: "days must be a positive integer"}
	}

	query := map[string]string{
		"bot_id": fmt.Sprintf("%d", botID),
		"days":   fmt.Sprintf("%d", days),
	}

	var raw struct {
		UserID    int64   `json:"user_id"`
		AccessEnd *string `json:"access_end"`
	}

	if err := c.post(ctx, fmt.Sprintf("/api/users/%d/access/extend", userID), query, true, nil, &raw); err != nil {
		return nil, err
	}

	var accessEnd *time.Time
	if raw.AccessEnd != nil {
		accessEnd = utils.ParseTime(*raw.AccessEnd)
	}

	return &ExtendAccessResult{
		UserID:    raw.UserID,
		AccessEnd: accessEnd,
	}, nil
}

func (c *Client) ExtendAILimit(ctx context.Context, userID int64, millions int64) (*ExtendAILimitResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if millions <= 0 {
		return nil, &ValidationError{Message: "millions must be a positive integer"}
	}

	query := map[string]string{
		"millions": fmt.Sprintf("%d", millions),
	}

	var raw struct {
		PreviousAILimit int64 `json:"previous_ai_limit"`
		AILimit         int64 `json:"ai_limit"`
	}

	if err := c.post(ctx, fmt.Sprintf("/api/users/%d/ai-limit/extend", userID), query, true, nil, &raw); err != nil {
		return nil, err
	}

	return &ExtendAILimitResult{
		PreviousAILimit: raw.PreviousAILimit,
		AILimit:         raw.AILimit,
	}, nil
}
