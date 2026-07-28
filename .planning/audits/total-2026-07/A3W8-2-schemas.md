# A3 W8 агент 2 — ent/schema part 2 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W8-2-schemas.md](A2W8-2-schemas.md) · Stage A: [A8-2-schemas.md](A8-2-schemas.md)  
> Assignment: [W8-ASSIGNMENTS.md](W8-ASSIGNMENTS.md) § A2 · Периметр: **17** · Прочитано: **17/17** · Пропущено: **0**  
> Фокус: UUID / version / soft-delete / indexes / OnDelete  
> Remediation: TASK-095 (VersionMixin) · TASK-097 (soft-delete CFD/integrations) · TASK-107 (indexes/links/secrets)

## Сверка с A2 (A2W8-2-*)

| A2 ID | Tag | Статус | Evidence в HEAD |
|-------|-----|--------|-----------------|
| A2W8-2-001 | FIXED-OK | закрыт | `comment.go:44-49` `VersionMixin{}` (TASK-095) |
| A2W8-2-002 | FIXED-OK | закрыт | `dashboardlayout.go:41-46` `VersionMixin{}` |
| A2W8-2-003 | FIXED-OK | закрыт | `component.go:36-41` + `environment.go:42-47` `VersionMixin{}`; soft-delete уже был (TASK-039) |
| A2W8-2-004 | FIXED-OK | закрыт | `customfielddefinition.go:37-50` `updated_at`+`deleted_at`+`VersionMixin` (TASK-097 + TASK-095) |
| A2W8-2-005 | FIXED-OK | закрыт | `crmcontact.go:70` `index.Fields("email")`; `crmdeal.go:79` `Edges("owner").Fields("deleted_at")` (TASK-107) |
| A2W8-2-006 | RESIDUAL | частично | **version** → ADR-ENT-VERSION-MIXIN exclusion `filemetadata` (docs+PO) FIXED; **хвост:** `entity_id` String, `size` без Min, `uploader` Optional — см. A3W8-2-001 |
| A2W8-2-007 | FIXED-OK | закрыт | `gitlabintegration.go:33-37` `Sensitive().NotEmpty()` (не `Optional`) |
| A2W8-2-008 | FIXED-OK | закрыт | `githublink.go:43-45,74` / `gitlablink.go:46-48,77` `version` + unique `(integration, external_id, link_type)` |
| A2W8-2-009 | FIXED-OK | закрыт | `hubbase.go:59` owner+`deleted_at`; `hubcolumn.go:65-66` unique `(table,name)` + table+`deleted_at`; `hubrecord.go:58` table+`deleted_at` |

**REGRESS:** 0 · **FIXED-OK (A2 actives):** 8 (+ version-часть A2W8-2-006) · **RESIDUAL (A2):** 1 (хвост FileMetadata) · **NEW:** 5

### Cascade FIXED-OK (Stage A / A2 уже закрытые классы в периметре)

| Класс | Tag | Evidence |
|-------|-----|----------|
| Component/Environment soft-delete + unique `(project,name)` | FIXED-OK | `component.go:27-30,59-62`; `environment.go:33-36,65-68` |
| GitHub `webhook_secret` Sensitive+NotEmpty | FIXED-OK | `githubintegration.go:33-36` |
| Comment `mentions` `[]uuid.UUID` | FIXED-OK | `comment.go:29` |
| Mojibake schema comments | FIXED-OK | UTF-8 «Рёбра»/comments; без mojibake |
| UUID PK + `nowUTC` | FIXED-OK | все 16 entity-схем + `helpers.go:14` |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/coffeeround.go` | 47 | 1 | NEW OnDelete pairs |
| `backend/ent/schema/comment.go` | 75 | 1 | VersionMixin OK; NEW OnDelete reactions |
| `backend/ent/schema/component.go` | 65 | 0 | soft+version+unique FIXED-OK |
| `backend/ent/schema/crmcontact.go` | 75 | 0 | email index + VersionMixin FIXED-OK |
| `backend/ent/schema/crmdeal.go` | 82 | 0 | owner+deleted_at FIXED-OK |
| `backend/ent/schema/customfielddefinition.go` | 74 | 0 | soft+version FIXED-OK |
| `backend/ent/schema/dashboardlayout.go` | 67 | 0 | VersionMixin FIXED-OK |
| `backend/ent/schema/environment.go` | 71 | 0 | soft+version+unique FIXED-OK |
| `backend/ent/schema/filemetadata.go` | 70 | 1 | RESIDUAL entity_id/size/uploader |
| `backend/ent/schema/githubintegration.go` | 79 | 2 | webhook OK; OnDelete links + access_token |
| `backend/ent/schema/githublink.go` | 77 | 0 | version+unique FIXED-OK |
| `backend/ent/schema/gitlabintegration.go` | 76 | 2 | webhook OK; OnDelete links + access_token |
| `backend/ent/schema/gitlablink.go` | 80 | 0 | version+unique FIXED-OK |
| `backend/ent/schema/helpers.go` | 14 | 0 | чисто UTC/`keyRegexp` |
| `backend/ent/schema/hubbase.go` | 62 | 1 | indexes OK; NEW OnDelete tables |
| `backend/ent/schema/hubcolumn.go` | 69 | 0 | unique+composite FIXED-OK |
| `backend/ent/schema/hubrecord.go` | 61 | 0 | table+deleted_at FIXED-OK |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W8-2-001 | RESIDUAL (A2W8-2-006 / W8-A2-010·011) | P2 | `filemetadata.go:26-35,50-54` | `entity_id` = `String`; `size` без `Min`/`Positive`; `uploader` не `Required`. `version` **не** дефект: exclusion ADR-ENT-VERSION-MIXIN §4 | TZ UUID / defense-in-depth: сервис валидирует size (`storage/service.go:92-94`) и всегда ставит uploader, но схема принимает отрицательный size и NULL uploader. `entity_id` нельзя слепо → UUID: storage допускает `"me"` / `"temp"` (`service.go:255-271`) | (1) `size` → `NonNegative()`/`Min(0)`; (2) `uploader` Required **или** ADR system-upload path; (3) `entity_id`: allowlist-валидация / typed polymorphic (UUID **или** sentinel), не голый String без контракта |
| A3W8-2-002 | NEW | P2 | `coffeeround.go:37-38` | `edge.To("pairs")` без `entsql.OnDelete(Cascade)` | ADR-ENT-ONDELETE-CASCADE owned_child; hard-purge round → FK trap / orphans `CoffeePair` | `Annotations(entsql.OnDelete(entsql.Cascade))` на `pairs` (+ migrate FK) |
| A3W8-2-003 | NEW | P2 | `hubbase.go:51-52` | `edge.To("tables")` без Cascade | owned HubTable/колонки/записи при PermanentDelete base — FK trap (симметрия Project→children) | Cascade на `tables`; каскад table→columns/records — слот A3 / HubTable |
| A3W8-2-004 | NEW | P2 | `githubintegration.go:65`; `gitlabintegration.go:62` | `To("github_links"/"gitlab_links")` без Cascade | Project Cascade удаляет integration; links Required→integration → FK trap, если Issue-links ещё живы / порядок DELETE | Cascade Integration→links (Issue→links уже Cascade в `issue.go`) |
| A3W8-2-005 | NEW | P3 | `comment.go:63` | `To("reactions")` без `entsql.OnDelete(Cascade)` | hard-delete comment → orphan reactions / FK trap | Cascade на `reactions` |
| A3W8-2-006 | NEW (хвост W8-A2-012) | P3 | `githubintegration.go:30-32`; `gitlabintegration.go:30-32` | `access_token` Sensitive без `NotEmpty` (webhook_secret уже Required) | схема допускает пустой PAT; асимметрия с webhook_secret | `NotEmpty()` **или** ADR «token optional until first sync» |

## Чисто (подтверждено)

- **UUID PK** `Default(uuid.New).Immutable()` — все entity-схемы периметра.
- **VersionMixin** на mutable: Comment, Component, Crm*, CustomFieldDefinition, DashboardLayout, Environment, GitHub/GitLab Integration, Hub*.
- **GitHub/GitLab Link** — inline `version` Default(1) + unique external (не дублировать Mixin).
- **CoffeeRound** — append-only (без `updated_at`); version exclusion ADR OK; индекс `created_at`.
- **Soft-delete** — Comment/Component/Environment/Crm*/CFD/FileMetadata/Integrations/Hub* имеют `deleted_at` (+ индексы где нужно).
- **helpers.go** — только `nowUTC()` / `keyRegexp`; UTC-канон.
- **OnDelete на родителях вне слота:** Project Cascade → components/environments/custom_fields/integrations/files; Issue Cascade → comments/github_links/gitlab_links — без регресса TASK-096 на этих рёбрах.

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 2 |
| **Total open** | **6** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 actives 001–005, 007–009) | 8 |
| RESIDUAL | 1 (`A3W8-2-001` ← A2W8-2-006 хвост) |
| NEW | 5 (`A3W8-2-002`…`006`) |
| REGRESS | 0 |

A2 baseline open classes: 9 → **FIXED-OK 8** · RESIDUAL 1 · NEW 5 · REGRESS 0 · open **6** (P2=4 · P3=2).

**Топ ID:** A3W8-2-001 (FileMetadata) · A3W8-2-002/003/004 (OnDelete owned_child).
