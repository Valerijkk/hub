# Hub — Единый Roadmap: расширение + микросервисы (2026-07-02)

> Принятые решения пользователя (2026-07-02): релиз **не гейтим**; v4 — **паритет Bitrix/Lark**; accounting — **своя бухгалтерия**; направление — **микросервисы по блокам**; новые модули — **wiki + agent-native AI/MCP + SSO/LDAP — все берём**.
> Основания: `ADR-ARCH-MICROSERVICES-2026`, A/B-планы `.planning/decisions/ab-microservices/*` и `.planning/archive/ab-tests/*`, `.planning/research/MARKET-AND-NEEDS-2026-07-02.md`, `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`.
> Оркестратор: `team-lead` (Fable 5), эскалация `team-lead-opus`.

---

## 1. Два параллельных трека

Работаем двумя переплетающимися треками (волны ≤4 агента, границы файлов не пересекаются):

- **Track F — Функционал** (главная ценность для пользователя): углубление v4 до паритета + новые модули.
- **Track P — Платформа** (архитектура): Modular Monolith Hardening → strangler-fig вынос сервисов.

Track P Фаза 0 (Hardening) идёт **рано и параллельно** Track F — она чинит tech-debt и полезна независимо от исхода split.

---

## 2. Каталог микросервисов (утверждён)

См. `ADR-ARCH-MICROSERVICES-2026` §3 (уточнено 2026-07-03 по gray-zones-решениям):
`hub-core` (ядро: auth/rbac/project/**tracker — исключён из выноса**/sprint/tms-условно/dashboard/…) · `hub-ai` (вынос №1) · **realtime-трио: `hub-ws` / `hub-messenger` / `hub-calls`** (вынос №2, три отдельных сервиса) · `hub-notification` (пилот) · уже вынесены `hub-mail`, `meilisearch-indexer` · внешний `LiveKit`.
БД — **schema-per-service в одном PostgreSQL**, cross-schema запрещён грантами; CQRS-проекции через NATS. Go/No-Go после пилота — композитная метрика. Полный реестр решений: `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`.

Детальные фазы Track P (M0 ADR+K3s+OTel → M1 Hardening → M2 пилот notification → M3 ai → M4 calls → M5 ws+messenger → M6 tms; tracker опционально) — в `.planning/decisions/ab-microservices/PLAN-team-lead-fable.md` (база) + глубина рисков/observability из `PLAN-team-lead-opus.md`.

---

## 3. Новые модули в scope (из рыночного анализа)

| Модуль | Обоснование | Куда | Предусловие |
|--------|-------------|------|-------------|
| **Knowledge base / wiki** | Крупнейший пробел vs Notion/Confluence/Yonote; «держатель» пользователей; окно ухода on-prem Confluence | новый домен в hub-core (позже — кандидат в сервис) | ba/sa спека, дизайн |
| **Agent-native AI + MCP-сервер** | Тренд 2026: автономные агенты + MCP как lingua franca; у Hub только копилот `@hub-ai` | `hub-ai` (совмещается с выносом №1) | arch ADR по агентам/памяти |

> **Уточнение scope hub-ai (решение пользователя 2026-07-02): «подключается любая ИИшка».** Референсы: **LiteLLM** (шлюз: OpenAI-совместимый endpoint, 100+ провайдеров, virtual keys, бюджеты, RBAC) и **LibreChat** (чат-хаб: мульти-провайдер в одном диалоге, ветвление, сравнение моделей, per-role доступ к моделям, агенты+MCP). Дельта к текущему `ai`-модулю (5 адаптеров + BYOK): (1) универсальный OpenAI-совместимый адаптер — провайдер добавляется конфигом (base_url+key+модели) без кода; (2) каталог моделей + fallback-маршрутизация + per-role доступ к моделям; (3) virtual keys / бюджеты per-user/team с hard-stop (согласовано в GZ-SEC-22) + трекинг расходов по диалогам; (4) внутренний OpenAI-совместимый endpoint hub-ai как единая точка для MCP и модулей Hub (аудит+лимиты). UI: ветвление диалогов и сравнение моделей side-by-side (LibreChat-паттерн), RAG по докам проекта (Open WebUI-паттерн). |
| **SSO / LDAP / AD** | Барьер для RU-корпоратов on-prem (у Bitrix24 On-Premise есть) | auth/identity | **ADR-AUTH-SSO (proposed)** — снимает forbidden из `tz_constraints.md`; требуется до реализации |

> **Важно:** SSO/LDAP сейчас в `docs/tz_constraints.md` — **FORBIDDEN до ADR**. Пользователь одобрил включение → `arch` готовит `ADR-AUTH-SSO-2026`, `arch-ver`+`security-auditor` верифицируют, только затем реализация. Magic Link остаётся дефолтным методом; SSO — дополнительный корпоративный.

---

## 4. Track F — приоритеты углубления (из EXPANSION эпики B/C)

- **v4 до паритета (эпик B):** intranet (Mighty-lite), calendar (room booking/availability/recurring/iCal), CRM (companies/активности/воронка), HR (org chart/handbook/баланс отпусков), **bases → Lark-паритет** (formula engine, relations, views — план готов: `.planning/archive/ab-tests/PLAN-bases-*.md`), accounting (своя бухгалтерия: баланс/cash flow/НДФЛ/мультивалюта — XL). FE-декомпозиция скелетов (1 файл → модуль).
- **Глубина ядра (эпик C):** custom fields builder UI, Gantt pro, TMS-отчёты уровня TestOps, notification digest, кастомные дашборды, issue↔message.
- **UX polish (эпик D)** и **quality/compliance (эпик E: WCAG/Lighthouse/coverage/perf)** — сквозняком.

---

## 5. Первые волны (согласовано со schedule-A/B)

**Волна 1 (дни 1–3, 4 потока, без пересечений):**
- `devops-engineer` — фоновый ops-трек: FA-06/PG-03 ротация секретов + BFG purge + SSH deploy + tag v2.5.0 (**не гейтит** разработку).
- `be1` — быстрый tech-debt: GIPHY-ключ → `import.meta.env` (`messenger/gif-picker.tsx`, уже в git status), magic-link `utf8.RuneCountInString` (`auth/handler.go`), N+1 в `audit/service.go`+`repository.go`.
- `security-auditor` — E5: LiveKit `iss` constant-time compare, RequireAuth fail-open (calls/auth middleware) + спека FA-02.
- `ba2` — US/AC для Intranet и Calendar (задел паритета, конвейер аналитики опережает реализацию).
- Ворота: регрессия BE+FE green, docs-sync, `be-ver`/`fe-ver`/security review, `/qa-gate`, приёмка `team-lead`. Параллельно `docs-writer` сверяет дрейф ROADMAP-чекбоксов Phase 8/9/12.

**Волна 2:** старт Track P Фаза 0 (Hardening: `*ent.Client` за интерфейс) ∥ Track F аналитика v4 (sa1/sa2 спеки) ∥ дизайн (designer) первых экранов.
**Волна 3:** ADR-волна (`arch`: ADR-MICROSERVICES детализация C4, ADR-AUTH-SSO, ADR-AI-agents) + первые реализации v4.

Полное расписание (~12–17 волн, ~2.5 мес) — `.planning/archive/ab-tests/PLAN-schedule-fable.md` (база, карта владения модулями + heatmap) + `PLAN-schedule-opus.md` (анти-боттлнек-конвейер).

---

## 6. Ближайший шаг

1. `arch` (через `team-lead`) детализирует C4 и три ADR (MICROSERVICES уже accepted; SSO, AI-agents — proposed).
2. `tech-lead` нарезает Волну 1 в атомарные планы.
3. `team-lead` запускает Волну 1.

Все ворота (регрессия/verify/docs-sync/qa-gate) — по `.cursor/REGULATION.md` и `.cursor/playbooks/team-workflow.md`.
