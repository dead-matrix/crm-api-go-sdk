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

// ServersStatus reports whether a user's bot worker is up, so a caller can
// poll the real completion of a restart instead of trusting the restart call
// (which returns as soon as the open command is accepted, before the worker
// is actually serving again). Lightweight read — safe to poll on a loop.
func (c *Client) ServersStatus(ctx context.Context, userID int64, botID int64) (*ServerStatusResult, error) {
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
		Bound bool `json:"bound"`
		Up    bool `json:"up"`
	}

	if err := c.get(ctx, "/api/servers/status", query, true, &raw); err != nil {
		return nil, err
	}

	return &ServerStatusResult{
		Bound: raw.Bound,
		Up:    raw.Up,
	}, nil
}
