package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) ReferralsInfo(ctx context.Context, userID int64) (*ReferralsInfoResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw struct {
		RefLink       string  `json:"ref_link"`
		Percent       int64   `json:"percent"`
		Registrations int64   `json:"registrations"`
		RefPayments   int64   `json:"ref_payments"`
		RefTotalSum   int64   `json:"ref_total_sum"`
		EarnedUSD     float64 `json:"earned_usd"`
		AvailableUSD  float64 `json:"available_usd"`
		Referrees     []struct {
			UserID           int64   `json:"user_id"`
			FullName         *string `json:"full_name"`
			Username         *string `json:"username"`
			PaymentsCount    int64   `json:"payments_count"`
			PaymentsSumMinor int64   `json:"payments_sum_minor"`
			Payments         []struct {
				Date          *string `json:"date"`
				AmountMinor   int64   `json:"amount_minor"`
				CommissionUSD float64 `json:"commission_usd"`
				Status        string  `json:"status"`
			} `json:"payments"`
		} `json:"referrees"`
	}

	if err := c.get(ctx, "/api/referrals/info", query, true, &raw); err != nil {
		return nil, err
	}

	referrees := make([]ReferreeInfo, 0, len(raw.Referrees))
	for _, r := range raw.Referrees {
		pays := make([]ReferreePayment, 0, len(r.Payments))
		for _, p := range r.Payments {
			var dt *time.Time
			if p.Date != nil {
				dt = utils.ParseTime(*p.Date)
			}

			pays = append(pays, ReferreePayment{
				Date:          dt,
				AmountMinor:   p.AmountMinor,
				CommissionUSD: p.CommissionUSD,
				Status:        p.Status,
			})
		}

		referrees = append(referrees, ReferreeInfo{
			UserID:           r.UserID,
			FullName:         r.FullName,
			Username:         r.Username,
			PaymentsCount:    r.PaymentsCount,
			PaymentsSumMinor: r.PaymentsSumMinor,
			Payments:         pays,
		})
	}

	return &ReferralsInfoResult{
		RefLink:       raw.RefLink,
		Percent:       raw.Percent,
		Registrations: raw.Registrations,
		RefPayments:   raw.RefPayments,
		RefTotalSum:   raw.RefTotalSum,
		EarnedUSD:     raw.EarnedUSD,
		AvailableUSD:  raw.AvailableUSD,
		Referrees:     referrees,
	}, nil
}
