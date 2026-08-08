# Hub — Playbook: Git workflow

> Для `be1`/`be2`, `fe1`/`fe2`, `qa-auto1`/`qa-auto2`, `devops-engineer`, `code-reviewer`, `team-lead`/`tech-lead`.
> Первоисточник: `.cursor/rules/50-git.mdc` (при конфликте правило важнее playbook'а). Стиль коммитов и код — `.cursor/rules/10-style.mdc`.
> Особый акцент — **параллельная работа двух+ исполнителей одного слоя** без конфликтов, и границы «что агенту разрешено в git, а что только пользователю».

Все примеры коммитов, веток и merge-сообщений в этом документе — **реальные, взятые из истории репозитория Hub** (`git log --oneline`), а не выдуманные. Это сделано намеренно: playbook должен отражать то, как реально работает команда, а не абстрактный идеал.

*Обновлено: 2026-07-18.*

---

## Оглавление

1. [Зачем дисциплина git в команде из многих агентов](#1-зачем-дисциплина-git-в-команде-из-многих-агентов)
2. [Ветки](#2-ветки)
3. [Conventional Commits](#3-conventional-commits)
4. [Что НЕЛЬЗЯ коммитить](#4-что-нельзя-коммитить)
5. [Параллельная работа be1/be2, fe1/fe2, qa-auto1/qa-auto2](#5-параллельная-работа-be1be2-fe1fe2-qa-auto1qa-auto2)
6. [Разрешение конфликтов (шаг за шагом)](#6-разрешение-конфликтов-шаг-за-шагом)
7. [PR-процесс](#7-pr-процесс)
8. [Интеграционные ветки волн](#8-интеграционные-ветки-волн)
9. [Обязательная регрессия перед коммитом/PR](#9-обязательная-регрессия-перед-коммитомpr)
10. [Связь с docs-sync](#10-связь-с-docs-sync)
11. [Теги и релизы](#11-теги-и-релизы)
12. [Безопасные практики (жёсткие запреты)](#12-безопасные-практики-жёсткие-запреты)
13. [Когда `--amend` нельзя (и когда можно)](#13-когда---amend-нельзя-и-когда-можно)
14. [Кто пушит/мержит в master — границы агента](#14-кто-пушитмержит-в-master--границы-агента)
15. [Антипаттерны git (сводка)](#15-антипаттерны-git-сводка)
16. [Быстрая справка по командам](#16-быстрая-справка-по-командам)
17. [Чек-лист перед PR](#17-чек-лист-перед-pr)

---

## 1. Зачем дисциплина git в команде из многих агентов

В Hub одновременно работают пары/потоки исполнителей: `be1`/`be2` (backend), `fe1`/`fe2` (frontend), `qa-auto1`/`qa-auto2` (автотесты), а сверху — `ba1`/`ba2`, `sa1`/`sa2` на аналитике. Без дисциплины они наступают друг другу на файлы, ломают историю и создают конфликты мержа, которые стоят дороже самой фичи. Правила ниже делают параллельную разработку предсказуемой.

Три опоры:
1. **Маленькие атомарные коммиты/PR** — легко ревьюить, легко откатывать (`git revert <sha>` затрагивает ровно одну логическую правку).
2. **Чёткие границы файлов** — параллельные агенты не редактируют одно и то же одновременно.
3. **Частый rebase на `master`** — конфликты ловятся рано и мелкими, а не одним гигантским конфликтом перед мержем.

### 1.1 Read-only preflight состояния checkout

Перед branch/worktree/review/cleanup записать факты, не угадывать по prompt:

```text
git rev-parse --show-toplevel
git rev-parse --git-dir
git rev-parse --git-common-dir
git rev-parse --show-superproject-working-tree
git branch --show-current
git rev-parse --verify HEAD
git status --short
```

Пустая branch строка означает detached HEAD; непустой superproject path —
submodule; различие git-dir/common-dir — linked worktree. Это read-only
диагностика, не разрешение cleanup. Удалять/перемещать worktree/submodule можно
только при доказанном task ownership точного canonical path и отдельной
authority; чужой или ambiguous checkout остаётся нетронутым.

Реальная история репозитория Hub подтверждает эти принципы на практике: волновая разработка идёт через feature-ветки (`feat/platform/otel-nats-trace`, `fix/platform/wave1-quickfixes`) и docs-ветки (`docs/wave2-specs`, `docs/wave2-adr-package`), которые сливаются в интеграционную ветку волны (`integration/wave1`) до финального мержа в `master` — подробнее в §8.

---

## 2. Ветки

```
master                              production-ready, protected
feat/<module>/<description>         новая фича
fix/<module>/<description>          багфикс
refactor/<module>/<description>     рефакторинг
chore/<description>                 deps, tooling, config
docs/<description>                  документация/аналитика/спеки без кода
hotfix/<issue-id>/<description>     критический прод-фикс
integration/<wave-name>             интеграционная ветка волны (см. §8)
```

- `<module>` — из списка scope (см. §3): `auth`, `tracker`, `sprint`, `messenger`, `tms`, `calls`, `dashboard`, `notification`, `search`, `ws`, `rbac`, `platform`, `frontend`, `infra`, `ent`, `ai`, `crm`, `intranet`, `calendar`, `hr`, `accounting`, `bases`, `marketplace`.
- `<description>` — kebab-case, коротко и по делу.
- Одна ветка = одна логическая задача/план. Не смешивай несвязанные изменения.
- Ветку держим короткоживущей: чем дольше живёт, тем больше конфликтов при финальном мерже.

### Реальные примеры имён веток из истории Hub

```
feat/platform/otel-nats-trace     — W3C trace context в NATS headers (4 атомарных коммита, см. §3)
fix/platform/wave1-quickfixes     — квик-фиксы волны 1, смерджена в integration/wave1
docs/wave1-analysis               — аналитика (US/AC/ERD) волны 1, смерджена в integration/wave1
docs/wave2-specs                  — SA-спеки волны 2 (Intranet/Calendar/Wiki)
docs/wave2-adr-package            — пакет ADR волны 2 (AUTH-SSO, AI-AGENTS-MCP)
feat/ux-1-palette                 — Ctrl+K палитра, PR #15
```

<good-example>

Хорошее имя ветки — модуль/scope виден сразу, описание конкретное:

```
feat/tracker/wip-limits
fix/ws/send-on-closed-channel
refactor/auth/extract-repository-interface
docs/wave3-specs
```
</good-example>

<bad-example>

```
feature-branch          (нет модуля, нет описания)
my-fixes                (чьи фиксы, какие?)
tmp                     (не отражает содержимое, легко забыть смерджить)
wave3                   (для интеграционной ветки нужен префикс integration/, см. §8)
```
</bad-example>

---

## 3. Conventional Commits

```
<type>(<scope>): <описание в повелительном наклонении>

[тело: что и ПОЧЕМУ; ссылка на PLAN/тикет]
```

**Типы:** `feat | fix | refactor | test | docs | chore | perf | ci`
**Scope:** `auth | tracker | sprint | messenger | tms | calls | dashboard | notification | search | ws | rbac | platform | frontend | infra | ent | ai | crm | intranet | calendar | hr | accounting | bases | marketplace`

### 3.0. FORBIDDEN — AI co-author (жёстко, все модели/агенты)

**Никогда** не добавлять в коммит:

- `Co-authored-by:` / `Co-Authored-By:` (Cursor, Claude, Codex, Cloud Agent, Copilot, Anthropic, OpenAI, `cursoragent@…`, любой AI)
- `Generated-by:` / «Made with Cursor|Claude|Codex»

Автор = только git user владельца репо. После каждого коммита агент проверяет `git log -1 --format=%B` и сдирает trailer, если среда его дописала. Правило: `.cursor/rules/51-no-ai-coauthor.mdc` · хук: `.githooks/commit-msg` (`git config core.hooksPath .githooks`).

#### Выключить attribution в Cursor навсегда (владелец, один раз)

1. **Cursor Settings** (не VS Code Settings): `Ctrl+Shift+J` / Command Palette → «Cursor Settings»
2. Раздел **Agents → Attribution** — выключить:
   - **Commit Attribution**
   - **PR Attribution**
3. Перезапустить Cursor.
4. CLI (файл вне репо): `%USERPROFILE%\.cursor\cli-config.json`:

```json
"attribution": {
  "attributeCommitsToAgent": false,
  "attributePRsToAgent": false
}
```

Cloud Agents могут форсить trailer серверным хуком — на IDE-тоггл не опираться;
для Hub cloud-коммиты без личного review владельца запрещены.

### 3.1. Реальные примеры по каждому типу (из `git log` Hub)

**`feat` — новая функциональность:**
```
feat(platform): propagate trace context via NATS headers on publish
feat(platform): continue producer trace in NATS consumers
feat(platform): persist W3C trace context on outbox events
feat(tracker): add WIP limits to Kanban columns
feat(messenger): Thread → Issue promotion (Feature #58)
feat(messenger): mobile-friendly channel drawer (UX-1)
feat(ui): add QueryStateView for unified loading/error/empty handling
feat(ui): surface NotificationBell in Topbar for discoverability
feat(frontend): extend Ctrl+K palette with calls, settings, admin (#15)
feat(hub): ship v2.0.2 — security, UX-08, coverage, a11y e2e
```

**`fix` — багфикс:**
```
fix(auth): count magic-link code length in runes, not bytes
fix(auth): propagate invalidateUserSessions errors to enforce single-session policy
fix(bases): scope all base operations to owner (403 for foreign bases)
fix(audit): batch GetResourceLimits settings lookup to remove N+1
fix(frontend): read GIPHY API key from VITE_GIPHY_API_KEY env
fix(ws): prevent send on closed channel panic (SEC-05)
fix(notification): ctx-aware rate-limit sleep in push worker
fix(messenger): close thread panel on Escape (a11y)
fix(messenger): wire lifecycle ctx into auto_channels NATS handler
fix(messenger): sanitize <a href> against javascript: scheme in render-content (A2)
fix(project): require membership for SaveAsTemplate (A1)
fix(tms): cap per-file size in multipart webhook uploads
fix(ui): replace window.confirm with AlertDialog in messenger and tms
fix(ui): replace window.confirm with AlertDialog in admin
fix(ui): remove dead Ctrl+N hotkey and Ctrl+K placeholder
fix(ui): add explicit type=button on remaining unmarked buttons (MF-3 closeout)
fix(cron): recover from panics in scheduled jobs to prevent server crash
fix(backup): use lifecycle ctx for async cleanup goroutine
fix(main): defer apiKeySvc.Close to drain touch-worker pool on shutdown
fix(infra): use 127.0.0.1 in compose healthchecks (IPv6 localhost)
fix(infra): bump Alpine to 3.21 for postgresql17-client in Docker image
fix(platform): strip UTF-8 BOM from handler.go for Go 1.25
fix(ci): retry docker compose up on registry timeout
```

**`refactor` — рефакторинг без изменения поведения:**
```
refactor(auth): extract AuthRepository interface
refactor(routes): apply QueryStateView to projects/board/issues (UX-2)
refactor(tms): replace EntClient escape hatch with EnvironmentRepository
refactor(ent): default schema timestamps to UTC via shared clock helper
refactor(*): UTC time.Now in DB write paths and timestamp comparisons
refactor(*): use UTC time for soft-delete timestamps
```

**`test` — тесты:**
```
test(auth): cover magic-link brute-force lockout (TC-AUTH-10)
test(bases): RBAC denial suite for foreign-base access
test(gitlab): mock GitLab API in routes integration test
test(platform): cover trace-context round-trip across outbox and NATS
```

**`docs` — документация/аналитика (в Hub включает ADR, спеки SA, vision US/AC):**
```
docs(spec): Wiki — новый модуль FR-WIKI-01..09, ERD, API-контракты, события hub.wiki.*
docs(spec): Calendar v2 — FR-CAL-01..08, ERD, availability и события hub.calendar.*
docs(vision): add US/AC for new Wiki module (space hierarchy, RBAC restrict, page versions)
docs(adr): add ADR-AUTH-SSO-2026 (accepted) - OIDC + LDAP/AD over Magic Link
docs(adr): add ADR-AI-AGENTS-MCP-2026 (accepted) - agent autonomy, MCP, LiteLLM-pattern gateway
docs(planning): STATE session note — Волна 2 SA-спеки, verified:false -> sa-ver
docs(platform): sync NATS trace propagation in playbook and state
```

**`chore` — deps, tooling, инфраструктурные скрипты без прод-поведения:**
```
chore(deps): bump @radix-ui/react-tooltip in /frontend
chore(deps-dev): bump tailwindcss from 4.2.2 to 4.3.1 in /frontend
chore(deps-dev): bump happy-dom from 20.10.5 to 20.10.6 in /frontend
chore(ops): FA-06 unified entry script + sync COVERAGE
chore(ops): K3s FA-06 P1 runbook and v2.5.0 release gate script
chore(platform): reorganize AI team, playbooks, roadmap and docs package
chore(infra): disable Dependabot version updates
```

**`perf` — производительность:**
```
perf(frontend): split recharts/plate and heavy routes into lazy chunks
perf(ui): default <Avatar> images to lazy loading
```

**`ci` — CI/CD:**
```
ci(deploy): remove DEV_MODE=true from production workflow
ci(frontend): FB-06 Lighthouse artifact on master push (#27)
ci(deploy): gate deploy on green CI
```

### 3.2. Хорошо / плохо

```
✅ feat(tracker): add WIP limits to Kanban columns
✅ fix(ws): prevent send on closed channel panic (SEC-05)
✅ test(auth): cover magic-link brute-force lockout (TC-AUTH-10)
✅ refactor(auth): extract AuthRepository interface
✅ ci(deploy): gate deploy on green CI

❌ update code                 (нет type/scope, бессмысленно)
❌ fix bug                     (какой?)
❌ feat: added stuff and fixed tests and refactored  (не атомарно)
❌ WIP                         (не коммитить в общие ветки)
❌ fix(auth): fix               (scope есть, описание — нет)
```

### 3.3. Атомарность

Один коммит = один проверяемый результат, обычно **≤200–300 строк** диффа. Если PR разросся — режь на несколько коммитов/PR. Крупная фича = серия атомарных коммитов, каждый оставляет `master` зелёным (после мержа).

<good-example>

Реальная атомарная серия из четырёх коммитов одной фичи (`feat/platform/otel-nats-trace`), каждый — самостоятельный проверяемый шаг:

```
3ad9383 feat(platform): persist W3C trace context on outbox events
c2ed605 feat(platform): propagate trace context via NATS headers on publish
47146cc feat(platform): continue producer trace in NATS consumers
dacb4ad test(platform): cover trace-context round-trip across outbox and NATS
7004df7 docs(platform): sync NATS trace propagation in playbook and state
```

Порядок логичен: сначала хранение trace-контекста (модель данных), затем публикация (producer), затем продолжение трейса у consumer'а, затем тест на весь round-trip, затем docs-sync. Каждый коммит компилируется и не ломает `master`.
</good-example>

<bad-example>

```
feat(platform): add tracing, fix bugs, update tests, refactor nats client
```

Один коммит смешивает фичу + фиксы + тесты + рефакторинг. Ревьюер не может проверить части отдельно, `git revert` откатит всё сразу, а не только сломавшуюся часть.
</bad-example>

Тело коммита — для *почему* и ссылок:
```
fix(auth): hash refresh tokens at rest (SHA-256)

Refresh-токены хранились в Redis в открытом виде — при утечке дампа
компрометировались все сессии (NF-12). Хешируем перед записью,
сверяем ConstantTimeCompare. См. docs/security_guidelines.md.

Refs: HUB-numbers, PLAN-CLOSE-GAPS.md#NF-12
```

### 3.4. Merge-коммиты волн (реальный формат Hub)

Когда параллельные ветки волны сходятся в интеграционную ветку, merge-коммит **не** переписывается в Conventional Commits — используется стандартное сообщение git с указанием источника:

```
Merge remote-tracking branch 'origin/docs/wave1-analysis' into integration/wave1
Merge remote-tracking branch 'origin/fix/platform/wave1-quickfixes' into integration/wave1
Merge branch 'docs/wave2-specs'
Merge branch 'docs/wave2-adr-package'
```

Это нормально и ожидаемо для интеграционных мержей (не для финального мержа фичи в `master` — там rebase-merge, см. §7.6).

---

## 4. Что НЕЛЬЗЯ коммитить

```gitignore
# Секреты
.env
.env.*                       (кроме .env.example)
livekit-prod.yaml            (содержит devkey: secret — только шаблон)
docker-compose.local.yml     (может содержать Resend/Telegram токены)

# Артефакты сборки/покрытия
backend/hub
frontend/dist/
coverage.out
backend/cover*  backend/cov_*  backend/*_cov*  backend/tmp*
*.out
tmp_*.py

# IDE
.idea/  *.swp
```

**Обязательный pre-commit скан на секреты:**
```bash
git diff --staged | rg -i "(secret|password|token|api[_-]?key|private[_-]?key|BEGIN .*PRIVATE)"
```
Если что-то нашлось — не коммитить, разобраться. Никогда не коммитить реальные ключи «для истории».

(2026) Обязательные CI-инструменты в дополнение к ручному скану: `govulncheck ./...` (+ `go mod verify`) для уязвимостей зависимостей и `gitleaks` для автоматического secrets-скана diff'а.

### 4.1. Реальный пример: почему это важно

Правило `livekit-prod.yaml` появилось не абстрактно: в конфиге LiveKit есть поле `devkey: secret`, которое выглядит как безобидный dev-дефолт, но в проде остаётся тем же значением, если не заменить его на реальный секрет через шаблонизацию (`deployments/.env.example` + подстановка). Коммит такого файла с реальным ключом = утечка на всю историю репозитория, которую нельзя вычистить без `git filter-repo`/BFG и форс-пуша (что для Hub запрещено правилом §12).

---

## 5. Параллельная работа be1/be2, fe1/fe2, qa-auto1/qa-auto2

Это ядро playbook'а. Правила, чтобы два+ агента одного слоя не конфликтовали.

### 5.1. Разделение по файлам/модулям

1. **Разделение по файлам/модулям.** На одну волну каждый агент владеет своим набором файлов. `be1` и `be2` **не редактируют один и тот же файл** одновременно. Границы назначает `tech-lead`/`team-lead` до старта (см. `team-workflow.md`).
2. **Общие контракты — заранее и заморожены.** Общие типы/DTO/интерфейсы/API фиксирует `sa`/`arch` до начала волны. Разработчик не меняет общий контракт в одностороннем порядке — сначала согласование.
3. **Частый rebase на `master`.** Перед пушем и хотя бы раз в день:
   ```bash
   git fetch origin && git rebase origin/master
   ```
   Rebase (а не merge) держит линейную историю; конфликты решаются локально и мелко.
4. **Маленькие PR, быстрый мерж.** Чем меньше PR и чем быстрее он вмержен, тем меньше расхождение веток.
5. **Порядок мержа** назначает `tech-lead`: если два PR трогают смежные области, они мержатся последовательно, второй ребейзится на первый.
6. **Кросс-ревью.** Перед персональным верификатором PR смотрит парный агент: `be1` ревьюит `be2` и наоборот (аналогично `fe`, `qa-auto`). Это ловит стилевые/логические расхождения между потоками.

### 5.2. Реальная модель нарезки модуля `tracker` под be1/be2

Модули Hub уже нарезаны по фичам именно для параллельной разработки — см. реальную структуру `backend/internal/tracker/`:

```
service_issues.go     ← be1: CRUD задач, optimistic locking
service_comments.go   ← be2: комментарии, треды
service_boards.go     ← владелец назначается на волну
service_workflow.go   ← workflow-переходы, статусы
service_time.go        ← time tracking
service_labels.go      ← метки
repository.go          ← ОБЩИЙ интерфейс Repository — зона риска, см. ниже
```

<good-example>

**Хороший сплит одной фичи (issue-linking):**

```
be1:  tracker/repo_links.go      (EntRepository методы CreateIssueLink/…)
      tracker/service_links.go   (бизнес-логика)
      tracker/handler_links.go   (HTTP)
be2:  tracker/repository.go      (добавил интерфейс IssueLinkRepository — смержил ПЕРВЫМ)
      затем работает над своим доменом (comments)
```

Общий `IssueLinkRepository` вынесен в отдельный интерфейс (ISP), `be2` смержил его первым, `be1` реализовал.
</good-example>

<bad-example>

**Плохо:** оба правят `service_issues.go` метод `UpdateIssue` в разных ветках → конфликт слияния + потерянная логика optimistic locking у одного из них.
</bad-example>

### 5.3. Аналогичный паттерн для fe1/fe2

Frontend-модули Hub нарезаны так же — по `modules/<domain>/` и `components/<feature>/`. Пример реальной параллельной волны (UX-1/UX-2, коммиты `3afd032`/`efe38d8`):

```
fe1: feat(messenger): mobile-friendly channel drawer (UX-1)
     — components/messenger/*, свой набор компонентов, не трогает routes/

fe2: refactor(routes): apply QueryStateView to projects/board/issues (UX-2)
     — routes/_authenticated/projects/**, свой набор роутов
```

Общий компонент `QueryStateView` (`components/query-state-view.tsx`) — вводится **одним** агентом, второй только потребляет через импорт, без правки самого компонента в параллельной ветке.

### 5.4. Аналогичный паттерн для qa-auto1/qa-auto2

`qa-auto1` владеет Python `testframework/{api,integration,load}/<domain>/**`,
`qa-auto2` — `testframework/{contract,e2e}/<domain>/**`. Genuine units принадлежат
разработчикам и живут только в `backend/tests/**` / `frontend/tests/**`.
Границы назначаются по exact paths на задачу; общий `testframework/core/**`
имеет одного writer-а и не редактируется параллельно.

### 5.5. Чек-лист координации (be1/be2, аналогично fe/qa-auto)

- [ ] Задача привязана к конкретному модулю/файлам; пересечений с параллельной задачей нет.
- [ ] Если трогается общий интерфейс/DTO — это анонсировано, PR мержится первым.
- [ ] Перед своим мержем сделан rebase на актуальный `master`.
- [ ] PR другого разработчика отревьюен (layering, errors, tests, регрессия).
- [ ] `.planning/STATE.md` обновлён session-заметкой.

---

## 6. Разрешение конфликтов (шаг за шагом)

### 6.1. Диагностика: конфликт — это сигнал плохо разделённых границ

Конфликт по общему файлу почти всегда означает, что границы были плохо разделены на старте волны → эскалация `tech-lead`/`team-lead` **до** попытки его вручную разрулить, если конфликт нетривиальный (затрагивает бизнес-логику, а не просто соседние строки).

### 6.2. Пошаговая процедура (rebase — стандартный путь для Hub)

```bash
# 1. Обновить локальный master
git fetch origin

# 2. Ребейзнуться на свежий master
git rebase origin/master

# 3. Если конфликт — git покажет список файлов
git status
#   both modified:   backend/internal/tracker/repository.go

# 4. Открыть файл, найти маркеры <<<<<<< / ======= / >>>>>>>, решить вручную
#    Правило: сохраняем ОБЕ смысловые правки, если они не взаимоисключающие
#    (например, оба добавили новый метод в интерфейс — оставляем оба метода).

# 5. Отметить файл как решённый
git add backend/internal/tracker/repository.go

# 6. Продолжить rebase (повторится для каждого коммита с конфликтом)
git rebase --continue

# 7. Если поняли, что рано начали rebase / нужно перегруппировать —
#    можно безопасно откатить всю операцию:
git rebase --abort

# 8. После успешного rebase — форс-пуш СВОЕЙ фича-ветки (не master!)
git push --force-with-lease origin <своя-ветка>
```

`--force-with-lease` (а не голый `--force`) — защита от перезаписи чужих коммитов, если кто-то запушил в эту же ветку между твоим fetch и push.

### 6.3. Кто решает конфликт, если оба агента правили общий файл

Решает **ребейзящийся** (тот, чей PR мержится позже) — он видит уже смерженную версию `master`/базовой ветки и разрешает конфликт локально, сохраняя семантику обеих правок. Автор более раннего (уже смерженного) PR не участвует в разрешении — его код уже часть базы.

### 6.4. После разрешения — обязательная регрессия

Конфликт мог незаметно сломать логику при выборе «неправильной» версии строки (например, потерять `version++` при optimistic locking). Поэтому после `git rebase --continue`/ручного мержа — **обязателен** полный прогон регрессии затронутого модуля (§9) до пуша, даже если до конфликта тесты были зелёными.

### 6.5. Конфликты в lock-файлах (`go.sum`, `package-lock.json`)

Не редактировать вручную построчно и не удалять lock-файл в общей dirty-копии.
Сначала определить владельца и выяснить, какой dependency change должен
победить. Если регенерация действительно нужна, отдельный writer выполняет
package-manager command только в task-owned ограниченном Docker, сверяет exact
diff/checksums и возвращает файл merge-owner-у; destructive удаление требует
отдельной authority на точный path.

---

## 7. PR-процесс

1. PR-описание ссылается на задачу/план (`PLAN.md`/тикет) и коротко объясняет *что и почему*.
2. Все CI-чеки зелёные: **lint + test + build** (backend: `go build -tags release` + `go test -tags hubtest -race`; frontend: `tsc --noEmit` + Biome + Vitest; e2e где применимо).
3. В диффе **нет** `DEV_MODE=true`, секретов, build/coverage-артефактов.
4. Пройдено кросс-ревью (парный агент) и **персональный `*-ver`** (`verified: true` + лог регрессии).
5. `code-reviewer` — approve перед мержем.
6. **Rebase-merge в `master`** (решение 2026-07-03: линейная история, атомарные коммиты планов сохранены для точечного revert; squash для фича-PR не используется).
7. **Push/мерж в `master` выполняет ТОЛЬКО пользователь лично.** Агенты коммитят и пушат фича-ветки свободно, но в master не пушат никогда. `team-lead` доводит PR до «merge-ready» и передаёт пользователю. Личное ревью пользователя обязательно для: security/auth, миграций БД, деплой/CI, ADR.

### 7.1. Реальные примеры PR из истории Hub (с номером)

```
feat(frontend): extend Ctrl+K palette with calls, settings, admin (#15)
ci(frontend): FB-06 Lighthouse artifact on master push (#27)
chore(ops): FA-06 unified entry script + sync COVERAGE (#26)
fix(ci): retry docker compose up on registry timeout (#16)
docs: полный аудит 2026-06-23 + sync docs + cleanup артефактов (#28)
```

Номер PR попадает в GitHub merge-сообщение автоматически — сохраняется как
есть, дополнительно переписывать не нужно.

### 7.2. Что проверяет `code-reviewer` перед approve

- Layering не нарушен (`*ent.Client` не утёк в Service/Handler — см. `backend.md` §2).
- Нет комментариев-мусора/пересказов (`code-comments.md`).
- Тесты покрывают новое поведение, а не только happy path.
- Docs-sync выполнен для user-visible/prod изменений (rule 55).
- Нет отладочных `console.log`/`fmt.Println`/закомментированного кода.

### 7.3. Размер PR

Ориентир — как и для коммита, но суммарно: PR **до ~500–800 строк диффа** ревьюится за один присест без потери качества. Больше — разбивать на серию (см. `split-to-prs` skill в общей команде, либо вручную по логическим этапам, как в примере §3.3 с `otel-nats-trace`).

### 7.4. Draft PR для промежуточных чекпоинтов волны

Если фича растянута на несколько дней/агентов, можно держать PR в статусе draft и пушить туда промежуточные атомарные коммиты — но **зелёная регрессия обязательна для каждого коммита**, а не только для финального перед переводом в ready-for-review.

### 7.5. Что делать при "red" CI после пуша

1. Не пушить новые коммиты «на удачу» — сначала посмотреть лог упавшего джоба (`gh run view <id> --log-failed` или UI).
2. Применить `test-triage` (`.cursor/skills/test-triage/SKILL.md`): сначала проверить, это баг в функционале или баг в тесте, **потом** чинить.
3. Фикс — отдельным атомарным коммитом с понятным сообщением (`fix(ci): retry docker compose up on registry timeout`), не через `--amend` уже запушенного коммита (см. §13).

### 7.6. Rebase-merge — почему не squash

Squash сплющивает атомарные коммиты плана в один — теряется возможность точечного `git revert` конкретного шага (например, откатить только тест, не откатывая фичу). Rebase-merge переносит все атомарные коммиты фича-ветки на вершину `master` без merge-коммита — история линейна, но детализация сохранена.

---

## 8. Интеграционные ветки волн

Для волновой разработки (несколько параллельных потоков — ba/sa/be/fe/qa на одну фичу-волну) Hub использует паттерн интеграционной ветки: `integration/<wave-name>`.

### 8.1. Реальный пример (`integration/wave1`)

```
669072c Merge remote-tracking branch 'origin/docs/wave1-analysis' into integration/wave1
b3bca09 Merge remote-tracking branch 'origin/fix/platform/wave1-quickfixes' into integration/wave1
954e6c3 docs(planning): wave1 closure - all streams verified, follow-ups to wave2
```

Каждый поток (аналитика, квик-фиксы) мержится в `integration/wave1` отдельным merge-коммитом со стандартным git-сообщением (не Conventional Commits — это техническая точка схождения, не смысловое изменение). После того как все потоки волны собраны и верифицированы, `team-lead` фиксирует closure-коммит (`docs(planning): wave1 closure — ...`) и передаёт интеграционную ветку пользователю на финальный мерж в `master`.

### 8.2. Когда использовать интеграционную ветку, а когда — прямой PR в master

| Сценарий | Путь |
|---|---|
| Одна фича, один/два агента, одна ветка | Прямой PR `feat/<module>/<desc>` → `master` |
| Волна из 3+ параллельных потоков (ba+sa+be+fe+qa) на одну крупную фичу | `integration/<wave-name>` собирает потоки → один PR в `master` |
| Быстрый квикфикс вне волны | Прямой PR `fix/<module>/<desc>` → `master` |

### 8.3. Правила для интеграционной ветки

- Владеет и мержит в неё **только** `team-lead` (или пользователь) — рядовые агенты пушат в свои фича-ветки, не напрямую в `integration/*`.
- Перед финальным мержем `integration/*` → `master` — полный `/qa-gate` на интеграции (rule 56), не по отдельности на каждый поток.
- `integration/*` тоже никогда не мержится в `master` агентом — только пользователем, как и любой другой путь в `master` (§14).

---

## 9. Обязательная регрессия перед коммитом/PR (REGULATION §7, rule 56)

Любое изменение, влияющее на поведение, проверяется **до** коммита только в
task-owned Docker с limits/cleanup. Канонические target jobs и команды —
`AGENTS.md` и `/qa-gate`: Go genuine units `backend/tests/**`, frontend genuine
units `frontend/tests/**`, весь non-unit QA — root `testframework/**` Python.
Host `go/npm/pytest/Playwright/build/lint` запрещены.

Результат (`PASS|FAIL|NOT_RUN`, exact revision, время, container/image) фиксируется
в evidence. Без required green нет handoff/PR. Цикл: новое поведение →
регрессия затронутого домена → на интеграции волны полный `/qa-gate`; детали —
`.cursor/rules/56-regression-verify.mdc` и `.cursor/playbooks/testing.md`.

**Важно:** «вроде зелёный» одним прогоном не считается закрытием цикла — при FAIL применяется `test-triage`, фикс, повтор до реального green (не переинтерпретация «наверное, флейк»).

---

## 10. Связь с docs-sync

Нельзя коммитить user-visible/prod-изменение без обновления документации в том же PR (rule 55):
- API → `system_specification.md` + `wiki/API.md`
- UI flow/hotkey → `wiki/SHORTCUTS.md`/`QUICKSTART.md`
- тесты/coverage → `docs/TESTING.md`, `.planning/COVERAGE.md`
- CI/deploy/Docker → `docs/stack_and_tools.md`, `deployments/.env.example`, `wiki/DEPLOY.md`
- security → `docs/security_guidelines.md`
- всегда → `.planning/STATE.md` (session note)

Реальный пример коммита, который правильно комбинирует код и docs-sync в одной фиче (но отдельными коммитами внутри одной ветки, см. §3.3):
```
c2ed605 feat(platform): propagate trace context via NATS headers on publish
dacb4ad test(platform): cover trace-context round-trip across outbox and NATS
7004df7 docs(platform): sync NATS trace propagation in playbook and state
```
Код → тест → docs-sync — три отдельных атомарных коммита одной ветки, каждый проверяемый по отдельности.

---

## 11. Теги и релизы

```
v1.0.0   major milestone
v1.0.1   hotfix
v2.0.0   новый milestone
```

Тег ставится на `master` после прохождения `/qa-gate` и sign-off. Релиз/деплой гейтится успешным CI (`devops-engineer`).

### 11.1. Реальный release gate Hub (`scripts/v25-release-gate.sh`)

Перед тегом `v2.5.0` в Hub используется скрипт-чек-лист (не авто-тег, только печать статуса):
```bash
./scripts/v25-release-gate.sh
```
Он проверяет наличие обязательных runbook-скриптов (`fa06-rotate-p1.sh`, `fa06-rotate-p1-k3s.sh`, `deploy-preflight.sh`), статус последнего CI на `master` (`gh run list --branch master --limit 1`) и отсутствие открытых PR. Только пользователь запускает финальный тег после того, как gate прошёл и sign-off получен — агенты **не создают** git-теги самостоятельно (эквивалент правила «push в master — только пользователь», §14).

---

## 12. Безопасные практики (жёсткие запреты)

| Нельзя | Почему |
|--------|--------|
| Менять `git config` | Может сломать идентичность/подписи команды |
| `push --force` в `master` | Уничтожает чужую историю |
| Пропускать хуки (`--no-verify`) | Обходит проверки безопасности/качества |
| `commit --amend` по запушенному коммиту | Требует force-push, ломает чужие клоны |
| **`Co-authored-by` / AI attribution в коммите** | Запрет пользователя: Cursor/Claude/Codex/Cloud — не соавторы (§3.0, `51-no-ai-coauthor.mdc`) |
| Коммитить секреты «временно» | Утечка навсегда в истории (нужен git-purge) |
| Гигантский PR «всё сразу» | Не ревьюится, не откатывается |
| `git rebase -i` в общих скриптах агента | Требует интерактивного ввода — агент зависнет |
| Тег/релиз без прохождения release gate | Прод без гарантии green CI |

`--amend` допустим только для **своего непзапушенного** последнего коммита (например, догрузить файлы, изменённые pre-commit хуком) — детали в §13.

---

## 13. Когда `--amend` нельзя (и когда можно)

### 13.1. Можно

- Коммит создан **вами** в этой же сессии.
- Коммит **ещё не запушен** (`git status` показывает «ahead of origin», не «up to date»/«diverged»).
- Причина — мелкая техническая правка того же коммита: pre-commit хук/линтер автоматически подправил файлы, которые нужно включить в тот же коммит, или опечатка в сообщении, которую вы заметили сразу же.

```bash
# Пример: хук отформатировал файл после коммита — догружаем его в тот же коммит
gofmt -w backend/internal/tracker/service_issues.go
git add backend/internal/tracker/service_issues.go
git commit --amend --no-edit
```

### 13.2. Нельзя

| Ситуация | Почему нельзя | Что делать вместо |
|---|---|---|
| Коммит уже запушен в свою фича-ветку, но ещё не в PR | Требует `--force-with-lease`, ломает чужие локальные копии ветки, если кто-то уже её тянул | Новый коммит `fix(...): ...` сверху |
| Коммит запушен и **виден в открытом PR** | Ревьюер уже мог оставить комментарии к конкретному diff'у коммита — amend их «осиротит» | Новый коммит; атомарная история сохранится при пользовательском rebase-merge |
| Коммит создан **не вами** (чужой коммит в общей ветке) | Переписывание чужой истории без согласия — потеря авторства и доверия | Попросить автора исправить, либо новый коммит с фиксом |
| CI-проверка/pre-commit hook **отклонил** коммит | Коммит не был создан — правило `--amend` вообще неприменимо | Исправить проблему → создать НОВЫЙ коммит (`always_applied` git-safety правило: если commit FAILED/REJECTED — никогда amend) |
| Коммит уже влит в `integration/*`/`master` | История разделяемая многими потоками волны | `git revert`, никогда amend |
| Нужно объединить 5 «WIP» коммитов в один осмысленный | `--amend` работает только с последним коммитом | `git rebase -i` **только интерактивно пользователем** — агентам запрещён (см. §12) |

### 13.3. Дерево решений

```
Коммит запушен?
├─ Нет → Мой последний коммит? → Да → amend можно (§13.1)
│                                → Нет (чужой) → новый коммит
└─ Да → Виден в открытом PR/влит в integration/master?
         → Да → НИКОГДА amend → новый коммит / revert
         → Нет, но кто-то мог утянуть ветку → тоже не amend без явной договорённости
```

---

## 14. Кто пушит/мержит в master — границы агента

Это самое строгое правило git-дисциплины Hub, вынесенное в `.cursor/rules/50-git.mdc` как always-applied.

| Действие | Кто может |
|---|---|
| Коммитить в фича-ветку | Любой агент, свободно, атомарно, после ворот регрессии (§9) |
| Пушить фича-ветку на remote | Только по явному поручению пользователя |
| Открывать PR из фича-ветки | Только по явному поручению пользователя |
| Мержить `integration/*` → доводить до merge-ready | `team-lead` |
| **Push/merge в `master`** | **ТОЛЬКО пользователь лично** |
| Создавать git-теги релиза | Только пользователь (после release gate, §11.1) |
| `git push --force` в любую ветку, включая свою | Только с `--force-with-lease` после rebase своей же ветки; force в `master` — никогда никому |

`team-lead` доводит ветку/PR до состояния «merge-ready» (green CI + верификация + `code-reviewer` approve) и **передаёт пользователю** — не мержит самостоятельно, даже если технически может выполнить команду. Личное ревью пользователя дополнительно обязательно для: security/auth, миграций БД, деплой/CI, ADR-изменений — независимо от того, что весь автоматический gate прошёл зелёным.

---

## 15. Антипаттерны git (сводка)

| Нельзя | Почему | Как правильно |
|--------|--------|---------------|
| Долгоживущая ветка | Ад конфликтов | Короткие ветки, частый мерж |
| Два агента на один файл | Конфликты, потеря работы | Разделить границы до старта (§5) |
| Merge вместо rebase на master | Шумная история | `git rebase origin/master` |
| Коммит без регрессии | Красный master | Прогон §9 до коммита |
| «update», «WIP», «fix» | Нечитаемая история | Conventional Commits (§3) |
| Смешать фичу + рефактор + формат | Не ревьюится | Отдельные атомарные коммиты |
| Ручное разрешение конфликта в `go.sum`/`package-lock.json` | Портит хеши | Пересобрать (`go mod tidy` / переустановка) |
| `--amend` запушенного/чужого коммита | Ломает историю других | Новый коммит / `git revert` (§13) |
| Агент мержит `master` самостоятельно | Нарушение rule 50-git | Передать пользователю как merge-ready (§14) |
| Тег релиза без release gate | Прод без гарантий | `scripts/v25-release-gate.sh` (или аналог) + sign-off |

---

## 16. Быстрая справка по командам

```bash
# Статус и синхронизация
git status
git fetch origin
git rebase origin/master
git rebase --continue
git rebase --abort

# Атомарный коммит
git add <файлы>
git commit -m "$(cat <<'EOF'
fix(tracker): prevent double version increment on concurrent update

Two goroutines could bump version twice in one request cycle under
high concurrency. See TC-TRCK-04.
EOF
)"

# Пуш своей фича-ветки (безопасный форс после rebase)
git push origin <ветка>
git push --force-with-lease origin <ветка>

# Секрет-скан перед коммитом
git diff --staged | rg -i "(secret|password|token|api[_-]?key|private[_-]?key|BEGIN .*PRIVATE)"

# Просмотр истории для стиля коммитов
git log --oneline -50

# PR (через gh)
gh pr create --title "feat(tracker): add WIP limits to Kanban columns" --body "..."
gh pr view --web
gh run list --branch master --limit 1
```

---

## 17. Чек-лист перед PR

- [ ] Ветка названа по конвенции; изменение атомарно и по теме.
- [ ] Коммиты — Conventional Commits; тело объясняет *почему* + ссылка на план/тикет.
- [ ] Ребейз на свежий `origin/master` выполнен; конфликты решены (§6), регрессия перепрогнана после разрешения.
- [ ] Регрессия green (лог + время); `/qa-gate` пройден (§9).
- [ ] Нет секретов/`DEV_MODE=true`/артефактов (скан §4 чист).
- [ ] Docs-sync выполнен; PR ссылается на задачу (§10).
- [ ] Кросс-ревью парного агента + персональный `*-ver` (verified:true).
- [ ] Никаких `--amend` запушенных/чужих коммитов (§13); история читаема.
- [ ] Готово к rebase-merge; ветка запушена, PR «merge-ready» — мерж в `master` выполняет **пользователь** (§14).
