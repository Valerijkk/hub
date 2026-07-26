# NFR-11 — Soft-delete vs Trash API classification

**Date measured:** 2026-07-27  
**Scope:** `backend/ent/schema/*.go` + `backend/internal/audit/service.go` (read-only)  
**NFR wording** (`docs/system_specification.md`):

> NFR-11 | Soft delete retention | 30 дней

**Related TZ:** soft delete via trash (**30 days**); hard delete only Admin/GDPR (`docs/tz_constraints.md`).

**Decision status:** owner options only — **no recommendation chosen** in this document.

---

## 1. Measurement (NOW)

| Metric | Value | Notes |
|---|---:|---|
| Entity schemas (`*.go` minus `helpers.go`) | **82** | same universe as NFR-24 D1 |
| Schemas with `field.Time("deleted_at")` | **15** | see §2 |
| Trash API entity types | **4** | `issue`, `test_case`, `file`, `project` |
| Trash retention constant | **30** days | `TrashRetentionDays` in `audit/service.go` |
| Prior audit (`00-PLANNING-AUDIT.md` §4.2) | 15 / 83, API 4 types | **15 deleted_at confirmed**; universe clarified to 82 entities |

### How measured

```powershell
# 2026-07-27, PowerShell, backend/ent/schema
Get-ChildItem *.go | Where-Object Name -ne 'helpers.go' |
  Where-Object { Select-String -Path $_.FullName -Pattern 'field\.Time\("deleted_at"\)' -Quiet }
# Trash types: backend/internal/audit/service.go ListTrash / RestoreItem / SoftDelete switches
```

---

## 2. Schemas with `deleted_at` (15)

| Schema | Ent type | In Trash API? | API type string |
|---|---|:---:|---|
| crmcontact.go | CrmContact | **no** | — |
| crmdeal.go | CrmDeal | **no** | — |
| filemetadata.go | FileMetadata | **yes** | `file` |
| hubbase.go | HubBase | **no** | — |
| hubcolumn.go | HubColumn | **no** | — |
| hubrecord.go | HubRecord | **no** | — |
| hubtable.go | HubTable | **no** | — |
| invoice.go | Invoice | **no** | — |
| issue.go | Issue | **yes** | `issue` |
| journalentry.go | JournalEntry | **no** | — |
| ledgeraccount.go | LedgerAccount | **no** | — |
| marketplacelisting.go | MarketplaceListing | **no** | — |
| newspost.go | NewsPost | **no** | — |
| project.go | Project | **yes** | `project` |
| testcase.go | TestCase | **yes** | `test_case` |

**Summary:** 15 soft-delete columns · **4** restoreable via unified Trash API · **11** have `deleted_at` but no Trash list/restore/purge path in audit service.

---

## 3. Trash API (code truth)

**Routes** (`backend/internal/audit/handler.go`):

- `GET /api/admin/trash`
- `POST /api/admin/trash/:type/:id/restore`

**Supported `type` values** (`backend/internal/audit/service.go`, switches in `ListTrash` / `RestoreItem` / `SoftDelete` / `PurgeExpired`):

| Type | Schema | Soft delete | Restore | Purge after 30d |
|---|---|:---:|:---:|:---:|
| `issue` | Issue | yes | yes | yes |
| `test_case` | TestCase | yes | yes | yes |
| `file` | FileMetadata | yes | yes | yes |
| `project` | Project | yes | yes | yes |

Unknown types: `ListTrash` returns empty; `RestoreItem` / `SoftDelete` return error `"неизвестный тип сущности"`.

Retention: `TrashRetentionDays = 30` (matches NFR-11 retention number).

Wiki docs (`docs/wiki/API.md`): admin trash endpoints documented; **type enum not fully listed** (gap for docs-sync after owner decision).

---

## 4. FR / product promises involving trash or soft-delete

Cross-check against `docs/system_specification.md` (+ supporting `docs/vision_and_scope.md` / `docs/wiki/API.md`).

### 4.1 Explicit FR → trash / restore

| ID | Promise | Schema / module today | `deleted_at` | In Trash API | Gap |
|---|---|---|:---:|:---:|---|
| **FR-INTR-07** | News post delete → корзина 30д + audit | `NewsPost` | **yes** | **no** | Soft-deleted but **not restorable** via Trash |
| **FR-INTR-08** | Restore post from trash returns post + poll + vote counts; restore via `/admin/trash` | `NewsPost` (+ Poll/PollVote) | yes (post) | **no** | Spec names unified Trash API; type missing |
| **FR-CAL-05** | Delete series → all occurrences to корзина 30д | `ScheduledCall` | **no** | **no** | No soft-delete column; no trash type |
| **FR-WIKI-03** | Page delete → subtree to корзина 30д; restore returns subtree | *(no wiki schemas)* | n/a | **no** | Module not implemented (planning debt) |
| **FR-WIKI-07/09** | History after trash restore; search excludes trash | *(no wiki)* | n/a | n/a | Blocked on wiki module |

### 4.2 Broader NFR / US / API promises

| Source | Promise | Code reality |
|---|---|---|
| NFR-11 | Soft delete retention 30 days | Retention constant = 30 for the **4** trash types only |
| TZ / `60-tz-constraints.mdc` | Soft delete via trash 30d; hard delete Admin/GDPR | Only 4 types in trash; 11 other `deleted_at` entities lack restore/purge API |
| US-S03 (`vision_and_scope.md`) | Admin: audit log + trash 30d, restore | Implemented for 4 types |
| Data model note (`system_specification.md` ~L109) | Soft delete: `deleted_at` **or** TrashItem | Mix: column-based soft delete; no separate `TrashItem` Ent schema |
| Wiki API: `DELETE /projects/:projectId` → trash; restore | Project path exists | Aligned with `project` trash type |
| Intranet DELETE news → trash + restore via `/admin/trash` | Documented in system_specification | **Broken vs code** (FR-INTR-08) |
| Calendar DELETE event → trash 30д | Documented | **Not implemented** (FR-CAL-05) |
| Wiki DELETE page → trash | Documented as planned | **No module** |

### 4.3 Soft-deleted in DB but not promised as Trash API types

These have `deleted_at` without an FR that explicitly names unified Trash restore (may still be product soft-delete):

| Schema | Likely intent | Trash gap |
|---|---|---|
| CrmContact, CrmDeal | CRM soft delete | No admin restore/list |
| HubBase, HubTable, HubColumn, HubRecord | Bases soft delete | No admin restore/list |
| Invoice, JournalEntry, LedgerAccount | Accounting soft delete | No admin restore/list; legal-hold may interact (GZ-SEC) |
| MarketplaceListing | Listing soft delete | No admin restore/list |
| NewsPost | Explicitly promised in FR-INTR-07/08 | Soft delete **without** restore API |

---

## 5. Gap matrix (schemas with `deleted_at` vs Trash)

```
deleted_at schemas (15)
├── In Trash API (4): issue, test_case, file, project
└── Not in Trash API (11):
    ├── FR-promised restore: newspost  ← FR-INTR-07/08
    └── Soft-delete only / no FR trash: crm*, hub*, invoice, journalentry,
        ledgeraccount, marketplacelisting
```

```
FR-promised trash without deleted_at / schema:
├── FR-CAL-05 → ScheduledCall (no deleted_at)
└── FR-WIKI-03 → wiki_page (no schema)
```

---

## 6. Owner options (DO NOT choose here)

### Option 1 — Expand trash

Grow Trash API (and purge cron) to cover every entity that must honor NFR-11 / FR trash promises — at minimum:

1. **Must-fix for open FR:** `newspost` (+ cascade poll/votes per FR-INTR-08); `scheduledcall` series (add `deleted_at` + type); wiki when/if built.
2. **Decide for soft-deleted orphans:** CRM, Bases, Accounting, Marketplace — either register trash types + purge, or document domain-local soft delete without admin Trash UI.

| Pros | Cons |
|---|---|
| Matches NFR-11 / TZ / US-S03 / FR-INTR-08 / FR-CAL-05 | Multi-module work (audit repo + handlers + purge + FE admin) |
| Ends “data soft-deleted but unrestorable” failure mode for NewsPost | Accounting soft-delete may conflict with legal-hold / immutability (GZ-SEC) |
| Single admin UX for restore | Larger purge surface and test matrix |
| Makes “30 days” meaningful beyond 4 types | Calendar needs schema migration (`deleted_at` on ScheduledCall) |

### Option 2 — Narrow NFR-11 wording

Rewrite NFR-11 (and TZ) so soft-delete retention / unified Trash applies only to an **explicit list** (today: `issue`, `test_case`, `file`, `project`), and:

- amend FR-INTR-07/08 / FR-CAL-05 / FR-WIKI-03 to match (or move them to a future milestone),
- clarify that other `deleted_at` columns are domain soft-delete **without** Trash API unless listed.

| Pros | Cons |
|---|---|
| Docs match the 4-type implementation | Product loses promised restore for news / calendar / wiki unless FR also narrowed |
| Smaller near-term code scope under freeze | TZ “soft delete via trash” becomes weaker |
| Forces honest FR edits instead of silent gap | Users may still expect US-S03 “restore anything deleted” |
| Clear gate: new trash type = explicit NFR/FR change | 11 soft-deleted entities still need a stated lifecycle (purge? never? GDPR only?) |

---

## 7. What this document does *not* do

- Does **not** pick Option 1 or 2.
- Does **not** change audit service, schemas, or NFR/FR text.
- Does **not** invent Trash types for CRM/Bases/Accounting without owner scope.
- Wiki remains a separate scope decision (`01-PLANNING-FIXES.md` B5).

---

## 8. Traceability

| Source | Use |
|---|---|
| `docs/system_specification.md` NFR-11, FR-INTR-07/08, FR-CAL-05, FR-WIKI-03 | Promises |
| `docs/vision_and_scope.md` US-S03, US-INTR-03, calendar/wiki edges | Product intent |
| `docs/wiki/API.md` admin trash + project restore | API surface |
| `backend/internal/audit/service.go` | Trash type switch + 30d retention |
| `backend/ent/schema/*.go` | `deleted_at` inventory, 2026-07-27 |
| `.planning/planning-audit-2026-07/00-PLANNING-AUDIT.md` §4.2–4.3 | Prior gap write-up |
| `.planning/planning-audit-2026-07/01-PLANNING-FIXES.md` B1/B2/B4 | Fix plan hooks |
