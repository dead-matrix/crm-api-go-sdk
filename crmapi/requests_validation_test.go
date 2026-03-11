package crmapi

import (
	"strings"
	"testing"
)

func TestRequestValidateRejectsInvalidInputs(t *testing.T) {
	reasonTooLong := strings.Repeat("r", 513)
	refTooLong := strings.Repeat("x", 2049)
	cases := []struct {
		name string
		fn   func() error
	}{
		{"create user", func() error { return (CreateUserInput{}).Validate() }},
		{"update user", func() error { return (UpdateUserInput{}).Validate() }},
		{"payments calculate", func() error { return (PaymentsCalculateInput{}).Validate() }},
		{"invoice draft", func() error { return (InvoiceDraftInput{}).Validate() }},
		{"invoice issue", func() error { return (InvoiceIssueInput{}).Validate() }},
		{"refund", func() error { return (RefundInput{Reason: &reasonTooLong}).Validate() }},
		{"add access", func() error { return (AddAccessInput{Ref: &refTooLong}).Validate() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var validationErr *ValidationError
			if err := tc.fn(); !errorsAsValidation(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

func TestRequestValidateAcceptsValidInputs(t *testing.T) {
	username := "john"
	refer := "ref"
	reason := "duplicate"
	amount := int64(10)
	paymentID := int64(1)
	cases := []struct {
		name string
		fn   func() error
	}{
		{"create user", func() error {
			return (CreateUserInput{UserID: 1, FullName: "John", Username: &username, BotID: 2, Refer: &refer}).Validate()
		}},
		{"update user", func() error { return (UpdateUserInput{FullName: "John", Username: &username}).Validate() }},
		{"payments calculate", func() error {
			return (PaymentsCalculateInput{ProductIDs: []int64{1}, DiscountPercent: 10, Months: 1}).Validate()
		}},
		{"invoice draft", func() error {
			return (InvoiceDraftInput{ClientID: 1, ProductIDs: []int64{1}, DiscountPercent: 10, Months: 1, Provider: "yookassa"}).Validate()
		}},
		{"invoice issue", func() error { return (InvoiceIssueInput{UUID: strings.Repeat("u", 20), ClientEmail: "a@b"}).Validate() }},
		{"refund", func() error { return (RefundInput{Reason: &reason, AmountMinor: &amount}).Validate() }},
		{"add access", func() error {
			return (AddAccessInput{UserID: 1, BotID: 2, Action: ActionAdd, PaymentID: &paymentID, Ref: &refer}).Validate()
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func errorsAsValidation(err error, target **ValidationError) bool {
	if err == nil {
		return false
	}
	v, ok := err.(*ValidationError)
	if ok {
		*target = v
	}
	return ok
}
