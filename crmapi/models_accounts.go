package crmapi

import "time"

type DayTotal struct {
	Day   int64 `json:"day"`
	Total int64 `json:"total"`
}

type AccountItem struct {
	SessionName *string  `json:"session_name,omitempty"`
	Valid       bool     `json:"valid"`
	SpamBlock   bool     `json:"spam_block"`
	IsConnected bool     `json:"is_connected"`
	Location    *string  `json:"location,omitempty"`
	FullName    *string  `json:"full_name,omitempty"`
	Username    *string  `json:"username,omitempty"`
	Phone       *string  `json:"phone,omitempty"`
	Premium     bool     `json:"premium"`
	Commented   DayTotal `json:"commented"`
	Invited     DayTotal `json:"invited"`
	Stories     DayTotal `json:"stories"`
	Tagged      DayTotal `json:"tagged"`
	Views       DayTotal `json:"views"`
	Reactions   DayTotal `json:"reactions"`
}

type ProfileStatistics struct {
	Subscriber        bool   `json:"subscriber"`
	AllAccountsAmount int64  `json:"all_accounts_amount"`
	AllInvited        int64  `json:"all_invited"`
	AllCommented      int64  `json:"all_commented"`
	AllStories        int64  `json:"all_stories"`
	AllTagged         int64  `json:"all_tagged"`
	AllViews          int64  `json:"all_views"`
	AllReactions      int64  `json:"all_reactions"`
	Tasks             int64  `json:"tasks"`
	Valid             int64  `json:"valid"`
	Work              int64  `json:"work"`
	Invalid           int64  `json:"invalid"`
	SpamBlock         int64  `json:"spam_block"`
	Invited           int64  `json:"invited"`
	Commented         int64  `json:"commented"`
	Stories           int64  `json:"stories"`
	Tagged            int64  `json:"tagged"`
	Views             int64  `json:"views"`
	Reactions         int64  `json:"reactions"`
	Quota             *int64 `json:"quota,omitempty"`
}

type PosterAccount struct {
	TelegramID     *int64     `json:"telegram_id,omitempty"`
	Valid          bool       `json:"valid"`
	IsConnected    bool       `json:"is_connected"`
	LastConnection *time.Time `json:"last_connection,omitempty"`
	Premium        bool       `json:"premium"`
	FullName       *string    `json:"full_name,omitempty"`
	Username       *string    `json:"username,omitempty"`
	Location       *string    `json:"location,omitempty"`
}

type PosterSubscription struct {
	Active    bool       `json:"active"`
	Access    any        `json:"access,omitempty"`
	AccessEnd *time.Time `json:"access_end,omitempty"`
}

type Bot3Summary struct {
	Subscription PosterSubscription `json:"subscription"`
	Account      *PosterAccount     `json:"account,omitempty"`
	Tasks        map[string]int64   `json:"tasks"`
}
