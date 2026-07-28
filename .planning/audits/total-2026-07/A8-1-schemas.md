# W8 агент A1 — ent/schema part 1

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: UUID v4 PK · `version` optimistic lock · soft-delete · edges/cascade · indexes.  
> Чек-лист TASK-008 адаптирован под schema (не handler/service).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/ent/schema/activitylog.go` | 61 | 0 | чисто |
| `backend/ent/schema/aiprovider.go` | 73 | 1 | замечания |
| `backend/ent/schema/aiusagedaily.go` | 50 | 0 | чисто |
| `backend/ent/schema/aiuserpreference.go` | 53 | 2 | замечания |
| `backend/ent/schema/apikey.go` | 64 | 2 | замечания |
| `backend/ent/schema/auditlog.go` | 62 | 0 | чисто |
| `backend/ent/schema/automationrule.go` | 68 | 1 | замечания |
| `backend/ent/schema/backuprecord.go` | 52 | 1 | замечания |
| `backend/ent/schema/board.go` | 47 | 4 | замечания |
| `backend/ent/schema/bookmark.go` | 51 | 1 | замечания |
| `backend/ent/schema/call.go` | 88 | 3 | замечания |
| `backend/ent/schema/callparticipant.go` | 59 | 0 | чисто |
| `backend/ent/schema/channel.go` | 89 | 2 | замечания |
| `backend/ent/schema/channelinvite.go` | 74 | 2 | замечания |
| `backend/ent/schema/channelmember.go` | 68 | 1 | замечания |
| `backend/ent/schema/coffeeenrollment.go` | 45 | 2 | замечания |
| `backend/ent/schema/coffeepair.go` | 48 | 0 | чисто |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 5 |
| P3 | 2 |
| **Σ** | **10** |

### Топ-3

1. **W8-A1-001 (P1)** — `CoffeeEnrollment`: задуман O2O «один enrollment на user», но нет `Unique` на стороне `User.To` и нет `index.Edges("user").Unique()` → в БД возможны дубликаты строк на одного пользователя.
2. **W8-A1-002 (P1)** — `Channel`: изменяемая сущность с `deleted_at` / `updated_at`, но без `field.Int("version")` — прямое нарушение TZ optimistic lock на горячем мессенджер-ресурсе.
3. **W8-A1-003 (P1)** — Остальные mutable-схемы A1 без `version` (`AIProvider`, `AutomationRule`, `AIUserPreference`, `Board`, `ChannelMember`, `ChannelInvite`, `Call`, `APIKey`, `CoffeeEnrollment`) — остаток системного gap A8-004 (~14 схем с `version` на весь `ent/schema`).

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W8-A1-001 | P1 | `coffeeenrollment.go:32-38` (+ `user.go:135` вне scope) | Edge `From("user")` с `Unique().Required()`, но у `User.To("coffee_enrollment")` нет `.Unique()`, и в `Indexes()` нет `index.Edges("user").Unique()` (в отличие от `AIUserPreference`) | Ent трактует это как O2M: несколько `CoffeeEnrollment` на одного `user`. Двойной enroll / гонка insert → дубликаты, сломанный Random Coffee | Добавить `index.Edges("user").Unique()` и `edge.To("coffee_enrollment", ...).Unique()` на User; миграция + unique constraint |
| W8-A1-002 | P1 | `channel.go:17-57` | Есть `updated_at`, `deleted_at`, нет `field.Int("version").Default(1)` | TZ: «`version` в каждой изменяемой Ent-схеме». Concurrent rename/mute-settings/soft-delete → last-write-wins | Добавить `version`; `VersionEQ` на update/soft-delete в messenger repo |
| W8-A1-003 | P1 | `aiprovider.go:17-55`; `automationrule.go:17-43`; `aiuserpreference.go:17-34`; `board.go:16-36`; `channelmember.go:17-44`; `channelinvite.go:17-48`; `call.go:19-61`; `apikey.go:17-45`; `coffeeenrollment.go:16-29` | Mutable-сущности (есть `updated_at` и/или конкурентные поля: `use_count`, `participant_count`, `role`, `filters`, `api_key_enc`) без `version` | Тот же TZ-инвариант; остаток A8-004. Счётчики Call/Invite частично смягчены `Add*` в сервисе, но метаданные/настройки — нет | Mixin `VersionMixin` + rollout; junction/append-only исключить явно (ADR) |
| W8-A1-004 | P2 | `channel.go:63-78`; `call.go:76-77` | `edge.To` на children (`members`, `invites`, `calls`, `polls`, … / `participants`, `recordings`) без `entsql.Annotation{OnDelete: Cascade\|SetNull}` — в пакете `schema/` **нет ни одного** `OnDelete` | Default FK = NoAction/Restrict. Hard-delete родителя при живых детях → ошибка FK или ручная orphan-очистка в сервисе (хрупко). Soft-delete Channel смягчает, но не закрывает invite/call hard paths | Явно аннотировать OnDelete по политике (Cascade для join, SetNull для optional FK) + Atlas migrate |
| W8-A1-005 | P2 | `board.go:16-47` | Нет `Indexes()`; нет `updated_at`; нет `deleted_at` | List boards по проекту без явного index на edge; нет аудита времени правки filters; soft-delete policy для user-managed board не выровнена с Project/Channel | `index.Edges("project")`; `updated_at`; решить soft-delete vs hard + зафиксировать |
| W8-A1-006 | P2 | `aiuserpreference.go:22-24` | `default_provider_id` — UUID-поле без `edge` на `AIProvider` | Нет FK-целостности: можно сохранить id удалённого/чужого провайдера; проверка только в app-слое | `edge.To/From` + `Field("default_provider_id")` (Optional) или SetNull OnDelete |
| W8-A1-007 | P2 | `bookmark.go:27-28` | `channel_id` — UUID без edge на `Channel` (есть только edge на `User`) | Orphan bookmarks при удалении канала; нельзя `HasChannelWith` в Ent-запросах | `edge.From("channel", Channel.Type).Field("channel_id")` + индекс |
| W8-A1-008 | P2 | `board.go:16-36` | User-managed сущность без `deleted_at` (в A1 soft-delete есть только у `Channel`) | TZ корзина 30д для удаляемого пользовательского контента; board hard-delete необратим | Добавить `deleted_at` + фильтр в list **или** ADR: boards — hard-delete config |
| W8-A1-009 | P3 | `apikey.go:61`; `call.go:85-86`; `channelinvite.go:71` | Индекс на поле, которое уже `Unique()` на field (`key_hash`, `room_name`, `invite_code`, `token`) | Дублирующий unique index в migrate — шум, не баг поведения | Убрать лишние `index.Fields(...).Unique()` / оставить один источник |
| W8-A1-010 | P3 | `apikey.go:48`; `backuprecord.go:41`; `board.go:39` | Комментарий «ибра (edges)» вместо «Рёбра» | Косметика / mojibake-след; путает читаемость | Заменить на «Рёбра» |

## Чисто

Проверено по чек-листу schema: UUID PK (`Default(uuid.New)` = v4) · `version` · `deleted_at` · edges/Unique/Required · indexes · timestamps UTC (`nowUTC`).

| Файл | Что проверено / почему чисто |
|------|------------------------------|
| `activitylog.go` | UUID PK; append-only (нет update-полей) — `version`/soft-delete не требуются; edge issue+user Required Unique; index `(issue, created_at)` |
| `aiusagedaily.go` | UUID PK; unique `(user, usage_date)`; счётчики — кандидаты на атомарный `Add*`, не на soft-delete |
| `auditlog.go` | UUID PK; append-only; индексы action/entity/user/project/created_at; `entity_id` string — осознанный polymorphic (UUID или ключ), не PK |
| `callparticipant.go` | UUID PK; O2M edges Unique+Required; unique `(call_id, user_id)`; join-таблица — soft-delete/`version` не канон |
| `coffeepair.go` | UUID PK; Required edges round+A+B; unique `(round, participant_a, participant_b)` |

**Перепроверка grok A8 (data-layer):**  
- A8-005 Channel `deleted_at` — **закрыто** в текущем `channel.go:47-50` + index `deleted_at`.  
- A8-004 version gap — **открыт**; в A1 подтверждён на Channel и списке mutable-схем (W8-A1-002/003).  
- Cascade/OnDelete — **по-прежнему отсутствует** во всём `ent/schema/` (W8-A1-004).

## Чек-лист (адаптация TASK-008 → schema)

```
Корректность схемы
  [x] UUID PK: Default(uuid.New).Immutable() — все 17
  [x] version на изменяемых — FAIL на Channel + 9 mutable (см. находки)
  [x] soft-delete deleted_at где TZ/корзина — Channel OK; Board gap
  [x] edges Unique/Required согласованы с кардинальностью — FAIL CoffeeEnrollment O2O
  [x] OnDelete/Cascade на parent→child — отсутствует
  [x] indexes под list/FK/soft-delete — Board без Indexes; прочие OK/избыточны

Данные / целостность
  [x] FK-поля без edge (orphan risk) — AIUserPreference.provider, Bookmark.channel
  [x] hard-delete trap при NoAction FK — Channel/Call children

Безопасность (уровень схемы)
  [x] Sensitive() на секретах — AIProvider.api_key_enc OK
  [x] APIKey хранит hash, не plaintext — OK
```
