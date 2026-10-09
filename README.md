# CRM API SDK for Go

Типобезопасный Go SDK для работы с CRM API, подготовленный для backend-сервисов, workers, bots и internal tools.

## Особенности

- Явная типизация моделей и запросов
- JWT-аутентификация с in-memory cache
- Автоматический refresh токена при `401`
- Bounded retry для временных транспортных ошибок и временных `502/503/504`
- Настраиваемые `User-Agent`, transport/doer и token cache
- Контекстная модель вызовов через `context.Context`
- SDK не читает `.env` и переменные окружения — конфигурация передаётся явно

## Установка

```bash
go get github.com/dead-matrix/crm-api-go-sdk
```

> Актуальный module path: `github.com/dead-matrix/crm-api-go-sdk`.

## Быстрый старт

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi"
)

func main() {
	client, err := crmapi.NewClient(crmapi.Config{
		BaseURL:        "https://your-crm.example",
		StaffID:        123,
		ServiceToken:   "YOUR_SERVICE_TOKEN",
		UserAgent:      "your-service/1.0",
		RequestRetries: 3,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	user, err := client.GetUser(ctx, 7014133383)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", user)
}
```

## Production-конфигурация

`crmapi.Config` поддерживает:

- кастомный transport/doer через `HTTPClient`
- инъекцию token cache через `TokenCache`
- кастомный `User-Agent`
- retry policy через `RequestRetries`, `RetryBaseDelay`, `RetryMaxDelay`, `RetryStatusCodes`
- опциональный retry для non-idempotent запросов через `RetryNonIdempotent`

По умолчанию status-retry применяется только к idempotent-методам (`GET`, `PUT`, `DELETE`, ...), что безопаснее для mutation-операций.

## Принципы использования

- Передавайте `context.Context` во все вызовы.
- Переиспользуйте один экземпляр `Client` между запросами.
- Вызывайте `Close()` при завершении работы процесса, чтобы закрыть idle connections кастомного транспорта.
- SDK не загружает `.env` автоматически.

## Платёжные провайдеры

SDK валидирует имя провайдера в `InvoiceDraftInput.Validate()` ещё до HTTP-запроса.
Используйте предопределённые константы, чтобы не опечататься:

```go
draft, err := client.CreateInvoiceDraft(ctx, crmapi.InvoiceDraftInput{
    ClientID:        123,
    ProductIDs:      []int64{1, 2},
    DiscountPercent: 10,
    Months:          1,
    Provider:        string(crmapi.PaymentProviderPlatega), // yookassa | cryptocloud | heleket | platega
})
```

Список поддерживаемых провайдеров доступен через `crmapi.SupportedPaymentProviders` —
удобно использовать для построения UI-списков:

```go
for _, p := range crmapi.SupportedPaymentProviders {
    fmt.Println(p) // yookassa, cryptocloud, heleket, platega
}
```

Проверить значение без обращения к CRM:

```go
if !crmapi.PaymentProvider("platega").IsValid() {
    // не поддерживается текущей версией SDK
}
```

Валидация case-insensitive и trim'ит пробелы; значение нормализуется к нижнему регистру
перед отправкой в CRM API (см. `InvoiceDraftInput.normalized()`).

## Обработка ошибок

SDK возвращает типизированные ошибки:

- `*crmapi.ConfigError`
- `*crmapi.ValidationError`
- `*crmapi.AuthError`
- `*crmapi.APIError`
- `*crmapi.HTTPError`

Пример:

```go
user, err := client.GetUser(ctx, 123)
if err != nil {
	var authErr *crmapi.AuthError
	if errors.As(err, &authErr) {
		// обработка auth ошибки
	}
}
```

## Real API smoke tests

Реальные smoke-тесты намеренно **не запускаются** при обычном `go test ./...`.

1. Скопируй `.env.example` в `.env`
2. Заполни локальные значения
3. Запускай явно:

```bash
RUN_REAL_API_TESTS=1 go test ./crmapi -run TestRealAPI_Smoke -v
```

Это защищает локальную разработку и CI от случайных вызовов в real API.

## Git hygiene

- `.env` с секретами должен оставаться локальным и не коммититься
- в репозитории держи только `.env.example`

## Migration notes (Unreleased)

См. `docs/SDK_PARITY.md` для полного списка изменений и обоснований.
Ключевые BREAKING-изменения по сравнению с предыдущей версией модуля:

- Ряд скалярных полей моделей ответа стал указательным (nullable), чтобы
  сохранить серверный `null` и не превращать его в zero-value:
  - `ChangeStatusResult.Status: string → *string`
  - `CreateUserResult.FullName: string → *string`
  - `ListUserItem.FullName: string → *string`
  - `DialogSearchItem.Status` / `StatusColor: string → *string`
  - `ActivationRedeemResult.PaymentID: int64 → *int64`
- `DeliveryRef.LastUsedAt` / `CreatedAt` / `UpdatedAt`:
  `*string → *time.Time` — SDK теперь парсит ISO 8601 в `time.Time` сам.
- `ReplyTemplateListItem` / `ReplyTemplateFull` получили `json`-теги
  `lastUsedAt,omitempty` / `createdAt,omitempty` / `updatedAt,omitempty` —
  re-marshal публичных структур теперь даёт camelCase, паритетный с
  серверным форматом.
- Default `RetryStatusCodes` дополнен `429` — единая retry-policy с Python SDK.
- Платежи (CRM SocialTraff):
  - `PaymentHistoryItem.ClientID` / `InvoiceInfoResult.ClientID: int64 → *int64`,
    `Sale.UserID: int64 → *int64` — у платежей без Telegram-пользователя приходит `null`;
  - новые поля `AccountID *int64` в `PaymentHistoryItem` и `InvoiceInfoResult`,
    `InvoiceInfoResult.WebReturnURL *string`;
  - `PaymentHistoryItem.Access *PaymentAccess` (`AccountID`, `Plans`, `AccessEnd`)
    заменяет `Activation`; `Activation` помечено Deprecated и остаётся пустым.
- Smoke против CRM SocialTraff:
  `RUN_REAL_API_TESTS=1 go test ./crmapi -run TestSocialTraffSmoke -v`.
- Рефералы (CRM SocialTraff, v0.0.0-socialtraff.3):
  - `ReferralsInfoResult.Partner *ReferralPartner`: сводка партнёра из ключа
    `partner` (ставки, холд, счётчики, суммы в USD-центах и копейках, промокоды
    `PartnerPromoCode`, открытая заявка `PartnerPendingWithdrawal`, последние
    начисления `PartnerAccrual`); `nil`, если CRM ключ не прислала;
  - `AvailableUSD` верхнего уровня теперь без холда (холд в `Partner.OnHoldUSDCents`);
  - `ReferralsWithdrawRequest` принимает только `method = "wallet"`, для
    `subscription` возвращает `ValidationError` без запроса к CRM;
    `ReferralsWithdrawSettle` по-прежнему принимает `wallet` и `subscription`.
- Данные покупателя и аккаунты (CRM SocialTraff). Старые сигнатуры и типы полей
  не менялись, кроме поведения `AIUsage`:
  - `AIUsage` больше не подставляет `bot_id=1`: бота выбирает CRM. Явный бот и
    глубина отчёта задаются новым `AIUsageWithOptions(ctx, userID,
    AIUsageOptions{BotID, Days, Months, Recent})`; кому нужен прежний запрос,
    передаёт `AIUsageOptions{BotID: 1}`. В отчёте появились `AccountID`,
    `KeyStats *AIKeyStats`, у функции `Unlimited` и `AdminSpentTokens` /
    `AdminSpentUSD` / `AdminGenerations`, у генерации `Source`;
  - карточка человека: `GetUserResult.Buyer *UserBuyer` (`nil`, если человек не
    покупатель) и `GetUserResult.Accounts []UserAccount` (всегда не `nil`),
    `UserBotInfo.AccountID`. `Buyer.BuyerID` и есть тот `userID`, который ждут
    `ReferralsInfo`, `ReferralsWithdrawRequest` и `ReferralsWithdrawSettle`;
  - новый `GetUserAccount(ctx, userID, accountID)` возвращает `UserAccountCard`:
    участники, чаты, боты, ожидающие промокоды. Человек вне аккаунта получает
    `*APIError` с `Code == "not_found"`;
  - `account_id` в доступе: `AddAccessInput.AccountID` и `IdempotencyKey`,
    действие `ActionRemove`, `AccessManageInput.AccountID`. `UserID` можно не
    задавать, если задан `AccountID` (нужен хотя бы один); в ответе
    `AccountID`, а `UserID` при вызове только по аккаунту равен 0;
  - новые `ExtendUserAccessForAccount(ctx, userID, botID, days, accountID)` и
    `SubscriptionsHistoryForAccount(ctx, userID, accountID)`; в ответах
    `ExtendAccessResult.AccountID`, `SubscriptionsHistoryResult.AccountID`,
    `AccessHistoryItem.Added` / `Removed`, `ActivationRedeemResult.AccountID`;
  - `GrantAITokensResult.AccountID` и `Unlimited`;
  - рефералы: `ReferralsInfoResult.RefBotLink`; у заявки на вывод
    `AmountMinor`, `AvailableMinor`, `MinMinor`, `MinUSD` и статус `below_min`;
    у проведения вывода `CurrentStatus` (только при `already_settled`);
  - блокировки бота: `ListBotBlocksByKind(ctx, botID, kind)` с константами
    `BotBlockKindBlocked` / `BotBlockKindUnreachable`, в списке `Kind` и
    `Counts`, в `BotBlockReportResult` признак `Ignored`.

Миграция: добавить nil-check перед разыменованием указателей.
Компилятор Go подсветит все места, где старый код полагался на zero-value
этих полей.

## Лицензия

Proprietary.
