---
name: portable-test-automation-ai-quality-layer
version: 0.1.0-alpha.1
status: pre-release
release_date: TODO
owner: automation-quality
validation_status: STRUCTURAL_PASS
last_validation: 2026-07-26
release_eligibility: blocked
provenance: MANIFEST.md
---

# Версия

Текущая semantic version: `0.1.0-alpha.1`.

Это reviewable pre-release placeholder для portable AI quality layer. Это не
заявление о completion, production-readiness, conformance, effectiveness или ROI.

Структурная валидация полного текущего тулкита проходит с 0 errors и
0 warnings. Combined security/tooling tests: 55 PASS / 1 environmental SKIP,
56 total (создание symlink недоступно; mocked Windows reparse detection
проходит). Portable package из 172 files, включая `build_task_packet.py`, был
собран и независимо провалидирован; local Hub profile/smoke/raw feedback
evidence исключены. Результаты eval effectiveness остаются `NOT_RUN`, поэтому
`release_eligibility` — `blocked`.

Продвижение требует:

1. проходящих validator unit tests;
2. зелёной структурной валидации на точном intended distribution;
3. approved source и governance review;
4. representative paired eval evidence для intended usage;
5. approval accountable owner.

Пока eval/governance/release gates не записаны, `validation_status` остаётся
`STRUCTURAL_PASS`, а `release_eligibility` — `blocked`.

См. [MANIFEST.md](MANIFEST.md) для ownership и границ артефактов.
