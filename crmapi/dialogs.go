package crmapi

import (
	"context"
	"fmt"
	"strings"
)

func (c *Client) GetDialogs(ctx context.Context, department string) ([]DialogItem, error) {
	department = strings.TrimSpace(department)
	if department == "" {
		return nil, &ValidationError{Message: "department must not be empty"}
	}

	var raw struct {
		Dialogs []struct {
			UserID                int64   `json:"user_id"`
			FullName              *string `json:"full_name"`
			HasActiveSubscription bool    `json:"has_active_subscription"`
			Status                *string `json:"status"`
			StatusColor           *string `json:"status_color"`
		} `json:"dialogs"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/dialogs/%s", department), nil, true, &raw); err != nil {
		return nil, err
	}

	items := make([]DialogItem, 0, len(raw.Dialogs))
	for _, r := range raw.Dialogs {
		items = append(items, DialogItem{
			UserID:                r.UserID,
			FullName:              r.FullName,
			HasActiveSubscription: r.HasActiveSubscription,
			Status:                r.Status,
			StatusColor:           r.StatusColor,
		})
	}

	return items, nil
}

func (c *Client) GetStatuses(ctx context.Context, departmentID int64) (*StatusesResult, error) {
	if departmentID <= 0 {
		return nil, &ValidationError{Message: "department_id must be a positive integer"}
	}

	var raw struct {
		DepartmentID    int64  `json:"department_id"`
		DefaultStatusID *int64 `json:"default_status_id"`
		Statuses        []struct {
			ID    int64   `json:"id"`
			Title string  `json:"title"`
			Color *string `json:"color"`
		} `json:"statuses"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/dialogs/statuses/%d", departmentID), nil, true, &raw); err != nil {
		return nil, err
	}

	statuses := make([]StatusItem, 0, len(raw.Statuses))
	for _, s := range raw.Statuses {
		statuses = append(statuses, StatusItem{
			ID:    s.ID,
			Title: s.Title,
			Color: s.Color,
		})
	}

	return &StatusesResult{
		DepartmentID:    raw.DepartmentID,
		DefaultStatusID: raw.DefaultStatusID,
		Statuses:        statuses,
	}, nil
}

func (c *Client) ChangeDialogStatus(ctx context.Context, userID int64, statusID int64) (*ChangeStatusResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if statusID <= 0 {
		return nil, &ValidationError{Message: "status_id must be a positive integer"}
	}

	body := map[string]int64{
		"user_id":   userID,
		"status_id": statusID,
	}

	var raw struct {
		Status string `json:"status"`
	}

	if err := c.post(ctx, "/api/dialogs/status", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &ChangeStatusResult{
		Status: raw.Status,
	}, nil
}

// ClearDialogStatus снимает активный статус диалога. POST /api/dialogs/status
// получает payload с явным status_id=null — это требуется текущей валидацией
// CRM API: payload без поля status_id отвергается с 422 "Validation error".
// Сервер удаляет запись DialogStatus для диалога в его текущем департаменте
// и возвращает status=null, который нормализуется в пустую строку в
// ChangeStatusResult.Status.
func (c *Client) ClearDialogStatus(ctx context.Context, userID int64) (*ChangeStatusResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	body := map[string]any{
		"user_id":   userID,
		"status_id": nil,
	}

	var raw struct {
		Status *string `json:"status"`
	}

	if err := c.post(ctx, "/api/dialogs/status", nil, true, body, &raw); err != nil {
		return nil, err
	}

	status := ""
	if raw.Status != nil {
		status = *raw.Status
	}
	return &ChangeStatusResult{Status: status}, nil
}

func (c *Client) TransferDialog(ctx context.Context, userID int64, toDepartment string) (*TransferDialogResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	toDepartment = strings.TrimSpace(toDepartment)
	if toDepartment == "" {
		return nil, &ValidationError{Message: "to_department must not be empty"}
	}

	body := map[string]any{
		"user_id":       userID,
		"to_department": toDepartment,
	}

	var raw struct {
		Transferred bool `json:"transferred"`
	}

	if err := c.post(ctx, "/api/dialogs/transfer", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &TransferDialogResult{
		Transferred: raw.Transferred,
	}, nil
}

func (c *Client) SearchDialogs(ctx context.Context, department string, q string, offset int64) (*DialogSearchResult, error) {
	department = strings.TrimSpace(department)
	if department == "" {
		return nil, &ValidationError{Message: "department must not be empty"}
	}

	q = strings.TrimSpace(q)
	if q == "" {
		return nil, &ValidationError{Message: "q must not be empty"}
	}

	if offset < 0 {
		return nil, &ValidationError{Message: "offset must be a non-negative integer"}
	}

	query := map[string]string{
		"q":      q,
		"offset": fmt.Sprintf("%d", offset),
	}

	var raw struct {
		Dialogs []struct {
			UserID                int64   `json:"user_id"`
			FullName              *string `json:"full_name"`
			HasActiveSubscription bool    `json:"has_active_subscription"`
			Status                string  `json:"status"`
			StatusColor           string  `json:"status_color"`
		} `json:"dialogs"`
		Limit  int64 `json:"limit"`
		Offset int64 `json:"offset"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/dialogs/%s/search", department), query, true, &raw); err != nil {
		return nil, err
	}

	dialogs := make([]DialogSearchItem, 0, len(raw.Dialogs))
	for _, r := range raw.Dialogs {
		dialogs = append(dialogs, DialogSearchItem{
			UserID:                r.UserID,
			FullName:              r.FullName,
			HasActiveSubscription: r.HasActiveSubscription,
			Status:                r.Status,
			StatusColor:           r.StatusColor,
		})
	}

	return &DialogSearchResult{
		Dialogs: dialogs,
		Limit:   raw.Limit,
		Offset:  raw.Offset,
	}, nil
}
