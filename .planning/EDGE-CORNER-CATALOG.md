# Hub — Каталог Happy / Edge / Corner cases

> **Статус:** planning/QA артефакт · создан 2026-07-10 · владелец: `qa-manual`  
> **Назначение:** максимально полный каталог сценариев для снижения багов в разработке и тестировании. Источник заготовок для будущих `TC-*` в `docs/TESTING.md` и автотестов.  
> **Не заменяет** `docs/TESTING.md` (канонические TC) — дополняет его foresight-слоем.  
> **Связи:** `.planning/JULY-2026-REBUILD-BRIEF.md` §3 (CAL-SYNC C1–C15) · `seeds/SEED-002-calendar-calls-email-sync.md` · будущий `testframework/TESTING_GUIDE.md` · `.cursor/playbooks/testing.md`  
> **SFU:** остаёмся на **LiveKit** (mediasoup / SEED-001 = REJECTED).

---

## Оглавление

1. [Методология](#методология)
2. [Маппинг на QA-конвейер](#маппинг-на-qa-конвейер)
3. [Домены](#домены)
   - [1. Auth](#1-auth)
   - [2. RBAC / Org](#2-rbac--org)
   - [3. Tracker](#3-tracker)
   - [4. Sprint](#4-sprint)
   - [5. Messenger](#5-messenger)
   - [6. Calls + LiveKit](#6-calls--livekit)
   - [7. Calendar](#7-calendar)
   - [8. TMS](#8-tms)
   - [9. Notification / Search / WS](#9-notification--search--ws)
   - [10. Storage / Files](#10-storage--files)
   - [11. Dashboard](#11-dashboard)
   - [12. v4: Intranet / CRM / HR / Accounting / Bases](#12-v4-intranet--crm--hr--accounting--bases)
   - [13. AI](#13-ai)
   - [14. Platform cross-cutting](#14-platform-cross-cutting)
4. [Top-50 P0 кросс-доменных рисков](#top-50-p0-кросс-доменных-рисков)
5. [Сводка счётчиков](#сводка-счётчиков)
6. [История изменений](#история-изменений)

---

## Методология

### Как пользоваться каталогом

| Шаг | Кто | Действие |
|-----|-----|----------|
| 1 | `qa-manual` | Берёт строку `ECC-*` → оформляет полный `TC-<MODULE>-NN` в `docs/TESTING.md` (шаги, AC, колонка «Автотест» = TODO/ручной) |
| 2 | `qa-manual-ver` | Верифицирует TC (техники, трассируемость US/FR) → `verified: true` |
| 3 | `qa-auto1/2` | Автоматизирует verified TC; обновляет Gap в этом каталоге и колонку «Автотест» |
| 4 | `be*`/`fe*` | Unit рядом с кодом (вне QA-конвейера) — при закрытии Gap помечают Partial→Covered |

### Типы сценариев

| Тип | Определение | Примеры |
|-----|-------------|---------|
| **Happy** | Основной успешный путь | Magic Link → JWT; create issue → key |
| **Neg** | Невалидный ввод, 401/403/404, business reject | wrong code, Viewer create issue, expired invite |
| **Edge** | Границы: лимиты, пустые списки, max size, TTL, concurrent | rate 3/10m, empty board, 25 MB, JWT 15m |
| **Corner** | Редкие комбинации / races / mid-flight failures | version race, double-submit, LiveKit restart mid-call, RSVP after delete |

### Приоритеты

| P | Критерий |
|---|----------|
| **P0** | Security / data loss / money-time / authz bypass / production blocker |
| **P1** | Core UX broken, wrong state, visible inconsistency |
| **P2** | Rare path, polish, nice-to-have coverage |

### Gap-легенда

| Gap | Значение |
|-----|----------|
| **Covered** | Есть автотест (Unit/API) и/или TC с привязкой |
| **Partial** | Частичное покрытие (happy/unit есть, edge/corner нет) |
| **TODO** | Нет dedicated теста; нужен TC + авто/ручной |
| **Manual** | Только ручной / E2E exploratory (сложно автоматизировать) |
| **Future** | Фича ещё не в коде (RSVP, CalendarEvent) — TC писать после ba/sa |

### Как пополнять

1. Новый FR/US / edge из discuss-phase → добавить `ECC-<DOM>-NNN` сюда **до** кода.
2. Баг в проде / flaky → reverse-engineer corner → каталог + TC.
3. Brief/seed (как CAL-SYNC C1–C15) → импорт в домен Calendar с префиксом `ECC-CAL-C*`.
4. Раз в milestone: `qa-manual` ревизия Gap-колонки vs `backend/testframework/` + E2E specs.
5. Не дублировать дословно TC из `docs/TESTING.md` — ссылаться (`TC-AUTH-01`); здесь держать **шире** (особенно Edge/Corner).

### Техники тест-дизайна (чеклист при добавлении)

- Классы эквивалентности + границы (TTL, sizes, counts)
- Таблица решений (RBAC × action)
- Переходы состояний (sprint planning→active→completed; call active→ended)
- Pairwise для комбинаций (channel type × role × readonly)
- Race / idempotency / clock skew / soft-delete restore

---

## Маппинг на QA-конвейер

```
ECC-каталог (.planning/EDGE-CORNER-CATALOG.md)
        │  qa-manual выбирает P0/P1 → полный TC
        ▼
docs/TESTING.md  (TC-*-NN + колонка «Автотест»)
        │  qa-manual-ver
        ▼
Автотест: backend/testframework/{domain}/  |  Vitest  |  Playwright E2E
        │  qa-auto-ver
        ▼
Gap: TODO → Partial → Covered  (обновить этот файл)

Юниты be/fe — параллельно, вне конвейера (rule 56).
```

**Целевой guide:** `testframework/TESTING_GUIDE.md` (будет создан другим агентом) — должен ссылаться на этот каталог как на источник edge/corner backlog.

**Brief:** `.planning/JULY-2026-REBUILD-BRIEF.md` §3.4 C1–C15 → строки `ECC-CAL-C01`…`ECC-CAL-C15` (+ расширения) в [§7 Calendar](#7-calendar).

---

## Домены

### 1. Auth

> US-M01, US-M08, US-M09 · NFR-05/09/10 · TC-AUTH-* · ADR-AUTH-01 (TG ≠ login)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AUTH-001 | Happy | Magic Link: send → email code → verify → access+refresh | 200, JWT 15m + refresh 30d | API/E2E | P0 | US-M01, TC-AUTH-01/02 | Covered |
| ECC-AUTH-002 | Happy | Refresh: валидный refresh → новый access | 200, старый access невалиден после expiry | API | P0 | TC-AUTH-04, NFR-09 | Covered |
| ECC-AUTH-003 | Happy | Logout → refresh revoked; sessions cleared | 200; refresh 401 | API/E2E | P0 | TC-AUTH-12 | Covered |
| ECC-AUTH-004 | Happy | GET/PUT /auth/me профиль + timezone | 200 | API/E2E | P1 | TC-AUTH-06/07 | Covered |
| ECC-AUTH-005 | Happy | Онбординг complete | onboarding=true | E2E | P1 | US-M08, TC-AUTH-11 | Covered |
| ECC-AUTH-006 | Happy | Список сессий + отзыв чужой сессии | 200, session gone | API | P1 | TC-AUTH-08/09 | Partial |
| ECC-AUTH-007 | Neg | Verify с неверным кодом | 400/401; fail counter++ | API | P0 | TC-AUTH-03 | Covered |
| ECC-AUTH-008 | Neg | Verify без кода в Redis | 400/401 | API | P0 | — | Covered |
| ECC-AUTH-009 | Neg | Refresh expired / revoked | 401 | API | P0 | TC-AUTH-05 | Covered |
| ECC-AUTH-010 | Neg | GET /auth/me без Authorization | 401 | API | P0 | — | Covered |
| ECC-AUTH-011 | Neg | Malformed Bearer / garbage JWT | 401 | API | P0 | — | Covered |
| ECC-AUTH-012 | Neg | Невалидный email (no @, empty, >254) | 400 | API | P1 | — | Partial |
| ECC-AUTH-013 | Neg | Invite-only: новый email без invite | blocked 403/400 | API | P0 | US-M09, TC invite | Covered |
| ECC-AUTH-014 | Neg | Deactivated user: любой authenticated call | 401/403 | API | P0 | TC-ADMIN-02 | Covered |
| ECC-AUTH-015 | Neg | Telegram login endpoints (legacy) | 404 — login удалён | API | P0 | ADR-AUTH-01 | Partial |
| ECC-AUTH-016 | Edge | Auth rate limit: 4-й запрос / 10 мин | 429 | API | P0 | NFR-05, TC-AUTH-10 | Covered |
| ECC-AUTH-017 | Edge | Brute-force lockout на verify | lockout после N fails | API | P0 | — | Covered |
| ECC-AUTH-018 | Edge | Lockout TTL истекает → снова можно | verify OK | API | P1 | — | Covered |
| ECC-AUTH-019 | Edge | Access token ровно на границе 15m | до TTL OK, после 401 | API | P1 | NFR-09 | TODO |
| ECC-AUTH-020 | Edge | Refresh на границе 30d | до TTL OK, после 401 | API | P1 | NFR-10 | TODO |
| ECC-AUTH-021 | Edge | Multibyte / emoji magic code length | reject по rune length | API | P1 | — | Covered |
| ECC-AUTH-022 | Edge | Concurrent refresh same token | один успех / rotation-safe | API | P0 | — | TODO |
| ECC-AUTH-023 | Edge | DEV_MODE magic link returns code in body | только non-prod | API | P1 | — | Covered |
| ECC-AUTH-024 | Corner | Double-submit verify одного кода | один 200, второй reject | API | P0 | — | TODO |
| ECC-AUTH-025 | Corner | Logout на одном устройстве, другой session жив | только target revoked | API | P1 | — | Partial |
| ECC-AUTH-026 | Corner | Clock skew: client clock +5m vs server JWT | 401 или clock-skew window | API/Manual | P1 | — | TODO |
| ECC-AUTH-027 | Corner | Pending user 7d без approve → purge | user deleted | API | P0 | org TZ | Covered |
| ECC-AUTH-028 | Corner | Registration mode flip mid-signup | consistent reject/accept | API | P1 | — | Partial |
| ECC-AUTH-029 | Corner | Constant-time compare на коде (timing) | no early-exit leak | Unit | P0 | security | Partial |
| ECC-AUTH-030 | Happy | TG notify bind (не login) | bind OK; login path absent | API | P1 | ADR-AUTH-01 | Covered |
| ECC-AUTH-031 | Neg | TG verify invalid / rate limited | 400/429 | API | P1 | — | Covered |
| ECC-AUTH-032 | Edge | Resend magic link while previous valid | new code invalidates old OR both TTL | API | P1 | — | Partial |
| ECC-AUTH-033 | Corner | Instant revoke refresh after GDPR/deactivate | all sessions dead | API | P0 | US-S03 | TODO |
| ECC-AUTH-034 | Neg | SSO endpoints до ADR-AUTH-SSO | 404 / not implemented | API | P2 | TZ SSO | Future |

**Домен auth: 34 кейса**

---

### 2. RBAC / Org

> US-M10, US-M09 · system_role founder · project roles Owner/Developer/Tester/Viewer · TC-RBAC-*, TC-PROJ-*

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-RBAC-001 | Happy | Owner: полный CRUD project/members | 200 | API | P0 | US-M10 | Partial |
| ECC-RBAC-002 | Happy | Developer: create/update issues | 200 | API | P0 | — | Partial |
| ECC-RBAC-003 | Happy | Custom role bitmask Add/Remove/Has | permissions correct | Unit | P0 | TC-RBAC-01..05 | Covered |
| ECC-RBAC-004 | Happy | Invite create + join by code | member added | API | P0 | TC-PROJ-07/08 | Covered |
| ECC-RBAC-005 | Happy | Org: founder exists; HasOrganization | true | API | P0 | TZ org | Covered |
| ECC-RBAC-006 | Happy | Admin approve pending user | user active | API | P0 | — | Covered |
| ECC-RBAC-007 | Neg | Viewer: create issue / delete project | 403 | API | P0 | US-M10 | Partial |
| ECC-RBAC-008 | Neg | Non-member: access project resources | 403/404 | API | P0 | — | Partial |
| ECC-RBAC-009 | Neg | Remove last Owner | 400 reject | API | P0 | TC-PROJ-10 | Covered |
| ECC-RBAC-010 | Neg | Non-admin: set registration mode | 403 | API | P0 | — | Covered |
| ECC-RBAC-011 | Neg | Expired invite code | 400/410 | API | P0 | TC-PROJ-08 | Covered |
| ECC-RBAC-012 | Neg | Join invite already member | idempotent 200 or 409 | API | P1 | — | TODO |
| ECC-RBAC-013 | Neg | Second founder attempt | reject — один founder | API | P0 | TZ | Partial |
| ECC-RBAC-014 | Neg | Reject pending user | not in org | API | P1 | — | Covered |
| ECC-RBAC-015 | Edge | Role change mid-request (cache stale) | RBAC cache invalidate | API | P0 | rbac cache tests | Partial |
| ECC-RBAC-016 | Edge | Empty members list after remove all non-owners | OK, owner remains | API | P1 | — | TODO |
| ECC-RBAC-017 | Edge | Permission bit overflow / unknown bit | safe no-op or 400 | Unit | P2 | — | Partial |
| ECC-RBAC-018 | Corner | Concurrent role update same member | last-write / version | API | P1 | NFR-24 | TODO |
| ECC-RBAC-019 | Corner | Demote self from Owner while sole owner | reject | API | P0 | — | TODO |
| ECC-RBAC-020 | Corner | Archived project: write ops | 403/400 read-only | API | P1 | US-S10 | Partial |
| ECC-RBAC-021 | Corner | Soft-deleted project: member APIs | 404 | API | P1 | NFR-11 | TODO |
| ECC-RBAC-022 | Happy | Default role permissions matrix | Owner>Dev>Tester>Viewer | Unit | P0 | TC-RBAC-05 | Covered |
| ECC-RBAC-023 | Neg | Cross-project IDOR (issue of other project) | 403/404 | API | P0 | security | Partial |
| ECC-RBAC-024 | Edge | Invite reuse after max uses / one-time | second join fail | API | P1 | — | TODO |
| ECC-RBAC-025 | Corner | Admin deactivate user mid-WS session | WS disconnect / 401 | Manual | P0 | — | TODO |
| ECC-RBAC-026 | Happy | Middleware RequirePermission smoke | allow/deny | API | P0 | — | Covered |
| ECC-RBAC-027 | Neg | Multi-org attempt (create 2nd org) | reject / out of scope | API | P1 | NFR-28 | TODO |

**Домен rbac/org: 27 кейсов**

---

### 3. Tracker

> US-M03, US-M11, US-M12, US-C04 · NFR-24 · TC-TRCK-* · optimistic `version`

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-001 | Happy | Create issue → KEY-N autoincrement | 201, key | API/E2E | P0 | TC-TRCK-01/02 | Covered |
| ECC-TRCK-002 | Happy | Update issue with correct version | 200, version++ | API | P0 | TC-TRCK-03 | Covered |
| ECC-TRCK-003 | Happy | Valid workflow transition | status changed | API | P0 | TC-TRCK-10 | Covered |
| ECC-TRCK-004 | Happy | Issue links blocks/relates/… | 201 | API | P1 | TC-TRCK-06 | Covered |
| ECC-TRCK-005 | Happy | Comment + @mention → notif | 201 + notif | API | P1 | TC-TRCK-05 | Partial |
| ECC-TRCK-006 | Happy | Kanban move column = status | board updated | E2E/API | P0 | US-M11 | Partial |
| ECC-TRCK-007 | Happy | JQL filter type+priority | filtered list | API | P1 | TC-TRCK-09 | Covered |
| ECC-TRCK-008 | Happy | Time log manual + timer start/stop | time_spent++ | API | P1 | TC-TRCK-07/08 | TODO |
| ECC-TRCK-009 | Happy | All 9 issue types create | 201 each | API | P1 | TC-TRCK-02 | Covered |
| ECC-TRCK-010 | Happy | Soft delete issue | deleted_at set | API | P1 | TC-TRCK-12 | Partial |
| ECC-TRCK-011 | Neg | Update with stale version | 409 conflict | API | P0 | TC-TRCK-04, NFR-24 | Covered |
| ECC-TRCK-012 | Neg | Invalid workflow transition | 400 | API | P0 | TC-TRCK-11 | Covered |
| ECC-TRCK-013 | Neg | Non-member create issue | 403 | API | P0 | — | Partial |
| ECC-TRCK-014 | Neg | Unknown issue key | 404 | API | P1 | — | Covered |
| ECC-TRCK-015 | Neg | Empty title / title > max | 400 | API | P1 | — | Partial |
| ECC-TRCK-016 | Neg | Link to self / duplicate link | 400 | API | P2 | — | TODO |
| ECC-TRCK-017 | Neg | JQL syntax error | 400 clear error | API | P1 | — | Partial |
| ECC-TRCK-018 | Edge | WIP limit exceeded on column | 400/409 warn | API | P1 | US-M11 | Partial |
| ECC-TRCK-019 | Edge | Empty board (0 issues) | 200 empty | API/E2E | P2 | — | TODO |
| ECC-TRCK-020 | Edge | Pagination offset/limit boundaries | page correct; overshoot empty | API | P1 | NFR-14 | Partial |
| ECC-TRCK-021 | Edge | Max checklist items / labels | reject over limit | API | P2 | — | TODO |
| ECC-TRCK-022 | Edge | Concurrent creates → unique keys | no duplicate KEY-N | API | P0 | — | Partial |
| ECC-TRCK-023 | Edge | Swimlanes + filters combo | correct subset | E2E | P2 | US-M11 | TODO |
| ECC-TRCK-024 | Corner | Two tabs: both edit, second 409, retry | UI shows conflict | E2E/Manual | P0 | NFR-24 | TODO |
| ECC-TRCK-025 | Corner | Double-submit create (network retry) | one issue or idempotent key | API | P0 | — | TODO |
| ECC-TRCK-026 | Corner | Move + delete race | consistent 404/409 | API | P1 | — | TODO |
| ECC-TRCK-027 | Corner | Restore soft-deleted then update version | version continues | API | P1 | NFR-11 | TODO |
| ECC-TRCK-028 | Corner | Assignee deactivated user | 400 or keep + warn | API | P1 | — | TODO |
| ECC-TRCK-029 | Corner | Automation rule fires on transition mid-edit | no lost update / ordered | API | P1 | automation tests | Partial |
| ECC-TRCK-030 | Happy | Custom fields CRUD on issue | 200 | API | P1 | — | Partial |
| ECC-TRCK-031 | Neg | Saved JQL share to non-member | no leak | API | P0 | US-C04, NF-03 | TODO |
| ECC-TRCK-032 | Edge | Board with 100+ issues load | <500ms target | Manual/Perf | P2 | perf | TODO |
| ECC-TRCK-033 | Corner | Parent epic delete with children | cascade policy | API | P1 | — | TODO |
| ECC-TRCK-034 | Happy | GitHub/GitLab key autolink | link created | API | P2 | US-S04/S06 | Partial |
| ECC-TRCK-035 | Neg | String ID instead of UUID in path | 400 | API | P1 | NFR-25 | Partial |
| ECC-TRCK-036 | Corner | Issue in completed sprint + reopen | sprint rules | API | P1 | sprint | TODO |
| ECC-TRCK-037 | Edge | Estimate 0 / negative / huge | 400 on neg; 0 OK | API | P2 | — | TODO |
| ECC-TRCK-038 | Happy | Multiple boards per project | isolated columns | API | P1 | US-M11 | Partial |
| ECC-TRCK-039 | Corner | Workflow edit removes status with issues | migrate or block | API | P0 | — | TODO |
| ECC-TRCK-040 | Neg | Comment on deleted issue | 404 | API | P1 | — | TODO |

**Домен tracker: 40 кейсов**

---

### 4. Sprint

> US-S01 · TC-SPRINT-*

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-001 | Happy | Create sprint planning | 201, status=planning | API | P0 | TC-SPRINT-01 | Covered |
| ECC-SPR-002 | Happy | Start sprint | active | API | P0 | TC-SPRINT-02 | Covered |
| ECC-SPR-003 | Happy | Complete → incomplete issues moved | completed + move | API | P0 | TC-SPRINT-07 | Covered |
| ECC-SPR-004 | Happy | Burndown points | 200 series | API | P1 | TC-SPRINT-04 | Covered |
| ECC-SPR-005 | Happy | Velocity after 3 sprints | planned vs done | API | P1 | TC-SPRINT-05 | Covered |
| ECC-SPR-006 | Happy | Planning Poker vote→reveal | scores visible | Manual/E2E | P2 | TC-SPRINT-06 | TODO |
| ECC-SPR-007 | Neg | End date before start | 400 | API | P1 | TC-SPRINT-01 | Covered |
| ECC-SPR-008 | Neg | Start when another active | 400 | API | P0 | TC-SPRINT-03 | Covered |
| ECC-SPR-009 | Neg | Complete non-active | 400 | API | P1 | — | Partial |
| ECC-SPR-010 | Neg | Non-member sprint ops | 403 | API | P0 | — | TODO |
| ECC-SPR-011 | Edge | Sprint with 0 issues start | allowed or warn | API | P2 | — | TODO |
| ECC-SPR-012 | Edge | Same-day start=end | 400 or 1-day OK | API | P2 | — | TODO |
| ECC-SPR-013 | Edge | WIP during sprint board | enforce | API | P1 | — | Partial |
| ECC-SPR-014 | Corner | Two users start two sprints race | one wins, one 400 | API | P0 | — | TODO |
| ECC-SPR-015 | Corner | Complete while issue mid-transition | consistent state | API | P1 | — | TODO |
| ECC-SPR-016 | Corner | Delete/archive project with active sprint | block or cascade | API | P1 | — | TODO |
| ECC-SPR-017 | Corner | Retro after complete | data available | Manual | P2 | US-S01 | TODO |
| ECC-SPR-018 | Edge | Burndown timezone day boundary UTC | points align UTC days | API | P1 | NFR-12 | TODO |
| ECC-SPR-019 | Happy | Add/remove issues in planning | backlog updated | API | P1 | — | Partial |
| ECC-SPR-020 | Neg | Modify completed sprint dates | 400 | API | P2 | — | TODO |
| ECC-SPR-021 | Corner | Calendar shows sprint as event | aggregator includes | API | P1 | US-V03 | Partial |

**Домен sprint: 21 кейс**

---

### 5. Messenger

> US-M04, US-S12 · NFR-07/15 · TC-MSG-* · Mongo messages · WS events

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MSG-001 | Happy | Create public channel | 201 | API/E2E | P0 | TC-MSG-01 | Covered |
| ECC-MSG-002 | Happy | DM dedup A↔B | same channel | API | P0 | TC-MSG-02 | Covered |
| ECC-MSG-003 | Happy | Send message → NATS/WS | 201 + event | API | P0 | TC-MSG-03 | Covered |
| ECC-MSG-004 | Happy | Reaction toggle add/remove | toggled | API | P1 | TC-MSG-04 | Covered |
| ECC-MSG-005 | Happy | Thread reply increments count | reply_count++ | API | P1 | TC-MSG-06 | Covered |
| ECC-MSG-006 | Happy | Unread: skip sender; mark read | counts correct | API | P0 | TC-MSG-09 | Covered |
| ECC-MSG-007 | Happy | Search messages | hits | API | P1 | TC-MSG-10 | Covered |
| ECC-MSG-008 | Happy | Presence heartbeat online/away/offline | status | API | P1 | TC-PRESENCE-* | Covered |
| ECC-MSG-009 | Happy | Pin / forward / poll | 200 | API/Manual | P2 | US-S12 | Partial |
| ECC-MSG-010 | Neg | Send in channel not member | 403 | API | P0 | TC-MSG-03 | Covered |
| ECC-MSG-011 | Neg | Read-only channel non-admin | 403 | API | P0 | TC-MSG-07 | Covered |
| ECC-MSG-012 | Neg | Empty content / oversize text | 400 | API | P1 | — | Partial |
| ECC-MSG-013 | Neg | Private channel invite leak | non-member 404 | API | P0 | — | Partial |
| ECC-MSG-014 | Neg | React to deleted message | 404 | API | P1 | — | TODO |
| ECC-MSG-015 | Edge | Message rate 61/min | 429 | API | P0 | NFR-07 | Partial |
| ECC-MSG-016 | Edge | Cursor pagination empty / last page | empty next | API | P1 | NFR-15 | Partial |
| ECC-MSG-017 | Edge | Channel 1000+ messages load | cursor works <300ms | Perf | P2 | — | TODO |
| ECC-MSG-018 | Edge | Mute until past/future | muted_until | API | P2 | TC-MSG-08 | TODO |
| ECC-MSG-019 | Edge | Typing indicator flood | throttle | Manual | P2 | — | TODO |
| ECC-MSG-020 | Corner | Double-send same client nonce | one message | API | P0 | — | TODO |
| ECC-MSG-021 | Corner | Delete parent with thread replies | count/cascade policy | API | P1 | TC-MSG-06 | Partial |
| ECC-MSG-022 | Corner | WS disconnect mid-send → retry | no dup or recoverable | E2E | P0 | — | TODO |
| ECC-MSG-023 | Corner | Soft-delete channel then restore | messages intact | API | P1 | NFR-11 | TODO |
| ECC-MSG-024 | Corner | Auto-channel on project create | #project-KEY exists | API | P1 | US-M02 | Covered |
| ECC-MSG-025 | Corner | Read-only #hub-ci admin override | admin can post | API | P1 | TC-MSG-07 | Covered |
| ECC-MSG-026 | Happy | Join public channel | member | API | P1 | — | Partial |
| ECC-MSG-027 | Neg | Edit/delete others' message (non-admin) | 403 | API | P1 | — | TODO |
| ECC-MSG-028 | Edge | Unicode / emoji-only / RTL content | stored OK | API | P2 | — | TODO |
| ECC-MSG-029 | Corner | Presence cleanup stale | offline | Unit | P1 | TC-PRESENCE-03 | Covered |
| ECC-MSG-030 | Corner | Schedule call from messenger (future RSVP) | invite flow | Future | P0 | SEED-002 | Future |
| ECC-MSG-031 | Edge | Attachment via storage presign in chat | file linked | API | P1 | storage | Partial |
| ECC-MSG-032 | Neg | Search without membership scope | no foreign hits | API | P0 | NF-03 | Partial |
| ECC-MSG-033 | Corner | Forward from private to public | redact or block | API | P1 | — | TODO |
| ECC-MSG-034 | Happy | Bookmark message | saved | Manual | P2 | US-S12 | TODO |
| ECC-MSG-035 | Edge | Bulk unread get all channels | map | API | P1 | — | Covered |
| ECC-MSG-036 | Corner | Member removed while viewing channel | 403 on next send | API | P1 | — | TODO |
| ECC-MSG-037 | Neg | Invalid UUID channel id | 400 | API | P2 | NFR-25 | Partial |
| ECC-MSG-038 | Corner | NATS down: send message | graceful error / outbox | API | P0 | NFR-26 | TODO |
| ECC-MSG-039 | Happy | Voice message upload (if enabled) | 201 | Manual | P2 | US-S12 | TODO |
| ECC-MSG-040 | Edge | Link preview SSRF-safe fetch | ValidateOutboundURL | API | P0 | NFR-21 | TODO |

**Домен messenger: 40 кейсов**

---

### 6. Calls + LiveKit

> US-M06, US-S14 · NFR-03/20 · TC-CALL-* · **LiveKit only** (SFU migration REJECTED)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-001 | Happy | Create call from channel → LiveKit JWT | 201 room+token | API | P0 | TC-CALL-01 | Covered |
| ECC-CALL-002 | Happy | Instant call + invite code | 201 | API | P0 | TC-CALL-02 | Covered |
| ECC-CALL-003 | Happy | Guest join by invite | guest token | API | P0 | TC-CALL-04 | Covered |
| ECC-CALL-004 | Happy | Leave / end call | counts / ended | API | P0 | TC-CALL-05/06 | Partial |
| ECC-CALL-005 | Happy | Start/stop recording egress | recording states | API | P0 | TC-CALL-07/08 | Covered |
| ECC-CALL-006 | Happy | Ephemeral call chat | message OK | API | P1 | TC-CALL-09 | Covered |
| ECC-CALL-007 | Happy | Scheduled call + RRULE | created | API | P0 | US-S14 | Covered |
| ECC-CALL-008 | Happy | Conflict check overlapping | conflicts listed | API | P1 | — | Covered |
| ECC-CALL-009 | Happy | Webhook participant_joined valid sig | count++ | API | P0 | TC-CALL-11, NFR-20 | Covered |
| ECC-CALL-010 | Neg | Webhook invalid / missing sha256 | 401 reject | API | P0 | SEC-02 | Covered |
| ECC-CALL-011 | Neg | Webhook tampered body | 401 | API | P0 | — | Covered |
| ECC-CALL-012 | Neg | Join ended call | 400/404 | API | P1 | — | Partial |
| ECC-CALL-013 | Neg | Non-member start recording | 403 | API | P0 | — | Covered |
| ECC-CALL-014 | Neg | Guest join invalid code | 404 | API | P1 | — | Partial |
| ECC-CALL-015 | Neg | Chat text too long | 400 | API | P2 | — | Covered |
| ECC-CALL-016 | Neg | Invalid RRULE on schedule | 400 | API | P1 | — | Covered |
| ECC-CALL-017 | Edge | Max 50 participants | 51st rejected/warn | Manual | P0 | NFR-03 | TODO |
| ECC-CALL-018 | Edge | Max active recordings limit | reject over | API | P1 | — | Covered |
| ECC-CALL-019 | Edge | Recording download not ready | 400/409 | API | P1 | — | Covered |
| ECC-CALL-020 | Edge | Screen share only (no webcam Phase16) | audio+screen OK | Manual | P1 | US-M06 | Manual |
| ECC-CALL-021 | Edge | Process due scheduled calls worker | room created | API | P1 | — | Covered |
| ECC-CALL-022 | Corner | **LiveKit SFU restart mid-call** | clients reconnect; state recover | Manual | P0 | LiveKit ops | Manual |
| ECC-CALL-023 | Corner | Webhook room_finished after local end | idempotent | API | P0 | — | Partial |
| ECC-CALL-024 | Corner | Double end call | second idempotent | API | P1 | — | TODO |
| ECC-CALL-025 | Corner | Token expiry mid-call | refresh/rejoin | Manual | P0 | — | Manual |
| ECC-CALL-026 | Corner | Egress fail mid-recording | status failed; UI | API/Manual | P1 | — | Partial |
| ECC-CALL-027 | Corner | Delete scheduled while participants joining | consistent | API | P1 | — | TODO |
| ECC-CALL-028 | Corner | Raise hand + leave race | no stuck hand | Manual | P2 | TC-CALL-10 | TODO |
| ECC-CALL-029 | Happy | Recording list pagination | pages | API | P2 | — | Covered |
| ECC-CALL-030 | Neg | Instant call recording unsupported | 400 | API | P1 | — | Covered |
| ECC-CALL-031 | Edge | Empty participant_ids conflict check | 400 | API | P2 | — | Covered |
| ECC-CALL-032 | Corner | Clock skew scheduled starts_at | worker still fires | API | P1 | — | TODO |
| ECC-CALL-033 | Neg | Webhook wrong issuer | 401 | API | P0 | — | Covered |
| ECC-CALL-034 | Happy | Update scheduled call time | 200 | API | P1 | — | Covered |
| ECC-CALL-035 | Corner | Channel deleted while call active | call ends or orphan policy | Manual | P1 | — | TODO |
| ECC-CALL-036 | Edge | Guest name empty / XSS name | sanitize 400 | API | P1 | — | Partial |
| ECC-CALL-037 | Corner | mediasoup migration paths | **N/A — REJECTED**; no tests required | — | P2 | SEED-001 | N/A |
| ECC-CALL-038 | Happy | Project recordings RBAC member | 200 | API | P1 | — | Covered |
| ECC-CALL-039 | Corner | Network partition one peer | others continue | Manual | P1 | WebRTC | Manual |
| ECC-CALL-040 | Edge | Concurrent guest joins same code | all get tokens | API | P1 | — | TODO |

**Домен calls: 40 кейсов** (ECC-CALL-037 = N/A, не в счётчике активных → **39 активных + 1 N/A**)

---

### 7. Calendar

> US-V03, US-CAL-* · as-is aggregator · **будущий RSVP** SEED-002 / JULY brief §3 · C1–C15

#### 7.1 As-is (read aggregator)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-001 | Happy | List events: calls+sprints+deadlines+leave | unified list | API | P0 | US-V03 | Covered |
| ECC-CAL-002 | Happy | Export iCal | text/calendar | API | P1 | US-CAL-03 | Covered |
| ECC-CAL-003 | Happy | Expand recurring RRULE calls | instances | Unit | P1 | — | Covered |
| ECC-CAL-004 | Neg | Invalid / missing range params | 400 | API | P1 | — | Covered |
| ECC-CAL-005 | Neg | Range too wide | 400 | API | P1 | — | Covered |
| ECC-CAL-006 | Edge | Empty range (no events) | 200 [] | API | P2 | — | TODO |
| ECC-CAL-007 | Edge | Leave pending not busy | not in busy | API | P1 | US-CAL-04 | Partial |
| ECC-CAL-008 | Edge | UTC store / local display | FE converts | E2E/Manual | P0 | NFR-12, C6 | Partial |
| ECC-CAL-009 | Corner | Overlapping sprint + call + leave same day | all shown | API | P1 | — | Partial |
| ECC-CAL-010 | Happy | Team availability busy/free (when shipped) | slots | API | P1 | US-CAL-04 | Future |
| ECC-CAL-011 | Neg | External iCal SSRF URL | blocked ValidateOutboundURL | API | P0 | NFR-21, US-CAL-03 | Future |
| ECC-CAL-012 | Edge | Resource double-book | 409 | API | P1 | US-CAL-01 | Future |
| ECC-CAL-013 | Edge | Monthly RRULE 31st in short month | last day | Unit | P1 | US-CAL-02 | Future |
| ECC-CAL-014 | Corner | Soft-delete series restore | all instances back | API | P1 | NFR-11 | Future |

#### 7.2 CAL-SYNC RSVP (brief §3.4 C1–C15 + расширения)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C01 | Corner | Accept после delete события организатором | 404/gone; inbox clear | API | P0 | Brief C1 | Future |
| ECC-CAL-C02 | Edge | Recurring: accept серии vs instance | v1: **серия целиком** | API | P0 | Brief C2 | Future |
| ECC-CAL-C03 | Edge | Два инвайта на одно время | conflict warning; soft | API | P0 | Brief C3 | Future |
| ECC-CAL-C04 | Neg | Участник не в project/channel | cannot add / guest-only | API | P0 | Brief C4 | Future |
| ECC-CAL-C05 | Neg | Organizer declines сам | запрет; только cancel | API | P0 | Brief C5 | Future |
| ECC-CAL-C06 | Edge | TZ UTC DB + ICS VTIMEZONE | correct local | API/E2E | P0 | Brief C6, NFR-12 | Future |
| ECC-CAL-C07 | Corner | Partial update title mid-flight | notif; rsvp **not** reset | API | P1 | Brief C7 | Future |
| ECC-CAL-C08 | Corner | Reschedule: reset RSVP **or** keep+banner | per product decision | API | P0 | Brief C8 | Future |
| ECC-CAL-C09 | Corner | Soft-delete / GDPR event | trash 30d; hard admin | API | P0 | Brief C9 | Future |
| ECC-CAL-C10 | Corner | Race simultaneous accept/decline | optimistic version attendee | API | P0 | Brief C10 | Future |
| ECC-CAL-C11 | Edge | Channel call без явных attendees | all members required | API | P1 | Brief C11 | Future |
| ECC-CAL-C12 | Edge | Leave/busy overlap availability | leave+accepted calls | API | P1 | Brief C12, US-CAL-04 | Future |
| ECC-CAL-C13 | Corner | Email bounce / no mailbox | in-app invite SoT | Manual | P1 | Brief C13 | Future |
| ECC-CAL-C14 | Neg | External organizer email-only | out of scope v1 | API | P2 | Brief C14 | Future |
| ECC-CAL-C15 | Corner | Messenger thread delete ≠ cancel call | cancel only via event API | API | P0 | Brief C15 | Future |
| ECC-CAL-C16 | Happy | Create from messenger → attendees + invite | NATS calendar.invite.created | API/E2E | P0 | SEED-002 | Future |
| ECC-CAL-C17 | Happy | Accept → event in both calendars | grid shows | E2E | P0 | SEED-002 | Future |
| ECC-CAL-C18 | Happy | Decline → hidden from grid; organizer notif | declined | API | P0 | SEED-002 | Future |
| ECC-CAL-C19 | Happy | Tentative → soft-busy | shown tentative | API | P1 | SEED-002 | Future |
| ECC-CAL-C20 | Edge | needs_action inbox badge not hard-busy | policy | API/E2E | P1 | Brief §3.2 | Future |
| ECC-CAL-C21 | Happy | L0 ICS attachment in email | hub-mail sends ICS | Manual | P1 | Brief L0 | Future |
| ECC-CAL-C22 | Neg | L3 OAuth without ADR | blocked / not shipped | — | P1 | GZ-V4-05 | Future |
| ECC-CAL-C23 | Corner | Double Accept idempotent | same state | API | P1 | — | Future |
| ECC-CAL-C24 | Corner | Accept then organizer cancel | gone + notif | API | P0 | — | Future |
| ECC-CAL-C25 | Edge | Attendee removed from channel after invite | policy revoke/keep | API | P1 | — | Future |

**Домен calendar: 39 кейсов** (14 as-is + 25 RSVP/sync)

---

### 8. TMS

> US-M05, US-S13, US-C02 · TC-TMS-* · Allure · Mongo cases/results

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-001 | Happy | Create classic case | 201 KEY-TC-N | API/E2E | P0 | TC-TMS-01 | Covered |
| ECC-TMS-002 | Happy | Create BDD case | 201 | API | P0 | TC-TMS-02 | Covered |
| ECC-TMS-003 | Happy | Update → version; rollback | version N; restore | API | P0 | TC-TMS-03/04 | Covered |
| ECC-TMS-004 | Happy | Run + add cases + execute result | stats++ | API | P0 | TC-TMS-05/06 | Covered |
| ECC-TMS-005 | Happy | Allure import match ASID/annotation | matched | API | P0 | TC-TMS-07 | Covered |
| ECC-TMS-006 | Happy | Bug from failed result | Bug issue | API | P1 | TC-TMS-09 | Covered |
| ECC-TMS-007 | Happy | Traceability / flaky reports | 200 | API | P1 | TC-TMS-10/11 | Covered |
| ECC-TMS-008 | Neg | Rollback to missing version | 404 | API | P1 | — | Partial |
| ECC-TMS-009 | Neg | Execute on completed run | 400 | API | P1 | — | TODO |
| ECC-TMS-010 | Neg | Non-member TMS write | 403 | API | P0 | — | Partial |
| ECC-TMS-011 | Neg | Malformed Allure JSON | 400 | API | P1 | TC-ALLURE-02 | Covered |
| ECC-TMS-012 | Neg | JUnit webhook bad signature/body | 401/400 | API | P0 | TC-TMS-08 | TODO |
| ECC-TMS-013 | Edge | Empty steps classic | 400 or allow | API | P2 | — | TODO |
| ECC-TMS-014 | Edge | Import 1000 Allure results | <5s | Perf | P2 | perf | TODO |
| ECC-TMS-015 | Edge | Case key autoincrement race | unique keys | API | P0 | — | TODO |
| ECC-TMS-016 | Corner | Concurrent edit same case version | 409 | API | P0 | NFR-24 | TODO |
| ECC-TMS-017 | Corner | Match priority order Allure strategies | deterministic | Unit | P1 | TC-ALLURE-05 | Covered |
| ECC-TMS-018 | Corner | Soft-delete case in active run | block or skip | API | P1 | — | TODO |
| ECC-TMS-019 | Corner | Export PDF/CSV on huge run | streams OK | API | P2 | TC-TMS-12 | Partial |
| ECC-TMS-020 | Happy | Suites hierarchy | CRUD | API | P1 | — | Partial |
| ECC-TMS-021 | Neg | Invalid BDD keyword | 400 | API | P2 | — | TODO |
| ECC-TMS-022 | Edge | Environments matrix | filter results | API | P2 | US-S13 | Partial |
| ECC-TMS-023 | Corner | TestRail import partial fail | report errors | API | P1 | — | Partial |
| ECC-TMS-024 | Happy | Link case↔issue | bidirectional | API | P1 | — | Covered |
| ECC-TMS-025 | Edge | Flaky detection boundary (1 fail/1 pass) | listed | API | P2 | US-C02 | TODO |
| ECC-TMS-026 | Corner | Webhook replay same payload | idempotent | API | P0 | — | TODO |
| ECC-TMS-027 | Neg | Result status unknown enum | 400 | API | P1 | — | Partial |
| ECC-TMS-028 | Corner | Mongo down during import | graceful 503 | API | P0 | — | TODO |
| ECC-TMS-029 | Happy | Compare runs (if UI) | diff | Manual | P2 | US-C02 | TODO |
| ECC-TMS-030 | Edge | Pagination runs offset | pages | API | P2 | NFR-14 | Partial |

**Домен tms: 30 кейсов**

---

### 9. Notification / Search / WS

> US-S05, US-M07 · NFR-18/19 · NF-02/03/07 · TC-NOTIF-*, TC-SEARCH-*, TC-WS-*

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-NSW-001 | Happy | Assign → in-app notification | type=assigned | API | P0 | TC-NOTIF-01 | Partial |
| ECC-NSW-002 | Happy | @mention → notification | mentioned | API | P0 | TC-NOTIF-02 | Partial |
| ECC-NSW-003 | Happy | Mark read / read-all | is_read | API/E2E | P1 | TC-NOTIF-03/04 | Partial |
| ECC-NSW-004 | Happy | Quiet hours defer non-urgent | delayed | API | P1 | TC-NOTIF-05 | Covered |
| ECC-NSW-005 | Happy | Urgent breaks quiet hours | delivered | API | P1 | TC-NOTIF-06 | TODO |
| ECC-NSW-006 | Happy | Telegram notify delivery (bound) | sent | Manual | P1 | US-S05 | Manual |
| ECC-NSW-007 | Happy | Web Push subscribe | 200 | API | P1 | TC-NOTIF-07 | TODO |
| ECC-NSW-008 | Neg | Web Push endpoint SSRF | ValidateOutboundURL reject | API | P0 | NF-07, NFR-21 | TODO |
| ECC-NSW-009 | Neg | Mark read others' notification | 403/404 | API | P0 | — | TODO |
| ECC-NSW-010 | Edge | Debounce burst same subject | coalesced | Unit | P1 | — | Covered |
| ECC-NSW-011 | Edge | Empty notification list | 200 [] | API | P2 | — | TODO |
| ECC-NSW-012 | Corner | Notif during user deactivate | no delivery / drop | API | P1 | — | TODO |
| ECC-NSW-013 | Happy | Global search Ctrl+K multi-type | ranked results | E2E | P0 | US-M07 | Partial |
| ECC-NSW-014 | Happy | JQL via search | filtered | API | P1 | TC-JQL-* | Covered |
| ECC-NSW-015 | Neg | Search without auth | 401 | API | P0 | — | Partial |
| ECC-NSW-016 | Neg | Search leaks other project | **no hits** user-scope | API | P0 | NF-03, NFR-19 | Partial |
| ECC-NSW-017 | Edge | Search limit / empty q | defaults | API | P1 | TC-SEARCH-01/04 | Covered |
| ECC-NSW-018 | Edge | Saved filter CRUD | DTO OK | API | P2 | TC-SEARCH-03 | Covered |
| ECC-NSW-019 | Corner | Search soft-deleted entities | excluded | API | P1 | — | TODO |
| ECC-NSW-020 | Happy | WS connect with JWT | connected | API | P0 | — | Covered |
| ECC-NSW-021 | Happy | Subscribe room with membership | events flow | API | P0 | NFR-18 | Partial |
| ECC-NSW-022 | Neg | Subscribe room **without** membership | reject (NF-02) | API | P0 | NFR-18 | Partial |
| ECC-NSW-023 | Neg | WS with expired/invalid token | close 401 | API | P0 | — | Covered |
| ECC-NSW-024 | Edge | Ping/pong protocol | alive | Unit | P1 | TC-WS-03 | Covered |
| ECC-NSW-025 | Edge | Hub shutdown drains clients | clean close | Unit | P1 | — | Covered |
| ECC-NSW-026 | Corner | Origin check fail | reject | Unit | P0 | — | Covered |
| ECC-NSW-027 | Corner | NATS bridge restart | resubscribe / no crash | API | P0 | — | Partial |
| ECC-NSW-028 | Corner | 100+ concurrent WS | stable | Perf | P1 | NFR-02 | TODO |
| ECC-NSW-029 | Edge | Message rate via WS path | still limited | API | P1 | NFR-07 | TODO |
| ECC-NSW-030 | Happy | issue_created pushed to project room | event received | E2E | P0 | regression WS | TODO |
| ECC-NSW-031 | Neg | Cross-user notification IDOR | 404 | API | P0 | — | TODO |
| ECC-NSW-032 | Corner | Quiet hours TZ boundary | correct local window | API | P1 | NFR-12 | TODO |
| ECC-NSW-033 | Edge | Search unicode / special chars | safe query | API | P2 | — | TODO |
| ECC-NSW-034 | Happy | Markdown escape Telegram format | no inject | Unit | P1 | notif tests | Covered |
| ECC-NSW-035 | Corner | Push unsubscribe + send | no error storm | API | P2 | — | TODO |
| ECC-NSW-036 | Neg | WS subscribe project:{other} | deny | API | P0 | NF-02 | Partial |
| ECC-NSW-037 | Edge | Authorizer membership cache stale | recheck | Unit | P0 | — | Partial |
| ECC-NSW-038 | Corner | Search + RBAC role Viewer only | scoped | API | P0 | — | TODO |

**Домен notification/search/ws: 38 кейсов**

---

### 10. Storage / Files

> NFR-04/08/27 · NF-09 known · TC-STORAGE-* · SeaweedFS presigned

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-STOR-001 | Happy | Presigned upload URL | URL + TTL | API | P0 | TC-STORAGE-01 | Covered |
| ECC-STOR-002 | Happy | Presigned download URL | URL + TTL | API | P0 | TC-STORAGE-02 | Covered |
| ECC-STOR-003 | Happy | Upload result fields | metadata | Unit | P1 | TC-STORAGE-03 | Covered |
| ECC-STOR-004 | Neg | Unauthenticated presign | 401 | API | P0 | — | Covered |
| ECC-STOR-005 | Neg | Non-member upload to project | 403 | API | P0 | — | Partial |
| ECC-STOR-006 | Neg | Path traversal filename | sanitized reject | API | P0 | filename tests | Covered |
| ECC-STOR-007 | Edge | Fiber BodyLimit 25 MB | 413 over | API | P0 | NFR-04 | Partial |
| ECC-STOR-008 | Edge | File rate 11/min | 429 | API | P0 | NFR-08 | TODO |
| ECC-STOR-009 | Edge | Presign TTL expiry | 403 from S3 | Manual | P1 | — | TODO |
| ECC-STOR-010 | Corner | **NF-09:** presigned PUT without size cap | document risk; monitor | Manual | P0 | NF-09 | Manual |
| ECC-STOR-011 | Corner | Orphan cleanup job | deletes orphans | Unit | P1 | orphan tests | Covered |
| ECC-STOR-012 | Corner | Soft-delete parent entity → file | retention policy | API | P1 | — | TODO |
| ECC-STOR-013 | Edge | Empty file 0 bytes | allow or 400 | API | P2 | — | TODO |
| ECC-STOR-014 | Edge | Content-Type mismatch | still store or reject | API | P2 | — | TODO |
| ECC-STOR-015 | Neg | Extremely long filename | 400 | API | P2 | — | Partial |
| ECC-STOR-016 | Corner | Concurrent uploads same key | no corrupt | API | P1 | — | TODO |
| ECC-STOR-017 | Happy | Attach file to issue/message | linked | API | P1 | — | Partial |
| ECC-STOR-018 | Neg | Download after GDPR purge | 404 | API | P0 | US-S03 | TODO |
| ECC-STOR-019 | Edge | SeaweedFS down | 503 graceful | API | P0 | — | TODO |
| ECC-STOR-020 | Corner | Double-complete upload callback | idempotent | API | P1 | — | TODO |

**Домен storage: 20 кейсов**

---

### 11. Dashboard

> US-S02, US-S11 · TC-DASH-*

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-DASH-001 | Happy | Personal layout save/load | persisted | API/E2E | P1 | TC-DASH-01 | Covered |
| ECC-DASH-002 | Happy | Project dashboard widgets | stats | API | P1 | TC-DASH-02 | Covered |
| ECC-DASH-003 | Happy | Admin stats health | services | API | P1 | TC-DASH-03 | Covered |
| ECC-DASH-004 | Happy | Drag-drop resize widgets | layout OK | E2E | P2 | US-S02 | Manual |
| ECC-DASH-005 | Neg | Save invalid layout JSON | 400 | API | P1 | TC-DASH-04 | Covered |
| ECC-DASH-006 | Neg | Non-admin admin dashboard | 403 | API | P0 | — | Partial |
| ECC-DASH-007 | Neg | Project dash non-member | 403 | API | P0 | — | TODO |
| ECC-DASH-008 | Edge | Empty widgets layout | defaults | API | P2 | — | Covered |
| ECC-DASH-009 | Edge | Max widgets limit | 400 over | API | P2 | — | TODO |
| ECC-DASH-010 | Corner | Concurrent layout saves | last-write/version | API | P1 | NFR-24 | TODO |
| ECC-DASH-011 | Corner | Widget data source deleted | empty/error state | E2E | P1 | — | TODO |
| ECC-DASH-012 | Happy | Templates Dev/QA/Manager | apply | API | P2 | US-S02 | TODO |
| ECC-DASH-013 | Edge | Burndown widget no active sprint | empty state | E2E | P2 | — | TODO |
| ECC-DASH-014 | Corner | TZ in date widgets | local display | Manual | P1 | NFR-12 | TODO |
| ECC-DASH-015 | Neg | XSS in widget title | escaped | E2E | P0 | — | TODO |
| ECC-DASH-016 | Happy | Mobile dashboard simplified | usable | Manual | P2 | US-C05 | Manual |

**Домен dashboard: 16 кейсов**

---

### 12. v4: Intranet / CRM / HR / Accounting / Bases

> US-V01..V06, US-INTR-*, US-CAL-* (partial) · E2E `v4-superapp.spec.ts`

#### 12.1 Intranet

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4I-001 | Happy | News feed list/create post | 200/201 | API/E2E | P1 | US-V01 | Partial |
| ECC-V4I-002 | Happy | Directory + org chart | 200 | API | P1 | US-V02 | Partial |
| ECC-V4I-003 | Happy | Kudos send | 201 | API | P2 | US-INTR-04 | Partial |
| ECC-V4I-004 | Happy | Random Coffee match + DM | pair+DM | API | P2 | US-INTR-05 | Partial |
| ECC-V4I-005 | Neg | Non-auth news write | 401 | API | P0 | — | Partial |
| ECC-V4I-006 | Neg | Admin delete post (moderation) | soft-deleted | API | P1 | US-INTR-03 | Future |
| ECC-V4I-007 | Edge | Poll on post; vote once | aggregates | API | P1 | US-INTR-01 | Future |
| ECC-V4I-008 | Edge | Edit poll options after vote | 422 | API | P1 | US-INTR-01 | Future |
| ECC-V4I-009 | Corner | Delete post with poll → cascade trash | hidden 30d | API | P1 | US-INTR-01 | Future |
| ECC-V4I-010 | Corner | Restore post restores poll counts | restored | API | P1 | US-INTR-03 | Future |
| ECC-V4I-011 | Happy | Onboarding checklist template | progress | API | P2 | US-INTR-02 | Future |
| ECC-V4I-012 | Edge | Kudos leaderboard empty period | empty | API | P2 | US-INTR-04 | Future |
| ECC-V4I-013 | Corner | Coffee opt-out mid-week | no match / skip | API | P2 | — | Partial |

#### 12.2 CRM

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4C-001 | Happy | Contact + deal CRUD | 201/200 | API | P1 | US-V04 | Covered |
| ECC-V4C-002 | Neg | Deal without contact policy | 400 or allow | API | P2 | — | Partial |
| ECC-V4C-003 | Neg | Non-member CRM access | 403 | API | P0 | — | Partial |
| ECC-V4C-004 | Edge | Empty pipeline | 200 [] | API | P2 | — | TODO |
| ECC-V4C-005 | Corner | Optimistic lock deal version | 409 | API | P1 | NFR-24 | TODO |
| ECC-V4C-006 | Corner | Soft-delete contact with deals | cascade policy | API | P1 | — | TODO |
| ECC-V4C-007 | Happy | Invoice link from CRM (if any) | 200 | API | P2 | — | Partial |
| ECC-V4C-008 | Edge | Search contacts unicode | hits | API | P2 | — | TODO |

#### 12.3 HR

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4H-001 | Happy | Leave request create | 201 pending | API | P0 | US-V05 | Covered |
| ECC-V4H-002 | Happy | Approve leave → calendar busy | approved event | API | P0 | US-CAL-04 | Partial |
| ECC-V4H-003 | Neg | Overlap leave same user | 409/400 | API | P1 | — | Partial |
| ECC-V4H-004 | Neg | Approve by non-manager | 403 | API | P0 | — | Partial |
| ECC-V4H-005 | Edge | Leave end < start | 400 | API | P1 | — | Covered |
| ECC-V4H-006 | Edge | Pending leave not in availability | free | API | P1 | US-CAL-04 | Partial |
| ECC-V4H-007 | Corner | Approve then cancel race | consistent | API | P1 | — | TODO |
| ECC-V4H-008 | Corner | Leave during scheduled calls | conflict warn | API | P1 | CAL-SYNC C12 | Future |
| ECC-V4H-009 | Happy | Reject leave | rejected | API | P1 | — | Partial |

#### 12.4 Accounting

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4A-001 | Happy | Journal entry create | 201 balanced | API | P1 | US-V06 | Partial |
| ECC-V4A-002 | Happy | Invoice CRUD | 201 | API | P1 | — | Partial |
| ECC-V4A-003 | Neg | Unbalanced entry | 400 | API | P0 | — | Partial |
| ECC-V4A-004 | Neg | Non-admin accounting write | 403 | API | P0 | — | Partial |
| ECC-V4A-005 | Edge | Zero-amount line | 400 | API | P2 | — | TODO |
| ECC-V4A-006 | Corner | Soft-delete invoice with payments | policy | API | P1 | — | TODO |
| ECC-V4A-007 | Corner | Concurrent post same period | lock/version | API | P1 | — | TODO |
| ECC-V4A-008 | Edge | Currency / decimal precision | exact | Unit | P1 | — | TODO |

#### 12.5 Bases (no-code DB)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4B-001 | Happy | Create base + table + row | 201 | API | P1 | vision bases | Covered |
| ECC-V4B-002 | Happy | RBAC on base | allow/deny | API | P0 | bases rbac tests | Covered |
| ECC-V4B-003 | Neg | Write without permission | 403 | API | P0 | — | Covered |
| ECC-V4B-004 | Neg | Invalid field type value | 400 | API | P1 | — | Partial |
| ECC-V4B-005 | Edge | Empty table list rows | [] | API | P2 | — | Partial |
| ECC-V4B-006 | Edge | Max columns / rows soft limit | 400 | API | P2 | — | TODO |
| ECC-V4B-007 | Corner | Schema change with existing rows | migrate/block | API | P1 | — | TODO |
| ECC-V4B-008 | Corner | version conflict row update | 409 | API | P0 | NFR-24 | Partial |
| ECC-V4B-009 | Corner | Soft-delete base restore | data back | API | P1 | NFR-11 | TODO |
| ECC-V4B-010 | Happy | Handler flow CRUD smoke | 200 | API | P1 | — | Covered |

**Домен v4: 48 кейсов** (13+8+9+8+10)

---

### 13. AI

> US-F01..F05 · content policy · quota · privacy consent · E2E ai-hub

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-001 | Happy | @hub-ai command in channel | bot reply | API | P1 | US-F01 | Partial |
| ECC-AI-002 | Happy | List models / select preference | 200 | API/E2E | P1 | US-F04 | Partial |
| ECC-AI-003 | Happy | Admin create provider | 201 | API | P1 | US-F02 | Partial |
| ECC-AI-004 | Happy | Usage tracking daily | counters | Unit | P1 | US-F05 | Covered |
| ECC-AI-005 | Neg | Jailbreak / blocked content policy | reject | Unit | P0 | — | Covered |
| ECC-AI-006 | Neg | Quota exceeded | 429/403 | Unit | P0 | US-F05 | Covered |
| ECC-AI-007 | Neg | Cloud provider without consent | blocked | Unit | P0 | privacy tests | Covered |
| ECC-AI-008 | Neg | Non-admin provider CRUD | 403 | API | P0 | — | Partial |
| ECC-AI-009 | Neg | Invalid provider URL | 400 | API | P1 | — | Covered |
| ECC-AI-010 | Edge | Empty prompt | 400 | API | P2 | — | Partial |
| ECC-AI-011 | Edge | Max prompt size | 400 | API | P1 | — | TODO |
| ECC-AI-012 | Corner | Provider timeout / 5xx | graceful error | API | P1 | — | TODO |
| ECC-AI-013 | Corner | Audit log without content leak | metadata only | Unit | P0 | US-F05 | Covered |
| ECC-AI-014 | Happy | Code review unified GH/GL | review text | Unit | P2 | — | Covered |
| ECC-AI-015 | Happy | Generate test cases prompt builder | prompt OK | Unit | P2 | US-F03 | Covered |
| ECC-AI-016 | Corner | Switch model mid-conversation | consistent | Manual | P2 | — | TODO |
| ECC-AI-017 | Edge | Rate limit AI endpoints | 429 | API | P1 | — | TODO |
| ECC-AI-018 | Neg | SSRF via custom provider base URL | blocked | API | P0 | NFR-21 | Partial |
| ECC-AI-019 | Corner | Ollama local vs cloud consent matrix | correct filter | Unit | P1 | — | Covered |
| ECC-AI-020 | Happy | AI settings page loads | E2E | P2 | ai-hub.spec | Partial |
| ECC-AI-021 | Corner | Bot in read-only channel | admin-only or deny | API | P1 | MSG | TODO |
| ECC-AI-022 | Edge | Concurrent chat requests same user | queue/limit | API | P2 | — | TODO |

**Домен ai: 22 кейса**

---

### 14. Platform cross-cutting

> NFR-11..15, NFR-24/25 · soft-delete · UTC · pagination · error shape · rate limits · TC platform

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-PLT-001 | Happy | Error body `{error, request_id}` | shape + UUID | API | P0 | NFR-13 | Partial |
| ECC-PLT-002 | Happy | RequestID middleware propagates | header+body | Unit | P1 | middleware | Covered |
| ECC-PLT-003 | Happy | Soft-delete → trash list → restore | deleted_at null | API | P0 | NFR-11, US-S03 | Partial |
| ECC-PLT-004 | Happy | Hard delete Admin/GDPR | gone | API | P0 | TC-ADMIN-03 | TODO |
| ECC-PLT-005 | Neg | Restore after 30d retention | 404 purged | API | P0 | NFR-11 | TODO |
| ECC-PLT-006 | Neg | Non-admin hard delete | 403 | API | P0 | — | TODO |
| ECC-PLT-007 | Edge | Soft-delete day 0 and day 29 | still restorable | API | P1 | — | TODO |
| ECC-PLT-008 | Edge | Soft-delete day 30+ purge job | hard removed | API | P0 | — | TODO |
| ECC-PLT-009 | Corner | Restore then re-delete | new retention window | API | P1 | — | TODO |
| ECC-PLT-010 | Happy | UTC timestamps in API | Zulu / offset 0 | API | P0 | NFR-12 | Partial |
| ECC-PLT-011 | Edge | FE TZ convert America/ / Europe/ | correct local | E2E | P0 | NFR-12 | Manual |
| ECC-PLT-012 | Corner | DST transition event display | no double/skip hour bug | Manual | P1 | — | Manual |
| ECC-PLT-013 | Happy | Issues offset/limit pagination | pages | API | P1 | NFR-14 | Partial |
| ECC-PLT-014 | Happy | Chat cursor pagination | next cursor | API | P1 | NFR-15 | Partial |
| ECC-PLT-015 | Edge | limit=0 / limit>max / negative offset | 400 or clamped | API | P1 | — | TODO |
| ECC-PLT-016 | Edge | cursor invalid / tampered | 400 | API | P1 | — | TODO |
| ECC-PLT-017 | Happy | API rate 100/sec under | 200 | API | P1 | NFR-06 | Partial |
| ECC-PLT-018 | Edge | API 101st /sec | 429 | API | P0 | NFR-06 | TODO |
| ECC-PLT-019 | Edge | Auth 3/10m (see auth) | 429 | API | P0 | NFR-05 | Covered |
| ECC-PLT-020 | Edge | Files 10/min / msgs 60/min | 429 | API | P0 | NFR-07/08 | TODO |
| ECC-PLT-021 | Neg | UUID string PK rejected | 400 | API | P1 | NFR-25 | Partial |
| ECC-PLT-022 | Happy | Optimistic version on Ent update | version++ | API | P0 | NFR-24 | Partial |
| ECC-PLT-023 | Corner | Stale version all modules sample | 409 | API | P0 | — | Partial |
| ECC-PLT-024 | Neg | panic in handler | 500 + request_id, no crash process | API | P0 | TZ no panic | Partial |
| ECC-PLT-025 | Corner | io.ReadAll without LimitReader absent | code audit | Manual | P0 | TZ forbid | Manual |
| ECC-PLT-026 | Corner | http.DefaultClient user URL absent | SafeHTTPClient | Manual | P0 | TZ forbid | Manual |
| ECC-PLT-027 | Happy | CORS / Security headers | present | Unit | P1 | NFR-22 | Covered |
| ECC-PLT-028 | Happy | Health / metrics endpoints | 200 | E2E | P1 | health.spec | Covered |
| ECC-PLT-029 | Edge | Large JSON body DoS | limited | API | P0 | — | Partial |
| ECC-PLT-030 | Corner | DB migration mid-deploy read | compatible | Manual | P1 | devops | Manual |
| ECC-PLT-031 | Neg | DEV_MODE in prod build | disabled paths | CI | P0 | TZ | Partial |
| ECC-PLT-032 | Happy | Audit log filters | 200 | API | P1 | TC-ADMIN-01 | Covered |
| ECC-PLT-033 | Corner | Outbox publish retry | eventually delivered | Unit | P1 | outbox tests | Covered |
| ECC-PLT-034 | Edge | NATS subject constants | correct stream | Unit | P2 | TC-NATS-01 | Covered |
| ECC-PLT-035 | Corner | Backup ZIP empty + non-empty | OK | Unit | P1 | TC-ADMIN-04 | Covered |
| ECC-PLT-036 | Neg | Webhook signature skip forbidden | always verify | API | P0 | SEC-02 | Covered |

**Домен platform: 36 кейсов**

---

## Top-50 P0 кросс-доменных рисков

Баги, которые «дорого» (security, data loss, authz, money-time, prod outage). Отсортированы по домену/критичности — использовать как regression must-have.

| # | Риск | Домены | Каталог ID | Почему дорого |
|---|------|--------|------------|---------------|
| 1 | Authz bypass: WS room subscribe без membership | ws, messenger | ECC-NSW-022 | NF-02 data leak realtime |
| 2 | Search без user-scope | search | ECC-NSW-016 | NF-03 cross-project leak |
| 3 | LiveKit webhook signature skip / empty sha256 | calls | ECC-CALL-010/011 | SEC-02 fake join/recording |
| 4 | Web Push / iCal URL SSRF | notif, calendar | ECC-NSW-008, ECC-CAL-011 | NFR-21 RCE/SSRF |
| 5 | Magic Link rate limit / brute-force bypass | auth | ECC-AUTH-016/017 | account takeover |
| 6 | Refresh token reuse after logout/revoke | auth | ECC-AUTH-003/022 | session hijack |
| 7 | Deactivated user still authenticated (cache) | auth, rbac | ECC-AUTH-014, ECC-RBAC-015 | access after fire |
| 8 | Invite-only registration bypass | auth, org | ECC-AUTH-013 | unauthorized join |
| 9 | IDOR cross-project issue/file | tracker, storage | ECC-RBAC-023, ECC-STOR-005 | data leak |
| 10 | Optimistic lock ignored → silent overwrite | tracker, tms, v4 | ECC-TRCK-011/024 | lost updates |
| 11 | Duplicate issue/case keys under race | tracker, tms | ECC-TRCK-022, ECC-TMS-015 | data integrity |
| 12 | Double-submit create (message/issue/payment) | msg, tracker, acct | ECC-MSG-020, ECC-TRCK-025 | dup entities |
| 13 | Soft-delete restore after purge / GDPR incomplete | platform | ECC-PLT-004/005 | compliance |
| 14 | Presigned upload unbounded (NF-09) | storage | ECC-STOR-010 | disk DoS cost |
| 15 | Message rate limit bypass via WS | messenger, ws | ECC-MSG-015, ECC-NSW-029 | spam/DoS |
| 16 | API global rate limit ineffective | platform | ECC-PLT-018 | DoS |
| 17 | LiveKit restart mid-call no recovery UX | calls | ECC-CALL-022 | meeting failure |
| 18 | Recording webhook spoof → fake ready URL | calls | ECC-CALL-010 | malware link |
| 19 | Guest invite code enumerable | calls | ECC-CALL-014/040 | meeting crash |
| 20 | NATS down → silent message loss | messenger | ECC-MSG-038 | lost chat |
| 21 | Calendar RSVP accept after delete (C1) | calendar | ECC-CAL-C01 | ghost events |
| 22 | RSVP race accept/decline (C10) | calendar | ECC-CAL-C10 | wrong attendance |
| 23 | Reschedule policy wrong → no-shows (C8) | calendar | ECC-CAL-C08 | meeting fail |
| 24 | Thread delete cancels call wrongly (C15) | calendar, msg | ECC-CAL-C15 | dropped meetings |
| 25 | Leave approved but not in busy/free | hr, calendar | ECC-V4H-002/006 | double-book |
| 26 | Organizer self-decline allowed (C5) | calendar | ECC-CAL-C05 | broken invariant |
| 27 | External attendee / multi-org leak (C14) | calendar | ECC-CAL-C14 | scope creep leak |
| 28 | Founder duplication / last owner remove | rbac | ECC-RBAC-013/009 | org lockout |
| 29 | Pending 7d purge not running | auth/org | ECC-AUTH-027 | stale PII |
| 30 | Quiet hours TZ wrong → missed alerts | notif | ECC-NSW-032 | ops miss |
| 31 | Urgent not breaking quiet hours | notif | ECC-NSW-005 | incident miss |
| 32 | AI content exfil in audit logs | ai | ECC-AI-013 | privacy |
| 33 | AI provider SSRF via base URL | ai | ECC-AI-018 | SSRF |
| 34 | AI quota bypass | ai | ECC-AI-006 | cost overrun |
| 35 | Consent bypass → cloud LLM | ai | ECC-AI-007 | compliance |
| 36 | Unbalanced accounting entry posted | accounting | ECC-V4A-003 | financial integrity |
| 37 | Bases RBAC bypass row read | bases | ECC-V4B-002/003 | data leak |
| 38 | Workflow status deleted with live issues | tracker | ECC-TRCK-039 | board corruption |
| 39 | Two active sprints race | sprint | ECC-SPR-014 | planning chaos |
| 40 | Read-only channel write bypass | messenger | ECC-MSG-011 | CI channel spam |
| 41 | Private channel search leak | messenger, search | ECC-MSG-032 | secret leak |
| 42 | Link preview SSRF | messenger | ECC-MSG-040 | SSRF |
| 43 | Constant-time compare fail on secrets | auth | ECC-AUTH-029 | timing attack |
| 44 | Error responses without request_id / stack leak | platform | ECC-PLT-001/024 | debug info leak |
| 45 | DEV_MODE in production | platform | ECC-PLT-031 | auth bypass |
| 46 | Webhook replay Allure/JUnit duplicates | tms | ECC-TMS-026 | bad metrics |
| 47 | Mongo down → partial write inconsistency | tms, msg | ECC-TMS-028 | split brain |
| 48 | Clock skew JWT / scheduled calls | auth, calls | ECC-AUTH-026, ECC-CALL-032 | flaky auth/meetings |
| 49 | Concurrent refresh token rotation | auth | ECC-AUTH-022 | random logouts / hijack |
| 50 | Soft-deleted entity still in search/WS events | platform, search | ECC-NSW-019 | ghost data |

---

## Сводка счётчиков

| # | Домен | Happy | Neg | Edge | Corner | **Всего** |
|---|-------|------:|----:|-----:|-------:|----------:|
| 1 | Auth | 7 | 9 | 8 | 10 | **34** |
| 2 | RBAC / Org | 6 | 8 | 4 | 9 | **27** |
| 3 | Tracker | 11 | 8 | 7 | 14 | **40** |
| 4 | Sprint | 7 | 4 | 4 | 6 | **21** |
| 5 | Messenger | 10 | 7 | 9 | 14 | **40** |
| 6 | Calls + LiveKit | 11 | 9 | 8 | 11* | **39** (+1 N/A) |
| 7 | Calendar (as-is + RSVP) | 9 | 5 | 12 | 13 | **39** |
| 8 | TMS | 9 | 7 | 6 | 8 | **30** |
| 9 | Notification / Search / WS | 12 | 8 | 9 | 9 | **38** |
| 10 | Storage / Files | 4 | 5 | 6 | 5 | **20** |
| 11 | Dashboard | 5 | 4 | 3 | 4 | **16** |
| 12 | v4 (all sub) | 14 | 12 | 11 | 11 | **48** |
| 13 | AI | 7 | 6 | 4 | 5 | **22** |
| 14 | Platform | 10 | 6 | 10 | 10 | **36** |
| | **ИТОГО (базовый каталог)** | **122** | **98** | **101** | **129** | **450** |

\* Calls: ECC-CALL-037 mediasoup = N/A (в активный итог не входит).

### Expansion wave 2026-07-10 (максимизация Edge + Corner)

> Полные таблицы — в `.planning/edge-expansions/` (не дублировать сюда тысячи строк).  
> ID: Edge `*100+` · Corner parent `*C200+` · Corner agents `*C300+`.

| Файл | Тип | Строк ECC |
|------|-----|----------:|
| `edge-expansions/01-AUTH-RBAC-PLATFORM.md` | Edge | 70 |
| `edge-expansions/02-TRACKER-SPRINT-TMS.md` | Edge | 97 |
| `edge-expansions/03-MESSENGER-WS-NOTIF-SEARCH.md` | Edge | 80 |
| `edge-expansions/04-CALLS-CALENDAR-STORAGE.md` | Edge | 84 |
| `edge-expansions/05-V4-AI-FUTURE.md` | Edge | 96 |
| `edge-expansions/00-CORNER-PARENT.md` | Corner | 130 |
| `edge-expansions/corners/C01-AUTH-RBAC-PLATFORM.md` | Corner | 99 |
| `edge-expansions/corners/C02-TRACKER-SPRINT-TMS.md` | Corner | 85 |
| `edge-expansions/corners/C03-MESSENGER-WS-NOTIF-SEARCH.md` | Corner | 94 |
| `edge-expansions/corners/C04-CALLS-CALENDAR-MAIL.md` | Corner | 85 |
| `edge-expansions/corners/C05-FUTURE-V4-SSO-MCP-MS-UX.md` | Corner | 95 |
| | **Edge expansions** | **427** |
| | **Corner expansions** | **588** |
| | **Базовый каталог** | **450** |
| | **ГРАНД-ИТОГ ECC (после wave1)** | **≈1465** |

### Expansion wave 2 (2026-07-10) — ещё +5 Edge +5 Corner

> Путь: `.planning/edge-expansions/wave2/` · ID Edge `*200+` · Corner `C400+`

| Файл | Тип | Строк |
|------|-----|------:|
| `wave2/edge/E01` … `E05` | Edge | **406** (80+94+77+69+86) |
| `wave2/corners/C01` … `C05` | Corner | **365** (72+69+80+71+73) |
| | **Wave2 сумма** | **771** |
| | **ГРАНД-ИТОГ после Wave2** | **≈2236** |

| Тип (оценка all waves) | |
|------------------------|------:|
| Edge | **~934** |
| Corner | **~1082** |
| Happy+Neg (base) | **~220** |

### Expansion waves 3–5 (2026-07-10) — ещё 3×(5 Edge + 5 Corner)

| Волна | Путь | ID Edge / Corner | Строк ECC |
|-------|------|------------------|----------:|
| Wave3 | `edge-expansions/wave3/` | `*300+` / `C500+` | **781** |
| Wave4 | `edge-expansions/wave4/` | `*400+` / `C600+` | **848** |
| Wave5 | `edge-expansions/wave5/` | `*500+` / `C700+` | **830** |
| | **W3–W5 сумма** | | **2459** |
| | **ГРАНД-ИТОГ ECC (base+W1+W2+W3–5)** | | **≈4696** |

| Тип (оценка после W5) | |
|------------------------|------:|
| Edge | **~2200+** |
| Corner | **~2300+** |
| Happy+Neg (base) | **~220** |

### Gap snapshot (оценка)

| Gap | ~доля | Комментарий |
|-----|------:|-------------|
| Covered / Partial | ~15–20% от гранд-итога | Сильные auth/calls webhook/tms; expansions почти все TODO/Future |
| TODO / Manual | ~55% | QH-P2 приоритет: Top-50 + C300 P0 |
| Future | ~25% | CAL-SYNC, Wiki, SSO, MCP, MS, Redesign |

### Рекомендованный порядок автоматизации (после formal TC)

1. P0 из Top-50 без Covered (WS NF-02, search NF-03, rate limits, version races)
2. Corner C300 P0 security (auth refresh reuse, RBAC mid-flight, WS membership)
3. CAL-SYNC C1–C15 + CAL-C300 сразу после ba/sa ERD
4. Soft-delete restore/purge + GDPR cross-domain
5. Messenger double-send + NATS failure + LiveKit webhook replay
6. Future: SSO break-glass, MCP autonomy, MS route-flip — до кода как foresight TC

---

---

## История изменений

| Дата | Автор | Изменение |
|------|-------|-----------|
| 2026-07-10 | qa-manual | Первичный каталог **450** ECC (+1 N/A); CAL-SYNC C1–C15; LiveKit-only; Top-50 P0 |
| 2026-07-10 | 5×qa-manual Edge | Expansions `01–05`: **+427 Edge** (`*100+`) |
| 2026-07-10 | parent + 5×qa-manual Corner | `00-CORNER-PARENT` **+130** + corners C01–C05 **+458** (`C200+`/`C300+`) |
| 2026-07-10 | parent merge | Гранд-итог **≈1465** ECC; индекс expansions в сводке |
| 2026-07-10 | 10×qa-manual Wave2 | `wave2/edge` **+406** · `wave2/corners` **+365** · гранд **≈2236** |
| 2026-07-10 | 30×qa-manual Wave3–5 | W3 **+781** · W4 **+848** · W5 **+830** · гранд **≈4696** |

---

*Канон TC: `docs/TESTING.md`. Полные expansion-таблицы: `.planning/edge-expansions/`. Этот файл — индекс + базовый каталог + Top-50.*
