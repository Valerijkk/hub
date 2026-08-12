# Encoding deep spot-check: P1 TASK backlog (124–176)

**Date:** 2026-07-29  
**Method:** full-file human read (5 mandated) + first-40-lines read (10 random) + UTF-8 decode + mojibake needle scan  
**Patterns:** `U+FFFD`, double-UTF8 Latin (`Ðа`/`Ñ`/`Ã`/`Â`/`â€`), prior Hub markers (`╨` `╤` `тЖТ` `тАФ` `ЁЯ` `пїЅ`), broken emoji / wrong quotes  
**Fixes applied:** none (0 FAIL)

## Counts

| Metric | Value |
|--------|------:|
| Full-file deep reads | 5 |
| Random first-40 sample | 10 |
| Total files checked | 15 |
| PASS | 15 |
| FAIL (real mojibake) | 0 |

## Mandatory deep (full content)

| File | Result | Notes |
|------|--------|-------|
| `TASK-126-2026-07-29-a3-10-3-fe-sprint-backlog-regress.md` | PASS | H1 `♻️🔒` intact; RU prose clean; ASCII/`«»` quotes OK |
| `TASK-129-2026-07-29-a3-11-1-fe-api-error-mapping.md` | PASS | H1 `♻️🔒` + `→` OK; wire JSON/`ApiError` quotes intentional ASCII |
| `TASK-130-2026-07-29-a3-11-2-fe-projects-sprint-version.md` | PASS | H1 emoji OK; tables/code fences clean RU |
| `TASK-134-2026-07-29-a3-1-2-tracker-services.md` | PASS | H1 `🔒` OK; Per-ID table Cyrillic OK |
| `TASK-171-2026-07-29-a3-8-5-schemas.md` | PASS | H1 `🔒` OK; soft-delete / Cascade RU OK |

## Random sample (first ~40 lines) — from TASK-124…176 excl. mandatory

Selection (spread): 124, 128, 137, 142, 147, 153, 158, 163, 168, 175.

| File | Result | H1 emoji / RU |
|------|--------|---------------|
| `TASK-124-…-a3-10-1-fe-routes-modules.md` | PASS | `🧹` · error≠empty prose OK |
| `TASK-128-…-a3-10-5-fe-tms-onboarding-auth.md` | PASS | `🔒` · canMutate / onboarding OK |
| `TASK-137-…-a3-1-5-project.md` | PASS | `🔒` · archive/trash OK |
| `TASK-142-…-a3-2-5-bases.md` | PASS | `🧹` · hubbase trash OK |
| `TASK-147-…-a3-3-5-calls.md` | PASS | `🐛` · LiveKit orphan OK |
| `TASK-153-…-a3-5-1-ai-handlers.md` | PASS | `🐛` · ErrNoProvider OK |
| `TASK-158-…-a3-6-2-auth-tests.md` | PASS | `🔒` · deactivated JWT OK |
| `TASK-163-…-a3-7-2-v4-business.md` | PASS | `♻️` · coffee notify OK |
| `TASK-168-…-a3-8-2-schemas.md` | PASS | `🐛` · OnDelete Cascade OK |
| `TASK-175-…-a3-9-4-components.md` | PASS | `🧹` · checklist/error UI OK |

## Automated corroboration

- Valid UTF-8 decode: 15/15  
- Cyrillic present: 15/15  
- Mojibake needles / `U+FFFD`: 0 hits  
- Emoji in H1 render as intended (♻️ 🔒 🧹 🐛), not `ЁЯ`-style broken sequences  

## Verdict

**ALL PASS** — no real mojibake, broken emoji, or wrong-quote garble in the deep P1 set or the random 10. No TASK files modified.
