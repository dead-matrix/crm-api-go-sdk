package crmapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) ListUserNotes(ctx context.Context, userID int64) ([]NoteItem, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	var raw []struct {
		Staff struct {
			ID   *int64  `json:"id"`
			Name *string `json:"name"`
		} `json:"staff"`
		Date *string `json:"date"`
		Text string  `json:"text"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/users/%d/notes", userID), nil, true, &raw); err != nil {
		return nil, err
	}

	items := make([]NoteItem, 0, len(raw))
	for _, n := range raw {
		var dt *time.Time
		if n.Date != nil {
			dt = utils.ParseTime(*n.Date)
		}

		items = append(items, NoteItem{
			Staff: NoteStaff{
				ID:   n.Staff.ID,
				Name: n.Staff.Name,
			},
			Date: dt,
			Text: n.Text,
		})
	}

	return items, nil
}

func (c *Client) CreateUserNote(ctx context.Context, userID int64, text string) (*NoteItem, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, &ValidationError{Message: "text must not be empty"}
	}

	body := map[string]string{
		"text": text,
	}

	var raw struct {
		Staff struct {
			ID   *int64  `json:"id"`
			Name *string `json:"name"`
		} `json:"staff"`
		Date *string `json:"date"`
		Text string  `json:"text"`
	}

	if err := c.post(ctx, fmt.Sprintf("/api/users/%d/notes", userID), nil, true, body, &raw); err != nil {
		return nil, err
	}

	var dt *time.Time
	if raw.Date != nil {
		dt = utils.ParseTime(*raw.Date)
	}

	return &NoteItem{
		Staff: NoteStaff{
			ID:   raw.Staff.ID,
			Name: raw.Staff.Name,
		},
		Date: dt,
		Text: raw.Text,
	}, nil
}
