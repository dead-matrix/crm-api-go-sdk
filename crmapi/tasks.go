package crmapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) GetActiveTasks(ctx context.Context, userID int64) (*ActiveTasksResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw struct {
		Text string `json:"text"`
	}

	if err := c.get(ctx, "/api/tasks/active", query, true, &raw); err != nil {
		return nil, err
	}

	return &ActiveTasksResult{
		Text: raw.Text,
	}, nil
}

func (c *Client) TasksTypes(ctx context.Context, userID int64, botID int64) (map[string]string, error) {
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

	var raw map[string]string
	if err := c.get(ctx, "/api/tasks/types", query, true, &raw); err != nil {
		return nil, err
	}

	if raw == nil {
		return map[string]string{}, nil
	}

	return raw, nil
}

func (c *Client) TasksList(ctx context.Context, userID int64, botID int64, taskType *string, limit int64, offset int64) ([]TaskListItem, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if limit <= 0 {
		return nil, &ValidationError{Message: "limit must be a positive integer"}
	}
	if offset < 0 {
		return nil, &ValidationError{Message: "offset must be a non-negative integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
		"bot_id":  fmt.Sprintf("%d", botID),
		"limit":   fmt.Sprintf("%d", limit),
		"offset":  fmt.Sprintf("%d", offset),
	}

	if taskType != nil {
		trimmed := strings.TrimSpace(*taskType)
		if trimmed != "" {
			query["task_type"] = trimmed
		}
	}

	var raw []struct {
		ID   int64  `json:"id"`
		Text string `json:"text"`
	}

	if err := c.get(ctx, "/api/tasks/list", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]TaskListItem, 0, len(raw))
	for _, i := range raw {
		items = append(items, TaskListItem{
			ID:   i.ID,
			Text: i.Text,
		})
	}

	return items, nil
}

func (c *Client) TasksInfo(ctx context.Context, userID int64, botID int64, taskType string, taskID int64) (*TaskInfoResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if taskID <= 0 {
		return nil, &ValidationError{Message: "task_id must be a positive integer"}
	}

	taskType = strings.TrimSpace(taskType)
	if taskType == "" {
		return nil, &ValidationError{Message: "task_type must not be empty"}
	}

	query := map[string]string{
		"user_id":   fmt.Sprintf("%d", userID),
		"bot_id":    fmt.Sprintf("%d", botID),
		"task_type": taskType,
		"task_id":   fmt.Sprintf("%d", taskID),
	}

	var raw struct {
		Text string `json:"text"`
	}

	if err := c.get(ctx, "/api/tasks/info", query, true, &raw); err != nil {
		return nil, err
	}

	return &TaskInfoResult{
		Text: raw.Text,
	}, nil
}

func (c *Client) TasksLog(ctx context.Context, userID int64, taskType string, taskID int64, botID int64) (*TaskLogResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if botID <= 0 {
		return nil, &ValidationError{Message: "bot_id must be a positive integer"}
	}
	if taskID <= 0 {
		return nil, &ValidationError{Message: "task_id must be a positive integer"}
	}

	taskType = strings.TrimSpace(taskType)
	if taskType == "" {
		return nil, &ValidationError{Message: "task_type must not be empty"}
	}

	query := map[string]string{
		"user_id":   fmt.Sprintf("%d", userID),
		"task_type": taskType,
		"task_id":   fmt.Sprintf("%d", taskID),
		"bot_id":    fmt.Sprintf("%d", botID),
	}

	content, headers, err := c.getFile(ctx, "/api/tasks/log", query, true)
	if err != nil {
		return nil, err
	}

	filename := utils.ParseContentDispositionFilename(headers.Get("Content-Disposition"))
	var filenamePtr *string
	if filename != "" {
		filenamePtr = &filename
	}

	return &TaskLogResult{
		Filename: filenamePtr,
		Content:  content,
	}, nil
}
