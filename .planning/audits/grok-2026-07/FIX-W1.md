# FIX-W1 — Security P0 wave (A11-001 … A1-001)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-26  
**Предшествующий docs-коммит:** `f3e1697` — downgrade A12-001 P0→P2 (ent gitignored + CI `go generate`); см. также `FIX-W0.md`.

| # | ID | Commit | Статус |
|---|-----|--------|--------|
| 1 | A11-001 | `da3fe4b` | ✅ |
| 2 | A5-001 | `ce3bbd2` | ✅ |
| 3 | A2-001 | `449c2c3` | ✅ |
| 4 | A1-002 / A1-003 | `48bedbb` | ✅ |
| 5 | A1-001 | `f8e17f9` | ✅ |

**Окружение:** Docker Desktop / Postgres:5432 / Mongo:27017 недоступны. Где нужны были интеграционные фикстуры PG — использован sqlite `enttest` (auth) или stub Repository + miniredis (rbac). Mongo-интеграционные search-тесты SKIP; A2-001 закрыт unit-тестом на mock membership.

---

## A11-001 — FE `/projects/join` vs BE `/invites/join`

### Что изменил

`frontend/src/api/projects.ts`: `joinByInvite` → `POST /invites/join` (как в `internal/project/handler.go`). BE не трогали.

### Тест

`testframework/stages/unit/frontend-test/domains/tracker/projects-join-api.test.ts` — путь URL контракта.

| | Результат |
|---|-----------|
| **ДО** | FAIL: ожидали `/invites/join`, был `/projects/join` |
| **ПОСЛЕ** | PASS (`vitest`, 1 test) |

```text
✓ projects-join-api.test.ts (1 test)
Test Files  1 passed (1)
```

### Решение

Чинить FE: BE-путь канонический (`/invites/join`), алиас не добавляли.

---

## A5-001 — imex CSV: `nil` mappings → все строки skip

### Что изменил

`backend/internal/imex/service.go` `buildFieldMap`: при пустых/nil `mappings` авто-маппинг колонок по имени (`title` / `type` / `priority`, case-insensitive) через уже строившийся `headerIdx`.

### Тест

`TestBuildFieldMap` подкейсы `empty/nil mappings auto-maps… (A5-001)`.

| | Результат |
|---|-----------|
| **ДО** | FAIL: `expected map[title/type/…] got map[]` |
| **ПОСЛЕ** | PASS |

```text
--- PASS: TestBuildFieldMap (0.00s)
    --- PASS: .../empty_mappings_auto-maps_Type/Priority_by_name_(A5-001)
    --- PASS: .../nil_mappings_auto-maps_Title/Type/Priority_export_headers_(A5-001)
```

### Решение

Авто-маппинг по имени колонки — задуманный путь (headers уже индексировались, но не использовались). Явные `mappings` по-прежнему имеют приоритет.

---

## A2-001 — messenger SearchMessages без membership

### Что изменил

- Handler передаёт `userID` в `SearchMessages`.
- При заданном `channel_id` — `IsMember`; иначе — `ListUserChannels` и Mongo `$in` по channel IDs.
- Пустой список каналов → пустая выдача (не org-wide `$text`).

### Тест

`TestSearchMessages_DeniesNonMember_A2_001` (mock repo, без Mongo).

| | Результат |
|---|-----------|
| **ДО** | FAIL: non-member получал hits |
| **ПОСЛЕ** | PASS |

```text
--- PASS: TestSearchMessages_DeniesNonMember_A2_001 (0.00s)
```

Mongo wave123 search — SKIP (Mongo недоступен); unit-контракт membership закрыт.

### Решение

Scope только к каналам членства; явный `channel_id` тоже требует membership (нельзя «подсмотреть» чужой канал зная UUID).

---

## A1-002 / A1-003 — RBAC role IDOR + Update системной роли

### Что изменил

- `UpdateCustomRole` / `DeleteCustomRole` / `AssignCustomRole`: `projectID` из URL сверяется с `GetRoleProjectID` → `ErrRoleProjectMismatch`.
- `UpdateCustomRole`: отказ на `IsSystem` → `ErrSystemRoleUpdate` (симметрия с Delete).
- Handler парсит `projectId`, маппит mismatch → 404, system → 403.

### Тест

`testframework/rbac/custom_roles_a1002_test.go` (stub Repository + miniredis, **без Postgres**).

| | Результат |
|---|-----------|
| **ДО** (guards временно сняты) | FAIL ×4: `Expected ErrRoleProjectMismatch/ErrSystemRoleUpdate but got nil` |
| **ПОСЛЕ** | PASS |

```text
--- FAIL: TestUpdateCustomRole_RejectsForeignProjectRole
--- FAIL: TestAssignCustomRole_RejectsForeignProjectRole
--- FAIL: TestDeleteCustomRole_RejectsForeignProjectRole
--- FAIL: TestUpdateCustomRole_RejectsSystemRole
FAIL  .../testframework/rbac

# после фикса:
ok  github.com/valerii/hub/backend/testframework/rbac  0.846s
```

### Решение

Middleware проверяет права в `:projectId`, но не связывает `roleId` с проектом — скоуп на service-слое. Update системной роли — отдельный gap аудита, закрыт вместе с IDOR.

---

## A1-001 — Pending user → JWT (ORG-06)

### Решение по семантике (до фикса)

По коду + planning:

| Источник | Смысл |
|----------|--------|
| `ECC-AUTH-224` | magic-link для pending: code может уйти, **login blocked** |
| ORG-06 / `ApprovePendingUser` / `PurgeExpiredPending` | pending → очередь admin → active; иначе purge 7d |
| `CreatePendingMember` | явный `StatusPending`, не «ограниченный токен» |

**Выбрано: полный отказ в токене** (`ErrAccountNotActive`), не JWT с урезанным доступом. Ограниченный токен размыл бы границу API и противоречил ECC-AUTH-224.

### Что изменил

1. `LoginWithMagicLink` (+ Telegram login paths): `u.Status != active` → `ErrAccountNotActive` до `CreateSession`.
2. Handler verify → HTTP 403 `"account pending approval"`.
3. `userIsDeactivated` (middleware + refresh): блокирует **любой** non-`active` (pending тоже), не только `deactivated`.

### Тест

`TestLoginWithMagicLink_PendingUserDenied` (sqlite `enttest` + miniredis).

| | Результат |
|---|-----------|
| **ДО** (gate временно убран) | FAIL: `Expected ErrAccountNotActive but got nil` + сессия создавалась |
| **ПОСЛЕ** | PASS; `Session.Count == 0` |

```text
--- FAIL: TestLoginWithMagicLink_PendingUserDenied
    Expected error with "account is not active" in chain but got nil.

# после фикса:
ok  .../testframework/auth  (TestLoginWithMagicLink_PendingUserDenied)
```

Смежные: `TestRequireAuth_DeactivatedUser`, `TestRefreshToken_RejectsDeactivatedUser`, `TestLoginWithMagicLink_Success` — PASS.

---

## Коммиты (порядок)

```text
da3fe4b fix(frontend): point joinByInvite at /invites/join (A11-001)
ce3bbd2 fix(imex): auto-map CSV columns when mappings empty (A5-001)
449c2c3 fix(messenger): scope message search to member channels (A2-001)
48bedbb fix(rbac): scope custom roles to project and block system role update (A1-002/A1-003)
f8e17f9 fix(auth): deny JWT for non-active users including pending (A1-001)
```

Ent-схемы не менялись.
