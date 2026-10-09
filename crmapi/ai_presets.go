package crmapi

import (
	"context"
	"fmt"
)

// AIPresets возвращает ИИ-пресеты клиента, «ИИ-настройки» профиля, ИИ-задачи
// с действующими настройками и фактические модели генераций по функциям
// comment и chatting. Только чтение.
//
// GET /api/ai-presets?user_id=
func (c *Client) AIPresets(ctx context.Context, userID int64) (*AIPresetsOverview, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var res AIPresetsOverview
	if err := c.get(ctx, "/api/ai-presets", query, true, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
