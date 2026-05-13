package crmapi

import (
	"context"
)

// Sales-decks endpoints (POST /api/scripts/price, POST /api/scripts/tools).
// Phase 5 removed the legacy text-quick-reply trio (ScriptsList / ScriptsGet
// / ScriptsCreate) on the CRM-API side; the SDK methods are deleted in
// parallel. Quick replies now go through reply_templates.go.

func (c *Client) ScriptsPrice(ctx context.Context, options []int64) ([]PriceMediaItem, error) {
	if len(options) == 0 {
		return nil, &ValidationError{Message: "options must contain at least one element"}
	}
	if len(options) > 5 {
		return nil, &ValidationError{Message: "options must contain at most 5 elements"}
	}
	for _, opt := range options {
		if opt < 0 || opt > 4 {
			return nil, &ValidationError{Message: "each option must be between 0 and 4"}
		}
	}

	body := map[string]any{
		"options": options,
	}

	var raw []struct {
		Text  string   `json:"text"`
		Media []string `json:"media"`
	}

	if err := c.post(ctx, "/api/scripts/price", nil, true, body, &raw); err != nil {
		return nil, err
	}

	items := make([]PriceMediaItem, 0, len(raw))
	for _, item := range raw {
		items = append(items, PriceMediaItem{
			Text:  item.Text,
			Media: item.Media,
		})
	}

	return items, nil
}

func (c *Client) ScriptsTools(ctx context.Context, options []int64, botID int64) (*ToolsMediaResult, error) {
	if len(options) == 0 {
		return nil, &ValidationError{Message: "options must contain at least one element"}
	}
	if len(options) > 5 {
		return nil, &ValidationError{Message: "options must contain at most 5 elements"}
	}
	for _, opt := range options {
		if opt < 0 || opt > 4 {
			return nil, &ValidationError{Message: "each option must be between 0 and 4"}
		}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}

	body := map[string]any{
		"options": options,
		"bot_id":  botID,
	}

	var raw struct {
		Text  string `json:"text"`
		Media []struct {
			VideoURL string `json:"video_url"`
			FileID   string `json:"file_id"`
		} `json:"media"`
	}

	if err := c.post(ctx, "/api/scripts/tools", nil, true, body, &raw); err != nil {
		return nil, err
	}

	media := make([]ToolsMediaItem, 0, len(raw.Media))
	for _, item := range raw.Media {
		media = append(media, ToolsMediaItem{
			VideoURL: item.VideoURL,
			FileID:   item.FileID,
		})
	}

	return &ToolsMediaResult{
		Text:  raw.Text,
		Media: media,
	}, nil
}
