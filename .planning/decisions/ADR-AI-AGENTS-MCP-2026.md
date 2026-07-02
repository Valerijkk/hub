# ADR-AI-AGENTS-MCP-2026: Agent-native AI, MCP-сервер и hub-ai по LiteLLM-паттерну

- **Status:** accepted (2026-07-03, решения пользователя — `.planning/GRAY-ZONES-DECISIONS-2026-07-03.md` §2, батч 2)
- **Дата:** 2026-07-03
- **Автор:** arch · **Верифицируют:** arch-ver + security-auditor
- **Связанные:** GZ-SEC-09/10/20/21/22/23, `ADR-ARCH-MICROSERVICES-2026` §3 (hub-ai — вынос №1), `.planning/MICROSERVICES-ROADMAP.md` §3 (уточнение scope hub-ai), `.planning/AI-PROVIDER-HUB-SPEC.md`, `docs/tz_constraints.md`
- **Расширяет:** текущий модуль `backend/internal/ai/` (5 адаптеров, BYOK, messenger-бот, fallback) до agent-native платформы

---

## Контекст

Текущий `ai/`-модуль — копилот: adapter-слой на 5 провайдеров (OpenAI/Anthropic/GigaChat/Yandex/stream), BYOK, `@hub-ai` в мессенджере, JQL-assist, PR-review. Тренд 2026 — **автономные агенты** (выполняют многошаговые задачи в системе) и **MCP** как lingua franca интеграции внешних агентов (Claude/Cursor/копилоты клиентов) с корпоративными системами.

Пользователь одобрил agent-native AI + MCP в scope и зафиксировал границы автономии, модель идентичности, лимиты и паттерн шлюза. Ключевой конфликт, который решает этот ADR: агент должен быть полезен (действовать), но не должен становиться неуправляемой поверхностью атаки (prompt injection → действия от имени пользователя).

## Рассмотренные варианты

1. **Полная автономия агента в правах пользователя** — максимум пользы, но prompt injection превращает агента в проводника атак (удаление данных, рассылка от имени). Отклонено.
2. **Всё через подтверждение (human-in-the-loop на каждый шаг)** — безопасно, но убивает ценность автономии (агент = медленная форма). Отклонено.
3. **Ступенчатая автономия + отдельный principal (принято)** — read и создание/черновики автономно, разрушающие/чувствительные действия через подтверждение; агент — отдельная идентичность с правами «агент ∩ пользователь».
4. Для шлюза: **самописный набор адаптеров под каждого провайдера** (текущее) vs **LiteLLM-паттерн** (принято) — универсальный OpenAI-совместимый адаптер конфигом; код адаптера один, провайдеры добавляются данными.
5. Для MCP: **полный CRUD-toolset сразу** vs **курируемый read-heavy v1 (принято)** — минимальный blast-radius на период, пока не отработаны prompt-injection-митигации.

## Решение

### 1. Ступенчатая автономия (зафиксировано пользователем)

| Уровень | Действия | Режим |
|---------|----------|-------|
| L1 Read | Чтение задач/каналов/wiki/отчётов в правах пользователя | Автономно |
| L2 Create/Draft | Создание задач, черновики сообщений/комментариев/страниц, свои артефакты | Автономно (в правах пользователя) |
| L3 Confirm | Delete любых сущностей; изменение **чужих** задач; отправка сообщений **от имени** пользователя; RBAC/настройки | **Только с явным подтверждением** пользователя (per-action, UI-карточка) |

Уровень — свойство инструмента/операции, не сессии: один диалог может содержать L1/L2 автономно и L3-паузы на подтверждение. Подтверждение одноразовое per-action (не «разрешить всё на час»).

### 2. Agent-principal: идентичность и RBAC/ERD

Агент — **отдельный principal**, не «пользователь с токеном». Эффективные права: **`агент ∩ пользователь`** — агент никогда не превышает права инициировавшего пользователя, а его собственный ceiling режется конфигурацией.

**AgentPrincipal** (новая Ent-схема):

```
id             UUID v4 PK
name           string                      # "hub-ai copilot", "mcp:<pat-name>"
kind           enum { internal, mcp }      # встроенный агент / внешний через MCP
max_level      enum { read, create, confirm }   # ceiling автономии (§1)
scopes         JSON []string               # доменные скоупы: tracker.read, wiki.read, tms.read, messenger.read, tracker.create ...
enabled        bool, Default true
version        int                         # optimistic lock
created_at / updated_at                    # UTC
```

**AgentActionLog** (audit-запись, дополняет существующий `audit/`):

```
id             UUID v4 PK
user_id        UUID FK → User              # от чьего имени (инициатор)
agent_id       UUID FK → AgentPrincipal    # какой агент
task_id        string                      # ID задачи/диалога/run — сквозная корреляция цепочки шагов
tool           string                      # "tracker.search", "wiki.read_page", "tracker.create_issue"
level          enum { read, create, confirm }
decision       enum { auto, confirmed, denied, blocked }   # blocked = срез прав/лимит
payload_digest string                      # SHA-256 аргументов (без хранения чувствительного тела)
result_status  string                      # ok / error / rate_limited
created_at     UTC, Immutable
Индексы: (user_id, created_at), (agent_id, created_at), (task_id)
```

Каждое действие агента порождает ровно одну запись **user + agent + task-id** — требование пользователя (GZ-SEC-10/21). RBAC-проверка выполняется дважды: стандартный user-RBAC (как если бы действовал пользователь) **и** scope/level-проверка agent-principal; отказ любой из них → `decision=blocked` + audit.

### 3. BYOK и лимиты (зафиксировано пользователем)

- BYOK-ключи провайдеров — **AES-GCM at rest** (`platform/crypto`, как сейчас), маскирование в API, запрет логирования.
- **Org-дневные лимиты токенов с hard-stop:** счётчик расхода per-org/day (расширение `AIUsageDaily`); при достижении лимита — немедленный отказ новых запросов (`429 {"error","request_id"}`) **и алерт** админу (in-app + Telegram-канал админа, GZ-SEC-26). Никакого «мягкого» перерасхода.
- Дополнительные слои квот (LiteLLM-паттерн, §5): virtual keys с бюджетами per-user/per-team, per-provider RPM (уже есть `provider_rpm.go`). Порядок проверки: virtual key budget → org daily hard-stop → provider RPM.

### 4. MCP-сервер (зафиксировано: PAT со скоупами + курируемый read-heavy toolset v1)

**Транспорт:** Streamable HTTP (`POST /api/mcp` + SSE-стрим ответов) — текущий канонический транспорт спецификации MCP; stdio не нужен (Hub — сетевой сервер). Реализация в hub-ai (официальный `modelcontextprotocol/go-sdk` как кандидат; финальный выбор — spike перед фазой AGENT-2, критерии §5.3 playbook).

**Аутентификация:** **PAT (Personal Access Token) со скоупами** — эволюция существующей `APIKey`-схемы: `+ scopes JSON`, `+ agent_principal_id FK`. Хранение — SHA-256 hash (как сейчас), префикс для отображения, TTL обязателен (дефолт 90 дней — политика ротации GZ-SEC-04). Каждый PAT привязан к пользователю и agent-principal → внешний MCP-клиент действует строго в срезе «агент ∩ пользователь». Сравнение токенов — constant-time.

**Стартовый toolset v1 (курируемый, read-heavy):**

| Инструмент | Скоуп | Описание |
|------------|-------|----------|
| `tracker.search_issues` | tracker.read | Поиск задач (JQL-подмножество), user-scope фильтр обязателен |
| `tracker.get_issue` | tracker.read | Карточка задачи + комментарии |
| `messenger.read_channel` | messenger.read | Последние N сообщений канала (только каналы, где user — member) |
| `wiki.search` / `wiki.read_page` | wiki.read | Поиск и чтение страниц (space-RBAC, GZ-SEC-07/11) |
| `tms.get_run_report` | tms.read | Отчёт прогона / Allure-summary |
| `tms.list_failed_cases` | tms.read | Упавшие кейсы прогона |
| `search.global` | search.read | Глобальный поиск (существующий user-scope инвариант NF-03) |
| `tracker.create_issue` | tracker.create | **Единственный** write v1 (уровень L2 — создание своего) |

Расширение toolset (write/L3-инструменты, календарь, CRM) — только через дополнение этого ADR после отработки митигаций §6 на v1.

### 5. hub-ai — LiteLLM-паттерн (зафиксировано пользователем)

Целевая форма шлюза (вынос №1 по ADR-MICROSERVICES, дельта к текущему `ai/`):

1. **Универсальный OpenAI-совместимый адаптер:** провайдер добавляется **конфигом** (`base_url + api_key + model list`) без нового Go-кода; существующие 5 адаптеров сохраняются для несовместимых API (GigaChat/Yandex), остальное мигрирует на generic-адаптер.
2. **Каталог моделей + fallback-маршрутизация:** реестр моделей с capabilities/ценой; per-role доступ к моделям; fallback-цепочки (уже есть `fallback.go` — обобщается на каталог).
3. **Virtual keys / бюджеты:** внутренние ключи per-user/team с бюджетами и трекингом расходов по диалогам; поверх — org-дневной hard-stop (§3).
4. **Внутренний OpenAI-совместимый endpoint** (`POST /v1/chat/completions` внутри hub-ai): единая точка для MCP-сервера и всех модулей Hub → каждый AI-вызов проходит через один слой аудита, лимитов и routing. Прямые вызовы провайдеров из других модулей запрещаются.

Agent-runtime и MCP-сервер живут в hub-ai (совмещается с выносом №1): изоляция BYOK-секретов и независимое масштабирование — драйверы выноса по ADR-MICROSERVICES.

### 6. Риски и митигации

| Риск | Митигация |
|------|-----------|
| **Prompt injection** (внешний текст в задаче/wiki/сообщении инструктирует агента) | Ступенчатая автономия — инъекция не может достичь delete/чужого/от-имени без явного подтверждения человека; tool-результаты помечаются как untrusted data в контексте; L3-подтверждение показывает пользователю **конкретное** действие и аргументы; read-heavy toolset v1 |
| **Exfiltration через MCP** (внешний клиент выкачивает данные PAT'ом) | Права = агент ∩ пользователь (не больше, чем user видит сам); скоупы PAT сужают домены; user-scope фильтр в каждом read-инструменте (инвариант NF-03); rate limit MCP-вызовов per-PAT; полный audit user+agent+task-id → аномалии видны; TTL и ротация PAT 90 дней; revoke PAT = мгновенный отказ |
| **Budget burn** (агентный цикл сжигает токены) | Org-дневной hard-stop + алерт (§3); virtual key бюджеты per-user/team; лимит шагов на task-id в agent-runtime; RPM per-provider |
| Compromised PAT | Hash-хранение, constant-time compare, TTL, скоупы, instant revoke, audit-корреляция по agent_id |
| Аудит-шум / нечитаемость | `task_id` группирует цепочку шагов; `payload_digest` вместо полного тела (без утечки секретов в лог) |
| Расползание toolset без ревью | Расширение только дополнением ADR + security-review каждого нового инструмента |
| Coverage-риск нового кода | MCP/agents — security-critical: **гейт ≥80%** (GZ-QA-01/02) |

### 7. План фаз реализации

| Фаза | Содержание | Выход |
|------|-----------|-------|
| AGENT-0 Spec | sa: FR/ERD (AgentPrincipal, AgentActionLog, PAT-scopes), контракты MCP-инструментов; qa-manual: TC-таблицы; spike выбора MCP SDK | verified спека + ADR-приложение по SDK |
| AGENT-1 Identity | Ent-схемы + RBAC-срез «агент ∩ пользователь» + audit-pipeline (user+agent+task-id) | Principal-модель в ядре |
| AGENT-2 MCP v1 | MCP-сервер (Streamable HTTP), PAT со скоупами, read-toolset §4 + `tracker.create_issue`, rate limits | Внешние клиенты (Claude/Cursor) читают Hub |
| AGENT-3 Gateway | LiteLLM-дельта: generic OpenAI-адаптер, каталог моделей, virtual keys/бюджеты, внутренний `/v1/chat/completions`, org hard-stop + алерт | Единый шлюз для модулей и MCP |
| AGENT-4 Autonomy | Agent-runtime: L1/L2 автономные цепочки, L3 confirm-UI, лимит шагов, internal-агент в мессенджере | Ступенчатая автономия end-to-end |
| AGENT-5 Hardening | security-auditor (§6), negative-тесты (injection/exfiltration-сценарии), coverage ≥80%, docs sync | Приёмка |

Совмещается с Track P M3 (вынос hub-ai): AGENT-1/2 могут идти в монолите, AGENT-3+ — уже в границах будущего сервиса (те же пакеты, вынос — route-flip).

## Последствия

- **+** Hub становится agent-native платформой (тренд 2026) и открывается для внешних AI-клиентов через MCP без кастомных интеграций; провайдеры подключаются конфигом («любая ИИшка»); расходы управляемы (hard-stop + бюджеты); каждый шаг агента атрибутирован и восстановим по audit.
- **−** Существенная новая поверхность атаки (MCP endpoint, PAT, агентные действия) → митигации §6 + гейт ≥80% + security-audit каждой фазы; сложность RBAC растёт (второй principal-слой); L3-подтверждения добавляют трение в UX (осознанный трейд-офф).
- **Долг:** read-heavy toolset ограничивает пользу MCP v1 (write-расширение — после отработки митигаций); LibreChat-паттерны UI (ветвление диалогов, side-by-side сравнение моделей) — отдельная FE-фаза, этим ADR не гейтится.

## Проверка соответствия (для arch-ver / security-auditor)

`ent/schema/` содержит AgentPrincipal + AgentActionLog с указанными полями/индексами; APIKey расширен scopes; каждое действие агента в тестах порождает audit-запись с user_id+agent_id+task_id; MCP-инструменты применяют user-scope фильтр (negative-тест: чужие данные недоступны); при исчерпании org-лимита API отвечает 429 и создаётся алерт; L3-действие без подтверждения невозможно (negative-тест `decision=blocked`); прямых вызовов AI-провайдеров вне hub-ai нет (grep по base_url провайдеров вне `internal/ai`).
