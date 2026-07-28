# Hub — Project Status Snapshot (2026-07-10)

> **Снимок «где мы»** перед кодовым freeze до возврата ~конец июля 2026.
> Индекс: `.planning/archive/july-2026/INDEX-JULY-2026.md` · Brief: `JULY-2026-REBUILD-BRIEF.md` · Операционный план: `ROADMAP-DETAILED-JULY-RETURN.md`
> Метрики baseline: `AGENTS.md` / `.planning/STATE.md` / `.planning/COVERAGE.md`

---

## 1. Вердикт одной строкой

**v1–v4 + coverage + TF + Wave C (кроме FA-06) — done на `master`.** Волна 1 ROADMAP-FINAL **исполнена и верифицирована в фича-ветках** (`integration/wave1` merge-ready). Код July Rebuild **не стартовал**. SFU = **LiveKit** (mediasoup REJECTED). Active planning: SEED-002 CAL-SYNC, EPIC-REDESIGN, EPIC-QA-HARDENING (каталог + Testing Guide).

---

## 2. Milestone / метрики (канон)

| Параметр | Значение | Источник |
|----------|----------|----------|
| Milestone | v4 Super-app ✅ · post-v4 planning (July Rebuild) | STATE |
| Phases 1–125 | 100% completed | STATE progress |
| COV overall | **83.0%** (2026-06-22 Docker baseline, `-tags hubtest`) | COVERAGE / AGENTS |
| Security pkgs ≥80% | **6/6** | AGENTS |
| Product score | **89/100** | AGENTS / AUDIT |
| REST API | **286+** endpoints | AGENTS |
| Ent schemas | **82** | AGENTS |
| Backend tests | **495** in `backend/testframework/` · **0** in `internal/` | STATE |
| Vitest | **99/99** · 37 files under testframework unit domains | STATE |
| Playwright E2E | **20** specs · smoke 3/3 | STATE / FA-10 |
| qa-gate (FA-10) | PASS | FA-10-QA-GATE |

### Milestones

| Milestone | Status | Date |
|-----------|--------|------|
| v1.0 MVP | ✅ SHIPPED | 2026-03-21 |
| v2.0 Hub Redesign | ✅ COMPLETE | 2026-05-15 |
| v2.5 Trust | ✅ COMPLETE | 2026-06-22 |
| v2.6 Automate + Arch + OBS | ✅ COMPLETE | 2026-06-24 |
| v3.0 AI Native | ✅ COMPLETE | 2026-06-22 |
| v4 Super-app | ✅ COMPLETE | 2026-06-25 |
| Coverage wave Ph 59–125 | ✅ COV 83.0% (2026-06-22 Docker baseline) | 2026-06-21..22 |
| TF migration Wave B | ✅ 0 internal tests | 2026-06-23 |
| Final audit Wave C | ✅ кроме FA-06 manual | 2026-06-22 |

---

## 3. Ветки / волны — что ждёт пользователя

| Ветка | Содержимое | Статус | Действие пользователя |
|-------|------------|--------|------------------------|
| `master` | Baseline метрик, v4, TF, Wave C | Канон prod-ready кода | Не трогать агентами |
| `integration/wave1` | Интеграция Волны 1 | ✅ Docker suite PASS · tsc · Vitest 99/99 · **merge-ready** | **Rebase-merge в master** (личное ревью: bases security) |
| `feat/bases/security-membership-fix` | Owner-scoping bases + tests | ✅ be-ver + security verified | Входит в integration/wave1 |
| `fix/platform/wave1-quickfixes` | audit N+1, magic-link runes, GIPHY env | ✅ be-ver | Входит в integration/wave1 |
| `docs/wave1-analysis` | US/AC intranet/calendar/wiki + ROADMAP drift + FA-06 runbook | ✅ ba-ver (14 US) | Входит в integration/wave1 |

**Политика:** агенты коммитят/пушат фича-ветки; **push/мерж `master` — только пользователь**; rebase-merge.

**Follow-ups Волны 1 → Волна 2 (после merge):**
- VITE_GIPHY_API_KEY в prod-сборке + `wiki/DEPLOY.md`
- bases: owner-only vs project-membership sharing → эпик BS-1
- NF-09 presigned PUT size limit → Волна 2 devops

---

## 4. Карта эпиков

| Эпик / трек | Статус | Примечание |
|-------------|--------|------------|
| Track F W1 quick-wins | ✅ done (ветки) | Ждёт merge пользователя |
| Track F W2–12 | ⏸ deferred | Не стартовать код до July align |
| Track P M0–M4 | ⏸ deferred | ADR microservices **accepted**; исполнение после align |
| EPIC DOCS-EXCELLENCE | 🟡 partial | LICENSE/README/SECURITY на master; wiki polish — docs-lane |
| TF-PY | ⏸ deferred | После M2–M4 |
| ~~EPIC-SFU (mediasoup)~~ | ❌ **REJECTED 2026-07-10** | LiveKit остаётся; SEED-001 rejected |
| **EPIC-CAL-SYNC** | 🟢 active (planning) | SEED-002; код с возврата |
| **EPIC-REDESIGN** | 🟢 active (planning) | Cloud Code Designer → DESIGN-SPEC |
| **EPIC-QA-HARDENING** | 🟢 planning | Catalog + Testing Guide; execution с возврата |
| FA-06 secret rotation | 🔴 ops blocker | Manual; runbook ready |
| SSO (ADR-AUTH-SSO) | ⏸ | Scope одобрен; реализация после ADR |
| Multi-org / SaaS billing / native apps | ❌ out of scope | TZ |

---

## 5. Блокеры

| ID | Тип | Блокирует | Владелец | Когда |
|----|-----|-----------|----------|-------|
| **FA-06** | Ops / prod cutover | SSH deploy, tag v2.5.0 | Пользователь (+ devops) | До prod; **не** гейтит July feature-dev |
| **Merge Wave 1** | Git | Чистый master для W2+ | Пользователь | День 0 возврата (рекоменд.) |
| **July Align** | Product | Приоритет CAL vs Redesign vs QA; mail L0–L3; reschedule-RSVP | product + пользователь | Неделя 0 |
| **SEED-001** | — | — | — | Закрыт (REJECTED) |

---

## 6. Архитектурный канон (не пересматривать без ADR)

- Magic Link only; Telegram ≠ login
- 1 org / instance; UUID; UTC; optimistic `version`; soft-delete 30d
- Fiber WS + NATS JetStream; SeaweedFS; TanStack Router; Shadcn className-only
- **Calls SFU = LiveKit** (M4 `hub-calls` = LiveKit-домен)
- Microservices: strangler-fig; core + hub-ai + realtime-трио + notification pilot; tracker не пилим
- Calendar v2 GZ-V4-05: iCal export/subscribe **без** OAuth; CAL-SYNC L3 = новый ADR

---

## 7. Рекомендуемый порядок при возврате (недели 0–N)

| Неделя | Фокус | Выход |
|--------|-------|-------|
| **0** Align (1–3 дн) | Brief §5 Qs; merge wave1; FA-06 если prod; приоритеты эпиков | Locked decisions + kickoff note |
| **1** Planning expansion | Edge catalog → TC; Testing Guide adoption; tech-lead дробление | DoR для фаз |
| **2** Specs | ba/sa/arch CAL-SYNC; designer Redesign inventory | US/AC + ERD/API + ADR-CAL-UNIFIED draft |
| **3+** Execute | QA-HARDENING TC→auto; CAL-SYNC be/fe; Redesign fe waves | Зелёные волны + merge user |
| **Параллельно** | Track P M0 только если align = «продолжать каркас» | Иначе freeze W2–11 |

Детали: `ROADMAP-DETAILED-JULY-RETURN.md`.

---

## 8. Ссылки

- Edge/corner → TC: `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`
- Testing Guide: `testframework/TESTING_GUIDE.md`
- Quality: `.planning/archive/july-2026/CODE-QUALITY-CHARTER.md`
- Seeds: `seeds/SEED-001-*.md` (REJECTED), `seeds/SEED-002-*.md` (active)
- Roadmap: `ROADMAP-FINAL-2026-07-03.md`

---

*Snapshot tech-lead 2026-07-10 · записано на диск parent-агентом (Ask mode у tech-lead). Не коммитить без запроса пользователя.*
