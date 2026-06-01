package crmapi

import "context"

func (c *Client) GetStaff(ctx context.Context) (*StaffInfo, error) {
	var raw struct {
		Name     *string        `json:"name"`
		Role     *string        `json:"role"`
		IsActive bool           `json:"is_active"`
		Access   map[string]any `json:"access"`
	}

	if err := c.get(ctx, "/api/staff", nil, true, &raw); err != nil {
		return nil, err
	}

	return &StaffInfo{
		Name:     raw.Name,
		Role:     raw.Role,
		IsActive: raw.IsActive,
		Access:   raw.Access,
	}, nil
}

// ListStaff — короткая инфа обо всех сотрудниках с user_id > 1000
// (GET /api/staff/list). user_id <= 1000 — внутренние/системные сотрудники,
// сервер их не возвращает.
func (c *Client) ListStaff(ctx context.Context) ([]StaffListItem, error) {
	var raw []struct {
		UserID int64  `json:"user_id"`
		Name   string `json:"name"`
		Role   string `json:"role"`
	}

	if err := c.get(ctx, "/api/staff/list", nil, true, &raw); err != nil {
		return nil, err
	}

	items := make([]StaffListItem, 0, len(raw))
	for _, s := range raw {
		items = append(items, StaffListItem{
			UserID: s.UserID,
			Name:   s.Name,
			Role:   s.Role,
		})
	}

	return items, nil
}
