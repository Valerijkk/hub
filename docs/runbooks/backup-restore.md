# Runbook — бэкапы и восстановление

> **Кто выполняет:** владелец инфраструктуры. **Когда:** настройка бэкапов; восстановление после потери/повреждения данных; перед опасным релизом.
> **Главное правило:** бэкап, который ни разу не восстанавливали, — это не бэкап, а надежда. Проверка восстановлением — ежемесячно (§4).

---

## 1. Что именно нужно бэкапить

Состояние живёт в именованных Docker volumes (`deployments/docker-compose.prod.yml`):

| Volume | Что внутри | Критичность | Потеря означает |
|--------|-----------|-------------|------------------|
| `pgdata` | PostgreSQL — основные данные (пользователи, задачи, проекты, базы) | **Критично** | Потеря продукта |
| `mongodata` | MongoDB — сообщения/чат-история, поисковые данные | **Критично** | Потеря переписки |
| `seaweeddata` | SeaweedFS — загруженные файлы и вложения | **Критично** | Потеря файлов |
| `redisdata` | Redis — сессии, кэш, refresh-токены | Средняя | Разлогин пользователей, потеря кэша |
| `natsdata` | NATS JetStream — очереди/стримы событий | Средняя | Потеря недоставленных событий |
| `caddydata` | Caddy — TLS-сертификаты | Низкая | Перевыпуск сертификатов (автоматом) |

Помимо volumes:

| Артефакт | Где | Замечание |
|----------|-----|-----------|
| `.env` production | `/opt/hub/deployments/.env` | ⚠️ Содержит секреты. Бэкапить **зашифрованным**, никогда не в git |
| Конфиги деплоя | git-репозиторий | Уже версионируется |
| Образы приложения | GHCR | Восстанавливаются пересборкой из git |

> `TODO(ops)` — заполнить владельцу, из репозитория не выводится:
> - есть ли сейчас автоматические бэкапы и куда они складываются;
> - есть ли снапшоты дисков у хостера (часто это самый простой уровень защиты);
> - целевые **RPO** (сколько данных допустимо потерять) и **RTO** (за сколько обязаны восстановиться).

---

## 2. Снятие бэкапа

### 2.1 PostgreSQL (логический дамп — предпочтительно)

```bash
cd /opt/hub/deployments
TS=$(date +%F-%H%M)
docker compose -f docker-compose.prod.yml -p hub exec -T postgres \
  pg_dump -U hub -d hub --format=custom > /var/backups/hub/pg-${TS}.dump
```

☐ Проверить, что дамп непустой и читается:

```bash
ls -lh /var/backups/hub/pg-${TS}.dump
docker compose -f docker-compose.prod.yml -p hub exec -T postgres \
  pg_restore --list /var/backups/hub/pg-${TS}.dump | head -5   # список объектов = дамп валиден
```

> `TODO(ops)`: сверить имя пользователя/БД (`-U hub -d hub`) с фактическими значениями в `.env` — здесь взяты имена из dev-конфигурации.

### 2.2 MongoDB

```bash
TS=$(date +%F-%H%M)
docker compose -f docker-compose.prod.yml -p hub exec -T mongo \
  mongodump --archive --gzip > /var/backups/hub/mongo-${TS}.archive.gz
```

☐ Проверить размер и что архив распаковывается (`gzip -t`).

### 2.3 Файлы (SeaweedFS)

Volume копируется целиком. ⚠️ Для консистентности предпочтительно остановить сервис на время копирования; при копировании «на живую» возможны частично записанные файлы.

```bash
TS=$(date +%F-%H%M)
docker run --rm -v hub_seaweeddata:/data -v /var/backups/hub:/backup alpine \
  tar czf /backup/seaweed-${TS}.tar.gz -C /data .
```

> Имя volume с префиксом проекта: `-p hub` → `hub_seaweeddata`. Проверить фактическое: `docker volume ls | grep seaweed`.

### 2.4 Секреты

⚠️ `.env` содержит боевые секреты. Бэкап только зашифрованным, хранение — отдельно от бэкапов данных.

```bash
gpg --symmetric --cipher-algo AES256 -o /var/backups/hub/env-$(date +%F).gpg /opt/hub/deployments/.env
```

Никогда: не класть `.env` в git, не отправлять в чат/тикет, не логировать.

### 2.5 Автоматизация

`TODO(ops)`: завести cron/systemd-timer, который выполняет §2.1–2.3 ежедневно, складывает в **отдельное от прод-хоста** хранилище и удаляет старше N дней. Бэкап, лежащий только на том же диске, что и данные, не защищает от отказа диска.

Минимальный ориентир политики (уточнить под RPO): ежедневный бэкап, хранение 14 дней, плюс еженедельный с хранением 8 недель.

---

## 3. Восстановление

⚠️ **Перед любым восстановлением:** снять бэкап *текущего* (даже повреждённого) состояния. Если восстановление пойдёт не так, откатиться будет некуда.

### 3.1 Оценить масштаб

☐ Что потеряно: одна таблица/коллекция, вся БД, весь хост?
☐ На какой момент восстанавливаем (какой бэкап последний валидный)?
☐ Сколько данных теряем между бэкапом и инцидентом — это фактический RPO, зафиксировать в постмортеме.
☐ Уведомить пользователей, если восстановление означает потерю их работы за период.

### 3.2 PostgreSQL

☐ Остановить приложение, чтобы оно не писало в восстанавливаемую БД:

```bash
cd /opt/hub/deployments
docker compose -f docker-compose.prod.yml -p hub stop backend hub-mail
```

☐ Восстановить (⚠️ `--clean` удалит существующие объекты перед восстановлением):

```bash
cat /var/backups/hub/pg-<TS>.dump | docker compose -f docker-compose.prod.yml -p hub exec -T postgres \
  pg_restore -U hub -d hub --clean --if-exists
```

☐ Поднять приложение и проверить:

```bash
docker compose -f docker-compose.prod.yml -p hub start backend hub-mail
curl -sf http://127.0.0.1:3000/api/health | head -c 200; echo
```

☐ Проверить данные реальным сценарием (логин + чтение записи), а не только health.

### 3.3 MongoDB

```bash
docker compose -f docker-compose.prod.yml -p hub stop backend
gunzip -c /var/backups/hub/mongo-<TS>.archive.gz | \
  docker compose -f docker-compose.prod.yml -p hub exec -T mongo mongorestore --archive --gzip --drop
docker compose -f docker-compose.prod.yml -p hub start backend
```

⚠️ `--drop` удаляет коллекции перед восстановлением — убедись, что восстанавливаешь полный, а не частичный дамп.

### 3.4 Файлы (SeaweedFS)

```bash
docker compose -f docker-compose.prod.yml -p hub stop seaweedfs
docker run --rm -v hub_seaweeddata:/data -v /var/backups/hub:/backup alpine \
  sh -c "rm -rf /data/* && tar xzf /backup/seaweed-<TS>.tar.gz -C /data"
docker compose -f docker-compose.prod.yml -p hub start seaweedfs
```

### 3.5 Полная потеря хоста

☐ Поднять новый хост, установить Docker.
☐ Склонировать репозиторий в `/opt/hub`.
☐ Восстановить `.env` из зашифрованного бэкапа (§2.4).
☐ Поднять инфраструктуру: `docker compose -f docker-compose.prod.yml --env-file .env -p hub up -d postgres mongo redis nats seaweedfs livekit`.
☐ Восстановить данные: §3.2 → §3.3 → §3.4.
☐ Задеплоить приложение: [deploy-and-rollback.md](deploy-and-rollback.md) §3 «Ручной деплой».
☐ Переключить DNS на новый хост, дождаться выпуска сертификатов Caddy.
☐ Проверка: §4.1 деплой-runbook + реальные пользовательские сценарии.

---

## 4. Ежемесячная проверка восстановлением (обязательна)

> Смысл: убедиться, что бэкапы **реально восстанавливаются**, до того как это понадобится в аварии. Проверка «файл существует и весит много» ничего не доказывает.

☐ Взять последний бэкап Postgres.
☐ Поднять **временный** контейнер Postgres (не трогая прод):

```bash
docker run -d --name pg-restore-test -e POSTGRES_PASSWORD=test -e POSTGRES_USER=hub -e POSTGRES_DB=hub postgres:16
sleep 10
cat /var/backups/hub/pg-<TS>.dump | docker exec -i pg-restore-test pg_restore -U hub -d hub --clean --if-exists
```

☐ Проверить, что данные на месте (несколько ключевых таблиц непусты):

```bash
docker exec -i pg-restore-test psql -U hub -d hub -c "\dt" | head -20
```

☐ Удалить временный контейнер: `docker rm -f pg-restore-test`.
☐ Записать результат проверки (дата, бэкап, успех/проблемы) — история проверок важнее одной успешной.

> `TODO(ops)`: сверить версию образа Postgres с фактической в `docker-compose.prod.yml` — восстановление в младшую версию не сработает.

---

## 5. Чек-лист готовности (заполнить один раз, пересматривать раз в квартал)

- [ ] Автоматические бэкапы Postgres настроены и работают (проверено по факту наличия свежих файлов)
- [ ] Автоматические бэкапы Mongo настроены
- [ ] Бэкапы файлов (SeaweedFS) настроены
- [ ] Бэкапы хранятся **вне** прод-хоста
- [ ] `.env` забэкаплен зашифрованным, ключ шифрования доступен владельцу отдельно
- [ ] RPO/RTO зафиксированы письменно
- [ ] Проверка восстановлением проведена за последний месяц
- [ ] Процедура восстановления пройдена хотя бы раз целиком (не только Postgres)
