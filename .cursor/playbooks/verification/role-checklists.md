# Верификация: чек-листы по ролям

> Часть плейбука. Канон и оглавление — [`../verification.md`](../verification.md).

Что проверяет каждый верификатор в своей зоне.

---

## 3. Чек-листы по ролям верификаторов

Каждый чек-лист — конкретные проверяемые пункты. Пункт считается пройденным только при
воспроизведении (не при заявлении). Рядом — как проверять и good/bad.

### 3.1 be-ver — верификатор Backend

Skill: `BE_VER_skill.md`. Артефакт: Go-код в `backend/internal/<module>`. Запускает тесты сам.

Чек-лист:

```
[ ] Соответствие system_specification и architecture (FR/контракт реализован как описан)
[ ] Layering: Handler → Service → Repository → DB; НЕТ *ent.Client в Service/Handler
[ ] Repository pattern соблюдён; нет GetDB() escape hatch в новом коде
[ ] Ошибки: обёрнуты (fmt.Errorf("...: %w")); формат ответа {"error","request_id"} (NFR-13)
[ ] context.Context первым аргументом; проброшен во все слои
[ ] UUID v4 PK; version int (optimistic lock, 409 при конфликте); UTC; soft delete
[ ] Нет hardcoded secrets; нет DEV_MODE=true в prod-путях
[ ] Security-инварианты: SafeHTTPClient+ValidateOutboundURL (SSRF); io.LimitReader (не ReadAll);
    subtle.ConstantTimeCompare для секретов; webhook signature обязателен; RBAC до сервиса;
    WS room authz по membership; search user-scope
[ ] Нет panic() в production-путях
[ ] go build -tags release ./... && go vet ./... — чисто
[ ] go test -race -count=1 ./... — PASS (лог приложен)
[ ] Unit-тесты: граничные случаи, классы эквивалентности, негативные сценарии
[ ] Атомарность: ≤300 строк осмысленного кода за итерацию
[ ] При FAIL теста: triage выполнен (ROOT=code|test, обоснование) — REGULATION §9
[ ] slog с контекстом (не fmt.Println/log.Printf)
```

Команды воспроизведения:

```bash
cd backend && go build -tags release ./... && go vet ./...
go test -race -count=1 -timeout=600s ./...
go test -cover ./internal/<module>/...
# канонично (Docker, -tags hubtest): см. REGULATION §7 / AGENTS.md
```

Как проверять layering (пример поиска нарушения):

```
Grep: "\*ent\.Client" в service.go/handler*.go модуля → должно быть пусто
Grep: "GetDB()" в новом коде → escape hatch, замечание HIGH
```

#### Good / Bad: be-ver проверка

```
GOOD
- Запустил go test -race сам, приложил лог "ok ./internal/tms 12.4s"
- Grep подтвердил отсутствие *ent.Client в service.go
- Проверил 409 при устаревшем version через тест — присутствует
- Замечание [CRITICAL]: webhook handler пропускает signature при отсутствии claim (SEC-02)
```

```
BAD
- "BE написал, что тесты проходят" → поверил без запуска (нарушение §1.2)
- Не проверил формат ошибки → пропустил {"message"} вместо {"error","request_id"}
- Поставил verified: true без лога регрессии (нарушение §6)
```

### 3.2 fe-ver — верификатор Frontend

Skill: `FE_VER_skill.md`. Артефакт: React/TS в `frontend/src`.

Чек-лист:

```
[ ] Rules of Hooks: все хуки до любого return; нет хуков после условия
[ ] Соответствие дизайн-системе (Shadcn/ui + Tailwind 4; только className, Radix не трогать)
[ ] Нет !important в CSS (TZ forbidden)
[ ] Стабильные key в списках (item.id, НЕ index)
[ ] Accessibility/WCAG 2.1 AA: контраст 4.5:1 текст / 3:1 UI; aria-атрибуты; keyboard nav;
    focus-visible; явный type у button
[ ] Адаптивность: mobile bottom nav / desktop sidebar; проверено на брейкпоинтах
[ ] Безопасность: токен в памяти, НЕ в localStorage; нет dangerouslySetInnerHTML без sanitize (XSS)
[ ] optional chaining/?? вместо non-null ! без guard
[ ] tsc --noEmit — чисто; npm run lint (Biome) — чисто
[ ] Vitest — PASS (лог); Playwright E2E критических путей — PASS
[ ] Lighthouse ≥ 90 (цель; если gate не blocking — зафиксировать факт)
[ ] Соответствие TanStack Router (единственный роутер); TanStack Query + Zustand
[ ] Описание от FE соответствует коду
[ ] Атомарность изменения
```

Команды:

```bash
cd frontend && node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit && npm run lint && npx vitest run
npx playwright test
```

Как проверять Rules of Hooks (пример):

```
Grep в компоненте: return перед useQuery/useState/useEffect → нарушение [HIGH]
```

#### Good / Bad: fe-ver

```
GOOD
- Нашёл useEffect после `if (!id) return null` → замечание [HIGH] Rules of Hooks
- Проверил key={item.id} в списках; axe login PASS приложил
- Подтвердил токен в Zustand-памяти, не в localStorage
```

```
BAD
- "Выглядит адаптивно" без проверки брейкпоинтов → «на глаз» (антипаттерн)
- Пропустил key={index} → баг реконсиляции
- verified: true без vitest-лога
```

### 3.3 qa-auto-ver — верификатор автотестов

Skill (базово): `QA_VER_skill.md`. Артефакт: `testframework/**`, Vitest, Playwright.

Чек-лист:

```
[ ] Детерминизм: тест проходит при повторных запусках (-count=3 / несколько прогонов)
[ ] Отсутствие flaky: нет зависимости от времени/порядка/сети без мока
[ ] НЕТ sleep/произвольных таймаутов; ожидание по условию (waitFor/poll), не по времени
[ ] Привязка TC ↔ autotest: каждый autotest ссылается на TC/AC (трассируемость)
[ ] Покрытие: новый код покрыт; overall ≥50% gate, security-critical ≥80%
[ ] Изоляция: тесты не делят состояние; setup/teardown чистые; параллельно-безопасны
[ ] Ассерты содержательны (проверяют результат, не «не упало»)
[ ] Негативные и граничные кейсы присутствуют (не только happy path)
[ ] race-detector зелёный (backend); нет data race
[ ] Тесты в правильном месте пайплайна (testframework/{domain}, stages/...)
[ ] При FAIL: triage (ROOT=code|test) — не «озеленение» без проверки функционала (§10)
```

Команды:

```bash
# backend детерминизм
go test -race -count=3 ./internal/<module>/...
# frontend
npx vitest run
npx playwright test --repeat-each=2
```

#### Good / Bad: qa-auto-ver

```
GOOD
- Запустил -count=3, все зелёные → детерминизм подтверждён
- Нашёл time.Sleep(2s) в тесте → замечание [MEDIUM], заменить на poll по условию
- Проверил, что flaky_test.go ссылается на TC-TMS-14 → трассировка есть
```

```
BAD
- Запустил один раз, объявил «стабильно» → flaky не выявлен
- Принял тест с assert «err == nil» без проверки результата → пустой тест
- Принял -count=1 без race → пропустил гонку
```

### 3.4 qa-manual-ver — верификатор ручного тест-плана

Skill (базово): `QA_VER_skill.md`. Артефакт: `docs/test_plan.md` / TMS-кейсы.

Чек-лист:

```
[ ] test_plan полон: пирамида, критерии входа/выхода, регрессия, инструменты
[ ] Каждая Acceptance Criteria из vision_and_scope имеет TC (трассируемость AC↔TC)
[ ] Техники тест-дизайна применены: классы эквивалентности, граничные значения,
    таблицы решений, переходы состояний, попарное тестирование где уместно
[ ] Три вида сценариев на каждую US: основной / альтернативный / исключительный
[ ] Матрица покрытия требований актуальна (COVERAGE.md)
[ ] Предусловия/шаги/ожидаемый результат TC однозначны и воспроизводимы
[ ] Баг-репорты корректны: шаги, факт vs ожидание, severity, окружение
[ ] Smoke-набор на билде определён и проходит
[ ] Регрессионный набор привязан к изменяемым областям
```

Как проверять трассировку AC↔TC:

```
Для каждой US в vision → найти TC в test_plan/TMS → если AC без TC = Gap [HIGH]
```

#### Good / Bad: qa-manual-ver

```
GOOD
- Сверил US-S16 AC (3 вида) → нашёл TC на happy/фильтр/409 → покрыто
- Проверил граничные значения limit (0,1,100,101) в TC листинга
```

```
BAD
- "Тест-план выглядит полным" без сверки AC↔TC → пропущенные критерии
- Принял TC только на happy path → нет негативных/граничных
```

### 3.5 ba-ver — верификатор Business Analyst

Skill: `BA_VER_skill.md`. Артефакт: `docs/vision_and_scope.md`.

Чек-лист:

```
[ ] Бизнес-цели SMART: конкретны, измеримы, достижимы, релевантны, со сроком
[ ] KPI измеримы и отражают ФАКТ (со ссылкой на источник), не желаемое
[ ] ≥5 User Stories, каждая атомарна и в формате «Как/хочу/чтобы»
[ ] Каждая US имеет AC (три вида: happy/alt/error), проверяемые
[ ] MoSCoW без противоречий (нет дублей приоритета одной фичи; напр. Web Push Should И Could)
[ ] Персоны покрывают ключевые роли; JTBD связаны с US
[ ] Сценарии: основной, альтернативный, исключительный
[ ] Нет противоречий с system_specification (числа/scope; напр. API count едино)
[ ] Соответствие TZ scope (Won't Have vs mandatory features by phase)
[ ] Описание BA (2–5 предложений) соответствует реальному содержимому
[ ] Нумерация US/G/P без коллизий (для параллельных ba1/ba2)
[ ] Открытые гипотезы помечены, не «протекли» в verified без риска
```

Как проверять (независимо):

```
- Сам пройти таблицу US, посчитать, проверить формат каждой
- Сам сверить MoSCoW: grep фичи по категориям, искать дубли
- Сам сверить число endpoints/метрик с system_specification
```

#### Good / Bad: ba-ver (реальный стиль из лога)

```
GOOD (FAIL с конкретикой)
FAIL: G-05 ✓ vs KPI pending; Web Push Should/Could; Lighthouse v2.0 vs NFR-17 v2.5;
      KPI API ~230 vs SA 293; нет US→PROJECT traceability matrix
```

```
BAD
- "US в целом нормальные" → нет пофактовой проверки
- Не сверил KPI с spec → пропустил drift чисел
```

### 3.6 sa-ver — верификатор System Analyst

Skill: `SA_VER_skill.md`. Артефакт: `docs/system_specification.md` + `wiki/API.md`.

Чек-лист:

```
[ ] Все User Stories из vision покрыты технически (FR c Source=US)
[ ] FR тестируемы (поведение, не «сделать»); NFR числовые/проверяемые
[ ] ERD корректна: UUID v4, version, UTC, soft delete; связи описаны явно с ключами
[ ] API-контракты непротиворечивы с architecture; формат ошибки {"error","request_id"};
    пагинация offset/cursor; коды 4xx/409/422; rate limits
[ ] Endpoint inventory сверён с кодом (независимый пересчёт RegisterRoutes)
[ ] Матрица трассируемости (COVERAGE) актуальна; gap помечены
[ ] Нет нарушений TZ mandatory (auth/data/api/real-time/storage)
[ ] Описание SA соответствует спецификации (counts верны)
[ ] Межмодульные связи согласованы (нет прямого доступа к чужому Ent)
```

Метод (из реального лога sa-ver):

```
Independent recount via RegisterRoutes block parser на всех handler-файлах +
cross-check таблицы wiki/API.md + spot-check CRITICAL fixes в коде.
Пример вывода: SA claimed 293, verified 285 (−8); per-module: search 7→5, notification 11→10...
```

#### Good / Bad: sa-ver

```
GOOD
- Сам пересчитал routes → нашёл расхождение 293 vs 285, замечание [HIGH] с разбивкой по модулям
- Проверил TestCase на version → отсутствует → [CRITICAL] нарушение NFR-24
```

```
BAD
- Принял «285 REST» на слово → drift не пойман
- Не проверил формат ошибки в контрактах → пропустил нарушение NFR-13
```

### 3.7 arch-ver — верификатор Architect

Skill: `ARCH_VER_skill.md`. Артефакт: `docs/architecture.md`, `stack_and_tools.md`, ADR.

Чек-лист:

```
[ ] Архитектура покрывает все NFR из system_specification
[ ] SOLID и Clean Architecture в module structure (handler/service/repository)
[ ] KISS: нет неоправданной сложности (лишние слои/абстракции/паттерны)
[ ] Реализуемо выбранным стеком (stack_and_tools) — Go 1.25/Fiber v3/React 19/Ent
[ ] Security на уровне архитектуры: WS RBAC, search scope, webhook signature, SSRF-защита,
    security headers
[ ] Нет запрещённых замен (Fiber, NATS JetStream, SeaweedFS, TanStack Router — фиксированы)
[ ] Соответствие TZ: mandatory соблюдены, forbidden отсутствуют; исключения только через ADR
[ ] ADR корректны: контекст, ≥2 варианта, решение, последствия, проверка соответствия
[ ] C4/числа синхронны с кодом (grep/go.mod/routeTree) — нет stale-claims
[ ] Tech debt зафиксирован (TD-XX) с evidence file:line
[ ] Описание ARCH соответствует codebase
```

Метод (из реального лога arch-ver): пофактовый spot-check каждого утверждения против кода.

```
| Claim | Status | Evidence |
| 20 domain modules + platform | ✅ | 21 dirs в backend/internal/ |
| Go 1.25, Fiber v3, ent v0.14.1 | ✅ | go.mod |
| TD-06 security headers not wired | ❌ | middleware.SecurityHeaders() на main.go:190 (уже wired) |
```

#### Good / Bad: arch-ver

```
GOOD
- Сверил «18 модулей» с кодом → фактически 21 → [MEDIUM] stale, обновить C4
- Проверил ADR-042: есть 2 варианта, последствия, проверка соответствия → PASS
- Подтвердил отсутствие net/http для WS (только Fiber) → TZ соблюдён
```

```
BAD
- Принял диаграмму C4 без сверки с кодом → пропустил drift числа модулей
- Пропустил ADR без раздела «последствия» → неполное решение
```

### 3.8 product-ver — верификатор соответствия продукту

Артефакт: любой пользовательский результат (фича) против vision/KPI/scope.

Чек-лист:

```
[ ] Реализованное соответствует US и AC из vision_and_scope (не «что-то похожее»)
[ ] Вклад в бизнес-цель G-XX прослеживается; KPI движется в нужную сторону
[ ] Scope соблюдён: не реализовано Won't Have; не выходит за границы milestone
[ ] Ценность для персоны реальна (JTBD выполнен), а не «фича ради фичи»
[ ] Нет регресса пользовательских сценариев (основной/альтернативный/исключительный)
[ ] UX соответствует ожиданиям vision (онбординг, hotkeys, доступность модуля)
[ ] Приоритет реализации соответствует MoSCoW/ROADMAP
```

#### Good / Bad: product-ver

```
GOOD
- Прошёл US-S16 как Мария (QA Lead): пометил flaky, отфильтровал, увидел автоподсветку → AC выполнены
- Убедился, что не добавлен функционал из Won't Have
```

```
BAD
- Принял фичу, которая «технически работает», но не решает JTBD персоны
- Пропустил выход за scope (реализовано то, что PO явно отложил)
```

### 3.9 designer-ver — верификатор дизайна

Артефакт: UI против дизайн-системы и a11y.

Чек-лист:

```
[ ] Консистентность токенов: цвета/отступы/типографика/радиусы из дизайн-системы (Tailwind/Shadcn),
    не произвольные значения
[ ] Нет !important; только className-кастомизация Shadcn, Radix-поведение не изменено
[ ] a11y: контраст WCAG AA (4.5:1 / 3:1), focus-visible, aria, порядок табуляции
[ ] Отзывчивость: mobile/tablet/desktop брейкпоинты; bottom nav / sidebar
[ ] Состояния компонентов: hover/active/disabled/loading/empty/error покрыты
[ ] Единообразие с существующими экранами (не «новый стиль» в одном месте)
[ ] Иконки/спейсинг из системы; нет визуального рассинхрона
[ ] Тёмная/светлая тема и accent color работают
```

#### Good / Bad: designer-ver

```
GOOD
- Проверил, что цвета из токенов темы, а не хардкод hex → консистентно
- Проверил empty/error/loading состояния списка flaky-кейсов → присутствуют
```

```
BAD
- "Красиво" без проверки контраста → пропустил AA-нарушение
- Пропустил произвольный отступ вне шкалы токенов → рассинхрон
```

### 3.10 code-reviewer — ревью diff перед мержем

GSD: `gsd-code-review`, `gsd-code-review-fix`. Артефакт: diff PR.

Чек-лист:

```
[ ] Корректность: логика делает то, что заявлено; нет очевидных багов/edge-case-дыр
[ ] Безопасность: SSRF, XSS, инъекции, секреты, RBAC, подпись webhook, timing-safe сравнение
[ ] Соответствие паттернам: layering, Repository, error wrapping, context-first
[ ] Стиль: code_style_guide (Go/TS), нет запрещённых паттернов, нет narrating-комментариев
[ ] Тесты: изменения покрыты; нет удаления тестов без причины; race зелёный
[ ] Атомарность и читаемость diff; нет несвязанных изменений
[ ] Нет DEV_MODE=true, hardcoded credentials, seaweedfs:latest
[ ] Docs sync выполнен (STATE + domain docs) для user-visible/prod изменений
[ ] Соответствие TZ; конвенции коммитов (feat/fix/refactor(scope))
[ ] PR ссылается на задачу/план; CI зелёный
```

code-reviewer фиксирует замечания в REVIEW.md (раздел 5), не мержит с CRITICAL/HIGH.

#### Good / Bad: code-reviewer

```
GOOD
- Нашёл io.ReadAll без LimitReader → [HIGH] memory DoS, предложил io.LimitReader
- Проверил, что diff атомарен и покрыт тестами, CI зелёный
```

```
BAD
- Апрув «LGTM» без чтения diff по существу
- Пропустил hardcoded token в конфиге
```

### 3.11 Общие проверки для всех верификаторов

Независимо от роли, каждый верификатор проверяет:

```
[ ] Описание исполнителя (2–5 предложений) соответствует фактическому содержимому артефакта
[ ] Docs sync выполнен: STATE.md + доменный doc обновлены (rule 55) для user-visible/prod
[ ] Соответствие TZ (mandatory/forbidden) в зоне артефакта
[ ] Трассируемость вверх: артефакт ссылается на порождающее требование (US/FR/AC)
[ ] Нет артефактов-мусора в git status (cover*/*.out/tmp_*) — rule 45
[ ] Атомарность изменения (не смешаны несвязанные правки)
[ ] Вердикт с датой/временем и severity-классификацией
```

Это «нижний слой» проверки — он не заменяет ролевой чек-лист (3.1–3.10), а дополняет его.

### 3.12 Разрешение спора «код vs требование»

Когда верификатор видит расхождение между кодом и требованием, он не решает сам, кто прав —
он применяет протокол test-triage к артефактам:

```
1. Прочитать первоисточник требования (vision/spec), не только код.
2. Определить: код отклонился от требования ИЛИ требование устарело/неверно.
3. Если код неверен → замечание исполнителю кода (be/fe).
4. Если требование неверно/неполно → замечание аналитику (ba/sa) + возможно возврат на анализ.
5. Зафиксировать: ROOT=code|spec, обоснование.
```

Верификатор не «подгоняет» вердикт под удобную сторону — он идёт к первоисточнику.

---
