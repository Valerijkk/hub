# OWNER QUESTIONS — 14 `[default]` gray-zone decisions (D4) + GDPR/wiki (D5)

> Источник: `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`  
> Батчи 1–5 — явные ответы владельца; батч 6 + три `[default]` из батчей 3/5 — **рекомендованные дефолты без явного подтверждения**.  
> Дата сверки с кодом: 2026-07-27. Только документация; код не менялся.

**Как отвечать:** для каждого Q — `CONFIRM` (оставить default) / `OVERTURN` (указать вариант) / `DEFER` (с датой/триггером).

---

## A. GDPR (GZ-SEC-17 / 18 / 19) — первоочередно

### Q1 — GZ-SEC-17: единая корзина 30д для новых типов

| | |
|---|---|
| **Default** | Единый trash 30д; удаление wiki-страницы **каскадит на детей**; relations в bases — **tombstone**; HR не исключён из корзины. |
| **В коде сейчас** | Корзина/soft-delete есть точечно: admin trash (`audit` ListTrash/Restore + cron `trash-purge-expired`), projects trash, CRM `deleted_at` на contact/deal, messenger soft-delete, storage soft-delete. **Нет** wiki-схем/`backend/internal/wiki`. HR `LeaveRequest` **без** `deleted_at`. Единого каскада «поддерево → trash» нет (спека FR-WIKI-03 описывает поведение, которого нет). |
| **Если overturn** | Per-module retention / HR без корзины → переписать NFR-11 scope, ERD HR/bases, `PurgeExpired`, UX восстановления; отложить wiki-каскад до появления модуля. |

**Вопрос владельцу:** подтверждаете единый trash+каскад wiki / tombstone bases, или HR/иерархии получают отдельные политики?

---

### Q2 — GZ-SEC-18: право на забвение (анонимизация + NATS)

| | |
|---|---|
| **Default** | Анонимизация авторства контента + hard-delete персональных записей; оркестрация через NATS `user.gdpr_delete` (подписчики per-service). Admin-endpoint для CRM-контактов (субъекты-неюзеры). |
| **В коде сейчас** | `DELETE /api/admin/users/:id/gdpr` → `AuditService.DeleteUserGDPR`: hard-delete user + каскад comments/reactions/sessions/activity/time_logs + reassign issues. **Нет** NATS-события GDPR-каскада; **нет** анонимизации авторства wiki/bases; CRM contact — обычный soft-delete владельцем, не отдельный admin-GDPR для субъекта данных. |
| **Если overturn** | Полный hard-delete везде → ломает историю документов/проводок; «только hub-core» → дыры в v4/microservice-модулях. Нужен явный список сущностей в каскаде. |

**Вопрос владельцу:** подтверждаете anonymize+NATS fan-out, или другой контракт erasure до появления wiki/split?

---

### Q3 — GZ-SEC-19: legal-hold HR / accounting

| | |
|---|---|
| **Default** | Legal-hold: accounting/HR исключаются из GDPR-каскада на срок обязательного хранения; связи анонимизируются. |
| **В коде сейчас** | Флага legal-hold / exempt-from-GDPR **нет**. `DeleteUserGDPR` не знает journal/invoice/leave. Accounting (`journalentry`, `invoice`, …) и HR leave живут без retention-policy в коде. |
| **Если overturn** | «Удаляем всё» → риск для RU self-hosted; «только юрист потом» → нельзя закрывать accounting/HR erasure в реализации. |

**Вопрос владельцу:** фиксируем legal-hold в продукте, перекладываем на клиента, или стоп до юрконсультации?

---

## B. HR / CRM sensitive data (GZ-SEC-13 / 14)

### Q4 — GZ-SEC-13/14: модульные роли + field-masking + AES-GCM зарплат/документов

| | |
|---|---|
| **Default** | Роли `hr:admin`, `crm:manager` + field-level masking; **AES-GCM** для зарплат и HR-документов at rest. |
| **В коде сейчас** | AES-GCM есть для integration/BYOK (`platform/crypto/aesgcm.go`, GitHub/GitLab/AI keys) — **не** для HR/payroll. Схем зарплаты/паспорта нет; `LeaveRequest` — plaintext reason. Ролей `hr:admin`/`crm:manager` в rbac **нет** (system_role / project permissions). Field-masking API для CRM amount — **нет**. |
| **Если overturn** | Только disk/PG encryption → проще, слабее при утечке БД; без модульных ролей — любой Member видит HR/CRM (повтор bases-класса багов). |

**Вопрос владельцу:** подтверждаете app-level AES-GCM + отдельные data-роли до углубления HR/CRM, или достаточно Postgres/диска + существующий RBAC?

---

## C. Остальные `[default]` (10)

### Q5 — GZ-SEC-05/12: секреты K3s = K8s Secrets + SOPS/age

| | |
|---|---|
| **Default** | K8s Secrets + SOPS/age-шифрованные манифесты (GitOps). |
| **В коде/инфре** | Prod — compose + env; K3s почти пуст (`networkpolicy-metrics.yaml`). **SOPS/age в репо нет.** Docs/SECURITY упоминают SOPS как целевое. |
| **Если overturn** | sealed-secrets / Vault / plain Secrets вручную → другой M0 devops-план и GitOps-модель. |

**Вопрос:** подтверждаете SOPS/age на горизонт M0–M2?

---

### Q6 — GZ-SEC-13/15: S2S = static bearer + NetworkPolicy deny-by-default

| | |
|---|---|
| **Default** | Static bearer per service из K8s Secrets + NetworkPolicy deny-by-default (M0–M2); mTLS позже. |
| **В коде/инфре** | Прецедент: `mail/internal_auth.go` (`X-Hub-Internal-Token` / Bearer). Полного S2S mesh нет (монолит). NetworkPolicy — только optional metrics scrape. |
| **Если overturn** | mTLS/mesh сразу → тяжелее M0; JWT service-accounts → связка с JWKS (GZ-ARCH-13). |

**Вопрос:** static bearer + NP на пилот — ОК?

---

### Q7 — GZ-SEC-04 (ops): регулярная ротация 90 дней

| | |
|---|---|
| **Default** | API keys + внешние токены — **90 дней**; инфраструктурные (JWT/DB/NATS/LiveKit) — при инцидентах. |
| **В коде** | API keys: optional `expires_at`, ручное создание — **нет** cron/policy «каждые 90 дней». FA-06 — разовая ручная ротация. |
| **Если overturn** | Без расписания → оставить manual-only; короче/длиннее TTL → правка admin UX + runbooks. |

**Вопрос:** вводим 90-дневную политику для API/внешних токенов?

---

### Q8 — GZ-SEC-14 (security): внешние поверхности

| | |
|---|---|
| **Default** | Только **guest-звонки** по одноразовой ссылке с lobby; status/wiki/bases public — вне scope. |
| **В коде** | `POST /api/calls/guest-join` (invite_code + guest_name → LiveKit guest JWT). **Lobby/admit** не найден. Публичной status page / public wiki / public bases share — нет (bases share-link — backlog по V4). |
| **Если overturn** | + status page / public wiki/forms → отдельный security-review, CSP/CORS, rate-limit анонимов. |

**Вопрос:** guest-calls only до первого внешнего клиента — ОК? Нужен ли lobby до prod?

---

### Q9 — GZ-SEC-26: алерты → Telegram админа + uptime-ping

| | |
|---|---|
| **Default** | Telegram-канал админа: page = downtime/5xx/GDPR-FAIL; дайджест = retention/cert; + внешний uptime-ping. |
| **В коде/инфре** | Alertmanager → UI (dev) / webhook на backend (prod example). Telegram — notification bind пользователя, **не** ops paging-канал. Внешнего uptime-ping в репо нет. |
| **Если overturn** | Email-only / PagerDuty / только Alertmanager UI → другой runbook on-call. |

**Вопрос:** Telegram admin channel + внешний ping — целевой prod-алертинг?

---

### Q10 — GZ-SEC-27: NF-09 и accepted-риски перед prod

| | |
|---|---|
| **Default** | **NF-09 закрыть** S3-policy / size limit **до prod**; остальные accepted (NF-10, DEV_MODE WS-Origin, BE-H1) подтвердить. |
| **В коде** | Presigned PUT без `ContentLength`/max — всё ещё **accepted** (storage service; TZ NF-09). Fiber BodyLimit 25 MB не закрывает direct PUT. |
| **Если overturn** | Оставить NF-09 accepted в prod → явный residual risk в PRODUCTION-HANDOFF; закрытие → bucket policy / Seaweed quota / presign constraint. |

**Вопрос:** NF-09 — блокер v2.5.0 cutover или accepted с документированием?

---

### Q11 — GZ-SEC-10 (ops): AD-группы → только system role

| | |
|---|---|
| **Default** | AD/IdP groups → только **system role** (Admin/Member); project-роли назначаются внутри Hub. |
| **В коде** | SSO/LDAP **не реализованы** (ADR-AUTH-SSO есть как решение, код — Magic Link only). Маппинга групп нет. |
| **Если overturn** | Авто-маппинг project roles → сложнее sync/конфликты с RBAC; влияет ADR-AUTH-SSO. |

**Вопрос:** подтверждаете «groups → system role only» до реализации SSO?

---

### Q12 — GZ-ARCH-04: search остаётся агрегатор в ядре

| | |
|---|---|
| **Default** | Search-агрегатор в ядре; шаг «search» в roadmap = стандартизация indexer, не вынос сервиса. |
| **В коде** | `backend/internal/search` в монолите; Meili indexer опционален. Отдельного hub-search нет. |
| **Если overturn** | hub-search сервис → новая фаза Track P, CI/path-filter, GDPR/search user-scope в отдельном деплое. |

**Вопрос:** search не выносим в горизонте M0–M6 — ОК?

---

### Q13 — GZ-ARCH-08: миграции = per-service initContainer

| | |
|---|---|
| **Default** | Каждый сервис применяет свои Atlas/Ent миграции в initContainer; схемы независимы. |
| **В коде** | Монолит: единые Ent-миграции при старте. initContainer-паттерна per-service ещё нет. |
| **Если overturn** | Центральный migration-job → единая очередь/владелец; ручной gate devops — медленнее. |

**Вопрос:** per-service initContainer — канон на split?

---

### Q14 — GZ-REL-05: rolling + rollback по тегу; canary отложить

| | |
|---|---|
| **Default** | Rolling update + быстрый rollback по тегу; blue-green/canary — после реального внешнего трафика. |
| **В коде/инфре** | Compose/GHCR tag deploy; K3s Deployment strategy / canary **не** описаны. |
| **Если overturn** | Canary/blue-green сейчас → infra + CI минут (студенческий пакет) + сложность rollback. |

**Вопрос:** canary откладываем до внешнего трафика — ОК?

---

## D5 — GDPR defaults и отсутствующий wiki

### Что в GZ-SEC-17/18/19 становится невалидным без модуля wiki

Спека (`docs/system_specification.md`, FR-WIKI-01…09) и default-решение опираются на артефакты, которых **нет** в `backend/internal/` и `backend/ent/schema/`:

| Фрагмент default | Почему невалиден / пуст сейчас |
|---|---|
| «wiki каскадит» при удалении в корзину (SEC-17) | Нет page/space/version; нечего каскадить. FR-WIKI-03 не на что натянуть. |
| «анонимизация авторства wiki-правок» (SEC-18) | Нет истории версий страниц (FR-WIKI-06/07); некого anonymize. |
| Restrict/space RBAC в GDPR-restore (смежно SEC-07/11) | Нет space membership — admin GDPR «служебный доступ к wiki» из спеки мёртв. |
| Search `types=wiki_page` в erasure/hide (FR-WIKI-09) | Индекса wiki нет; GDPR hide-from-search для wiki — no-op. |
| Handbook = wiki space (GZ-V4-13) | HR handbook ждать wiki — блокер паритета, не GDPR-кода, но тот же gap. |

**Что остаётся валидным без wiki:** legal-hold accounting/HR (SEC-19); единый trash для **существующих** типов с `deleted_at`; идея NATS `user.gdpr_delete` для уже живых модулей (crm/bases/hr leave — когда добавят подписчиков); admin CRM-contact erasure (когда сделают endpoint).

### Что делать, пока wiki нет

1. **Снять wiki из исполнимого GDPR-default:** пометить SEC-17/18 wiki-ветки как `BLOCKED: waits FR-WIKI implementation` (не реализовывать каскад «впрок»).
2. **Подтвердить владельцем Q1–Q3** на **текущем** периметре: user GDPR, CRM contacts, bases records, HR leave, accounting journals — с/без legal-hold.
3. **Не обещать в STATE/roadmap 100% GDPR-паритет спеки**, пока FR-WIKI не в ROADMAP с фазой.
4. **При появлении wiki:** первым коммитом схемы — `deleted_at` + subtree trash (FR-WIKI-03) + anonymize author on `user.gdpr_delete` + wire в тот же NATS-контракт, что выберет Q2.
5. **До wiki:** handbook (если нужен) — либо временные страницы в intranet/HR, либо явный defer (уже заложено в V4-решении «wiki раньше HR-волны»).

---

## Сводка для владельца (чеклист)

| # | ID | Тема | Приоритет |
|---|-----|------|-----------|
| Q1 | GZ-SEC-17 | Trash 30д / wiki cascade / bases tombstone | P0 legal/product |
| Q2 | GZ-SEC-18 | Anonymize + NATS GDPR fan-out | P0 legal |
| Q3 | GZ-SEC-19 | Legal-hold HR/accounting | P0 legal |
| Q4 | GZ-SEC-13/14 | HR/CRM roles + AES-GCM fields | P0 security |
| Q5 | GZ-SEC-05/12 | SOPS/age | P1 infra |
| Q6 | GZ-SEC-13/15 | S2S bearer + NP | P1 infra |
| Q7 | GZ-SEC-04 | Rotation 90d | P1 ops |
| Q8 | GZ-SEC-14 | Guest-only surfaces (+ lobby?) | P1 security |
| Q9 | GZ-SEC-26 | Telegram + uptime alerts | P2 ops |
| Q10 | GZ-SEC-27 | Close NF-09 before prod? | P0 prod gate |
| Q11 | GZ-SEC-10 | AD → system role only | P2 (pre-SSO) |
| Q12 | GZ-ARCH-04 | Search stays in core | P2 arch |
| Q13 | GZ-ARCH-08 | Per-service migrations | P2 arch |
| Q14 | GZ-REL-05 | No canary yet | P2 release |

Связанные аудит-находки: `00-PLANNING-AUDIT.md` §5–6, `01-PLANNING-FIXES.md` C1 / B5.
