# ADR-STORAGE-NF09-SIZE-2026 — Presigned upload: обязательный size + ContentLength

- Статус: **Accepted** (owner 2026-07-28: отдельный approve ADR не требуется)
- Дата: 2026-07-28
- Автор: be2 · Верифицирует: be-ver / security-auditor
- Связанные: TZ NF-09, NFR-04/NFR-27, W7-A4-001, P1-BE-38, TASK-024 (partial),
  TASK-099, GZ-SEC-27

## Контекст

Fiber `BodyLimit` 25 MB не действует на direct PUT по presigned URL в SeaweedFS.
TASK-024 закрыл путь `size > 0` (bind `ContentLength` в presign + Confirm HEAD).
Legacy `size == 0` / omitted size оставлял unsigned PUT → клиент мог залить >25 MB
(NF-09 residual W7-A4-001).

## Рассмотренные варианты

1. **Reject `size <= 0` + всегда `ContentLength` в PresignPutObject** — выбрано
2. S3 bucket policy `content-length-range` only (GZ-SEC-27) — дополнение infra,
   не заменяет API-контракт; отложено как defense-in-depth вне этой задачи
3. Оставить legacy + только Confirm HEAD — отклонено: диск уже заполнен до Confirm
4. Подписывать `ContentLength = maxUploadBytes` при size==0 — отклонено:
   S3 PUT ContentLength — exact match, не upper bound

## Решение

1. `POST /api/files/upload-url` **требует** `size` (байты) в диапазоне
   `1 … maxUploadBytes` (25 MiB). `size <= 0` / omitted → `400` `ErrInvalidFileSize`.
2. `GenerateUploadURL` всегда ставит `PutObjectInput.ContentLength = size`
   (oversized PUT отклонит S3/совместимый бэкенд).
3. Confirm сохраняет `assertObjectWithinLimit` (HEAD + delete orphan) как
   defense-in-depth для уже созданных pending и сбоев S3.
4. Клиенты (FE `uploadFile`) обязаны передавать `file.size`.

## Finding IDs (TASK-099)

| ID | Вердикт | Evidence |
|----|---------|----------|
| W7-A4-001 | **FIXED** | reject size≤0 + ContentLength always; тест `TestGenerateUploadURL_RejectsZeroSize_W7_A4_001` |

## Последствия

- (+) NF-09 закрыт на API-границе; нет unsigned oversized PUT
- (−) Breaking для клиентов без `size` (канонический FE обновляется вместе с фиксом)
- (−) Bucket-policy upper-bound остаётся желательным ops-слоем (не блокер)

## Approval record

- Owner: **апрув ADR не требуется** (2026-07-28, parent TASK-099)
- Migrate: N/A (без Ent schema)
