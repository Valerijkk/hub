---
name: amend-push
description: git add → commit --amend → push --force-with-lease → дождаться CI. Используй когда нужно добавить правки в последний коммит и синхронизировать с remote.
disable-model-invocation: true
---

# /amend-push — Amend Last Commit & Push

⚠️ Использовать ТОЛЬКО если последний коммит ещё не смержен в main.

## Step 1: Проверить состояние

```bash
cd C:/Users/Valerii/Desktop/hub
git status
git log --oneline -5
```

Убедиться что HEAD commit создан в текущей сессии.

## Step 2: Stage правки

```bash
git add -p   # интерактивно (рекомендуется)
# или
git add <specific-files>
# или
git add -u   # все изменённые
```

## Step 3: QA Gate перед amend

Запусти `/qa-gate` или минимум:
```bash
cd backend && go build ./... && go test -race ./... -count=1
cd frontend && npx tsc --noEmit
```

## Step 4: Amend

```bash
git commit --amend --no-edit
# или если нужно изменить сообщение:
git commit --amend -m "feat(module): updated description"
```

## Step 5: Push с защитой

```bash
git push --force-with-lease
```

`--force-with-lease` безопаснее `--force` — откажет если кто-то другой пушнул после тебя.

## Step 6: Дождаться CI

```bash
# Проверить статус через gh CLI
gh run list --limit 3
gh run watch <run-id>
```

Или открыть GitHub Actions в браузере.

## Чеклист

- [ ] Коммит существует (не пустой amend)
- [ ] QA gate прошёл
- [ ] `force-with-lease` использован, не `force`
- [ ] CI зелёный после пуша
