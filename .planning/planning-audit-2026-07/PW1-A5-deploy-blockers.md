# PW1-A5 — `deploy_blockers` for STATE.md

> Audit task A5 · documentation only · 2026-07-27  
> Do **not** treat phase `percent: 100` as “prod deploy unlocked”.

## Sources

| Source | Finding |
|--------|---------|
| `.planning/phases/63-production-gate/VERIFICATION.md` | Verdict: ✅ PASS (code gate) · ⚠️ **deploy blocked on commit + PG-03 ops**. PG-03 row: runbook exists; **manual ops pending**. |
| `.planning/phases/67-production-gate/VERIFICATION.md` | Same verdict. Sign-off: code gate-ready; deploy needs user-approved squash commit **and PG-03** if secrets were ever in git history. |
| `.planning/phases/{63,67}-production-gate/PLAN.md` | Non-blocking (manual ops): **PG-03** + COMMIT/PR user approval. |
| `.planning/production/PG-03-SECRET-ROTATION.md` | Full secret-rotation checklist (FA-06 Wave C). |
| `.planning/production/FA-06-EXECUTION-LOG.md` | Phase 0 git scan ✅; **Phase 1–3 not executed**; CTO sign-off blank. |
| `docs/runbooks/fa06-secret-rotation.md` | Owner-only Resend + Telegram rotation (INF-03); GZ-SEC-01/02. |

## What PG-03 / FA-06 are

- **PG-03** = Production Gate item **#03**: **secret rotation** before prod cutover (JWT, Resend, Telegram, DBs, LiveKit, S3, …). Canonical checklist: `.planning/production/PG-03-SECRET-ROTATION.md`.
- **FA-06** = Wave C final-audit task that **executes PG-03**. Log: `.planning/production/FA-06-EXECUTION-LOG.md`. Owner runbook (P1 Resend + Telegram): `docs/runbooks/fa06-secret-rotation.md`.
- **Why still open:** Phase 0 found INF-03 history exposure (`RESEND_API_KEY`, `TELEGRAM_BOT_TOKEN` before `859ecef`) and SEC-03 LiveKit history. Working tree is clean (mocks only). **Phase 1–3 rotation + CTO sign-off not done** → sole remaining **ops prod blocker** (does not gate feature development / July Rebuild planning).
- **Not a code gate failure:** Phases 63/67 code gates PASSED; deploy is operationally blocked.

## Paste into `.planning/STATE.md` frontmatter

Insert under existing `progress:` (or after `last_activity:`). **A4 owns rewriting `progress` / percent fields — do not conflate.**

```yaml
deploy_blockers:
  - id: PG-03
    aka: FA-06
    title: Secret rotation (Resend/Telegram P1 + inventory)
    status: open
    blocks: production_cutover
    does_not_block: feature_dev_july_rebuild_planning
    phase_refs:
      - "63-production-gate (VERIFICATION: code PASS, deploy blocked)"
      - "67-production-gate (VERIFICATION: code PASS, deploy blocked)"
    docs:
      - .planning/production/PG-03-SECRET-ROTATION.md
      - .planning/production/FA-06-EXECUTION-LOG.md
      - docs/runbooks/fa06-secret-rotation.md
    progress:
      phase_0_git_scan: done
      phase_1_rotate: pending
      phase_2_smoke: pending
      phase_3_cto_signoff: pending
    priority_secrets:
      - RESEND_API_KEY
      - TELEGRAM_BOT_TOKEN
    note: >-
      Единственный prod-ops блокер. SSH deploy + tag v2.5.0 — после FA-06
      Phase 1–3 + CTO sign-off. Code gates 63/67 PASS.
```

### Optional human-readable list (body § «Блокеры / риски»)

Replace the opaque row `PG-03 not executed | FA-06 manual ops…` with:

```markdown
| ID | Блокер | Статус | Ссылки |
|----|--------|--------|--------|
| **PG-03 / FA-06** | Secret rotation (P1: Resend + Telegram; full inventory in PG-03) | ⏳ Phase 0 ✅ · Phase 1–3 **manual pending** | `PG-03-SECRET-ROTATION.md` · `FA-06-EXECUTION-LOG.md` · `docs/runbooks/fa06-secret-rotation.md` |
```

## Acceptance for A4/STATE merge

- [ ] Frontmatter has `deploy_blockers:` as above (or equivalent YAML).
- [ ] `percent` / `phases_percent: 100` does **not** imply deploy unlocked.
- [ ] No secret values committed; only env **names**.
