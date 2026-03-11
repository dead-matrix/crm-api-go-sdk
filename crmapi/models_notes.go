package crmapi

import "time"

type NoteStaff struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type NoteItem struct {
	Staff NoteStaff  `json:"staff"`
	Date  *time.Time `json:"date,omitempty"`
	Text  string     `json:"text"`
}
