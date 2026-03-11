package crmapi

type PromptUpdateResult struct {
	Reset   *bool   `json:"reset,omitempty"`
	Message *string `json:"message,omitempty"`
	Updated *bool   `json:"updated,omitempty"`
	Created *bool   `json:"created,omitempty"`
}
