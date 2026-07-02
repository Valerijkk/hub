# Hub — Playbook: Git workflow

> Для be1/be2, fe1/fe2, qa-auto1/qa-auto2, devops, code-reviewer. Первоисточник: `.cursor/rules/50-git.mdc` (при конфликте правило важнее).
> Особый акцент — **параллельная работа двух разработчиков** одного слоя без конфликтов.

*Обновлено: 2026-07-02.*

---

## 1. Зачем дисциплина git в команде из многих агентов

В Hub одновременно работают пары исполнителей (be1/be2, fe1/fe2, qa-auto1/qa-auto2). Без дисциплины они наступают друг другу на файлы, ломают историю и создают конфликты мержа, которые стоят дороже самой фичи. Правила ниже делают параллельную разработку предсказуемой.

Три опоры:
1. **Маленькие атомарные коммиты/PR** — легко ревьюить, легко откатывать.
2. **Чёткие границы файлов** — параллельные агенты не редактируют одно и то же.
3. **Частый rebase на `master`** — конфликты ловятся рано и мелкими.

---

## 2. Ветки

```
master                              production-ready, protected
feat/<module>/<description>         новая фича
fix/<module>/<description>          багфикс
refactor/<module>/<description>     рефакторинг
chore/<description>                 deps, tooling, config
hotfix/<issue-id>/<description>     критический прод-фикс
```

- `<module>` — из списка scope (см. §3).
- `<description>` — kebab-case, коротко и по делу: `feat/tracker/wip-limits`.
- Одна ветка = одна логическая задача/план. Не смешивай несвязанные изменения.
- Ветку держим короткоживущей: чем дольше живёт, тем больше конфликтов.

---

## 3. Conventional Commits

```
<type>(<scope>): <описание в повелительном наклонении>

[тело: что и ПОЧЕМУ; ссылка на PLAN/тикет]
```

**Типы:** `feat | fix | refactor | test | docs | chore | perf | ci`
**Scope:** `auth | tracker | sprint | messenger | tms | calls | dashboard | notification | search | ws | rbac | platform | frontend | infra | ent | ai | crm | intranet | calendar | hr | accounting | bases | marketplace`

### Хорошо / плохо

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
```

### Атомарность

Один коммит = один проверяемый результат, обычно **≤200–300 строк** диффа. Если PR разросся — режь на несколько коммитов/PR. Крупная фича = серия атомарных коммитов, каждый оставляет `master` зелёным (после мержа).

Тело коммита — для *почему* и ссылок:
```
fix(auth): hash refresh tokens at rest (SHA-256)

Refresh-токены хранились в Redis в открытом виде — при утечке дампа
компрометировались все сессии (NF-12). Хешируем перед записью,
сверяем ConstantTimeCompare. См. docs/security_guidelines.md.

Refs: HUB-numbers, PLAN-CLOSE-GAPS.md#NF-12
```

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

---

## 5. Параллельная работа be1/be2, fe1/fe2, qa-auto1/qa-auto2

Это ядро playbook. Правила, чтобы два агента одного слоя не конфликтовали:

1. **Разделение по файлам/модулям.** На одну волну каждый агент владеет своим набором файлов. be1 и be2 **не редактируют один и тот же файл** одновременно. Границы назначает `tech-lead`/`team-lead` до старта (см. `team-workflow.md`).
2. **Общие контракты — заранее и заморожены.** Общие типы/DTO/интерфейсы/API фиксирует `sa`/`arch` до начала волны. Разработчик не меняет общий контракт в одностороннем порядке — сначала согласование.
3. **Частый rebase на `master`.** Перед пушем и хотя бы раз в день:
   ```bash
   git fetch origin && git rebase origin/master
   ```
   Rebase (а не merge) держит линейную историю; конфликты решаются локально и мелко.
4. **Маленькие PR, быстрый мерж.** Чем меньше PR и чем быстрее он вмержен, тем меньше расхождение веток.
5. **Порядок мержа** назначает `tech-lead`: если два PR трогают смежные области, они мержатся последовательно, второй ребейзится на первый.
6. **Кросс-ревью.** Перед персональным верификатором PR смотрит парный агент: be1 ревьюит be2 и наоборот (аналогично fe, qa-auto). Это ловит стилевые/логические расхождения между потоками.

### Разрешение конфликтов
- Конфликт по общему файлу = сигнал, что границы были плохо разделены → эскалация `tech-lead`.
- Решаем на стороне ребейзящегося (более позднего) PR, сохраняя обе смысловые правки.
- После разрешения — обязательно перепрогнать регрессию (конфликт мог сломать логику).

---

## 6. PR-процесс

1. PR-описание ссылается на задачу/план (`PLAN.md`/тикет) и коротко объясняет *что и почему*.
2. Все CI-чеки зелёные: **lint + test + build** (backend: `go build -tags release` + `go test -tags hubtest -race`; frontend: `tsc --noEmit` + Biome + Vitest; e2e где применимо).
3. В диффе **нет** `DEV_MODE=true`, секретов, build/coverage-артефактов.
4. Пройдено кросс-ревью (парный агент) и **персональный `*-ver`** (`verified: true` + лог регрессии).
5. `code-reviewer` — approve перед мержем.
6. **Rebase-merge в `master`** (решение 2026-07-03: линейная история, атомарные коммиты планов сохранены для точечного revert; squash для фича-PR не используется).
7. **Push/мерж в `master` выполняет ТОЛЬКО пользователь лично.** Агенты коммитят и пушат фича-ветки свободно, но в master не пушат никогда. team-lead доводит PR до «merge-ready» и передаёт пользователю. Личное ревью пользователя обязательно для: security/auth, миграций БД, деплой/CI, ADR.

---

## 7. Обязательная регрессия перед коммитом/PR (REGULATION §7)

Любое изменение, влияющее на поведение → прогон **до** коммита:

```bash
# Backend
cd backend && go build -tags release ./... && go vet ./... && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Frontend
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

Или `/qa-gate` (останавливается на первой FAIL). Результат (PASS/FAIL + время) фиксируется в skill-логе. Без green — нет handoff/PR. Детали — `.cursor/playbooks/testing.md`.

---

## 8. Связь с docs-sync

Нельзя коммитить user-visible/prod-изменение без обновления документации в том же PR (rule 55):
- API → `system_specification.md` + `wiki/API.md`
- UI flow/hotkey → `wiki/SHORTCUTS.md`/`QUICKSTART.md`
- тесты/coverage → `docs/TESTING.md`, `.planning/COVERAGE.md`
- всегда → `.planning/STATE.md` (session note)

---

## 9. Теги и релизы

```
v1.0.0   major milestone
v1.0.1   hotfix
v2.0.0   новый milestone
```

Тег ставится на `master` после прохождения `/qa-gate` и sign-off. Релиз/деплой гейтится успешным CI (devops).

---

## 10. Безопасные практики (жёсткие запреты)

| Нельзя | Почему |
|--------|--------|
| Менять `git config` | Может сломать идентичность/подписи команды |
| `push --force` в `master` | Уничтожает чужую историю |
| Пропускать хуки (`--no-verify`) | Обходит проверки безопасности/качества |
| `commit --amend` по запушенному | Требует force-push, ломает чужие клоны |
| Коммитить секреты «временно» | Утечка навсегда в истории (нужен git-purge) |
| Гигантский PR «всё сразу» | Не ревьюится, не откатывается |

`--amend` допустим только для **своего непзапушенного** последнего коммита (например, догрузить файлы, изменённые pre-commit хуком).

---

## 11. Антипаттерны git (сводка)

| Нельзя | Почему | Как правильно |
|--------|--------|---------------|
| Долгоживущая ветка | Ад конфликтов | Короткие ветки, частый мерж |
| Два агента на один файл | Конфликты, потеря работы | Разделить границы до старта |
| Merge вместо rebase на master | Шумная история | `git rebase origin/master` |
| Коммит без регрессии | Красный master | Прогон §7 до коммита |
| «update», «WIP», «fix» | Нечитаемая история | Conventional Commits |
| Смешать фичу + рефактор + формат | Не ревьюится | Отдельные атомарные коммиты |

---

## 12. Чек-лист перед PR

- [ ] Ветка названа по конвенции; изменение атомарно и по теме.
- [ ] Коммиты — Conventional Commits; тело объясняет *почему* + ссылка на план/тикет.
- [ ] Ребейз на свежий `origin/master` выполнен; конфликты решены.
- [ ] Регрессия green (лог + время); `/qa-gate` пройден.
- [ ] Нет секретов/`DEV_MODE=true`/артефактов (скан §4 чист).
- [ ] Docs-sync выполнен; PR ссылается на задачу.
- [ ] Кросс-ревью парного агента + персональный `*-ver` (verified:true).
- [ ] Готово к rebase-merge; ветка запушена, PR «merge-ready» — мерж в master выполняет пользователь.
