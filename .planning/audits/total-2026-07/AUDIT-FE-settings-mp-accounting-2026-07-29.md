# AUDIT FE — settings + marketplace + accounting

> TASK-008 · File-auditor · READ-ONLY · 2026-07-29  
> HEAD: `617af22372be61b98e7ea63d825b35075f495588` (`617af223`) · branch `cursor-grok-4.5`  
> Модель: `cursor-grok-4.5`  
> Периметр: `frontend/src/modules/settings/**/*.{ts,tsx}` · `frontend/src/modules/marketplace/**/*.{ts,tsx}` · `frontend/src/modules/accounting/**/*.{ts,tsx}`  
> Фокус: **error≠empty** · **sanitizeUrl** · **version optimistic lock UI** · **XSS**  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 остаётся открытым**

## Сверка с предшественниками (кратко)

| Источник | Статус в HEAD (этот периметр) |
|----------|-------------------------------|
| A3W10-1-001 (accounting invoices/trial/P&L `isError`) | **FIXED-OK** — `trialBalanceError` / `invoicesError` / `incomeStatementError` + `QueryErrorAlert` per section |
| A3W10-1-002 (`TemplateSelector` fail→empty grid) | **FIXED-OK** — `loadError` + `QueryErrorAlert` + loading |
| A3W10-1-004 (marketplace pending silently hide) | **FIXED-OK** — `pendingError \|\| pendingItems.length \|\| pendingLoading` + alert |
| A2W10-A1-005 / A3W10-2 webhooks list error≠empty | **FIXED-OK** — `loadError` + `QueryErrorAlert`; create toast |
| A2W10-A1-005 api-keys / A2W10-A1-006 ai-privacy load | **FIXED-OK** — list/load error UI |
| A3W4-5-002 (webhook Update VersionEQ на BE) | BE **FIXED-OK**; **FE не доведён** → **FE-SMA-001** |
| A3W5-3-006 (AI preferences VersionEQ на BE) | BE **FIXED-OK**; **FE не доведён** → **FE-SMA-002** |

**REGRESS:** 0 · **RESIDUAL (переоткрытие A3 actives в этом слоте):** 0 · **NEW:** 3

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `frontend/src/modules/settings/ai-privacy.tsx` | 109 | 1 | **NEW version** (load error≠empty OK) |
| `frontend/src/modules/settings/api-keys.tsx` | 233 | 0 | чисто (loading/error/empty) |
| `frontend/src/modules/settings/appearance.tsx` | 139 | 0 | чисто (локальный UI, без fetch) |
| `frontend/src/modules/settings/project-templates.tsx` | 205 | 0 | чисто (page + `TemplateSelector` error≠empty) |
| `frontend/src/modules/settings/webhooks.tsx` | 337 | 1 | **NEW version** (+ list error≠empty OK) |
| `frontend/src/modules/marketplace/marketplace-page.tsx` | 520 | 1 | **NEW loading≠empty** mine; pending/browse error OK; version mutate OK |
| `frontend/src/modules/accounting/accounting-page.tsx` | 456 | 0 | чисто по фокусу (secondary `isError` FIXED) |

Файлов в списке: **7**. Прочитано: **7**. Пропущено: **0**.

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| FE-SMA-001 | NEW | P2 | `webhooks.tsx:112–115` (+ типы `api/webhooks.ts:3–10,30–36`) | `updateWebhook(ep.id, { is_active })` **без** `version`; DTO/`WebhookEndpoint` без поля `version` | BE после A3W4-5-002: `version` ≥1 обязателен → toggle «Включить/Отключить» стабильно **400**; optimistic lock UI отсутствует. `handleToggleActive`/`handleDelete` без `try/catch` → silent reject | Добавить `version` в `WebhookEndpoint` + `UpdateWebhookRequest`; `updateWebhook(id, { is_active, version: ep.version })`; toast на fail toggle/delete/deliveries |
| FE-SMA-002 | NEW | P2 | `ai-privacy.tsx:20–21,40` (+ `api/ai.ts:92–102`) | `getAIPreferences`/`updateAIPreferences` без `version`; save шлёт только `allow_external_ai` | BE `SaveUserPreference`: update при `version<=0` → `ErrVersionConflict` (**409**). Первый create может пройти; повторное «Сохранить» / смена чекбокса для существующей записи — сломано | Тип `AIUserPreference` + `version`; хранить `pref.version` после load; `updateAIPreferences({ allow_external_ai, version })`; 409 → toast «обновите» + `refresh` |
| FE-SMA-003 | NEW | P3 | `marketplace-page.tsx:58–66,348–350,479–497` | Вкладка «Мои»: нет `isLoading`/`isPending` у `mine` query; empty/error гейтятся через `!browseLoading` | При `tab==="mine"` browse-query `enabled:false` → `browseLoading===false` → на первом fetch mine UI = «У вас нет объявлений» (loading≠empty) | `isPending: mineLoading`; loading UI для mine; empty/error: `!(tab==="browse"?browseLoading:mineLoading)` |

## Фокус-чек

### error≠empty

| Зона | Результат |
|------|-----------|
| api-keys list | **OK** — loading / `QueryErrorAlert` / «Нет API-ключей» |
| ai-privacy load | **OK** — `loadError \|\| !loaded` → alert+retry (форма не на default) |
| webhooks list | **OK** — loading / error / «Нет вебхуков»; missing `projectId` → отдельный alert |
| project-templates page + `TemplateSelector` | **OK** — A3W10-1-002 **FIXED-OK** |
| marketplace browse + pending admin | **OK** — A3W10-1-004 **FIXED-OK**; browse error≠empty |
| marketplace **mine** | **FE-SMA-003** (loading маскируется empty) |
| accounting accounts+journal early-return | **OK** |
| accounting invoices / trial / P&L | **OK** — A3W10-1-001 **FIXED-OK** |
| webhooks deliveries load | soft: fail `listDeliveries` без toast/UI (панель не открывается) — не заводилось отдельно |

### sanitizeUrl

| Проверка | Результат |
|----------|-----------|
| `href=` / `window.open` / external URL в 7 файлах | **0** — N/A |
| webhook `ep.url` | рендер как **text children**, не `<a href>` — XSS-навигация через scheme не применима |
| marketplace / accounting | нет user-controlled URL в DOM-ссылках |

### version optimistic lock UI

| Проверка | Результат |
|----------|-----------|
| marketplace approve/reject/sold/withdraw/update | **FIXED-OK** — `version: item.version` во всех mutate |
| webhooks update (toggle active) | **FE-SMA-001** |
| AI privacy save | **FE-SMA-002** |
| accounting post/issue/pay | N/A на FE — API body без client `version` (status-transition на BE; `version` только в ответе DTO) |
| api-keys / appearance / project-templates | N/A (нет concurrent update OL в UI) |

### XSS

| Проверка | Результат |
|----------|-----------|
| `dangerouslySetInnerHTML` / `innerHTML` / `eval` | **0** в 7 файлах |
| User text (titles, descriptions, names, webhook URL, delivery error body) | React text children |
| Created API key | `<code>{createdKey}</code>` — text, OK |
| ConfirmDialog marketplace title interpolation | React children / string prop — не HTML inject |

## FIXED-OK (сводка)

| Зона | Доказательство HEAD |
|------|---------------------|
| Accounting secondary error≠empty | `accounting-page.tsx` invoices/trial/P&L `isError` + alert |
| TemplateSelector error≠empty | `project-templates.tsx` loadError ветка |
| Marketplace pending queue | `pendingError` + visible admin section |
| Marketplace browse/mine **error** (не loading) | `listError` + retry vs empty copy |
| Marketplace OL mutate | `version` в approve/reject/sold/withdraw/update |
| Settings list loads (webhooks/api-keys/ai) | `loadError` + `QueryErrorAlert` |
| XSS / sanitizeUrl в периметре | нет raw HTML / external href |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 1 |
| **Total** | **3** |

| Метрика | Значение |
|---------|----------|
| Файлов | **7 / 7 / 0** |
| Активных | **3** (NEW 3 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (фокус) | A3W10-1-001/002/004 · list error≠empty settings · marketplace version mutate · XSS N/A |

### Топ-3

1. **FE-SMA-001 (P2, NEW)** — webhooks toggle без `version` после BE VersionEQ (фича «Вкл/Выкл» мертва).
2. **FE-SMA-002 (P2, NEW)** — AI privacy save без `version` → 409 на update существующей preference.
3. **FE-SMA-003 (P3, NEW)** — marketplace «Мои»: loading = empty «нет объявлений».

---

Файлов в списке: 7. Прочитано: 7. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=2 · P3=1 · Total=3.  
**TASK-008: открыт** · коммит не создавался.
