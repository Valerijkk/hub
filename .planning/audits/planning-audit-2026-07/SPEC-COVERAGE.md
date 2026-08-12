# SPEC-COVERAGE — FR → phase → status → evidence

> Planning-audit **C3** · docs only · 2026-07-27  
> Source of FR IDs: `docs/system_specification.md` (numbered `FR-*` only).  
> Module prefix rows (`FR-INTR` / `FR-CAL` / `FR-WIKI` without number) are section labels, not requirements.

## Method

| Field | Rule |
|-------|------|
| **FR** | Unique `FR-(INTR\|CAL\|WIKI)-\d+` in `docs/system_specification.md` |
| **Phase** | Numbered GSD phase in `.planning/ROADMAP.md` / `.planning/phases/` **or** planned gap phase `126+`. July wave tasks from `.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md` listed as **Wave** (secondary). |
| **Status** | `done` · `partial` · `planned` · `missing` (code vs FR wording) |
| **Evidence** | Prefer `path:line` in production code; else schema / PLAN / «absent» |

**Inventory:** **28** FRs (INTR 11 + CAL 8 + WIKI 9).

**C5 `spec_coverage` rule:** FR counts as «has a phase» only if **Phase** column is non-empty (numbered GSD / planned 126+). July Wave alone does **not** count — matches audit A6 (Wiki in July waves but no GSD phase).

| Metric | Value |
|--------|-------|
| FR total | 28 |
| FR with numbered/planned phase | **9** |
| FR with no numbered/planned phase | **19** |
| `spec_coverage` | **9/28** (≈ **0.321**) |

---

## Matrix

### Intranet (FR-INTR)

| FR | Spec (short) | Phase | Wave | Status | Evidence |
|----|--------------|-------|------|--------|----------|
| FR-INTR-01 | Poll attached to news post (2–10 options) | — | W1.10 US · W2.8 spec · W3.e FE poll-in-post | **missing** | No news↔poll link. Messenger polls only: `backend/ent/schema/poll.go:22` (`channel_id`). Intranet news routes: `backend/internal/intranet/handler.go:38-44` — no `/poll`. |
| FR-INTR-02 | One vote / user; % distribution | — | W3.e / W4 | **missing** | No intranet poll vote API. Messenger: `backend/internal/messenger/service_polls.go` (channel polls ≠ FR). |
| FR-INTR-03 | Lock options after first vote; anonymous aggregates | — | W3.e / W4 | **missing** | Same — no news-poll domain. |
| FR-INTR-04 | Admin assigns onboarding **template** copy | — | W1.10 · W2.8 · W4.a | **missing** | Auth profile onboarding only: `backend/internal/auth/service_profile.go:34-39` (`CompleteOnboarding`). No `onboarding_template` schema under `backend/ent/schema/`. |
| FR-INTR-05 | Owner-only checklist items; Admin progress list | — | W4.a | **missing** | No onboarding assignment/items endpoints in `backend/internal/intranet/handler.go:22-45`. |
| FR-INTR-06 | Re-assign needs `confirm=true`; deactivate freezes | — | W4.a | **missing** | Same — no assignment API. |
| FR-INTR-07 | Soft-delete news → trash 30d + audit; cascade hide poll | **Phase 55** (intranet coverage) · news soft-delete shipped | W4 | **partial** | Soft-delete: `backend/internal/intranet/service.go:168-172`, `repository.go:124`; schema `deleted_at`: `backend/ent/schema/newspost.go:38-40`. Trash types omit `newspost`: `backend/internal/audit/service.go:197-204`. No news poll to cascade. |
| FR-INTR-08 | Restore from trash returns post + poll + counters | **Phase 126** (planned) | — | **planned** | Gap confirmed; PLAN: `.planning/phases/126-intranet-newspost-trash/PLAN.md`. Trash restore arms: `backend/internal/audit/service.go:226-237` (issue/test_case/file/project only). |
| FR-INTR-09 | Kudos **leaderboard** by period month/quarter/all | **Phase 44** (Kudos) | W4.a | **partial** | Feed/received exist: `backend/internal/intranet/handler.go:28-29`, `service.go:250`, `268`. **No** `/kudos/leaderboard` route in `handler.go:27-30`. |
| FR-INTR-10 | Weekly matching: pairs, odd→triple, anti-repeat | **Phase 47** (Random Coffee) | W4.a | **partial** | Matching: `backend/internal/intranet/coffee_service.go:114-137`, `buildCoffeePairs:201-229`. Pairs only (2-person); odd leftover → `unpaired`, **not** a triple. Anti-repeat via `prevPairs`. Manual admin `POST /coffee/run` (`handler.go:37`), not weekly cron. |
| FR-INTR-11 | Notify-only per pair/triple (ADR-INTRANET-U05); auto-DM deferred | **Phase 47** | W4.a | **covered** | `CoffeeMatchDeliveryMode=notify_only`; `notifyCoffeeMatch` + slog; IT `coffee_notify_task106_test.go`. Auto-DM out of scope until new ADR. |

### Calendar (FR-CAL)

| FR | Spec (short) | Phase | Wave | Status | Evidence |
|----|--------------|-------|------|--------|----------|
| FR-CAL-01 | Resource catalog CRUD (Admin); kind room/equipment/other | — | W1.11 · W2.9 · W3.e/f | **missing** | No `resource` / `Resource` schema in `backend/ent/schema/`. Calendar routes only list+export: `backend/internal/calendar/handler.go:22-25`. |
| FR-CAL-02 | Book resource on event; overlap → 409 | — | W3.f · W4.b | **missing** | No `resource_booking` / exclusion constraint in schemas. |
| FR-CAL-03 | Recurrence freq d/w/m + until XOR count | **Phase 41** (unified calendar) + calls scheduled | W3.f · W4.b | **partial** | `ScheduledCall.rrule` RFC5545: `backend/ent/schema/scheduledcall.go:41-43`; parse: `backend/internal/calls/scheduled.go:46`. **No** `calendar_event` / `event_recurrence` entities from Wave-2 FR/ERD. |
| FR-CAL-04 | Edit series = all occurrences | **Phase 41** / calls | W3.f | **partial** | Update path can set RRULE: `backend/internal/calls/repository.go:367`. Spec’s first-class `calendar_event` series absent. |
| FR-CAL-05 | Delete series → trash 30d; per-occurrence resource check | **Phase 127** (planned) | — | **planned** | Delete = deactivate only: `backend/internal/calls/repository.go:377-379` (`SetIsActive(false)`); **no** `deleted_at` on `scheduledcall.go`. PLAN: `.planning/phases/127-calendar-series-trash/PLAN.md`. |
| FR-CAL-06 | Personal iCal link; ≥128-bit token; rotate invalidates | **Phase 41** | W4.b | **partial** | Auth-gated export: `backend/internal/calendar/handler.go:25`, `63-79` (`/export.ics`). **No** secret token / rotate / hash-at-rest (`ical_export_token` schema absent). |
| FR-CAL-07 | Subscribe external iCal URL; SSRF guard; ≤10 | — | W4.b | **missing** | No `/calendar/subscriptions` in `handler.go:22-25`. No `ical_subscription` schema. |
| FR-CAL-08 | Availability busy/free (+ approved HR leave) | — | W4.b | **missing** | No `/calendar/availability`. Leave exists in HR module separately (Phase 43) but not aggregated here. |

### Wiki (FR-WIKI)

| FR | Spec (short) | Phase | Wave | Status | Evidence |
|----|--------------|-------|------|--------|----------|
| FR-WIKI-01 | Space CRUD; page tree; move; depth≤10 | — | W1.12 US · W2.10 spec · **W5.e** start | **missing** | No `backend/internal/wiki/`; no wiki schemas in `backend/ent/schema/`; no FE wiki routes (`frontend/src` wiki = 0). Spec marks «Волны 5/8»: `docs/system_specification.md:41`. |
| FR-WIKI-02 | Optimistic lock version; presigned images | — | W5.e | **missing** | Absent module. |
| FR-WIKI-03 | Soft-delete subtree → trash 30d | — | W5.e | **missing** | Absent; trash types unchanged (`audit/service.go:197-204`). |
| FR-WIKI-04 | Space RBAC read/write/admin; ≥1 admin | — | W5.e | **missing** | Absent. |
| FR-WIKI-05 | Page restrict allow-list | — | **W8.a** | **missing** | Absent. |
| FR-WIKI-06 | Version on each save; diff | — | W8.a | **missing** | Absent. |
| FR-WIKI-07 | Rollback → new version N+1 | — | W8.a | **missing** | Absent. |
| FR-WIKI-08 | Issue↔page M:N links | — | W8.a | **missing** | Absent. |
| FR-WIKI-09 | Search `types=wiki_page` user-scoped | — | W8.a | **missing** | Absent wiki type in search module. |

---

## FRs with no numbered / planned GSD phase

These have **July wave** references (US/spec/implementation bullets) but **zero** entry in `.planning/ROADMAP.md` phase checklist / `.planning/phases/<n>-*` (and no 126+ PLAN). Per C5/A6 they do **not** count toward `spec_coverage`.

| FR | July wave (if any) | Notes |
|----|--------------------|-------|
| FR-INTR-01 | W3.e | Poll-in-post not implemented |
| FR-INTR-02 | W3.e / W4 | |
| FR-INTR-03 | W3.e / W4 | |
| FR-INTR-04 | W4.a | Distinct from auth `CompleteOnboarding` |
| FR-INTR-05 | W4.a | |
| FR-INTR-06 | W4.a | |
| FR-CAL-01 | W3.e/f | Resource catalog |
| FR-CAL-02 | W3.f / W4.b | Booking |
| FR-CAL-07 | W4.b | iCal subscriptions |
| FR-CAL-08 | W4.b | Availability |
| FR-WIKI-01 | W5.e | Core wiki — audit §5 |
| FR-WIKI-02 | W5.e | |
| FR-WIKI-03 | W5.e | |
| FR-WIKI-04 | W5.e | |
| FR-WIKI-05 | W8.a | |
| FR-WIKI-06 | W8.a | |
| FR-WIKI-07 | W8.a | |
| FR-WIKI-08 | W8.a | |
| FR-WIKI-09 | W8.a | |

**Count:** 19 / 28.

### FRs that *do* have a numbered/planned phase (9)

| FR | Phase | Status |
|----|-------|--------|
| FR-INTR-07 | 55 | partial |
| FR-INTR-08 | 126 | planned |
| FR-INTR-09 | 44 | partial |
| FR-INTR-10 | 47 | partial |
| FR-INTR-11 | 47 | covered (notify-only ADR-U05) |
| FR-CAL-03 | 41 (+ calls scheduled) | partial |
| FR-CAL-04 | 41 / calls | partial |
| FR-CAL-05 | 127 | planned |
| FR-CAL-06 | 41 | partial |

---

## Related artifacts

- Audit: `.planning/audits/planning-audit-2026-07/00-PLANNING-AUDIT.md` §4.3, §5  
- Fixes: `.planning/audits/planning-audit-2026-07/01-PLANNING-FIXES.md` A6, B1, B2, B5  
- Gap PLANs: `PW4-D3-FR-PLANS.md` · phases `126-*`, `127-*`  
- Wiki scope options: `WIKI-SCOPE-OPTIONS.md`  
- C5 paste block: `PW3-C5-spec-coverage.md` (after this file)

---

*Generated 2026-07-27 · C3 SPEC-COVERAGE*
