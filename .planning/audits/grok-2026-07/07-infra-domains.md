# API Gateway / Mail / Intranet / Status / Dashboard / Backup / Dev

> READ-ONLY аудит зоны A7 · 2026-07-26 · ветка `cursor-grok-4.5`  
> Объём: `backend/internal/{apigateway,mail,intranet,status,dashboard,backup,dev}` — 45 `.go` файлов (реальный код, не спеки).

## Что проверено

| Пакет | Файлы / фокус |
|-------|----------------|
| `apigateway` | webhook HMAC/SSRF/worker pool/Ack-NAK, API-key touch pool, RBAC IDOR на `/webhooks/:id`, repository |
| `mail` | InternalTokenAuth (constant-time), SendAlert, admin settings/broadcast, settings store |
| `intranet` | news/kudos authz (admin), directory, coffee matching/enrollment/repository, pagination caps |
| `status` | публичный `/api/status`, probe aggregation, HTTP 503 mapping |
| `dashboard` | personal/project/admin handlers, Redis cache, layout/templates, TopProjects query |
| `backup` | PerformBackup/Restore/cleanup, maintenance mode, S3, pg_restore/mongorestore, admin routes |
| `dev` | `!release` vs `release` build tags, DEV_MODE guard, seed/login |

Искали: `_ =` errors, nil panic, races/goroutines, context, N+1, tx, resources, injection, missing authz, webhook HMAC skip, SSRF, secrets in logs/HTTP, pagination, DEV_MODE leftovers, backup restore risks, off-by-one.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A7-001 | P1 | `dashboard/handler.go:71-81` | `Project` берёт `userID` в `_` и зовёт `GetProjectStats` без membership/RBAC | Любой аутентифицированный юзер читает метрики любого проекта (tasks/bugs/activity). Кеш Redis `dashboard:project:{id}` усиливает утечку | Проверить membership + permission (как tracker/RBAC); не отдавать stats чужого проекта (403/404) |
| A7-002 | P1 | `dashboard/handler.go:111` | Admin dashboard: `SystemRole != "admin"` → 403 | Founder (канон `auth.IsSystemAdmin` / intranet) отрезан от `/dashboard/admin`, хотя это единственный system owner | Разрешить `admin` **и** `founder` (тот же хелпер, что в audit/intranet) |
| A7-003 | P1 | `backup/service.go:206-214` | `setMaintenanceMode` только логирует; флаг/гейт трафика нет | `RestoreBackup` делает `pg_restore --clean` / `mongorestore --drop` на живом инстансе → гонки запросов, частичная порча данных | Реальный maintenance gate (middleware + Redis/flag) до restore; отклонять non-health API на время операции |
| A7-004 | P1 | `backup/service.go:75-77` (и 82-84, 90-91, 102) | При ошибке dump/zip/S3: `markFailed` + `return toBackupResponse(record), nil` | Handler отдаёт **201** с телом, где `status` всё ещё `in_progress` (локальный `record` не перечитан); клиент думает, что бэкап ок | После `markFailed` перечитать запись **или** вернуть error; HTTP → 5xx / 200 с актуальным `failed` |
| A7-005 | P1 | `mail/service.go:32-39` | При непустом `RESEND_API_KEY` SendAlert только логирует `"queued"` и `return nil` | Реальной отправки в Resend нет; `ResendReady=true` + Broadcast `Sent++` создают ложное чувство доставки | Реализовать Resend HTTP client (как `auth/resend`) **или** явно фейлить / `ResendReady=false` пока stub |
| A7-006 | P1 | `intranet/coffee_repository.go:77` | `GetPreviousRoundPairKeys` использует `Offset(1)` | Вызов из `RunCoffeeMatching` **до** `CreateCoffeeRound`: «предыдущий» раунд = latest (`Offset(0)`). Сейчас избегаются пары **предыдущего-предыдущего**; при 1 раунде avoid-map пуст → rematch тех же пар | Брать latest round (`First` / без Offset) **или** переименовать и сменить call-site; поправить тест `coffee_repository_test.go`, который закрепляет off-by-one |
| A7-007 | P2 | `backup/service.go:256-258`, `274` + `handler.go:99` | `CombinedOutput` вшит в `fmt.Errorf`; handler отдаёт `err.Error()` клиенту | В stderr/`--uri=`/`--dbname=` могут светиться connection string / пароли БД в HTTP 500 | Логировать output серверно; клиенту — generic `"restore failed"` + `request_id` |
| A7-008 | P2 | `mail/admin_broadcast.go:39-41` | Broadcast: `EmailNEQ("")` без фильтра статуса | Письма уходят pending/deactivated (и любым с email), не только active | `Where(user.StatusEQ(user.StatusActive), …)` |
| A7-009 | P2 | `backup/service.go:148-151` | Pre-restore backup failure → log + продолжение restore | При сбое safety-net идёт разрушительный restore без отката текущего состояния | Fail-closed: abort restore, если pre-backup не `completed` |
| A7-010 | P2 | `dashboard/service.go:382-412` + `repository.go:191-194` | Top-5 проектов: `WithIssues()` грузит **все** issues всех проектов в память | На больших инстансах — OOM / latency spike на admin dashboard | SQL `GROUP BY` / `Count` по project_id, `ORDER BY count DESC LIMIT 5` |
| A7-011 | P2 | `intranet/service.go:300-312` | Kudo уже в БД; ошибка `NotifyKudosReceived` → error клиенту | Клиент ретраит → дубликаты kudos; частичный успех без компенсации | Логировать notify failure и возвращать 201; либо outbox/async notify |
| A7-012 | P2 | `intranet/coffee_service.go:138-153` | Round + pairs без транзакции | Ошибка на N-й `CreateCoffeePair` оставляет полусобранный раунд и рассылку | Ent Tx: round+pairs атомарно; notify после commit |
| A7-013 | P3 | `backup/service.go:305-308`, `376-379` | `_ =` на `DeleteObject` | S3-объект может остаться после удаления записи / ротации | Логировать/ретраить delete; не удалять DB-row при S3 fail (или orphan GC) |
| A7-014 | P3 | `intranet/service.go:291`, `321` | `len(msg)` / `len(title)` — байты, не руны | Non-ASCII режется раньше лимита (как исторический auth rune-баг) | `utf8.RuneCountInString` |
| A7-015 | P3 | `backup/repository.go:60-64` | `SetRestoredNote` пишет в `ErrorMessage` | Успешный restore перезаписывает/маскирует поле ошибки; семантика поля ломается | Отдельное поле `restored_at` / note, не `error_message` |
| A7-016 | P3 | `backup/handler.go:102` | Ответ `"восстановление запущено"` при синхронном `RestoreBackup` | Клиент думает, что операция async; таймауты прокси режут длинный restore | Либо async job + polling, либо `"восстановление завершено"` после success |
| A7-017 | P3 | `dashboard/service.go:250-262` | Velocity: 10 последовательных `Count…` | Классический N+1 к БД на каждый cache miss | Один агрегирующий запрос / bucket в SQL |

## Чисто

- **apigateway webhooks:** HMAC всегда ставится (`X-Hub-Signature-256`); SSRF через `ValidateOutboundURL` + `SafeHTTPClient`; bounded worker pool; Ack только после успешного enqueue; Stop ждёт `inFlight` (нет write-after-close race); RBAC + `EndpointProjectID` против IDOR на `/webhooks/:id`; secret не отдаётся в DTO.
- **apigateway API keys:** hash-at-rest, expiry, deactivated user reject, bounded `last_used_at` touch pool (не unbounded `go`).
- **dev:** `//go:build release` → no-op `RegisterRoutes`; в non-release — guard `DEV_MODE` → 404. Нет «тайных» prod-маршрутов при `-tags release`.
- **status:** публичный агрегат без секретов; probes с timeout 5s; wiring всегда передаёт ненулевые pinger'ы.
- **intranet news/directory:** admin checks на мутациях; published-only для member GET; soft-delete; pagination capped в repo (≤100).
- **backup routes:** `/admin/backups` под `authMw` + `adminMw` + confirm `RESTORE`.
- **mail InternalTokenAuth:** empty expected → 503; сравнение через `subtle.ConstantTimeCompare`.

## Оценка зоны: 5/10

**Обоснование:** API Gateway и Dev выглядят зрелыми (SSRF/HMAC/RBAC/build-tags). Intranet news/kudos в целом аккуратны. Но зона тянет вниз системными дефектами: **IDOR проектного dashboard**, **founder cut-off на admin stats**, **restore без реального maintenance + ложный 201 на failed backup**, **coffee rematch off-by-one**, **mail stub с ложным ResendReady/Sent**. Это не косметика — authz и data-safety. После закрытия P1 зона может подняться до ~7.5/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 6 |
| P2 | 6 |
| P3 | 5 |
| **Total** | **17** |
| **Score** | **5/10** |
