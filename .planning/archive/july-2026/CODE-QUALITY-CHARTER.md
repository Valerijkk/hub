# Hub — Code Quality Charter

> **Статус:** принято для разработки 2026-07-10 · приоритет = корректности  
> **Владелец:** team-lead + code-reviewer · верификация: `be-ver` / `fe-ver` / `*-ver`  
> **Связь:** EPIC-QA-HARDENING (`.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md` § July Rebuild)

---

## 1. Приоритеты

| Приоритет | Правило |
|-----------|---------|
| **Correctness = Maintainability** | Одинаковый вес. Нельзя жертвовать читаемостью/слоями ради «быстрее закрыть задачу агентом». |
| **Security non-negotiables** | TZ + `docs/security_guidelines.md` — выше скорости и «удобства». |
| **YAGNI поверх speculative refactor** | Не плодить абстракции «на будущее»; чистить god-files только с причиной (дубль, рост LOC, нарушение слоя). |

Агент, который «работает», но оставляет нечитаемый diff, **не** считается done.

---

## 2. Принципы в контексте Hub

### SOLID (практично)

- **S** — один сервис/компонент = одна ответственность домена (не «utils на всё»).
- **O/L** — расширение через интерфейсы Repository/Service; не ломать контракты handler'ов без ADR.
- **I** — узкие интерфейсы репозитория под use-case, не god-`Repository` на 50 методов без нужды.
- **D** — Handler → Service → Repository; **запрещено** `*ent.Client` в Service или Handler (layering).

### DRY / KISS / YAGNI

- **DRY:** общая логика — в service/shared helper с тестом; копипаст RBAC/authz — red flag.
- **KISS:** простой путь в Fiber/React предпочтительнее «framework в framework».
- **YAGNI:** нет adapter SFU «на всякий случай» (LiveKit остаётся каноном); нет premature microservices вне ADR.

### Backend (Go)

```
Handler (HTTP/WS) → Service (бизнес) → Repository (Ent/Mongo/Redis)
```

- Context первым аргументом; ошибки wrap (`fmt.Errorf("…: %w", err)`).
- Optimistic locking `version`; UUID v4 PK; UTC в БД.
- Канон: `docs/code_style_guide.md` · `.cursor/playbooks/backend.md` · rule `10-style` / `20-backend`.

### Frontend (React)

- Rules of Hooks: все хуки до любого early return.
- TanStack Router + Query — канон; Shadcn: только `className`, Radix не ломать.
- Стабильные `key` (id), не `key={index}`.
- Канон: `docs/code_style_guide.md` · `.cursor/playbooks/frontend.md` · rule `30-frontend`.

### Comments

- Только **WHY** (инвариант, workaround, security). Не WHAT.
- Playbook: `.cursor/playbooks/code-comments.md`.

---

## 3. Definition of Done (be / fe PR)

PR merge-ready только если:

1. **Читаемость** — новый читатель понимает поток за ≤15 мин без устных пояснений.
2. **Слои** — нет `*ent.Client` в service/handler; нет бизнес-логики в handler сверх map/validate.
3. **Тесты** — юниты/интеграция в `testframework/` на happy + отказ (403/404/validation); доменная регрессия green.
4. **Нет god-files без причины** — рост файла >~400–500 LOC или смешение concerns → split в том же PR или follow-up task с ссылкой.
5. **Comments** — WHY only; нет placeholder/`TODO` без ticket.
6. **Docs-sync** — STATE + domain wiki/API при user-visible/API change (rule `55-docs-sync`).
7. **Security** — нет антипаттернов из §4; секреты не в diff.

---

## 4. Антипаттерны (запрещено)

| Антипаттерн | Почему | Альтернатива |
|-------------|--------|--------------|
| `panic()` в production paths | Crash процесса | return error / typed sentinel |
| `http.DefaultClient` на user URL | SSRF / no timeout | `safehttp.NewSafeHTTPClient` |
| `io.ReadAll` без `LimitReader` | Memory DoS | `io.LimitReader` + max |
| `!=` для секретов | Timing attack | `subtle.ConstantTimeCompare` / hmac.Equal |
| `*ent.Client` в Service/Handler | Layering | Repository interface |
| Hardcoded credentials / `DEV_MODE=true` в prod | Secrets leak | env + examples only |
| `key={index}` в списках | Reconciliation bugs | stable id |
| `!important` в CSS | Специфичность | токены `--hub-*` / className |
| Skip webhook signature / WS без membership | SEC-02 / NF-02 | fail-closed checks |
| Search без user-scope | NF-03 | filter by membership |

Полный список TZ: `docs/tz_constraints.md` · rule `60-tz-constraints`.

---

## 5. Ревью и верификация

| Роль | Что проверяет |
|------|----------------|
| **code-reviewer** | Diff pre-merge: баги, security, стиль, тесты, docs-sync, god-files |
| **be-ver** | Layering, errors, race, context, Ent boundaries |
| **fe-ver** | Hooks, a11y, дизайн-система, типы, регрессия UI |
| **qa-auto-ver** / **qa-manual-ver** | Детерминизм тестов, трассируемость TC↔автотест |
| **security-auditor** | OWASP/ASVS на auth/RBAC/webhook/SSRF |

`verified: true` — только при green доменной регрессии + лог (rule `56-regression-verify`).

---

## 6. Связанные артефакты

| Документ | Назначение |
|----------|------------|
| `docs/code_style_guide.md` | Go/TS стиль |
| `.cursor/playbooks/backend.md` | BE практики |
| `.cursor/playbooks/frontend.md` | FE практики |
| `.cursor/playbooks/code-comments.md` | Комментарии WHY |
| `.cursor/rules/10-style.mdc` | Универсальный стиль |
| `.cursor/rules/20-backend.mdc` | Backend rules |
| `.cursor/rules/30-frontend.mdc` | Frontend rules |
| `.planning/archive/july-2026/INDEX-JULY-2026.md` | Индекс July planning |
| `testframework/TESTING_GUIDE.md` | Как писать/гонять тесты (QA-HARDENING) |

---

*Обновлять при ADR, меняющем layering или запреты TZ.*
