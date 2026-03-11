package crmapi

type StaffInfo struct {
	Name     *string        `json:"name,omitempty"`
	Role     *string        `json:"role,omitempty"`
	IsActive bool           `json:"is_active"`
	Access   map[string]any `json:"access,omitempty"`
}
