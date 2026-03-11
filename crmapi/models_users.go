package crmapi

import "time"

type UserBotInfo struct {
	BotID      int64      `json:"bot_id"`
	BotName    string     `json:"bot_name"`
	Registered *time.Time `json:"registered,omitempty"`
	Refer      *string    `json:"refer,omitempty"`
	Access     any        `json:"access,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
}

type GetUserResult struct {
	UserID   int64         `json:"user_id"`
	FullName *string       `json:"full_name,omitempty"`
	Username *string       `json:"username,omitempty"`
	Status   *string       `json:"status,omitempty"`
	BotsInfo []UserBotInfo `json:"bots_info"`
}

type CreateUserResult struct {
	Created bool `json:"created"`
}

type UpdateUserResult struct {
	UserID   int64   `json:"user_id"`
	FullName string  `json:"full_name"`
	Username *string `json:"username,omitempty"`
}

type AddAccessResult struct {
	Created    bool       `json:"created"`
	ID         *int64     `json:"id,omitempty"`
	UserID     int64      `json:"user_id"`
	BotID      int64      `json:"bot_id"`
	Action     string     `json:"action"`
	ActionDate *time.Time `json:"action_date,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
}

type ExtendAccessResult struct {
	UserID    int64      `json:"user_id"`
	AccessEnd *time.Time `json:"access_end,omitempty"`
}

type ExtendAILimitResult struct {
	PreviousAILimit int64 `json:"previous_ai_limit"`
	AILimit         int64 `json:"ai_limit"`
}
