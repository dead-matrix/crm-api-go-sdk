package crmapi

import (
	"context"
	"fmt"
)

// AccountsCount возвращает число аккаунтов пользователя (дешёвый COUNT на стороне
// CRM). Мессенджер зовёт его для удалённых (removedOnly=true) ПЕРЕД их загрузкой:
// если удалённых слишком много, предлагается выбрать период вместо выгрузки всего
// разом. includeRemoved учитывать ли удалённые в общем подсчёте; removedOnly=true
// считает ТОЛЬКО удалённые (перекрывает includeRemoved).
func (c *Client) AccountsCount(ctx context.Context, userID int64, includeRemoved bool, removedOnly bool) (int64, error) {
	if userID <= 0 {
		return 0, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var raw struct {
		Total int64 `json:"total"`
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}
	if includeRemoved {
		query["include_removed"] = "true"
	}
	if removedOnly {
		query["removed_only"] = "true"
	}

	if err := c.get(ctx, "/api/accounts/count", query, true, &raw); err != nil {
		return 0, err
	}
	return raw.Total, nil
}

// AccountsList возвращает аккаунты пользователя. includeRemoved=true просит
// CRM отдать и удалённые строки (поле Removed=true у них); false сохраняет
// прежнее поведение - только живые аккаунты. removedOnly=true отдаёт ТОЛЬКО
// удалённые (перекрывает includeRemoved) - мессенджер грузит их отдельным
// запросом по галочке «Показывать удаленные». days>0 ограничивает выборку
// аккаунтами, впервые загруженными за последние N дней (по first_load) - для
// удалённых у «тяжёлых» пользователей; days=0 - без ограничения.
func (c *Client) AccountsList(ctx context.Context, userID int64, includeRemoved bool, removedOnly bool, days int) ([]AccountItem, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var raw []struct {
		SessionName *string `json:"session_name"`
		Valid       bool    `json:"valid"`
		SpamBlock   bool    `json:"spam_block"`
		IsConnected bool    `json:"is_connected"`
		Location    *string `json:"location"`
		FullName    *string `json:"full_name"`
		Username    *string `json:"username"`
		Phone       *string `json:"phone"`
		Premium     bool    `json:"premium"`
		Commented   struct {
			Day   int64 `json:"day"`
			Total int64 `json:"total"`
		} `json:"commented"`
		Invited struct {
			Day   int64 `json:"day"`
			Total int64 `json:"total"`
		} `json:"invited"`
		Stories struct {
			Day   int64 `json:"day"`
			Total int64 `json:"total"`
		} `json:"stories"`
		Tagged struct {
			Day   int64 `json:"day"`
			Total int64 `json:"total"`
		} `json:"tagged"`
		Views struct {
			Day   int64 `json:"day"`
			Total int64 `json:"total"`
		} `json:"views"`
		Reactions struct {
			Day   int64 `json:"day"`
			Total int64 `json:"total"`
		} `json:"reactions"`
		FirstLoad *string `json:"first_load"`
		Removed   bool    `json:"removed"`
		Proxy     *string `json:"proxy"`
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}
	if includeRemoved {
		query["include_removed"] = "true"
	}
	if removedOnly {
		query["removed_only"] = "true"
	}
	if days > 0 {
		query["days"] = fmt.Sprintf("%d", days)
	}

	if err := c.get(ctx, "/api/accounts/list", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]AccountItem, 0, len(raw))
	for _, a := range raw {
		items = append(items, AccountItem{
			SessionName: a.SessionName,
			Valid:       a.Valid,
			SpamBlock:   a.SpamBlock,
			IsConnected: a.IsConnected,
			Location:    a.Location,
			FullName:    a.FullName,
			Username:    a.Username,
			Phone:       a.Phone,
			Premium:     a.Premium,
			Commented: DayTotal{
				Day:   a.Commented.Day,
				Total: a.Commented.Total,
			},
			Invited: DayTotal{
				Day:   a.Invited.Day,
				Total: a.Invited.Total,
			},
			Stories: DayTotal{
				Day:   a.Stories.Day,
				Total: a.Stories.Total,
			},
			Tagged: DayTotal{
				Day:   a.Tagged.Day,
				Total: a.Tagged.Total,
			},
			Views: DayTotal{
				Day:   a.Views.Day,
				Total: a.Views.Total,
			},
			Reactions: DayTotal{
				Day:   a.Reactions.Day,
				Total: a.Reactions.Total,
			},
			FirstLoad: a.FirstLoad,
			Removed:   a.Removed,
			Proxy:     a.Proxy,
		})
	}

	return items, nil
}
