package crmapi

import (
	"context"
	"fmt"
	"strings"
)

func (c *Client) ScriptsList(ctx context.Context) ([]ScriptItem, error) {
	var raw []struct {
		ID      int64   `json:"id"`
		Title   string  `json:"title"`
		Creator *string `json:"creator"`
	}

	if err := c.get(ctx, "/api/scripts", nil, true, &raw); err != nil {
		return nil, err
	}

	items := make([]ScriptItem, 0, len(raw))
	for _, item := range raw {
		items = append(items, ScriptItem{
			ID:      item.ID,
			Title:   item.Title,
			Creator: item.Creator,
		})
	}

	return items, nil
}

func (c *Client) ScriptsGet(ctx context.Context, scriptID int64) (*ScriptFull, error) {
	if scriptID <= 0 {
		return nil, &ValidationError{Message: "script_id must be a positive integer"}
	}

	var raw struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Text  string `json:"text"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/scripts/%d", scriptID), nil, true, &raw); err != nil {
		return nil, err
	}

	return &ScriptFull{
		ID:    raw.ID,
		Title: raw.Title,
		Text:  raw.Text,
	}, nil
}

func (c *Client) ScriptsCreate(ctx context.Context, title string, text string) (*ScriptFull, error) {
	title = strings.TrimSpace(title)
	text = strings.TrimSpace(text)

	if title == "" {
		return nil, &ValidationError{Message: "title must not be empty"}
	}
	if text == "" {
		return nil, &ValidationError{Message: "text must not be empty"}
	}
	if len(title) > 255 {
		return nil, &ValidationError{Message: "title must be at most 255 characters"}
	}
	if len(text) > 4096 {
		return nil, &ValidationError{Message: "text must be at most 4096 characters"}
	}

	body := map[string]string{
		"title": title,
		"text":  text,
	}

	var raw struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Text  string `json:"text"`
	}

	if err := c.post(ctx, "/api/scripts", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &ScriptFull{
		ID:    raw.ID,
		Title: raw.Title,
		Text:  raw.Text,
	}, nil
}

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
