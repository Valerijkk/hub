# Реестр первичных источников

Дата доступа для каждого источника ниже: **2026-07-26**.

Этот реестр содержит официальные спецификации и документацию vendor/project.
Тулкит фиксирует actionable control, выведенный из каждого источника; он не
копирует source prose. Запись источника не доказывает, что control реализован —
это доказывают validator, eval record или governance evidence.

## Cursor

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-CURSOR-RULES-001 | [Cursor Rules](https://cursor.com/docs/rules) | Держать persistent project instructions versioned и scoped; валидировать frontmatter, где платформа этого требует; тестировать path-specific behavior, а не предполагать, что rule загрузился. |
| SRC-CURSOR-SKILLS-001 | [Cursor Agent Skills](https://cursor.com/docs/skills) | Упаковывать repeatable procedures как discoverable skills, держать selection metadata краткой и проверять skill directory contract до distribution. |
| SRC-CURSOR-SUBAGENTS-001 | [Cursor Subagents](https://cursor.com/docs/subagents) | Делегировать только задачи, которым полезен isolated context; в assisted eval runs фиксировать model, readonly status, task boundary и returned evidence. |
| SRC-CURSOR-HOOKS-001 | [Cursor Hooks](https://cursor.com/docs/hooks) | Трактовать hooks как executable policy points: ревьюить scripts, fail closed только для approved gates, redact event data и тестировать allow/block behavior до team rollout. |

Controls fed: frontmatter checks, Agent Skills checks, scope-control fixture,
tool-call accounting, approval boundaries и hook rollout guidance.

## Agent Skills open specification

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-AGENTSKILLS-SPEC-001 | [Agent Skills specification](https://agentskills.io/specification) | Требовать `SKILL.md` frontmatter, обязательные `name` и `description`, lowercase hyphenated names и `name`, равный parent directory. Scripts и references оставлять optional bundled resources. |
| SRC-AGENTSKILLS-REPO-001 | [Agent Skills specification repository](https://github.com/agentskills/agentskills) | Отслеживать изменения спецификации и сравнивать этот stdlib validator с reference implementation до заявления о полной conformance. |

Controls fed: `SKILL001`–`SKILL005`. Локальный validator проверяет portable
subset, используемый этим тулкитом; он не заявляет complete specification
conformance.

## GSD inspiration and license

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-GSD-REPO-001 | [gsd-build/get-shit-done](https://github.com/gsd-build/get-shit-done) | Трактовать GSD как inspiration/upstream, а не как неявную часть Hub; pin revision до копирования любого substantial upstream material. |
| SRC-GSD-LICENSE-001 | [GSD MIT License](https://github.com/gsd-build/get-shit-done/blob/main/LICENSE) | Сохранять MIT copyright/license для скопированных substantial portions. Текущая local inventory revision — `UNKNOWN`, поэтому external redistribution остаётся gated. |

| SRC-GSD-CONTEXT-001 | [CONTEXT.md](https://github.com/gsd-build/get-shit-done/blob/main/CONTEXT.md) | Запрет проверять исходник грепом вместо вызова; обязательные граничные фикстуры для параметров-лимитов; тест паритета при намеренном дублировании; таймауты на внешние вызовы с деградацией вместо исключения. |
| SRC-GSD-SUITES-001 | [docs/TESTING-SUITES.md](https://github.com/gsd-build/get-shit-done/blob/main/docs/TESTING-SUITES.md) | Разделение сьютов маркером в имени файла, а не каталогом; разный состав прогона на мержреквесте и на главной ветке; структурные утверждения вместо сравнения текста ошибки. |
| SRC-GSD-STANDARDS-001 | [docs/contributor-standards.md](https://github.com/gsd-build/get-shit-done/blob/main/docs/contributor-standards.md) | Поправка к принятому решению дописывается отдельной секцией; исходный текст не переписывается. |

| SRC-GSD-CORE-001 | [open-gsd/gsd-core — TESTING-STANDARDS.md](https://github.com/open-gsd/gsd-core/blob/main/TESTING-STANDARDS.md) | Шесть контрактов теста; политика удаления плохих тестов вместо отключения; свойства вместо примеров для парсинга и обратимых пар; часы как параметр вместо подмены системного времени; классы непереносимости между ОС. |

Дата доступа к пяти записям GSD: **2026-08-04**. Репозиторий
`gsd-build/get-shit-done` на эту дату архивирован, разработка продолжена в
`open-gsd/gsd-core`. Изучены обе: архивная `main` первого и `main` второго.

Вынесены **правила**, сформулированные заново под объект этого пакета.
Существенные фрагменты текста не копировались, поэтому license gate по
`SRC-GSD-LICENSE-001` не срабатывает; при копировании кода или прозы он
применяется в полном объёме.

Controls fed: [`NOTICE.md`](../NOTICE.md), provenance/reuse review, package
license gate. Этот тулкит заявляет adapted patterns, а не official GSD status.

## OpenAI GPT-5.6, reasoning, and Codex

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-OPENAI-GPT56-001 | [Prompting guidance for GPT-5.6 Sol](https://developers.openai.com/api/docs/guides/prompt-guidance-gpt-5p6) | Задавать outcome, constraints, available evidence, approval boundaries, completion criteria и stopping conditions. Tuning reasoning effort — только после representative eval, показавшего quality need. |
| SRC-OPENAI-REASONING-001 | [Reasoning models](https://developers.openai.com/api/docs/guides/reasoning) | Фиксировать model и reasoning settings в каждом run; сравнивать settings на paired fixtures; запрашивать concise rationale и evidence, а не raw private reasoning. |
| SRC-OPENAI-CODEX-001 | [Codex prompting guide](https://developers.openai.com/cookbook/examples/gpt-5/codex_prompting_guide) | Называть desired behavior, релевантный code или reproduction, constraints и verification command. Измерять tool use и artifact correctness, а не длину prompt. |
| SRC-OPENAI-REASONING-ITEMS-001 | [Reasoning items in the Responses API](https://developers.openai.com/cookbook/examples/responses_api/reasoning_items) | Не требовать raw reasoning disclosure. Если платформа даёт reasoning summaries или accounting, хранить только минимум approved metadata и соблюдать retention controls. |

Controls fed: forbidden explicit reasoning-disclosure scan, paired model-setting
records, first-pass acceptance, token/tool/latency metrics и evidence-first
output contracts.

## GitHub instructions and Actions security

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-GITHUB-INSTRUCTIONS-001 | [Adding repository custom instructions](https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/add-custom-instructions/add-repository-instructions) | Держать repository-wide и path-specific instructions раздельно, версионировать их вместе с кодом и включать build/test/validation facts из authoritative project sources. |
| SRC-GITHUB-ACTIONS-SECURE-001 | [GitHub Actions secure use reference](https://docs.github.com/en/actions/reference/security/secure-use) | Применять least privilege к workflow tokens, трактовать event и context fields как untrusted, pin third-party actions по policy, предотвращать secret output и ревьюить self-hosted runner exposure. |
| SRC-GITHUB-OIDC-001 | [OpenID Connect for GitHub Actions](https://docs.github.com/en/actions/concepts/security/openid-connect) | Предпочитать short-lived federated credentials long-lived cloud secrets, когда deployment automation нужна external authentication. |

Controls fed: secret redaction, prompt-injection fixture, scope и
external-action approvals, CI policy review и repository-instruction provenance.

## OWASP

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-OWASP-LLM10-2025-001 | [OWASP Top 10 for LLM Applications 2025](https://owasp.org/www-project-top-10-for-large-language-model-applications/assets/PDF/OWASP-Top-10-for-LLMs-v2025.pdf) | Тестировать prompt injection, sensitive information disclosure, unsafe output use, excessive agency, supply-chain assumptions и unbounded resource use в pilot threat set. |
| SRC-OWASP-LLMSVS-001 | [OWASP LLMSVS v2.0](https://owasp.org/www-project-llm-verification-standard/LLMSVS-v2.0-en.html) | Определять verification scope и exclusions, сохранять reproducible evidence и сочетать automated checks с human review для context-sensitive controls. |
| SRC-OWASP-CODE-REVIEW-001 | [OWASP Secure Code Review Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secure_Code_Review_Cheat_Sheet.html) | Использовать risk-based, data-flow-aware manual review для authentication, authorization, input handling, secrets и business logic; не считать scanner output финальным вердиктом. |

Controls fed: injection и review fixtures, secret-like scanning, defect-first
review, false-positive measurement, evidence retention и security approval
boundaries.

## Model Context Protocol

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-MCP-AUTH-001 | [MCP authorization specification](https://modelcontextprotocol.io/specification/latest/basic/authorization) | Привязывать tokens к intended resource, использовать least-privilege scopes, валидировать authorization metadata и никогда не передавать upstream tokens как замену server authorization. |
| SRC-MCP-SECURITY-001 | [MCP security best practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices) | Валидировать redirect и authorization URLs, защищать consent и state, предотвращать confused-deputy flows, минимизировать scopes и логировать metadata только после credential redaction. |

Controls fed: MCP tool approval boundaries, network policy fields, tool-call
audit records, secret redaction и injection resistance.

## NIST

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-NIST-AI600-1-001 | [NIST AI 600-1: Generative AI Profile](https://www.nist.gov/publications/artificial-intelligence-risk-management-framework-generative-artificial-intelligence) | Назначать accountable owners, документировать context и limitations, измерять risks confabulation и misuse, мониторить incidents и принимать deployment decisions по recorded evidence и risk tolerance. |

Controls fed: governance roles, provenance, hallucinated-path metrics, pilot
decision gates, incident response и explicit limitation statements.

## Engineering and observability

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-GO-TESTING-001 | [Go `testing` package](https://pkg.go.dev/testing) | Опирать generated Go tests на конвенции `_test.go` и `TestXxx`, использовать subtests/fuzzing только когда репозиторий их поддерживает, и выполнять canonical `go test` command репозитория. |
| SRC-GO-FUZZ-001 | [Go fuzzing](https://go.dev/doc/security/fuzz) | Использовать fuzzing для exploration пространства входов с deterministic seed corpus и reproducible failures; не подменять fuzzing requirement-based examples. |
| SRC-REACT-ACT-001 | [React `act`](https://react.dev/reference/react/act) | Дождаться settle state updates перед assertions; предпочитать user-observable outcomes и approved testing helpers проекта. |
| SRC-TYPESCRIPT-STRICT-001 | [TypeScript strictness](https://www.typescriptlang.org/docs/handbook/2/basic-types.html#strictness) | Сохранять project strictness, не ослаблять type checks ради компиляции generated tests и трактовать unchecked `any` или null paths как review risks. |
| SRC-TYPESCRIPT-NARROWING-001 | [TypeScript narrowing](https://www.typescriptlang.org/docs/handbook/2/narrowing.html) | Использовать control-flow narrowing и exhaustive checks для typed fixtures и result parsers вместо unsafe assertions. |
| SRC-OTEL-LIBRARIES-001 | [OpenTelemetry instrumentation libraries](https://opentelemetry.io/docs/concepts/instrumentation/libraries/) | Инжектить test exporters или processors, assert spans без production export и включать instrumentation name/version для provenance. |
| SRC-OTEL-TRACES-001 | [OpenTelemetry traces](https://opentelemetry.io/docs/concepts/signals/traces/) | Коррелировать eval или automation operations с trace/span context при approval; никогда не прикреплять prompts, source code или credentials как unrestricted attributes. |

Controls fed: критерии генерации и оценки риска нестабильности, compile/test evidence, assertion
sensitivity, strict typed records, deterministic telemetry tests и
privacy-aware observability.

## Test-framework profiles

Эти источники применяются только когда generated profile подтверждает
соответствующий framework/version. Они не являются универсальными правилами.

| ID | Official source | Actionable mapping |
|---|---|---|
| SRC-PLAYWRIGHT-BP-001 | [Playwright best practices](https://playwright.dev/docs/best-practices) | Сохранять test isolation, предпочитать resilient user-facing locators и web-first assertions, избегать fixed waits, когда проект использует Playwright. |
| SRC-TESTING-LIBRARY-001 | [Testing Library](https://testing-library.com/docs/) | Предпочитать user-observable behavior внутренностям компонента; использовать project-approved accessible query order. |
| SRC-PYTEST-FIXTURES-001 | [pytest fixtures](https://docs.pytest.org/en/stable/how-to/fixtures.html) | Держать изменения fixture state явными, а cleanup — безопасным; использовать fixture scope/dependencies, подтверждённые project profile. |
| SRC-JUNIT-GUIDE-001 | [JUnit User Guide](https://docs.junit.org/current/user-guide/) | Опирать lifecycle, parameterization и extension choices на detected JUnit version; не переносить version-specific annotations вслепую. |
| SRC-CYPRESS-BP-001 | [Cypress best practices](https://docs.cypress.io/app/core-concepts/best-practices) | Сохранять независимость тестов и project-approved resilient selectors; трактовать disabled isolation как явный risk. |
| SRC-VITEST-MOCKING-001 | [Vitest mocking](https://vitest.dev/guide/mocking) | Restore/reset mock state, учитывать ограничения hoisting/environment и следовать detected Browser/Node mode. |

Controls fed: framework-specific profile cards, UI selector/isolation review,
fixture/mock cleanup и version-aware generation.

## Registry maintenance

1. Перепроверяй ссылки и material guidance перед каждым toolkit release.
2. Меняй access-date только после того, как человек или approved retrieval job
   снова открыл источник.
3. Если источник материально меняет control, обновляй validation manifest или
   критерий оценки в том же reviewed change.
4. Держи vendor-specific controls за capability checks; не предполагай, что
   каждый host реализует каждое поле.
5. Помечай superseded sources вместо silent удаления их decision history.
