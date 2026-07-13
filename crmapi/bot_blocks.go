package crmapi

import (
	"context"
	"fmt"
)

// BotBlocksListResult is the result of GET /api/bot-blocks: active per-bot
// "user blocked the bot" flags (crm_bot_blocks).
type BotBlocksListResult struct {
	BotID   int64
	UserIDs []int64
	Count   int64
}

// BotBlockUnblockResult is the result of POST /api/bot-blocks/unblock.
// Removed=false means the flag did not exist — that is also a success.
type BotBlockUnblockResult struct {
	Removed bool
}

// BotBlockReportResult is the result of POST /api/bot-blocks/report.
type BotBlockReportResult struct {
	Added bool
}

// ListBotBlocks fetches every user id flagged as "blocked the bot" for the
// given bot (crm_bot_blocks). Intended usage: prime a local cache at startup
// and refresh it hourly; clear entries via UnblockBotBlock when the user
// shows any activity.
func (c *Client) ListBotBlocks(ctx context.Context, botID int64) (*BotBlocksListResult, error) {
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}

	query := map[string]string{"bot_id": fmt.Sprintf("%d", botID)}

	var raw struct {
		BotID   int64   `json:"bot_id"`
		UserIDs []int64 `json:"user_ids"`
		Count   int64   `json:"count"`
	}

	if err := c.get(ctx, "/api/bot-blocks", query, true, &raw); err != nil {
		return nil, err
	}

	return &BotBlocksListResult{
		BotID:   raw.BotID,
		UserIDs: raw.UserIDs,
		Count:   raw.Count,
	}, nil
}

// UnblockBotBlock clears the "blocked the bot" flag for (botID, userID).
// Idempotent: calling it when no flag exists succeeds with Removed=false.
func (c *Client) UnblockBotBlock(ctx context.Context, botID, userID int64) (*BotBlockUnblockResult, error) {
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	body := map[string]int64{"bot_id": botID, "user_id": userID}

	var raw struct {
		Removed bool `json:"removed"`
	}

	if err := c.post(ctx, "/api/bot-blocks/unblock", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &BotBlockUnblockResult{Removed: raw.Removed}, nil
}

// ReportBotBlock sets the "blocked the bot" flag for (botID, userID).
// Idempotent (unique per bot+user). reason: "blocked" | "deactivated" |
// "chat_not_found"; if errText (raw Telegram error description) is non-empty,
// the CRM classifies the reason from it instead.
func (c *Client) ReportBotBlock(ctx context.Context, botID, userID int64, reason, errText string) (*BotBlockReportResult, error) {
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if reason == "" {
		reason = "blocked"
	}

	body := map[string]any{
		"bot_id":  botID,
		"user_id": userID,
		"reason":  reason,
	}
	if errText != "" {
		body["error"] = errText
	}

	var raw struct {
		Added bool `json:"added"`
	}

	if err := c.post(ctx, "/api/bot-blocks/report", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &BotBlockReportResult{Added: raw.Added}, nil
}
