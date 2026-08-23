---
name: test-triage
description: При FAIL сначала установить требуемое и наблюдаемое product behavior, затем классифицировать code/test/infra/flaky/stop. Обязателен для QA, BE, FE, debugger и verifier.
---

# Test Triage — behavior before fix

Падающий test — свидетельство расхождения, а не автоматический verdict о code.
Зелёный test также может быть ложным, если oracle не чувствителен.

## Scope

Применять к target jobs:

- `backend-unit` → `backend/tests/**`;
- `frontend-unit` → `frontend/tests/**`;
- `python-contract|api|integration|e2e|load` → root `testframework/**`.

Legacy Go component/static Vitest/TS Playwright не чинить и не портировать;
они удаляются по founder manifest. Если FAIL пришёл только из legacy runner,
зафиксировать исторический контекст и проверить, относится ли behavior к
current target requirement/risk.

## 1. Зафиксировать симптом и provenance

```text
Job/case: <target job + exact case/TC>
Source SHA/image: <values>
Exact Docker command/resources/services: <values>
Observed: <error/output/state/side effect>
Collection/skips/duration/cleanup: <values>
```

Host-run, старый лог или неизвестный source SHA не воспроизводят baseline.

## 2. Проверить product behavior первым

До изменения code/test:

1. Найти task, requirement, approved spec/API/ADR и связанный TC.
2. Сформулировать expected observable outcome и forbidden side effects.
3. Трассировать production boundary; при необходимости сделать минимальный
   repro только в task-owned Docker/Compose с CPU/RAM limits.
4. Сравнить docs, фактический product behavior и test oracle.
5. Для security/tenant риска проверить deny path, не только happy path.

Нельзя считать test assertion источником истины. Если docs и approved
decisions расходятся — STOP, не выбирать удобную версию.

## 3. Выбрать ровно один ROOT

| ROOT | Диагноз | Действие |
|---|---|---|
| `code` | production behavior нарушает однозначный contract | минимальный code fix; test сохраняет/усиливает oracle |
| `test` | product соответствует contract, а fixture/oracle/metadata неверны | минимальный test fix; production не менять |
| `infra` | image/config/service/data/port/resource/cleanup не соответствует runner contract | исправить task-owned infra, не product assertion |
| `flaky` | недетерминизм time/order/shared state/network/race подтверждён | устранить cause; retry/quarantine только bounded с owner+expiry |
| `stop` | contract неоднозначен, authority/secret/destructive risk или нужный scope вне ownership | остановиться, передать BA/SA/product/owner |

`code + test` как один root запрещён. Если approved contract изменился, сначала
owner обновляет normative docs/decision; затем это отдельные согласованные
изменения с общей traceability.

## 4. False-confidence checks

- case красный при временном удалении защищаемого behavior?
- oracle наблюдает output/state/effect/no-side-effect, а не private call order?
- mock interaction является public contract?
- zero collection/skip/retry не замаскировали FAIL?
- suite соответствует boundary, а не историческому path?
- historical 82% coverage/parity не использованы как rationale?

## 5. Запрещено

- менять expected/assertion только ради green;
- одновременно править code и test «наугад»;
- удалять case или добавлять skip без risk/ticket/owner/expiry;
- увеличивать timeout/retry без доказанного `ROOT=flaky|infra`;
- восстанавливать legacy non-unit ради coverage;
- запускать host tests/validators/build/app process;
- выдавать static review или NOT_RUN за verified PASS.

## 6. Verification

Исправление проверяется только через
`docker compose -f docker-compose.local.yml run --rm <exact-service>`, затем смежными fixed services. Evidence содержит
source SHA/image/command/resources/result/duration/artifacts/cleanup. Если
воспроизведение не разрешено task scope, записать `NOT_RUN`.

## Log

```text
TRIAGE: <job/case/TC>
CONTRACT: <task + requirement + expected observable behavior>
OBSERVED: <product vs oracle>
ROOT: code|test|infra|flaky|stop
EVIDENCE: <Docker provenance or NOT_RUN>
FIX: <exact files/owner or pending clarification>
VERIFIED: <target jobs + results + cleanup, or NOT_RUN>
RESIDUAL: <unverified risk>
```

## Hub calibration examples

- Magic Link test ожидает plaintext Redis key, а approved security contract
  требует hash → `ROOT=test`.
- Cross-tenant object доступ фактически разрешён, requirement требует deny →
  `ROOT=code`.
- Integration suite не видит NATS из-за неверного task Compose network →
  `ROOT=infra`.
- E2E иногда падает из-за shared account/order → `ROOT=flaky`.
- Overlap policy отсутствует и в spec, и в ADR → `ROOT=stop` к SA/product.

Сложная code/flaky investigation может перейти в `gsd-debug`; root classification
и evidence сохраняются.
