# August 2026 — phase dependency map (TASK-005)


> ⚠️ **Зависимости на Э3–Э7 неактуальны.** Решением владельца 2026-07-30
> эти этапы отложены до появления нагрузки
> ([`04-REVISED-PRIORITIES.md`](04-REVISED-PRIORITIES.md)). Строки
> «фаза 133 → Э3» и производные от них не применяются.
>
> Актуальны только зависимости внутри трека тестового фреймворка.

Machine-readable companion to [`00-AUGUST-PLAN.md`](00-AUGUST-PLAN.md).

## Edges

| From | To | Strength | Note |
|------|-----|----------|------|
| — | 126 | none | FR-INTR-08; schema ready |
| 126 | 127 | hard | shared trash registration |
| A3 coverage baseline | 128 | hard | fresh COV before TF migrate |
| 128 | 129 | hard | |
| 129 | 130 | hard | |
| 129 | 131 | hard | |
| 130 | 131 | soft | smoke before rewrite |
| 128 | 132 | hard | |
| A3 coverage | 132 | hard | before claiming no cov drop |
| 128 | 133 | hard | |
| 128 | 134 | hard | |
| 128 | 135 | hard | |
| 132 | 134 | soft | |
| 134 | 135 | soft | |
| 130+131+132+133 | 136 | hard | cutover |
| **133** | **MS Э3** | **hard block** | no Module Federation without FE units |

## Microservices track

Architect §8 questions in [`01-MICROSERVICES-SPLIT.md`](01-MICROSERVICES-SPLIT.md) remain **open** — do not start Э1 until answered.
Э3 must not start until phase **133** is verified.
