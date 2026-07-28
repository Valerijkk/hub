# W8 агент A2 — ent/schema part 2

> Фокус: UUID+version · soft-delete · missing indexes · insecure defaults  
> Read-only · TASK-008 · 2026-07-28

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/ent/schema/coffeeround.go | 46 | 0 | чисто |
| backend/ent/schema/comment.go | 68 | 1 | замечания |
| backend/ent/schema/component.go | 51 | 2 | замечания |
| backend/ent/schema/crmcontact.go | 68 | 1 | замечания |
| backend/ent/schema/crmdeal.go | 75 | 1 | замечания |
| backend/ent/schema/customfielddefinition.go | 58 | 1 | замечания |
| backend/ent/schema/dashboardlayout.go | 60 | 1 | замечания |
| backend/ent/schema/environment.go | 56 | 2 | замечания |
| backend/ent/schema/filemetadata.go | 70 | 2 | замечания |
| backend/ent/schema/githubintegration.go | 64 | 2 | замечания |
| backend/ent/schema/githublink.go | 73 | 1 | замечания |
| backend/ent/schema/gitlabintegration.go | 61 | 2 | замечания |
| backend/ent/schema/gitlablink.go | 76 | 1 | замечания |
| backend/ent/schema/helpers.go | 16 | 0 | чисто |
| backend/ent/schema/hubbase.go | 55 | 1 | замечания |
| backend/ent/schema/hubcolumn.go | 61 | 1 | замечания |
| backend/ent/schema/hubrecord.go | 54 | 1 | замечания |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Сводка (counts)

| Категория | Count |
|-----------|------:|
| Файлов прочитано | 17 |
| Файлов с находками | 15 |
| Файлов чисто | 2 |
| Находок всего | **20** |
| P0 | 0 |
| P1 | 4 |
| P2 | 14 |
| P3 | 2 |

### Топ-3

1. **W8-A2-001 / W8-A2-002 (P1)** — `Comment` и `DashboardLayout` изменяемые (`updated_at` / concurrent JSON), но без `version` → last-write-wins против TZ optimistic lock.
2. **W8-A2-003 / W8-A2-004 (P1)** — `Component` и `Environment` без `deleted_at`; TMS hard-delete (`DeleteOneID`) против TZ-корзины 30 дней.
3. **W8-A2-012 / W8-A2-013 (P2→security)** — `webhook_secret` у GitHub/GitLab без `Sensitive()`; у GitLab ещё `Optional()` (пустой секрет на уровне схемы).

### Перепроверка прошлых A8 (grok-2026-07)

| Старый ID | Статус в этом срезе | Комментарий |
|-----------|---------------------|-------------|
| A8-006 Comment soft-delete | **CLOSED** — не дублировать | `deleted_at` + индекс есть |
| A8-004 Comment без `version` | **живо** → W8-A2-001 | soft-delete закрыт, optimistic lock — нет |
| A8-015 FileMetadata soft без version | **живо** → W8-A2-013 | `deleted_at` есть, `version` нет |
| A8-001/002 composite indexes | паттерн не перенесён на Hub* | → W8-A2-015…017 |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W8-A2-001 | P1 | `backend/ent/schema/comment.go:17-42` | Нет `field.Int("version")` при `updated_at` + `UpdateComment` | TZ: optimistic lock в каждой изменяемой схеме. Repo обновляет body без `VersionEQ` → тихая потеря правок при concurrent edit | Добавить `version` Default(1); Update с `VersionEQ` + `AddVersion(1)` |
| W8-A2-002 | P1 | `backend/ent/schema/dashboardlayout.go:17-39` | Нет `version` при mutable `items` JSON + `updated_at` | Concurrent SaveLayout last-write-wins; TZ optimistic lock | `version` + проверка в SaveDashboardLayout |
| W8-A2-003 | P1 | `backend/ent/schema/component.go:17-30` | Нет `deleted_at`; delete = hard (`tms/ent_repository.go` `DeleteOneID`) | TZ soft-delete / корзина 30 дней для user-visible сущности | `deleted_at` + soft-delete в repo; фильтр `DeletedAtIsNil` |
| W8-A2-004 | P1 | `backend/ent/schema/environment.go:17-36` | Нет `deleted_at`; `DeleteEnvironment` = hard delete; есть `UpdateEnvironment` без `version` | Soft-delete gap + mutable entity без optimistic lock | `deleted_at` + `version` + soft-delete path |
| W8-A2-005 | P2 | `backend/ent/schema/customfielddefinition.go:17-37` | Нет `version` / `deleted_at` / `updated_at` | Определение поля мутируется (name/options/required/sort); delete без корзины | Выровнять с Issue/CustomField policy: version + soft-delete или явный ADR «config hard-delete» |
| W8-A2-006 | P2 | `backend/ent/schema/component.go:46-49` | Индекс только `name`; нет unique `(project, name)` | Дубликаты имён компонента в одном проекте на уровне БД разрешены | `index.Edges("project").Fields("name").Unique()` |
| W8-A2-007 | P2 | `backend/ent/schema/environment.go:51-54` | Индекс только `env_label`; нет unique `(project, name)` | Дубликаты окружений в проекте | `index.Edges("project").Fields("name").Unique()` |
| W8-A2-008 | P2 | `backend/ent/schema/crmcontact.go:62-66` | Нет индекса по `email` | Lookup контакта по email — full scan | `index.Fields("email")` (или partial WHERE email IS NOT NULL) |
| W8-A2-009 | P2 | `backend/ent/schema/crmdeal.go:69-73` | Нет `index.Edges("owner").Fields("deleted_at")` | Список «мои сделки» фильтрует owner + `deleted_at IS NULL` без composite | Как A8-001: `index.Edges("owner").Fields("deleted_at")` |
| W8-A2-010 | P2 | `backend/ent/schema/filemetadata.go:34-35` | `entity_id` — `field.String`, не UUID | TZ: UUID для идентификаторов сущностей; строка принимает мусор/невалидный ID | `field.UUID("entity_id", …).Optional().Nillable()` (+ миграция) |
| W8-A2-011 | P2 | `backend/ent/schema/filemetadata.go:17-46` | Soft-delete есть, `version` нет; `size` без Min; `uploader` не Required | Concurrent confirm/delete без OL (A8-015); отрицательный size; orphan-файлы без uploader | `version`; `Positive()`/`Min(0)` на size; `uploader` Required или явный system-upload path |
| W8-A2-012 | P2 | `backend/ent/schema/githubintegration.go:30-34` | `webhook_secret` без `Sensitive()`; `access_token` Sensitive, но без `NotEmpty()` | Секрет может попасть в Ent String()/дампы; пустой PAT допустим схемой | `Sensitive()` + `NotEmpty()` на оба секретных поля |
| W8-A2-013 | P2 | `backend/ent/schema/gitlabintegration.go:30-35` | `webhook_secret` `Optional()` и без `Sensitive()` | Схема разрешает интеграцию без секрета (verify отвергает пустой — но запись «дырявая»); не Sensitive | Required/`NotEmpty` + `Sensitive()`; align с GitHub |
| W8-A2-014 | P2 | `backend/ent/schema/githublink.go:17-49` | Есть `updated_at`, нет `version`; нет unique `(integration, external_id, link_type)` | Гонки обновления статуса PR; дубликаты линков | `version`; unique index Edges("integration")+fields |
| W8-A2-015 | P2 | `backend/ent/schema/gitlablink.go:17-52` | Симметрично GitHubLink: нет `version`, нет unique по external | Те же риски для MR/commit links | Как W8-A2-014 |
| W8-A2-016 | P2 | `backend/ent/schema/hubbase.go:50-53` | Индекс только `deleted_at`; нет `Edges("owner")+deleted_at` | `ListBases` WHERE owner + `deleted_at IS NULL` — паттерн A8-001 не применён | `index.Edges("owner").Fields("deleted_at")` |
| W8-A2-017 | P2 | `backend/ent/schema/hubcolumn.go:55-59` | Нет unique `(table, name)`; нет `Edges("table")+deleted_at` | Дубликаты колонок; list columns by table+soft-delete без composite | Unique name per table; `index.Edges("table").Fields("deleted_at")` |
| W8-A2-018 | P2 | `backend/ent/schema/hubrecord.go:49-52` | Только `deleted_at`; нет `Edges("table")+deleted_at` | Hot-path `ListRecords` table + DeletedAtIsNil | `index.Edges("table").Fields("deleted_at")` |
| W8-A2-019 | P3 | `backend/ent/schema/comment.go:29-30` | `mentions` — `[]string`, не `[]uuid.UUID` | User ID как строки обходят UUID-инвариант на уровне схемы | `JSON("mentions", []uuid.UUID{})` |
| W8-A2-020 | P3 | `backend/ent/schema/githubintegration.go:28` (+ githublink/gitlablink comments) | Mojibake в комментариях (`Рмя`) | UTF-8 corruption в schema comments (TASK-007) | Перекодировать комментарии в UTF-8 |

## Чисто

Файлы без находок — что проверялось:

- **coffeeround.go** — UUID v4 PK + `Default(uuid.New).Immutable()`; append-only (нет `updated_at`) → `version`/soft-delete не требуются; индекс `created_at`; defaults `pair_count`/`unpaired_count` = 0 безопасны; timestamps через `nowUTC`.
- **helpers.go** — только `keyRegexp` и `nowUTC()` (`time.Now().UTC()`); канон UTC для всех схем пакета; секретов/defaults нет.

Частично чистые по фокусу (находки только индексы/мелочь):

- **crmcontact.go** — UUID + `version` + `deleted_at` + индексы `deleted_at`/`company`; timestamps UTC. Gap: email index (W8-A2-008).
- **crmdeal.go** — UUID + `version` + `deleted_at` + `stage`/`rank` index; defaults `lead`/`RUB` осмысленны. Gap: owner+deleted_at composite (W8-A2-009).
- **hubbase / hubcolumn / hubrecord** — UUID + `version` + `deleted_at` OK; gaps только missing composite/unique indexes (W8-A2-016…018).
