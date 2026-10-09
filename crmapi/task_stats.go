package crmapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// TaskStats возвращает подробную статистику задачи лайкера или чаттинга:
// итоги, разбивку по аккаунтам, проблемы, проблемные чаты и динамику по дням.
// days<=0 - окно по умолчанию CRM. Чужая задача - APIError со Status 404.
//
// GET /api/tasks/stats?task_type=&task_id=&user_id=&days=
func (c *Client) TaskStats(ctx context.Context, taskType string, taskID int64, userID int64, days int) (*TaskStats, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if taskID <= 0 {
		return nil, &ValidationError{Message: "task_id must be a positive integer"}
	}
	taskType = strings.TrimSpace(taskType)
	if taskType == "" {
		return nil, &ValidationError{Message: "task_type must not be empty"}
	}

	query := map[string]string{
		"task_type": taskType,
		"task_id":   fmt.Sprintf("%d", taskID),
		"user_id":   fmt.Sprintf("%d", userID),
	}
	if days > 0 {
		query["days"] = fmt.Sprintf("%d", days)
	}

	var res TaskStats
	if err := c.get(ctx, "/api/tasks/stats", query, true, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ActivityDaily возвращает лайки, списанные лайки и сообщения чаттинга
// клиента по дням (МСК, по возрастанию даты, дни без событий нулями).
// days<=0 - окно по умолчанию CRM.
//
// GET /api/profile/activity-daily?user_id=&days=
func (c *Client) ActivityDaily(ctx context.Context, userID int64, days int) (*ActivityDaily, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}
	if days > 0 {
		query["days"] = fmt.Sprintf("%d", days)
	}

	var res ActivityDaily
	if err := c.get(ctx, "/api/profile/activity-daily", query, true, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Schedules возвращает расписания (планировщик) задач клиента.
//
// GET /api/schedules?user_id=
func (c *Client) Schedules(ctx context.Context, userID int64) ([]Schedule, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var res []Schedule
	if err := c.get(ctx, "/api/schedules", query, true, &res); err != nil {
		return nil, err
	}
	if res == nil {
		res = []Schedule{}
	}
	return res, nil
}

// ScheduleRuns возвращает историю запусков расписания. limit<=0 - значение
// по умолчанию CRM. Чужое расписание - APIError со Status 404.
//
// GET /api/schedules/{id}/runs?user_id=&limit=
func (c *Client) ScheduleRuns(ctx context.Context, scheduleID int64, userID int64, limit int) (*ScheduleRuns, error) {
	if scheduleID <= 0 {
		return nil, &ValidationError{Message: "schedule_id must be a positive integer"}
	}
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}
	if limit > 0 {
		query["limit"] = fmt.Sprintf("%d", limit)
	}

	var res ScheduleRuns
	if err := c.get(ctx, fmt.Sprintf("/api/schedules/%d/runs", scheduleID), query, true, &res); err != nil {
		return nil, err
	}
	if res.Runs == nil {
		res.Runs = []ScheduleRun{}
	}
	return &res, nil
}

// ScheduleLog скачивает лог расписания файлом, как TasksLog. Имя файла берётся
// из Content-Disposition (пустое, если заголовка нет). Нет лога или чужое
// расписание - APIError со Status 404.
//
// GET /api/schedules/{id}/log?user_id=
func (c *Client) ScheduleLog(ctx context.Context, scheduleID int64, userID int64) (string, []byte, error) {
	if scheduleID <= 0 {
		return "", nil, &ValidationError{Message: "schedule_id must be a positive integer"}
	}
	if userID <= 0 {
		return "", nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	content, headers, err := c.getFile(ctx, fmt.Sprintf("/api/schedules/%d/log", scheduleID), query, true)
	if err != nil {
		return "", nil, err
	}

	filename := utils.ParseContentDispositionFilename(headers.Get("Content-Disposition"))
	return filename, content, nil
}
