# Runbooks — эксплуатационные процедуры Hub

> Пошаговые процедуры для дежурного/владельца: что делать, когда что-то происходит в проде. Формат: цель → предусловия → шаги с чек-боксами → верификация → откат.
>
> **Правило runbook-а:** процедура пишется так, чтобы её выполнил человек в 3 часа ночи без контекста. Никаких «очевидно, что», никаких неуказанных путей.

*Создано: 2026-07-19.*

---

## Индекс

| Runbook | Когда открывать | Время |
|---------|-----------------|-------|
| [incident-response.md](incident-response.md) | Прод лежит / деградирует / подозрение на инцидент — **начинать отсюда** | 5 мин до первого действия |
| [monitoring-and-alerts.md](monitoring-and-alerts.md) | Прилетел алерт Prometheus/Alertmanager — что он значит и что делать | 5–15 мин |
| [deploy-and-rollback.md](deploy-and-rollback.md) | Плановый деплой; откат неудачного релиза | 15–30 мин |
| [database-operations.md](database-operations.md) | Миграции, ручные запросы, разбор блокировок, восстановление данных | варьируется |
| [backup-restore.md](backup-restore.md) | Настройка бэкапов; восстановление после потери данных | 30+ мин |
| [fa06-secret-rotation.md](fa06-secret-rotation.md) | Ротация секретов (Resend, Telegram) | ~55 мин |

## Быстрая навигация по симптому

| Симптом | Куда идти |
|---------|-----------|
| Сайт не открывается | [incident-response](incident-response.md) §2 → [monitoring-and-alerts](monitoring-and-alerts.md) `HubBackendMetricsDown` |
| 5xx на части запросов | [monitoring-and-alerts](monitoring-and-alerts.md) `HubHighHTTP5xxRate` |
| Всё медленно | [monitoring-and-alerts](monitoring-and-alerts.md) `HubHighHTTPLatencyP95` |
| Задеплоили и стало хуже | [deploy-and-rollback](deploy-and-rollback.md) §4 «Откат» |
| Не приходят письма/уведомления | [fa06-secret-rotation](fa06-secret-rotation.md) (проверить ключи) + [incident-response](incident-response.md) §5 |
| Кончилось место / БД не пишет | [database-operations](database-operations.md) §5 |
| Утёк секрет | [fa06-secret-rotation](fa06-secret-rotation.md) немедленно |
| Потеряны данные | [backup-restore](backup-restore.md) §3 |

## Инфраструктура (справка, детали — `deployments/README.md`)

**Production-стек** (`deployments/docker-compose.prod.yml`): `postgres`, `mongo`, `redis`, `nats`, `seaweedfs`, `livekit`, `backend`, `hub-mail`, `frontend`, `caddy`.

**Мониторинг** (`deployments/docker-compose.monitoring.yml`, overlay): `prometheus`, `alertmanager`, `loki`, `promtail`, `grafana`.

**Деплой:** GitHub Actions `.github/workflows/deploy.yml` — сборка и push образов, затем SSH-деплой на хост. Каталог приложения на сервере: `/opt/hub` (см. `scripts/deploy-preflight.sh`).

**Health-эндпоинты:** `/api/health` (backend, проверяется в preflight), `/metrics` (Prometheus-метрики, `backend/internal/platform/observability/prometheus.go`).

> ⚠️ **Не подтверждено этим репозиторием:** конкретный хост, домен, объём дисков, наличие внешнего мониторинга/пейджера, политика бэкапов у хостера. Отмечено `TODO(ops)` в соответствующих местах — заполнить владельцу при первой же эксплуатации.

## Конвенции runbook-ов

- `☐` — шаг, который нужно отметить выполненным; runbook проходится сверху вниз.
- ⚠️ — шаг, необратимый или влияющий на прод; читать целиком до выполнения.
- `TODO(ops)` — факт, который нельзя вывести из репозитория; заполняет владелец инфраструктуры.
- Команды даны для Docker Compose на хосте. Для K3s (`deployments/k3s/`) эквиваленты помечены отдельно.
- Каждый инцидент завершается записью: [incident-response](incident-response.md) §7 «Постмортем».
