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
