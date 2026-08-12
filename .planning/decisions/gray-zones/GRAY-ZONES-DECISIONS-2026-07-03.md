# Hub — Решения по серым зонам (зафиксировано 2026-07-03)

> Источник вопросов: `.planning/decisions/gray-zones/AREA-*.md` (6 областей, ~90 зон).
> Батчи 1–5 отвечены пользователем явно; батч 6 (GDPR/ops-детали) — приняты рекомендованные дефолты, помечены `[default]`, пересматриваются по запросу.
> Этот документ — вход для: ADR-детализаций, финального roadmap, правок правил (rule 50/56) и планов tech-lead.

---

## 1. Релиз, git-процесс, качество (батч 1)

| Зона | Решение |
|------|---------|
| GZ-SEC-01 ротация FA-06 | **Пользователь выполняет сам по runbook** в ближайшие дни; агент готовит чек-лист и обновляет env/Secrets |
| GZ-SEC-02 + GZ-REL-01 история/тег | **Только ротация, BFG-purge не делаем** (репо приватный); тег v2.5.0 после ротации |
| GZ-REL-09/07 коммиты/мерж | **Двухуровневая политика (уточнено):** агенты свободно коммитят в фича-ветки и пушат их на remote; **push/мерж в master выполняет только пользователь лично** — агенты в master не пушат никогда. → правка `50-git.mdc` |
| GZ-REL-06/08 PR/мерж-метод | **PR = фича/фаза** с атомарными коммитами внутри; **rebase-merge** (линейная история, откат по плану сохранён); починить «main»→«master» в правиле |
| GZ-REL-10 личное ревью | **Только критичное:** security/auth, миграции БД, деплой/CI, ADR. Остальное — цепочка агентов |
| GZ-REL-11 лимит агентов | **Мягкий:** дефолт 4, до 6–8 по решению team-lead при явной карте непересекающихся файлов |
| GZ-REL-14 дистрибуция | **Пока внутренний/проприетарный**; решение о дистрибуции — отдельным ADR при первом клиенте |
| GZ-QA-13 регрессия в волнах | **Агент: сборка + тесты своего домена; полный /qa-gate — один раз на интеграции волны + CI на merge.** → правка rule 56 |
| GZ-QA-01/02 coverage floor | **≥82% сразу для всего нового кода; MCP/agents ≥80% как security-critical** |

## 2. SSO и AI-агенты/MCP (батч 2)

| Зона | Решение |
|------|---------|
| GZ-SEC-03 протоколы SSO | **OIDC + LDAP/AD bind** в v1; SAML позже |
| GZ-SEC-04/08 SSO×MagicLink | **Org-политика magic-link/sso/both; при enforce-SSO Magic Link = break-glass для founder.** Меняет MANDATORY ТЗ → фиксируется в ADR-AUTH-SSO |
| GZ-SEC-05/09 провижининг | **JIT при первом входе + периодический LDAP-sync деактиваций;** invite остаётся для non-SSO |
| GZ-SEC-06 MFA | **Без собственного MFA** — email-фактор + MFA на стороне IdP |
| GZ-SEC-09/20 автономия AI | **Ступенчатая:** read+создание/черновики автономно (в правах пользователя); delete/чужие задачи/сообщения от имени/RBAC — с подтверждением |
| GZ-SEC-10/21 идентичность AI | **Отдельный agent-principal;** права = агент ∩ пользователь; audit: user + agent + task-id |
| GZ-SEC-22 BYOK | **AES-GCM хранение + org-дневные лимиты токенов с hard-stop и алертом** |
| GZ-SEC-23 MCP | **PAT со скоупами + курируемый read-heavy toolset в v1** |
| GZ-SEC-07/11 wiki RBAC | **Space-level права (read/write/admin) + наследование + restrict-флаг на страницу** |
| GZ-SEC-08/12 bases sharing | **База+таблица (read/edit/manage), view наследует; только project-membership в v1** |

## 3. Микросервисы, БД, deployment (батч 3)

| Зона | Решение |
|------|---------|
| GZ-ARCH-01 состав realtime | **ТРИ сервиса: hub-ws / hub-messenger / hub-calls** (не единый hub-realtime). ⚠️ Отличается от ADR §3 → ADR обновлён |
| GZ-ARCH-02/03 tms/tracker | **tms = условная фаза M6** (только после Go по пилоту); **tracker — явно исключён из горизонта**, возврат отдельным ADR |
| GZ-ARCH-05 Go/No-Go | **Композит:** lead-time деплоя ↓, p95 не хуже +10%, 0 инцидентов из-за split за 2 недели |
| GZ-ARCH-06 драйвер split | **Масштабирование → ограничиться hub-ai + realtime-трио** (+ пилот notification как обучение шаблону). Полный распил не самоцель |
| GZ-ARCH-07 бэкапы | **Решить на пилоте M2** (отложено осознанно) |
| GZ-ARCH-09 сироты UUID | **Идемпотентный consumer на user.deleted/project.deleted + периодический reconcile-job** |
| GZ-ARCH-10/11 топология | **Одна K3s-нода + local-path PVC;** пересмотр при HA-потребности |
| GZ-ARCH-12 dev-профиль | **После приёмки каждого выноса дефолт = split для этого сервиса** (dev ≈ prod) |
| GZ-ARCH-13/14 JWKS/OTel | **HS256 допустим для пилота (JWKS обязателен до hub-ai); OTel-трейсинг через NATS — жёсткий gate до пилота** |
| GZ-ARCH-15 rate limits | **Слоями:** auth-лимиты в hub-core, доменные в своём сервисе, общий 100/с на Caddy |
| GZ-ARCH-16/17 CI/flips | **Path-filter обязателен в M0** (остаёмся в 3000 мин); переключения — Caddy route-flip, окно в минуты допустимо |
| GZ-ARCH-04 search | **[default]** Search-агрегатор в ядре; шаг «search» = стандартизация indexer |
| GZ-ARCH-08 миграции | **[default]** Per-service initContainer (схемы независимы) |

## 4. Границы паритета v4 (батч 4)

| Зона | Решение |
|------|---------|
| GZ-V4-22 порядок | **Bases-security-fix немедленно (S)** → intranet+calendar → crm+hr → accounting; bases-эпик параллельно на be-потоке; FE-декомпозиция по мере захода в модуль |
| GZ-V4-16 accounting класс | **Управленка:** баланс + P&L + cash flow по своему плану счетов, без регламентных форм |
| GZ-V4-17/18 accounting scope | **Волна 1: баланс/cash flow/PDF-инвойсы/AR aging;** поле валюты в схеме сразу (мультивалюта волной 2); зарплата — только проводки факта; **1C-sync расширяем как экспорт-мост** |
| GZ-V4-05 calendar sync | **iCal двусторонне-упрощённо:** export + подписка на внешние iCal-фиды, без OAuth |
| GZ-V4-04/06/07 calendar | Универсальные «ресурсы» без панелей; recurring daily/weekly/monthly + until/count (правка серии целиком); availability busy/free + отпуска |
| GZ-V4-01/02/03 intranet | Опрос-в-посте; 1–2 onboarding-шаблона с чекбоксами; постмодерация + kudos-лидерборд + coffee-матчинг раз в неделю с авто-DM |
| GZ-V4-08/09/10 CRM | Companies + лёгкая activity-сущность (timeline); импорт CSV/XLSX + пресеты amoCRM/Bitrix; email — ручное логирование + отправка через hub-mail |
| GZ-V4-11 CRM-аналитика | **Ждать BI-builder и делать на нём.** ⚠️ Следствие: BI-builder входит в backlog как зависимость CRM-аналитики |
| GZ-V4-12/13/14/15 HR | user.manager_id + авто-chart; handbook = пространство общей wiki (wiki раньше HR-волны); **авто-начисление отпусков (accrual policy) + поле грейда** (расширенный вариант); производственный календарь РФ; review-циклы — backlog |
| GZ-V4-19/20 bases | **Параметры PLAN-bases подтверждены:** 30 функций + 4 views + лимиты 50k/200; realtime в BS-5; attachment-поля да; share-link — backlog |
| GZ-V4-21 marketplace/status | **Заморозить до завершения P1-модулей** |

## 5. Качество/тесты/релизный процесс (батч 5)

| Зона | Решение |
|------|---------|
| GZ-QA-03 отстающие пакеты | **Совместить uplift с волнами переработки** (ai/bases переписываются); отдельный uplift не делать |
| GZ-QA-04/05 Lighthouse | **Baseline сейчас, gate «baseline−5» (анти-регрессия), ≥90 после UX-polish; 4–5 страниц на master push, login-only на PR** |
| GZ-QA-06/07 WCAG | **Критичные пути dark+light + axe в E2E на остальное; агентский sign-off, человек точечно** |
| GZ-QA-08/09 contract/E2E | **С пилота: JSON-схемы NATS-событий + schema-тесты с обеих сторон; PR = smoke+затронутый сервис, master = полный compose** |
| GZ-QA-10 нагрузочные | **Locust (Python)** — уточнение пользователя 2026-07-03 (вместо k6): baseline-профиль (WS/auth/board) локально/nightly **до старта выноса realtime**; позже интегрируется в Python-testframework (`testframework/load/`) фазы TF-PY |
| GZ-QA-11 flaky | **Retries=2 только E2E; 2 падения/неделя → quarantine с issue и SLA 1 волна; unit-flaky — немедленный фикс** |
| GZ-QA-12 coverage при split | **Per-service gates:** hub-core ≥82%, каждый сервис ≥80% (security жёстче); суммарную метрику убрать |
| GZ-QA-14 TC-таблицы | **Полные TC-таблицы сразу для всех новых модулей** (максимум трассируемости — выбор пользователя, больше нагрузка qa-manual) |
| GZ-QA-16 FE-coverage | **Мерить и публиковать в COVERAGE.md без blocking; gate после baseline** |
| GZ-REL-02/03/04 релизы | **Тег на milestone + промежуточные на крупные фичи; единая платформенная версия пока ≤6 сервисов; master всегда releasable, деплоим непрерывно** |
| GZ-REL-12/13/16 процесс | **Cross-cutting владелец — per-волна в PLAN; конфликт: автор-агент ребейзит + повторная регрессия; CHANGELOG автогенерация + notes на milestone** |
| GZ-REL-05 canary | **[default]** Rolling update + rollback по тегу; canary отложить |

## 6. GDPR/данные/ops (батч 6 — рекомендованные дефолты `[default]`, пересматриваются по запросу)

| Зона | Решение |
|------|---------|
| GZ-SEC-05/12 секреты K3s | [default] K8s Secrets + SOPS/age-шифрованные манифесты (GitOps) |
| GZ-SEC-13/15 S2S auth | [default] Static bearer per service из K8s Secrets + NetworkPolicy deny-by-default (M0–M2) |
| GZ-SEC-13/14 HR/CRM данные | [default] Модульные роли (hr:admin, crm:manager) + field-level masking; AES-GCM для зарплат/документов |
| GZ-SEC-17/18/19 GDPR | [default] Единая корзина 30д (wiki каскадит, relations tombstone); GDPR: анонимизация авторства + hard-delete персонального через NATS-событие; accounting/HR — legal-hold; admin-endpoint для CRM-контактов |
| GZ-SEC-14 внешние поверхности | [default] Только guest-звонки по одноразовой ссылке с lobby; остальное за auth |
| GZ-SEC-04 регулярная ротация | [default] API keys + внешние токены — 90 дней; инфраструктурные — при инцидентах |
| GZ-SEC-26 алерты | [default] Telegram-канал админа (page: downtime/5xx/GDPR-FAIL; дайджест: retention/cert) + внешний uptime-ping |
| GZ-SEC-27 accepted-риски | [default] NF-09 закрыть S3-policy лимитом до prod; остальные accepted подтвердить |
| GZ-SEC-10 маппинг групп | [default] AD-группы → только system role (Admin/Member); project-роли внутри Hub |

## 7. QA-конвейер (уточнение пользователя, 2026-07-03)

| Решение | Содержание |
|---------|-----------|
| Порядок QA строго последователен | **Бизнес-дока → qa-manual пишет TC → qa-manual-ver верифицирует тестовую доку → qa-auto1/2 автоматизируют верифицированные TC → qa-auto-ver верифицирует автотесты.** Автоматизация неверифицированной доки запрещена |
| Юнит-тесты | **Пишут разработчики (`be*`/`fe*`) вместе с кодом** — не qa-auto. Зона qa-auto: integration/API/компонентные/E2E поверх TC |
| Целевой тестовый фреймворк (решение 2026-07-03) | **После дробления на микросервисы:** юниты остаются с кодом (Go — бэк, TS — фронт); **всё остальное (integration/API/contract/E2E) переписывается на Python** (pytest + allure-pytest + httpx + playwright-python) в `testframework/`, раскидано по доменам/сервисам по best practices. Фаза TF-PY в ROADMAP-FINAL, Depends on: Track P M2–M4 |

Применено: `testing.md` §1.1/§1.3, `team-workflow.md` §3, `REGULATION.md` §5, агенты qa-manual/qa-manual-ver/qa-auto1/qa-auto2/be1/be2/fe1/fe2.

---

## Следствия для документов (исполнить)

1. **ADR-ARCH-MICROSERVICES-2026:** каталог — realtime как **три сервиса** (hub-ws/hub-messenger/hub-calls); scope — пилот notification + hub-ai + realtime-трио; tms условно; tracker исключён. ✅ обновлено 2026-07-03
2. **`50-git.mdc`:** двухуровневая политика коммитов; rebase-merge; main→master.
3. **Rule 56:** модульная регрессия у агента + полный gate на интеграции волны.
4. **ADR-AUTH-SSO (писать):** OIDC+LDAP, org-toggle, JIT+LDAP-sync, без MFA, маппинг только system role.
5. **ADR-AI-agents (писать):** ступенчатая автономия, agent-principal, BYOK hard-stop, MCP PAT+curated toolset.
6. **Backlog +BI-builder** (зависимость CRM-аналитики).
7. **Bases security-fix** — немедленная задача Волны 1.
8. **tz_constraints.md:** SSO из FORBIDDEN → «за ADR-AUTH-SSO»; правки rate-limit размещения при split.
