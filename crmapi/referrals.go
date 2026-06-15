package crmapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// withdrawMethods — допустимые методы вывода: 'wallet' (на кошелёк) |
// 'subscription' (в обмен на подписку).
var withdrawMethods = map[string]bool{"wallet": true, "subscription": true}

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

// ReferralsWithdrawRequest — заявка реферера на вывод всего доступного баланса.
//
// method: "wallet" | "subscription". Result.Status: "no_balance" |
// "already_pending" (бот показывает call.answer show_alert) | "created"
// (создана заявка + outbox-событие в мессенджер).
func (c *Client) ReferralsWithdrawRequest(ctx context.Context, userID int64, method string) (*WithdrawRequestResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	m := strings.ToLower(strings.TrimSpace(method))
	if !withdrawMethods[m] {
		return nil, &ValidationError{Message: "method must be 'wallet' or 'subscription'"}
	}

	body := map[string]any{"user_id": userID, "method": m}

	var raw struct {
		Status       string   `json:"status"`
		WithdrawalID *int64   `json:"withdrawal_id"`
		AmountUSD    *float64 `json:"amount_usd"`
		Method       *string  `json:"method"`
		AvailableUSD *float64 `json:"available_usd"`
	}

	if err := c.post(ctx, "/api/referrals/withdraw/request", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &WithdrawRequestResult{
		Status:       raw.Status,
		WithdrawalID: raw.WithdrawalID,
		AmountUSD:    raw.AmountUSD,
		Method:       raw.Method,
		AvailableUSD: raw.AvailableUSD,
	}, nil
}

// ReferralsWithdrawSettle — провести вывод: перевести amountMinor (USD-центы)
// из «доступно» в «выплачено» и зафиксировать method. Поддерживает частичный
// вывод. withdrawalID (опц.) — закрыть конкретную заявку; иначе закрывается
// открытая заявка пользователя либо создаётся запись вывода.
func (c *Client) ReferralsWithdrawSettle(ctx context.Context, userID int64, amountMinor int64, method string, withdrawalID *int64) (*WithdrawSettleResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	if amountMinor <= 0 {
		return nil, &ValidationError{Message: "amount_minor must be a positive integer (USD cents)"}
	}
	m := strings.ToLower(strings.TrimSpace(method))
	if !withdrawMethods[m] {
		return nil, &ValidationError{Message: "method must be 'wallet' or 'subscription'"}
	}

	body := map[string]any{"user_id": userID, "amount_minor": amountMinor, "method": m}
	if withdrawalID != nil {
		body["withdrawal_id"] = *withdrawalID
	}

	var raw struct {
		Status            string  `json:"status"`
		WithdrawalID      int64   `json:"withdrawal_id"`
		PaidUSD           float64 `json:"paid_usd"`
		AvailableAfterUSD float64 `json:"available_after_usd"`
		Method            string  `json:"method"`
	}

	if err := c.post(ctx, "/api/referrals/withdraw/settle", nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &WithdrawSettleResult{
		Status:            raw.Status,
		WithdrawalID:      raw.WithdrawalID,
		PaidUSD:           raw.PaidUSD,
		AvailableAfterUSD: raw.AvailableAfterUSD,
		Method:            raw.Method,
	}, nil
}
