package crmapi

import (
	"context"
	"fmt"
)

func (c *Client) ServersRestart(ctx context.Context, userID int64, botID int64) (*ServerRestartResult, error) {
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
		Message string `json:"message"`
	}

	if err := c.post(ctx, "/api/servers/restart", query, true, nil, &raw); err != nil {
		return nil, err
	}

	return &ServerRestartResult{
		Message: raw.Message,
	}, nil
}
