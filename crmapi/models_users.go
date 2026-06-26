package crmapi

import "time"

type UserBotInfo struct {
	BotID      int64      `json:"bot_id"`
	BotName    string     `json:"bot_name"`
	Registered *time.Time `json:"registered,omitempty"`
	Refer      *string    `json:"refer,omitempty"`
	Access     any        `json:"access,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
	// Subscription-freeze fields, populated when the latest access row for this
	// bot is action="freeze". Frozen flags the state; FrozenAt is the moment of
	// freezing; FrozenExpiry is the per-feature end-date snapshot {feature:
	// end_iso} captured at freeze time (remaining days per feature =
	// FrozenExpiry[feat] − FrozenAt, since the frozen period doesn't elapse).
	Frozen       bool              `json:"frozen,omitempty"`
	FrozenAt     *time.Time        `json:"frozen_at,omitempty"`
	FrozenExpiry map[string]string `json:"frozen_expiry,omitempty"`
}

type GetUserResult struct {
	UserID   int64         `json:"user_id"`
	FullName *string       `json:"full_name,omitempty"`
	Username *string       `json:"username,omitempty"`
	Status   *string       `json:"status,omitempty"`
	BotsInfo []UserBotInfo `json:"bots_info"`
}

// CreateUserResult is the result of POST /api/users (idempotent).
//
// If a registration for (UserID, BotID) already existed, Created is false and
// the rest of the fields contain the existing record. Otherwise Created is
// true and the fields describe the newly created record.
//
// FullName is nullable: on the idempotent path the server returns
// `user.full_name` directly from the database row, which is technically a
// nullable column. New registrations always have a non-nil FullName.
type CreateUserResult struct {
	Created  bool       `json:"created"`
	UserID   int64      `json:"user_id"`
	FullName *string    `json:"full_name,omitempty"`
	Username *string    `json:"username,omitempty"`
	BotID    int64      `json:"bot_id"`
	Refer    *string    `json:"refer,omitempty"`
	DateReg  *time.Time `json:"date_reg,omitempty"`
}

// ListUserItem is one element of GET /api/users?bot_id=... response.
//
// FullName is nullable: the server returns `user.full_name` directly from
// the database row, which is technically a nullable column.
type ListUserItem struct {
	UserID     int64      `json:"user_id"`
	FullName   *string    `json:"full_name,omitempty"`
	Username   *string    `json:"username,omitempty"`
	DateReg    *time.Time `json:"date_reg,omitempty"`
	Refer      *string    `json:"refer,omitempty"`
	Restricted bool       `json:"restricted"`
}

// ListUsersResult is the response of GET /api/users?bot_id=...&limit=...&offset=...
type ListUsersResult struct {
	BotID  int64          `json:"bot_id"`
	Limit  int64          `json:"limit"`
	Offset int64          `json:"offset"`
	Count  int64          `json:"count"`
	Items  []ListUserItem `json:"items"`
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
