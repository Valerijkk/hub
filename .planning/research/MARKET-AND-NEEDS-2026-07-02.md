# Hub — Исследование рынка, конкурентов и потребностей сотрудников (2026-07-02)

> **Тип:** внешнее исследование (WebSearch 2026) + сверка с внутренними docs (read-only).
> **Цель:** дать фундамент для продуктового и архитектурного решения — **«сколько микросервисов нужно и какие пилить»** для Hub (self-hosted корпоративный super-app для команд 10–50: трекер + мессенджер + TMS + звонки + дашборды + интранет/календарь/CRM/HR/учёт/no-code БД/AI-хаб).
> **Метод:** сначала прочитаны внутренние документы (см. §0), затем веб-исследование 2026 по конкурентам, трендам и потребностям ролей. Отчёт **дополняет**, а не дублирует уже зафиксированное.
> **Оговорка:** веб-цены/фичи — по данным вендоров и обзоров на начало–середину 2026; проверять перед решениями о закупке. Раздел §6 (микросервисы) — **вход** для архитектурного решения, НЕ финальная архитектура.

---

## 0. Что уже зафиксировано во внутренних docs (чтобы не дублировать)

Прочитаны: `docs/vision_and_scope.md`, `docs/user_needs_analysis_2026.md`, `docs/competitive_analysis_2026.md`, `docs/competitive_audit.md`, `.planning/decisions/specs/PLATFORM-VISION-v4.md`, `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`, `docs/tz_constraints.md`.

**Уже покрыто внутри (не повторяем подробно):**
- Матрица «Hub vs лидеры» по категориям (Jira/Linear/Plane, Slack/Discord/Mattermost, TestRail/Allure, Zoom, Mighty/Bitrix24/Lark) — `competitive_analysis_2026.md` §2–§8.
- Top-10 gaps и Top-10 дифференциаторов Hub — там же §5–§6.
- Персоны × JTBD × боли (Dev/QA/TL/PO/Admin/Member/DevOps/Employee) — `user_needs_analysis_2026.md` §2.
- Тренды 2025–2026 и ответ Hub (AI-native, consolidation, automation, keyboard-first) — `competitive_analysis_2026.md` §4.
- Логические домены и намёк на split (hub-ai-gateway, hub-intranet при >500 concurrent WS) — `PLATFORM-VISION-v4.md` §2.
- Глубина модулей по факту кода: ядро «полное», v4-модули «скелеты» — `EXPANSION-ANALYSIS-2026-07-02.md` §2.
- Ограничения TZ (Magic Link only, нет SSO/LDAP до v5, NATS, Fiber WS, SeaweedFS, 1 org/инстанс) — `tz_constraints.md`.

**Чего во внутренних docs НЕ было или было слабо (главный вклад этого отчёта):**
1. **RU-контекст конкурентов** (Kaiten, YouGile, Yonote, Пачка, VK Teams, Яндекс 360) и реестр Минцифры/152-ФЗ как драйвер спроса — раньше почти не разбирался.
2. **Свежие цены 2026** и TCO-цифры (Linear/Jira/Confluence, Bitrix24 On-Premise, Notion/ClickUp).
3. **Сдвиг тренда 2026: «workspace platform / agent-native»** — агенты живут внутри воркспейса с памятью, **MCP как lingua franca**. Это выше, чем «AI-копилот», который Hub уже закрыл.
4. **Явные новые пробелы**: knowledge base/wiki (на фоне ухода on-prem Confluence), MCP/агенты, e-sign, BI-builder, collaborative docs/whiteboard.
5. **Предварительная нарезка доменных микросервисов** как отдельный аналитический раздел (§6).

---

## 1. Резюме

**Где Hub силён (подтверждено рынком 2026):**
- Единственный self-hosted продукт, который совмещает **полноценный dev/QA-стек (трекер + TMS + Allure-CI + звонки)** и **корпоративный super-app** в одном инстансе. Ни Bitrix24 (слабый dev/TMS), ни Mattermost (чат+ChatOps, без TMS/CRM/HR), ни Яндекс 360 (без TMS) этого не дают.
- **RU-friendly AI без VPN** (GigaChat/YandexGPT/Ollama, BYOK) — прямое попадание в 152-ФЗ и запрет облачных OpenAI/Anthropic для ПДн.
- **TCO**: против связки Jira Premium + Confluence + TestRail + Zoom + Slack (для 45–50 чел. это ~$110–207k/год у Linear/Jira-расчётов, отдельная строка Confluence ~$3–6k/год) Hub = стоимость инфраструктуры. Atlassian поднял цены Data Center на 15% в феврале 2026 — усиливает аргумент.

**Где Hub отстаёт (глазами сотрудника, не только «фича-паритет»):**
- **Нет knowledge base / wiki** — а это ежедневный инструмент. Notion, Confluence, Yonote держат людей внутри именно вики+доками. При этом Atlassian сворачивает on-prem Confluence (EOL) → **окно возможности**, которым уже пользуется Mattermost («Mattermost Docs», март 2026).
- **AI — это чат-бот, а не агент.** Рынок 2026 ушёл в «agent-native»: агенты с памятью внутри воркспейса + **MCP** как стандарт интеграции. У Hub `@hub-ai` — копилот, но нет MCP-сервера, нет автономных агентов/триажа.
- **Коллаборативные документы/доски и e-sign/BI** — есть у Bitrix24/Lark/Notion, у Hub нет.
- **Enterprise-вход (SSO/LDAP/AD)** — RU-корпораты on-prem ждут AD/LDAP; Bitrix24 On-Premise это даёт. Hub запрещает до v5 — трение при продаже в компанию >30 чел.
- **v4-модули — «скелеты»** (intranet/calendar/crm/hr/accounting/bases): вширь есть, вглубь нет (подтверждено `EXPANSION-ANALYSIS`).

**Чего не хватает глазами сотрудника (что держит/выталкивает из инструмента):**
- Держит: один UI без Alt-Tab, автосинк задача↔git↔статус, поиск Ctrl+K, звонок из контекста, AI под рукой.
- Выталкивает: ручное обновление статусов в 3–4 местах, шум уведомлений без дайджеста, «Загрузка…» вместо skeleton, формы без валидации, отсутствие вики (где лежат решения), отсутствие мобильного нативного клиента.

**Предварительная рекомендация по числу микросервисов (детали §6):** НЕ дробить. Оставить **модульный монолит-ядро** + уже вынесенные `LiveKit` (медиа) и `hub-mail`. **Пилить отдельно только то, что имеет иной профиль нагрузки/технологии:** (1) **AI-gateway** (GPU/стриминг/долгие запросы/изоляция ключей), (2) **Realtime/WS-хаб** (presence + fan-out сообщений/уведомлений; при росте concurrent WS). Итог целевого состояния: **~4–5 деплой-юнитов**, не 15.

---

## 2. Матрица конкурентов (2026)

### 2.1 Глобальные / западные

| Продукт | Ключевые модули | Что делает лучше всего | Модель / цена (2026) | Self-hosted? | AI |
|---|---|---|---|---|---|
| **Bitrix24** | CRM, задачи+Gantt, мессенджер, видео, календарь, доки+доски, HR, e-sign, RPA, BI Builder | «Всё-в-одном» для SMB; **flat-цена** (не per-user); CRM | Cloud flat-rate; **On-Premise Enterprise ~$24 990** (source code, MSA в топ-редакции) | ✅ On-Premise (полный исходник, AD/LDAP/SSO) | CoPilot AI везде |
| **Lark / Feishu** | Чат, видео, календарь, Docs, **Base (no-code)**, approval | Docs+Base+календарь как единое полотно; глобальные команды | SaaS, per-user | ❌ (cloud) | AI в Docs/Base |
| **Notion** | Docs/wiki, БД-таблицы, лёгкие задачи | **Knowledge base / вики №1**, гибкие БД | $10–15/user/мес; **Notion AI ~$10/user доп.** | ❌ | Notion AI (add-on), External Agent API |
| **ClickUp** | Задачи, доки, цели, тайм-трекинг, Gantt | Широта PM + мощные автоматизации (10k+/мес) | $7–19/user/мес; **ClickUp Brain включён** в платные | ❌ | ClickUp Brain (включён) |
| **Slack** | Каналы, треды, huddles, marketplace ботов | Экосистема интеграций/ботов; huddles | ~$10/user/мес; Slack AI — доп. | ❌ | Slack AI (premium) |
| **Microsoft Teams** | Чат, видео, M365, Copilot Studio | Глубина M365; Copilot как «super app» | ~$7/user/мес + Copilot отдельно | ❌ | Copilot + Agent 365 |
| **Atlassian (Jira/Confluence)** | Трекер, вики, roadmaps, marketplace 3000+ | Enterprise-workflow, JQL, экосистема | Jira ~$8–16/user; **Confluence ~$6/user отдельно**; DC-цены +15% (фев 2026) | ⚠️ Data Center ($$$); **on-prem Confluence сворачивается (EOL)** | Atlassian Intelligence + **Rovo** (Premium) |
| **Linear** | Issues, cycles, initiatives, triage | **Скорость (<100мс), keyboard-first**, dev-UX | Free / $10 / $16 user | ❌ | **Linear Agents** (без metered токенов), Triage Intelligence |
| **Basecamp** | To-do, message board, docs, chat | Простота, async, flat-fee | Flat (~$299/мес unlimited) | ❌ | минимально |
| **Zoom** | Видео, webinar, phone, Team Chat | Качество видео/ВКС | per-host | ❌ | AI Companion |
| **Mattermost** | Чат, workflow, voice, screen share, **Docs (вики, в разработке)**, AI | **ChatOps** (GitHub/GitLab/Jenkins/PagerDuty/Prometheus), суверенность | MIT (Team free); Enterprise (SSO/compliance) | ✅ (Go+React+Postgres, один бинарь) | AI-workflows + sovereign LLM (Ollama) |

### 2.2 RU-контекст (ключ для позиционирования Hub)

| Продукт | Ключевые модули | Лучше всего | Модель | Self-hosted? | AI |
|---|---|---|---|---|---|
| **Яндекс 360 / Трекер** | Трекер (Jira-аналог), Мессенджер, Почта, Диск, Доки, Вики, Телемост | Серьёзная разработка + экосистема Яндекса; >45k орг. | SaaS per-user | ⚠️ **on-prem обещан к концу 2026** | YandexGPT в сервисах |
| **VK Teams (VK WorkSpace)** | Мессенджер, почта, ВКС, диск, доки, доска | Импортозамещение M365 (перекос в коммуникации), госсектор | Cloud/**коробка** | ✅ коробка | базово |
| **Kaiten** | Kanban/канбан-потоки, процессы, отчёты | **Визуализация Agile/матричных процессов** на одном экране | SaaS + on-prem | ✅ | базово |
| **YouGile** | Задачи + **чат в каждой задаче** + простая CRM | Скорость/минимализм, «чат в контексте задачи», малый бизнес | SaaS + коробка | ✅ | базово |
| **Yonote** | Доки/вики, БД, лёгкие задачи | **Notion-аналог для РФ** (knowledge base) | SaaS + on-prem | ✅ | AI-функции |
| **Пачка** | Мессенджер (каналы/треды), интеграции | Структурный корп-мессенджер без «тяжёлой» экосистемы | SaaS | ⚠️ | базово |
| **ПланФикс** | No-code процессы, CRM, проекты | Конструктор процессов (ClickUp/Monday-аналог) | SaaS | ⚠️ | базово |

**Что перенять (приоритетно):**
1. **Knowledge base/вики** уровня Yonote/Notion/Confluence — самый частый «держатель» пользователей; усилено уходом on-prem Confluence.
2. **Flat-цена / предсказуемый TCO** как явный маркетинг-аргумент (у Bitrix24 это главный крючок против per-user).
3. **«Чат в контексте задачи»** (YouGile) — у Hub есть deep links, но не полноценный тред-в-карточке; закрыть issue↔thread двусторонне.
4. **BI Builder / кастомные панели** (Bitrix24) — против «статичных виджетов» Hub.
5. **E-signature + approval/RPA** (Bitrix24) — для HR/accounting/CRM-сценариев в РФ.
6. **ChatOps-глубина** (Mattermost) — Hub уже близко (#hub-ci, github/gitlab), добить PagerDuty/alert→issue.
7. **Реальные MCP-серверы** — чтобы Hub «говорил MCP» (бинарный сигнал закупки по трендам, см. §3).

---

## 3. Тренды 2026 и следствия для Hub

| # | Тренд 2026 (источник) | Следствие для Hub | Обязательно/Опц. |
|---|---|---|---|
| T1 | **Workspace platform / re-aggregation.** Средняя орг. держит ~305 SaaS-приложений; рынок консолидируется в платформы, которые «запускают другой софт внутри себя». | Hub концептуально уже здесь (super-app). Усилить нарратив «одна платформа вместо 5–8 подписок». | Обязательно (позиционирование) |
| T2 | **Agent-native, не copilot.** Агенты живут в воркспейсе с **персистентной памятью**, а не вызываются извне. Copilot→автономные агенты (MS Build 2026, Notion External Agent API, Atlassian Rovo, Linear Agents). | `@hub-ai` — это копилот. Нужны: агенты с памятью на проект, авто-триаж задач, draft-тикеты из чата, agent-действия под RBAC/аудит. | **Обязательно** (иначе отставание на поколение) |
| T3 | **MCP как lingua franca.** «Говорит ли инструмент на MCP?» — бинарный сигнал закупки; Zoom/Notion/GitHub/Atlassian/Slack уже поставляют MCP-серверы. | Дать **Hub MCP-сервер** (задачи/тесты/доки/сообщения как MCP-tools) + OpenAI-compatible endpoint (уже в backlog US-F06). | **Обязательно** |
| T4 | **Суверенный AI / self-hosted (152-ФЗ, EU AI Act).** Облачные OpenAI/Anthropic для ПДн вне закона в РФ; штрафы 10–15 млн ₽; политика geofencing на уровне шлюза. | Killer-feature Hub. Добавить **audit-логи на каждый AI-запрос**, geofencing/allow-list провайдеров, режим air-gap (только Ollama). | **Обязательно** (RU-рынок) |
| T5 | **Импортозамещение РФ.** Реестр Минцифры 30k+ продуктов, но полная миграция у 1–2%; требования совместимости с рос. ОС (офис с 01.09.2026, облако с 01.01.2027). | Заявить совместимость с рос. ОС/реестр Минцифры как цель; это прямой драйвер спроса для целевого сегмента. | Обязательно (go-to-market) |
| T6 | **Context switching как измеримый налог.** 70% инженеров >5 инструментов/нед; 88% теряют продуктивность; ~23 мин на восстановление фокуса; 6–15 ч/нед на tool sprawl. | Двойная ставка на «no Alt-Tab»: авто-статусы задача↔git, deep-work режим, тред-решения в карточке, дайджест-уведомления. | Обязательно |
| T7 | **Knowledge management как отдельная категория; уход on-prem Confluence.** Mattermost Docs (март 2026) как «суверенный преемник Confluence». | **Пилить wiki/knowledge base** (self-hosted) — крупнейший продуктовый пробел Hub. | **Обязательно** |
| T8 | **Автоматизация → агентная.** Deterministic Zapier/Make уступают агентам, рассуждающим о цели; но требуют MCP + память + аудит. | Углубить automation Hub от lite-правил к «целевым» сценариям + связать с T2/T3. | Опционально (после T2/T3) |
| T9 | **Встроенные звонки/видео + async.** ВКС встроен в платформы (Teams/Lark/Bitrix24), плюс рост async-культуры. | LiveKit уже даёт это. Добавить async-примитивы (видео/voice-заметки, записи с транскриптом/AI-резюме). | Опционально |
| T10 | **Governance/аудит агентов** (Agent 365, control plane). | Расширить существующий audit на AI-агентов: идентичность агента, действия, откат. | Опционально (по мере T2) |

---

## 4. Персоны × must-have (сотрудник, не только менеджер)

> Дополняет `user_needs_analysis_2026.md` §2 — фокус на «что нужно как сотруднику ежедневно» и свежих пробелах из §3.

| Персона | Топ-потребности (JTBD, ежедневно) | Покрыто в Hub? | Пробел (2026) |
|---|---|---|---|
| **Разработчик** | Задача→ветка→PR→статус без ручного апдейта; поиск; звонок из контекста; AI-помощь по коду; **вики с решениями** | Частично: git-автолинк ✅, Ctrl+K ✅, `@hub-ai` ✅ | **Нет вики/ADR-хранилища**; AI без MCP/агентов; issue↔thread не двусторонний; статусы иногда вручную |
| **QA / тестировщик** | Кейсы + Allure из CI; traceability; баг из failed; **тренды/flaky/run-diff**; отчёт к релизу | Хорошо: TMS ✅, Allure ✅, traceability ✅ | **Flaky ML / run-diff — частично** (vs Allure TestOps); bulk-bug из прогона |
| **PM / тимлид** | Спринт/burndown/velocity; **roadmap/timeline**; загрузка команды; дайджест; **BI-панели** | Частично: спринты ✅, дашборды ⚠️ статичные | **Нет roadmap/timeline**, custom fields UI, capacity planning, BI-builder |
| **Менеджер (CRM/продажи)** | Контакты/компании/сделки; активности по сделке; **won→счёт**; e-sign договора | Скелет: CRM-lite ✅, won→invoice ✅ | Компании как сущность, активности, воронка-аналитика, **e-sign**, импорт |
| **HR** | Заявки на отпуск+апрув; **org chart**; **handbook/вики**; онбординг-чеклисты; баланс отпусков | Скелет: leave+1C ✅ | Визуальный org chart, **wiki-handbook**, баланс отпусков, онбординг не-dev |
| **Ops / админ** | Пользователи/аудит/бэкапы/лимиты; **SSO/LDAP**; наблюдаемость; секьюрити | Хорошо: audit ✅, OBS ✅, backups ✅ | **Нет SSO/LDAP/AD** (ждут on-prem корпораты); дашборд slow-query/rate-limit; ротация секретов (PG-03) |
| **Сотрудник (не-dev)** | Новости компании; календарь; отпуск; **вики/база знаний**; мобайл; kudos | Скелет: intranet/calendar/HR ✅ | **Нет вики**; мобильный только PWA; лента pin/reactions/модерация — мелко |

**Сквозной вывод по персонам:** самый повторяющийся дефицит через *все* роли — **knowledge base/wiki**, **дайджест-уведомления**, **углубление v4** и (для enterprise-продажи) **SSO/LDAP**.

---

## 5. Функциональные пробелы Hub vs рынок (приоритизировано) и кандидаты на новые модули/сервисы

> P0 = стратегически критично 2026 · P1 = высокая ценность · P2 = средняя · P3 = nice-to-have. Учтены ограничения `tz_constraints.md` (часть требует ADR).

| # | Пробел / кандидат-модуль | Эталон рынка | Почему важно (тренд/персона) | Приоритет | Комментарий по scope |
|---|---|---|---|---|---|
| G1 | **Knowledge base / Wiki** (страницы, вложенность, коллаб-редактирование, ADR) | Confluence(EOL on-prem), Notion, Yonote | T7 + все персоны; крупнейший «держатель» пользователей | **P0** | Новый модуль; переиспользовать help/pages, storage, search |
| G2 | **AI-агенты + MCP-сервер** (память на проект, авто-триаж, draft из чата, OpenAI-compatible API) | Linear Agents, Rovo, Notion Agent API | T2+T3; иначе отставание на поколение | **P0** | Расширение `ai/`; US-F06 уже в backlog |
| G3 | **AI governance для 152-ФЗ** (audit каждого запроса, geofencing/allow-list, air-gap Ollama) | Sovereign-AI шлюзы | T4; RU-комплаенс = продажный аргумент | **P0** | Расширение `ai/` + audit |
| G4 | **SSO / LDAP / AD (read-only)** | Bitrix24 On-Premise, Mattermost EE | Enterprise on-prem РФ ждёт; сейчас FORBIDDEN до v5 | **P1** (нужен ADR) | Меняет TZ; решение PO/arch |
| G5 | **BI / кастомные дашборды** (конструктор панелей, alerting) | Bitrix24 BI Builder, Grafana | PM/Admin; «статичные виджеты» — известный gap | **P1** | Углубление `dashboard/`; OBS-стек уже есть |
| G6 | **Workflow-automation вглубь → агентная** (правила по цели, triage inbox) | Jira Automation, ClickUp, Make | T8; TL/PO/DevOps | **P1** | Углубление automation lite + связка с G2 |
| G7 | **Collaborative docs / whiteboard** (онлайн-редактирование, доски) | Lark Docs, Notion, Bitrix24 Boards | T1; сотрудник/PM | **P2** | Может слиться с G1 (wiki) |
| G8 | **E-signature + approvals/RPA** | Bitrix24 (e-sign, RPA) | HR/CRM/accounting сценарии РФ | **P2** | Модуль поверх HR/CRM/accounting |
| G9 | **TMS reporting уровня TestOps** (flaky ML, run-diff, категории дефектов) | Allure TestOps | QA-удержание | **P2** | Углубление `tms/` |
| G10 | **Roadmap / timeline / initiatives** | Linear, Jira | TL/PO | **P2** | Углубление `tracker/sprint` |
| G11 | **Notification digest + smart delivery** | Slack | T6; все | **P2** | Углубление `notification/` |
| G12 | **Мобильный клиент вглубь** (PWA offline / нативный) | Slack/Teams/Bitrix mobile | Сотрудник/Member | **P2** (натив — ADR) | PWA polish разрешён; натив — FORBIDDEN до ADR |
| G13 | **Углубление v4 (intranet/calendar/crm/hr/accounting/bases)** | Bitrix24/Lark | «широко, но неглубоко» | **P1** | Главный фронт «полноты» (см. EXPANSION §5 эпик B) |
| G14 | **Real-time перевод сообщений** | Rocket.Chat | глобальные/смешанные команды | **P3** | Ниша |

**Кандидаты именно на новые модули/сервисы (сводно):** `wiki/knowledge-base` (G1), `ai-agents/mcp` (G2/G3, расширение ai), `bi/analytics` (G5), `workflow-automation` вглубь (G6), `e-sign` (G8). Прочее — углубление существующих модулей.

---

## 6. Предварительные кандидаты доменных микросервисов (вход для решения о числе сервисов)

> **Не финальная архитектура.** Это группировка доменов Hub по критериям декомпозиции для последующего решения `arch`/`arch-ver`. Опирается на консенсус практики 2026: **начинать с модульного монолита, выделять сервис только при доказанном ограничении** (иной профиль масштабирования / технологии / частоты релизов / автономии команды). Дробление «ради чистоты/безопасности» — антипаттерн (кейсы отката auth-в-сервис, рост latency 80мс→450мс).

### 6.1 Критерии выделения (rubric)

| Критерий | Что означает | Кого выталкивает из монолита |
|---|---|---|
| **Профиль нагрузки/ресурсов** | GPU/RAM/сеть кардинально иные | AI-inference, real-time WS (десятки-сотни k соединений), медиа |
| **Технология/стек** | нужен иной рантайм/протокол | сигналинг/медиа (WebRTC), потенциально ML |
| **Изоляция данных/секретов** | отдельная схема/ключи, blast-radius | AI-ключи (BYOK), медиа-записи |
| **Частота релизов/автономия** | отдельный релиз-цикл, отдельная команда | обычно нерелевантно при команде <50 инж. |
| **Строгая консистентность (ACID)** | общие транзакции, in-process | **держит вместе**: auth/rbac/tracker/tms/project |

### 6.2 Группировка доменов Hub

| Группа | Домены Hub | Профиль | Решение |
|---|---|---|---|
| **CORE (модульный монолит)** | auth, rbac, project, tracker, sprint, tms, search, github/gitlab, imex, backup, audit, dashboard | REST, ACID, сильная связность (traceability задача→тест→PR) | **Оставить вместе.** Дробить нельзя — общие транзакции и трассируемость. Границы модулей — в коде (bounded contexts), не по сети |
| **CORP (модуль внутри CORE или отдельный позже)** | intranet, calendar, crm, hr, accounting, bases, marketplace, status, **wiki (новый)** | Низконагруженный CRUD, слабо связан с dev-ядром | **Пока в монолите** как отдельные модули. Кандидат №3 на вынос (`hub-workspace`) — но по автономии/релизам, не по нагрузке |
| **REALTIME/WS** | ws (presence, fan-out), messenger real-time, notification delivery | Десятки-сотни k concurrent WS; иной профиль масштабирования, чем REST | **Кандидат на вынос №2** при росте concurrent WS (порог из PLATFORM-VISION: >500). До порога — в монолите |
| **AI-GATEWAY** | ai (провайдеры, стриминг, BYOK, будущие агенты/MCP) | GPU/латентность, долгие/стриминговые запросы, изоляция ключей, независимый релиз | **Кандидат на вынос №1** (самый обоснованный). Стриминг + иной ресурс-профиль + blast-radius ключей |
| **MEDIA/CALLS** | calls (сигналинг) + LiveKit (SFU) | WebRTC, отдельный рантайм | **LiveKit уже отдельный.** Сигналинг оставить тонким в CORE; медиа не гонять через backend |
| **MAIL** | hub-mail (`cmd/hub-mail`) | Внешний I/O, ретраи, отдельный контур | **Уже отдельный сервис** — оставить |

### 6.3 Предварительная рекомендация по числу/составу

- **Сейчас (факт):** CORE-монолит + LiveKit + hub-mail = **3 деплой-юнита**. Это правильно для команды <50 инж. и текущей нагрузки. Не дробить.
- **Целевое (по мере роста), приоритет выноса:**
  1. **AI-gateway** — вынести первым (иной ресурс-профиль, стриминг, изоляция BYOK-ключей, агенты/MCP как отдельный релиз-цикл). Уже намечено в `PLATFORM-VISION-v4` §2.
  2. **Realtime/WS-хаб** — вынести при достижении порога concurrent WS (presence + сообщения + fan-out уведомлений). До порога — не трогать.
  3. **CORP/`hub-workspace`** (intranet+calendar+crm+hr+accounting+bases+wiki) — опционально и позже, по автономии команды, а НЕ по нагрузке.
- **Итог целевого состояния: ~4–5 сервисов** (CORE + LiveKit + hub-mail + AI-gateway + WS-хаб), не 10–15.
- **Обязательные предусловия перед любым сплитом** (strangler fig): чёткие контракты (OpenAPI/gRPC) и event-схемы через **NATS JetStream** (уже есть), изоляция данных по схемам, feature-flags + shadow-traffic, OTel-наблюдаемость на новой сетевой границе (OBS-стек уже есть). Держать монолит-путь «тёплым» для отката 1–2 релиза.
- **Что НЕ выносить:** auth/rbac (кейсы отката «auth-в-сервис» из-за дублирования JWT-валидации и latency), tracker/tms/project (общие ACID-транзакции traceability).

---

## 7. Открытые вопросы к продукту

1. **Wiki/knowledge base (G1) — делаем сейчас как P0?** Это крупнейший пробел и совпадает с уходом on-prem Confluence. Отдельный модуль или расширение help/pages? Коллаб-редактирование (CRDT) в scope или «просто markdown-страницы» на первой итерации?
2. **AI: копилот → агенты + MCP (G2/G3).** Готовы ли инвестировать в agent-native (память, авто-триаж, MCP-сервер, OpenAI-compatible API)? Это ответ на главный тренд 2026, но объём L–XL.
3. **SSO/LDAP/AD (G4)** — снимаем запрет `tz_constraints` через ADR? Без этого затруднён вход в RU-корпораты >30 чел. Или держим «Magic Link only» как принцип и теряем этот сегмент?
4. **152-ФЗ-позиционирование (T4/T5):** заявляем цель «реестр Минцифры + совместимость с рос. ОС»? Это меняет go-to-market и часть требований (ОС, упаковка).
5. **Глубина v4 vs ширина ядра:** сначала углублять v4-скелеты (G13, эпик B в EXPANSION) или закрывать пробелы ядра (wiki, BI, roadmap, TMS-reports)? Ресурсы конечны.
6. **Микросервисы: триггеры, а не даты.** Согласуем ли численные пороги для выноса (например: AI-gateway — при GPU-нагрузке/квоте X; WS-хаб — при >500 concurrent WS)? Иначе риск преждевременного дробления.
7. **E-sign/BI/whiteboard (G5/G7/G8):** строим сами или интегрируемся (коннекторы)? Для accounting уже выбран путь «1C-коннектор, не своя бухгалтерия» — распространяем ли принцип «connector-first» на e-sign/BI?
8. **Мобильный клиент (G12):** остаёмся на PWA или инициируем ADR на нативный (сейчас FORBIDDEN)?

---

## Источники (2026)

- Bitrix24 pricing / On-Premise (flat-rate, ~$24 990 Enterprise, MSA, source code, AD/LDAP/SSO): bitrix24.com/prices, /prices/self-hosted.php, /self-hosted/
- Notion vs ClickUp 2026 (цены, AI add-on vs включён): workmanagementhub.com
- Linear vs Jira 2026 TCO (Confluence отдельно ~$6/user, DC +15% фев 2026, Linear Agents без metered): ctaio.dev/labs/linear-vs-jira, tech-insider.org, linear.app/pricing
- RU таск-трекеры/мессенджеры 2026 (Kaiten, YouGile, Яндекс Трекер on-prem к концу 2026, VK Teams, Пачка, Yonote): habr.com (kaiten/yougile), vc.ru, planfix.ru/blog, top107.ru
- Тренды workspace platform / agent-native / MCP: coommit.com (workspace-platform-era, agentic-workspace), 2tolead.com (MS Build 2026), research.isg-one.com (Collaborative AI Suites 2026)
- Sovereign AI / MCP / data residency: frends.com (MCP for regulated enterprises)
- 152-ФЗ / импортозамещение / RU AI-стек: prem.agmind.dev, chimitdorzhi.tech, aibot.direct, it-institute.ru (реестр Минцифры 2026)
- Context switching / tool sprawl: Chainguard 2026 Engineering Reality Report, dev.to (context switching), lucaberton.com (developer burden 2026)
- Микросервисы vs модульный монолит 2026 (когда дробить, strangler fig): dev.to (uaslimcreate), codingplainenglish.medium.com, todayamerican.medium.com, doctree-ai.com
- Self-hosted collaboration / Confluence EOL: selfhosting.sh, listicler.com, github.com/mattermost, mattermost.com/newsroom (Mattermost Docs, март 2026)

---
*Составлено 2026-07-02. Read-only анализ кода/планов; код и тесты не изменялись. Отчёт — вход для продуктового приоритезирования и архитектурного решения по числу микросервисов (передать `product`/`arch`).*
