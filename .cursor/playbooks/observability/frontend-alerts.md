# Наблюдаемость: фронтенд, алерты, дашборды

> Часть плейбука. Канон и оглавление — [`../observability.md`](../observability.md).

Клиентская телеметрия, пороги алертов, различия prod и dev, дашборды.

---

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
