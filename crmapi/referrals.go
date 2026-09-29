package crmapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// withdrawMethods: допустимые методы проведения вывода (settle): 'wallet'
// (на кошелёк) | 'subscription' (в обмен на подписку).
var withdrawMethods = map[string]bool{"wallet": true, "subscription": true}

// withdrawRequestMethods: методы, с которыми CRM принимает заявку на вывод.
// Оплата подписки балансом идёт не через заявку, поэтому subscription CRM
// отклоняет с 400; SDK отсекает его локально.
var withdrawRequestMethods = map[string]bool{"wallet": true}

// ReferralsInfo: реферальная сводка пользователя (GET /referrals/info).
// Result.Partner заполнен, только если CRM прислала ключ partner.
func (c *Client) ReferralsInfo(ctx context.Context, userID int64) (*ReferralsInfoResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}

	query := map[string]string{
		"user_id": fmt.Sprintf("%d", userID),
	}

	var raw struct {
		RefLink                  string  `json:"ref_link"`
		Percent                  int64   `json:"percent"`
		Registrations            int64   `json:"registrations"`
		RefPayments              int64   `json:"ref_payments"`
		RefTotalSum              int64   `json:"ref_total_sum"`
		EarnedUSD                float64 `json:"earned_usd"`
		AvailableUSD             float64 `json:"available_usd"`
		WithdrawnWalletUSD       float64 `json:"withdrawn_wallet_usd"`
		WithdrawnSubscriptionUSD float64 `json:"withdrawn_subscription_usd"`
		Referrees                []struct {
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
		Partner *rawReferralPartner `json:"partner"`
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
		RefLink:                  raw.RefLink,
		Percent:                  raw.Percent,
		Registrations:            raw.Registrations,
		RefPayments:              raw.RefPayments,
		RefTotalSum:              raw.RefTotalSum,
		EarnedUSD:                raw.EarnedUSD,
		AvailableUSD:             raw.AvailableUSD,
		WithdrawnWalletUSD:       raw.WithdrawnWalletUSD,
		WithdrawnSubscriptionUSD: raw.WithdrawnSubscriptionUSD,
		Referrees:                referrees,
		Partner:                  mapReferralPartner(raw.Partner),
	}, nil
}

type rawReferralPartner struct {
	RefLink                       string   `json:"ref_link"`
	RefBotLink                    string   `json:"ref_bot_link"`
	Percent                       int64    `json:"percent"`
	FirstPercent                  int64    `json:"first_percent"`
	RecurringPercent              int64    `json:"recurring_percent"`
	HoldDays                      int64    `json:"hold_days"`
	Registrations                 int64    `json:"registrations"`
	PaidReferralsCount            int64    `json:"paid_referrals_count"`
	PromoActivationsCount         int64    `json:"promo_activations_count"`
	ReferredPaymentsCount         int64    `json:"referred_payments_count"`
	ReferredTurnoverRUBKopecks    int64    `json:"referred_turnover_rub_kopecks"`
	EarnedUSDCents                int64    `json:"earned_usd_cents"`
	AvailableUSDCents             int64    `json:"available_usd_cents"`
	OnHoldUSDCents                int64    `json:"on_hold_usd_cents"`
	SpentOnSubscriptionsUSDCents  int64    `json:"spent_on_subscriptions_usd_cents"`
	WithdrawnWalletUSDCents       int64    `json:"withdrawn_wallet_usd_cents"`
	WithdrawnSubscriptionUSDCents int64    `json:"withdrawn_subscription_usd_cents"`
	MinWithdrawalUSDCents         int64    `json:"min_withdrawal_usd_cents"`
	WithdrawMethods               []string `json:"withdraw_methods"`
	PromoCodes                    []struct {
		Code string `json:"code"`
		Kind string `json:"kind"`
	} `json:"promo_codes"`
	PendingWithdrawal *struct {
		WithdrawalID   string  `json:"withdrawal_id"`
		AmountUSDCents int64   `json:"amount_usd_cents"`
		Method         string  `json:"method"`
		CreatedAt      *string `json:"created_at"`
	} `json:"pending_withdrawal"`
	RecentAccruals []struct {
		AmountUSDCents int64   `json:"amount_usd_cents"`
		Kind           string  `json:"kind"`
		AvailableAt    *string `json:"available_at"`
		Reverted       bool    `json:"reverted"`
	} `json:"recent_accruals"`
}

func mapReferralPartner(raw *rawReferralPartner) *ReferralPartner {
	if raw == nil {
		return nil
	}

	methods := make([]string, 0, len(raw.WithdrawMethods))
	methods = append(methods, raw.WithdrawMethods...)

	promoCodes := make([]PartnerPromoCode, 0, len(raw.PromoCodes))
	for _, p := range raw.PromoCodes {
		promoCodes = append(promoCodes, PartnerPromoCode{Code: p.Code, Kind: p.Kind})
	}

	var pending *PartnerPendingWithdrawal
	if pw := raw.PendingWithdrawal; pw != nil {
		pending = &PartnerPendingWithdrawal{
			WithdrawalID:   pw.WithdrawalID,
			AmountUSDCents: pw.AmountUSDCents,
			Method:         pw.Method,
			CreatedAt:      parseOptionalTime(pw.CreatedAt),
		}
	}

	accruals := make([]PartnerAccrual, 0, len(raw.RecentAccruals))
	for _, a := range raw.RecentAccruals {
		accruals = append(accruals, PartnerAccrual{
			AmountUSDCents: a.AmountUSDCents,
			Kind:           a.Kind,
			AvailableAt:    parseOptionalTime(a.AvailableAt),
			Reverted:       a.Reverted,
		})
	}

	return &ReferralPartner{
		RefLink:                       raw.RefLink,
		RefBotLink:                    raw.RefBotLink,
		Percent:                       raw.Percent,
		FirstPercent:                  raw.FirstPercent,
		RecurringPercent:              raw.RecurringPercent,
		HoldDays:                      raw.HoldDays,
		Registrations:                 raw.Registrations,
		PaidReferralsCount:            raw.PaidReferralsCount,
		PromoActivationsCount:         raw.PromoActivationsCount,
		ReferredPaymentsCount:         raw.ReferredPaymentsCount,
		ReferredTurnoverRUBKopecks:    raw.ReferredTurnoverRUBKopecks,
		EarnedUSDCents:                raw.EarnedUSDCents,
		AvailableUSDCents:             raw.AvailableUSDCents,
		OnHoldUSDCents:                raw.OnHoldUSDCents,
		SpentOnSubscriptionsUSDCents:  raw.SpentOnSubscriptionsUSDCents,
		WithdrawnWalletUSDCents:       raw.WithdrawnWalletUSDCents,
		WithdrawnSubscriptionUSDCents: raw.WithdrawnSubscriptionUSDCents,
		MinWithdrawalUSDCents:         raw.MinWithdrawalUSDCents,
		WithdrawMethods:               methods,
		PromoCodes:                    promoCodes,
		PendingWithdrawal:             pending,
		RecentAccruals:                accruals,
	}
}

// ReferralsWithdrawRequest — заявка реферера на вывод всего доступного баланса.
//
// method: только "wallet" (subscription CRM отклоняет с 400, SDK возвращает
// ValidationError без HTTP-вызова). Result.Status: "no_balance" |
// "already_pending" (бот показывает call.answer show_alert) | "created"
// (создана заявка + outbox-событие в мессенджер).
func (c *Client) ReferralsWithdrawRequest(ctx context.Context, userID int64, method string) (*WithdrawRequestResult, error) {
	if userID <= 0 {
		return nil, &ValidationError{Message: "user_id must be a positive integer"}
	}
	m := strings.ToLower(strings.TrimSpace(method))
	if !withdrawRequestMethods[m] {
		return nil, &ValidationError{Message: "method must be 'wallet'"}
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
// method: "wallet" | "subscription" (в отличие от заявки, settle принимает оба).
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
