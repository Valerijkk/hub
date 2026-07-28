# ADR-INTRANET-U05 — Random Coffee: notify-only (FR-INTR-11)

- Статус: **Accepted** (owner 2026-07-28: отдельный approve ADR не требуется)
- Дата: 2026-07-28
- Автор: be2 · Верифицирует: be-ver / product-ver
- Связанные: FR-INTR-10, FR-INTR-11, NFR-26, P2-BE-U05 / W7-A2-008,
  TASK-106, GZ-V4-03 coffee-минимум,
  `docs/system_specification.md` (Intranet Random Coffee)

## Контекст

A2 W7 residual **W7-A2-008 / P2-BE-U05**: `RunCoffeeMatching` вызывает только
`NotifyCoffeeMatch` (ошибки глотались через `_ =`); auto-DM в messenger и
`CoffeePair.dm_channel_id` отсутствуют. Спека FR-INTR-11 требовала оба канала
(DM + notification), из‑за чего поведение выглядело «полусделанным».

W12 U05: без scope «greenfield DM vs только notification» мутация на
отсутствующем API неоднозначна. Greenfield auto-DM тянет messenger
CreateDM/reuse, Ent migrate и риск спама DM без явного решения.

## Рассмотренные варианты

1. **Auto-DM + notification** (полная спека) — отклонено для этой волны:
   messenger-coupling + Ent `dm_channel_id` + migrate; риск спама DM;
   вне минимального coffee-scope intranet-модуля.
2. **Notify-only** (in-app notification через `CoffeeNotifier`) — **выбрано**
3. **Ничего не слать** — отклонено: участники не узнают о паре.

## Решение

1. Канон FR-INTR-11 (v4 coffee-минимум): после создания пары/тройки каждый
   участник получает **in-app notification** через `CoffeeNotifier.NotifyCoffeeMatch`.
2. **Auto-DM запрещён** до отдельного ADR/задачи (messenger + schema).
   Константа `CoffeeMatchDeliveryMode = "notify_only"` — явный контракт в коде.
3. Ошибки notify **не** глотаются молча: `slog.WarnContext` с `user_id` /
   `partner_id` / `error`; матчинг раунда не откатывается (best-effort notify).
4. Opt-out до запуска: только `ListActiveCoffeeEnrollments` — без изменений
   (выключивший не матчится).
5. Migrate / `dm_channel_id` — **не делать**.

## Finding IDs (TASK-106)

| ID | Вердикт |
|----|---------|
| W7-A2-008 / P2-BE-U05 | FIXED — ADR notify-only + slog + mutation IT |

## Последствия

- (+) Однозначный канон; нет спама DM; закрывает U05 без greenfield messenger
- (+) Спека FR-INTR-11 синхронизирована с notify-only
- (−) Авто-чат пары отложен; пользователи открывают контакт сами из notify/UI
- (−) Полный FR с DM — только после нового ADR

## Approval record

- Owner: **апрув ADR не требуется** (2026-07-28)
- Auto-DM: **запрещён** до отдельного ADR (не спамить)
