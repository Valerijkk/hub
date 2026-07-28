# Runbook — операции с базами данных

> **Кто выполняет:** владелец / дежурный. **Когда:** миграции, разбор медленных запросов и блокировок, отказ БД, ручные операции с данными.
> **Главное правило:** любая ручная операция записи в прод-БД делается только после бэкапа ([backup-restore.md](backup-restore.md) §2) и внутри транзакции с проверкой перед `COMMIT`.

---

## 1. Хранилища Hub и их роль

| Хранилище | Что хранит | Отказ означает |
|-----------|-----------|----------------|
| **PostgreSQL** | Основные доменные данные (пользователи, проекты, задачи, базы) | Прод не работает |
| **MongoDB** | Сообщения/чат, поисковые структуры | Не работают чат/поиск |
| **Redis** | Сессии, кэш, refresh-токены | Разлогин, рост нагрузки на PG |
| **NATS (JetStream)** | События между модулями | Не срабатывают автоматизации/нотификации |
| **SeaweedFS** | Файлы и вложения (S3-совместимо) | Не грузятся/не отдаются файлы |

Схема Postgres управляется **Ent** (`backend/ent/migrate/`) — миграция применяется приложением при старте, отдельного шага «накатить SQL» в деплое нет.

---

## 2. Подключение (диагностика)

⚠️ Подключение к прод-БД — операция с риском. Читать можно; писать — только по §6.

```bash
cd /opt/hub/deployments

# PostgreSQL
docker compose -f docker-compose.prod.yml -p hub exec postgres psql -U hub -d hub

# MongoDB
docker compose -f docker-compose.prod.yml -p hub exec mongo mongosh

# Redis
docker compose -f docker-compose.prod.yml -p hub exec redis redis-cli
```

> `TODO(ops)`: сверить имена пользователя/БД с фактическими в `/opt/hub/deployments/.env` — здесь приведены значения по образцу dev-конфигурации.

---

## 3. Медленно / деградация производительности

> Симптом со стороны мониторинга: `HubHighHTTPLatencyP95` ([monitoring-and-alerts.md](monitoring-and-alerts.md) §3.3).

### 3.1 Активные и «зависшие» запросы (Postgres)

```sql
-- Что выполняется прямо сейчас, дольше 5 секунд
SELECT pid, now() - query_start AS duration, state, left(query, 120) AS query
FROM pg_stat_activity
WHERE state <> 'idle' AND now() - query_start > interval '5 seconds'
ORDER BY duration DESC;
```

☐ Долгий запрос (минуты) — кандидат на прерывание, но **сначала пойми, что это**: миграция/бэкап прерывать нельзя.

```sql
-- Мягкая отмена запроса (предпочтительно)
SELECT pg_cancel_backend(<pid>);
-- ⚠️ Жёсткий разрыв соединения — только если отмена не сработала
SELECT pg_terminate_backend(<pid>);
```

### 3.2 Блокировки

```sql
-- Кто кого блокирует
SELECT blocked.pid AS blocked_pid, blocking.pid AS blocking_pid,
       left(blocked.query, 80) AS blocked_query,
       left(blocking.query, 80) AS blocking_query
FROM pg_stat_activity blocked
JOIN pg_stat_activity blocking ON blocking.pid = ANY(pg_blocking_pids(blocked.pid))
WHERE cardinality(pg_blocking_pids(blocked.pid)) > 0;
```

☐ Цепочка блокировок обычно начинается с одной долгой транзакции — снимать её, а не всех подряд.
☐ Частая причина в приложении: транзакция, открытая на время внешнего вызова. Это баг кода → тикет, а не только ручное лечение.

### 3.3 Исчерпание пула соединений

```sql
SELECT count(*) AS total, state FROM pg_stat_activity GROUP BY state;
SHOW max_connections;
```

☐ `total` близко к `max_connections` → приложение не отдаёт соединения (утечка) либо пул настроен больше лимита БД.
☐ Много `idle in transaction` — самый опасный признак: транзакции открыты и держат блокировки. Это дефект кода.

### 3.4 Размер и раздувание таблиц

```sql
SELECT relname, pg_size_pretty(pg_total_relation_size(relid)) AS size,
       n_live_tup, n_dead_tup
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 15;
```

☐ Большой `n_dead_tup` относительно `n_live_tup` → autovacuum не справляется; разово: `VACUUM (ANALYZE) <table>;`.

---

## 4. БД недоступна

☐ Контейнер жив?

```bash
docker compose -f docker-compose.prod.yml -p hub ps postgres mongo redis nats
docker compose -f docker-compose.prod.yml -p hub logs --tail=100 postgres
```

☐ Частые причины в порядке проверки:

| Причина | Признак в логах | Действие |
|---------|-----------------|----------|
| Кончилось место | `No space left on device` | [incident-response.md](incident-response.md) §5.1 — освободить место, затем рестарт |
| Исчерпаны соединения | `too many clients already` | §3.3, перезапуск приложения освободит соединения |
| Повреждение данных | ошибки консистентности при старте | ⚠️ **не рестартовать по кругу** — [backup-restore.md](backup-restore.md) §3 |
| OOM | контейнер убит, `Exit 137` | Проверить `free -m`, лимиты памяти |

⚠️ При подозрении на повреждение данных повторные рестарты могут усугубить. Снять копию volume **до** дальнейших действий.

---

## 5. Место кончается

☐ Общая картина: `df -h`, `docker system df`.
☐ Размер volumes:

```bash
docker system df -v | grep -E 'hub_(pgdata|mongodata|seaweeddata|redisdata|natsdata)'
```

☐ Внутри Postgres — что растёт: §3.4.
☐ Что можно чистить безопасно: старые Docker-образы (`docker image prune -f`), кэш сборки (`docker builder prune -f`), логи контейнеров при настроенной ротации.
☐ ⚠️ Что **нельзя**: `docker system prune -a --volumes` (снесёт данные), ручное удаление файлов внутри volume БД.

---

## 6. Ручные операции с данными (крайний случай)

> Ручная правка прод-данных — последнее средство. Сначала спроси: можно ли починить через приложение/фичу?

☐ Снять бэкап **до** операции ([backup-restore.md](backup-restore.md) §2.1) — не «после проверки», а до.
☐ Сформулировать письменно: что меняем, сколько строк ожидаем затронуть, как проверим.
☐ Выполнять **в транзакции** с проверкой до фиксации:

```sql
BEGIN;
-- 1. Сначала SELECT: убедиться, что затрагиваются ровно те строки
SELECT count(*) FROM <table> WHERE <condition>;
-- 2. Затем UPDATE/DELETE
UPDATE <table> SET <...> WHERE <condition>;
-- 3. Проверить, что затронуто ожидаемое количество
-- Если число НЕ совпало с ожиданием:
ROLLBACK;
-- Только при совпадении:
COMMIT;
```

☐ ⚠️ `UPDATE`/`DELETE` без `WHERE` — запрещены категорически.
☐ Записать выполненную операцию в журнал (`.planning/ops/production/`): что, зачем, когда, кто, сколько строк.
☐ Если операция понадобилась из-за бага — завести тикет на исправление причины. Ручная правка лечит симптом.

---

## 7. Миграции схемы

Схема применяется Ent при старте приложения (`backend/ent/migrate/`).

☐ Перед релизом с изменением схемы — бэкап Postgres (обязательно, [deploy-and-rollback.md](deploy-and-rollback.md) §6.1).
☐ Предпочитать аддитивные изменения — они совместимы со старым кодом, откат безопасен.
☐ Разрушительные изменения (удаление/переименование) разбивать на два релиза: сначала код перестаёт использовать поле, следующим релизом поле удаляется.
☐ После деплоя проверить логи старта backend — ошибки миграции проявляются именно там.
☐ Помнить: **откат кода не откатывает схему**. План отката для миграции — отдельный (восстановление из бэкапа либо компенсирующая миграция).

---

## 8. Redis / NATS

### Redis

```bash
docker compose -f docker-compose.prod.yml -p hub exec redis redis-cli INFO memory | head -10
docker compose -f docker-compose.prod.yml -p hub exec redis redis-cli DBSIZE
```

⚠️ `FLUSHALL` на проде — разлогинивание всех пользователей и потеря кэша. Только осознанно, с уведомлением.

### NATS / JetStream

☐ Проверить, обрабатываются ли сообщения (растёт ли необработанная очередь) — накопление означает, что консьюмер в приложении не справляется или упал.
☐ `TODO(ops)`: зафиксировать конкретные команды `nats` CLI / способ инспекции стримов для этой инсталляции.

---

## 9. Чек-лист перед любой операцией с прод-БД

- [ ] Понимаю, что именно и зачем делаю (сформулировано письменно)
- [ ] Бэкап снят и проверен
- [ ] Операция обратима (транзакция) либо есть план восстановления
- [ ] Ожидаемое число затронутых строк известно **до** выполнения
- [ ] Есть кому эскалировать, если пойдёт не так
- [ ] Действие будет записано в журнал
