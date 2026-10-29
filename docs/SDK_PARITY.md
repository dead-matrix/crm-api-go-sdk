# SDK Parity — Python ↔ Go

Внутренний инженерный документ. Описывает соответствие между публичным API
двух клиентов CRM API и серверным контрактом.

- Python SDK — пакет `crm_api` (этот репозиторий или соседний).
- Go SDK — пакет `crmapi` (модуль `github.com/dead-matrix/crm-api-go-sdk`).
- Сервер — CRM-API на FastAPI, маршруты под префиксом `/api`.

**Дата актуализации:** 2026-10-09.

Строки и разделы, добавленные 2026-10-09, сверены с Go SDK и кодом CRM
SocialTraff. Python SDK в эту итерацию не сверялся: в таких строках столбец
Python помечен «не сверялось», это не утверждение, что метода там нет.

## Как читать этот документ

1. Таблица «Покрытие модулей» — high-level сводка.
2. Таблица «Эндпоинты» — endpoint → Python метод → Go метод → request → response.
3. «Nullable-sensitive поля» — поля, где сервер реально возвращает null
   и где SDK обязан сохранить это значение (а не превратить в `"None"`/`""`).
4. «Intentional differences» — где Python и Go намеренно расходятся.
5. «Закрытые parity-баги (2026-05-20)» — что было поправлено в этой итерации.
6. «Изменения 2026-10-09» - что добавлено в Go SDK под CRM SocialTraff.

Если меняете SDK — обновите соответствующую строку и дату актуализации.

## Покрытие модулей

| Модуль | Покрыто Python | Покрыто Go | Notes |
|---|---|---|---|
| accounts | ✅ | ✅ | |
| activation | ✅ | ✅ | |
| bot_blocks | не сверялось | ✅ | флаги «бот не может писать человеку» |
| catalog | ✅ | ✅ | |
| departments | ✅ | ✅ | |
| dialogs | ✅ | ✅ | |
| notes | ✅ | ✅ | |
| payments | ✅ | ✅ | |
| profile | ✅ | ✅ | |
| prompts | ✅ | ✅ | |
| proxy | ✅ | ✅ | |
| referrals | ✅ | ✅ | |
| reply_templates | ✅ | ✅ | включая `delivery-refs` |
| sales_decks (scripts) | ✅ | ✅ | URL префикс `/scripts/*` |
| servers | ✅ | ✅ | |
| staff | ✅ | ✅ | без `/staff/{user_id}` — нет такого маршрута |
| subscriptions | ✅ | ✅ | |
| tasks | ✅ | ✅ | |
| users | ✅ | ✅ | |

Webhook-эндпоинты CRM-API наружу не выставляются и в SDK отсутствуют намеренно.

## Эндпоинты

Таблица: HTTP method + path → Python метод → Go метод → request model → response model.

Все маршруты ниже под префиксом `/api`. Все методы требуют JWT (бабушек делает `POST /api/staff/{staff_id}/auth` под капотом).

### accounts
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /accounts/list?user_id=&include_removed=&removed_only=&days=` | `accounts_list(user_id, include_removed, days, removed_only)` | `AccountsList(ctx, userID, includeRemoved, removedOnly, days)` | — | `[]AccountItem` (`DayTotal` вложенный; `first_load` ISO-строка, `removed`); `removed_only` только удалённые, `days>0` ограничивает период по `first_load` |
| `GET /accounts/count?user_id=&include_removed=&removed_only=` | `accounts_count(user_id, include_removed, removed_only)` | `AccountsCount(ctx, userID, includeRemoved, removedOnly)` | — | `{total}` (дешёвый COUNT удалённых перед их загрузкой) |

### activation
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `POST /activation/redeem` | `activation_redeem(data)` | `ActivationRedeem(ctx, input)` | `ActivationRedeemInput { token, recipient_user_id, bot_id }` | `ActivationRedeemResult { success, error_code, error_message, account_id?, user_id, bot_id, action, access, access_end, quantity, activation_code_id, payment_id?, idempotent_replay }` |

### bot_blocks
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /bot-blocks?bot_id=` | не сверялось | `ListBotBlocks(ctx, botID)` | - | `BotBlocksListResult { bot_id, kind?, user_ids, count, counts { blocked, unreachable, total } }`; `kind` в ответе `null` |
| `GET /bot-blocks?bot_id=&kind=` | не сверялось | `ListBotBlocksByKind(ctx, botID, kind)` | `kind`: `blocked` \| `unreachable` (константы `BotBlockKind*`), проверяется в SDK: на неизвестное значение CRM отвечает успехом с пустым списком | то же; `count` считает отфильтрованное, `counts` все флаги бота |
| `POST /bot-blocks/unblock` | не сверялось | `UnblockBotBlock(ctx, botID, userID)` | `{ bot_id, user_id }` | `BotBlockUnblockResult { removed }` |
| `POST /bot-blocks/report` | не сверялось | `ReportBotBlock(ctx, botID, userID, reason, errText)` | `{ bot_id, user_id, reason, error? }` | `BotBlockReportResult { added, ignored }`; `ignored=true`: CRM флаги этого бота не ведёт |

### catalog
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /products/active` | `products_active()` | `ProductsActive(ctx)` | — | `Map[str, CategoryBucket]` |

### departments
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /departments` | `list_departments()` | `ListDepartments(ctx)` | — | `[]DepartmentItem` |

### dialogs
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /dialogs/{department}` | `get_dialogs(department)` | `GetDialogs(ctx, department)` | — | `[]DialogItem` |
| `GET /dialogs/{department}/search?q=&offset=` | `search_dialogs(department, q, offset)` | `SearchDialogs(ctx, department, q, offset)` | — | `DialogSearchResult { dialogs, limit, offset }` |
| `GET /dialogs/statuses/{department_id}` | `get_statuses(department_id)` | `GetStatuses(ctx, departmentID)` | — | `StatusesResult { department_id, default_status_id, statuses }` |
| `POST /dialogs/status` | `change_dialog_status(user_id, status_id)` / `clear_dialog_status(user_id)` | `ChangeDialogStatus(ctx, userID, statusID)` / `ClearDialogStatus(ctx, userID)` | `{ user_id, status_id }` (status_id может быть `null`) | `ChangeStatusResult { status: str?/string? }` |
| `POST /dialogs/transfer` | `transfer_dialog(user_id, to_department)` | `TransferDialog(ctx, userID, toDepartment)` | `{ user_id, to_department }` | `TransferDialogResult { transferred: bool }` |

### notes
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /users/{user_id}/notes` | `list_user_notes(user_id)` | `ListUserNotes(ctx, userID)` | — | `[]NoteItem` |
| `POST /users/{user_id}/notes` | `create_user_note(user_id, text)` | `CreateUserNote(ctx, userID, text)` | `{ text }` | `NoteItem` |

### payments
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `POST /payments/calculate` | `calculate_payment(data)` | `CalculatePayment(ctx, input)` | `PaymentsCalculateInput { product_ids, discount_percent, months }` | `PaymentsCalculateResult` |
| `POST /payments/invoice/draft` | `create_invoice_draft(data)` | `CreateInvoiceDraft(ctx, input)` | `InvoiceDraftInput { client_id, product_ids, discount_percent, months, provider, payment_method? }` (`payment_method` опц., шлётся только для platega через `exclude_none`/`omitempty`) | `InvoiceDraftResult { uuid, pay_link, status }` |
| `POST /payments/invoice/issue` | `issue_invoice(data)` | `IssueInvoice(ctx, input)` | `InvoiceIssueInput { uuid, client_email }` | `InvoiceIssueResult { pay_url, status }` |
| `GET /payments/invoice/{uuid}` | `get_invoice_info(uuid)` | `GetInvoiceInfo(ctx, uuid)` | — | `InvoiceInfoResult` |
| `GET /payments?user_id=&limit=&offset=` | `get_payments(user_id?, limit, offset)` | `GetPayments(ctx, userID*, limit, offset)` | — | `PaymentsListResult { limit, offset, count, items: []PaymentHistoryItem }` |
| `GET /payments/sales` | `get_monthly_sales()` | `GetMonthlySales(ctx)` | — | `MonthlySalesResult { month_start, payments: []Sale }` |
| `POST /payments/purchase-summary` | — | `PurchaseSummaries(ctx, userIDs)` | `{ user_ids: []int64 }` (≤5000) | `map[int64]PurchaseSummary { first_paid_at, last_paid_at, payments_count, total_minor }` — каждый запрошенный id присутствует; нули = не покупал |
| `GET /payments/confirm/{uuid}` | `confirm_payment(uuid)` | `ConfirmPayment(ctx, uuid)` | — | `ConfirmPaymentResult { uuid, status }` |
| `POST /payments/refund/{uuid}` | `refund_payment(uuid, data?)` | `RefundPayment(ctx, uuid, input*)` | `RefundInput { reason?, amount_minor? }` | `RefundResult { uuid, provider, allowed, message, status? }` |

### profile
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /profile/statistics?user_id=` | `profile_statistics(user_id)` | `ProfileStatistics(ctx, userID)` | — | `ProfileStatistics` |
| `GET /profile/bot3/summary?user_id=` | `profile_bot3_summary(user_id)` | `ProfileBot3Summary(ctx, userID)` | — | `Bot3Summary { subscription, account, tasks }` |

### prompts
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /prompt?user_id=` | `prompt_get(user_id)` | `PromptGet(ctx, userID)` | — | `Optional[str]` / `*string` |
| `POST /prompt/update?user_id=` | `prompt_update(user_id, prompt)` | `PromptUpdate(ctx, userID, prompt)` | `{ prompt }` | `PromptUpdateResult { reset?, updated?, created?, message? }` |

### proxy
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `POST /proxy/check?user_id=` | `proxy_check(user_id)` | `ProxyCheck(ctx, userID)` | — | `ProxyCheckResult` |
| `GET /proxy/list?user_id=` | `proxy_list(user_id)` | `ProxyList(ctx, userID)` | — | `[]ProxyItem` |

### referrals
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /referrals/info?user_id=` | `referrals_info(user_id)` | `ReferralsInfo(ctx, userID)` | `user_id` в CRM SocialTraff: `buyer_id` покупателя (`GetUserResult.Buyer.BuyerID`), не Telegram id | `ReferralsInfoResult` (в том числе `ref_bot_link?`, `partner?`) |
| `POST /referrals/withdraw/request` | не сверялось | `ReferralsWithdrawRequest(ctx, userID, method)` | `{ user_id, method }`; `method` только `wallet`, `user_id` как выше | `WithdrawRequestResult { status, withdrawal_id?, amount_minor?, amount_usd?, method?, available_minor?, available_usd?, min_minor?, min_usd? }`; `status`: `created` \| `already_pending` \| `no_balance` \| `below_min` |
| `POST /referrals/withdraw/settle` | не сверялось | `ReferralsWithdrawSettle(ctx, userID, amountMinor, method, withdrawalID*)` | `{ user_id, amount_minor, method, withdrawal_id? }`; `method`: `wallet` \| `subscription` | `WithdrawSettleResult { status, withdrawal_id, paid_usd, available_after_usd, method, current_status? }`; `status`: `settled` \| `already_settled` \| `not_found` |

### reply_templates
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /reply-templates?limit=&offset=` | `reply_templates_list(limit, offset)` | `ReplyTemplatesList(ctx, limit, offset)` | — | `[]ReplyTemplateListItem` |
| `GET /reply-templates/{id}` | `reply_templates_get(id)` | `ReplyTemplatesGet(ctx, templateID)` | — | `ReplyTemplateFull` |
| `POST /reply-templates` | `reply_templates_create(input)` | `ReplyTemplatesCreate(ctx, input)` | `CreateReplyTemplateInput` | `ReplyTemplateFull` |
| `DELETE /reply-templates/{id}` | `reply_templates_delete(id)` | `ReplyTemplatesDelete(ctx, templateID)` | — | `DeleteReplyTemplateResult { id, public_id }` |
| `GET /reply-templates/{id}/delivery-refs?provider=&providerScope=` | `reply_templates_delivery_refs_list(id, provider, scope)` | `ReplyTemplatesDeliveryRefsList(ctx, templateID, provider, providerScope)` | — | `[]DeliveryRef` |
| `PUT /reply-templates/{id}/delivery-refs` | `reply_templates_delivery_refs_upsert(id, input)` | `ReplyTemplatesDeliveryRefsUpsert(ctx, templateID, input)` | `UpsertDeliveryRefsInput` | `[]DeliveryRef` |

### sales_decks (под URL `/scripts/*`)
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `POST /scripts/price` | `scripts_price(options)` | `ScriptsPrice(ctx, options)` | `{ options: []int }` | `[]PriceMediaItem` |
| `POST /scripts/tools` | `scripts_tools(options, bot_id)` | `ScriptsTools(ctx, options, botID)` | `{ options, bot_id }` | `ToolsMediaResult` |

### servers
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `POST /servers/restart?user_id=&bot_id=` | `servers_restart(user_id, bot_id=1)` | `ServersRestart(ctx, userID, botID)` | — | `ServerRestartResult { message }` |
| `GET /servers/version?user_id=&bot_id=` | — | `ServersVersion(ctx, userID, botID)` | — | `ServerVersionResult { version }` |
| `GET /servers/status?user_id=&bot_id=` | — | `ServersStatus(ctx, userID, botID)` | — | `ServerStatusResult { bound, up }` |
| `GET /servers/changelog?version=` | — | `ServersChangelog(ctx, currentVersion)` | — | `ChangelogResult { latest, known, up_to_date, versions[] }` |

### staff
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /staff` | `get_staff()` | `GetStaff(ctx)` | — | `StaffInfo { name?, role?, is_active, access }` |
| `GET /staff/list` | `list_staff()` | `ListStaff(ctx)` | — | `[]StaffListItem { user_id, name, role }` (только user_id > 1000) |

### subscriptions
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `POST /access/add` | `add_access(input)` | `AddAccess(ctx, input)` | `AddAccessInput { user_id?, account_id?, bot_id, action, access?, action_date?, access_end?, payment_id?, ref?, days?, deltas?, idempotency_key? }`: нужен `user_id` или `account_id`; `action`: `add` \| `extend` \| `revoke` \| `refund` \| `remove` \| `custom`; опциональные поля опускаются (`exclude_none` / `omitempty`), `user_id=0` тоже | `AddAccessResult { created, id?, account_id?, user_id, bot_id, action, action_date?, access_end? }` |
| `POST /access/manage` | не сверялось | `ManageAccess(ctx, input)` | `AccessManageInput { user_id?, account_id?, bot_id, op, features?, days?, end?, note?, idempotency_key? }`: нужен `user_id` или `account_id`; `op`: `grant` \| `extend` \| `remove_features` \| `revoke_all` | `AccessManageResult { account_id?, user_id, bot_id, op, action, access, access_end?, crm_access_id? }` |
| `GET /users/{user_id}/subscriptions/history` | `subscriptions_history(user_id)` | `SubscriptionsHistory(ctx, userID)` | - | `SubscriptionsHistoryResult { user_id, account_id?, history: []AccessHistoryItem }`; у строки истории `added` / `removed`: ключи фич, появившиеся и пропавшие относительно предыдущей строки |
| `GET /users/{user_id}/subscriptions/history?account_id=` | не сверялось | `SubscriptionsHistoryForAccount(ctx, userID, accountID)` | - | то же по указанному аккаунту; человек вне аккаунта: 404 `not_found` |
| `POST /users/subscription-state` | не сверялось | `SubscriptionState(ctx, userIDs)` | `{ user_ids: []int64 }` (1..5000, дубли схлопывает CRM) | `[]SubscriptionStateItem { user_id, has_active_subscription, frozen }`; неизвестный id: `false/false` |
| `GET /access/definitions` | `access_definitions()` | `AccessDefinitions(ctx)` | — | `AccessDefinitionsResult { main, poster, categories }` |
| `POST /subscriptions/transfer-link?user_id=&bot_id=` | `subscriptions_transfer_link(user_id, bot_id)` | `SubscriptionsTransferLink(ctx, userID, botID)` | — | `TransferLinkResult` |
| `POST /subscriptions/transfer/redeem` | `subscriptions_transfer_redeem(data)` | `SubscriptionsTransferRedeem(ctx, input)` | `TransferRedeemInput` | `TransferRedeemResult` |

### tasks
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /tasks/active?user_id=` | `get_active_tasks(user_id)` | `GetActiveTasks(ctx, userID)` | — | `ActiveTasksResult { text }` |
| `GET /tasks/types?user_id=&bot_id=` | `tasks_types(user_id, bot_id)` | `TasksTypes(ctx, userID, botID)` | — | `Dict[str, str]` / `map[string]string` |
| `GET /tasks/list?user_id=&bot_id=&task_type=&limit=&offset=` | `tasks_list(...)` | `TasksList(ctx, ...)` | — | `[]TaskListItem` |
| `GET /tasks/info?user_id=&bot_id=&task_type=&task_id=` | `tasks_info(...)` | `TasksInfo(ctx, ...)` | — | `TaskInfoResult { text }` |
| `GET /tasks/log?...` (бинарный файл) | `tasks_log(...) → TaskLogResult` | `TasksLog(ctx, ...) → *TaskLogResult` | — | байты + Content-Disposition filename |

### users
| Endpoint | Python | Go | Request | Response |
|---|---|---|---|---|
| `GET /users?bot_id=&limit=&offset=` | `list_users(bot_id, limit, offset)` | `ListUsers(ctx, botID, limit, offset)` | — | `ListUsersResult` |
| `GET /users/{user_id}` | `get_user(user_id)` | `GetUser(ctx, userID)` | - | `GetUserResult { user_id, full_name?, username?, status?, has_active_subscription, frozen, bots_info: []UserBotInfo, buyer?: UserBuyer, accounts: []UserAccount }`; `bots_info[].account_id?`; `UserBuyer { buyer_id, email?, display_name?, ref_code?, refer?, landing?, created_at }`; `UserAccount { account_id, title, role, access?, access_expiry?, access_end?, is_personal, is_primary, owner_buyer_id, created_at, joined_at, members_count, chats_count, bots_count }` |
| `GET /users/{user_id}/accounts/{account_id}` | не сверялось | `GetUserAccount(ctx, userID, accountID)` | - | `UserAccountCard { account_id, title, is_personal, is_primary, created_at, owner_buyer_id, role, access?, access_end?, access_expiry?, members[], chats[], bots[], promo_pending[] }`; человек вне аккаунта: 404 `not_found` (`APIError`) |
| `POST /users` | `create_user(input)` | `CreateUser(ctx, input)` | `CreateUserInput` | `CreateUserResult` (идемпотентно) |
| `PUT /users/{user_id}` | `update_user(user_id, input)` | `UpdateUser(ctx, userID, input)` | `UpdateUserInput` | `UpdateUserResult` |
| `POST /users/{user_id}/access/extend?bot_id=&days=` | `extend_user_access(user_id, bot_id, days)` | `ExtendUserAccess(ctx, userID, botID, days)` | без тела: продлевается основной аккаунт | `ExtendAccessResult { user_id, account_id?, access_end? }` |
| `POST /users/{user_id}/access/extend?bot_id=&days=` | не сверялось | `ExtendUserAccessForAccount(ctx, userID, botID, days, accountID)` | `{ account_id }` | то же; человек вне аккаунта: 404 `not_found` |
| `POST /users/{user_id}/ai-limit/extend?millions=` | `extend_ai_limit(user_id, millions)` | `ExtendAILimit(ctx, userID, millions)` | — | `ExtendAiLimitResult` |
| `POST /users/{user_id}/ai-tokens/grant?tokens=&function=&ref=&bot_id=` | не сверялось | `GrantAITokens(ctx, userID, tokens, function, ref, botID)` | - (всё в query; `bot_id` обязателен в SDK) | `GrantAITokensResult { granted, account_id, function, tokens, previous_ai_limit, ai_limit, balance_tokens, unlimited }`; `previous_ai_limit` / `ai_limit` у CRM SocialTraff всегда `null` (в Go нули) |
| `GET /users/{user_id}/ai-usage` | не сверялось | `AIUsage(ctx, userID)` | - (без `bot_id`: бота выбирает CRM) | `AIUsageResult { bot_id, account_id, by_function, daily, monthly, recent, key_stats? }`; у функции `unlimited`, `admin_spent_tokens`, `admin_spent_usd`, `admin_generations`; у генерации `source?`; `AIKeyStats { mask, usage_usd, limit_usd?, limit_remaining_usd?, disabled }` |
| `GET /users/{user_id}/ai-usage?bot_id=&days=&months=&recent=` | не сверялось | `AIUsageWithOptions(ctx, userID, AIUsageOptions{BotID, Days, Months, Recent})` | в query уходят только поля больше нуля; границы CRM: `days` 1..180, `months` 1..24, `recent` 1..500 | то же |

## Auth flow

Оба SDK реализуют идентично:

1. `POST /api/staff/{staff_id}/auth` с заголовком `X-Service-Token: <service_token>`.
2. Кэш `(token, expires_at)` под ключом `(base_url, staff_id)`.
3. При `401` от защищённого эндпоинта — refresh токена один раз + повтор запроса.
4. Per-key lock против гонок при одновременном refresh.

## Retry policy

Оба SDK после правок 2026-05-20 используют единые правила:

- Транспортные ошибки (соединение, EOF, таймауты) ретраятся для **любого** HTTP метода.
- Статус-коды `{429, 502, 503, 504}` ретраятся только для **идемпотентных** методов (`GET/HEAD/OPTIONS/TRACE/PUT/DELETE`).
- `POST`/`PATCH` на эти статусы по умолчанию **не** ретраятся (можно включить через `retry_non_idempotent=True` в Python или `RetryNonIdempotent: true` в Go).
- `500` не входит в дефолтный список — чтобы не маскировать реальные баги повторами.
- Backoff экспоненциальный с jitter.

## Маппинг ошибок

Оба SDK после правок 2026-05-20 единообразно:

- `401`/`403` → `AuthError(code, status, message)`.
- `400`/`422` или `code=VALIDATION_ERROR` → `ValidationError(code, status, message)`.
- остальные не-success → `ApiError(code, status, message)` / `APIError`.

Все три исключения имеют атрибут `code`, заполняемый из envelope.

`activation_redeem`, `subscriptions_transfer_link`, `subscriptions_transfer_redeem`
ловят `(ApiError, AuthError, ValidationError)` и превращают известные
бизнес-коды в `Result(success=False, error_code=...)` вместо exception.

## Nullable-sensitive поля

Поля, где сервер реально может вернуть `null`, и где SDK обязан сохранить
эту nullability в публичной модели (не превращая в `"None"` / `""` / `0`).

| Endpoint | Field | Server gives null when | Python type | Go type |
|---|---|---|---|---|
| `POST /api/activation/redeem` | `payment_id` | activation_code без `payment_id` | `Optional[int]` | `*int64` |
| `GET /api/dialogs/{dept}/search` | `dialogs[].status` | в департаменте без default_status | `Optional[str]` | `*string` |
| `GET /api/dialogs/{dept}/search` | `dialogs[].status_color` | то же | `Optional[str]` | `*string` |
| `POST /api/dialogs/status` | `status` | clear-кейс | `Optional[str]` | `*string` |
| `POST /api/users` (идемп. путь) | `full_name` | nullable DB column | `Optional[str]` | `*string` |
| `POST /api/users` (идемп. путь) | `username` | nullable | `Optional[str]` | `*string` |
| `GET /api/users` (list) | `items[].full_name` | nullable DB column | `Optional[str]` | `*string` |
| `GET /api/users/{id}` | `full_name`/`username`/`status` | nullable | `Optional[str]` | `*string` |
| `GET /api/staff` | `name`/`role` | nullable DB column | `Optional[str]` | `*string` |
| Все timestamps в reply_templates / delivery-refs | `lastUsedAt`/`createdAt`/`updatedAt` | never used / before persistence | `Optional[datetime]` | `*time.Time` |
| `GET /api/payments/invoice/{uuid}` | `client_email`/`referer_id`/`staff_id`/`pay_link`/`pay_url` | nullable | `Optional[T]` | `*T` |
| `GET /api/payments/invoice/{uuid}` | `payment_method` | non-platega провайдеры | `Optional[str]` | `*string` |
| `GET /api/payments` | `items[].*` (аналогично) | nullable | `Optional[T]` | `*T` |
| `POST /api/activation/redeem` | `account_id` | повтор погашенного кода по старой строке доступа без аккаунта | не сверялось | `*int64` |
| `GET /api/users/{id}` | `buyer` | человек не покупатель продукта | не сверялось | `*UserBuyer` |
| `GET /api/users/{id}` | `buyer.email` / `display_name` / `ref_code` / `refer` / `landing` | не заполнено | не сверялось | `*string` |
| `GET /api/users/{id}` | `bots_info[].account_id` | у человека нет аккаунта | не сверялось | `*int64` |
| `GET /api/users/{id}`, `GET /api/users/{id}/accounts/{account_id}` | `access` / `access_expiry` | у аккаунта нет активного доступа; `access_expiry` ещё и у строк без карты сроков | не сверялось | `map[string]bool` / `map[string]string` (nil) |
| `GET /api/users/{id}`, `GET /api/users/{id}/accounts/{account_id}` | `access_end` | нет активного доступа | не сверялось | `*time.Time` |
| `GET /api/users/{id}/accounts/{account_id}` | `members[].tg_id` | покупатель без привязанного Telegram | не сверялось | `*int64` |
| `GET /api/users/{id}/accounts/{account_id}` | `members[].email` / `display_name` | не заполнено | не сверялось | `*string` |
| `POST /api/access/add`, `POST /api/access/manage` | `user_id` | вызов только по `account_id` | не сверялось | `int64` (null даёт 0, см. Intentional differences) |
| `POST /api/access/add`, `POST /api/access/manage`, `POST /api/users/{id}/access/extend` | `account_id` | CRM без этого ключа | не сверялось | `*int64` |
| `GET /api/users/{id}/subscriptions/history` | `account_id` | у человека нет аккаунта | не сверялось | `*int64` |
| `GET /api/users/{id}/ai-usage` | `key_stats` | ключа ещё нет или провайдер не ответил | не сверялось | `*AIKeyStats` |
| `GET /api/users/{id}/ai-usage` | `key_stats.limit_usd` / `limit_remaining_usd` | ключ без лимита (ноль значил бы «исчерпан») | не сверялось | `*float64` |
| `GET /api/users/{id}/ai-usage` | `recent[].source` | CRM без этого ключа | не сверялось | `*string` |
| `GET /api/referrals/info` | `ref_bot_link` | у покупателя нет реф-кода | не сверялось | `*string` |
| `POST /api/referrals/withdraw/request` | `withdrawal_id` / `amount_minor` / `amount_usd` / `method` / `available_minor` / `available_usd` / `min_minor` / `min_usd` | поле не относится к статусу ответа (ключ отсутствует) | не сверялось | `*int64` / `*float64` / `*string` |
| `POST /api/referrals/withdraw/settle` | `current_status` | любой статус, кроме `already_settled` | не сверялось | `*string` |
| `GET /api/bot-blocks` | `kind` | запрос без фильтра | не сверялось | `*string` |

## Wire-format request body

Опциональные поля **не отправляются** (вместо явного `null`) в обоих SDK:

- Python: `model_dump(exclude_none=True)` или `model_dump(mode="json", exclude_none=True)` — `add_access`, `create_invoice_draft`.
- Go: `omitempty` на поле struct.

Сервер принимает оба варианта (Pydantic-модель с `Optional[X] = None`), но
единая wire-форма упрощает диагностику и логи.

Исключение: `POST /api/dialogs/status` всегда шлёт **явный** `status_id: null`
для clear-кейса (как Python, так и Go) — это паритет в обоих направлениях.

Go, с 2026-10-09:

- `POST /api/access/add` и `POST /api/access/manage`: `user_id` опускается при
  нуле. CRM отвергает `user_id: 0`, а вызов только по `account_id` законен.
- `POST /api/users/{id}/access/extend`: `ExtendUserAccess` идёт без тела,
  тело `{ "account_id": N }` шлёт только `ExtendUserAccessForAccount`.
- `GET /api/users/{id}/ai-usage`: `bot_id` и остальные параметры уходят, только
  если заданы в `AIUsageOptions`.

## Intentional differences

Намеренные расхождения между SDK; не баги.

1. **`clear_dialog_status`** — в Python это алиас `change_dialog_status(uid, None)`,
   в Go — отдельная функция `ClearDialogStatus`. Wire-формат идентичен.
2. **ID-типы** — Python `int`, Go `int64`. Сериализация JSON одинакова.
3. **TransferLinkResult / TransferRedeemResult / ActivationRedeemResult** —
   в обоих SDK все поля для success-кейса формально опциональны (Python через
   `Optional[T]`, Go через non-pointer scalars + `omitempty`). При `success=False`
   их zero-value не имеет смысла — нужно ориентироваться на `success`+`error_code`.
4. **POST /api/users request body** — Python шлёт явные `username: null`/`refer: null`
   если поля не заданы (`model_dump()` без `exclude_none`), Go опускает (`omitempty`).
   Сервер принимает оба варианта. Унификация желательна, но не блокирующая.
5. **PaymentHistoryItem / InvoiceInfoResult — `description`/`provider`** —
   объявлены `Optional` в обоих SDK, хотя DB-колонка non-null. Защитное расширение
   nullability — не баг, перестраховка от будущих изменений сервера.
6. **`refund_payment`** на сервере сейчас возвращает 501 (заморожен); оба SDK его
   проксируют без специальной обработки.
7. **Локальная валидация** (sales_decks options 0..4, max 5; reply-template kinds
   `single|album`, item types) хардкожена в обоих SDK как константы. Совпадают
   между SDK и сервером.
8. **Go `AddAccessResult.UserID` / `AccessManageResult.UserID` остались `int64`**,
   хотя при вызове только по `account_id` CRM отдаёт `user_id: null`. Тип не
   менялся ради совместимости: null даёт 0, а адресата в этом случае несёт
   `AccountID`.
9. **Go `UserAccount.Access` / `UserAccountCard.Access` типизированы как
   `map[string]bool`**, а `UserBotInfo.Access` остался `any`. У аккаунта значение
   приводится по истинности, как это делает CRM: у строк доступа без карты
   сроков там бывает не булево значение, и оно не должно ронять разбор карточки.

## Закрытые parity-баги (2026-05-20)

В порядке исправления:

1. **Python `clear_dialog_status` шёл без `status_id` в body** — сервер принимал, но Go SDK явно слал `null`. Выровнено: оба SDK теперь шлют явный `null`. Файлы: `crm_api/client/dialogs.py`.
2. **Python retry policy ретраил только транспортные ошибки** — Go ретраит ещё и 5xx. Python переписан без `tenacity`, добавлен ретрай на `{429, 502, 503, 504}` для идемпотентных методов. Файлы: `crm_api/client/base.py`.
3. **Go `defaultRetryStatusCodes` не включал 429** — добавлено. Файлы: `crmapi/config.go`.
4. **Go `ChangeStatusResult.Status string`** — терял различие `null` vs `""`. Стал `*string`. Файлы: `crmapi/models_dialogs.go`, `crmapi/dialogs.go`.
5. **Python `AuthError`/`ValidationError` теряли `code` сервера** — методы `activation_redeem`/`subscriptions_transfer_*` ловили только `ApiError` и пробрасывали бизнес-ошибки 400/403 исключением. Добавлены `code`/`status` атрибуты в `AuthError`/`ValidationError`; методы ловят все три класса. Файлы: `crm_api/exceptions.py`, `crm_api/client/base.py`, `crm_api/client/activation.py`, `crm_api/client/subscriptions.py`.
6. **Go `ReplyTemplateListItem`/`ReplyTemplateFull` без json-тегов на `LastUsedAt`/`CreatedAt`/`UpdatedAt`** — re-marshal давал `LastUsedAt` вместо `lastUsedAt`. Добавлены теги. Файлы: `crmapi/models_reply_templates.go`.
7. **Python `add_access` слал `access: null`** — Go опускал. Выровнено: оба SDK теперь опускают опциональные поля. Файлы: `crm_api/client/subscriptions.py`.
8. **Python `get_payments` имел недостижимую legacy-ветку** — её срабатывание уронило бы клиент (`data.get('limit')` на list). Удалена; добавлены безопасные `or ""` против null-coercion в string-полях. Файлы: `crm_api/client/payments.py`.
9. **`CreateUserResult.full_name` и `ListUserItem.full_name` non-nullable** — идемпотентный путь сервера может вернуть `null`. Стали `Optional[str]` / `*string`. Файлы: `crm_api/models/users.py`, `crm_api/client/users.py`, `crmapi/models_users.go`, `crmapi/users.go`.
10. **`DialogSearchItem.status` / `status_color` non-nullable** — поведенчески давали `"None"` в Python. Стали `Optional[str]` / `*string`. Файлы: `crm_api/models/dialogs.py`, `crm_api/client/dialogs.py`, `crmapi/models_dialogs.go`, `crmapi/dialogs.go`.
11. **Go `ActivationRedeemResult.PaymentID int64`** — `null` от сервера декодировался в `0`, неотличимо от валидного payment_id. Стал `*int64`. Файлы: `crmapi/models_activation.go`, `crmapi/activation.go`.
12. **Go `DeliveryRef.LastUsedAt/CreatedAt/UpdatedAt *string`** — сырая ISO-строка вместо `*time.Time`. Стали `*time.Time` для паритета с Python `Optional[datetime]`. Файлы: `crmapi/models_reply_templates.go`, `crmapi/reply_templates.go`.

## Изменения 2026-10-09 (Go SDK, CRM SocialTraff)

Подписка в CRM SocialTraff принадлежит аккаунту, а не человеку. Правки дают
Go SDK все поля карточки человека и все параметры, которые CRM принимает.
Экспортированные сигнатуры и типы существующих полей не менялись.

1. **Карточка человека.** `GetUserResult.Buyer`, `GetUserResult.Accounts`,
   `UserBotInfo.AccountID`; новый `GetUserAccount`. Файлы: `crmapi/users.go`,
   `crmapi/models_users.go`.
2. **`account_id` в доступе.** `AddAccessInput.AccountID` / `IdempotencyKey`,
   действие `ActionRemove`, `AccessManageInput.AccountID`; `user_id` не
   обязателен при заданном аккаунте; `account_id` в ответах `AddAccess`,
   `ManageAccess`, `ExtendUserAccess`, `SubscriptionsHistory`,
   `ActivationRedeem`; новые `ExtendUserAccessForAccount` и
   `SubscriptionsHistoryForAccount`; `added` / `removed` в строке истории.
   Файлы: `crmapi/subscriptions.go`, `crmapi/requests_subscriptions.go`,
   `crmapi/models_subscriptions.go`, `crmapi/users.go`, `crmapi/activation.go`.
3. **`AIUsage` больше не шлёт `bot_id=1`.** Изменение поведения: раньше метод
   всегда подставлял бота 1, у CRM SocialTraff бот один и это 10. Теперь бота
   выбирает CRM; явный бот и глубина задаются `AIUsageWithOptions`. Вызывающий,
   которому нужен именно бот 1, передаёт `AIUsageOptions{BotID: 1}`. В отчёте
   добавлены `account_id`, `key_stats`, `unlimited`, `admin_*`, `source`.
   Файлы: `crmapi/ai_usage.go`.
4. **`GrantAITokens`:** `account_id` и `unlimited` в ответе. Файлы:
   `crmapi/models_users.go`.
5. **Рефералы:** `ref_bot_link`, суммы заявки в центах, статус `below_min`,
   `current_status` у проведения вывода. Файлы: `crmapi/referrals.go`,
   `crmapi/models_referrals.go`.
6. **Блокировки бота:** `ListBotBlocksByKind`, `kind` и `counts` в списке,
   `ignored` в report. Файлы: `crmapi/bot_blocks.go`.

## Breaking changes от этой итерации

Совместимость для пользователей SDK:

1. **Python `AuthError(message)` теперь принимает `code` и `status`** — сигнатура расширена с keyword-аргументами, дефолты `None`. Не ломает старых вызывающих.
2. **Python `ValidationError(message)` — то же**, расширение сигнатуры.
3. **Python `CreateUserResult.full_name: str → Optional[str]`** — потребители, которые делали `result.full_name.lower()`, должны добавить None-check.
4. **Python `ListUserItem.full_name: str → Optional[str]`** — то же.
5. **Python `DialogSearchItem.status: str → Optional[str]`** и `status_color` — то же.
6. **Go `ChangeStatusResult.Status string → *string`** — потребители читают `*res.Status`.
7. **Go `CreateUserResult.FullName string → *string`** и `ListUserItem.FullName` — то же.
8. **Go `DialogSearchItem.Status / StatusColor string → *string`** — то же.
9. **Go `ActivationRedeemResult.PaymentID int64 → *int64`** — потребители читают `*res.PaymentID`.
10. **Go `DeliveryRef.LastUsedAt / CreatedAt / UpdatedAt *string → *time.Time`** — потребители получают распарсенное время, а не сырую ISO-строку.

Миграционный риск низкий — изменения единообразны и затрагивают только
nullable-поля, которые на практике редко читаются без None-check.

## Команды проверки

Подтверждают паритет на 2026-05-20:

```bash
# Python SDK (запускать из корня пакета CRM-API-SDK)
.venv/Scripts/python -m pytest \
    --ignore=tests/test_extend_ai_limit_run.py \
    --ignore=tests/test_new_endpoints_run.py \
    --ignore=tests/test_payments_api_run.py \
    --ignore=tests/test_scripts_tools_run.py

# Go SDK (запускать из корня модуля crm-api-go-sdk)
go vet ./...
go test -count=1 ./...
gofmt -l crmapi
```

`*_run.py` в Python SDK — интеграционные тесты с фикстурой `client`, требующие
живого CRM сервера. В unit-прогон не входят, не связаны с правками.

Линтеры (`ruff`/`mypy`/`black` в Python venv, `staticcheck` для Go) на момент
правок не установлены в окружении разработки — отдельные проверки не запускались.
Если будут добавлены — обновите этот документ.
