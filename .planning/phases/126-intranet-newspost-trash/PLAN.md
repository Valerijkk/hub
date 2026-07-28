# Phase 126 — FR-INTR-08 NewsPost trash restore

**Status:** IMPLEMENTED 2026-07-28 (TASK-001) — see `VERIFICATION.md`  
**Source:** `docs/system_specification.md` FR-INTR-08 (US-INTR-03), NFR-11  
**Audit:** `.planning/audits/planning-audit-2026-07/00-PLANNING-AUDIT.md` §4.3 · `01-PLANNING-FIXES.md` B1  
**Index note:** `.planning/audits/planning-audit-2026-07/PW4-D3-FR-PLANS.md`  
**Suggested branch (when implementing):** `feat/intranet/newspost-trash-restore`

## Goal

Close confirmed gap **FR-INTR-08**: restoring a soft-deleted news post from the unified admin trash returns the post **together with its poll and vote counters** (when a poll exists). Today `NewsPost.deleted_at` exists and delete soft-sets it, but `newspost` is **not** registered in trash list/restore/purge — data remains in DB with no admin restore path.

## Gap (confirmed)

| Fact | Path |
|------|------|
| Spec restore via Trash API | `docs/system_specification.md` (FR-INTR-07/08, `DELETE /api/intranet/news/:newsId` + `/admin/trash`) |
| Schema already has `deleted_at` | `backend/ent/schema/newspost.go` |
| Soft delete on delete | `backend/internal/intranet/service.go` (`DeleteNews`) · `backend/internal/intranet/repository.go` (`SoftDeleteNewsPost`) |
| Trash types = issue / test_case / file / project only | `backend/internal/audit/service.go` (`ListTrash`, `RestoreItem`, `SoftDelete`, `PurgeExpired`) |
| Routes | `backend/internal/audit/handler.go` (`GET /admin/trash`, `POST /admin/trash/:type/:id/restore`) |
| Wiki surface | `docs/wiki/API.md` (admin trash rows) |

## Tasks

- [x] Register trash type `newspost` in `backend/internal/audit/service.go` switch arms: `ListTrash`, `RestoreItem`, `SoftDelete`, `PurgeExpired` (retention = `TrashRetentionDays` = 30, NFR-11).
- [x] Extend `backend/internal/audit/repository.go` (+ Ent impl): `ListDeletedNewsPosts`, `RestoreNewsPost` (clear `deleted_at`), `SoftDeleteNewsPost` (if not already shared), `PurgeExpiredNewsPosts` (`deleted_at <= cutoff` hard delete).
- [x] Ensure restore is **poll-aware** per FR-INTR-08: after restore, GET news (+ poll endpoints when present) returns the same poll options and denormalized `votes_count` / vote rows as before delete. Soft-delete of the post must continue to hide poll/votes from public reads (FR-INTR-07 cascade hide — one trash entry for the post). *(poll schemas absent — contract ready via ClearDeletedAt only)*
- [x] Wire intranet delete path so soft-deleted posts are visible to trash listing (`deleted_at` set, published list filters unchanged). Prefer reusing audit SoftDelete/register pattern rather than a second restore API.
- [ ] Audit log on delete/restore (actor, post_id) — align with FR-INTR-07 wording; emit `hub.intranet.news.restored` when outbox/NATS pattern already used for `hub.intranet.news.deleted` (spec events table). *(deferred: no existing news.deleted outbox in intranet)*
- [x] Tests under `backend/testframework/audit/` and/or `backend/testframework/intranet/`: soft-delete → appears in `GET /admin/trash?type=newspost` → `POST .../restore` → post (+ poll/votes if fixture has poll) readable again; purge after 30d cutoff; unknown type still errors; non-admin → 403.
- [x] Docs sync on implement: `docs/wiki/API.md` trash type list; session note in `.planning/STATE.md`; mark FR-INTR-08 covered in planning index / this PLAN checkboxes.

## Acceptance criteria

1. Soft-deleted `NewsPost` appears in `GET /api/admin/trash` for `type=newspost` with `days_remaining` within 30 days (NFR-11).
2. `POST /api/admin/trash/newspost/:id/restore` clears `deleted_at`; subsequent intranet GET returns the post.
3. If the post had a poll with votes, restore returns poll + option `votes_count` / aggregates unchanged (FR-INTR-08). Post without poll restores cleanly.
4. After retention, purge hard-deletes expired soft-deleted posts (and cascade poll/votes rows if schemas exist) so trash does not retain forever.
5. Member cannot restore; admin/founder can. Public/list news endpoints never return soft-deleted posts.
6. Regression: existing trash types (`issue`, `test_case`, `file`, `project`) behavior unchanged.

## Out of scope

- Implementing full Wave 2 poll CRUD (FR-INTR-01..03) if `news_poll*` schemas are still absent — restore contract must be ready; poll fixtures only when schemas land.
- Frontend trash UI redesign (admin trash page may already list types; only add `newspost` if UI is type-driven).
- Hard delete without trash, GDPR export, or changing Magic Link / auth.
- Wiki module (FR-WIKI-*) or other NFR-11 entity rollouts (B4).
- Production code in this planning commit — **PLAN only**.

## Depends on

- Нет фазовой зависимости (схема `NewsPost.deleted_at` уже есть).
- Soft: единый Trash API (`backend/internal/audit`) — расширяется в этой фазе.

## Dependencies / notes

- Depends on existing soft-delete field in `backend/ent/schema/newspost.go` (no schema migration for `deleted_at`).
- If `news_poll` / `news_poll_option` / `news_poll_vote` are added in a parallel Wave 2 phase, cascade hide/restore must be one transaction with the post.
- Pattern to copy: issue restore/list/purge in `backend/internal/audit/service.go` + `repository.go`.

## Verify (when implementing)

```bash
cd backend && go test -tags hubtest -race -count=1 ./testframework/audit/... ./testframework/intranet/...
```
