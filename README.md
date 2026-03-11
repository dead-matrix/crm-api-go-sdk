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

## Лицензия

Proprietary.
