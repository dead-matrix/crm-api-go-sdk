package crmapi

type StaffInfo struct {
	Name     *string        `json:"name,omitempty"`
	Role     *string        `json:"role,omitempty"`
	IsActive bool           `json:"is_active"`
	Access   map[string]any `json:"access,omitempty"`
}

// StaffListItem — короткая запись сотрудника из GET /api/staff/list (user_id > 1000).
type StaffListItem struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}
