package crmapi

type TaskListItem struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type TaskInfoResult struct {
	Text string `json:"text"`
}

type TaskLogResult struct {
	Filename *string `json:"filename,omitempty"`
	Content  []byte  `json:"content"`
}

type ActiveTasksResult struct {
	Text string `json:"text"`
}
