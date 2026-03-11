package crmapi

import (
	"context"
	"fmt"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

func (c *Client) CalculatePayment(ctx context.Context, input PaymentsCalculateInput) (*PaymentsCalculateResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var raw struct {
		AmountMinor int64    `json:"amount_minor"`
		AmountUSD   *float64 `json:"amount_usd"`
		Currency    string   `json:"currency"`
		Items       []struct {
			ID                       int64  `json:"id"`
			Title                    string `json:"title"`
			UnitPriceMinor           int64  `json:"unit_price_minor"`
			DiscountPercent          int64  `json:"discount_percent"`
			UnitPriceDiscountedMinor int64  `json:"unit_price_discounted_minor"`
			Quantity                 int64  `json:"quantity"`
			LineTotalMinor           int64  `json:"line_total_minor"`
		} `json:"items"`
	}

	if err := c.post(ctx, "/api/payments/calculate", nil, true, input, &raw); err != nil {
		return nil, err
	}

	items := make([]CalcItem, 0, len(raw.Items))
	for _, i := range raw.Items {
		items = append(items, CalcItem{
			ID:                       i.ID,
			Title:                    i.Title,
			UnitPriceMinor:           i.UnitPriceMinor,
			DiscountPercent:          i.DiscountPercent,
			UnitPriceDiscountedMinor: i.UnitPriceDiscountedMinor,
			Quantity:                 i.Quantity,
			LineTotalMinor:           i.LineTotalMinor,
		})
	}

	currency := raw.Currency
	if currency == "" {
		currency = "RUB"
	}

	return &PaymentsCalculateResult{
		AmountMinor: raw.AmountMinor,
		AmountUSD:   raw.AmountUSD,
		Currency:    currency,
		Items:       items,
	}, nil
}

func (c *Client) CreateInvoiceDraft(ctx context.Context, input InvoiceDraftInput) (*InvoiceDraftResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		UUID    string `json:"uuid"`
		PayLink string `json:"pay_link"`
		Status  string `json:"status"`
	}

	if err := c.post(ctx, "/api/payments/invoice/draft", nil, true, input, &raw); err != nil {
		return nil, err
	}

	return &InvoiceDraftResult{
		UUID:    raw.UUID,
		PayLink: raw.PayLink,
		Status:  raw.Status,
	}, nil
}

func (c *Client) IssueInvoice(ctx context.Context, input InvoiceIssueInput) (*InvoiceIssueResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	input = input.normalized()

	var raw struct {
		PayURL string `json:"pay_url"`
		Status string `json:"status"`
	}

	if err := c.post(ctx, "/api/payments/invoice/issue", nil, true, input, &raw); err != nil {
		return nil, err
	}

	return &InvoiceIssueResult{
		PayURL: raw.PayURL,
		Status: raw.Status,
	}, nil
}

func (c *Client) GetInvoiceInfo(ctx context.Context, uuid string) (*InvoiceInfoResult, error) {
	if uuid == "" {
		return nil, &ValidationError{Message: "uuid must not be empty"}
	}

	var raw struct {
		UUID            string           `json:"uuid"`
		Status          string           `json:"status"`
		StatusRU        string           `json:"status_ru"`
		ClientID        int64            `json:"client_id"`
		ClientEmail     *string          `json:"client_email"`
		RefererID       *int64           `json:"referer_id"`
		StaffID         *int64           `json:"staff_id"`
		AmountMinor     int64            `json:"amount_minor"`
		FXRateRUBUSD    *float64         `json:"fx_rate_rub_usd"`
		Currency        string           `json:"currency"`
		DiscountPercent *int64           `json:"discount_percent"`
		Description     string           `json:"description"`
		Items           []map[string]any `json:"items"`
		Provider        string           `json:"provider"`
		PayLink         *string          `json:"pay_link"`
		PayURL          *string          `json:"pay_url"`
		DateCreate      *string          `json:"date_create"`
		DateInvoiced    *string          `json:"date_invoiced"`
		DatePaid        *string          `json:"date_paid"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/payments/invoice/%s", uuid), nil, true, &raw); err != nil {
		return nil, err
	}

	var dateCreate *time.Time
	var dateInvoiced *time.Time
	var datePaid *time.Time

	if raw.DateCreate != nil {
		dateCreate = utils.ParseTime(*raw.DateCreate)
	}
	if raw.DateInvoiced != nil {
		dateInvoiced = utils.ParseTime(*raw.DateInvoiced)
	}
	if raw.DatePaid != nil {
		datePaid = utils.ParseTime(*raw.DatePaid)
	}

	currency := raw.Currency
	if currency == "" {
		currency = "RUB"
	}

	return &InvoiceInfoResult{
		UUID:            raw.UUID,
		Status:          raw.Status,
		StatusRU:        raw.StatusRU,
		ClientID:        raw.ClientID,
		ClientEmail:     raw.ClientEmail,
		RefererID:       raw.RefererID,
		StaffID:         raw.StaffID,
		AmountMinor:     raw.AmountMinor,
		FXRateRUBUSD:    raw.FXRateRUBUSD,
		Currency:        currency,
		DiscountPercent: raw.DiscountPercent,
		Description:     raw.Description,
		Items:           raw.Items,
		Provider:        raw.Provider,
		PayLink:         raw.PayLink,
		PayURL:          raw.PayURL,
		DateCreate:      dateCreate,
		DateInvoiced:    dateInvoiced,
		DatePaid:        datePaid,
	}, nil
}

func (c *Client) GetPayments(ctx context.Context, userID *int64) ([]PaymentHistoryItem, error) {
	query := map[string]string{}
	if userID != nil {
		if *userID <= 0 {
			return nil, &ValidationError{Message: "user_id must be a positive integer"}
		}
		query["user_id"] = fmt.Sprintf("%d", *userID)
	}
	if len(query) == 0 {
		query = nil
	}

	var raw []struct {
		UUID            string           `json:"uuid"`
		Status          string           `json:"status"`
		StatusRU        string           `json:"status_ru"`
		ClientID        int64            `json:"client_id"`
		ClientEmail     *string          `json:"client_email"`
		RefererID       *int64           `json:"referer_id"`
		StaffID         *int64           `json:"staff_id"`
		AmountMinor     int64            `json:"amount_minor"`
		FXRateRUBUSD    *float64         `json:"fx_rate_rub_usd"`
		Currency        string           `json:"currency"`
		DiscountPercent *int64           `json:"discount_percent"`
		Description     *string          `json:"description"`
		Items           []map[string]any `json:"items"`
		Provider        *string          `json:"provider"`
		PayLink         *string          `json:"pay_link"`
		PayURL          *string          `json:"pay_url"`
		DateCreate      *string          `json:"date_create"`
		DateInvoiced    *string          `json:"date_invoiced"`
		DatePaid        *string          `json:"date_paid"`
		Activation      []struct {
			BotID  *int64 `json:"bot_id"`
			Code   string `json:"code"`
			IsUsed bool   `json:"is_used"`
			URL    string `json:"url"`
		} `json:"activation"`
	}

	if err := c.get(ctx, "/api/payments", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]PaymentHistoryItem, 0, len(raw))
	for _, p := range raw {
		activation := make([]ActivationLink, 0, len(p.Activation))
		for _, ac := range p.Activation {
			activation = append(activation, ActivationLink{
				BotID:  ac.BotID,
				Code:   ac.Code,
				IsUsed: ac.IsUsed,
				URL:    ac.URL,
			})
		}

		var dateCreate *time.Time
		var dateInvoiced *time.Time
		var datePaid *time.Time

		if p.DateCreate != nil {
			dateCreate = utils.ParseTime(*p.DateCreate)
		}
		if p.DateInvoiced != nil {
			dateInvoiced = utils.ParseTime(*p.DateInvoiced)
		}
		if p.DatePaid != nil {
			datePaid = utils.ParseTime(*p.DatePaid)
		}

		currency := p.Currency
		if currency == "" {
			currency = "RUB"
		}

		items = append(items, PaymentHistoryItem{
			UUID:            p.UUID,
			Status:          p.Status,
			StatusRU:        p.StatusRU,
			ClientID:        p.ClientID,
			ClientEmail:     p.ClientEmail,
			RefererID:       p.RefererID,
			StaffID:         p.StaffID,
			AmountMinor:     p.AmountMinor,
			FXRateRUBUSD:    p.FXRateRUBUSD,
			Currency:        currency,
			DiscountPercent: p.DiscountPercent,
			Description:     p.Description,
			Items:           p.Items,
			Provider:        p.Provider,
			PayLink:         p.PayLink,
			PayURL:          p.PayURL,
			DateCreate:      dateCreate,
			DateInvoiced:    dateInvoiced,
			DatePaid:        datePaid,
			Activation:      activation,
		})
	}

	return items, nil
}

func (c *Client) ConfirmPayment(ctx context.Context, uuid string) (*ConfirmPaymentResult, error) {
	if uuid == "" {
		return nil, &ValidationError{Message: "uuid must not be empty"}
	}

	var raw struct {
		UUID   string `json:"uuid"`
		Status string `json:"status"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/payments/confirm/%s", uuid), nil, true, &raw); err != nil {
		return nil, err
	}

	return &ConfirmPaymentResult{
		UUID:   raw.UUID,
		Status: raw.Status,
	}, nil
}

func (c *Client) RefundPayment(ctx context.Context, uuid string, payload *RefundInput) (*RefundResult, error) {
	if uuid == "" {
		return nil, &ValidationError{Message: "uuid must not be empty"}
	}
	if payload != nil {
		if err := payload.Validate(); err != nil {
			return nil, err
		}
	}

	var body any
	if payload != nil {
		body = payload
	}

	var raw struct {
		UUID     string  `json:"uuid"`
		Provider string  `json:"provider"`
		Allowed  bool    `json:"allowed"`
		Message  string  `json:"message"`
		Status   *string `json:"status"`
	}

	if err := c.post(ctx, fmt.Sprintf("/api/payments/refund/%s", uuid), nil, true, body, &raw); err != nil {
		return nil, err
	}

	return &RefundResult{
		UUID:     raw.UUID,
		Provider: raw.Provider,
		Allowed:  raw.Allowed,
		Message:  raw.Message,
		Status:   raw.Status,
	}, nil
}
