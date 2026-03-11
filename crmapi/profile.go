package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) ProfileStatistics(ctx context.Context, userID int64) (*ProfileStatistics, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw struct {
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
		Quota             *int64 `json:"quota"`
	}

	if err := c.get(ctx, "/api/profile/statistics", query, true, &raw); err != nil {
		return nil, err
	}

	return &ProfileStatistics{
		Subscriber:        raw.Subscriber,
		AllAccountsAmount: raw.AllAccountsAmount,
		AllInvited:        raw.AllInvited,
		AllCommented:      raw.AllCommented,
		AllStories:        raw.AllStories,
		AllTagged:         raw.AllTagged,
		AllViews:          raw.AllViews,
		AllReactions:      raw.AllReactions,
		Tasks:             raw.Tasks,
		Valid:             raw.Valid,
		Work:              raw.Work,
		Invalid:           raw.Invalid,
		SpamBlock:         raw.SpamBlock,
		Invited:           raw.Invited,
		Commented:         raw.Commented,
		Stories:           raw.Stories,
		Tagged:            raw.Tagged,
		Views:             raw.Views,
		Reactions:         raw.Reactions,
		Quota:             raw.Quota,
	}, nil
}

func (c *Client) ProfileBot3Summary(ctx context.Context, userID int64) (*Bot3Summary, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw struct {
		Subscription struct {
			Active    bool    `json:"active"`
			Access    any     `json:"access"`
			AccessEnd *string `json:"access_end"`
		} `json:"subscription"`
		Account *struct {
			TelegramID     *int64  `json:"telegram_id"`
			Valid          bool    `json:"valid"`
			IsConnected    bool    `json:"is_connected"`
			LastConnection *string `json:"last_connection"`
			Premium        bool    `json:"premium"`
			FullName       *string `json:"full_name"`
			Username       *string `json:"username"`
			Location       *string `json:"location"`
		} `json:"account"`
		Tasks map[string]int64 `json:"tasks"`
	}

	if err := c.get(ctx, "/api/profile/bot3/summary", query, true, &raw); err != nil {
		return nil, err
	}

	var accessEnd *time.Time
	if raw.Subscription.AccessEnd != nil {
		accessEnd = utils.ParseTime(*raw.Subscription.AccessEnd)
	}

	subscription := PosterSubscription{
		Active:    raw.Subscription.Active,
		Access:    raw.Subscription.Access,
		AccessEnd: accessEnd,
	}

	var account *PosterAccount
	if raw.Account != nil {
		var lastConnection *time.Time
		if raw.Account.LastConnection != nil {
			lastConnection = utils.ParseTime(*raw.Account.LastConnection)
		}

		account = &PosterAccount{
			TelegramID:     raw.Account.TelegramID,
			Valid:          raw.Account.Valid,
			IsConnected:    raw.Account.IsConnected,
			LastConnection: lastConnection,
			Premium:        raw.Account.Premium,
			FullName:       raw.Account.FullName,
			Username:       raw.Account.Username,
			Location:       raw.Account.Location,
		}
	}

	tasks := raw.Tasks
	if tasks == nil {
		tasks = map[string]int64{}
	}

	return &Bot3Summary{
		Subscription: subscription,
		Account:      account,
		Tasks:        tasks,
	}, nil
}
