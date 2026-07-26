# W7 — Docs drift (API / QUICKSTART / vision ↔ код)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** docs-writer (Grok 4.5), readonly  
> **Дата:** 2026-07-23  
> **Scope (выборка):** `docs/wiki/API.md`, `docs/wiki/QUICKSTART.md`, `docs/wiki/SHORTCUTS.md` (смежно), `docs/vision_and_scope.md` vs `backend/internal/{auth,notification,search,calls,ai,bases,project}/*`, `frontend/src/{components/layout/app-shell.tsx,components/search/command-palette.tsx,routes/onboarding.lazy.tsx,routes/.../gantt.tsx}`  
> **Код не менялся** · правки wiki/vision — backlog docs-sync (rule 55)

---

## Summary

Выборочная сверка показывает **реальный drift**: wiki/API отстаёт от живых admin/Telegram-роутов; QUICKSTART устарел по онбордингу и командам регрессии; vision противоречит TZ по SSO и помечает как «отложено» уже существующий Gantt UI. Базовые инварианты Magic Link / Ctrl+K / LiveKit / bases owner-scope в docs в целом **PASS**.

**Вердикт:** docs drift **подтверждён** (0 CRITICAL security-doc-lie; **3 HIGH** omission/contradiction; несколько MEDIUM/LOW). `verified: false` для «wiki = код» до закрытия HIGH.

---

## Findings

| Sev | Doc locus | Issue | Evidence (код) | Fix (docs) |
|-----|-----------|-------|----------------|------------|
| **HIGH** | `API.md` § Auth + § Admin / Audit | **ORG-06 admin API не документирован** в таблицах маршрутов: pending queue / approve / reject / org invite. Vision уже ссылается на invite; wiki — нет | `auth/handler.go:52-55`: `GET /admin/users/pending`, `POST .../approve`, `POST .../reject`, `GET /admin/organization/invite`. Swagger в `backend/docs/` есть. § Admin перечисляет deactivate/ban/gdpr, **без** pending | Добавить 4 строки в Auth и/или Admin; обновить счётчик «Auth» / «Admin (12)» |
| **HIGH** | `API.md` § Notifications (заголовок «10 endpoints») + intro ADR-AUTH-01 | Intro обещает `/notifications/telegram/*`, таблица **не содержит** Telegram bind API → читатель не найдёт контракт | `notification/handler.go:38-42`: `POST .../telegram/bind`, `.../confirm`, `DELETE .../telegram`, `POST .../test`, `GET .../status` (под `authMw`) | Дописать 5 эндпоинтов; заголовок → **15**; согласовать с QUICKSTART «Настройки → Telegram» |
| **HIGH** | `vision_and_scope.md` Won't Have (v1–v2 / v4) | **SSO/LDAP = Won't Have**, тогда как TZ: **одобрено в scope** (2026-07-03), реализация после `ADR-AUTH-SSO` | `docs/tz_constraints.md` ⚠️ SSO/LDAP одобрено; Magic Link остаётся break-glass. Vision L118, L150: «SSO… Won't» | Переписать: Won't = соц. OAuth; SSO/LDAP = **in scope, blocked on ADR**; ссылка на `ADR-AUTH-SSO-2026.md` |
| **MEDIUM** | `QUICKSTART.md` § Онбординг | «**3 шага**» (профиль / обзор / настройки) | `onboarding.lazy.tsx:153-156` — **4** шага: profile → tour → theme → project. `vision` US-M08 уже корректен (4 шага) | Синхронизировать QUICKSTART с US-M08 / UI |
| **MEDIUM** | `QUICKSTART.md` § Минимальная регрессия | `go test … ./internal/... ./testframework/...` — вводит в заблуждение (тестов в `internal/` нет) | CI: `go test -tags hubtest … ./testframework/...` (`.github/workflows/ci.yml:61`). `AGENTS.md`: 0 tests in `internal/` | Убрать `./internal/...`; добавить `node scripts/link-testframework-node-modules.mjs` перед Vitest |
| **MEDIUM** | `vision_and_scope.md` § Текущий статус / polish | Фокус «Wave C · FA-06»; Gantt в «отложенный polish Phase 16» | Gantt UI жив: `routes/.../gantt.tsx` + `modules/tracker/gantt-chart.tsx` (GANTT-01). Актуальный фокус: July Rebuild planning + AI-expansion (`AGENTS.md` / `STATE.md`) | Обновить «Текущий фокус»; Gantt — **shipped MVP** (ограничения: due_date-only), webcam/hub-mail — Phase 16 |
| **MEDIUM** | `vision_and_scope.md` US-F06 | OpenAI-compatible API помечен **⏳ v3.5 backlog** | `ai/handler.go:42` — `POST /ai/v1/chat/completions`; `API.md` § AI уже документирует | Статус US-F06 → ✅ (или Partial, если нет полного паритета OpenAI) |
| **MEDIUM** | `SHORTCUTS.md` + Settings hotkeys UI | Нет `Ctrl+Shift+A` (AI panel); Settings/hotkeys не зеркалит shell | `app-shell.tsx:78-87` — `Ctrl/Cmd+Shift+A` → `toggleAIPanel`. `settings/hotkeys.tsx:9-15` — только K/N/Esc/Enter | Дописать в SHORTCUTS + HOTKEYS-массив UI |
| **LOW** | `API.md` § Search | Описание: «задачи, тест-кейсы, сообщения» — неполный список типов | `search/service.go` + `command-palette.tsx` GROUP_ORDER: issue, project, test_case, channel, message, user, file | Расширить описание; параметр `types=` |
| **LOW** | `API.md` § Wiki planned | Утверждает `GET /search?types=wiki_page` | В `search/service.go` типа `wiki_page` **нет** (модуля wiki нет — planned OK) | Пометить search-тип как planned вместе с модулем; не обещать query сейчас |
| **LOW** | `API.md` § Admin HTTP audit | «подключён … в `main.go`» | `cmd/server/wiring_fiber.go:54` — `HTTPAuditMiddleware` | Исправить путь файла |
| **LOW** | `API.md` шапка «Обновлено: 2026-06-22» / QUICKSTART «2026-06-21» | Даты устарели относительно кода (Wave 1 merge 07-11, AI-expansion) | Факт: роуты pending/telegram/gantt новее дат wiki | Bump даты после docs-fix |
| **LOW** | `vision` KPI «31 пакет в internal/» | Подсчёт каталогов ≈ **33** | `backend/internal/*` dirs | Пересчитать KPI или уточнить «domain vs platform» |
| **INFO** | `API.md` Wiki section | Корректно **planned/spec**, кода нет | Нет `backend/internal/wiki/` | Держать planned до реализации |
| **PASS** | Auth Magic Link only (HTTP surface) | Wiki ↔ handler | `auth/handler.go:35-46` без `/telegram` login | — |
| **PASS** | JWT 15m / refresh 30d / error shape | Wiki intro | Согласовано с TZ / session service | — |
| **PASS** | Ctrl+K command palette | QUICKSTART + SHORTCUTS + vision US-M07 | `app-shell.tsx:54-57`, `command-palette.tsx` | — |
| **PASS** | Ctrl+N новая задача | QUICKSTART + SHORTCUTS | `app-shell.tsx:60-75` → `/issues/new` | — |
| **PASS** | Bases owner-scoped | `API.md` Hub Base | `bases/handler.go` + service authorize (W1/bases) | — |
| **PASS** | Calls raise-hand / LiveKit webhook | `API.md` Calls | `calls/handler.go:51,80` | — |
| **PASS** | Go 1.25 в QUICKSTART | Версия стека | `backend/go.mod` `go 1.25.0`; CI `go-version: '1.25'` | — |
| **PASS** | `.env.local.example` | QUICKSTART seed path | Файл в корне репо существует | — |

---

## Матрица выборки (doc claim → код)

| Claim | Источник | Статус |
|-------|----------|--------|
| Magic Link only; TG ≠ login | API / QUICKSTART / vision US-M01 | **PASS** (HTTP); service-layer TG login — вне wiki (см. W1-auth) |
| Telegram notify bind API | API intro | **FAIL** — код есть, таблица пуста |
| Pending approve / org invite | vision Org; отсутствует в API tables | **FAIL** wiki omission |
| Онбординг 3 шага | QUICKSTART | **FAIL** (код/US-M08 = 4) |
| Ctrl+K сущности | QUICKSTART / SHORTCUTS / US-M07 | **PASS** (+ message в коде, часто не упомянут) |
| `types=wiki_page` | API Wiki | **FAIL** premature (planned) |
| SSO Won't Have | vision | **FAIL** vs `tz_constraints.md` |
| Gantt Phase 16 only | vision polish | **FAIL** — UI уже в master |
| US-F06 backlog | vision | **FAIL** — route shipped |
| Регрессия `./internal/...` | QUICKSTART | **FAIL** |
| ~286 REST / 82 Ent | API / vision KPI | **PASS** (order-of-magnitude; точный recount не делался) |
| Wiki module planned | API | **PASS** |

---

## 5 top risks (для читателя docs)

1. **Admin pending API скрыт в wiki** — онбординг invite_only по vision не собрать по `API.md` alone → операционный/интеграционный drift.
2. **Telegram bind недокументирован в таблице** — противоречие ADR-блоку в том же файле; риск «вернуть login-роуты» из-за путаницы surface.
3. **SSO Won't vs TZ in-scope** — BA/PO/архитектор могут планировать в разные стороны до ADR.
4. **QUICKSTART regression path** — агенты/новички гоняют пустой `./internal/...` и думают, что «тестов нет».
5. **Gantt «ещё Phase 16»** при живом UI — ложный backlog и дублирование работы.

---

## Рекомендации (docs-only, без кода)

1. **P0:** `API.md` — секции Auth/Admin (pending*) + Notifications (telegram*).  
2. **P0:** `vision_and_scope.md` — SSO wording ↔ `tz_constraints.md`; статус Gantt / US-F06 / «Текущий фокус».  
3. **P1:** `QUICKSTART.md` — 4-step onboarding + canonical test commands из `AGENTS.md`.  
4. **P1:** `SHORTCUTS.md` (+ UI hotkeys page) — `Ctrl+Shift+A`.  
5. **P2:** Search types; Wiki search note; `wiring_fiber.go` вместо `main.go`; bump «Обновлено».

---

## Verification

| Этап | Результат |
|------|-----------|
| Файлы docs прочитаны | API.md (шапка + Auth/Admin/Notif/Search/Wiki/AI/Calls/Bases), QUICKSTART, SHORTCUTS (частично), vision (KPI, MoSCoW, Won't, status, US-F/V) |
| Факты сверены с кодом | auth/notification/search/calls/ai/bases handlers; app-shell; command-palette; onboarding.lazy; gantt; go.mod; ci.yml; tz_constraints |
| Код не менялся | да |
| Выборка, не полный recount 286 routes | да — счётчики секций могут иметь доп. drift |
| Severity | HIGH = omission живого API / contradiction TZ; не security exploit |

---

## Handoff

- **Артефакт:** `.planning/audits/2026-07/repo-scan/W7-docs-drift.md`  
- **Next:** `docs-writer` (write-mode) или волна docs-fix по P0; `team-lead` — учесть в SCORECARD волны 8  
- **Связанные аудиты:** W1-auth (pending session), W2-notification-search (TG bind PASS в коде)
