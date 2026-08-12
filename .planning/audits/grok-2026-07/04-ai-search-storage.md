# AI / Search / Bases / Storage

**Аудит:** A4 · READ-ONLY · 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Зона:** `backend/internal/{ai,search,bases,storage}`  
**Метод:** Glob/Grep/Read реального `.go` кода; находки только с подтверждаемой строкой.

## Что проверено

| Модуль | Файлы | Фокус |
|--------|-------|--------|
| **ai** | adapters (openai/anthropic/gigachat/yandex/stream), `service.go`, `fallback.go`, `repository_ent.go`, `privacy.go`, `handler`/`admin`/`user`, `content_policy.go`, `doc_search.go` | SSRF (`ValidateOutboundURL` + `NewSafeHTTPClient`), secrets в ответах, quota/usage, stream goroutine/context, BYOK owner scope |
| **search** | `service.go`, `handler.go`, `repository.go`, `jql.go`, `meilisearch_indexer.go` | NF-03 user-scope (projects/channels/messages/users), soft-delete в SQL, saved filters authz, JQL predicates, goroutines+errors |
| **bases** | `service.go`, `repository.go`, `handler.go` | owner scope (`authorizeBase` + `GetTable` baseID), IDOR table↔base, tx на create, pagination records |
| **storage** | `service.go`, `handler.go`, `repository.go`, `filename.go`, `orphan_cleanup.go` | aws-sdk-go-v2 + path-style SeaweedFS, sanitize/path traversal, RBAC по `FileProjectID`, presign TTL/BodyLimit, confirm/download status |

**Чисто по критичным инвариантам (см. §Чисто):** AI adapters валидируют URL и читают body через `LimitReader`; bases owner-scope на CRUD; storage path sanitization и CR-04 RBAC по project файла; search messages scoped через `channel_members`.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A4-001 | P1 | `backend/internal/search/jql.go:104-108` | `assignee = …` всегда даёт `issue.HasAssignee()`, обе ветки идентичны; `HasAssigneeWith` не используется | Фильтр «мои задачи» / конкретный assignee не работает: выбираются любые задачи с любым assignee | `return issue.HasAssigneeWith(user.IDEQ(parsedUUID))`; для `!=` — отрицание; убрать мёртвую ветку после подстановки `currentUser()` |
| A4-002 | P1 | `backend/internal/search/service.go:227-234` (аналогично `:266-273`, `:432-439`, `:470-477`) | tsvector-запросы к `issues` / `test_cases` / `projects` / `file_metadata` **без** `deleted_at IS NULL` (у files также без `status`) | Soft-deleted сущности остаются в Ctrl+K выдаче при живом `search_vector` | Добавить `AND deleted_at IS NULL`; для files — `AND status = 'confirmed' AND deleted_at IS NULL` |
| A4-003 | P1 | `backend/internal/search/service.go:628-645` + `repository.go:65-71` | `CreateSavedFilter` принимает любой `project_id` без проверки membership | Любой auth-user создаёт фильтры в чужом проекте (мусор / pollution / подготовка к share) | Перед create: membership/RBAC на `projectID`; иначе 403 |
| A4-004 | P1 | `backend/internal/search/repository.go:74-84` + `service.go:648-659` | `ListSavedFilters` отдаёт own+shared по `project_id` без membership | Утечка имён/конфигов shared-фильтров чужого проекта (аналог NF-03) | Проверять членство в проекте до list; иначе 403 |
| A4-005 | P1 | `backend/internal/ai/fallback.go:51-61` + `adapter_openai.go:32` + `platform/http/safedialer.go:122-124` | Offline/Ollama путь считает `localhost`/`127.0.0.1` «offline», но `NewSafeHTTPClient` **блокирует** loopback при dial | `AI_OFFLINE_ONLY` / local Ollama фактически неработоспособны: create URL проходит `ValidateOutboundURL`, запрос падает `ErrBlockedAddress` | Отдельный dialer/allowlist для trusted offline endpoints **или** sidecar с публичным DNS; не ослаблять SSRF для user BYOK URL |
| A4-006 | P1 | `backend/internal/storage/repository.go:62-67` | `ConfirmUpload` — `UpdateOneID` без `status=pending` / `deleted_at IS NULL` | Повторный confirm перепривязывает `entity_type`/`entity_id` у уже confirmed файла; нет проверки, что PUT в S3 состоялся | `Where(statusEQ(pending), deletedAtIsNil())`; опционально S3 HeadObject; 409 если уже confirmed |
| A4-007 | P2 | `backend/internal/storage/repository.go:77-80` + `service.go:119-134` | Download/presign GET для любого non-deleted файла, включая `pending` | Presigned URL на незавершённый/пустой объект; обход «confirm before use» | Разрешать download только `status=confirmed` |
| A4-008 | P2 | `backend/internal/storage/service.go:84-88` | Presigned PUT без `ContentLength` / max size (известный NF-09) | Клиент может залить объект >> Fiber BodyLimit (25 MB) напрямую в SeaweedFS | Presign с ограничением размера (если бэкенд поддерживает) + lifecycle/quota; документировать остаточный риск |
| A4-009 | P2 | `backend/internal/ai/service.go:467` | `StartChatStream` → `go runChatStream(context.Background(), …)` | Отмена HTTP/WS клиента не останавливает LLM-вызов → расход токенов/квоты после disconnect | `context.WithoutCancel` только для audit **или** отдельный timeout ctx; прокидывать cancel от stream consumer |
| A4-010 | P2 | `backend/internal/ai/service.go:499` и `:562` | `_ = s.repo.IncrementUsage(...)` / audit errors глотаются | Тихий провал учёта → обход daily quota при ошибках БД | Логировать + метрика; при fail usage не считать chat успешным для quota-critical путей |
| A4-011 | P2 | `backend/internal/ai/repository_ent.go:223-248` | `IncrementUsage`: read → create **или** update без unique upsert/tx | Гонка двух первых запросов дня: один create падает → при A4-010 usage теряется | `ON CONFLICT (user, usage_date) DO UPDATE` / Ent upsert |
| A4-012 | P2 | `backend/internal/ai/repository_ent.go:38-43` | `decryptField`: `pt, _ := MaybeDecryptString` — ошибка decrypt игнорируется | В адаптер может уйти ciphertext как API key → ложные 401 / утечка формата в provider logs | При err decrypt возвращать ошибку наверх, не подставлять raw |
| A4-013 | P2 | `backend/internal/bases/service.go:72-97` | `CreateBase` + default table/column без транзакции | При ошибке на column остаётся base(+table) без дефолтной колонки | Одна Ent/SQL tx на base+table+column |
| A4-014 | P2 | `backend/internal/bases/repository.go:123-126` | `GetTable` грузит **все** records без limit/cursor | Большая таблица → память/latency DoS на `GetTableView` | Offset/limit или cursor; дефолтный потолок |
| A4-015 | P2 | `backend/internal/search/service.go:155-170` | Ошибки шардов только `slog.Warn`; `Search` возвращает `nil` error при частичном fail | Клиент видит урезанную выдачу как полный успех (ложные «нет результатов») | Если все шарды failed → error; иначе `partial: true` / заголовок warning |
| A4-016 | P2 | `backend/internal/search/jql.go:46-56` | `ToEntPredicates` игнорирует `Logic` (AND/OR) — все условия склеиваются как AND | `OR` в JQL семантически неверен | Собирать дерево `issue.And` / `issue.Or` по `cond.Logic` |
| A4-017 | P2 | `backend/internal/search/handler.go:183-185` | Forbidden delete filter → `500` + `err.Error()` | Неверный статус; возможна утечка текста ошибки | Map ownership deny → 403; not found → 404 |
| A4-018 | P2 | `backend/internal/search/meilisearch_indexer.go:33-36` | Meili client = `&http.Client{Timeout:…}` без SafeDial / `ValidateOutboundURL` | `MEILI_HOST` из env без SSRF-guard (паттерн TZ); риск при компромете env | `ValidateOutboundURL` + Safe client; pin internal network policy |
| A4-019 | P3 | `backend/internal/ai/privacy.go:27-29` | Нет preference → `AllowExternalAI = true` | External LLM по умолчанию opt-out, не opt-in | Default `false` для новых пользователей; явный consent |
| A4-020 | P3 | `backend/internal/search/service.go:613-614` | `title = title[:100]` по байтам | Битая UTF-8 граница в snippet сообщений | `utf8`/`runes` truncate |
| A4-021 | P3 | `backend/internal/storage/filename.go:31-32` | `out[:200]` по байтам после sanitize | Редкий mid-rune truncate в S3 key | Truncate по рунам / grapheme |
| A4-022 | P3 | `backend/internal/storage/service.go:91` | `_ = s.repo.DeleteByID` при ошибке presign | Orphan pending остаётся, если delete тоже упал (частично закрыто orphan cleanup) | Логировать delete err; cleanup уже есть — усилить алерт |

## Чисто

- **AI SSRF (user-controlled provider URL):** все chat adapters вызывают `ValidateOutboundURL` + `safehttp.NewSafeHTTPClient`; body через `io.LimitReader` (`adapter_openai.go`, `adapter_anthropic.go`, `adapter_gigachat.go`, `adapter_yandex.go`, `adapter_stream.go`). Create/Update provider тоже валидирует BaseURL (`repository_ent.go:64`, `:101`).
- **AI keys в API list:** `GetUserProvider` обнуляет `APIKeyEnc` (`repository_ent.go:219`); DTO отдаёт `has_api_key`, не ключ.
- **Search NF-03 (основные шарды):** issues/test_cases/projects/files/channels фильтруются `projectIDs` пользователя; users — common-project / admin unscoped; messages — `channel_members` (`service.go:565-576`). Auth на `/api/search` через Group+`authMw`.
- **Bases owner scope:** `authorizeBase` + `ListBases` owner filter; `GetTable(baseID, tableID)` требует принадлежность таблицы базе — IDOR table из чужой базы закрыт.
- **Storage path traversal:** `sanitizeFilename` (`filename.go`) режет path segments/`..`; S3 key = `{project}/{fileID}/{sanitized}`.
- **Storage authz (CR-04):** confirm/download/delete берут `projectID` из метаданных файла, не из body (`handler.go`); upload проверяет `PermFileUpload`.
- **SeaweedFS:** aws-sdk-go-v2, path-style, static creds (`service.go:50-53`) — не minio-go.

## Оценка зоны: 6.5/10

**Обоснование:** базовые security-паттерны зоны (SSRF-guard на AI adapters, bases owner-scope после фикса, storage sanitize+RBAC, search project/channel scope) на месте и выглядят осознанно. Однако есть несколько **P1** с прямым impact: сломанный JQL `assignee`, soft-delete дыра в глобальном поиске, saved-filters без membership, confirm без status-gate, конфликт offline-Ollama vs SafeDial. P2 по quota/stream/tx/pagination ухудшают надёжность, но не открывают полный IDOR на файлы/базы. До «8+/10» нужны фиксы всех P1 и soft-delete/confirm/filters.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 6 |
| P2 | 12 |
| P3 | 4 |
| **Всего** | **22** |

**Score: 6.5/10**
