# Milestone: v4 Hub One Super-app

> **Shipped:** 2026-06-25 (code on `feat/v2.5-trust`, uncommitted)  
> **Phases:** 39–57 (19 phases)  
> **Vision:** `.planning/decisions/specs/PLATFORM-VISION-v4.md`

---

## Executive Summary

Hub evolved from a dev-suite (tracker + messenger + TMS + calls) into a **corporate super-app** comparable to Tinkoff Mighty / Bitrix24 lite: intranet news, org directory, unified calendar, CRM-lite, HR leave, kudos, marketplace, public status page, Random Coffee, 1C connector, Lark-style no-code bases, and **accounting MVP with invoices + P&L**.

All v4 backend modules meet **≥50% test coverage**. Full qa-gate green: release build, race tests, Biome 0 errors, Vitest 80/80, v4 E2E in CI.

---

## Phase Map

| Phase | Module | Key deliverable |
|-------|--------|-----------------|
| 39 | Intranet news | `NewsPost` + `/intranet` feed |
| 40 | Org directory | MOD-02 + Ctrl+K people search |
| 41 | Calendar | MOD-04 `/calendar` + iCal export |
| 42 | CRM-lite | Contacts + deals kanban |
| 43 | HR-lite | Leave requests + calendar sync |
| 44 | Kudos | Thanks feed + notifications |
| 45 | Marketplace | Classifieds + mod queue |
| 46 | Status page | Public `/status` |
| 47 | Random Coffee | Intranet matching |
| 48 | 1C connector | Leave webhook + admin settings |
| 49 | Lark Base | No-code tables `/bases` |
| 50 | v4 gate | qa-gate + Vitest smoke |
| 51 | Accounting MVP | Ledger + trial balance |
| 52 | Accounting v4.5 | Invoices + P&L + CRM won hook |
| 53 | v4 module gate | CRM↔invoice tests + E2E |
| 54 | Coverage uplift | bases 61% + marketplace 69% |
| 55 | onec + intranet | 72% + 61% coverage |
| 56 | v4 final gate | CRM 73% + qa-gate |
| 57 | CI handoff | v4 E2E in CI + admin auth fix |
| 58 | Archive | Milestone SUMMARY + production gate |

---

## Architecture Additions

```
backend/internal/
├── intranet/      # news, kudos, org directory, random coffee
├── calendar/      # unified events + iCal
├── crm/           # contacts, deals (+ invoice hook on won)
├── hr/            # leave requests
├── marketplace/   # classifieds
├── status/        # public health probes
├── bases/         # Lark-style tables
├── accounting/    # ledger, journal, invoices, income statement
└── integrations/onec/

frontend/src/modules/
├── intranet/  calendar/  crm/  hr/  marketplace/
├── status/  bases/  accounting/
```

**Ent schemas added:** `NewsPost`, CRM entities, HR leave, marketplace listing, base tables, ledger accounts, journal entries, `Invoice`.

**Wiring:** `main.go` registers all v4 routes; `crmSvc.SetDealInvoiceCreator(acctSvc)` for won-deal → draft invoice.

---

## Coverage (final)

| Package | Coverage |
|---------|----------|
| status | 100.0% |
| crm | 72.8% |
| integrations/onec | 72.0% |
| marketplace | 69.2% |
| calendar | 65.6% |
| accounting | 65.0% |
| intranet | 60.8% |
| bases | 60.8% |
| hr | 53.1% |

Overall `internal/`: **53.6%** · Security-critical 6/6 ≥80%.

---

## Frontend & E2E

- **Vitest:** 32 files, 80 tests (v4 module smokes + accounting v4.5)
- **Playwright:** `e2e/v4-superapp.spec.ts` in CI `e2e-full` (13 specs total)
- **Routes:** `/intranet`, `/calendar`, `/crm`, `/hr`, `/marketplace`, `/status`, `/bases`, `/accounting`

---

## Known Limitations

| Item | Notes |
|------|-------|
| Single-tenant | One org per instance (by design) |
| Accounting | MVP chart of accounts; not full 1C replacement |
| Bases | No formula engine (v4.5+ backlog) |
| Deploy | Branch uncommitted; PG-03 secret rotation pending |
| Schema | Auto-migrate via `Schema.Create` at startup |

---

## Success Criteria — Met

1. ✅ All MOD-01..08 + V4-03..06 modules implemented
2. ✅ CRM won → draft invoice automation
3. ✅ v4 module coverage ≥50%
4. ✅ qa-gate green (build, race, lint, vitest)
5. ✅ v4 E2E in GitHub Actions

---

## Next Milestone

**v4.5+ (2027+):** Advanced accounting, Lark Base formulas, deeper 1C sync.  
**Immediate:** Squash commit `feat/v2.5-trust` → PR → tag `v4.0.0`.
