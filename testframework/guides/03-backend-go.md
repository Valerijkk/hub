# 03 — Backend Go (`backend/testframework`)

> Часть [Testing Guide](../TESTING_GUIDE.md). Playbook: [`.cursor/playbooks/testing.md`](../../.cursor/playbooks/testing.md) §4.  
> Pipeline: [`../PIPELINE.md`](../PIPELINE.md) · job `backend-test`.

---

## Где живут тесты

```
backend/testframework/
├── env/           # Builder (sqlite Ent)
├── factory/       # UserFactory и др.
├── auth/          # домен auth
├── tracker/
├── messenger/
├── tms/
├── calls/         # LiveKit (SFU migration rejected)
└── {domain}/
```

Почему не `testframework/stages/unit/backend-test/`? Go запрещает импорт `backend/internal/*` из другого модуля. Логически — unit-стадия CI; физически — модуль `backend`.

Владелец automation-слоя: **`qa-auto1`**. Unit-логику сервисов пишут **`be1`/`be2`** вместе с фичей.

---

## Каркас файла

```go
package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.

import (
	"testing"
	"github.com/valerii/hub/backend/internal/auth"
	// …
)

// TC-AUTH-…: краткое описание сценария.
func TestVerifyTelegramAuth_InvalidHash(t *testing.T) {
	// Arrange → Act → Assert
}
```

White-box через `export_hooks.go` + `//go:build hubtest` — playbook §4.10. В prod (`-tags release`) hooks не компилируются.

---

## Выбор изоляции

| Нужно | Инструмент | Пример |
|-------|------------|--------|
| Логика сервиса | `mockRepo` (map) | `ai/privacy_test.go`, playbook §4.7 |
| Реальная схема Ent | `env.New(t, prefix).Build()` | `tracker/repository_test.go` |
| Специфика PG/Mongo | testcontainers / CI Mongo | `*_mongo_test.go` |

**По умолчанию — mock.** Sqlite Builder — когда важны каскады/индексы. Docker в каждом unit — anti-YAGNI.

### Builder (реальный API)

```go
// backend/testframework/env/builder.go
e := env.New(t, "tracker").Build()
// e.Client, e.Ctx, e.UserID, e.ProjectID, e.StatusID
// DSN: file:{prefix}_{uuid}?mode=memory — изоляция + t.Cleanup(Close)
```

### Factory

```go
// backend/testframework/factory/user.go
uf := factory.NewUserFactory(t, e.Client, e.Ctx)
alice := uf.Create("alice") // email = alice-{uuid8}@test.com
```

### mockRepo (контракт)

```go
// Good — эмуляция optimistic lock
if m.updateVersionError || i.Version != input.Version {
	return nil, apperrors.NewConflict("version mismatch")
}

// Bad
return nil, nil // «зелёный» тест без смысла
```

Partial mock: embed интерфейс, переопредели нужные методы (playbook §4.7, `tms/allure_parser_test.go`).

---

## Table-driven + Parallel

```go
func TestGenerateCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{name: "ok", in: "a@b.c", want: nil},
		{name: "empty", in: "", want: auth.ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// изолированные данные на итерацию
		})
	}
}
```

- `require` — стоп при ошибке setup / перед разыменованием.  
- `assert` — мягкие проверки после успешного act.  
- `t.Helper()` во всех setup-хелперах.  
- `t.Cleanup` для Close/Terminate.

---

## Именование файлов

| Паттерн | Когда |
|---------|-------|
| `service_test.go` | базовые unit сервиса |
| `service_wave<NN>_test.go` | волна coverage / фича-волна |
| `handler_<feature>_test.go` | HTTP/httptest |
| `repository_test.go` | Ent repo + Builder |
| `*_mongo_test.go` | Mongo-специфика |
| `*_security_test.go` / `*_limits_test.go` | security / границы |

Имена функций: `Test{Subject}_{Scenario}` — см. `auth/service_test.go` (`_ValidSignature`, `_InvalidHash`, `_ExpiredAuth`).

---

## Happy / negative на backend

| Тип | Пример из репо |
|-----|----------------|
| Happy | `TestVerifyTelegramAuth_ValidSignature` |
| Negative | `TestVerifyTelegramAuth_InvalidHash`, `_MissingHash` |
| Edge / expiry | `TestVerifyTelegramAuth_ExpiredAuth` |
| Repo counter | `TestEntRepository_CreateIssue_IncrementsCounter` |

Error paths: 401/403/404/409/429 — на handler/service с `apperrors`, не только «err != nil».

---

## Прогон и coverage

```bash
cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
cd backend && go test -tags hubtest -race -count=1 ./testframework/tracker/...
```

Coverage-профиль — в TMP (rule 45), gate ≥82%, security-пакеты ≥80%.

---

## Чек-лист backend-теста

- [ ] `package *_test` (или hubtest hooks осознанно)
- [ ] TC-ID в комментарии; колонка в `docs/TESTING.md`
- [ ] Изоляция (uuid DSN / свой mock)
- [ ] `require` на setup; контрактные ошибки в mock
- [ ] `-race` зелёный; нет `time.Sleep`
- [ ] Артефакты `*.out` / `cover*` удалены
