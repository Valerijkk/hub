# Политика безопасности Hub

> Технические детали, статусы аудитов и чек-листы для верификаторов: [docs/security_guidelines.md](docs/security_guidelines.md)
> Инварианты ТЗ: [docs/tz_constraints.md](docs/tz_constraints.md) · Архитектурные решения: [.planning/decisions/](.planning/decisions/)

---

## Сообщение об уязвимостях

Hub — проприетарное self-hosted ПО. Если вы обнаружили уязвимость:

1. **Сообщите приватно** правообладателю: `<security-contact@placeholder>` (плейсхолдер — заменить на реальный контакт). Не создавайте публичных issues и не обсуждайте уязвимость в открытых каналах.
2. Приложите: описание, шаги воспроизведения, затронутые модуль/endpoint/версию, оценку воздействия (что может сделать атакующий).
3. **Не публикуйте** детали уязвимости до выпуска исправления и явного согласования с правообладателем (coordinated disclosure).

**Сроки:** подтверждение получения — до **72 часов**; первичная оценка severity — до **7 дней**; исправление CRITICAL/HIGH — приоритетно, в ближайшем релизе.

## Поддерживаемые версии

| Версия | Поддержка |
|--------|-----------|
| `master` (текущая разработка) | ✅ Security-фиксы |
| Последний релизный тег | ✅ Security-фиксы (hotfix-ветка) |
| Более старые теги | ❌ Обновитесь до последнего тега |

## Security-инварианты Hub

Эти гарантии встроены в платформу и **не могут быть ослаблены** без принятого ADR (см. `docs/tz_constraints.md`):

### Аутентификация и сессии

- **Magic Link (email) — единственный метод входа.** Без паролей; Telegram — только доставка уведомлений, не login. SSO (OIDC + LDAP) — только после принятого `ADR-AUTH-SSO`.
- **JWT access-токен: 15 минут TTL.** Refresh-токен — в Redis (хранится как SHA-256 hash), **30 дней TTL, мгновенный отзыв** (revocation); пользователь видит список сессий (device/IP) и может отозвать любую.
- **Rate limits:** auth — 3 попытки / 10 минут; API — 100/сек; файлы — 10/мин; сообщения — 60/мин.
- Magic-link код: TTL 10 минут, сравнение только через `subtle.ConstantTimeCompare`.

### Сетевая безопасность и SSRF

- Все обращения к user-controlled URL — **только** через `safehttp.NewSafeHTTPClient` + `ValidateOutboundURL` (запрет private/link-local адресов). Применяется к GitLab-интеграции, OpenGraph-превью, Web Push endpoints.
- `http.DefaultClient` для внешних запросов запрещён; чтение тел — только через `io.LimitReader`.

### Целостность и доступ

- **Timing-safe сравнения** для всех секретов (`subtle.ConstantTimeCompare` / `hmac.Equal`) — сравнение через `!=` запрещено.
- **Подписи webhook обязательны** (LiveKit sha256-claim, GitHub, GitLab): отсутствие подписи/claim = отказ, «skip if absent» запрещён.
- **RBAC и user-scope везде:** подписка на WS-комнаты — только после проверки membership (`CanSubscribe`, fail-closed + Origin-check); поиск (Meilisearch/Mongo) всегда фильтруется по каналам/проектам пользователя — глобальной выдачи не существует.
- **Шифрование токенов интеграций** (GitHub/GitLab) — AES-GCM at rest.

### Секреты

- **Секреты только в env / K8s Secrets** — `.env` в `.gitignore`, в репо только `.env.example` без значений. Hardcoded-credentials в конфигурациях запрещены.
- CI-секреты — GitHub Secrets, не в yaml. `DEV_MODE=true` в production CI/CD запрещён.

## Hardening-guide для self-hosted

Рекомендации для операторов инстанса Hub:

### TLS и периметр

- Весь внешний трафик — через **Caddy** (авто-HTTPS/TLS, HSTS). Никогда не выставляйте backend (:3000) напрямую.
- **Никогда не выставлять наружу:** PostgreSQL (5432), MongoDB (27017), Redis (6379), NATS (4222/8222), SeaweedFS (9333/8333), LiveKit admin-порты. Наружу смотрит только Caddy (80/443).
- В K3s — **NetworkPolicy**: инфраструктурные StatefulSets доступны только сервисам своего namespace.

### Секреты и ротация

- Хранение: env-файлы вне репозитория или K8s Secrets; для GitOps — шифрование через **SOPS**.
- **Ротация API-ключей и интеграционных токенов — каждые 90 дней** (Admin panel → API keys). Скомпрометированные токены (в т.ч. попавшие в git-историю) ротировать немедленно (runbook FA-06).
- JWT-секрет и LiveKit-ключи — уникальные per-инстанс, генерировать криптостойко (не dev-значения из примеров).

### Обновления и образы

- Все Docker-образы — **pinned версии** (`postgres:17`, `mongo:8.0`, `redis:7.4-alpine`, `nats:2.11`, `chrislusf/seaweedfs:3.97`, `livekit/livekit-server:v1.8`); тег `latest` запрещён.
- Деплой — только после green CI (`workflow_run` gate); production-бинарник — `go build -tags release` (dev-маршруты исключены).
- Следите за Dependabot-обновлениями (weekly go/npm/actions).

### Бэкапы

- Регулярные бэкапы PostgreSQL, MongoDB и S3-бакетов SeaweedFS (встроенный backup-модуль + внешнее хранилище).
- Проверяйте восстановимость: бэкап без теста восстановления — не бэкап.
- Корзина (soft-delete) хранит данные 30 дней; retention-cron и GDPR-каскады мониторить по WARN-логам.

## CI security

| Проверка | Инструмент | Статус |
|----------|-----------|--------|
| Уязвимости зависимостей Go | `govulncheck` + `go mod verify` | Волна 2 (BP-2026), Dependabot уже активен |
| Сканирование секретов | `gitleaks` | Волна 2 (BP-2026) |
| Race conditions | `go test -race` (все тесты) | ✅ в CI |
| Coverage security-пакетов | auth / ws / rbac / audit / calls / middleware **≥ 80%** | ✅ 6/6 (gate в CI) |
| Общий coverage-gate | backend ≥ 82% | ✅ 83.0% |
| Lint frontend | Biome (blocking gate) | ✅ в CI |

## Ссылки

- [docs/security_guidelines.md](docs/security_guidelines.md) — OWASP Top 10 чек-лист, статусы всех findings (12/12 CRITICAL fixed), grep-паттерны аудита
- [docs/tz_constraints.md](docs/tz_constraints.md) — mandatory/forbidden инварианты ТЗ
- [.planning/decisions/](.planning/decisions/) — ADR (auth, микросервисы, SSO)
- [.planning/audits/code-review-final/SECURITY.md](.planning/audits/code-review-final/SECURITY.md) — полный security-аудит
- [LICENSE](LICENSE) — проприетарная лицензия

---

*Обновлено: 2026-07-03 (DOCS-EXCELLENCE, DOC-03)*
