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

// --- Payment provider validation ------------------------------------------------

func TestInvoiceDraftAcceptsAllSupportedProviders(t *testing.T) {
	// Все провайдеры из SupportedPaymentProviders должны проходить Validate().
	for _, p := range SupportedPaymentProviders {
		t.Run(string(p), func(t *testing.T) {
			err := (InvoiceDraftInput{
				ClientID:        1,
				ProductIDs:      []int64{1},
				DiscountPercent: 0,
				Months:          1,
				Provider:        string(p),
			}).Validate()
			if err != nil {
				t.Fatalf("expected nil error for provider=%q, got %v", p, err)
			}
		})
	}
}

func TestInvoiceDraftAcceptsProviderCaseInsensitive(t *testing.T) {
	// Потребитель может передать " Platega " — Validate() всё равно должен
	// принять (normalized() приведёт к каноническому виду перед отправкой).
	cases := []string{"PLATEGA", "Platega", " platega ", "pLaTeGa"}
	for _, v := range cases {
		t.Run(v, func(t *testing.T) {
			err := (InvoiceDraftInput{
				ClientID:        1,
				ProductIDs:      []int64{1},
				DiscountPercent: 0,
				Months:          1,
				Provider:        v,
			}).Validate()
			if err != nil {
				t.Fatalf("expected nil error for provider=%q, got %v", v, err)
			}
		})
	}
}

func TestInvoiceDraftRejectsUnsupportedProvider(t *testing.T) {
	cases := []string{"", "stripe", "paypal", "yoomoney", "tinkoff"}
	for _, v := range cases {
		t.Run("provider="+v, func(t *testing.T) {
			err := (InvoiceDraftInput{
				ClientID:        1,
				ProductIDs:      []int64{1},
				DiscountPercent: 0,
				Months:          1,
				Provider:        v,
			}).Validate()
			var ve *ValidationError
			if !errorsAsValidation(err, &ve) {
				t.Fatalf("expected ValidationError for provider=%q, got %v", v, err)
			}
			// Сообщение должно перечислять все поддерживаемые провайдеры.
			if !strings.Contains(ve.Message, "platega") {
				t.Errorf("error message should mention 'platega': %s", ve.Message)
			}
		})
	}
}

func TestInvoiceDraftNormalizesProviderToLowerCase(t *testing.T) {
	in := InvoiceDraftInput{
		ClientID:        1,
		ProductIDs:      []int64{1},
		DiscountPercent: 0,
		Months:          1,
		Provider:        " PLATEGA ",
	}
	n := in.normalized()
	if n.Provider != "platega" {
		t.Fatalf("expected normalized provider=%q, got %q", "platega", n.Provider)
	}
}

func TestPaymentProviderIsValid(t *testing.T) {
	if !PaymentProviderPlatega.IsValid() {
		t.Fatalf("PaymentProviderPlatega must be valid")
	}
	if PaymentProvider("stripe").IsValid() {
		t.Fatalf("stripe must not be valid")
	}
	// Case-insensitive
	if !PaymentProvider(" Platega ").IsValid() {
		t.Fatalf("trimmed/upper-case platega must be valid")
	}
}

func TestSupportedPaymentProvidersList(t *testing.T) {
	want := map[PaymentProvider]bool{
		PaymentProviderYooKassa:    true,
		PaymentProviderCryptoCloud: true,
		PaymentProviderHeleket:     true,
		PaymentProviderPlatega:     true,
	}
	if len(SupportedPaymentProviders) != len(want) {
		t.Fatalf("expected %d providers, got %d", len(want), len(SupportedPaymentProviders))
	}
	for _, p := range SupportedPaymentProviders {
		if !want[p] {
			t.Errorf("unexpected provider: %q", p)
		}
	}
}
