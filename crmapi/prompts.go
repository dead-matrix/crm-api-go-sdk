package crmapi

import (
	"context"
	"fmt"
)

func (c *Client) PromptGet(ctx context.Context, userID int64) (*string, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var prompt *string
	if err := c.get(ctx, "/api/prompt", query, true, &prompt); err != nil {
		return nil, err
	}

	return prompt, nil
}

func (c *Client) PromptUpdate(ctx context.Context, userID int64, prompt string) (*PromptUpdateResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	body := map[string]string{
		"prompt": prompt,
	}

	var raw struct {
		Reset   *bool   `json:"reset"`
		Message *string `json:"message"`
		Updated *bool   `json:"updated"`
		Created *bool   `json:"created"`
	}

	if err := c.post(ctx, "/api/prompt/update", query, true, body, &raw); err != nil {
		return nil, err
	}

	return &PromptUpdateResult{
		Reset:   raw.Reset,
		Message: raw.Message,
		Updated: raw.Updated,
		Created: raw.Created,
	}, nil
}
