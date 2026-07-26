# API Contract Backend↔Frontend

## Что проверено

- Backend: `RegisterRoutes` / Group-монтирование в `backend/internal/{auth,project,tracker,sprint,tms,messenger,calls,notification,search,github,gitlab,apigateway,storage,dashboard,audit,backup,ai,mail,bases,crm,hr,calendar,intranet,marketplace,accounting,rbac,imex,status,integrations/onec}` + wiring (`cmd/server/wiring_*.go`)
- Frontend: все клиенты `frontend/src/api/*.ts` (35 файлов) + точечные вызовы вне api (`telegram-button.tsx`, `onboarding.lazy.tsx`, `profile.lazy.tsx`, `import-dialog.tsx`)
- Сверка: route path + HTTP method, multipart/blob transport через `api/client.ts`, ключевые DTO (auth User/Session, IssueResponse, birthday, error shape `AppError`)
- Не аудит: WS `/ws`, metrics, hub-mail internal `/v1`, webhooks внешних систем (GitHub/GitLab/LiveKit) как «FE должен звать»

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A11-001 | P0 | FE `frontend/src/api/projects.ts:133-135` (+ `routes/onboarding.lazy.tsx:312`); BE `backend/internal/project/handler.go:49` | Join по инвайту: FE `POST /projects/join`, BE `POST /invites/join` | Онбординг «вступить в проект» бьёт в несуществующий роут → 404 | Выровнять path на `/invites/join` (тело `{code}` уже совпадает с `JoinByInviteRequest`) |
| A11-002 | P0 | FE `frontend/src/api/tms.ts:681-691` (+ `modules/tms/components/import-dialog.tsx:49`); BE `backend/internal/tms/handler.go:105` + `handler_import.go:97` / `importer.go:268` | TMS import: `FormData` уходит через `api.post` → `JSON.stringify` + `Content-Type: application/json` (`client.ts:44-67`) | Multipart ломается (`{}` body), BE ждёт `FormFile("file")` → импорт из UI не работает | Отдельный multipart-fetch без JSON headers / `fetch` + `FormData` (как в `modules/tracker/import-export.tsx`) |
| A11-003 | P0 | FE `frontend/src/components/auth/telegram-button.tsx:25-27`; BE `backend/internal/auth/handler.go:31-55` | UI шлёт `POST /auth/telegram-code` — роута нет | Кнопка Telegram-login всегда 404; нарушает TZ Magic-Link-only + ADR-AUTH-01 | Убрать UI login через Telegram **или** (только после ADR) добавить согласованный endpoint; сейчас — удалить/скрыть кнопку |
| A11-004 | P1 | FE `frontend/src/api/auth.ts:94-96`; BE `backend/internal/auth/handler.go:31-55` | `loginTelegram` → `POST /auth/telegram` — роута нет (DTO в `auth/dto.go:20-22` мёртвый) | Мёртвый клиент; если кто-то вызовет — 404; путаница с forbidden Telegram-login | Удалить FE-функцию и BE DTO либо явно пометить dead + запретить импорт |
| A11-005 | P1 | FE `frontend/src/api/messenger.ts:44-50`; BE `backend/internal/messenger/handler.go:55-67` (+ schema `channelmember.go` имеет `notify_mode`) | `PUT …/notify-mode` не зарегистрирован | При вызове — 404; поле в Ent есть, API нет | Добавить handler (или писать `notify_mode` через существующий `PUT /channels/:id` / mute API) и покрыть тестом |
| A11-006 | P1 | FE `frontend/src/api/auth.ts:35,64` + `profile.lazy.tsx:136,388`; BE `backend/internal/auth/dto.go:40,91` | Birthday: FE шлёт `YYYY-MM-DD` с `<input type="date">`, BE `*time.Time` ждёт RFC3339 JSON | `encoding/json` не парсит date-only → 400 на сохранении профиля | Custom unmarshal / принимать `2006-01-02` на BE **или** FE конвертировать в RFC3339 перед PUT |
| A11-007 | P1 | FE `frontend/src/api/tms.ts:695-707`; BE `tms/handler.go:106-107` (export CSV/file); transport `client.ts:99-103` | `api.get<Blob>(…/export)` всегда делает `JSON.parse(text)` | CSV/XLSX export через этот клиент невозможно; тип `Blob` — ложный контракт | `fetch` + `response.blob()` / отдельный binary helper (как calendar `.ics`) |
| A11-008 | P2 | FE `frontend/src/api/tracker.ts:116-122`; BE `backend/internal/tracker/dto.go:150-175` | FE `IssueResponse.attachments[]` — в BE `IssueResponse` поля нет | Типы обещают вложения на issue, API их не отдаёт → UI/типы врут | Убрать с FE **или** добавить attachments в BE DTO/handler |
| A11-009 | P2 | FE `frontend/src/api/crm.ts` (есть create/list contacts, нет update/delete); BE `backend/internal/crm/handler.go:26-27` | Контакты: BE `PUT/DELETE /crm/contacts/:id` без FE-обёрток | Админ/CRM UI не может редактировать/удалять контакты через api-слой | Добавить `updateContact`/`deleteContact` в `crm.ts` |
| A11-010 | P2 | FE `frontend/src/api/bases.ts` (нет update/delete base); BE `backend/internal/bases/handler.go:27-28` | `PUT/DELETE /bases/:baseId` не обёрнуты на FE | Нельзя переименовать/удалить базу из api-клиента | Добавить `updateBase`/`deleteBase` |
| A11-011 | P2 | FE `frontend/src/api/marketplace.ts` (нет update/delete); BE `marketplace/handler.go:30,35` | `PUT/DELETE /marketplace/listings/:id` без FE | Редактирование/удаление листинга недоступно клиенту | Добавить клиентские методы |
| A11-012 | P2 | FE `frontend/src/api/dashboard.ts:116-142`; BE `dashboard/handler.go:30-31` | BE `GET /dashboard/templates` + `POST /dashboard/templates/apply` не вызываются FE | Шаблоны дашборда есть на BE, UI api их не знает | Добавить FE api **или** срезать BE если out of scope |
| A11-013 | P2 | FE `frontend/src/api/admin-org.ts` / `admin.ts`; BE `auth/handler.go:50-51` | `GET/PUT /admin/registration-mode` нет на FE | Admin не может переключить open/invite_only из SPA | Добавить в `admin-org.ts` |
| A11-014 | P3 | TZ `docs/tz_constraints.md` / rules: `{"error","request_id"}`; факт BE `platform/errors/errors.go:7-10` + FE `client.ts:10-18` | Документ TZ vs реальный контракт `code`+`message`+`request_id` | BE↔FE согласованы друг с другом; drift только с TZ | Обновить TZ/wiki под `AppError` (не ломать FE) |

## Неиспользуемые backend endpoints (выборка / топ)

Осознанно не полный dump (~286+ REST). Ниже — живые `RegisterRoutes`, для которых **нет** зеркала в `frontend/src/api` (или вызова с тем же path):

| Method | Path | Backend | Примечание |
|--------|------|---------|------------|
| GET/PUT | `/admin/registration-mode` | `auth/handler.go:50-51` | нет FE |
| GET/POST | `/dashboard/templates`, `…/apply` | `dashboard/handler.go:30-31` | нет FE |
| POST | `/projects/:id/unarchive` | `project/handler.go:33` | FE только archive/delete/restore |
| GET | `/projects/trash` | `project/handler.go:29` | нет FE |
| GET/DELETE | `/projects/:id/invites`, `…/:inviteId` | `project/handler.go:45-46` | FE только create invite |
| PUT | `/projects/:id/members/:userId` | `project/handler.go:40` | нет FE (смена роли) |
| CRUD | `/teams`, `/teams/:id/…` | `project/handler.go:53-57` | нет FE api |
| CRUD | `/projects/:id/roles…` | `rbac/handler.go:22-27` | нет FE api |
| POST/GET/DELETE | templates `/projects/:id/templates…` | `project/templates.go:210-214` | нет FE |
| POST/GET | imex import/export/pdf | `imex/handler.go:28-35` | FE tracker import — raw fetch, не api-модуль |
| PUT/DELETE | `/crm/contacts/:id` | `crm/handler.go:26-27` | нет FE |
| PUT/DELETE | `/bases/:baseId` | `bases/handler.go:27-28` | нет FE |
| POST | `/accounting/accounts` | `accounting/handler.go:24` | FE list only |
| PUT/DELETE | `/marketplace/listings/:id` | `marketplace/handler.go:30,35` | нет FE |
| GET | `/intranet/coffee/stats` | `intranet/handler.go:35` | нет FE |
| GET/PUT | `/intranet/news/:id`, `/news/admin` | `intranet/handler.go:40-43` | FE list/create/delete only |
| PUT | `/messenger/channels/:id` | `messenger/handler.go:59` | нет FE updateChannel |
| POST | `/messenger/channels/:id/join\|leave\|members` | `handler.go:61-63` | нет FE |
| POST | `/messages/:messageId/promote-to-issue` | `handler.go:117` | нет FE |
| POST | `/calls/webhook` | `calls/handler.go:80` | внешний, ок |
| * | TMS webhook/allure/junit, часть reports | `tms/handler.go` | CI/external; FE покрывает основной UI |
| * | `/dev/*` | `dev/handler.go` | только non-release |

## Frontend calls to missing routes

| Method | FE path | Где | Backend reality |
|--------|---------|-----|-----------------|
| POST | `/projects/join` | `projects.ts:135`, onboarding | Есть `POST /invites/join` |
| POST | `/auth/telegram-code` | `telegram-button.tsx:25` | Роута нет |
| POST | `/auth/telegram` | `auth.ts:96` | Роута нет (deprecated) |
| PUT | `/messenger/channels/:id/notify-mode` | `messenger.ts:48` | Роута нет (`notify_mode` только в Ent) |

Совпадающие «сложные» пути (проверены, **не** missing): `/invites/join` body `code`; `/issues/:key/github-links`; `/link-preview`; `/ai/v1/chat/completions`; GitLab `/issues/:key/gitlab/create-branch`; message-links; calendar `/export.ics` (отдельный fetch).

## Чисто

- Error JSON BE↔FE: `code` + `message` + `request_id` (`AppError` ↔ `ApiError`) — согласованы; 429 + `Retry-After` обрабатывается
- Auth core: magic-link / verify / me / sessions / logout / refresh (`client.ts:128` ↔ `auth/handler.go:37`)
- Tracker issues/boards/workflow/labels/custom-fields/automation paths зеркалят `tracker/handler.go`
- Sprint + poker paths зеркалят `sprint/handler.go`
- Calls (guest-join, scheduled-calls, recordings, conflicts) ↔ `calls/handler.go`
- Notifications (вкл. telegram bind, push subscribe/unsubscribe body `{endpoint}`) ↔ `notification/handler.go`
- Search `/search` + `/filters` + JQL; AI models/bot/preferences/jql/composer; admin AI / mail / onec / backups confirm=`RESTORE`
- Status `/status`, files, apikeys `/user/api-keys`, webhooks project+global — paths OK
- Даты в list/responses в целом RFC3339 `time.Time` ↔ `string` на FE (кроме birthday date-only)

## Оценка зоны: 5/10

Критичные живые разрывы (join path, TMS multipart import, Telegram login UI → missing routes) рядом с в целом зрелым зеркалом основных доменов. Score снижен из‑за P0 в user-facing flows; после фикса A11-001…003 зона уйдёт к ~7–8/10.

### Counts

| Метрика | Значение |
|---------|----------|
| Findings total | **14** (A11-001…014) |
| P0 / P1 / P2 / P3 | **3 / 4 / 6 / 1** |
| Missing FE→BE routes | **4** |
| Unused BE sample rows | **20+** (таблица выше) |
| Zone score | **5/10** |
