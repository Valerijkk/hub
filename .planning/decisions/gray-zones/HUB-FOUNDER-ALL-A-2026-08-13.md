# HUB-FOUNDER-ALL-A-2026-08-13 — canonical founder decision

**Date:** 2026-08-13

**Authority:** founder (product authority)

**Status:** LOCKED product decision; execution acceptance is separate.

**Decision:** the founder selected option **A for every question** in the
209-item master Hub questionnaire, with no exceptions.

## Provenance

The questionnaire source is the assistant message in
`C:\Users\Valerii\.codex\sessions\2026\08\13\rollout-2026-08-13T01-34-06-019ff81c-84df-7f60-b665-59d9e5bfa469.jsonl`:

- `response_item.timestamp=2026-08-13T18:06:48.687Z`;
- `payload.type=message`, `role=assistant`, `content[0].text`;
- UTF-8 SHA-256 of the complete source message text:
  `a9840534683a9b9558eadc40e2e094e7eb0c7c1a7b8323479629e23d12b0dbdf`.

The approval source is the founder's user message in the same transcript:

- `response_item.timestamp=2026-08-13T18:55:21.388Z`;
- `payload.type=message`, `role=user`;
- exact raw `content[0].text`, escaped: **`"на все вопросы а \n"`**
  (one trailing U+0020 space, then U+000A LF; 18 Unicode characters);
- SHA-256 of the exact raw UTF-8 reply bytes:
  `527b917699f0d381d442c7f5930b4a708accec46c4f5f324339638e2af9ffea4`;
- normalized display after removing that trailing space and LF:
  **`на все вопросы а`**.

The transcript establishes provenance, but this repository record is
self-contained: the complete approved Hub map, all approved premises, and all
209 exact selected-A wordings are embedded below.

Repository precedence was checked against the concrete current sources
`AGENTS.md`, `.planning/STATE.md`, `docs/README.md`,
`docs/vision_and_scope.md`,
`.planning/decisions/specs/PLATFORM-VISION-v4.md`,
`.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`, and
`docs/mail/FOUNDER-QUESTIONS.md`.

## Product horizons

- **Functional MVP of all approved modules by end of 2026.**
- **A genuinely accepted, coherent Hub 1.0 by June 2027.**
- Historical tags and historical `SHIPPED`/`COMPLETE` labels do not establish
  product, module, or production readiness.

## Canonical approved extract

### Reproducible normalization and digest

Two reproducible canonical extracts are maintained. The original
premises+register extract is reconstructed from this file as follows:

1. Read this Markdown as UTF-8.
2. Take, in order, the 16 lines strictly between
   `canonical-premises:start` and `canonical-premises:end`. Each is preserved
   verbatim, including its leading `- `.
3. Take, in order, the 209 lines inside the `text` fence strictly between
   `canonical-register:start` and `canonical-register:end`. Every line has
   exactly `ID<TAB>A wording`; the tab is U+0009.
4. Concatenate the 16 premise lines and then the 209 register lines with LF
   (U+000A), with no blank lines, no trailing spaces, no BOM, and one final LF.
   Markers, headings, and fence lines are excluded.

SHA-256 of the premises+register canonical UTF-8 bytes:

`5665899481be9f876b380355ba9de3f87884bed58b978906a7e5ad61d4368fbe`

The aggregate map+premises+register extract uses the same encoding and line
rules, but prepends, in order, the 10 lines strictly between
`canonical-map:start` and `canonical-map:end`: the two Markdown table header
lines followed by all eight block rows. It then appends the same 16 premises
and 209 register lines. There are 235 lines total, no blank lines, no trailing
spaces, no BOM, and one final LF. Markers, headings, and blank separator lines
are excluded.

SHA-256 of the aggregate map+premises+register canonical UTF-8 bytes:

`2d39635cb60a900e0cc3f11327227b5b75a7c2b596338f242886c32fa0c7617a`

### Предлагаемая карта Hub — exact source table

<!-- canonical-map:start -->
| Блок | Модули |
|---|---|
| Основа платформы | Identity & Directory, Roles & Access, Requests & Approvals, Admin, Security, Audit, eDiscovery, Files, Storage, Backup |
| Общие сервисы | Notifications, Search, API, Webhooks, Integrations, Import/Export, Observability, Status |
| Разработка | Projects, Teams, Tracker, Kanban, Scrum, Roadmap, TMS, Allure, GitHub/GitLab, CI visibility, Time, Automation |
| Коммуникации | Messenger, Calls, Mail, Calendar |
| Компания и знания | Intranet, Wiki |
| Бизнес | CRM-lite, HR, Management Accounting, 1C Bridge, Bases, Marketplace |
| Аналитика | Dashboards, Reports |
| AI | AI Hub, assistants, agents, knowledge search |
<!-- canonical-map:end -->

### Previously accepted premises — exact source wording

<!-- canonical-premises:start -->
- Один self-hosted Hub-инстанс — одна организация.
- PWA, без обязательных нативных приложений.
- Hybrid RBAC + ABAC/relationships, deny-by-default.
- Два Super Admin; критические действия требуют обоих.
- Защищённые системные роли плюс настраиваемые шаблоны.
- Три уровня риска: обычный, чувствительный, критический.
- Security Admin активирует сотрудников, почту и сервисные проекции.
- Одна человеческая identity; сервисные аккаунты связаны с ней, отдельных паролей нет.
- JML: preboarding, mover reconciliation, немедленный revoke при увольнении.
- Админский доступ — JIT; eDiscovery — только через расследование.
- Админы Mail/Security не получают скрытого доступа к содержимому.
- Shared resources — отдельные identities, без общих паролей.
- Calendar по умолчанию показывает только free/busy; private всегда только busy.
- AI-агент — отдельный principal; права равны пересечению прав пользователя и агента.
- Почта: Stalwart sidecar, OAuth, отдельный origin, DNS-проверка без DNS-ключей.
- CalendarEvent является канонической сущностью для связанного созвона.
<!-- canonical-premises:end -->

### Complete register — exact selected A wording

The register is tab-separated. It contains only the selected A wording; rejected
alternatives are intentionally absent.

<!-- canonical-register:start -->
```text
MAP-01	утвердить карту выше
MAP-02	отдельный сквозной модуль для доступов, HR, финансов, расследований
MAP-03	к 1.0 сделать очередь внутренних IT/HR/Finance-заявок внутри Requests
MAP-04	общая библиотека вложений и командных файлов, без клона Google Drive
MAP-05	dashboards в MVP, простой report builder в 1.0
MAP-06	в MVP только email/calendar guests, в 1.0 ограниченные project/call guests
MAP-07	внутренние объявления сотрудников
MAP-08	отложить multi-org, payroll, statutory accounting, собственный CI, PSTN/SIP, public marketplace, full BI/BPMN
IAM-01	Hub identity плюс OIDC для внутренних/внешних сервисов; LDAP только при реальном заказчике после 1.0
IAM-02	passkey предпочтительно, OIDC и magic-link recovery; без паролей
IAM-03	HR владеет кадровыми данными, Security — активацией и доступом
IAM-04	sponsor, ограниченный scope, обязательная дата окончания
IAM-05	для админов dual control, для сотрудников Security+проверка личности
IAM-06	пользователь видит устройства и завершает сессии; критические действия требуют свежего MFA
IAM-07	workload identities с короткими токенами
IAM-08	рабочие поля видны организации, личные контакты скрыты/добровольны
GOV-01	типизированные шаблоны с владельцем, SLA и риском
GOV-02	автоматически из риска, ресурса и оргструктуры
GOV-03	запрещено всегда
GOV-04	временное, аудируемое и не обходит separation of duties
GOV-05	напоминание, эскалация и резервный согласующий
GOV-06	break-glass с причиной, коротким TTL и обязательным разбором
GOV-07	месячная критических и квартальная чувствительных прав
GOV-08	роли, effective permissions, источник права, срок и конфликтующие запреты на одном экране
SEC-01	Public/Internal/Confidential/Restricted
SEC-02	метка родителя задаёт минимум, дочерний объект может стать строже
SEC-03	централизованная policy с разрешёнными исключениями модулей
SEC-04	soft delete → anonymization/deletion после срока; legal hold блокирует
SEC-05	1 год обычный аудит, 3 года privileged/security/eDiscovery, настраиваемо вверх
SEC-06	TLS+disk encryption и отдельное field encryption для зарплат, документов и секретов
SEC-07	никаких silent overwrites; показать diff, drafts сохранять обеими версиями
SEC-08	отдельный approved case на export/delete/anonymize
SEC-09	governed campaigns с аудиторией, preview и аудитом
SEC-10	история, preview и rollback
FILE-01	файл наследует доступ сущности-владельца
FILE-02	нет в MVP; в 1.0 только expiring/revocable/password-protected
FILE-03	обязательные malware scan и quarantine до скачивания
FILE-04	версии важных документов, корзина 30 дней
FILE-05	mail 25 MB, обычный файл 250 MB по умолчанию, recordings отдельно; всё настраивается
FILE-06	org/project/user quotas с предупреждением
FILE-07	RPO около 15 минут для БД, nightly backup, RTO до 4 часов, quarterly restore drill
FILE-08	удаление ссылки не уничтожает файл, пока есть другие владельцы
NTF-01	отдельные Action required, Mentions, Updates, System
NTF-02	in-app канон; push/Telegram/webhook best-effort; email только для выбранных и критических сценариев
NTF-03	пользовательские, но security/emergency могут пробить
NTF-04	daily/weekly opt-in
SRC-01	все модули, строго в пределах effective permissions
SRC-02	редактированные безопасные фрагменты
SRC-03	exact search в MVP, permission-aware semantic в 1.0 opt-in
SRC-04	archived включается фильтром, trash исключён
API-01	scoped, expiring, rotatable service-principal tokens
API-02	подпись, idempotency, retries, dead-letter и ручный replay
API-03	Admin владеет секретом, owner привязывает разрешённый ресурс
API-04	preview+validation, затем atomic commit
API-05	понятный статус интеграции и безопасный retry
OPS-01	внутренний status всегда, внешний опционально
OPS-02	Investigating/Identified/Monitoring/Resolved
OPS-03	автоматика создаёт draft, человек подтверждает публичное сообщение
OPS-04	in-app плюс внешний Telegram/webhook для critical
OPS-05	rolling+tag rollback до 1.0, canary позже
PRJ-01	динамическое членство плюс direct overrides
PRJ-02	оргбиблиотека, clone при создании
PRJ-03	полностью read-only, интеграции перестают писать, export доступен
PRJ-04	нет в MVP, scoped Viewer в 1.0
TRK-01	системный набор, проект включает нужное подмножество
TRK-02	Project Owner настраивает поддерживаемые типы
TRK-03	relates/blocks при доступе к обоим, без cross-project parent
TRK-04	initiatives/read-only timeline в 1.0
AGL-01	общий backlog, Kanban всегда, Scrum включается настройкой
AGL-02	warning по умолчанию, hard limit opt-in
AGL-03	owner выбирает backlog или следующий sprint для незавершённых
AGL-04	разрешить, но явно маркировать и учитывать отдельно
TMS-01	TMS хранит test intent, автотест — stable ID и связь
TMS-02	QA создаёт один deduplicated bug из выбранных failures
TMS-03	advisory в MVP, opt-in required suites/failure budget в 1.0
TMS-04	deterministic flaky signal, run diff и trends
VCS-01	одинаковый базовый GitHub/GitLab контур
VCS-02	Admin владеет credential, Owner связывает repo
VCS-03	auto-transition только при явном mapping и зелёном required CI
VCS-04	связать с issue/PR, человек создаёт deduplicated triage/bug
VCS-05	health/status/retry, но Hub не становится CI-системой
TIME-01	estimate-vs-actual для delivery, не payroll/billing
TIME-02	один активный таймер на пользователя плюс ручные логи и audit correction
AUTO-01	curated trigger/action catalogue
AUTO-02	одно объяснимое действие на правило до отдельного расширения
IMEX-01	переносимый project package плюс отдельный полный backup
MSG-01	public/private/project/announcement
MSG-02	создаёт Admin, публикует разрешённая роль
MSG-03	окно 24 часа, история edit видима, delete оставляет tombstone
MSG-04	полноценные threads и создание issue из сообщения с backlink
MSG-05	индивидуальные в DM/group DM, агрегат в больших каналах
MSG-06	добровольный статус, никакой аналитики «продуктивности»
MSG-07	org policy плюс одобренные исключения
MSG-08	новый админ не получает ретроспективный доступ без approved case
MSG-09	после 1.0
CALL-01	ad-hoc huddle из DM/channel и scheduled call из Calendar
CALL-02	обычное событие допускает email guest, linked call ждёт безопасную expiring/revocable ссылку
CALL-03	внешний гость всегда через lobby, host admit/deny
CALL-04	default off, явное согласие/баннер, доступ только участникам/проекту
CALL-05	90 дней по умолчанию, configurable/legal hold
CALL-06	captions в MVP при наличии, transcript+AI summary opt-in в 1.0
CALL-07	одна deduplicated overdue occurrence, затем нормальный график
CALL-08	до 50 участников, screen share, reactions, recording; без PSTN/SIP
CALL-09	честно заявлять transport/storage encryption, не E2EE при серверной записи
MAIL-01	только вручную Admin после успешных DNS/domain gates
MAIL-02	внутренняя Hub-почта работает независимо
MAIL-03	сохранить подходящий уникальный login local-part, иначе стабильный нейтральный адрес; aliases админские
MAIL-04	plain-safe MVP, ограниченный rich HTML после vetted sanitizer
MAIL-05	типизированные Hub rules, без raw Sieve/scripts
MAIL-06	MVP с read/send-as/send-on-behalf раздельно
MAIL-07	Inbox/Sent не удалять автоматически; Trash/Spam 30 дней; legal hold сильнее
MAIL-08	настраиваемая, предупреждение заранее, graceful receive-only перед полным отказом
MAIL-09	OOO в 1.0, read receipts privacy-respecting opt-in
MAIL-10	выключен по умолчанию, allowlist/approved case
MAIL-11	все folders, bounded sanitized snippets и attachment filenames
MAIL-12	один primary domain в MVP, несколько verified domains в 1.0
MAIL-13	показывать delivery unknown и сверять состояние до retry
MAIL-14	сохранять обе версии и предложить сравнение
CAL-01	personal/project/shared/resource плюс calls, sprints, deadlines и leave как типизированные источники
CAL-02	explicit Viewer/Editor/Delegate, по умолчанию free/busy
CAL-03	показывает свободные интервалы, working hours, timezone и ресурсы без раскрытия деталей
CAL-04	organizer видит индивидуальные ответы, остальные агрегат
CAL-05	member запрашивает, manager решает conflict/blackout
CAL-06	export+read-only iCal до 1.0, OAuth write sync позже
CAL-07	обязательный выбор this occurrence/following/series
CAL-08	durable exactly-once in-app, push/Telegram best-effort, без email fallback
CAL-09	напоминание от первого существующего локального времени дня с учётом DST
CAL-10	пользовательские часы, timezone и focus blocks
CAL-11	busy без причины, HR/manager видят необходимую детализацию
CAL-12	только после безопасной revocable credential
INT-01	хронологический editorial feed, pinned posts, без алгоритмического ранжирования
INT-02	обычные posts по policy, organization news через publisher/moderation
INT-03	для policies/security announcements можно требовать acknowledgement
INT-04	opt-in match, DM создаётся после согласия хотя бы одного участника
INT-05	публично по умолчанию, можно private-to-recipient
INT-06	org-wide MVP, department/location/role targeting в 1.0
WIKI-01	Markdown-compatible rich editor
WIKI-02	team spaces создают Members, handbook/org spaces — Admin
WIKI-03	restriction только сужает доступ space
WIKI-04	owner, review date и verified badge для handbook/runbooks
WIKI-05	версии, diff и restore
WIKI-06	optimistic conflict UI в MVP, realtime co-edit после 1.0
WIKI-07	нет до 1.0, использовать approved export
CRM-01	companies, contacts, deals, activities, tasks, meetings
CRM-02	стандартный pipeline с admin-editable stages
CRM-03	individual owner плюс team visibility
CRM-04	suggestions+manual merge с audit
CRM-05	только явно linked message после подтверждённого send/receive
CRM-06	consent/do-not-contact/retention поля обязательны для внешних контактов
CRM-07	типизированные follow-up/reminder/stage rules
CRM-08	CRM остаётся внутренним до после 1.0
HR-01	Employee record является source of truth жизненного цикла
HR-02	org chart, leave, onboarding/offboarding, grade, documents, RU calendar
HR-03	manager approves, HR override/audit
HR-04	одна authoritative manager chain, dotted-line как metadata в 1.0
HR-05	коллеги видят отсутствие, не причину
HR-06	Restricted, field encryption, строго HR/Finance grants
HR-07	upload, versioning и acknowledgement; юридическую e-signature не обещать
HR-08	после 1.0
HR-09	grade/history в 1.0, review cycles после 1.0
HR-10	только факты/exports в Accounting/1C, без payroll engine
FIN-01	management accounting, не statutory ledger
FIN-02	Hub для operational facts/budgets, 1C для regulated accounting
FIN-03	outbound approved facts сначала, reconciliation до inbound
FIN-04	posted fact не редактируется, только reversal/correction
FIN-05	draft/approved/sent/part-paid/paid/overdue/cancelled
FIN-06	одна базовая валюта MVP, multi-currency в 1.0
FIN-07	project/cost-center budget, plan-vs-actual
FIN-08	пороги суммы и risk-tier маршруты
FIN-09	агрегированные budget/margin показатели без зарплат и ledger details
FIN-10	CSV/import и ручное подтверждение в 1.0
BASE-01	grid/kanban/calendar/gallery
BASE-02	typed fields, relations, lookup, formulas
BASE-03	base/table permissions плюс field masking; row-level rules после 1.0
BASE-04	Manage schema отдельно от Edit records
BASE-05	internal forms в MVP, expiring external forms в 1.0
BASE-06	curated triggers/actions, один action на правило
BASE-07	record history и conflict diff
BASE-08	50k records/base и 200 fields/table как управляемые defaults
BASE-09	export/import boundary до 1.0
BASE-10	scoped API/webhooks через service principal
MKT-01	внутренние объявления/вещи/услуги сотрудников
MKT-02	никаких платежей, escrow и споров
MKT-03	через Messenger, без публикации личного телефона
MKT-04	reports, moderator queue, expiry и sold/reserved
MKT-05	вся организация MVP, approved department targeting в 1.0
AI-01	Admin-managed cloud/local providers; personal BYOK после 1.0
AI-02	Ollama/OpenAI-compatible endpoint поддерживается как privacy-first option
AI-03	Admin allowlist, пользователь выбирает разрешённую модель
AI-04	permission-aware retrieval на момент запроса, citations и uncertainty
AI-05	Hub-данные не используются для обучения provider; sensitive data cloud-off by default
AI-06	настраиваемая история, Restricted prompts off by default, audit хранит metadata/hash
AI-07	org/team/user quotas, warnings и hard stop
AI-08	allowlist, risk tier, confirmation, rate limit и kill switch
AI-09	search/summarize/draft mail/wiki/issues/tests, без автономных критических изменений
AI-10	opt-in agents для утверждённых повторяемых процессов
AI-11	показывать uncertainty, sources и действия агента
AI-12	user+agent+task+model+tool+input class+outcome
DASH-01	role defaults плюс personal/project customization
DASH-02	каждая показывает источник и freshness
DASH-03	saved filters, таблица/chart, schedule/export в 1.0
UX-01	responsive PWA с install/offline shell
UX-02	cached read и drafts для Mail/Tracker/Wiki; без offline admin/finance mutations
UX-03	русский основной, английский полностью к 1.0
UX-04	WCAG 2.2 AA как release gate 1.0
UX-05	единый search/command palette, recent и favorites
UX-06	role-based стартовый экран и checklist
UX-07	понятное состояние, retry/reconcile, correlation ID для поддержки
UX-08	risky/unfinished capability выключена по умолчанию и видна Admin
```
<!-- canonical-register:end -->

## Precedence and supersession

This decision is the current product-choice baseline. Its precedence rule is
narrow:

1. A newer explicit selected-A statement above supersedes an older source only
   for the same overlapping product choice.
2. Non-overlapping constraints, accepted ADR/spec technical semantics, security
   invariants, and evidence gates continue to apply.
3. A more detailed implementation contract may refine an A choice but cannot
   weaken, reverse, or silently extend it.
4. Historical documents remain historical evidence and are not rewritten by
   this record.

Concrete overlap ledger:

| Older source or decision | Exact new A ID | Scope and effect |
|---|---|---|
| `.planning/decisions/specs/PLATFORM-VISION-v4.md`: SSO/LDAP excluded from v1–v4 and LDAP marked Future 999.x | IAM-01, IAM-02 | Superseded for product scope: Hub identity plus OIDC is approved; LDAP is conditional on a real customer after 1.0; passkey is preferred and OIDC/magic-link is recovery, without passwords. Implementation and authentication assurance still require accepted contracts. |
| `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`, GZ-SEC-03: OIDC plus LDAP/AD bind in v1 | IAM-01 | Superseded only for LDAP timing: OIDC remains; LDAP moves to a real-customer need after 1.0. |
| Same file, GZ-SEC-06: no Hub MFA, email factor plus IdP MFA | IAM-02, IAM-06 | The new outcome is controlling: passwordless entry and fresh MFA for critical actions. The exact authenticator/provider implementation is not inferred here. |
| Same file, GZ-ARCH-02/03 and its consequences: Tracker excluded from the horizon; TMS conditional | MAP-01, TRK-01–TRK-04, TMS-01–TMS-04 | Superseded for product scope: Tracker and TMS are approved modules and belong to the approved-module MVP horizon. Service extraction and rollout order remain separate architecture/planning choices. |
| Same file, GZ-V4-04/06/07: recurring edits only for the whole series | CAL-07 | Superseded for recurrence UX: the user must choose this occurrence, following, or series. Existing recurrence safety/oracle requirements remain. |
| Same file, GZ-V4-01/02/03: weekly coffee matching with automatic DM | INT-04 | Superseded for consent: matching is opt-in and the DM is created only after at least one participant consents. |
| Same file, GZ-SEC-14 default: guest calls are the only external surface | MAP-06, CALL-02, CALL-03, CAL-12 | Superseded for product scope: MVP includes email/calendar guests; linked-call guests remain fail-closed until a safe revocable credential exists and external guests use a lobby. |
| Same file, GZ-V4-21: Marketplace frozen until P1 modules complete | MAP-07, MKT-01–MKT-05 | Superseded only if read as removal from the approved-module horizon. Internal employee classifieds are approved; sequencing may remain gated if the end-2026 horizon is still met. |
| Same file, GZ-V4-05: export plus read-only external iCal subscription, without OAuth | CAL-06 | Compatible, not superseded: export plus read-only iCal remains through 1.0; OAuth write sync is later. |
| Same file, GZ-V4-19/20: Bases defaults 50k records and 200 fields | BASE-08 | Compatible and reaffirmed as manageable defaults. |
| `docs/mail/FOUNDER-QUESTIONS.md` #7 working default excluded read receipts and deferred OOO | MAIL-09 | Direct product overlap is superseded: OOO is in 1.0 and read receipts are privacy-respecting opt-in. Implementation remains unaccepted. |
| Same file #13 working default deprecated unrestricted admin broadcast | SEC-09 | Governed campaigns are approved only with an audience, preview, and audit. This does not accept the existing unrestricted broadcast or waive consent/legal controls. |
| Same file #31 working default used 50 MiB total-message maximum | FILE-05 | Direct product quantity is superseded by the exact selected A wording: mail is 25 MB. Attachment-level and protocol enforcement details still need a technical contract. |
| Same file #6, #15, #17, #21, #23, #24, #25 and #26 working defaults | MAP-06, CALL-02, CAL-08, CAL-09, CALL-07, CAL-12, MAIL-03, MAIL-04, MAIL-05 | The named IDs control only their direct product behavior. More detailed nonconflicting security, retry, rendering, migration, and adapter semantics remain technical contracts rather than new founder answers. |

Only questions **#11** and **#29** in
`docs/mail/FOUNDER-QUESTIONS.md` are explicitly marked and accepted as
`ANSWERED`; they control the Mail branch routing and provenance exception.
Every other row remains unresolved or at its documented working default unless
a specific embedded A ID in the ledger above directly controls the same product
behavior. The all-A reply to the master questionnaire must not be treated as an
answer to the separate numbered Mail question list.

Named accepted ADR/spec contracts continue to control technical semantics where
they do not conflict with an embedded A choice. In particular, Mail/Calendar
fail-closed guest access, migration oracles, idempotency, DNS verification, and
bounded adapter contracts are not weakened by this product decision.

## Exact model routing and provenance locks

The global routing lock is:

- root `team-lead`: `gpt-5.6-sol`, `max`;
- deterministic mechanics, hashes, and provenance:
  `gpt-5.6-terra`, `low`;
- ordinary analysis, authoring, and documentation:
  `gpt-5.6-terra`, `medium`;
- independent review: `gpt-5.6-terra`, `high`;
- unusually complex downstream audits:
  `gpt-5.6-terra`, `xhigh`.

The task-local implementation exception applies **only** on branch
`codex/mail-module-20260813`, **only** to Mail workstream Phases 143–148, and
**only** to atomic production/test-code packets:
`gpt-5.6-terra`, `xhigh`.

Dispatcher-enforced provenance is sufficient only when a native collaboration
dispatch explicitly sets the exact model and effort, the dispatcher accepts it,
and the worker runs without rejection; record
`provenance_source: dispatcher-enforced`. If telemetry is exposed, it must
match. A model/effort mismatch, dispatch rejection, omitted model or effort, or
any fallback is `BLOCKED`. This exception does not waive frozen-contract,
scope, independent-review, launcher, security, or runtime gates.

## Non-acceptance boundaries

This record selects product choices. It does **not** mean or prove:

- implementation, module, migration, or product completion;
- stability, production readiness, or deployment authorization;
- dependency, sanitizer, provider, protocol, AGPL/SEL/commercial, or other
  licence/legal approval;
- a real mail domain, DNS operator, PTR/MX/SPF/DKIM/DMARC, TTL comparison,
  HTTPS origin, reverse proxy, split DNS, or network boundary is ready;
- schemas or data migrations are designed, reviewed, applied, reversible, or
  safe against the production dataset;
- tests, validators, builds, Docker launchers, security review, legal review,
  external integrations, runtime checks, or acceptance gates have passed.

Unproved execution facts retain their truthful state: `UNBOUND`,
`NOT_READY`, `NOT_RUN`, `UNKNOWN`, or `BLOCKED`, as applicable.

## Roadmap and specification consequences

Roadmap and specification work must use the 16 premises and 209 selected A
entries as the product baseline, preserve every qualifier, quantity, deferral,
failure behavior, and privacy/security boundary, and map them to traceable
requirements, implementation packets, and acceptance evidence.

Review finding I-05 is resolved operationally later by exact-path staging.
This record does not claim to be the only worktree change; unrelated Calendar
work-in-progress remains outside this decision record and outside its
ownership.

## Remaining unresolved matters

No unresolved contradiction remains inside the embedded approved extract.
Execution details deliberately left open by its wording remain open, including
licensing, real domains and DNS authority/evidence, exact OAuth exposure,
dependency approval, production migration mechanism and data oracle, provider
contracts, external egress, and runtime acceptance. Those are evidence-bearing
gates, not implied approvals.
