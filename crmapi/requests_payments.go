package crmapi

import "strings"

type PaymentsCalculateInput struct {
	ProductIDs      []int64 `json:"product_ids"`
	DiscountPercent int64   `json:"discount_percent"`
	Months          int64   `json:"months"`
}

func (in PaymentsCalculateInput) Validate() error {
	if len(in.ProductIDs) == 0 {
		return &ValidationError{Message: "product_ids must contain at least one element"}
	}
	for _, id := range in.ProductIDs {
		if id <= 0 {
			return &ValidationError{Message: "each product_id must be a positive integer"}
		}
	}
	if in.DiscountPercent < 0 || in.DiscountPercent > 90 {
		return &ValidationError{Message: "discount_percent must be between 0 and 90"}
	}
	if in.Months <= 0 {
		return &ValidationError{Message: "months must be a positive integer"}
	}
	return nil
}

type InvoiceDraftInput struct {
	ClientID        int64   `json:"client_id"`
	ProductIDs      []int64 `json:"product_ids"`
	DiscountPercent int64   `json:"discount_percent"`
	Months          int64   `json:"months"`
	Provider        string  `json:"provider"`
}

func (in InvoiceDraftInput) Validate() error {
	if in.ClientID <= 0 {
		return &ValidationError{Message: "client_id must be a positive integer"}
	}
	if len(in.ProductIDs) == 0 {
		return &ValidationError{Message: "product_ids must contain at least one element"}
	}
	for _, id := range in.ProductIDs {
		if id <= 0 {
			return &ValidationError{Message: "each product_id must be a positive integer"}
		}
	}
	if in.DiscountPercent < 0 || in.DiscountPercent > 90 {
		return &ValidationError{Message: "discount_percent must be between 0 and 90"}
	}
	if in.Months <= 0 {
		return &ValidationError{Message: "months must be a positive integer"}
	}

	switch strings.ToLower(strings.TrimSpace(in.Provider)) {
	case "yookassa", "cryptocloud", "heleket":
		return nil
	default:
		return &ValidationError{Message: "provider must be one of: yookassa, cryptocloud, heleket"}
	}
}

func (in InvoiceDraftInput) normalized() InvoiceDraftInput {
	in.Provider = strings.ToLower(strings.TrimSpace(in.Provider))
	return in
}

type InvoiceIssueInput struct {
	UUID        string `json:"uuid"`
	ClientEmail string `json:"client_email"`
}

func (in InvoiceIssueInput) Validate() error {
	if l := len(strings.TrimSpace(in.UUID)); l < 20 || l > 64 {
		return &ValidationError{Message: "uuid length must be between 20 and 64 characters"}
	}
	if l := len(strings.TrimSpace(in.ClientEmail)); l < 3 || l > 128 {
		return &ValidationError{Message: "client_email length must be between 3 and 128 characters"}
	}
	return nil
}

func (in InvoiceIssueInput) normalized() InvoiceIssueInput {
	in.UUID = strings.TrimSpace(in.UUID)
	in.ClientEmail = strings.TrimSpace(in.ClientEmail)
	return in
}

type RefundInput struct {
	Reason      *string `json:"reason,omitempty"`
	AmountMinor *int64  `json:"amount_minor,omitempty"`
}

func (in RefundInput) Validate() error {
	if in.Reason != nil && len(*in.Reason) > 512 {
		return &ValidationError{Message: "reason must be at most 512 characters"}
	}
	if in.AmountMinor != nil && *in.AmountMinor <= 0 {
		return &ValidationError{Message: "amount_minor must be a positive integer"}
	}
	return nil
}
