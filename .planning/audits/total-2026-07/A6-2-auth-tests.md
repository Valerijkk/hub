# W6 агент A2 — auth remaining tests

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: assert-free · telegram login attempts must NOT exist (ADR-AUTH-01) · magiclink runes · fail-closed middleware.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/testframework/auth/repository_organization_test.go` | 121 | 0 | чисто |
| `backend/testframework/auth/repository_test.go` | 73 | 0 | чисто |
| `backend/testframework/auth/resend_telegram_test.go` | 184 | 1 | замечания |
| `backend/testframework/auth/service_login_test.go` | 135 | 0 | чисто |
| `backend/testframework/auth/service_organization_test.go` | 97 | 2 | замечания |
| `backend/testframework/auth/service_profile_test.go` | 144 | 0 | чисто |
| `backend/testframework/auth/service_registration_race_test.go` | 53 | 0 | чисто |
| `backend/testframework/auth/service_registration_test.go` | 86 | 0 | чисто |
| `backend/testframework/auth/service_session_test.go` | 448 | 0 | чисто |
| `backend/testframework/auth/service_test.go` | 214 | 1 | замечания |
| `backend/testframework/auth/telegram_bot_coverage_test.go` | 26 | 1 | замечания |
| `backend/testframework/auth/telegram_bot_polling_test.go` | 136 | 1 | замечания |
| `backend/testframework/auth/telegram_widget_test.go` | 97 | 1 | замечания |

Файлов в списке: 13. Прочитано: 13. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W6-A2-001 | P1 | `telegram_bot_polling_test.go:57-88` | `TestTelegramBotAuth_HandleLogin_StoresCode` гоняет `/login` → `handleLogin` и **требует** ключ `tglogin:` в Redis | ADR-AUTH-01 / TZ: Telegram — не login. Прод-коммент: polling disabled; JWT только Magic Link. Позитивный тест login-кода **запрещён** фокусом волны («telegram login attempts must NOT exist») и держит мёртвый login-путь зелёным | Удалить тест (и ideally `handleLogin` + export). Оставить/усилить только denial: polling/handlers не пишут `tglogin:` (как `:130`) |
| W6-A2-002 | P2 | `resend_telegram_test.go:76-184`; `telegram_bot_coverage_test.go:21-25`; `telegram_bot_polling_test.go:90-112` | Suite `VerifyCode_*` / `HandleLogin_Shutdown*` / `GenerateTGCode` / `RecordVerifyFailure` — покрытие **legacy Telegram login** (`TelegramBotAuth.VerifyCode` «сохранён для тестов») | В `handler`/`wiring` нет вызовов `VerifyCode`/`handleLogin`. Тесты = единственные потребители login-бота → circular coverage мёртвого auth | Удалить login-suite; nil-receiver / StartPolling_Disabled оставить как AUTH-01. Bind — в `notification` |
| W6-A2-003 | P2 | `telegram_widget_test.go:45-63`; `service_test.go:18-98` | `TestVerifyTelegramAuth_LoginWidget` + WebApp `VerifyTelegramAuth_*` — crypto Login Widget / initData | `VerifyTelegramAuth` в prod **нигде не вызывается** (только определение + тесты). Login Widget = вход через Telegram → противоречит AUTH-01 | Удалить тесты и мёртвый `telegram.go` **или** перенести crypto в notification-bind с явным non-login контрактом + тесты bind, не «LoginWidget» |
| W6-A2-004 | P3 | `service_organization_test.go:52-74` | `PendingUsersCRUD`: после Approve/Reject только `NoError`, без статуса/удаления | Approve без `StatusActive` / Reject без `IsNotFound` зелёный при no-op service | Как в `repository_organization_test.go:49-58`: re-get + Equal Active; Reject → NotFound |
| W6-A2-005 | P3 | `service_organization_test.go:76-89` | `PurgeExpiredPending` сеет только fresh, ждёт `n==0` | Не проверяет purge просроченных (7 дней). Сломанный cutoff → всё равно PASS | Как repo-тест: expired `PendingSince` −8d → `n==1`, fresh остаётся |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **assert-free** | Во всех 13 файлах у каждого `Test*` есть `assert.*` / `require.*` и/или `t.Fatal`/`t.Error`. Строго assert-free **нет**. |
| `repository_organization_test.go` | Pending→approve/reject, invite token, purge expired (−8d), CountOrganizations — asserts на status/IDs/counts; слой **реальный Ent**. |
| `repository_test.go` | ExistsByEmail, UpsertSetting, CreateFirstOrMemberUser founder, IsNotFound. |
| `service_login_test.go` | Только Magic Link: pending → `ErrAccountNotActive` + 0 sessions; success JWT; invalid code; existing user. **Telegram login отсутствует** (правильно). |
| `service_session_test.go` | Hash refresh (не plaintext Redis key), rotate, logout, ownership revoke, deactivated/expired/hash mismatch, single-session, NATS publish. |
| `service_profile_test.go` | CRUD профиля; `TelegramID` в response = **bind-поле профиля**, не login. |
| `service_registration_*` | invite_only default, CheckRegistrationAllowed, race upsert, second user = member. |
| `telegram_bot_polling_test.go:114-131` | `StartPolling_Disabled`: `Empty` keys — **положительный AUTH-01** («must not store login codes»). |
| `resend_telegram_test.go` (Resend часть) | SendMagicLinkEmail success/APIError — email Magic Link, не Telegram login. |

### Вердикт по фокусу

| Чек-лист TASK-008 / фокус A2 | Статус |
|------------------------------|--------|
| assert-free тест | **Не найдено** |
| telegram login attempts must NOT exist | **Нарушено:** W6-A2-001 (явный `/login`→код), W6-A2-002/003 (VerifyCode + Login Widget / WebApp verify) |
| magiclink runes | **Вне среза A2** (канон: `handler_magiclink_runes_test.go` у A1). В remaining tests rune-кейсов нет — не дефект этого списка |
| fail-closed middleware | **Вне среза A2** (канон: `middleware_failclosed_test.go` у A1). В 13 файлах middleware-тестов нет |
| Тест на моке там, где guard в репозитории | **Не нарушено** — org/pending/login/session бьют Ent/miniredis, не мок поверх repo-guard |
| Тест, который пройдёт и на сломанном коде | **Да:** W6-A2-004, W6-A2-005 |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | 13 / 13 / 0 |
| Находок | **5** |
| P0 | 0 |
| P1 | 1 |
| P2 | 2 |
| P3 | 2 |

### Топ-3

1. **W6-A2-001** — позитивный `HandleLogin_StoresCode` прямо тестирует Telegram login (`tglogin:`), запрещённый ADR-AUTH-01.  
2. **W6-A2-002** — suite `VerifyCode`/`handleLogin` держит legacy login-бота «для тестов» без prod-потребителя.  
3. **W6-A2-003** — Login Widget / WebApp `VerifyTelegramAuth` тесты на мёртвом crypto (не в wiring).

P0 нет. Magic-link login + session/pending repo — в порядке; главный риск среза — **тесты, которые нормализуют Telegram-as-login** вместо denial-only.
