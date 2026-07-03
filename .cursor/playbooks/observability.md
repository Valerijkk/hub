# Hub — Playbook: логирование и наблюдаемость

> Эталонный документ для AI-команды Hub. На него ссылаются `be1`/`be2`/`fe1`/`fe2`/`qa`/`devops`.
> Обновлено: 2026-07-02 · Связанные правила: `.cursor/rules/10-style.mdc`, `.cursor/rules/20-backend.mdc`, `.cursor/rules/30-frontend.mdc`, `.cursor/rules/56-regression-verify.mdc`
> База фактов: `backend/internal/platform/observability/*`, `backend/internal/platform/middleware/*`, `docs/architecture.md`, `docs/stack_and_tools.md`, `docs/TESTING.md`

Этот playbook описывает **как именно** в Hub писать логи, снимать метрики и трассировать запросы. Он основан на **реальном коде** репозитория, а не на абстрактных рекомендациях. Все примеры `Good`/`Bad` привязаны к тому, что уже работает в `backend/internal/platform/observability` и `backend/internal/platform/middleware`.

Правило прочтения: если в тексте написано «есть в коде» — это проверяемый факт (файл указан). Если написано «ввести минимально / рекомендация» — этого пока нет, и раздел объясняет, как добавить, не ломая инварианты ТЗ.

---

## Оглавление

1. [Кто и когда использует. Три столпа наблюдаемости](#1-кто-и-когда-использует)
2. [Структурированное логирование Go (slog)](#2-структурированное-логирование-go-slog)
3. [Логирование ошибок](#3-логирование-ошибок)
4. [Замер времени и метрики (требование)](#4-замер-времени-и-метрики)
5. [Трейсинг (OpenTelemetry)](#5-трейсинг-opentelemetry)
6. [Frontend-наблюдаемость](#6-frontend-наблюдаемость)
7. [Логи в тестах](#7-логи-в-тестах)
8. [Формат логов prod vs dev, уровни, ротация](#8-формат-логов-prod-vs-dev)
9. [Алёрты и дашборды](#9-алёрты-и-дашборды)
10. [Definition of Done по наблюдаемости](#10-definition-of-done)
11. [Антипаттерны логирования](#11-антипаттерны-логирования)
12. [Приложения: шпаргалки и ключи полей](#12-приложения)

---

<a name="1-кто-и-когда-использует"></a>
## 1. Кто и когда использует. Три столпа наблюдаемости

### 1.1 Зачем это вообще нужно

Наблюдаемость (observability) — это свойство системы, при котором по её **внешним выходным данным** (логи, метрики, трейсы) можно понять её **внутреннее состояние**, не подключаясь дебаггером к продакшену. Для Hub это критично: self-hosted корпоративная платформа на 10–50 человек в команде, где падение мессенджера или звонков блокирует работу всей организации, а инцидент нельзя воспроизвести на ноутбуке разработчика.

Наблюдаемость отвечает на три вопроса при инциденте:

- **Что сломалось?** — логи (конкретное событие, стектрейс, `request_id`).
- **Насколько плохо и когда началось?** — метрики (error rate, latency, счётчики).
- **Где именно в цепочке вызовов?** — трейсы (span'ы между HTTP → service → NATS → БД).

Без наблюдаемости отладка инцидента превращается в гадание. С ней — в чтение: берём `request_id` из ответа пользователю, находим все логи запроса, смотрим span длительности, видим, что 800 мс ушло в один медленный запрос к Postgres.

### 1.2 Три столпа

| Столп | Что это | Где в Hub | Основной инструмент |
|---|---|---|---|
| **Logs** (логи) | Дискретные события с контекстом | `log/slog` (JSON stdout) | `slog.InfoContext(ctx, ...)` |
| **Metrics** (метрики) | Числовые агрегаты во времени | Prometheus + in-memory `Snapshot()` | `observability.Timed`, `/metrics` |
| **Traces** (трейсы) | Путь одного запроса через компоненты | OpenTelemetry OTLP (опционально) | `HTTPTraceMiddleware`, `Timed` |

Три столпа **дополняют**, а не заменяют друг друга:

- Метрика говорит «error rate вырос до 5%». Она **дешёвая** (число), хранится долго, годится для алёртов и дашбордов. Но не говорит **почему**.
- Лог говорит «ошибка X при обработке сущности Y, `request_id=Z`». Он **дорогой** (текст), хранится недолго, годится для разбора конкретного случая.
- Трейс говорит «запрос Z прошёл HTTP → auth → tracker.CreateIssue → Postgres, узкое место — вставка активности». Он связывает всё воедино через `request_id`/trace context.

Правильный рабочий цикл инцидента: **алёрт по метрике → находим окно времени → фильтруем логи по модулю/уровню → берём `request_id` конкретной ошибки → (если есть OTel) смотрим трейс**.

### 1.3 Кто использует и когда

| Роль | Когда обращается к playbook | Что обязан сделать |
|---|---|---|
| `be1` / `be2` | Любой новый handler/service/consumer | `slog.InfoContext(ctx, ...)` с обязательными полями; обернуть критичные операции в `Timed`; не логировать секреты |
| `fe1` / `fe2` | Новый route/модуль/error boundary | Обернуть модуль в `RouteErrorBoundary`; `console.*` только dev; не логировать токены |
| `qa` | Написание тестов, разбор падений | `t.Log`/`t.Logf` для диагностики; запись таймингов прогонов; проверка, что логи не текут секретами |
| `devops` | CI/CD, деплой, мониторинг | Настройка `LOG_LEVEL`/`ENVIRONMENT`; `/metrics` allowlist; `docker-compose.monitoring.yml`; алёрты |
| Верификаторы (`be-ver`/`fe-ver`) | Ревью перед merge | Прогон чек-листа §10; поиск антипаттернов §11 |

Момент подключения наблюдаемости — **не «потом»**, а в той же сессии, что и код фичи. По правилу `.cursor/rules/56-regression-verify.mdc` фича без логов/метрик критичного пути не проходит DoD (§10).

### 1.4 Инварианты ТЗ, которые наблюдаемость обязана уважать

Из `.cursor/rules/60-tz-constraints.mdc` и `docs/security_guidelines.md`:

- **UTC в логах и БД всегда.** Локальная TZ — только на фронте. Никаких `time.Local` в серверных логах.
- **`request_id` в каждом ответе об ошибке** — формат `{"error"/"message", "request_id"}` (реально `AppError`, см. §3).
- **Секреты/токены/PII не логируются никогда** — в том числе query string с `?token=` (реализовано через `stripQuery`, см. §2.7).
- **Нет `panic()` в production-путях** — паника ловится `Recover`, а не роняет процесс.
- **WebSocket через Fiber, NATS JetStream** — метрики и логи этих подсистем обязательны (очереди, дропы буфера).

---

<a name="2-структурированное-логирование-go-slog"></a>
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
## 4. Замер времени и метрики (требование)

Наблюдаемость Hub **обязана** включать замер длительности критичных операций и экспорт метрик. Это не опция: без таймингов нельзя проверить SLO (§4.5) и нельзя настроить алёрты по latency (§9).

### 4.1 Базовый паттерн: замер длительности через defer

Минимальный, всегда доступный приём — засечь время в начале и залогировать в `defer`:

```go
func (s *Service) HeavyOp(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	defer func() {
		// duration_ms логируется ВСЕГДА, даже при раннем return/ошибке
		slog.InfoContext(ctx, "heavy op done",
			"module", "tms",
			"entity_id", id,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	}()

	// ... тяжёлая работа ...
	return nil
}
```

Почему `defer`, а не строка в конце: при раннем `return` (ошибка, ветвление) финальный лог с длительностью не выполнится, если он написан просто последней строкой. `defer` гарантирует замер на **любом** пути выхода.

Bad — замер, который теряется при ошибке:

```go
func (s *Service) HeavyOp(ctx context.Context) error {
	start := time.Now()
	if err := s.step(ctx); err != nil {
		return err // duration НЕ залогирован -> не видим латентность неудачных вызовов
	}
	slog.InfoContext(ctx, "done", "dur_ms", time.Since(start).Milliseconds())
	return nil
}
```

Важно: `dur_ms` для **неудачных** операций так же ценен, как для успешных — часто именно медленные таймауты предшествуют ошибке. Замеряйте оба исхода.

### 4.2 Предпочтительный путь: `observability.Timed`

Для критичных операций используйте готовую обёртку `observability.Timed` / `TimedVoid` (`backend/internal/platform/observability/timing.go`). Она делает три вещи разом: **лог длительности + метрика + span трейса**.

```10:37:backend/internal/platform/observability/timing.go
// Timed выполняет fn и логирует длительность в миллисекундах.
// Ошибки логируются на уровне Error, успех — Info.
func Timed[T any](ctx context.Context, operation string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := startOperationSpan(ctx, operation)
	start := time.Now()
	result, err := fn(ctx)
	ms := time.Since(start).Milliseconds()
	recordOperation(operation, ms, err != nil)
	finishOperationSpan(span, err)
	attrs := []any{
		"operation", operation,
		"duration_ms", ms,
	}
	if err != nil {
		slog.ErrorContext(ctx, "operation failed", append(attrs, "error", err)...)
		return result, err
	}
	slog.InfoContext(ctx, "operation completed", attrs...)
	return result, nil
}

// TimedVoid — вариант Timed для функций без возвращаемого значения.
func TimedVoid(ctx context.Context, operation string, fn func(context.Context) error) error {
	_, err := Timed(ctx, operation, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}
```

Что даёт `Timed` автоматически:

1. Открывает child-span OTel (если трейсинг включён) — `startOperationSpan`.
2. Замеряет длительность.
3. Пишет метрику через `recordOperation` (in-memory `Snapshot()` **и** Prometheus histogram/counter).
4. Закрывает span, помечая ошибку при провале.
5. Логирует `operation completed` (Info) или `operation failed` (Error) с `duration_ms`.

Good — критичная операция обёрнута в `Timed`:

```go
func (s *Service) GenerateReport(ctx context.Context, projectID uuid.UUID) (*Report, error) {
	// имя operation стабильно и низкокардинально -> хорошая метка метрики/span
	return observability.Timed(ctx, "tms.generate_report",
		func(ctx context.Context) (*Report, error) {
			return s.buildReport(ctx, projectID)
		},
	)
}
```

Good — `TimedVoid` для операции без результата:

```go
func (w *Worker) flushOutbox(ctx context.Context) error {
	return observability.TimedVoid(ctx, "outbox.flush", func(ctx context.Context) error {
		return w.publishPending(ctx)
	})
}
```

Bad — ручной дубляж того, что уже делает `Timed` (лог + метрика + span вручную, с риском рассинхрона):

```go
func (s *Service) GenerateReport(ctx context.Context) (*Report, error) {
	start := time.Now()
	r, err := s.buildReport(ctx)
	// забыли метрику и span; продублировали то, что даёт Timed
	slog.InfoContext(ctx, "report", "dur_ms", time.Since(start).Milliseconds())
	return r, err
}
```

### 4.3 Именование операций (кардинальность меток)

`operation` становится **меткой** Prometheus (`operation="tms.generate_report"`). Метки с высокой кардинальностью взрывают память Prometheus и делают метрики бесполезными.

Правила именования:

- Формат: `<module>.<action>` в snake_case — `tracker.create_issue`, `messenger.send_message`, `auth.issue_magic_link`.
- **Никаких переменных данных в имени**: не включайте `user_id`, `issue_id`, timestamp. Это создаёт бесконечное число меток.
- Стабильность: одно и то же имя для одной и той же логической операции во всех вызовах.

Good:

```go
observability.Timed(ctx, "messenger.send_message", fn) // конечное множество имён
```

Bad — кардинальность метки бесконечна:

```go
observability.Timed(ctx, "send_message_to_"+channelID.String(), fn) // взрыв меток Prometheus
```

### 4.4 Метрики: что уже собирается

В Hub две системы метрик, обе питаются из `recordOperation` (`observability/metrics.go`):

**A. In-memory агрегат (`Snapshot()`)** — быстрый обзор без внешних систем:

```9:18:backend/internal/platform/observability/metrics.go
// OperationMetric — агрегированная статистика по имени операции (OBS-01).
type OperationMetric struct {
	Operation   string  `json:"operation"`
	Count       int64   `json:"count"`
	ErrorCount  int64   `json:"error_count"`
	TotalMs     int64   `json:"total_ms"`
	MaxMs       int64   `json:"max_ms"`
	AvgMs       float64 `json:"avg_ms"`
	LastSeenUTC string  `json:"last_seen_utc,omitempty"`
}
```

`Snapshot()` возвращает срез по всем операциям (count, error_count, avg, max). Полезно для отладочного эндпоинта/health-дашборда без Prometheus.

**B. Prometheus** (`observability/prometheus.go`) — для дашбордов и алёртов. Зарегистрированные метрики:

| Метрика | Тип | Метки | Смысл |
|---|---|---|---|
| `hub_operation_calls_total` | Counter | `operation`, `result` (success/error) | Число вызовов `Timed()` |
| `hub_operation_duration_milliseconds` | Histogram | `operation` | Длительность операций (buckets 1..5000мс) |
| `hub_http_requests_total` | Counter | `method`, `route`, `status` | Число HTTP-запросов |
| `hub_http_request_duration_milliseconds` | Histogram | `method`, `route` | Латентность HTTP (DefBuckets) |

HTTP-метрики снимаются в `HTTPMetricsMiddleware` (по `c.Route().Path`, не по конкретному пути — снова забота о кардинальности: `/api/issues/:id`, а не `/api/issues/123`).

Счётчики (Counter) — для «сколько раз»: запросы, ошибки, отправленные сообщения. Гистограммы (Histogram) — для распределения длительностей, из которых Prometheus считает p50/p95/p99 (§4.6).

### 4.5 Эндпоинт `/metrics` и его защита

Prometheus-scrape доступен на `GET /metrics` (`MetricsHandler`). Он защищён CIDR-allowlist (`MetricsEndpoint` / `MetricsAllowMiddleware`):

```71:89:backend/internal/platform/observability/prometheus.go
// MetricsEndpoint — /metrics с опциональным CIDR allowlist (OBS-01).
func MetricsEndpoint(allowedCIDRs []string) fiber.Handler {
	scrape := MetricsHandler()
	nets := parseCIDRs(allowedCIDRs)
	denyAll := len(allowedCIDRs) > 0 && len(nets) == 0

	return func(c fiber.Ctx) error {
		if len(allowedCIDRs) > 0 {
			if denyAll {
				return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
			}
			ip := net.ParseIP(c.IP())
			if ip == nil || !ipAllowed(ip, nets) {
				return fiber.NewError(fiber.StatusForbidden, "metrics endpoint forbidden")
			}
		}
		return scrape(c)
	}
}
```

Правила для `devops`:

- В проде **обязательно** задать allowlist (CIDR сети Prometheus/мониторинга). Пустой список = открыто всем — допустимо только в изолированной dev-сети.
- `/metrics` исключён из access-лога и трейс/метрик middleware (`if c.Path() == metricsPath { return c.Next() }`), чтобы scrape не засорял собственные метрики. При добавлении новых middleware учитывайте это исключение.

### 4.6 p50 / p95 / p99 и целевые SLO

Средняя длительность (`avg`) **лжёт**: один медленный запрос из тысячи не поднимет среднее, но испортит опыт этого пользователя. Поэтому мерим **перцентили**.

- **p50** (медиана) — типичный пользователь.
- **p95** — «почти худший» опыт; основной SLO-показатель Hub.
- **p99** — хвост; ловит редкие деградации, таймауты, GC-паузы.

Из гистограмм Prometheus перцентили считаются через `histogram_quantile`:

```promql
# p95 латентности HTTP по маршруту за 5 минут
histogram_quantile(0.95,
  sum(rate(hub_http_request_duration_milliseconds_bucket[5m])) by (le, route)
)

# p99 длительности бизнес-операции
histogram_quantile(0.99,
  sum(rate(hub_operation_duration_milliseconds_bucket[5m])) by (le, operation)
)
```

**Целевые SLO Hub** (источник — `docs/TESTING.md`, раздел производительности):

| Показатель | Цель | Метрика для проверки |
|---|---|---|
| API response time (p95) | **< 200 мс** | `hub_http_request_duration_milliseconds` |
| WebSocket message latency | **< 100 мс** | замер доставки в WS-слое |
| Kanban board load (100 задач) | **< 500 мс** | `tracker.*` операции + фронт |
| Messenger: загрузка канала (1000 сообщений) | **< 300 мс** | `messenger.*` операции + фронт |

Правило: если операция входит в один из SLO выше — она **обязана** быть обёрнута в `Timed` (или иметь HTTP-метрику через middleware), иначе SLO нечем проверять.

### 4.7 Где мерить

| Место | Как | Что покрывает |
|---|---|---|
| HTTP-граница | `HTTPMetricsMiddleware` (авто) | Латентность/счётчики всех запросов по маршруту |
| Критичные бизнес-операции | `observability.Timed(ctx, "module.action", fn)` | Отчёты, импорт/экспорт, генерация, тяжёлые выборки |
| Фоновые задачи | `TimedVoid` в теле job | outbox flush, cron-задачи, консьюмеры |
| Внешние вызовы | `Timed` вокруг вызова интеграции | GitHub/GitLab/Telegram/LiveKit/SMTP латентность |
| WS-доставка | ручной `start/defer` с `dur_ms` | latency доставки сообщения (SLO < 100 мс) |

Не мерьте **всё подряд**: тривиальные getter'ы и мгновенные операции не нуждаются в `Timed` — это лишний шум логов Info. Мерьте то, что: (а) входит в SLO, (б) обращается к БД/сети, (в) исторически бывало медленным.

### 4.8 Запись таймингов прогонов тестов

Тайминги — это не только рантайм, но и тесты. По `docs/TESTING.md` и `.cursor/rules/56-regression-verify.mdc` при регрессии фиксируйте длительность прогонов, чтобы отслеживать деградацию CI и находить медленные тесты.

Практика для `qa`:

- Прогон бэкенда: `go test -race -count=1 -timeout=600s ./...` — Go сам печатает время по пакетам; логируйте суммарное время в `.planning/COVERAGE.md`/отчёт сессии при заметном росте.
- Для точечного профилирования медленного теста используйте `-v` и `t.Log` с временем этапов (§7).
- Целевые пороги производительности из `docs/TESTING.md` (§4.6) проверяются перф-тестами; их результаты — часть DoD (§10).

Артефакты покрытия/профилей **не коммитим** (`.cursor/rules/45-artifacts-cleanup.mdc`): пишите профили в системный temp, не в `backend/`.

---

<a name="5-трейсинг-opentelemetry"></a>
## 5. Трейсинг (OpenTelemetry)

### 5.1 Что уже есть в коде

Hub **уже** имеет OpenTelemetry-трейсинг (OTel v1.40.0 по `docs/stack_and_tools.md`), реализованный в `backend/internal/platform/observability/otel.go`. Он **опциональный**: включается только при заданном OTLP-эндпоинте.

Инициализация (`InitTracing`) регистрирует global `TracerProvider` **только** если задан `OTEL_EXPORTER_OTLP_ENDPOINT` (или `cfg.OTLPEndpoint`):

```42:49:backend/internal/platform/observability/otel.go
func InitTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
	endpoint := strings.TrimSpace(cfg.OTLPEndpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	}
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}
```

Ключевые факты:

- Экспорт по **OTLP HTTP** (`otlptracehttp`). Эндпоинт — коллектор (Grafana Tempo/Jaeger/OTel Collector).
- Пропагация — `TraceContext` + `Baggage` (W3C). Это значит, что `traceparent`-заголовок принимается от клиента и пробрасывается дальше.
- Флаг `tracingActive` (atomic) хранит, включён ли реальный провайдер. `TracingEnabled()` возвращает его. Если трейсинг выключен, span'ы **не создаются** (нулевой оверхед).
- Имя сервиса: `cfg.ServiceName` → `OTEL_SERVICE_NAME` → `hub-backend`.
- Возвращается функция `Shutdown`, которую bootstrap регистрирует в cleanup с таймаутом 5 сек (graceful flush трейсов при остановке).

### 5.2 Серверные span'ы: `HTTPTraceMiddleware`

Каждый HTTP-запрос оборачивается в server-span:

```97:127:backend/internal/platform/observability/otel.go
// HTTPTraceMiddleware — server span на запрос и propagation context.
func HTTPTraceMiddleware() fiber.Handler {
	tracer := otel.Tracer(httpTracerName)
	prop := otel.GetTextMapPropagator()

	return func(c fiber.Ctx) error {
		if c.Path() == metricsPath {
			return c.Next()
		}

		carrier := fiberHeaderCarrier{c: c}
		ctx := prop.Extract(c.UserContext(), carrier)

		route := c.Path()
		if r := c.Route(); r != nil && r.Path != "" {
			route = r.Path
		}
		spanName := c.Method() + " " + route

		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
			attribute.String("http.target", c.Path()),
		)
		if requestID, ok := c.Locals("requestID").(string); ok && requestID != "" {
			span.SetAttributes(attribute.String("hub.request_id", requestID))
		}

		c.SetUserContext(ctx)
```

Важные детали:

- Span называется `METHOD /route` по **шаблону** маршрута (низкая кардинальность), не по конкретному пути.
- В span кладётся `hub.request_id` — так трейс склеивается с логами и ответом пользователю. Это мост между тремя столпами.
- `prop.Extract` вытаскивает входящий trace-контекст: если запрос пришёл от другого сервиса/фронта с `traceparent`, span станет частью распределённого трейса.
- `c.SetUserContext(ctx)` — важно: дальше по стеку `c.UserContext()` уже несёт span. Поэтому `Timed` (§4.2), вызванный внутри handler, создаст **child**-span автоматически.

### 5.3 Child-span'ы бизнес-операций

`Timed`/`TimedVoid` создают дочерние span'ы через `startOperationSpan`, но **только** если трейсинг включён:

```178:195:backend/internal/platform/observability/otel.go
// startOperationSpan — child span при активном tracing.
func startOperationSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	if !TracingEnabled() {
		return ctx, nil
	}
	return otel.Tracer(operationTracerName).Start(ctx, operation)
}

func finishOperationSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
```

Следствие: **обёртка `Timed` — это и метрика, и лог, и span одновременно.** Обернув критичную операцию в `Timed`, вы бесплатно получаете узел в трейсе. Отдельно вызывать OTel API почти никогда не нужно.

Good — вложенность даёт полезный трейс:

```go
// HTTPTraceMiddleware создал server-span для запроса;
// Timed добавит child-span "tracker.create_issue" внутри него
func (s *Service) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	return observability.Timed(ctx, "tracker.create_issue",
		func(ctx context.Context) (*Issue, error) {
			return s.repo.Create(ctx, req) // выполняется внутри child-span
		},
	)
}
```

### 5.4 Пропагация в исходящих вызовах

Для входящих запросов пропагация уже настроена. Для **исходящих** (Hub → GitHub/GitLab/внешний сервис) при необходимости распределённого трейса нужно инжектить контекст в заголовки запроса:

```go
// пробрасываем trace-контекст во внешний вызов, чтобы трейс не рвался
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
resp, err := client.Do(req)
```

Напоминание безопасности: для user-controlled URL используйте `safehttp.NewSafeHTTPClient` и `ValidateOutboundURL` (SSRF), а не `http.DefaultClient` — это ортогонально трейсингу, но обязательно.

**NATS / outbox (есть в коде, 2026-07-03).** Трейс не рвётся на асинхронной границе REST → outbox → NATS → consumer:

- `Publisher.Publish/PublishInTx` (`platform/outbox/publisher.go`) сохраняют W3C trace-context из ctx в колонку `trace_context` outbox-события; worker при publish переносит его в NATS headers (`traceparent` + `Hub-Event-Id`).
- Хелперы: `natsclient.InjectTraceContext/ExtractTraceContext` (carrier поверх `nats.Header`), `TraceContextMap/ExtractTraceContextFromMap` (`platform/nats/trace.go`). Пропагатор явный (W3C TraceContext+Baggage), не глобальный — работает и без OTLP endpoint.
- Прямой publish в рамках запроса: `nc.PublishJSONContext(ctx, subject, payload)` или `PublishWithTraceContext` вместо `PublishJSON` без ctx.
- Consumer-сторона: `ctx = natsclient.ExtractTraceContext(ctx, msg.Headers())`, затем `observability.StartConsumerSpan(ctx, subject, eventID)` / `EndConsumerSpan(span, err)` — span `consume <subject>` (kind=Consumer, атрибуты `messaging.*`, `hub.event_id`). Применено в notification consumer, audit consumer, ws bridge (`traced`-обёртка).

### 5.5 Если трейсинг выключен

Когда `OTEL_EXPORTER_OTLP_ENDPOINT` не задан (типичный dev/минимальный prod):

- `InitTracing` возвращает no-op shutdown, `tracingActive=false`.
- `startOperationSpan` возвращает `(ctx, nil)` — span'ы не создаются, оверхед нулевой.
- Логи и метрики продолжают работать полностью.

Это значит: **логи и метрики — базовый уровень (всегда), трейсинг — усиление (по возможности)**. Не полагайтесь на трейсы там, где обязаны быть лог и метрика.

### 5.6 Минимальное включение (для devops)

Чтобы поднять трейсинг, достаточно окружения — код уже готов:

```bash
# указать OTLP-коллектор (например, Grafana Tempo / OTel Collector)
OTEL_EXPORTER_OTLP_ENDPOINT="http://otel-collector:4318"
OTEL_SERVICE_NAME="hub-backend"
```

Коллектор поднимается опциональным `deployments/docker-compose.monitoring.yml` (Prometheus + Loki + Grafana, pinned v3.x — `docs/stack_and_tools.md`). Sampling трейсов настраивается на уровне SDK/коллектора: в проде под нагрузкой не экспортируйте 100% трейсов — используйте head/tail sampling на коллекторе, чтобы не платить за хранение.

### 5.7 Тестовый сброс

Для изоляции тестов есть `ResetTracingForTest()` и (под `//go:build hubtest`) `ExportSetTracingActiveForTest`. Используйте их в тестах наблюдаемости, чтобы включать/выключать трейсинг детерминированно и не оставлять global state между тестами.

---

<a name="6-frontend-наблюдаемость"></a>
## 6. Frontend-наблюдаемость

### 6.1 Что есть в коде: Error Boundaries

На фронте (`frontend/src`) есть два React Error Boundary:

- `components/ui/error-boundary.tsx` — базовый `ErrorBoundary` (FE-H4): перехватывает ошибки рендера, показывает fallback вместо белого экрана.
- `components/route-error-boundary.tsx` — `RouteErrorBoundary` с пропсом `module`: локализует падение в пределах раздела и даёт кнопки retry/reload.

Оба логируют в `console.error` в `componentDidCatch`. Внешнего сборщика (Sentry) пока **нет** — это осознанно вне скоупа FE-H4 (комментарий в коде прямо об этом говорит).

Реальный `RouteErrorBoundary`:

```37:45:frontend/src/components/route-error-boundary.tsx
	componentDidCatch(error: Error, info: ErrorInfo): void {
		// Лог в консоль — для разработки. В продакшене сюда стоит подключить
		// внешний сборщик (Sentry/etc.), но это вне скоупа FE-H4.
		console.error(
			`[${this.props.module}] route crashed:`,
			error.message,
			info.componentStack,
		);
	}
```

### 6.2 Обязательное правило: оборачивать модули

По `.cursor/rules/30-frontend.mdc` **каждый крупный модуль** оборачивается в error boundary. Обязательно для: `calls`, `dashboard`, `tracker`, `TMS`, `messenger`.

Good — модуль под boundary с осмысленным `module`:

```tsx
// падение звонков не роняет всё приложение; module -> в лог и в UI
<RouteErrorBoundary module="Звонки">
	<CallsModule />
</RouteErrorBoundary>
```

Bad — тяжёлый модуль без boundary:

```tsx
// любой throw в рендере -> белый экран всего SPA
<CallsModule />
```

### 6.3 `console.*` — только dev

Правило: в проде фронт **не** должен сыпать `console.log`. Причины: (1) утечка внутренностей/данных в консоль браузера пользователя, (2) шум, (3) возможная утечка чувствительных значений.

- `console.error` в error boundary — допустимо (диагностика падения), но в идеале дополняется отправкой в сборщик (см. §6.4).
- `console.log`/`console.debug` в бизнес-коде — **удалять перед merge** или гейтить на `import.meta.env.DEV`.

Good — dev-only лог:

```typescript
// логи разработки не попадают в prod-бандл
if (import.meta.env.DEV) {
	console.debug("[messenger] socket state", state);
}
```

Bad — безусловный лог в проде, ещё и с данными:

```typescript
console.log("current user", user);      // PII в консоли prod
console.log("token", accessToken);       // КАТАСТРОФА: токен в консоли
```

### 6.4 Отправка ошибок (reporting): как ввести минимально

Сборщика ошибок пока нет. Когда его подключат (Sentry или собственный endpoint `/api/telemetry/error`), точка интеграции — `componentDidCatch` обоих boundary плюс глобальные обработчики. Минимальная безопасная схема:

```typescript
// централизованный reporter: одно место, легко отключить/подменить в тестах
export function reportError(error: Error, context: Record<string, unknown>): void {
	// НЕ отправляем токены/пароли: context собирается вызывающим осознанно
	if (import.meta.env.DEV) {
		console.error("[reportError]", error.message, context);
		return;
	}
	// prod: отправка в сборщик (пример на будущее, endpoint TBD)
	void fetch("/api/telemetry/error", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		// X-Request-ID отдаёт бэк в ответах; можно прикладывать для корреляции
		body: JSON.stringify({
			message: error.message,
			stack: error.stack,
			...context,
		}),
	}).catch(() => {
		// reporting не должен ронять приложение: глотаем сетевую ошибку молча
	});
}
```

Глобальные обработчики необработанных ошибок (для ошибок вне React-дерева):

```typescript
// ловим то, что не попало в error boundary (async, обработчики событий)
window.addEventListener("error", (e) => {
	reportError(e.error ?? new Error(e.message), { kind: "window.onerror" });
});
window.addEventListener("unhandledrejection", (e) => {
	reportError(
		e.reason instanceof Error ? e.reason : new Error(String(e.reason)),
		{ kind: "unhandledrejection" },
	);
});
```

Правила reporting:

- Reporter **никогда не роняет** приложение: сетевые ошибки самого reporting проглатываются.
- В `context` кладём только безопасное: `module`, `route`, `request_id` (из заголовка ответа), но **не** токены, пароли, содержимое форм авторизации.
- Дедупликация/семплинг на клиенте: не слать одну и ту же ошибку сотни раз (например, при рендер-цикле).

### 6.5 Корреляция фронт ↔ бэк через `X-Request-ID`

Бэк возвращает `X-Request-ID` в каждом ответе (§2.8). Фронт может: (1) показывать его пользователю в тосте ошибки («сообщите код поддержке»), (2) прикладывать к отчёту об ошибке. Это связывает клиентский инцидент с серверными логами.

Good — показываем request_id в ошибке для поддержки:

```typescript
// пользователь сообщит request_id -> поддержка найдёт весь серверный лог
function handleApiError(err: ApiError): void {
	toast.error("Не удалось выполнить действие", {
		description: err.requestId ? `Код: ${err.requestId}` : undefined,
	});
}
```

### 6.6 Web Vitals и Lighthouse

Производительность фронта — часть наблюдаемости UX. Из `AGENTS.md`/`docs`: Lighthouse CI существует как **artifact-job** (FB-06, PR #27), пока **не блокирующий** гейт (NFR-17). Web-vitals (LCP, CLS, INP/FID, TTFB) — целевые метрики UX.

- `web-vitals` можно снимать на клиенте и слать в тот же telemetry-endpoint (когда появится), либо в консоль в dev.
- Lighthouse гоняется в CI как артефакт; при переводе в блокирующий гейт пороги задаются в конфиге CI (`devops`).
- Никаких секретов в web-vitals payload — это чисто числовые метрики + маршрут.

Пример снятия web-vitals (dev-логирование):

```typescript
import { onCLS, onINP, onLCP } from "web-vitals";

// в dev — в консоль; в prod — в reporter/telemetry (endpoint TBD)
const sink = (metric: { name: string; value: number }) => {
	if (import.meta.env.DEV) {
		console.debug("[web-vitals]", metric.name, Math.round(metric.value));
	}
};
onLCP(sink);
onCLS(sink);
onINP(sink);
```

### 6.7 Никаких секретов в консоль/сеть

Сводно, фронт-запреты (зеркало §2.6):

- Не `console.log` токены, пароли, содержимое формы логина/смены пароля.
- Не класть токены в URL/query (бэк их вырезает из логов, но фронт не должен их туда класть вовсе; для WS-токена — это осознанный вектор ME-03, обрабатываемый на сервере).
- Reporting-payload — только безопасные поля.

---

<a name="7-логи-в-тестах"></a>
## 7. Логи в тестах

### 7.1 `t.Log` / `t.Logf` вместо `slog`/`fmt` в тестах

В тестах используйте `t.Log`/`t.Logf`, а не `slog`/`fmt.Println`. Причины:

- Вывод `t.Log` **привязан к тесту**: показывается только при `-v` или при падении теста (`go test` печатает лог упавшего теста). Не засоряет успешные прогоны.
- `t.Log` уважает параллельность (`t.Parallel()`) и не мешается между тестами.
- `fmt.Println` в тесте пишется всегда и в общий stdout — шум, который невозможно связать с конкретным тестом.

Good:

```go
func TestCreateIssue(t *testing.T) {
	issue, err := svc.CreateIssue(ctx, req)
	if err != nil {
		// покажется только при падении/‑v; привязано к этому тесту
		t.Fatalf("create issue: %v", err)
	}
	t.Logf("created issue id=%s in %s", issue.ID, elapsed)
}
```

Bad:

```go
func TestCreateIssue(t *testing.T) {
	issue, _ := svc.CreateIssue(ctx, req)
	fmt.Println("issue:", issue) // печатается всегда, не привязано к тесту
}
```

### 7.2 Что логировать при падении

Хороший тест при падении печатает **достаточно для диагностики без повторного запуска**:

- Ожидаемое vs фактическое значение (`want`/`got`).
- Ключевые входные данные, приведшие к падению.
- `request_id`/id сущности, если тест бил по HTTP/сервису.

Good — информативное падение:

```go
if got := resp.StatusCode; got != want {
	// сразу видно и код, и тело с request_id -> не нужно перезапускать
	t.Fatalf("status = %d, want %d; body=%s", got, want, string(body))
}
```

Bad — падение без контекста:

```go
if resp.StatusCode != 200 {
	t.Fail() // что вернулось? почему? -> перезапуск с отладкой
}
```

### 7.3 Проверка наблюдаемости в тестах

Наблюдаемость сама по себе тестируема. В Hub есть хуки:

- `observability.ResetMetrics()` / `ResetPrometheusMetrics()` — сброс метрик между тестами.
- `observability.Snapshot()` — проверка, что операция записала метрику (count/error_count).
- `ExportRecordOperation`, `ExportSetTracingActiveForTest` (под `//go:build hubtest`) — прямой доступ к внутренним функциям.
- `ResetTracingForTest()` — детерминированный сброс трейсинга.

Good — тест проверяет, что операция залогировала метрику:

```go
func TestReport_RecordsMetric(t *testing.T) {
	observability.ResetMetrics()
	_, _ = svc.GenerateReport(ctx, projectID) // внутри Timed("tms.generate_report", ...)

	snap := observability.Snapshot()
	var found bool
	for _, m := range snap {
		if m.Operation == "tms.generate_report" {
			found = true
			if m.Count != 1 {
				t.Fatalf("count = %d, want 1", m.Count)
			}
		}
	}
	if !found {
		t.Fatal("metric tms.generate_report not recorded")
	}
}
```

### 7.4 Артефакты и тайминги прогонов

- По `.cursor/rules/45-artifacts-cleanup.mdc`: coverage-профили пишите в `${TMPDIR}`/`/tmp`, не в `backend/`. Не коммитьте `*.out`, `cover*`, `cov_*`.
- По `docs/TESTING.md`: фиксируйте суммарное время регрессии в отчёте сессии/`.planning/COVERAGE.md` при заметной деградации.
- При падении в CI сохраняйте лог упавшего теста как артефакт job (devops), чтобы разбор не требовал локального воспроизведения.
- Для перф-тестов SLO (§4.6) логируйте измеренные p95/max — это доказательная база DoD.

### 7.5 Не логировать секреты и в тестах

Тестовые данные тоже не должны утекать: не печатайте реальные токены/ключи в `t.Log`. Для фикстур используйте очевидно фейковые значения (`"test-token"`, `"user@example.test"`), чтобы даже случайный вывод не выглядел как реальный секрет и не триггерил секрет-сканеры.

---

<a name="8-формат-логов-prod-vs-dev"></a>
## 8. Формат логов prod vs dev, уровни, ротация

### 8.1 Текущий формат: JSON в stdout

Как показано в §2.1, `initLogger` использует **`slog.NewJSONHandler(os.Stdout, ...)`** с `Level: slog.LevelInfo` и глобальным полем `service`. Это единый формат для всех окружений на текущий момент.

Почему JSON везде:

- Продакшн (Docker/K3s): stdout собирается сборщиком логов (Loki через Promtail/агент), JSON парсится в поля без регулярок.
- Единый формат исключает «работает в dev, ломается парсинг в prod».

### 8.2 Prod (JSON) vs Dev (текст): рекомендация

Для локальной разработки JSON тяжело читать глазами. Рекомендуемое (не ломающее прод) улучшение — выбирать хендлер по окружению. Окружение уже доступно: `cfg.Server.Environment` (`ENVIRONMENT`, дефолт `development`) и `DEV_MODE`.

Рекомендуемый паттерн (ввести в `initLogger`):

```go
func initLogger() {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "hub-backend"
	}

	level := parseLevel(os.Getenv("LOG_LEVEL")) // см. 8.3
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.EqualFold(os.Getenv("ENVIRONMENT"), "development") {
		// текстовый handler читаем человеком в терминале
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// JSON для машинного сбора в prod/staging
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler).With("service", serviceName))
}
```

Инвариант: **prod/staging всегда JSON**. Текст — только `ENVIRONMENT=development`. Никогда не отдавайте текстовый формат в прод — он ломает парсинг и агрегацию.

### 8.3 Управление уровнем через env

Сейчас уровень захардкожен `LevelInfo`. Рекомендуемое улучшение — управлять через `LOG_LEVEL`, чтобы в инциденте временно включить `Debug` без пересборки:

```go
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // безопасный дефолт: соответствует текущему поведению
	}
}
```

Правила уровней по окружению:

| Окружение | Формат | Уровень по умолчанию |
|---|---|---|
| development | text | `Debug` или `Info` |
| staging | JSON | `Info` |
| production | JSON | `Info` (временно `Debug` при инциденте) |

`Debug` в проде — **временная** мера: включили, сняли диагностику, вернули `Info`. Постоянный `Debug` в проде = шум + стоимость хранения.

### 8.4 Ротация и хранение

Приложение **не ротирует логи само** — оно пишет в stdout. Ротация/ретеншн — ответственность инфраструктуры:

- **Docker**: драйвер логов (`json-file` с `max-size`/`max-file`, либо отправка в Loki). Без лимитов `json-file` заполнит диск.
- **K3s**: сборщик (Promtail/агент) → Loki с настроенным retention.
- **Loki**: ретеншн по времени/размеру; чувствительные логи не хранить дольше необходимого (GDPR).

Рекомендация devops для `docker-compose.prod.yml`:

```yaml
# ограничить рост логов контейнера; иначе диск заполнится в инциденте
logging:
  driver: "json-file"
  options:
    max-size: "50m"
    max-file: "5"
```

### 8.5 UTC и таймстемпы

`slog` по умолчанию проставляет `time` в записи. Инвариант ТЗ: **всё в UTC**. Не переопределяйте таймзону логгера на локальную. `Snapshot()` уже форматирует `LastSeenUTC` через `time.Now().UTC()` — придерживайтесь этого во всех кастомных полях времени.

---

<a name="9-алёрты-и-дашборды"></a>
## 9. Алёрты и дашборды

### 9.1 Стек мониторинга

По `docs/architecture.md` и `docs/stack_and_tools.md` мониторинг — **опциональный** стек `deployments/docker-compose.monitoring.yml`: **Prometheus + Loki + Grafana** (pinned v3.x). Prometheus скрейпит `/metrics` (§4.5), Loki собирает JSON-логи stdout, Grafana строит дашборды и алёрты.

Разделение ответственности:

- **Prometheus** — числовые метрики, алёрты по порогам (error rate, latency).
- **Loki** — логи; поиск по `request_id`, фильтр по `module`/уровню; логовые алёрты (например, всплеск `Error`).
- **Grafana** — визуализация обоих + управление алёртами; при включённом OTel — ещё и трейсы (Tempo).

### 9.2 Что мониторить: обязательный набор

| Категория | Метрика/источник | Порог/цель | Почему |
|---|---|---|---|
| **Error rate (HTTP)** | `hub_http_requests_total{status=~"5.."}` / всего | алёрт при > 1% за 5м | Рост 5xx = сбой сервиса |
| **Latency API (p95)** | `hub_http_request_duration_milliseconds` | SLO < 200 мс | Деградация UX |
| **Latency операций (p99)** | `hub_operation_duration_milliseconds` | зависит от операции | Медленные бизнес-операции |
| **Ошибки операций** | `hub_operation_calls_total{result="error"}` | всплеск | Провалы фоновых/критичных операций |
| **WS-подключения** | счётчик активных клиентов (WS-слой) | аномальный рост/обвал | Health реального времени |
| **WS дропы буфера** | счётчик `send buffer full` (§2.9) | > 0 устойчиво | Клиенты не успевают читать / бэкпрешер |
| **Очереди NATS** | pending/redelivery JetStream consumer | рост pending | Консьюмеры не справляются |
| **Логи Error** | Loki: rate `level="ERROR"` | всплеск по модулю | Быстрое обнаружение инцидента |
| **Ресурсы** | CPU/RAM/диск контейнеров | стандартные | OOM/диск под логи |

### 9.3 Примеры алёртов (Prometheus)

```promql
# HTTP error rate > 1% за 5 минут -> страница/уведомление
(
  sum(rate(hub_http_requests_total{status=~"5.."}[5m]))
  /
  sum(rate(hub_http_requests_total[5m]))
) > 0.01

# нарушение SLO: p95 латентности API выше 200 мс
histogram_quantile(0.95,
  sum(rate(hub_http_request_duration_milliseconds_bucket[5m])) by (le)
) > 200

# всплеск ошибок конкретной бизнес-операции
sum(rate(hub_operation_calls_total{result="error"}[5m])) by (operation) > 0.5
```

### 9.4 Дашборды: минимальный набор панелей

Рекомендованные дашборды Grafana:

1. **Обзор сервиса**: RPS, error rate, p50/p95/p99 латентности, топ медленных маршрутов.
2. **Бизнес-операции**: `hub_operation_*` — count/error по `operation`, p95 по операциям.
3. **Реальное время (WS/NATS)**: активные WS-клиенты, дропы буфера, pending/redeliveries JetStream.
4. **Логи**: Loki-панель с фильтром по `module`/`level`, быстрый поиск по `request_id`.
5. **SLO-панель**: явные пороги (§4.6) с цветовой индикацией нарушения.

### 9.5 Корреляция при инциденте (runbook в миниатюре)

Пошагово, как связать три столпа при разборе:

1. Алёрт «error rate 5xx > 1%» в Grafana → определяем окно времени.
2. Панель латентности/операций → видим, какой `route`/`operation` деградировал.
3. Loki: `{service="hub-backend", level="ERROR"} | json | module="tracker"` за окно → находим конкретные ошибки, берём `request_id`.
4. Loki по `request_id` → вся цепочка запроса (access-лог + сервисные логи).
5. (Если OTel включён) Grafana Tempo по `hub.request_id`/trace → видим span-дерево и узкое место.

Именно поэтому §2.8 и §5.2 настаивают на сквозном `request_id` — он единственная нить, связывающая логи, ответ пользователю и трейс.

### 9.6 Гигиена алёртов

- Алёрт без действия — шум. Каждый алёрт → runbook (что делать).
- Пороги по **перцентилям и rate**, не по абсолютным разовым значениям (один медленный запрос ≠ инцидент).
- `for: 5m` в правилах — не будить дежурного из-за секундного всплеска.
- Не алёртить на ожидаемое: `context.Canceled` при деплое, 4xx (это клиентские ошибки, не сбой сервиса — кроме всплеска 401/403, который может значить атаку).

---

<a name="10-definition-of-done"></a>
## 10. Definition of Done по наблюдаемости

Фича/фикс **не считается завершённой**, пока не выполнен чек-лист ниже. Это гейт `.cursor/rules/56-regression-verify.mdc` и обязательная часть ревью `be-ver`/`fe-ver`.

### 10.1 DoD для backend-фичи

- [ ] Значимые бизнес-события логируются через `slog.*Context(ctx, ...)` с константным сообщением и полями.
- [ ] Есть обязательные поля: `module`, `request_id` (в HTTP-пути), `user_id`/`entity_id` где применимо.
- [ ] Ошибки логируются **один раз, на границе**; нижние слои оборачивают через `%w` (нет log-and-return).
- [ ] Ожидаемые исходы (not found/conflict/validation) возвращают типизированную `AppError`, а не логируются как `Error`.
- [ ] HTTP-ошибки идут через `ErrorHandler`; клиент получает `{code, message, request_id}`; детали/стектрейс — только в лог.
- [ ] Критичные/SLO-операции обёрнуты в `observability.Timed`/`TimedVoid` (лог + метрика + span).
- [ ] Имена операций/маршрутов низкокардинальны (нет id/токенов в метках метрик и именах span).
- [ ] Нет `fmt.Println`/`log.Printf`/`log.Fatal`/`panic()` в рантайм-путях.
- [ ] Нет секретов/токенов/PII/полных `Authorization`-заголовков/query с токеном в логах.
- [ ] Горутины (WS/consumer/cron) защищены от паники; `context.Canceled` при shutdown не логируется как `Error`.
- [ ] Уровни выставлены осмысленно (Debug шумное, Info значимое, Warn деградация, Error сбой).
- [ ] Регрессия зелёная: `go build -tags release ./... && go test -race -count=1 -timeout=600s ./...`.

### 10.2 DoD для frontend-фичи

- [ ] Крупный модуль обёрнут в `RouteErrorBoundary`/`ErrorBoundary` с осмысленным `module`.
- [ ] Нет безусловных `console.log`/`console.debug` в проде (только `import.meta.env.DEV` или удалены).
- [ ] Нет токенов/PII в консоли или в URL/query.
- [ ] Ошибки API показывают `request_id` пользователю (для поддержки) где уместно.
- [ ] (Если reporter подключён) ошибки уходят в reporter без чувствительных полей; reporter не роняет приложение.
- [ ] Регрессия зелёная: `npx tsc --noEmit && npm run lint && npm run test:run`.

### 10.3 DoD для devops-изменения

- [ ] `/metrics` защищён CIDR-allowlist в проде.
- [ ] Формат логов в prod/staging — JSON; уровень управляем через env.
- [ ] Ротация/ретеншн логов настроены (Docker `max-size`/`max-file` или Loki retention).
- [ ] Алёрты имеют `for:` и привязанный runbook; нет алёртов на ожидаемое.
- [ ] (Если включают OTel) `OTEL_EXPORTER_OTLP_ENDPOINT` задан, sampling настроен на коллекторе.

### 10.4 Чек-лист ревью логов (для верификатора)

Быстрый проход по diff:

1. **Греп на запреты**: `fmt.Println`, `fmt.Printf`, `log.Print`, `log.Fatal`, `panic(`, `console.log`. Каждое совпадение — обосновать или убрать.
2. **Греп на секреты**: `token`, `password`, `secret`, `authorization`, `Cookie` рядом с `slog`/`console` — проверить, не логируется ли значение.
3. **Уровни**: есть ли `Error` на рутинных событиях? Есть ли реальные сбои на `Info`?
4. **Дубли**: одна ошибка логируется на нескольких слоях? → оставить один лог на границе.
5. **Контекст**: используется ли `...Context(ctx)` там, где ctx доступен?
6. **Поля**: есть `module`? есть `request_id`/`entity_id`? сообщение — константа?
7. **Метрики**: критичная/SLO-операция обёрнута в `Timed`? метки низкокардинальны?
8. **Кардинальность**: нет ли переменных id в именах операций/меток?

### 10.5 Связь с общим QA-гейтом

Наблюдаемость — часть, а не замена регрессии. Полный порядок сдачи по `.cursor/rules/56-regression-verify.mdc`:

```
1. Изменил код (+ логи/метрики по этому playbook)
2. Минимальная регрессия затронутого слоя
3. /qa-gate перед handoff
4. Docs sync (.cursor/rules/55-docs-sync.mdc)
5. be-ver/fe-ver для критичных изменений (прогон чек-листа 10.4)
6. Обновить .planning/STATE.md
```

---

<a name="11-антипаттерны-логирования"></a>
## 11. Антипаттерны логирования

Сводная таблица «нельзя → почему → как правильно». Это чек-лист для ревью (§10.4) и для самопроверки перед merge.

| # | Нельзя | Почему | Как правильно |
|---|---|---|---|
| 1 | Логирование в цикле (по элементу) | Взрыв объёма логов, забивает Loki, скрывает важное | Логировать агрегат после цикла: `"processed", "count", n` |
| 2 | Логирование секретов/токенов/PII | Утечка учёток; логи хранятся и менее защищены | Логировать id/ссылку/хэш; см. §2.6 |
| 3 | log-and-return (лог + возврат ошибки) | Дубли одной ошибки на каждом слое | Либо `%w`-wrap, либо один лог на границе (§3.3) |
| 4 | Конкатенация строк вместо полей | Не фильтруется, не агрегируется, высокая кардинальность | Пары ключ-значение: `slog.Info("msg", "k", v)` |
| 5 | `fmt.Println` / `log.Printf` | Не структурировано, без уровня/времени/service | `slog.*Context(ctx, ...)` |
| 6 | `log.Fatal` / `os.Exit` в рантайме | Убивает процесс без graceful shutdown | Вернуть ошибку; `os.Exit` только на старте |
| 7 | `panic()` в handler/service | Крах запроса/процесса | Вернуть `error`; `Recover` только как сеть безопасности |
| 8 | `Error` на рутинных событиях | Ложный рост error rate, усталость от алёртов | Понизить уровень (Info/Debug) |
| 9 | Реальный сбой на `Info`/проглочен | Инцидент утонет в шуме или исчезнет | `Error`/`Warn` с полем `error` |
| 10 | Переменные данные в имени операции/метки | Взрыв кардинальности Prometheus | `<module>.<action>`, id — в поля лога |
| 11 | Query string с токеном в access-логе | Утечка JWT (ME-03) | `stripQuery` (§2.7) |
| 12 | Логирование без `ctx`, когда он есть | Теряется trace/request корреляция | `...Context(ctx, ...)` |
| 13 | Ручной JSON ошибки в handler | Теряется `request_id`, утечка внутренностей | Типизированная `AppError` + `ErrorHandler` |
| 14 | Замер времени без `defer` | Латентность неудачных вызовов не логируется | `defer` с `dur_ms` или `Timed` |
| 15 | Дублирование `Timed` вручную | Рассинхрон лог/метрика/span | Использовать `observability.Timed` |
| 16 | `console.log` в prod-фронте | Утечка данных в консоль браузера, шум | `import.meta.env.DEV`-гейт или удалить |
| 17 | Reporting, роняющий приложение | Ошибка телеметрии = падение UX | Проглатывать сетевые ошибки reporter |
| 18 | Текстовый формат логов в prod | Ломает парсинг Loki | JSON в prod/staging всегда |
| 19 | Постоянный `Debug` в prod | Стоимость хранения, шум | `Info` дефолт; `Debug` временно при инциденте |
| 20 | Логи в локальной TZ | Рассинхрон времени между узлами | UTC везде |

### 11.1 Развёрнуто: антипаттерн №1 (лог в цикле)

Bad:

```go
for _, u := range users {
	// N логов на одну операцию -> тысячи строк в Loki
	slog.InfoContext(ctx, "notifying user", "user_id", u.ID)
	_ = notify(ctx, u)
}
```

Good:

```go
var sent, failed int
for _, u := range users {
	if err := notify(ctx, u); err != nil {
		failed++
		continue
	}
	sent++
}
// один агрегированный лог с итогом
slog.InfoContext(ctx, "notification batch done",
	"module", "notification",
	"sent", sent,
	"failed", failed,
	"total", len(users),
)
```

### 11.2 Развёрнуто: антипаттерн №4 (конкатенация vs поля)

Bad:

```go
slog.Info("user " + userID.String() + " joined channel " + channelID.String())
```

Good:

```go
slog.InfoContext(ctx, "user joined channel",
	"module", "messenger",
	"user_id", userID,
	"channel_id", channelID,
)
```

Разница на практике: в Loki `{module="messenger"} | json | user_id="<id>"` мгновенно найдёт все события пользователя во втором варианте. В первом — только медленная регулярка по подстроке, и агрегировать «сколько присоединений» невозможно.

### 11.3 Развёрнуто: антипаттерн №8/№9 (перепутанные уровни)

Bad:

```go
slog.Error("healthcheck ok")               // рутинное на Error -> ложный алёрт
if err != nil { slog.Info("db write failed", "error", err) } // сбой на Info -> утонет
```

Good:

```go
slog.DebugContext(ctx, "healthcheck ok")    // рутина -> Debug
if err != nil {
	slog.ErrorContext(ctx, "db write failed", "module", "tracker", "error", err) // сбой -> Error
}
```

---

<a name="12-приложения"></a>
## 12. Приложения

### 12.1 Словарь обязательных ключей полей

Единый словарь ключей — чтобы фильтры в Loki работали одинаково во всех модулях. Не изобретайте синонимы (`uid` vs `user_id`).

| Ключ | Тип | Значение | Пример |
|---|---|---|---|
| `service` | string | Имя сервиса (глобально) | `hub-backend` |
| `module` | string | Домен | `tracker`, `messenger`, `auth`, `calls`, `tms`, `ws`, `notification` |
| `request_id` | string | Сквозной id запроса | UUID |
| `user_id` | string | Инициатор действия | UUID |
| `issue_id` / `channel_id` / `message_id` / `call_id` | string | Конкретная сущность | UUID |
| `entity_id` + `entity_type` | string | Обобщённая сущность (audit/outbox) | `<uuid>` + `issue` |
| `operation` | string | Имя `Timed`-операции | `tracker.create_issue` |
| `duration_ms` / `dur_ms` | int64 | Длительность | `142` |
| `status` | int | HTTP-статус | `200`, `500` |
| `method` / `path` / `route` | string | HTTP-атрибуты | `POST`, `/api/issues`, `/api/issues/:id` |
| `error` | error | Ошибка (цепочка `%w`) | `create issue: ent: ...` |
| `subject` | string | NATS subject | `hub.issues.created` |

Соглашение: предпочитайте `duration_ms` (как в `Timed`/`Logger`); `dur_ms` допустим в ручных замерах, но лучше единообразно `duration_ms`.

### 12.2 Переменные окружения наблюдаемости

| Переменная | Назначение | Дефолт | Кто ставит |
|---|---|---|---|
| `OTEL_SERVICE_NAME` | Имя сервиса в логах/трейсах | `hub-backend` | devops |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP-коллектор; пусто = трейсинг выключен | пусто | devops |
| `ENVIRONMENT` | `production`/`staging`/`development` | `development` | devops |
| `DEV_MODE` | Небезопасные dev-эндпоинты; **запрещён в prod** | `false` | devops |
| `LOG_LEVEL` (рекомендация) | Уровень логов | `info` | devops |
| Metrics CIDR allowlist | Ограничение `/metrics` | пусто (открыто) | devops |

### 12.3 Шпаргалка для `be1`/`be2`

```go
import (
	"log/slog"
	"github.com/valerii/hub/backend/internal/platform/observability"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// 1. событие
slog.InfoContext(ctx, "issue created", "module", "tracker", "issue_id", id, "user_id", uid)

// 2. критичная операция (лог + метрика + span)
res, err := observability.Timed(ctx, "tracker.create_issue", func(ctx context.Context) (*Issue, error) {
	return repo.Create(ctx, req)
})

// 3. ошибка на границе (один раз)
if err != nil {
	slog.ErrorContext(ctx, "create issue failed", "module", "tracker", "error", err)
	return apperrors.NewInternal("failed to create issue")
}

// 4. ожидаемый исход -> типизированная ошибка, без Error-лога
if errors.Is(err, tracker.ErrNotFound) {
	return apperrors.NewNotFound("issue not found")
}
```

Запреты: `fmt.Println`, `log.Printf`, `log.Fatal`, `panic()`, секреты в логах, лог в цикле, log-and-return.

### 12.4 Шпаргалка для `fe1`/`fe2`

```tsx
// 1. обернуть модуль
<RouteErrorBoundary module="Трекер">
	<TrackerModule />
</RouteErrorBoundary>

// 2. dev-only лог
if (import.meta.env.DEV) console.debug("[tracker] state", state);

// 3. показать request_id в ошибке
toast.error("Ошибка", { description: err.requestId ? `Код: ${err.requestId}` : undefined });
```

Запреты: `console.log` в проде, токены/PII в консоли или URL.

### 12.5 Шпаргалка для `qa`

```go
// t.Log вместо fmt/slog
t.Logf("created id=%s in %s", id, elapsed)

// информативное падение
t.Fatalf("status = %d, want %d; body=%s", got, want, body)

// проверка метрики
observability.ResetMetrics()
_, _ = svc.Op(ctx)
snap := observability.Snapshot() // проверить Count/ErrorCount
```

Артефакты покрытия — в `${TMPDIR}`, не в `backend/`; не коммитить.

### 12.6 Шпаргалка для `devops`

```bash
# prod: JSON-логи (дефолт), уровень через env (рекомендация)
ENVIRONMENT=production
LOG_LEVEL=info

# опциональный трейсинг
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=hub-backend

# мониторинг-стек
docker compose -f deployments/docker-compose.monitoring.yml up -d
```

```promql
# ключевые алёрты
# 5xx error rate > 1%
(sum(rate(hub_http_requests_total{status=~"5.."}[5m])) / sum(rate(hub_http_requests_total[5m]))) > 0.01
# SLO p95 API > 200ms
histogram_quantile(0.95, sum(rate(hub_http_request_duration_milliseconds_bucket[5m])) by (le)) > 200
```

Обязательно: CIDR-allowlist на `/metrics`; ротация логов (`max-size`/`max-file` или Loki retention).

### 12.7 Карта источников (реальные файлы кода)

| Тема | Файл |
|---|---|
| Инициализация логгера | `backend/cmd/server/wiring_bootstrap.go` |
| Access-лог + stripQuery (ME-03) | `backend/internal/platform/middleware/logging.go` |
| RequestID middleware | `backend/internal/platform/middleware/requestid.go` |
| Централизованный ErrorHandler | `backend/internal/platform/middleware/middleware.go` |
| AppError (`{code,message,request_id}`) | `backend/internal/platform/errors/errors.go` |
| Timed/TimedVoid | `backend/internal/platform/observability/timing.go` |
| In-memory метрики / Snapshot | `backend/internal/platform/observability/metrics.go` |
| Prometheus метрики + /metrics + HTTP middleware | `backend/internal/platform/observability/prometheus.go` |
| /metrics CIDR allowlist | `backend/internal/platform/observability/metrics_allowlist.go` |
| OpenTelemetry (InitTracing, HTTPTraceMiddleware, spans) | `backend/internal/platform/observability/otel.go` |
| Тестовые хуки (hubtest) | `backend/internal/platform/observability/export_hooks.go` |
| Frontend ErrorBoundary | `frontend/src/components/ui/error-boundary.tsx` |
| Frontend RouteErrorBoundary | `frontend/src/components/route-error-boundary.tsx` |

### 12.8 Глоссарий

| Термин | Значение |
|---|---|
| Столпы наблюдаемости | Logs (события), Metrics (агрегаты), Traces (путь запроса) |
| Structured logging | Логи как пары ключ-значение (JSON), а не свободный текст |
| `request_id` | Сквозной идентификатор запроса; мост между логами, ответом и трейсом |
| Correlation | Связывание событий разных подсистем по общему id |
| Cardinality (кардинальность) | Число уникальных значений метки метрики; высокая — опасна для Prometheus |
| p50/p95/p99 | Перцентили латентности (медиана/почти худший/хвост) |
| SLO | Целевой уровень качества (например, p95 API < 200 мс) |
| Span | Узел трейса — одна операция с длительностью и атрибутами |
| Propagation | Проброс trace-контекста между сервисами (W3C `traceparent`) |
| Sampling | Логирование/трассировка части событий для экономии |
| OTLP | OpenTelemetry Protocol — формат экспорта телеметрии |
| Error boundary | React-компонент, перехватывающий ошибки рендера поддерева |

### 12.9 Быстрый принцип на одну строку

> **Логируй событие один раз на границе, структурированно, с `request_id` и `module`, без секретов; меряй критичное через `Timed`; метрики и логи обязательны, трейсинг — усиление; уровни не путай; в цикле не логируй.**

---

*Playbook основан на реальном коде Hub на 2026-07-02. При изменении кода наблюдаемости (`backend/internal/platform/observability/*`, `middleware/*`) — обновить соответствующие разделы (`.cursor/rules/55-docs-sync.mdc`).*
