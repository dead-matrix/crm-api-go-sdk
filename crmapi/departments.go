package crmapi

import "context"

func (c *Client) ListDepartments(ctx context.Context) ([]DepartmentItem, error) {
	var raw []struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Title string `json:"title"`
	}

	if err := c.get(ctx, "/api/departments", nil, true, &raw); err != nil {
		return nil, err
	}

	items := make([]DepartmentItem, 0, len(raw))
	for _, d := range raw {
		items = append(items, DepartmentItem{
			ID:    d.ID,
			Name:  d.Name,
			Title: d.Title,
		})
	}

	return items, nil
}
