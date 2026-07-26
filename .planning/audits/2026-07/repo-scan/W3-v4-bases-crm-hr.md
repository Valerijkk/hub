# W3 — v4 bases / crm / hr (owner-scope · layering · IDOR)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Агент:** be-ver + security (Grok 4.5) · **Дата:** 2026-07-23 · **Режим:** readonly repo-scan
> **Scope:** `backend/internal/bases/`, `backend/internal/crm/`, `backend/internal/hr/`
> **Фокус:** owner-scope (аналог NF-03), layering, IDOR
> **Код не менялся.**

## Вердикт (кратко)

| Модуль | Owner-scope / IDOR | Layering | Тесты denial |
|--------|--------------------|----------|--------------|
| **bases** | **PASS** — `authorizeBase` на всех Get/mutate; List по owner | OK | `testframework/bases/rbac_test.go` |
| **crm** | **FAIL** — org-wide list + **любой JWT** Update/Delete contact/deal без owner/роли | OK | нет foreign-owner suite |
| **hr** | **PASS** (почти) — my-leave + admin approve; cancel scoped requester | OK | foreign cancel / non-admin deny |

- **Регрессия в этом прогоне:** не запускалась (scan-only; Ask/readonly). Канон: `go test -tags hubtest -race ./testframework/{bases,crm,hr}/...`
- **verified:** `false` — блокируют CRITICAL/HIGH по **crm** (+ HIGH optimistic lock в bases/crm).

---

## Summary

| Severity | Count |
|----------|------:|
| CRITICAL | 4 |
| HIGH     | 3 |
| MEDIUM   | 4 |
| LOW      | 2 |

**OK (не findings):**
- **bases:** `authorizeBase` (`service.go:45-56`) до Get/Update/Delete/table/column/record; `ListBases` + `HasOwnerWith` (`repository.go:56-69`); `GetTable` привязан к `baseID` (`repository.go:112-128`); `normalizeCells` allowlist (`service.go:287-301`); soft-delete UTC; `*ent.Client` только в `EntRepository`.
- **crm:** layering OK; `IsProjectMember` при привязке `project_id` (`service.go:226-234`); soft-delete UTC; stage allowlist.
- **hr:** `ListLeaveByRequester`; `requireAdmin` на pending/approve/reject; `CancelLeave` + `HasRequesterWith` в repo; `VersionEQ` на approve/reject/cancel; тесты `CancelLeave_ForbiddenForOtherUser`, `ApproveLeave_ForbiddenForMember`.

---

## Findings

| Sev | Файл:строка | Evidence | Fix |
|-----|-------------|----------|-----|
| **CRITICAL** | `crm/service.go:80-123`; `crm/handler.go:59-84`; `crm/repository.go:66-72,81-112` | `UpdateContact` / `DeleteContact` принимают только `id` — **нет** `userID`, нет проверки `created_by`, нет RBAC-роли. Любой аутентифицированный член org может изменить/soft-delete **любой** контакт по UUID (IDOR). `ListContacts`/`GetContact` без scope — org-wide (wiki не обещает owner-scope, но мутации без authz = дыра целостности). | `authorizeContact` (created_by **или** system admin/founder / будущий CRM-role) **до** Update/Delete; denial-suite как `bases/rbac_test.go`. |
| **CRITICAL** | `crm/service.go:172-214`; `crm/handler.go:110-138`; `crm/repository.go:140-200` | `UpdateDeal` / `DeleteDeal`: edge `owner` есть в схеме (`ent/schema/crmdeal.go:55-58`) и пишется при Create, но **никогда не сверяется** с actor. `DeleteDeal(ctx, id)` вообще без userID. Любой JWT → правка stage/суммы/soft-delete чужой сделки. | `authorizeDeal`: `deal.OwnerID == userID` (или admin); Delete/Update только после authorize. |
| **CRITICAL** | `crm/service.go:172-204`, `247-257`; `handler.go:110-127` | Следствие: любой member может `PUT /crm/deals/:id` с `stage=won` + `amount_cents` → `tryCreateInvoice` создаёт draft invoice по **чужой** сделке (фин. побочный эффект). | Authorize owner/admin **до** Update; invoice hook только после успешной authz. |
| **CRITICAL** | `crm/handler.go:21-32`; `wiring_modules.go:43` | Маршруты `/api/crm/*` — только `authMw`, без `rbac.RequirePermission` и без service-level owner check. В отличие от tracker/tms. Org-wide CRM-lite допустим для **чтения**, но write без роли = полный wipe CRM любым member. | Минимум: service authorize на mutate; опционально RBAC `crm.edit` / admin-only delete. |
| **HIGH** | `bases/service.go:105-107,251-252`; `bases/repository.go:83-93,207-214` | Optimistic lock TZ **частично**: конфликт только если `req.Version > 0`; при `version=0`/omit — тихая перезапись. В repo **нет** `Where(VersionEQ)` — только check-then-act + `AddVersion(1)` (TOCTOU). | `Version` required; `Update…Where(VersionEQ)`; 0 rows → `ErrVersionConflict`. Эталон tracker. |
| **HIGH** | `crm/service.go:88-90,180-181`; `crm/repository.go:81-103,163-191` | Тот же класс: `Version > 0` в service, repo без `VersionEQ` на contact/deal Update. | Как выше для CRM Update. |
| **HIGH** | `testframework/crm/` | Есть тесты `ErrProjectForbidden` при чужом `project_id`; **нет** suite «user B мутирует contact/deal user A → 403». Happy-path покрывает wipe чужих данных как «успех». | Denial-suite: create as A, mutate as B → 403/404. |
| **MEDIUM** | `crm/service.go:43-53,125-147`; `repository.go:66-71,140-151` | List contacts/deals — полный org dump без фильтра. Для 1-org CRM-lite может быть by design (wiki не пишет owner-scope), но усиливает blast radius CRITICAL-мутаций и утечку PII (notes/email/phone) всем JWT. | Зафиксировать ADR «org-shared read»; иначе фильтр по created_by/owner или role. |
| **MEDIUM** | `hr/service.go:106-120,145-159,178-191` | `Approve`/`Reject`/`Cancel` при `version==0` подставляют текущую версию — клиентский optimistic lock обходится (гонка двух админов всё ещё ловится `VersionEQ` в repo). | Require `version` в body; 0 → 400. |
| **MEDIUM** | `hr/handler.go:144-149` vs `service.go:187-188` | `ErrForbidden` всегда мапится в `"admin access required"`. Non-owner `CancelLeave(..., 0)` тоже Forbidden → вводит в заблуждение (не admin-кейс). | Разделить `ErrForbidden` / `ErrNotOwner` или разный HTTP message. |
| **MEDIUM** | `bases/service.go:126-130,284` | `DeleteBase`/`DeleteRecord`: `return s.repo.SoftDelete*` без `fmt.Errorf("%w")` — теряется wrap на границе. | Wrap + map NotFound. |
| **LOW** | `bases/repository.go:131-151,169-194` | `UpdateTable`/`SoftDeleteTable`/`UpdateColumn`/`SoftDeleteColumn` в repo **без** HTTP-маршрутов — мёртвый API surface / будущий риск, если вызовут без authorize. | Удалить или провести через `authorizeBase` + handler. |
| **LOW** | `crm/handler.go:141-156` | `parseContactID`/`parseDealID` — ручной `uuid.Parse` вместо `httputil.ParseUUIDParam` (bases-канон). | Унифицировать. |

---

## Чек-лист be-ver (сводка)

| Пункт | bases | crm | hr |
|-------|-------|-----|-----|
| Layering `*ent.Client` вне repository | OK (`repository.go:39-40`) | OK (`repository.go:35-36`) | OK (`repository.go:29-30`) |
| Handler без ent | OK | OK | OK |
| Owner/member-scope List | OK | **FAIL** (org-all; design?) | OK (my) / admin pending |
| Owner-scope Get + mutate | OK | **FAIL** CRITICAL | OK cancel; admin review |
| Authz **до** побочного эффекта | OK | **FAIL** | OK |
| Allowlist user JSON | OK `normalizeCells` | N/A | N/A |
| Soft delete UTC | OK | OK | N/A (status cancel) |
| Optimistic `VersionEQ` в DB | **FAIL** HIGH | **FAIL** HIGH | OK approve/reject/cancel |
| `panic` / SSRF / DefaultClient | OK | OK | OK |
| Denial-тесты IDOR | OK `rbac_test.go` | **FAIL** | OK partial |

---

## Воспроизведение CRITICAL (crm)

1. User A создаёт контакт `C` и сделку `D` (`POST /api/crm/contacts`, `POST /api/crm/deals`).
2. User B (обычный member, JWT) вызывает:
   - `PUT /api/crm/contacts/{C}` — 200, поля изменены;
   - `DELETE /api/crm/deals/{D}` — 204;
   - `PUT /api/crm/deals/{D}` `{"stage":"won","amount_cents":10000,"version":N}` — draft invoice (если hook включён).
3. Ожидание (эталон bases): 403/404 для non-owner (или явная CRM-роль). Факт: успех.

---

## bases — эталон (для сравнения)

```45:56:backend/internal/bases/service.go
func (s *Service) authorizeBase(ctx context.Context, baseID, userID uuid.UUID) (*ent.HubBase, error) {
	b, err := s.repo.GetBase(ctx, baseID)
	// ...
	if b.Edges.Owner == nil || b.Edges.Owner.ID != userID {
		return nil, ErrForbidden
	}
	return b, nil
}
```

CRM должен получить аналог `authorizeContact` / `authorizeDeal` (или org-role RBAC) — сейчас edge `owner`/`created_by` только для метаданных ответа.

---

## Verification

- Воспроизведено: по коду (handler→service→repo); crm CRITICAL сценарий выше.
- Layering: OK во всех трёх модулях.
- Security: bases OK; crm FAIL; hr OK с MEDIUM по version=0 / error text.
- Регрессия: не прогонялась (repo-scan).
- Вердикт: **verified: false** — 2026-07-23 01:10 UTC+3
