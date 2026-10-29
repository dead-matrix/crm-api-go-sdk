package crmapi

import (
	"context"
	"fmt"
	"strings"
)

// Категории флага "бот не может писать человеку" для ListBotBlocksByKind.
// Различать их нужно потому, что судьба у них разная: заблокировавший сам
// может вернуться, а удалённый аккаунт или пропавший чат сам не вернётся.
const (
	// BotBlockKindBlocked - человек сам заблокировал бота.
	BotBlockKindBlocked = "blocked"
	// BotBlockKindUnreachable - Telegram не доставляет: аккаунт деактивирован
	// или чат не найден.
	BotBlockKindUnreachable = "unreachable"
)

// BotBlockCounts - число флагов бота по категориям. Считается по всем флагам
// бота и от фильтра kind в запросе не зависит.
type BotBlockCounts struct {
	Blocked     int64 `json:"blocked"`
	Unreachable int64 `json:"unreachable"`
	Total       int64 `json:"total"`
}

// BotBlocksListResult is the result of GET /api/bot-blocks: active per-bot
// "user blocked the bot" flags (crm_bot_blocks).
//
// Kind повторяет фильтр запроса: nil у полного списка (ListBotBlocks) и у CRM
// без этого ключа. Count - длина UserIDs, то есть с учётом фильтра; разбивка
// по всем флагам бота лежит в Counts (нули у CRM без этого ключа).
type BotBlocksListResult struct {
	BotID   int64
	Kind    *string
	UserIDs []int64
	Count   int64
	Counts  BotBlockCounts
}

// BotBlockUnblockResult is the result of POST /api/bot-blocks/unblock.
// Removed=false means the flag did not exist — that is also a success.
type BotBlockUnblockResult struct {
	Removed bool
}

// BotBlockReportResult is the result of POST /api/bot-blocks/report.
//
// Ignored=true: CRM не ведёт флаги для этого бота и сигнал отбросила. Added
// при этом false, как и у повторного сигнала, поэтому без Ignored "флаг уже
// стоит" не отличить от "флаг никогда не встанет".
type BotBlockReportResult struct {
	Added   bool
	Ignored bool
}

// ListBotBlocks fetches every user id flagged as "blocked the bot" for the
// given bot (crm_bot_blocks). Intended usage: prime a local cache at startup
// and refresh it hourly; clear entries via UnblockBotBlock when the user
// shows any activity.
func (c *Client) ListBotBlocks(ctx context.Context, botID int64) (*BotBlocksListResult, error) {
	return c.listBotBlocks(ctx, botID, "")
}

// ListBotBlocksByKind - тот же список, что ListBotBlocks, но одной категории:
// BotBlockKindBlocked или BotBlockKindUnreachable.
//
// Категория проверяется здесь, а не на сервере: на неизвестное значение CRM
// отвечает успехом с пустым списком, и опечатка выглядела бы как "флагов нет".
func (c *Client) ListBotBlocksByKind(ctx context.Context, botID int64, kind string) (*BotBlocksListResult, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != BotBlockKindBlocked && kind != BotBlockKindUnreachable {
		return nil, &ValidationError{Message: "kind must be 'blocked' or 'unreachable'"}
	}
	return c.listBotBlocks(ctx, botID, kind)
}

func (c *Client) listBotBlocks(ctx context.Context, botID int64, kind string) (*BotBlocksListResult, error) {
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}

	query := map[string]string{"bot_id": fmt.Sprintf("%d", botID)}
	if kind != "" {
		query["kind"] = kind
	}

	var raw struct {
		BotID   int64          `json:"bot_id"`
		Kind    *string        `json:"kind"`
		UserIDs []int64        `json:"user_ids"`
		Count   int64          `json:"count"`
		Counts  BotBlockCounts `json:"counts"`
	}

	if err := c.get(ctx, "/api/bot-blocks", query, true, &raw); err != nil {
		return nil, err
	}

	return &BotBlocksListResult{
		BotID:   raw.BotID,
		Kind:    raw.Kind,
		UserIDs: raw.UserIDs,
		Count:   raw.Count,
		Counts:  raw.Counts,
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
		Added   bool `json:"added"`
		Ignored bool `json:"ignored"`
	}

	if err := c.post(ctx, "/api/bot-blocks/report", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &BotBlockReportResult{Added: raw.Added, Ignored: raw.Ignored}, nil
}
