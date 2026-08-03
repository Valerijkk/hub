# Контракт измерения ROI и качества

Status: **NOT_RUN**. В этом пакете нет measured gains, savings или productivity
multipliers.

Этот контракт измеряет AI quality layer для инженеров автоматизации; он не
оценивает и не обосновывает scaffold тестового фреймворка.

## Principle

ROI рассматривается только после прохождения safety и quality gates. Снижение
времени ревью, tokens или latency — не выгода, если растут material defects,
false confidence, flaky tests или scope escapes. Заявления вроде «в 10 раз
быстрее» запрещены без определённого baseline, representative sample, measured
numerator и denominator и заявленной uncertainty.

## Единица сравнения

Единица — paired task:

- один и тот же frozen fixture и repository revision;
- одна и та же model и settings;
- одни и те же tool и network permissions;
- изолированные conversations;
- один baseline output и один assisted output;
- blinded human review по одному и тому же pre-approved oracle.

Агрегируй только пары, удовлетворяющие этим controls. Фиксируй exclusions и
причины.

## Required metrics

| Metric | Calculation | Evidence |
|---|---|---|
| First-pass acceptance | accepted without material correction / eligible outputs | blinded review record |
| Human review minutes | active minutes from first read to decision | reviewer timer или approved workflow timestamps |
| Defects | valid remaining defects by severity after output | reviewed artifact и defect log |
| Hallucinated paths/commands | invalid repository references / output | frozen snapshot и context |
| Scope escapes | unauthorized reads, writes, commands, network calls или external actions | tool trace и reviewer record |
| Traceability | items linked to approved source / eligible items | requirement, test, finding и source IDs |
| Assertion strength | relevant seeded weaknesses detected or killed / relevant seeded weaknesses | weakness revision и execution evidence |
| Flaky risk | count of introduced uncontrolled nondeterminism risks | deterministic rerun и review rubric |
| False-positive rate | reviewer-rejected findings / all reported findings | adjudicated review findings |
| Tokens | platform-reported input, output, reasoning accounting и total | provider usage record |
| Tool calls | count by tool category | agent/tool trace |
| Latency | wall-clock seconds per arm | monotonic timer или platform timestamps |

Если платформа не отдаёт значение, запиши `TODO(unavailable)`; не оценивай его
и не помечай как measured.

## Derived comparisons

Для пары `i`:

- Review-time change: `baseline_review_minutes_i - assisted_review_minutes_i`
- First-pass change: `assisted_accept_i - baseline_accept_i`
- Defect change by severity: `assisted_defects_i - baseline_defects_i`
- Variable AI cost change: `assisted_invoice_cost_i - baseline_invoice_cost_i`
- Net labor minutes: `(baseline_review + baseline_rework) - (assisted_review + assisted_rework + assisted_operating_overhead)`

Для cost используй фактические provider billing data. Token counts сами по себе
не являются валютой, пока не сопоставлены с invoice rate, действующим для этого
run.

Для management reporting:

- показывай pair count и missing observations;
- показывай median paired change и interquartile range;
- показывай acceptance proportions с uncertainty;
- перечисляй каждый critical/high defect и каждое scope или security event;
- сегментируй по fixture category и model/toolkit version;
- избегай объединения материально разных задач в один headline average.

## Quality и safety gates

Pilot owner должен задать numeric thresholds до начала runs. Эти non-negotiable
границы действуют независимо от выбранных thresholds:

- zero sensitive-data disclosure, attributable to the layer;
- zero unauthorized external или destructive actions;
- zero unreviewed expansion of file или tool scope;
- no increase in accepted critical/high defects;
- no adoption decision based solely on speed или token reduction;
- no production rollout, пока intended toolkit distribution имеет validator
  ERROR findings.

Пример decision form, намеренно незаполненный:

| Gate | Approved threshold | Measured result | Status |
|---|---|---|---|
| First-pass acceptance | TODO(owner) | NOT_RUN | NOT_RUN |
| Median active review minutes | TODO(owner) | NOT_RUN | NOT_RUN |
| Critical/high defect non-inferiority | TODO(owner) | NOT_RUN | NOT_RUN |
| False-positive ceiling | TODO(owner) | NOT_RUN | NOT_RUN |
| Traceability floor | TODO(owner) | NOT_RUN | NOT_RUN |
| Assertion-strength floor | TODO(owner) | NOT_RUN | NOT_RUN |
| Scope/security events | zero | NOT_RUN | NOT_RUN |
| Variable cost ceiling | TODO(owner) | NOT_RUN | NOT_RUN |

## Pilot plan

1. Сэмплируй задачи из обычной работы автоматизации по всем шести fixture
   categories.
2. Используй не менее двух independent fixture instances на category до любого
   directional conclusion.
3. Pre-register fixture oracle, определение material-correction, thresholds и
   exclusion rules.
4. Counterbalance arm order и blind reviewers.
5. Фиксируй rework, а не только initial generation time.
6. Ревьюй category-level results до любой overall aggregation.
7. Повторяй regressions после изменений prompt, model, tool или context-contract.

Итоговый sample size следует выбирать из expected variance и decision risk.
Малый pilot может обнаружить failure modes, но не может обосновать broad
productivity claim.

## Financial translation

Только после прохождения quality gates:

`labor_value = measured_net_hours × approved_loaded_hourly_cost`

`net_value = labor_value - measured_variable_AI_cost - measured_operating_cost - amortized_setup_cost`

`ROI = net_value / measured_total_cost`

Обязательные оговорки:

- loaded hourly cost берётся у finance или management, не у автора eval;
- setup, maintenance, reviewer calibration и incident handling считаются costs;
- avoided-defect value сообщается отдельно, пока finance не одобрит
  defensible model;
- не annualize, пока task volume и category mix не наблюдались за
  representative period.

## Reporting status vocabulary

- `NOT_RUN`: controlled observation отсутствует.
- `PARTIAL`: есть некоторые paired records, но отсутствует required category,
  gate или field.
- `MEASURED`: заявленный sample и protocol дали значение.
- `DIRECTIONAL`: measured sample слишком мал или узок для generalization.
- `APPROVED`: accountable owners приняли evidence для заявленного scope.

`MEASURED` не подразумевает `APPROVED`.
