# Hub — Playbook: логирование и наблюдаемость

> Эталонный документ для AI-команды Hub. На него ссылаются `be1`/`be2`/`fe1`/`fe2`/`qa`/`devops`.
> Обновлено: 2026-07-02 · Связанные правила: `.cursor/rules/10-style.mdc`, `.cursor/rules/20-backend.mdc`, `.cursor/rules/30-frontend.mdc`, `.cursor/rules/56-regression-verify.mdc`
> База фактов: `backend/internal/platform/observability/*`, `backend/internal/platform/middleware/*`, `docs/architecture.md`, `docs/stack_and_tools.md`, `docs/TESTING.md`

Этот playbook описывает **как именно** в Hub писать логи, снимать метрики и трассировать запросы. Он основан на **реальном коде** репозитория, а не на абстрактных рекомендациях. Все примеры `Good`/`Bad` привязаны к тому, что уже работает в `backend/internal/platform/observability` и `backend/internal/platform/middleware`.

Правило прочтения: если в тексте написано «есть в коде» — это проверяемый факт (файл указан). Если написано «ввести минимально / рекомендация» — этого пока нет, и раздел объясняет, как добавить, не ломая инварианты ТЗ.

---

## Оглавление

Канон здесь. Технологические разделы — отдельными файлами.

| Файл | Когда открывать |
|------|-----------------|
| [`observability/logging-go.md`](observability/logging-go.md) | Пишешь логи в Go: slog, уровни, поля, ошибки |
| [`observability/metrics-tracing.md`](observability/metrics-tracing.md) | Метрики, замер времени, OpenTelemetry |
| [`observability/frontend-alerts.md`](observability/frontend-alerts.md) | Клиентская телеметрия, алерты, дашборды |
| [`observability/reference.md`](observability/reference.md) | Приложения и справочные материалы |

Нумерация разделов сохранена — ссылки вида «§5» остаются рабочими.

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
- [ ] Регрессия выполнена через применимый именованный backend-job будущего versioned outer launcher `/qa-gate`. Evidence: source SHA, имя job, immutable image, CPU/RAM/PID limits, duration + result, task-owned artifacts и exact cleanup.
- [ ] Пока launcher не реализован, launcher/package имеют статус `NOT_READY`, runtime — `NOT_RUN`: запрещены прямые Docker/Compose, legacy runner и host fallback. Этот чек-лист не подтверждает G0-B: G0-B остаётся `NOT_VERIFIED`.

### 10.2 DoD для frontend-фичи

- [ ] Крупный модуль обёрнут в `RouteErrorBoundary`/`ErrorBoundary` с осмысленным `module`.
- [ ] Нет безусловных `console.log`/`console.debug` в проде (только `import.meta.env.DEV` или удалены).
- [ ] Нет токенов/PII в консоли или в URL/query.
- [ ] Ошибки API показывают `request_id` пользователю (для поддержки) где уместно.
- [ ] (Если reporter подключён) ошибки уходят в reporter без чувствительных полей; reporter не роняет приложение.
- [ ] Регрессия выполнена через применимый именованный frontend-job будущего versioned outer launcher `/qa-gate`. Evidence: source SHA, имя job, immutable image, CPU/RAM/PID limits, duration + result, task-owned artifacts и exact cleanup.
- [ ] Пока launcher не реализован, launcher/package имеют статус `NOT_READY`, runtime — `NOT_RUN`: запрещены прямые Docker/Compose, legacy runner и host fallback. Этот чек-лист не подтверждает G0-B: G0-B остаётся `NOT_VERIFIED`.

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
