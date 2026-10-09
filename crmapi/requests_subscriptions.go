package crmapi

import (
	"strings"
	"time"
	"unicode/utf8"
)

type ActionType string

const (
	ActionAdd    ActionType = "add"
	ActionExtend ActionType = "extend"
	ActionRevoke ActionType = "revoke"
	ActionRefund ActionType = "refund"
	// ActionRemove - синоним снятия: CRM обрабатывает его как revoke/refund,
	// но в строке доступа остаётся именно это слово, поэтому константа своя.
	ActionRemove ActionType = "remove"
	// ActionCustom - кастомная правка сроков: у каждой фичи свой сдвиг конца
	// на ±дни (см. AddAccessInput.Deltas), CRM применяет mode=adjust.
	ActionCustom ActionType = "custom"
)

// accessIdempotencyKeyMaxLen - предел CRM на длину idempotency_key (символы).
const accessIdempotencyKeyMaxLen = 64

type AddAccessInput struct {
	// UserID - Telegram id человека. Нужен хотя бы один из UserID и AccountID:
	// без AccountID действие идёт в основной аккаунт человека. Ноль в запрос
	// не попадает: CRM отвергает user_id=0, а вызов только по аккаунту законен.
	UserID int64 `json:"user_id,omitempty"`
	// AccountID - аккаунт, чей доступ меняется. Вместе с UserID человек обязан
	// состоять в этом аккаунте, иначе CRM ответит 404 not_found.
	AccountID  *int64     `json:"account_id,omitempty"`
	BotID      int64      `json:"bot_id"`
	Action     ActionType `json:"action"`
	Access     any        `json:"access,omitempty"`
	ActionDate *time.Time `json:"action_date,omitempty"`
	AccessEnd  *time.Time `json:"access_end,omitempty"`
	PaymentID  *int64     `json:"payment_id,omitempty"`
	Ref        *string    `json:"ref,omitempty"`
	// Days - частичное снятие: при revoke/refund/remove с Days>0 CRM отнимает
	// N дней у выданных фич (mode=subtract) вместо полного снятия. Игнорируется
	// на add/extend.
	Days *int64 `json:"days,omitempty"`
	// Deltas - кастомная правка (Action=custom): у каждой фичи свой сдвиг конца
	// на ±дни {ключ: дни} (mode=adjust). Минус ограничен полом now+1день на
	// стороне CRM. Игнорируется на прочих действиях.
	Deltas map[string]int64 `json:"deltas,omitempty"`
	// IdempotencyKey - ключ повтора (до 64 символов): второй вызов с тем же
	// ключом возвращает прежнюю строку доступа и не создаёт новых событий.
	// Без ключа CRM схлопывает только одинаковые операции в окне около минуты.
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
}

// validateAccessTarget проверяет адресата операции над доступом: человек,
// аккаунт или оба. Общая для AddAccess и ManageAccess, потому что CRM
// разрешает аккаунт одинаково в обоих маршрутах.
func validateAccessTarget(userID int64, accountID *int64) error {
	if userID < 0 {
		return &ValidationError{Message: "user_id must be a positive integer"}
	}
	if accountID != nil && *accountID <= 0 {
		return &ValidationError{Message: "account_id must be a positive integer"}
	}
	if userID == 0 && accountID == nil {
		return &ValidationError{Message: "user_id or account_id must be provided"}
	}
	return nil
}

func (in AddAccessInput) Validate() error {
	if err := validateAccessTarget(in.UserID, in.AccountID); err != nil {
		return err
	}
	if in.BotID <= 0 {
		return &ValidationError{Message: "bot_id must be a positive integer"}
	}

	action := ActionType(strings.ToLower(strings.TrimSpace(string(in.Action))))
	switch action {
	case ActionAdd, ActionExtend, ActionRevoke, ActionRefund, ActionRemove, ActionCustom:
	default:
		return &ValidationError{Message: "action must be one of: add, extend, revoke, refund, remove, custom"}
	}
	if action == ActionCustom && len(in.Deltas) == 0 {
		return &ValidationError{Message: "custom action requires a non-empty deltas map"}
	}
	for key, n := range in.Deltas {
		if strings.TrimSpace(key) == "" {
			return &ValidationError{Message: "deltas keys must be non-empty feature names"}
		}
		if n < -3650 || n > 3650 {
			return &ValidationError{Message: "deltas values must be between -3650 and 3650"}
		}
	}

	if in.PaymentID != nil && *in.PaymentID <= 0 {
		return &ValidationError{Message: "payment_id must be a positive integer"}
	}
	if in.Ref != nil && len(*in.Ref) > 2048 {
		return &ValidationError{Message: "ref must be at most 2048 characters"}
	}
	if in.Days != nil && (*in.Days < 1 || *in.Days > 3650) {
		return &ValidationError{Message: "days must be between 1 and 3650"}
	}
	// Длина в символах, а не в байтах: так её считает CRM.
	if in.IdempotencyKey != nil && utf8.RuneCountInString(*in.IdempotencyKey) > accessIdempotencyKeyMaxLen {
		return &ValidationError{Message: "idempotency_key must be at most 64 characters"}
	}

	return nil
}

func (in AddAccessInput) normalized() AddAccessInput {
	in.Action = ActionType(strings.ToLower(strings.TrimSpace(string(in.Action))))
	return in
}
