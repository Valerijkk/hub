# ADR-INTRANET-U03 — News polls + author-delete (FR-INTR-07/08)

- Статус: **Accepted** (owner 2026-07-28: отдельный approve ADR/migrate не требуется)
- Дата: 2026-07-28
- Автор: be1 · Верифицирует: be-ver / arch-ver
- Связанные: FR-INTR-01..03, FR-INTR-07, FR-INTR-08, NFR-11, NFR-31,
  P1-BE-U03 / W7-A2-001, TASK-098, GZ-V4-01/03,
  `docs/system_specification.md` (Intranet ERD `news_poll*`)

## Контекст

A2 W7 residual **W7-A2-001 / P1-BE-U03**: `DeleteNews` только `requireAdmin`;
сущностей `news_poll*` нет; restore из корзины формально не может вернуть
опрос со счётчиками. W12 зафиксировал U-section: нужен scope
«полный greenfield polls» vs «только author-or-admin delete».

Спека уже канонична (опрос-в-посте, admin **или** автор на DELETE,
soft-delete 30д, restore с голосами). Messenger `Poll*` — channel-scoped;
имена `news_poll*` обязательны, чтобы не конфликтовать.

## Рассмотренные варианты

1. **Только author-or-admin delete** — отклонено: FR-INTR-08 / restore-with-poll
   остаются невыполнимы; U03 не закрывается.
2. **Полный `news_poll*` + author-or-admin delete** — выбрано
3. **Company-wide polls (GZ-V4-01 B)** — отклонено для этой задачи;
   остаётся опрос-в-посте (GZ-V4-01 A / спека).

## Решение

1. Ent: `NewsPoll` (1:0..1 к `NewsPost`), `NewsPollOption` (`votes_count`
   денормализация), `NewsPollVote` unique `(poll_id, user_id)` (NFR-31).
2. Soft-hide опроса/голосов — **через** `NewsPost.deleted_at` (без отдельного
   `deleted_at` на poll-строках). List/Get poll фильтруют пост `DeletedAtIsNil`.
   Restore = `ClearDeletedAt` → опрос и счётчики снова видны (FR-INTR-08).
3. Retention 30д (NFR-11) без изменений: purge hard-delete поста; FK
   `ON DELETE CASCADE` на `news_poll*` (SQL migrate + Ent Schema.Create).
4. `DELETE /api/intranet/news/:id`: admin **или** автор поста; member на чужой → 403.
   Domain audit `(actor_id, post_id)` в той же Tx (как сейчас).
5. API polls по спеке: create news `poll?`, `GET/PUT .../poll`, `POST .../poll/vote`.
6. Migrate: additive CREATE TABLE; owner-approve **не требуется** (2026-07-28).

## Finding IDs (TASK-098)

| ID | Вердикт |
|----|---------|
| W7-A2-001 / P1-BE-U03 | FIXED — ADR + код + mutation IT |

## Последствия

- (+) Закрывает FR-INTR-01..03 и 07/08; trash restore осмыслен с опросом
- (−) Новый codegen Ent + SQL migrate; FE должен показать poll UI отдельно
- (−) Company-wide polls остаются вне scope

## Approval record

- Owner: **апрув ADR/migrate не требуется** (2026-07-28)
- Soft-delete 30d: **сохранить** (не обходить)
