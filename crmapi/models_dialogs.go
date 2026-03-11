package crmapi

type DialogItem struct {
	UserID                int64   `json:"user_id"`
	FullName              *string `json:"full_name,omitempty"`
	HasActiveSubscription bool    `json:"has_active_subscription"`
	Status                *string `json:"status,omitempty"`
	StatusColor           *string `json:"status_color,omitempty"`
}

type TransferDialogResult struct {
	Transferred bool `json:"transferred"`
}

type StatusItem struct {
	ID    int64   `json:"id"`
	Title string  `json:"title"`
	Color *string `json:"color,omitempty"`
}

type StatusesResult struct {
	DepartmentID    int64        `json:"department_id"`
	DefaultStatusID *int64       `json:"default_status_id,omitempty"`
	Statuses        []StatusItem `json:"statuses"`
}

type ChangeStatusResult struct {
	Status string `json:"status"`
}

type DialogSearchItem struct {
	UserID                int64   `json:"user_id"`
	FullName              *string `json:"full_name,omitempty"`
	HasActiveSubscription bool    `json:"has_active_subscription"`
	Status                string  `json:"status"`
	StatusColor           string  `json:"status_color"`
}

type DialogSearchResult struct {
	Dialogs []DialogSearchItem `json:"dialogs"`
	Limit   int64              `json:"limit"`
	Offset  int64              `json:"offset"`
}
