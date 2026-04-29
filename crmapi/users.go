package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// CreateUser creates a user via POST /api/users (idempotent).
//
// If a registration for (UserID, BotID) already exists on the server, the
// response carries Created=false plus the existing record — no side effects.
// Otherwise Created=true and the fields describe the freshly created record.
func (c *Client) CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		Created  bool    `json:"created"`
		UserID   int64   `json:"user_id"`
		FullName string  `json:"full_name"`
		Username *string `json:"username"`
		BotID    int64   `json:"bot_id"`
		Refer    *string `json:"refer"`
		DateReg  *string `json:"date_reg"`
	}

	if err := c.post(ctx, "/api/users", nil, true, input, &raw); err != nil {
		return nil, err
	}

	var dateReg *time.Time
	if raw.DateReg != nil {
		dateReg = utils.ParseTime(*raw.DateReg)
	}

	return &CreateUserResult{
		Created:  raw.Created,
		UserID:   raw.UserID,
		FullName: raw.FullName,
		Username: raw.Username,
		BotID:    raw.BotID,
		Refer:    raw.Refer,
		DateReg:  dateReg,
	}, nil
}

// ListUsers fetches a paginated list of users registered for the given bot.
//
// botID is required (>0). limit must be >0 (CRM allows up to 1_000_000;
// the recommended default mirrors the Python SDK at 100_000). offset must be >=0.
//
// Each ListUserItem includes a Restricted flag — clients that consume CRM
// messages should skip restricted users.
func (c *Client) ListUsers(ctx context.Context, botID int64, limit int64, offset int64) (*ListUsersResult, error) {
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if limit <= 0 {
		return nil, &ValidationError{Message: "limit must be a positive integer"}
	}
	if offset < 0 {
		return nil, &ValidationError{Message: "offset must be non-negative"}
	}

	query := map[string]string{
		"bot_id": fmt.Sprintf("%d", botID),
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}

	var raw struct {
		BotID  int64 `json:"bot_id"`
		Limit  int64 `json:"limit"`
		Offset int64 `json:"offset"`
		Count  int64 `json:"count"`
		Items  []struct {
			UserID     int64   `json:"user_id"`
			FullName   string  `json:"full_name"`
			Username   *string `json:"username"`
			DateReg    *string `json:"date_reg"`
			Refer      *string `json:"refer"`
			Restricted bool    `json:"restricted"`
		} `json:"items"`
	}

	if err := c.get(ctx, "/api/users", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]ListUserItem, 0, len(raw.Items))
	for _, it := range raw.Items {
		var dateReg *time.Time
		if it.DateReg != nil {
			dateReg = utils.ParseTime(*it.DateReg)
		}
		items = append(items, ListUserItem{
			UserID:     it.UserID,
			FullName:   it.FullName,
			Username:   it.Username,
			DateReg:    dateReg,
			Refer:      it.Refer,
			Restricted: it.Restricted,
		})
	}

	return &ListUsersResult{
		BotID:  raw.BotID,
		Limit:  raw.Limit,
		Offset: raw.Offset,
		Count:  raw.Count,
		Items:  items,
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
