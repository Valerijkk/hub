# Hub — Операционный план возврата (конец июля 2026)

> Companion к `JULY-2026-REBUILD-BRIEF.md` и `PROJECT-STATUS-2026-07-10.md`.
> Цель: день/неделя, RACI, DoR/DoD, риски, чек-лист первого дня.

---

## 1. Таймлайн

### Неделя 0 — Align (D1–D3)

| День | Что | Who | Выход |
|------|-----|-----|-------|
| D1 AM | Ревью INDEX + PROJECT-STATUS + Brief + seeds | user + team-lead | Shared context |
| D1 PM | Ответы Brief §5 (mail L*, reschedule RSVP, redesign scope, приоритет эпиков, судьба W2–11) | user + product | Kickoff decisions MD |
| D2 | Merge `integration/wave1` → master (rebase-merge); личное ревью bases | **user** | master = Wave1 |
| D2–D3 | FA-06 Phase 1–3 если нужен prod/tag | user + devops | Secrets rotated или явно «отложить cutover» |
| D3 | Приоритет очереди: QA-HARDENING / CAL-SYNC / Redesign / Track P | product + tech-lead | Ordered backlog |

**Не делать в W0:** код CAL-SYNC, пилить LiveKit, OAuth calendar, старт W2 вслепую.

### Неделя 1 — Planning expansion

| Поток | Who | Работа | Depends |
|-------|-----|--------|---------|
| Edge → TC | qa-manual | TC из `EDGE-CORNER-CATALOG.md` (P0 домены) | Catalog exists |
| Guide adoption | docs-writer + team-lead | `TESTING_GUIDE.md` в playbooks/PR checklist | Guide exists |
| CAL prep | tech-lead | Атомарные PLAN.md CAL-1..10 | Align CAL-0 |
| Redesign prep | designer | RD-1 inventory | Mockups + RD-0 |
| Charter | all be/fe | CODE-QUALITY-CHARTER в DoD | — |

### Неделя 2 — Specs / ADR / Design

| Поток | Who | Артефакт |
|-------|-----|----------|
| CAL US/AC | ba2 → ba-ver | vision US-CAL-RSVP-* |
| CAL ERD/API | sa2 → sa-ver | system_spec + contracts |
| ADR-CAL-UNIFIED | arch → arch-ver | `.planning/decisions/` |
| Redesign SPEC | designer → designer-ver | DESIGN-SPEC |
| QH-P2 | qa-manual-ver | Verified TC batch #1 |

### Неделя 3+ — Execute waves

Обычный конвейер: `ba→sa→arch→designer→be/fe→qa→*-ver→/qa-gate`.
Параллель A/B без пересечения файлов. Master — только user.

Порядок по умолчанию (если Align не сказал иначе):
1. QH-P2/P3 на security-critical (auth/ws/rbac) — быстрый ROI
2. CAL-SYNC CAL-5..8
3. Redesign RD-3..N на затронутых экранах
4. Track F W2 / Track P M0 — только если Align = continue framework

---

## 2. RACI (агенты)

| Активность | R | A | C | I |
|------------|---|---|---|---|
| Kickoff / приоритеты | product | user | team-lead, tech-lead | all |
| Декомпозиция планов | tech-lead | team-lead | arch, sa | be/fe/qa |
| Merge master | **user** | user | team-lead, code-reviewer | agents |
| FA-06 | user | user | devops, security-auditor | team-lead |
| CAL US/AC | ba2 | product | sa2 | ba-ver |
| CAL ADR/ERD | arch / sa2 | arch | security-auditor | be |
| CAL BE/FE | be1/be2, fe1/fe2 | team-lead | designer | *-ver |
| Redesign | designer, fe* | product | fe-ver | qa |
| QA-HARDENING catalog→TC | qa-manual | team-lead | ba/sa | qa-auto* |
| Автотесты | qa-auto1/2 | qa-auto-ver | be/fe (юниты) | team-lead |
| Wave gate | team-lead | user (merge) | *-ver, code-reviewer | — |
| Docs sync | docs-writer | team-lead | domain owners | — |

Модели: team-lead = Fable 5; team-lead-opus = эскалация; остальные Sonnet 5.

---

## 3. Definition of Ready (DoR) — план/фаза

- [ ] Цель и success criteria измеримы
- [ ] Depends on явны; нет скрытого «и» двух слоёв в одном плане
- [ ] Границы файлов не пересекаются с параллельным агентом
- [ ] US/AC или FR/API есть (для фич) / TC-ID есть (для QA)
- [ ] Security/TZ инварианты учтены
- [ ] Entry criteria волны выполнены
- [ ] Для CAL: ADR-CAL-UNIFIED хотя бы draft; mail level ≤ L1 без нового ADR

## 4. Definition of Done (DoD) — план

- [ ] Атомарный коммит в фича-ветке (≤200–300 LOC ориентир)
- [ ] Доменная регрессия green
- [ ] `*-ver` + лог → `verified: true`
- [ ] Docs-sync (STATE + domain doc)
- [ ] CODE-QUALITY-CHARTER соблюдён
- [ ] Нет секретов / `DEV_MODE` в prod paths
- [ ] Для волны: `/qa-gate` на интеграции (team-lead)
- [ ] Merge master — только по команде user

---

## 5. Риски и митигации

| Риск | Митигация |
|------|-----------|
| Merge wave1 отложен → drift | Rebase `integration/wave1` на master; не копить W2 на outdated base |
| OAuth calendar «втихую» | GZ-V4-05; L3 только с ADR override |
| SFU снова всплывёт | SEED-001 REJECTED; M4 = LiveKit |
| CAL + Calendar v2 дубли | ADR-CAL-UNIFIED канон; ScheduledCall specialization |
| Redesign раздувает scope | RD-0 MoSCoW; точечные экраны first |
| QA catalog без автотестов | QH-P2 DoD = verified TC; P3/P4 в той же волне |
| CI минут (3000/мес) | path-filter; не full gate у каждого агента |
| FA-06 забыт до prod | Ops checklist в D2; tag v2.5.0 только после FA-06 |

---

## 6. Чек-лист «первый день возврата»

1. [ ] Открыть `.planning/INDEX-JULY-2026.md`
2. [ ] Прочитать `PROJECT-STATUS-2026-07-10.md` + Brief §1–§5
3. [ ] Подтвердить: **LiveKit stay** (не mediasoup)
4. [ ] Ответить на открытые Qs Brief §5
5. [ ] `git fetch` · checkout `integration/wave1` · smoke / FA-10 лог
6. [ ] Rebase-merge Wave1 → master (личное ревью bases)
7. [ ] Решить FA-06: сейчас / после feature freeze
8. [ ] Забрать макеты Cloud Code Designer → designer
9. [ ] Проверить `EDGE-CORNER-CATALOG.md` + `TESTING_GUIDE.md`
10. [ ] Запустить team-lead: очередь QH / CAL / RD по приоритету
11. [ ] tech-lead: нарезка первой execute-волны
12. [ ] Не стартовать Track P M2/M4 и TF-PY в день 1

---

## 7. Порядок работ при возврате (кратко)

1. INDEX → PROJECT-STATUS → Brief
2. SFU = LiveKit; SEED-001 не поднимать
3. Brief §5 answers
4. Merge wave1 → master
5. FA-06 или явный defer cutover
6. Приоритет: QH security → CAL → RD (или Align override)
7. Testing Guide + CODE-QUALITY-CHARTER в PR DoD
8. qa-manual: TC из catalog (P0)
9. ba2/sa2/arch: CAL-SYNC US → ERD → ADR
10. designer: Cloud Code inventory + SPEC
11. tech-lead: атомарные PLAN
12. be/fe execute; qa-auto по verified TC
13. Track F W2 / Track P M0 — только если Align = continue
14. M4 hub-calls = LiveKit; Locust baseline до split
15. Волна: домен regression → `/qa-gate` → merge user
16. Docs-sync + STATE в той же сессии

---

*tech-lead 2026-07-10 · записано на диск parent-агентом · код не трогать до Align*
