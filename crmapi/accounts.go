package crmapi

import (
	"context"
	"fmt"
)

func (c *Client) AccountsList(ctx context.Context, userID int64) ([]AccountItem, error) {
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
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
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
		})
	}

	return items, nil
}
