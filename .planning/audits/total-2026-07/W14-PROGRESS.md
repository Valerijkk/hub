# W12/W14 progress — Stage B (2026-07-28)

> Обновлено оркестратором после P0 + P1 фиксов и старта P2

## P0 / P1 на проверке (ветка `cursor-grok-4.5`)

| TASK | Scope | Commit(s) |
|------|-------|-----------|
| 009–012 | P0 project/promote/webhooks | b81ee8f7 … b7281200 |
| 013 | tracker IDOR | 41ed2b88 |
| 014 | project ACL + sprint tx | bdb1cf31, 484c091a |
| 015 | tms IDOR | c11413e9 |
| 016 | tms import limits | в 22456c7c (вместе с search) |
| 017 | messenger IDOR | 76ff1e9d |
| 018 | ws race + calls authz | eaf8076d, b3367f68 |
| 019 | platform SSRF/outbox/pool | bfb37eec |
| 020 | webhook create contract | 9e892d74 |
| 021 | AI policy + bot authz | 76907211 |
| 022 | search soft-delete/NF-03 | 22456c7c |
| 023 | rbac cache + GDPR erase | 4b011375, 139ba2c1, 4cc54065 |
| 024 | notif/git*/storage/backup/coffee | aa169ddd … b045fedd |
| 025 | Ent unique/O2O | 29ee1991 |
| 026–028 | FE api/calls | 1eb793f4, 15dd222e, 70905c77 |
| 029–035 | FE messenger/board/sprint/trash/a11y/TMS | 2a2d575a … 3ffe8251, d314e6fd |

## P2 реестры

- `W12-REGISTRY-P2-BACKEND.md` — 27 классов / 11 групп → TASK-036…046
- `W12-REGISTRY-P2-FRONTEND.md` — 20 классов / 10 групп → TASK-047…056

## Далее

1. Исполнение P2 (приоритет: fail-closed, dto/error, error≠empty)
2. P3-сводка реестров
3. Stage C mutation (независимые агенты ≠ авторы фиксов)
