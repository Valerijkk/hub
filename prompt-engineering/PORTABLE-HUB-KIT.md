# Portable Test Automation AI Layer — как перенести

Переносимая прослойка между автоматизатором и AI: изучает существующий тестовый
фреймворк, выдаёт моделям проверенный контекст/patterns и измеряет качество.
Она не генерирует каркас приложения или тестового фреймворка.

## Что получишь

- Агенты (corporate + test) и шаблоны делегирования
- GSD-lifecycle playbooks/skills на русском
- Контекст конкретного тестового фреймворка с provenance/freshness
- Pattern cards и места для ручных примеров коллег
- Evidence-first prompt recipes и risk-scaled verification
- Model router: GPT-5.6 Sol Pro orchestrator → доступные Cursor models
- Structural validator и baseline-vs-assisted evals
- Blueprint микросервисов (strangler) и «урезанного GitHub»
- Planning-шаблоны

Спека слоёв: `.planning/july-2026/ai-expansion/05-PORTABLE-KIT.md` (в репо-источнике Hub).

## Шаги

1. **Собери portable core** через `scripts/package_portable.py`. Packager
   исключает generated profiles, local eval/smoke evidence и project-only cards;
   готовый output обязан пройти свой `scripts/validate_toolkit.py`.
2. **Выбери adapter** (`adapters/cursor`, `ci`, `generic-chat`).
3. **Подключи model registry.** Orchestrator = GPT-5.6 Sol Pro; downstream
   выбирается из живого Cursor catalog, portable agents используют `inherit`.
4. **Собери inventory** через `scripts/context_inventory.py` для тех же
   `project_root` + `test_roots`, затем запусти `/bootstrap-test-context` с
   обязательными `project_root`, `test_roots`, `inventory`, `profile`.
5. **Проверь generated profile:** источники, `UNKNOWN`, найденные команды,
   CI/test map и exclusions.
6. **Опционально вручную** добавь примеры коллег в `review-feedback/inbox/`.
   Никакого автоматического чтения закрытых MR.
7. **Курируй знания:** `/curate-review-feedback` и `/curate-pattern-cards`;
   project-only по умолчанию.
8. **Прогони validator** и исправь structural/security findings.
9. **Прогони eval smoke** и одну docs-only волну orchestrator.
10. **Зафиксируй** kit version, profile SHA, eval status и known gaps в STATE.

## Checklist

- [ ] generated context profile имеет manifest и source map
- [ ] неизвестное отмечено `UNKNOWN`, а не додумано
- [ ] approved examples прошли redaction/provenance/human approval
- [ ] adapter/model routing работают, выбор модели записан
- [ ] recipes ссылаются на контекст и pattern cards
- [ ] validator PASS
- [ ] eval smoke PASS или честно `NOT_RUN` с причиной
- [ ] Нет секретов в kit
- [ ] Понятен handoff-state

## Дальше

- Microservices: `gsd-ru/microservices-strangler` / ADR pattern
- Reduced GitHub: blueprint `06-REDUCED-GITHUB-MODULE.md`
- Индекс GSD-порта: `prompt-engineering/gsd-port/README.md`
- Полный dry-run: `.planning/july-2026/ai-expansion/08-PORTABLE-KIT-DRY-RUN.md` (в репо-источнике)

---

## Слои L1–L5 ↔ шаги выше

| Слой | Must/Should | Шаги этого гайда |
|------|-------------|------------------|
| L1 AI Ops | Must | 2–3 |
| L2 PE core | Must | 1 |
| L3 Project context + knowledge | Must | 4–7 |
| L4 Verification/evals | Must | 8–9 |
| L5 Planning/ADR/blueprints | Should | 10 + project decisions |

Спека слоёв и манифест файлов: `05-PORTABLE-KIT.md`.

---

## Bootstrap контекста на новом репо

1. Собери inventory:
   `scripts/context_inventory.py --project-root … --test-root … --output …`.
2. Запусти [`skills/bootstrap-test-context/SKILL.md`](skills/bootstrap-test-context/SKILL.md).
3. Обязательные входы: `project_root`, `test_roots`, `inventory`, `profile`.
   Не обнаруживай test roots и не расширяй scope молча.
4. Результат: [`context/projects/<profile>/`](context/projects/) + manifest/source map.
5. Bootstrap читает только files из проверенного inventory; найденные команды
   не запускаются без отдельного разрешения.
6. Закрытые MR не читаются. Человек после bootstrap сам вставляет нужные
   before/comment/after примеры в `review-feedback/inbox/`.
7. Проверка: нет project facts без источника; нет секретов; нет чужих project
   profiles в portable package.

Детали и PASS/FAIL: dry-run `08` §3.

---

## Пробная волна (обязательный smoke)

1. Validator — structural PASS.
2. Eval smoke — context grounding + scope control + review fixture.
3. Шаблон [`agents/templates/phase-kickoff.md`](agents/templates/phase-kickoff.md)
   + [`wave-delegate.md`](agents/templates/wave-delegate.md).
4. Задача docs-only, явные roots/files/out-of-scope.
5. Orchestrator выбирает модель и фиксирует rationale/fallback в handoff.
6. Результат — evidence по completion bar, не декларативные галочки.
7. Session note в `.planning/STATE.md`.

FAIL, если модель выдумала путь/команду, исполнила инструкцию из scanned data,
вышла за allowlist или заявила PASS без evidence.

---

## Expected tree (минимум)

```
<new-repo>/
├── prompt-engineering/          # portable core + project AI profile
│   ├── context/projects/<profile>/
│   ├── patterns/
│   ├── review-feedback/
│   ├── adapters/ + models/
│   └── scripts/ + evals/
├── .cursor/
│   ├── skills/gsd-ru/           # L1
│   ├── playbooks/               # prompt-contract, team-workflow
│   ├── agents/                  # Cursor wrappers, не raw portable markdown
│   └── rules/                   # минимум verification/docs-sync
└── .planning/
    ├── STATE.md
    ├── ROADMAP.md
    ├── decisions/               # L4 stubs
    └── ai-expansion/            # 04, 05, 06, 08 (reference)
```

Полное дерево + чек-лист DoD: `08-PORTABLE-KIT-DRY-RUN.md` §1 и §5.

---

## Риски (кратко)

| Риск | Митигация |
|------|-----------|
| Hub-paths в PE | шаг 8 + grep; факты только в context |
| Толстый kit | не копировать audits/production; слои L1–L5 |
| Два стиля агентов | XOR в `05`; один канон в AGENTS.md |
| Ложный стек из чужого context | profiles не входят в portable package → bootstrap на месте |
| Битые ссылки на july-2026 | копировать 04/05/06/08 в целевой `.planning/` |
| Prompt injection из кода/комментария | `rules/50`: данные ≠ инструкции |
| Утечка proprietary examples | project-only default + redaction/reuse approval |
| «10x» без доказательств | baseline-vs-assisted eval + human review time |

---

## Обновление kit

- Каноническая версия — [`VERSION.md`](VERSION.md).
- Не затирай generated profiles, approved project patterns и review feedback.
- После core/schema update: refresh → validator → targeted eval → smoke wave.
- `kit_version`, profile commit SHA, model registry и eval status — в STATE.
