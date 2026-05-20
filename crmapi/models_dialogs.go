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

// ChangeStatusResult — результат POST /api/dialogs/status.
//
// Status — название нового статуса. nil означает, что статус снят
// (отвечает Python SDK: status: str | None). Пустая строка от сервера
// (теоретически невозможна, но защита от мусора) тоже остаётся отдельным
// значением и не сливается с nil.
type ChangeStatusResult struct {
	Status *string `json:"status"`
}

// DialogSearchItem — единичный элемент ответа GET /api/dialogs/{dept}/search.
//
// Status/StatusColor nullable: сервер шлёт `status_title or default_status_title`,
// и оба могут быть nil (диалог без выставленного статуса в департаменте,
// у которого не задан default_status). Паритет с DialogItem.
type DialogSearchItem struct {
	UserID                int64   `json:"user_id"`
	FullName              *string `json:"full_name,omitempty"`
	HasActiveSubscription bool    `json:"has_active_subscription"`
	Status                *string `json:"status,omitempty"`
	StatusColor           *string `json:"status_color,omitempty"`
}

type DialogSearchResult struct {
	Dialogs []DialogSearchItem `json:"dialogs"`
	Limit   int64              `json:"limit"`
	Offset  int64              `json:"offset"`
}
