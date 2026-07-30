# 13 — Corporate delivery fragments

> Фрагменты для корпоративной поставки: планирование, волны, ревью, strangler микросервисов, урезанный GitHub PR-flow, handoff, верификация, portable-kit, AI-team ops. Обезличены — подставь слоты из `context/projects/<profile>/`.

Слоты: `{{goal}}`, `{{module}}`, `{{files}}`, `{{role}}`, `{{ac}}`, `{{service}}`.

> **Recipe assembly.** Фрагменты ниже — слоты `[ЧТО]` [мастер-шаблона](../../prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md). Профиль категории: **[РОЛЬ]** orchestrator / tech-lead / {{role}} · **[КОНТЕКСТ]** `context/projects/<profile>/` + STATE + шаблоны `agents/templates/` · **[ФОРМАТ]** цель · границы · артефакт · Verification · **rationale** goal-backward от критерия приёмки · **assumptions** project-agnostic (без хардкода путей чужого репо) · **evidence** DoD проверяем командой/артефактом · **[УСЛОВИЯ]** верификация `rules/40`.

## Planning (1–15)

1. Сформулируй SMART-goal для «{{goal}}» на одну фазу: Specific, Measurable, Achievable, Relevant, Time-bound — по одному предложению на букву.
2. Разбей «{{goal}}» на атомарные планы: каждый план ≤1 артефакт или ≤1 логический коммит; таблица plan_id | артефакт | DoD.
3. Построй MoSCoW для «{{goal}}»: Must / Should / Could / Won't — Won't обязателен и явным списком.
4. Напиши success criteria для «{{goal}}»: каждый критерий проверяем командой, сценарием или артефактом (не «готово / выглядит ок»).
5. Найди depends между задачами модуля {{module}}: рёбра A→B с причиной (контракт / схема / UI / infra).
6. Оцени риски для «{{goal}}»: вероятность × impact; для каждого — mitigation и owner-роль {{role}}.
7. Составь wave map для «{{goal}}»: что параллельно, что строго последовательно; конфликтные пути файлов.
8. Подготовь kickoff для orchestrator по шаблону phase-kickoff: goal, files, AC, regression_scope, verifier.
9. Перечисли out-of-scope для «{{goal}}», чтобы не расползтись по смежным модулям.
10. Сверь goal «{{goal}}» с KPI/метриками продукта из context: что двигаем, что не трогаем.
11. Напиши 3–5 acceptance criteria в Given/When/Then для «{{goal}}»; свяжи с {{ac}} если задан.
12. Оцени effort по атомарным планам «{{goal}}»: S/M/L и критический путь.
13. Составь definition of ready для старта волны по «{{goal}}»: спека, границы, AC, owner.
14. Выдели gray-zone решения для «{{goal}}»: вопрос | варианты | рекомендованный default | кто решает.
15. Составь RACI на фазу «{{goal}}»: Responsible / Accountable / Consulted / Informed по ролям.

## Waves & delegation (16–28)

16. Заполни wave-delegate для {{role}} на файлах {{files}}: цель, границы, формат ответа, regression_scope, verifier.
17. Проверь пересечение файлов в волне вокруг {{module}}: перечисли конфликты и предложи сериализацию.
18. Сформулируй regression_scope для {{module}} после изменений: команды/пакеты/сценарии до green.
19. Собери статус волны по «{{goal}}» в 5 буллетов для STATE: сделано / блокеры / next.
20. Верни агента: в задаче не хватает границ файлов — что именно запросить у оркестратора (чеклист).
21. Разбей перегруженный план «{{goal}}» на 3 атомарных плана с явными depends.
22. Определи verifier (роль) для артефакта модуля {{module}} и критерии verified:true.
23. Составь corporate-wave на goal «{{goal}}»: роли, потоки A/B, точки синхронизации, handoff.
24. Назначь параллельные потоки: кто берёт {{module}} vs соседний сервис {{service}} без пересечения {{files}}.
25. Сформулируй stop-условия волны: когда остановить параллель и эскалировать orchestrator.
26. Напиши delegation-brief для {{role}}: входные артефакты, запрещённые файлы вне {{files}}, DoD.
27. Спланируй интеграционный слот волны: кто мержит ветки модулей, кто гоняет полный gate.
28. Построй очередь задач волны: priority | owner {{role}} | files | blocker_if.

## Review & quality (29–42)

29. Defect-first review: найди CRITICAL/HIGH в описании изменений {{files}}; без CRITICAL не хвали стиль.
30. Проверь, что AC «{{ac}}» покрыт тестами или явным планом тестов; gaps таблицей.
31. Найди security risks в {{module}}: authz, injections, SSRF, secrets, unsafe URL — файл:симптом.
32. Проверь secrets/credentials в diff {{files}}: паттерны поиска и red-flag имена файлов.
33. Anti-AI check по {{files}}: narrative comments, fake tests, invented expectations, boilerplate.
34. Трассировка требование → задача → тест для {{ac}}: матрица с дырами.
35. Оформи вердикт PASS / FAIL / PASS_WITH_GAPS по ревью {{files}} с списком gaps.
36. Чек-лист rules/40 — пройди этапы 1–10 кратко для сдачи по «{{goal}}».
37. Cross-review: что обязан проверить {{role}} в чужом PR по {{module}} (чеклист 8 пунктов).
38. Найди нарушения layering/границ модуля в {{files}}: зависимости «сквозь» слои.
39. Оцени тестовый долг после изменений {{module}}: чего не хватает до безопасного merge.
40. Проверь согласованность API-контракта {{module}} с AC «{{ac}}» и docs-sync needs.
41. Составь review summary для человека-мержера: риск | что смотреть вручную | рекомендация.
42. Найди flaky-риски в новых тестах вокруг {{module}}: sleep, порядок, shared state, сеть.

## Microservices strangler (43–56)

43. Составь strangler-план выноса {{module}} в {{service}}: фазы extract → dual-run → route-flip → retire.
44. Перечисли события/subjects messaging для {{module}} при выносе в {{service}} (publish/consume).
45. Проверь shared-table assumptions для {{module}}: какие таблицы нельзя шарить с {{service}}.
46. Опиши route-flip и rollback для трафика {{module}} → {{service}} (флаг, процент, kill-switch).
47. CI path-filter для {{module}} и {{service}}: какие пути запускают какие джобы.
48. OTel/tracing: какие headers/context прокинуть через messaging между core и {{service}}.
49. Auth между сервисами: JWT/mTLS validate checklist для вызовов к {{service}}.
50. Go/No-Go метрики после пилота выноса {{module}}: latency, error rate, lag, rollback count.
51. Границы core vs extracted {{service}} для {{module}}: владение данными, API, UI.
52. Schema-per-service: миграционный план данных {{module}} → БД {{service}} без shared writes.
53. Outbox/inbox: нужен ли transactional outbox при выносе {{module}}; где точка истины.
54. Контракт ошибок между core и {{service}}: формат, request_id, идемпотентность.
55. Cutover runbook: шаг за шагом dual-write → verify → flip для {{service}}.
56. Анти-паттерны strangler для {{module}}: distributed monolith, chatty sync, shared DB — как избежать.

## Reduced GitHub PR-flow (57–68)

57. Опиши PR template для изменения {{files}}: summary, test plan, risk, rollback.
58. Чек-лист review: approve / request-changes / comment-only — критерии для {{role}}.
59. Какие CI checks должны блокировать merge PR по {{module}} (минимум и желательно).
60. Связь issue/goal «{{goal}}» с веткой и PR metadata (title, labels, blueprint полей).
61. Conventional commit message для изменений {{module}} в {{files}} (type/scope/why).
62. Merge policy: кто мержит protected branch; что агентам запрещено.
63. Rollback после плохого merge по {{module}}: revert / hotfix / feature-flag — шаги.
64. Агент code-review: defect-first промпт на diff {{files}} с форматом findings.
65. Branch naming для «{{goal}}» / {{module}}: предложи имя и rationale.
66. PR size gate: когда дробить PR по {{files}}; критерии «слишком большой».
67. Required reviewers matrix: security / docs / {{role}} — когда обязательны.
68. Squash vs rebase-merge: рекомендация для истории PR «{{goal}}» и почему.

## Handoff (69–80)

69. Напиши handoff от backend-dev к security-reviewer по {{module}}: diff, угрозы, что проверить.
70. Напиши handoff от frontend-dev к orchestrator: UI AC «{{ac}}», остатки, блокеры.
71. Напиши handoff от {{role}} к verifier: артефакты, evidence, regression_scope, open questions.
72. Составь STATE session note на 5–8 строк по завершению «{{goal}}».
73. Handoff между волнами: что обязан передать поток A потоку B по файлам {{files}}.
74. Docs-sync handoff: какие документы обновить после изменения {{module}} (матрица тип→doc).
75. Инцидентный handoff: симптомы в {{service}}, уже сделано, гипотезы, next owner.
76. Пауза mid-phase: checkpoint handoff для «{{goal}}» — где остановились, как resume.
77. Handoff к человеку-мержеру: merge-ready чеклист (CI, review, verified, docs).
78. Передача portable-kit потребителю: что скопировать, что заполнить в context, smoke-check.
79. Handoff QA: что ручное / что авто по {{ac}}; данные и окружение.
80. Закрытие фазы: финальный handoff-пакет (артефакты, метрики, known gaps) для «{{goal}}».

## Verification (81–92)

81. Goal-verify «{{goal}}»: есть ли evidence по каждому success criterion; FAIL если narrative-only.
82. Регрессия: что прогнать после изменения {{module}} (новое поведение + старый домен).
83. Проверь AC «{{ac}}» end-to-end: шаги сценария, ожидаемый результат, фактический evidence.
84. Independent verification: что должен перепроверить другой агент (не автор) по {{files}}.
85. Verification block: сформируй компактный блок Verification для сдачи задачи {{role}}.
86. Цикл качества: опиши loop implement→test→triage→fix для бага в {{module}} до green.
87. Definition of done gate: чеклист перед verified:true для «{{goal}}».
88. Сравни план vs факт по «{{goal}}»: отклонение scope, AC, files — таблица.
89. Security verification pass для {{service}}: authn/z, secrets, outbound URL, rate limits.
90. Контрактный тест: что должно остаться зелёным при смене реализации {{module}}.
91. Smoke после деплоя {{service}}: 5 команд/проверок Go/No-Go за 10 минут.
92. Аудит gaps: список того, что заявлено done по «{{goal}}», но без evidence.

## Portable-kit (93–102)

93. Составь чеклист переноса kit в новый репо: copy → bootstrap-test-context → smoke wave.
94. Проверь обезличенность: найди хардкод путей/доменов старого проекта в {{files}} kit.
95. Минимальный набор файлов kit для старта «{{goal}}» в чужом репо (must-copy list).
96. Адаптируй имена ролей: карта {{role}} → локальные роли целевого проекта.
97. Dry-run portable-kit: сценарий пробной волны на docs-задаче без прод-кода.
98. Заполни context-заготовки под модуль {{module}} / сервис {{service}} (какие секции).
99. Версионирование kit: что писать в STATE при переносе (версия, дата, known gaps).
100. Риски переносимости на другой стек (не тот язык/CI): что переписать vs что оставить.
101. Контракт промпта: какие секции обязательны в делегировании после переноса kit.
102. Post-port verification: 7 проверок, что kit «жив» (delegate, verify, handoff, no secrets).

## AI-team ops (103–112)

103. Составь roster волны: роль → агент → зона {{module}}/{{service}} → verifier.
104. Выбери модель/профиль для задачи «{{goal}}»: rationale cost vs risk (без имён вендоров — tier).
105. Эскалация: когда {{role}} обязан остановиться и вернуть задачу orchestrator (триггеры).
106. Параллель be/fe/qa: синхро-точки и запрет писать в чужие {{files}}.
107. Micro-task vs full wave: классифицируй «{{goal}}» — micro / single / wave; почему.
108. Prompt-contract audit: разбери входящий промпт — чего не хватает по шаблону §5.2.
109. Anti-pattern ops: ad-hoc «сделай X» без границ — перепиши в валидный delegate на {{files}}.
110. Метрики команды за волну: throughput планов, reopen rate, verification FAIL rate по {{module}}.
111. Ротация ownership: кто становится Accountable за {{service}} после выноса из монолита.
112. Ретро волны «{{goal}}»: 3 keep / 3 improve / 1 experiment на следующую волну.

## Композиции (типовые связки)

- Планирование фазы: `#1` → `#2` → `#4` → `#8` → `#23`
- Старт волны: `#16` → `#17` → `#26` → `#28`
- Strangler: `#43` → `#45` → `#52` → `#46` → `#50`
- PR-flow: `#57` → `#64` → `#58` → `#77`
- Закрытие: `#81` → `#82` → `#72` → `#80`
- Перенос kit: `#93` → `#94` → `#97` → `#102`
