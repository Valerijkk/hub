# Наблюдаемость: структурное логирование Go

> Часть плейбука. Канон и оглавление — [`../observability.md`](../observability.md).

slog, уровни, поля, логирование ошибок.

---

## 2. Структурированное логирование Go (slog)

### 2.1 Единственный логгер — `log/slog`

В Hub стандарт — стандартная библиотека `log/slog`. Она инициализируется один раз в `backend/cmd/server/wiring_bootstrap.go`:

```16:24:backend/cmd/server/wiring_bootstrap.go
func initLogger() {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "hub-backend"
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", serviceName)
	slog.SetDefault(logger)
}
```

Ключевые следствия для всех разработчиков:

- Хендлер — **JSON**, вывод в **stdout**. Логи собираются оркестратором контейнеров (Docker/K3s → Loki), не пишутся в файлы приложением.
- Уровень по умолчанию — **`LevelInfo`**. `Debug` в проде по умолчанию **не виден** (см. §8 про управление уровнем).
- Поле `service` добавлено глобально через `.With(...)`. Не дублируйте его в каждом вызове.
- Логгер установлен как **default** (`slog.SetDefault`), поэтому в любом пакете можно вызывать `slog.InfoContext(...)` без передачи логгера через DI.

Запрещено создавать свои `slog.Logger` с другим форматом/выводом в модулях. Единый формат — обязательное условие для парсинга в Loki/Grafana.

### 2.2 Структурированность: пары ключ-значение, а не форматная строка

Главный принцип: **сообщение — константа, переменные — в полях**. Так логи машиночитаемы, фильтруемы и агрегируемы.

Good — сообщение стабильно, данные в полях (как в `.cursor/rules/10-style.mdc`):

```go
// message остаётся константой -> легко грепать и агрегировать в Loki
slog.InfoContext(ctx, "issue created",
	"issue_id", issue.ID,
	"project_id", projectID,
	"module", "tracker",
)
```

Bad — конкатенация и форматирование прямо в сообщении:

```go
// каждое сообщение уникально: нельзя сгруппировать, нельзя отфильтровать по полю
slog.Info("issue " + issue.ID.String() + " created in project " + projectID.String())
slog.Info(fmt.Sprintf("issue %s created", issue.ID)) // то же самое, только через Sprintf
```

Почему это важно: в Loki/Grafana запрос `{module="tracker"} | json | issue_id="..."` работает только если `issue_id` — отдельное поле. Если оно вклеено в текст, найти все события по конкретной задаче можно лишь регуляркой по подстроке — медленно и ненадёжно.

### 2.3 Уровни: Debug / Info / Warn / Error

`slog` имеет четыре базовых уровня. Смысловая матрица для Hub:

| Уровень | Когда использовать | Видно в prod по умолчанию | Примеры из кода |
|---|---|---|---|
| `Debug` | Детальная диагностика, шумные события жизненного цикла | Нет (Level=Info) | `slog.Debug("ws client registered", ...)` (`ws/hub.go`) |
| `Info` | Значимое бизнес-событие, успешное завершение операции | Да | `slog.InfoContext(ctx, "operation completed", ...)` (`observability/timing.go`) |
| `Warn` | Аномалия, которую система пережила (деградация, дроп, ретрай) | Да | `slog.Warn("ws client send buffer full, dropping message", ...)` (`ws/hub.go`) |
| `Error` | Операция провалилась, требуется внимание | Да | `slog.ErrorContext(ctx, "operation failed", ..., "error", err)` (`observability/timing.go`) |

Правила выбора уровня:

- **`Debug`** — для того, что интересно только при активной отладке. В `ws/hub.go` регистрация/подписка клиента логируется на `Debug`, потому что при 50 пользователях это тысячи событий — в проде они бы забили лог.
- **`Info`** — для событий, которые вы хотите видеть в проде **всегда**: создан ресурс, отправлено сообщение, завершена фоновая задача. Не злоупотребляйте: Info на каждой итерации цикла — это тот же шум (см. §11).
- **`Warn`** — система справилась, но что-то не так: буфер WS переполнен и сообщение отброшено, сработал fallback AI-провайдера, включён `DEV_MODE`. Warn — сигнал «посмотри, если станет часто».
- **`Error`** — конкретная операция не выполнена. Error **обязан** нести `error`-поле с исходной ошибкой и достаточно контекста, чтобы понять «что именно не удалось».

Bad — паника уровней (всё Error или всё Info):

```go
// НЕЛЬЗЯ: рутинное событие на Error -> ложные срабатывания алёртов error_rate
slog.Error("user opened dashboard")
// НЕЛЬЗЯ: реальная потеря данных на Info -> утонет в шуме, никто не заметит
slog.Info("failed to persist message", "error", err)
```

### 2.4 Контекстное логирование: всегда `...Context(ctx, ...)`

Если в области видимости есть `context.Context` — используйте `slog.InfoContext`/`WarnContext`/`ErrorContext`/`DebugContext`, а не `slog.Info` без ctx.

Причины:

1. **Отмена/дедлайны**: контекстные методы уважают отменённый контекст и не выполняют лишнюю работу при завершении.
2. **Трейс-корреляция**: при активном OpenTelemetry (§5) `ctx` несёт trace/span id. Контекстные методы позволяют хендлеру связать лог с трейсом.
3. **Единый стиль**: `.cursor/rules/10-style.mdc` показывает канонический пример именно с `slog.InfoContext(ctx, ...)`.

Good — сервис получает ctx первым аргументом (по правилу «context always first») и логирует с ним:

```go
func (s *Service) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	issue, err := s.repo.Create(ctx, req)
	if err != nil {
		// ctx несёт request_id/trace -> лог склеивается с запросом
		slog.ErrorContext(ctx, "create issue failed",
			"module", "tracker",
			"project_id", req.ProjectID,
			"error", err,
		)
		return nil, fmt.Errorf("create issue: %w", err)
	}
	slog.InfoContext(ctx, "issue created", "module", "tracker", "issue_id", issue.ID)
	return issue, nil
}
```

Bad — ctx есть, но не передан в лог:

```go
func (s *Service) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	issue, err := s.repo.Create(ctx, req)
	// теряем trace-контекст; лог невозможно склеить с трейсом запроса
	slog.Info("issue created", "issue_id", issue.ID)
	return issue, err
}
```

Исключение, когда `ctx` **отсутствует** легитимно: код без доступа к запросу (например, регистрация клиента в WS-хабе, background-тикер до создания per-job контекста). Там допустим `slog.Info/Debug` без ctx — как в `ws/hub.go`. Но если ctx доступен — он обязателен.

### 2.5 Обязательные поля

Каждое значимое событие должно нести достаточно полей для фильтрации. Обязательный минимум в Hub:

| Поле | Тип | Когда обязательно | Источник |
|---|---|---|---|
| `request_id` | string | Всё, что происходит в рамках HTTP-запроса | `c.Locals("requestID")` |
| `module` | string | Всегда (домен: `tracker`, `messenger`, `auth`, `calls`, `tms`, `ws`, ...) | вручную |
| `user_id` | string/uuid | Действие от имени пользователя | из auth-контекста |
| `entity_id` | string/uuid | Операция над сущностью (`issue_id`, `channel_id`, `message_id`, ...) | доменный объект |
| `error` | error | Всегда на уровне `Error`/`Warn` при провале | `err` |
| `duration_ms` | int64 | Замер длительности операции (§4) | `time.Since(start).Milliseconds()` |

Замечание про именование `entity_id`: предпочитайте **конкретное** имя (`issue_id`, `channel_id`), а не абстрактное `entity_id` — это точнее для фильтров. Абстрактный `entity_id` уместен в общих утилитах (audit, outbox), где тип сущности переменный и есть отдельное поле `entity_type`.

Good — полный контекст:

```go
slog.InfoContext(ctx, "message sent",
	"module", "messenger",
	"channel_id", channelID,
	"message_id", msg.ID,
	"user_id", userID,
)
```

Bad — событие без опознавательных полей:

```go
// какое сообщение? в каком канале? от кого? -> лог бесполезен при разборе
slog.InfoContext(ctx, "message sent")
```

### 2.6 Что НЕ логировать (безопасность)

Абсолютный запрет (нарушение = security-инцидент, блокер merge):

| Категория | Примеры | Почему нельзя |
|---|---|---|
| Пароли/секреты | пароли, `INTEGRATION_ENCRYPTION_KEY`, webhook-секреты, приватные ключи | Утечка учёток; логи часто менее защищены, чем БД |
| Токены | JWT access/refresh, magic-link токены, API-ключи, Telegram bot token | Захват сессии/аккаунта |
| PII | email целиком, телефон, полное ФИО в связке с идентификаторами | GDPR/152-ФЗ; корзина/soft-delete не спасает логи |
| Тела с секретами | request body логина, смены пароля, bind-запросов | Пароль/токен утечёт вместе с payload |
| Query string с токеном | `/ws?token=<JWT>` | Реальный вектор ME-03 (см. ниже) |
| Полные заголовки | `Authorization`, `Cookie`, `X-Api-Key` | Содержат креды |

Good — логируем ссылку на объект, а не сам секрет:

```go
// логируем факт и идентификатор пользователя, не сам токен
slog.InfoContext(ctx, "magic link issued",
	"module", "auth",
	"user_id", userID,
	// НЕ логируем сам token/email в открытом виде
)
```

Bad — секрет прямо в логе:

```go
// КАТАСТРОФА: токен утекает в stdout -> Loki -> навсегда в хранилище логов
slog.InfoContext(ctx, "login", "email", email, "magic_token", token)
slog.InfoContext(ctx, "request", "authorization_header", c.Get("Authorization"))
```

Если очень нужно связать событие с чувствительным значением — логируйте **необратимый идентификатор**: последние 4 символа, хэш (например, `sha256` первых байт) или суррогатный `user_id`. Никогда не логируйте достаточную для восстановления часть секрета.

### 2.7 Реальный кейс ME-03: query string в access-логах

В Hub WebSocket-апгрейд использует `/ws?token=<JWT>`. Наивный access-лог записал бы полный путь с токеном. Это закрыто в `backend/internal/platform/middleware/logging.go`:

```11:44:backend/internal/platform/middleware/logging.go
// stripQuery возвращает путь без query-строки.
// Используется в access-логах, чтобы не утекали токены через
// `?token=...` (например, WebSocket-апгрейд: /ws?token=<JWT>).
// См. ME-03.
func stripQuery(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		return p[:i]
	}
	return p
}

// Logger логирует каждый запрос: метод, путь, статус, длительность и request_id через slog.
// Путь логируется без query string во избежание утечки токенов в логи (ME-03).
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		requestID, _ := c.Locals("requestID").(string)

		slog.Info("request",
			"method", c.Method(),
			"path", stripQuery(c.Path()),
			"status", c.Response().StatusCode(),
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"request_id", requestID,
		)

		return err
	}
}
```

Урок: **любой** новый access-лог, аудит-запись или отладочный вывод, включающий URL, обязан прогонять путь через `stripQuery` или иным способом гарантировать отсутствие query string с секретами.

### 2.8 Сквозной correlation id (`request_id`)

`request_id` — это сквозной идентификатор запроса. В Hub он генерируется в `RequestID()` middleware:

```8:22:backend/internal/platform/middleware/requestid.go
// RequestID генерирует уникальный X-Request-ID для каждого запроса и сохраняет в c.Locals.
// Использует uuid.NewString(), возвращающий copy-safe строку (Pitfall 1: переиспользование Fiber RequestCtx).
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals("requestID", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}
```

Поведение и следствия:

- Если клиент прислал `X-Request-ID` — он **переиспользуется** (позволяет связать логи фронта и бэка). Иначе генерируется новый UUID.
- Значение кладётся в `c.Locals("requestID")` и в **ответный** заголовок `X-Request-ID`. Пользователь/фронт видит его.
- При ошибке тот же `request_id` попадает в тело ответа (см. §3) — пользователь может сообщить его в поддержку, и по нему находится вся цепочка логов.

Как пробросить `request_id` в логи сервисного слоя: сервис получает `ctx`. Чтобы `request_id` был в логах глубоко в стеке, извлекайте его из `c.Locals` в handler и добавляйте в поля (либо кладите в `ctx` через свой ключ). Минимально надёжный приём — передавать его явным полем в тех логах, где он критичен:

```go
// handler достаёт request_id и передаёт вниз как поле лога
func (h *Handler) Create(c fiber.Ctx) error {
	requestID, _ := c.Locals("requestID").(string)
	ctx := c.UserContext()
	issue, err := h.svc.CreateIssue(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "create issue failed",
			"module", "tracker",
			"request_id", requestID, // сквозная корреляция с access-логом и ответом
			"error", err,
		)
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(issue)
}
```

Примечание о будущем улучшении (рекомендация, пока не в коде): чтобы не таскать `request_id` руками, можно добавить кастомный `slog.Handler`, который извлекает `request_id`/trace id из `ctx` и подмешивает их в каждый `...Context` вызов автоматически. Это устранит забытые поля. Ввести можно без изменения сигнатур — обёрткой над `slog.NewJSONHandler` в `initLogger`.

### 2.9 Сэмплирование шумных логов

Некоторые события повторяются тысячи раз в секунду (heartbeat WS, presence-обновления, ретраи). Логировать каждое — значит утопить полезные логи и раздуть счёт за хранение.

Стратегии в порядке предпочтения:

1. **Понизить уровень до `Debug`.** Самый простой приём: рутинное событие на `Debug` в проде не пишется (Level=Info). Так сделано в `ws/hub.go` для register/subscribe.
2. **Логировать агрегат, а не каждое событие.** Вместо «отправлено сообщение» × N — «отправлено N сообщений за интервал» одним `Info` по таймеру.
3. **Rate-limit/сэмплинг: 1 из N.** Для предсказуемо шумных `Warn` (например, «буфер полон») логируйте каждое K-е с счётчиком.

Good — сэмплинг счётчиком для шумного Warn:

```go
// логируем каждое 100-е переполнение буфера, но считаем все через метрику
if dropped := atomic.AddInt64(&bufferDropCounter, 1); dropped%100 == 1 {
	slog.WarnContext(ctx, "ws send buffer full (sampled 1/100)",
		"module", "ws",
		"dropped_total", dropped,
	)
}
```

Bad — полный лог на каждое высокочастотное событие:

```go
// при 50 клиентах и heartbeat раз в 5с это забьёт лог и Loki
for _, client := range clients {
	slog.InfoContext(ctx, "heartbeat received", "client_id", client.ID) // шум
}
```

Важно: сэмплирование применяется к **логам**, но **метрика считает всё** (§4). Дроп буфера должен инкрементить счётчик Prometheus на каждом событии, даже если лог пишется 1/100.

### 2.10 Запрет `fmt.Println` / `log.Printf` / `log.Fatal` в prod

Категорический запрет из `.cursor/rules/10-style.mdc`:

| Запрещено | Почему | Замена |
|---|---|---|
| `fmt.Println` / `fmt.Printf` | Не структурировано, без уровня, без времени/service | `slog.InfoContext(ctx, ...)` |
| `log.Printf` / `log.Println` | Стандартный `log` пишет в свой формат, не JSON | `slog.*` |
| `log.Fatal` / `log.Panic` | Убивает процесс без graceful shutdown; в проде = падение сервиса | вернуть ошибку; на старте — `slog.Error` + `os.Exit(1)` |
| `panic()` в хендлерах | Крашит запрос/процесс | вернуть `error` |
| `println()` (builtin) | Debug-мусор в stderr | удалить |

Единственное легитимное место `os.Exit(1)` — **фаза старта** (`bootstrapServer`), когда конфиг невалиден и сервис не может подняться. Там паттерн такой (реальный код):

```31:33:backend/cmd/server/wiring_bootstrap.go
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
```

Это допустимо только **до** начала обслуживания запросов. В рантайме (handler/service/consumer) `os.Exit`/`log.Fatal` запрещены — они не дают отработать graceful shutdown (outbox, cron, WS, consumers), который завязан на единый `rootCtx` (см. §3.5).

Good/Bad сводно:

```go
// GOOD
slog.InfoContext(ctx, "backup finished", "module", "backup", "files", n)

// BAD
fmt.Println("backup finished", n)         // не структурировано
log.Printf("backup finished %d", n)       // не JSON, без ctx
log.Fatalf("backup failed: %v", err)      // убивает процесс в рантайме
```

### 2.11 Язык сообщений логов

Соглашение Hub: **проза документации — русская, код — английский**. Сообщения логов — это код, поэтому предпочтительны **английские** константные сообщения (`"issue created"`, `"operation failed"`). В репозитории встречаются исторические русскоязычные сообщения (например, `slog.Error("не удалось записать активность", ...)` в `tracker/service_issues.go`) — это легаси; **новые** логи пишите по-английски для единообразия и грепа. Не смешивайте языки в пределах одного события.

---

<a name="3-логирование-ошибок"></a>
## 3. Логирование ошибок

### 3.1 Главный принцип: логировать один раз, на границе

Самая частая ошибка наблюдаемости — **дублирование**: одна и та же ошибка логируется на каждом уровне стека, поднимаясь вверх. В результате один сбой даёт 5 записей `Error`, error rate на дашборде завышен, а найти «первопричину» невозможно.

Правило Hub:

> **Ошибку логирует тот, кто её обрабатывает (принимает решение), — обычно граница системы.** Все нижние слои её только **оборачивают** (`%w`) и возвращают.

Граница в Hub — это:

- HTTP: централизованный `ErrorHandler` (см. §3.4) — логирует/отвечает.
- Фоновые задачи: `Timed`/`TimedVoid` (§4) — логирует провал операции один раз.
- Consumers (NATS): обработчик сообщения на верхнем уровне — логирует и решает про ack/nack.

Good — нижние слои оборачивают, верхний логирует:

```go
// repository: только оборачивает, добавляя контекст операции
func (r *entRepo) Create(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	row, err := r.client.Issue.Create().SetTitle(req.Title).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("ent create issue: %w", err) // wrap, НЕ log
	}
	return toModel(row), nil
}

// service: тоже оборачивает бизнес-контекстом, НЕ логирует
func (s *Service) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	issue, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create issue for project %s: %w", req.ProjectID, err)
	}
	return issue, nil
}

// handler/граница: логирует ОДИН раз с полным контекстом и отвечает
func (h *Handler) CreateIssue(c fiber.Ctx) error {
	issue, err := h.svc.CreateIssue(c.UserContext(), req)
	if err != nil {
		slog.ErrorContext(c.UserContext(), "create issue failed",
			"module", "tracker",
			"project_id", req.ProjectID,
			"request_id", reqID(c),
			"error", err, // wrapped chain читается целиком
		)
		return apperrors.NewInternal("failed to create issue")
	}
	return c.Status(fiber.StatusCreated).JSON(issue)
}
```

Bad — лог на каждом слое:

```go
// repository
if err != nil {
	slog.Error("ent create failed", "error", err) // лог #1
	return err
}
// service
if err != nil {
	slog.Error("service create failed", "error", err) // лог #2 — та же ошибка
	return err
}
// handler
if err != nil {
	slog.Error("handler create failed", "error", err) // лог #3 — снова та же
	return err
}
```

### 3.2 Wrap vs log: как выбирать

| Ситуация | Действие | Почему |
|---|---|---|
| Внутренний слой (repo/service), ошибка пойдёт выше | `fmt.Errorf("op: %w", err)` — **wrap** | Сохраняет контекст без дублей логов |
| Граница, дальше ошибка не идёт (или конвертируется в HTTP) | `slog.ErrorContext(...)` — **log** один раз | Здесь принимается решение |
| Ошибка «ожидаемая» (not found, conflict, validation) | Не `Error`; вернуть типизированную `AppError` | Это не сбой системы, а нормальный исход |
| Ошибка проглатывается осознанно (best-effort) | `slog.Warn(...)` с пояснением **почему** проглочена | Явно фиксируем деградацию |

Всегда оборачивайте с `%w` (не `%v`, не `%s`) — это сохраняет цепочку для `errors.Is`/`errors.As`. Правило `.cursor/rules/10-style.mdc`: `fmt.Errorf("create issue: %w", err) // always wrap`.

Good — осознанное проглатывание best-effort с Warn:

```go
// запись активности не критична для основной операции -> не валим запрос,
// но фиксируем Warn, чтобы деградация была видна
if err := s.activity.Record(ctx, ev); err != nil {
	slog.WarnContext(ctx, "activity record skipped",
		"module", "tracker",
		"issue_id", issue.ID,
		"error", err,
	)
}
```

Bad — молчаливое проглатывание:

```go
// ошибка исчезла бесследно: ни лога, ни возврата -> невидимая деградация
_ = s.activity.Record(ctx, ev)
```

### 3.3 Никогда не «log-and-return» одновременно на одном слое

Антипаттерн `log-and-return` — логировать ошибку **и** вернуть её вызывающему, который тоже её залогирует. Это прямой путь к дублям (§3.1).

Правило: на конкретном слое делайте **либо** log (если это граница и дальше не отдаёте), **либо** return-wrap (если отдаёте выше). Не то и другое.

Bad — классический log-and-return:

```go
func (s *Service) DoThing(ctx context.Context) error {
	if err := s.step(ctx); err != nil {
		slog.ErrorContext(ctx, "step failed", "error", err) // логируем...
		return fmt.Errorf("do thing: %w", err)              // ...и отдаём выше, где залогируют снова
	}
	return nil
}
```

Good — либо только wrap...:

```go
func (s *Service) DoThing(ctx context.Context) error {
	if err := s.step(ctx); err != nil {
		return fmt.Errorf("do thing: %w", err) // граница выше залогирует один раз
	}
	return nil
}
```

...либо только log (когда это верхняя граница, например consumer, который сам решает про ack):

```go
func (c *Consumer) handle(ctx context.Context, msg *nats.Msg) {
	if err := c.process(ctx, msg); err != nil {
		// это верхняя граница обработки сообщения: логируем и решаем про nak
		slog.ErrorContext(ctx, "message processing failed",
			"module", "notification",
			"subject", msg.Subject,
			"error", err,
		)
		_ = msg.Nak()
		return
	}
	_ = msg.Ack()
}
```

### 3.4 Централизованный HTTP-обработчик ошибок и формат `{error, request_id}`

Все HTTP-ошибки конвертируются в единый JSON через `ErrorHandler` (`backend/internal/platform/middleware/middleware.go`) и структуру `AppError` (`backend/internal/platform/errors/errors.go`). Это реализация инварианта ТЗ «ответ на ошибки: `{"error", "request_id"}`».

Структура ответа:

```6:18:backend/internal/platform/errors/errors.go
// AppError — структурированная ошибка приложения с HTTP-кодом, сообщением и ID запроса.
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
	Details    any    `json:"details,omitempty"`
	RetryAfter int    `json:"-"` // секунды для заголовка Retry-After (429)
}
```

Обработчик:

```13:41:backend/internal/platform/middleware/middleware.go
// ErrorHandler — централизованный обработчик ошибок Fiber.
// Проверяет тип ошибки (AppError, fiber.Error, generic) и возвращает JSON с кодом, сообщением и request_id.
func ErrorHandler(c fiber.Ctx, err error) error {
	requestID, _ := c.Locals("requestID").(string)

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		appErr.RequestID = requestID
		if appErr.RetryAfter > 0 {
			c.Set("Retry-After", strconv.Itoa(appErr.RetryAfter))
		}
		return c.Status(appErr.Code).JSON(appErr)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(apperrors.AppError{
			Code:      fiberErr.Code,
			Message:   fiberErr.Message,
			RequestID: requestID,
		})
	}

	return c.Status(500).JSON(apperrors.AppError{
		Code:      500,
		Message:   "internal server error",
		RequestID: requestID,
	})
}
```

Следствия для разработчиков:

- В handler возвращайте **типизированные** ошибки: `apperrors.NewNotFound(...)`, `NewBadRequest(...)`, `NewForbidden(...)`, `NewRateLimitedAfter(...)`, `NewInternal(...)`. Обработчик сам проставит `request_id` и HTTP-код.
- **Не** формируйте JSON ошибки вручную в handler — это дублирует логику и рискует потерять `request_id`.
- Для 500 клиенту отдаётся нейтральное `"internal server error"` — **детали и стектрейс идут только в лог**, не в ответ (не раскрываем внутренности). `request_id` связывает нейтральный ответ с подробным логом.
- Для 4xx (`AppError`) сообщение осмысленное и безопасное — его можно показать пользователю.

Good — handler отдаёт типизированную ошибку, логирует только неожиданное:

```go
func (h *Handler) GetIssue(c fiber.Ctx) error {
	issue, err := h.svc.GetIssue(c.UserContext(), id)
	switch {
	case errors.Is(err, tracker.ErrNotFound):
		// ожидаемый исход: не логируем как Error, просто 404
		return apperrors.NewNotFound("issue not found")
	case err != nil:
		// неожиданный сбой: логируем + нейтральный 500
		slog.ErrorContext(c.UserContext(), "get issue failed",
			"module", "tracker", "issue_id", id, "error", err)
		return apperrors.NewInternal("failed to load issue")
	}
	return c.JSON(issue)
}
```

Bad — ручной JSON без request_id и с утечкой внутренностей:

```go
// теряем request_id; отдаём внутреннюю ошибку клиенту (info leak)
return c.Status(500).JSON(fiber.Map{"error": err.Error()})
```

### 3.5 Паника и recover

`panic()` в production-путях запрещён (`.cursor/rules/10-style.mdc`, ТЗ). Но паника может прилететь из зависимости или из недосмотра. Защита — `Recover()` middleware в цепочке (порядок из `.cursor/rules/20-backend.mdc`):

```go
app.Use(middleware.RequestID())
app.Use(middleware.Recover())   // ловит панику ДО того, как она уронит процесс
app.Use(middleware.Logger())
```

Правила:

- `Recover` превращает панику в 500-ответ (через `ErrorHandler`) и **логирует** её с уровнем `Error`, включая стектрейс и `request_id`. Процесс продолжает работу — один запрос упал, сервис жив.
- В своём коде **не восстанавливайтесь молча**: `recover()` в бизнес-логике допустим только если вы логируете панику и превращаете её в нормальную ошибку. «Проглоченная» паника скрывает баг.
- В горутинах, которые вы запускаете сами (WS pump, consumers, cron), оборачивайте тело в защиту от паники, иначе паника в горутине **обрушит весь процесс** (её не поймает HTTP-`Recover`).

Good — защита горутины с логированием и без ресурса-паники:

```go
func (h *Hub) runClientPump(ctx context.Context, cl *client) {
	defer func() {
		if r := recover(); r != nil {
			// без этого паника в отдельной горутине убьёт весь сервер
			slog.ErrorContext(ctx, "ws client pump panic recovered",
				"module", "ws",
				"client_id", cl.id,
				"panic", r,
			)
		}
	}()
	cl.pump(ctx)
}
```

Bad — паника как поток управления и молчаливый recover:

```go
func (s *Service) mustLoad(ctx context.Context, id uuid.UUID) *Entity {
	e, err := s.repo.Get(ctx, id)
	if err != nil {
		panic(err) // НЕЛЬЗЯ: паника вместо возврата ошибки
	}
	return e
}

func worker() {
	defer func() { _ = recover() }() // НЕЛЬЗЯ: паника проглочена без следа
	// ...
}
```

### 3.6 Ошибки graceful shutdown

Все долгоживущие горутины Hub (outbox, cron, notification/audit consumers, WS hub, webhook dispatcher) получают единый `rootCtx`, отменяемый по `SIGINT`/`SIGTERM` (`wiring_bootstrap.go`). При остановке ошибки вида `context.Canceled` — **ожидаемые**, их не логируют на `Error`.

Good — отличаем отмену от реального сбоя:

```go
if err := c.consume(ctx); err != nil {
	if errors.Is(err, context.Canceled) {
		slog.InfoContext(ctx, "consumer stopped", "module", "notification") // норма при shutdown
		return
	}
	slog.ErrorContext(ctx, "consumer crashed", "module", "notification", "error", err)
}
```

Bad — шум `Error` при штатной остановке:

```go
if err := c.consume(ctx); err != nil {
	slog.ErrorContext(ctx, "consumer error", "error", err) // на каждом деплое ложный Error
}
```

---

<a name="4-замер-времени-и-метрики"></a>
