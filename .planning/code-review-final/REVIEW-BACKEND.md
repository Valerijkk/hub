---
review: backend-go-final
date: 2026-05-15
scope: backend/ (post Phase 14 + 6 audit-fix waves)
files_reviewed: ~120 production .go files
status: issues_found
score: 78/100
---

# Backend Code Review — Post Phase 14

**Контекст:** обзор после закрытия 41 finding (CR-01..LO-08), SRP-разделения AuthService/CallService и выноса `*ent.Client` из 3 сервисов. Ниже — **только новые/незакрытые** проблемы.

## Score: 78 / 100

Сильные стороны: graceful shutdown через `rootCtx`, SafeDialer для webhook SSRF, sliding-window Redis-лимитер, worker-pool для webhook-доставки, корректная WS-сериализация через канал, sync.Once для idempotent Stop. Слабости — ниже.

---

## CRITICAL

### C-01. SSRF в GitLab-интеграции (`http.DefaultClient`)
**File:** `backend/internal/gitlab/service.go:265,367,453,508`
**Проблема:** `integration.GitlabURL` — user-controlled (берётся из `CreateIntegrationRequest.GitLabURL` без валидации `safehttp.ValidateOutboundURL`). Используется `http.DefaultClient.Do(req)` — без SafeDialer. Атакующий с правом создать GitLab-интеграцию может указать `http://169.254.169.254/...` (AWS metadata), `http://localhost:6379`, `http://10.0.0.1` и получить ответы наружу.
**Рекомендация:** заменить `http.DefaultClient` на `webhookHTTPClient` (или вынести общий `safehttp.SharedClient`); валидировать `GitLabURL` в `CreateIntegration`/`UpdateIntegration` через `ValidateOutboundURL`. То же самое надо проверить в GitHub Enterprise URL (если поддерживается).

### C-02. Send-on-closed-channel race в WS Hub
**File:** `backend/internal/ws/hub.go:200-210,229-238,281-284`
**Проблема:** `BroadcastToRoom`/`SendToUser` вызываются из NATS-bridge горутин (`bridge.go`) и шлют в `client.send`. `removeClient` (горутина `Run`) закрывает `client.send` через `closeOnce`. Между `members.Range` и моментом `case client.send <-` в bridge-горутине Run может выполнить `removeClient` → close(send) → `send on closed channel` → паника. Закрытие отправляющей стороной — антипаттерн Go: канал должен закрывать только отправитель.
**Рекомендация:** не закрывать `client.send` в `removeClient`; вместо этого использовать только `close(client.done)` как сигнал writePump'у выйти. Альтернатива — обернуть отправку в `defer recover()` или ввести RWMutex per client.

### C-03. Race на первого admin при регистрации
**File:** `backend/internal/auth/service_login.go:107-115` (+ `service_registration.go`)
**Проблема:** `IsFirstUser(ctx)` → `db.User.Create().SetSystemRole(role)` — между двумя запросами нет блокировки. Две параллельные регистрации в свежей системе могут обе получить `SystemRoleAdmin`. На первой "горячей" установке это даёт привилегий двум аккаунтам.
**Рекомендация:** оборачивать в Tx с `SELECT ... FOR UPDATE` или использовать atomic check-and-set: попытка `INSERT ... WHERE NOT EXISTS (SELECT FROM users)` с retry. Минимум — установить advisory lock в Postgres на ключе `first_user_bootstrap`.

### C-04. Timing attack в magic-link verification
**File:** `backend/internal/auth/magiclink.go:132`
**Проблема:** `if stored != code` — обычное string-сравнение, не constant-time. Brute-force атакующий с измерением CPU-time может постепенно угадать символ-за-символом. HI-04 lockout (5 попыток) помогает, но при `magicLinkVerifyMax >= 3` всё равно мерится время первых попыток.
**Рекомендация:** `subtle.ConstantTimeCompare([]byte(stored), []byte(code)) == 1`. То же проверить в `telegram_bot.go:VerifyCode`.

---

## HIGH

### H-01. Unbounded `io.ReadAll(resp.Body)` для внешних API
**File:** `backend/internal/gitlab/service.go:272,374,460,515`; `backend/internal/auth/resend.go:73`
**Проблема:** ответ от GitLab/Resend читается полностью в память без лимита. Скомпрометированный/злоумышленный GitLab-инстанс (см. C-01) или злонамеренный SMTP-провайдер могут вернуть 10 GiB и положить процесс.
**Рекомендация:** `io.ReadAll(io.LimitReader(resp.Body, 1<<20))` (1 МБ хватает для ошибок).

### H-02. `respBuf.ReadFrom(resp.Body)` без лимита в webhook
**File:** `backend/internal/apigateway/webhook.go:275`
**Проблема:** комментарий говорит про `maxWebhookResponseBodyBytes = 4096`, но `ReadFrom` читает **всё тело** и только потом обрезает строку. Эндпоинт, возвращающий 1 ГБ, выест 1 ГБ RAM на каждом ретрае. Также ошибка `ReadFrom` молча игнорируется.
**Рекомендация:** `io.Copy(&respBuf, io.LimitReader(resp.Body, maxWebhookResponseBodyBytes))`; обработать err.

### H-03. Personal Dashboard возвращает чужие звонки
**File:** `backend/internal/dashboard/service.go:114-118`
**Проблема:** `s.db.Call.Query().Where().WithChannel().Limit(10).All(ctx)` — пустой `Where()` тянет 10 произвольных звонков **всей системы** без фильтра по `userID` и без `Status == "active"`. Утечка метаданных по соседним проектам/каналам.
**Рекомендация:** фильтр `call.HasParticipantsWith(user.ID(userID))` + `call.Status("active")` + сортировка по `CreatedAt`.

### H-04. Goroutine spawn-per-request в API-key authn
**File:** `backend/internal/apigateway/apikey.go:126-134`
**Проблема:** `go func() { ... UpdateOneID(...).Exec(bgCtx) }()` стартует горутину на **каждый** аутентифицированный запрос. Под нагрузкой 10K rps это 10K горутин/сек на запись `last_used_at`, без бэкпрешера → connection pool exhaustion → каскадный сбой Postgres.
**Рекомендация:** buffered-канал + 1-2 воркера, как в `PushSender`/`TelegramSender`. Дополнительно — дебаунс по `ak.ID` (обновлять не чаще раз в минуту).

### H-05. `/health` открывает новый raw DB-connection на каждый запрос
**File:** `backend/cmd/server/main.go:316-328`
**Проблема:** `database.OpenRawDB(cfg.Postgres.URL)` + `rawPgDB.Close()` на каждом hit. k8s liveness probe (1/сек) = 86400 коннектов/день, лишних handshake. Risk: исчерпание `pg_hba` слотов под массовой проверкой.
**Рекомендация:** держать общий `*sql.DB` (уже есть `rawDB` для search — переиспользовать) или брать pinger из Ent driver: `db.Driver().(*entsql.Driver).DB().PingContext(ctx)`.

### H-06. GDPR-удаление молча проглатывает ошибки
**File:** `backend/internal/audit/service.go:309,324,332,335,339`
**Проблема:** `_, _ = s.repo.DeleteUserReactions(...)`, то же для activity-logs / time-logs / audit-log create. GDPR Art.17 требует verifiable erasure — silent failure означает, что пользователь "удалён" по UI, но реакции/тайм-логи остались в БД. Юридический риск.
**Рекомендация:** собирать ошибки в `errs []error`, возвращать после `HardDeleteUser` (если есть failures — `errors.Join(errs...)` или хотя бы Error-log + статус "partial").

### H-07. Recording retention не удаляет S3-объект
**File:** `backend/internal/audit/retention.go:131-144`
**Проблема:** `Recording.Delete().Where(CreatedAtLT)` удаляет только метаданные в Postgres. MP4-файлы в S3/SeaweedFS остаются вечно → orphaned data, storage cost growth, GDPR-violation (retention policy декларирована, но реально не применена).
**Рекомендация:** перед `Recording.Delete` — пройтись `s3Client.DeleteObject(Bucket, FilePath)` (или batch `DeleteObjects`). Залогировать failures отдельно.

---

## MEDIUM

### M-01. `err == redis.Nil` вместо `errors.Is`
**File:** `auth/magiclink.go:125`, `auth/telegram_bot.go:151`, `auth/service_session.go:96`, `auth/middleware.go:110`, `messenger/presence.go:50,80,139`, `imex/service.go:82,119` (8 мест)
**Проблема:** `err == redis.Nil` сломается, если go-redis обернёт ошибку (например, в pipeline-результате или при wrapping). Защита от регрессии при апгрейде драйвера — `errors.Is`.
**Рекомендация:** `errors.Is(err, redis.Nil)` повсеместно. То же для `io.EOF`.

### M-02. `time.Sleep(50ms)` в воркере не уважает ctx
**File:** `backend/internal/notification/push.go:73`; `backend/internal/notification/telegram.go:65`
**Проблема:** при graceful shutdown воркер залипает на 35-50мс перед проверкой `<-workerCtx.Done()`. Не критично, но при cascading shutdown ×N воркеров это десятки мс.
**Рекомендация:** `select { case <-time.After(50*time.Millisecond): case <-workerCtx.Done(): return }`.

### M-03. `sendMessage` в Telegram без context
**File:** `backend/internal/notification/telegram.go:96-106`
**Проблема:** `t.bot.SendMessage(params)` без передачи ctx — нельзя прервать HTTP-запрос при shutdown. Telego поддерживает `BotWithCtx` или per-call ctx.
**Рекомендация:** передавать `workerCtx` в `sendMessage(ctx, msg)` и далее в telego.

### M-04. Variadic `entDB ...*ent.Client` — анти-паттерн SRP
**File:** `backend/internal/github/service.go:38`, `backend/internal/gitlab/service.go:35`
**Проблема:** SC-14-3 убрал `*ent.Client` из Messenger/Audit/Search, но GitHub/GitLab сервисы всё ещё хранят `entDB *ent.Client` для прямых TMS-обновлений (`service.go:530`). Variadic — для "опциональности" — скрывает обязательную зависимость. Это незакрытая часть SC-14-3.
**Рекомендация:** вынести TMS-обновления в `TMSRepository` interface, передать в конструктор явно. Удалить variadic.

### M-05. Search worker errs[0] вместо errors.Join
**File:** `backend/internal/search/service.go:154-156`
**Проблема:** 7 параллельных searchXxx-воркеров — если 3 упали, видно только `errs[0]`. Дебаг проблем с tsvector станет хуже.
**Рекомендация:** `slog.Warn(..., "errs", errors.Join(errs...))`.

### M-06. `_, _ := s.repo.MemberCount(ctx, projectID)` 6 мест
**File:** `backend/internal/project/service.go:82,119,132,145,269`; `invite.go:111`
**Проблема:** ошибка `MemberCount` молча трактуется как `count=0` → проверка лимита членов проекта может пройти, когда реально проект пустой из-за упавшего запроса. Не критично, но идиоматически плохо.
**Рекомендация:** хотя бы `if err != nil { slog.Warn(...); return ... }` или `count, err := ...; if err != nil { ... }`.

### M-07. `purgeOldMessages` без `index` на created_at — N×deletes
**File:** `backend/internal/audit/retention.go:97`
**Проблема:** `coll.DeleteMany({"created_at": {"$lt": cutoff}})` — если индекса `created_at_1` нет в `messages`, retention-job сканирует всю коллекцию ежедневно. Проверьте `mongo_repository.EnsureIndexes`.
**Рекомендация:** убедиться, что индекс есть; рассмотреть TTL-индекс `expireAfterSeconds` — Mongo удалит сам.

### M-08. `service_polls.go:65` — `msgID, _ := s.repo.InsertMessage`
**File:** `backend/internal/messenger/service_polls.go:65`
**Проблема:** ошибка вставки сообщения-опроса в MongoDB проглочена. Опрос создаётся в Postgres, но в чате не появится — UX-баг без следа в логах.
**Рекомендация:** `if err != nil { slog.Error(...); }` минимум. Не блокировать создание опроса, но залогировать.

### M-09. RateLimit cleanup-горутины без stop-канала
**File:** `backend/internal/platform/middleware/ratelimit.go:48,65,255-285`
**Проблема:** `go limiter.cleanup()` — горутина с `time.NewTicker(5*time.Minute)` и `for range ticker.C` без ctx. При интеграционных тестах `RegisterRoutes` создаёт N лимитеров → N невытаскиваемых горутин. Не leak в проде (1 регистрация), но в тестах накапливается.
**Рекомендация:** принимать ctx в `RateLimit(ctx, config)` или возвращать `Cleanup() func()` из конструктора.

---

## LOW

### L-01. `dashboard/service.go:107`: `fmt.Sscanf(v, "%d", &n)` без обработки ошибки
**Проблема:** если Redis вернёт "broken" значение, `n=0`, total получит мусор. Очень редко.
**Рекомендация:** `strconv.Atoi(v)` с обработкой.

### L-02. `bridge.go` — `b.subs` без mutex
**File:** `backend/internal/ws/bridge.go:38,98`
**Проблема:** `Start()` append'ит в `subs`, `Stop()` диапазонит и обнуляет. Если Start/Stop вызывают из разных горутин — race. На практике сейчас вызывается синхронно из main, но контракт нерушимый не задокументирован.
**Рекомендация:** sync.Mutex или явный комментарий "не thread-safe".

### L-03. `audit_log` 5×bool/timestamp полей — code smell
**File:** `backend/internal/notification/consumer.go:197-203, 245-251, ...` (повторяющийся `Dispatch(...)` с 7-8 позиционными аргументами)
**Проблема:** `Dispatch(ctx, userID, type, title, body, entityType, entityID, projectID, actorID, isHighPrio)` — 10 параметров, легко перепутать порядок (особенно `*uuid.UUID, *uuid.UUID, bool`). DRY: 7 однотипных обработчиков в `consumer.go`.
**Рекомендация:** `type DispatchInput struct { ... }` и `Dispatch(ctx, in DispatchInput)`. Заодно вынести логику "проверь UUID-поля → если ошибка, log+return" в хелпер `parseEventUUIDs(evt)`.

### L-04. `service_test.go` не покрывает обновлённый CallService
**File:** `backend/internal/calls/service.go` (247 строк)
**Заметка:** после SC-14-2 split'а unit-тесты `service_test.go` тестируют agglomerator, но `CallRoomService`/`CallEventPublisher`/`CallChatService` сами по себе — без тестов (видно по `recording_webhook_test.go` фокусируется только на webhook).
**Рекомендация:** добавить таблицы тестов для `CallRoomService.GenerateToken` (валидация TTL/прав), `CallEventPublisher.Started/Ended` (асссерт payload в fake NATS).

### L-05. `ent_repository.go` (TMS) — 932 строки = один god-репозиторий
**File:** `backend/internal/tms/ent_repository.go`
**Заметка:** SC-14-1/14-02 разнесли сервисы по SRP, но TMS репозиторий — единственный god-объект, оставшийся в коде. Если хочется консистентности, разбить на `repo_test_cases.go`/`repo_runs.go`/`repo_suites.go` по аналогии с `tracker/repo_issues.go`/`repo_boards.go`/...
**Рекомендация:** не блокер; запланировать в Phase 15.

### L-06. `cmd/server/main.go` — 629 строк
**Заметка:** wire-код всех модулей в одной `main()`. Phase 14-03 добавил репозиторий-разделения, что увеличило bootstrap. Идиоматически — выделить `internal/app/wire.go` с `BuildApp(cfg) (*App, error)`.
**Рекомендация:** не блокер; рефакторинг под `wire`/manual DI.

### L-07. `service_session.go:31` — конструктор с 6 позиционными аргументами
**File:** `backend/internal/auth/service_session.go:31`
**Проблема:** `NewSessionService(db, rdb, nc, jwtSecret, accessTTL, refreshTTL)` — легко перепутать accessTTL/refreshTTL (оба time.Duration).
**Рекомендация:** `SessionConfig` struct.

### L-08. `gitlab/service.go` имеет hardcoded `https://gitlab.com`
**File:** `backend/internal/gitlab/service.go:51`
**Проблема:** дефолт прибит гвоздями. Для self-hosted GitLab — окей, но это magic-string.
**Рекомендация:** константа `defaultGitLabURL` рядом с issueKeyRegex.

---

## INFO

- **`*-converters.go`** (tracker, tms) — большие mapper-файлы (319/389 строк). Это OK для DRY, но рассмотреть кодген (через `ent` template) если их рост продолжится.
- **`SystemSetting` lookup в `getRetentionDays`** делается 3 раза за `RunRetentionCleanup` — мелочь, но можно одним запросом + map.
- **`io.ReadAll(io.LimitReader(...))`** хелпер — стоит вынести в `platform/http` пакет, использовать везде где сейчас `io.ReadAll(resp.Body)`.
- **CallService:247 строк** — на грани, но в рамках SRP-цели "<250". OK.
- **MessengerService разбит на 5 файлов** (service.go + service_channels/messages/polls/channel_invites) — идёт хорошо, ent_repository.go (572) — но это репозиторий-агрегат, его роль допускает размер.

---

## Сводка по фокусу

| Категория          | Critical | High | Medium | Low | Info |
|--------------------|----------|------|--------|-----|------|
| Bugs               | 2 (C-02, C-03) | 1 (H-03) | 2 (M-06, M-08) | 1 (L-01) | – |
| SOLID/DRY          | –        | –    | 1 (M-04) | 3 (L-05/06/07) | 2 |
| Error handling     | 1 (C-04) | 2 (H-06, H-07) | 4 (M-01/05/08, M-06) | – | – |
| Concurrency        | 1 (C-02) | 1 (H-04) | 2 (M-02/03/09) | 1 (L-02) | – |
| Resource leaks     | –        | 2 (H-01/02, H-05) | 1 (M-07) | – | 1 |
| Security           | 1 (C-01) | – | – | – | – |

## Итог: 78/100

**-22 балла за:** 4 критикала (SSRF в gitlab, WS-race, admin-race, timing-attack), 7 hi-severity (unbounded reads, ресурсные leaks, dashboard-leak, GDPR-failure), 9 mid-severity.

**+78 за:** добротный Phase 14 (SRP-сплиты выполнены аккуратно), graceful shutdown, SafeDialer в webhook, Redis sliding-window, worker-pool в WebhookDispatcher, nil-safe sender'ы, sync.Once в Stop, корректная сериализация WS Hub через канал, использование `errors.Is(ent.IsNotFound)` повсеместно, переключение `_ := redis.Get` на `errors.Is` уже частично сделан.

**Что добить в первую очередь:**
1. C-01 (SSRF GitLab) — security-критикал, патч ~30 строк.
2. C-02 (WS send-on-closed) — потенциальная prod-паника, патч ~10 строк.
3. C-04 (timing attack) — `subtle.ConstantTimeCompare`, одна строка.
4. H-04 (apikey unbounded goroutines) — переиспользовать паттерн из push.go.
5. H-06 (GDPR errors) — legal compliance.

---

_Reviewed: 2026-05-15_
_Reviewer: Claude (gsd-code-reviewer, Opus 4.7 1M)_
_Depth: standard + targeted deep checks_
