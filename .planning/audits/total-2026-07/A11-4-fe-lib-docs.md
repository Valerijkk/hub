# W11 агент A4 — fe lib+docs part 4

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус: hooks (unread/ws), lib utils, docs drift  
> Список: `W11-ASSIGNMENTS.md` §A4 (19 файлов)

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/hooks/use-unread.ts | 57 | 2 | замечания |
| frontend/src/hooks/use-websocket.ts | 110 | 1 | замечания |
| frontend/src/lib/onboarding/profile-schema.ts | 56 | 0 | чисто |
| frontend/src/lib/onboarding/project-schema.ts | 59 | 0 | чисто |
| frontend/src/lib/format-api-error.ts | 102 | 1 | замечания |
| frontend/src/lib/format-time.ts | 28 | 2 | замечания |
| frontend/src/lib/is-system-admin.ts | 4 | 0 | чисто |
| frontend/src/lib/keyboard.ts | 11 | 0 | чисто |
| frontend/src/lib/permissions.ts | 41 | 0 | чисто |
| frontend/src/lib/push-notifications.ts | 167 | 1 | замечания |
| frontend/src/lib/sanitize-color.ts | 8 | 0 | чисто |
| frontend/src/lib/utils.ts | 21 | 0 | чисто |
| docs/architecture.md | 236 | 3 | замечания |
| docs/BUSINESS.md | 464 | 2 | замечания |
| docs/code_style_guide.md | 286 | 2 | замечания |
| docs/competitive_analysis_2026.md | 198 | 1 | замечания |
| docs/competitive_audit.md | 174 | 1 | замечания |
| docs/GLOSSARY.md | 159 | 0 | чисто |
| docs/README.md | 97 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W11-A4-001 | P1 | frontend/src/hooks/use-unread.ts:24-29 | Синхронизация unread только `updateUnread` по ключам из ответа; каналы, исчезнувшие из Redis hash (`HDel` при MarkRead), в store не обнуляются | Backend `GetAll` = `HGetAll` — нулевые каналы отсутствуют в map. После прочтения на другом клиенте/вкладке локальный badge остаётся >0 до ручного `markRead` | Заменять весь `unreadCounts` ответом (`setUnreadCounts(query.data)`) или обнулять ключи store ∉ response |
| W11-A4-002 | P1 | docs/architecture.md:56 | Точка композиции: `main.go` (~820 LOC). Факт: `main.go` = 34 строки; wiring в `wiring_*.go` | Агенты/разработчики ищут middleware/DI по устаревшим якорям → ломает audits и onboarding | Переписать: `cmd/server/main.go` (thin) + `wiring_fiber.go` / `wiring_modules.go` / …; убрать «~820 LOC» |
| W11-A4-003 | P1 | docs/architecture.md:127 | «Фактический порядок middleware (`main.go:185-250`)» — строки не существуют; список без `HTTPTrace`/`Metrics`/`MaintenanceGate` | Drift от `wiring_fiber.go:48-75`; неверные ссылки = ложные находки в ревью | Ссылки → `wiring_fiber.go`; дополнить фактический порядок (RequestID→Recover→HTTPTrace→…→Maintenance→Audit) |
| W11-A4-004 | P2 | frontend/src/hooks/use-unread.ts:51-53 | `useMarkRead`: после optimistic `resetUnread` ошибки API глотаются (`catch {}`) без отката | При 5xx/сети badge = 0, сервер всё ещё считает unread → рассинхрон до следующего удачного poll (и даже poll не чинит W11-A4-001) | Rollback `updateUnread` из предыдущего значения / invalidate без предварительного reset; или reset только в `onSuccess` |
| W11-A4-005 | P2 | frontend/src/lib/format-time.ts:18-27 | Хардкод русских строк («Только что», «мин», «Вчера»); нет i18n | Онбординг/профиль поддерживают `language: ru\|en` — EN-UI всё равно показывает RU relative time | `i18n.t(...)` / `Intl.RelativeTimeFormat` по locale пользователя |
| W11-A4-006 | P2 | frontend/src/lib/format-time.ts:13-18 | При `dateStr` в будущем `diffMin < 0` → ветка `< 1` → «Только что» | Некорректный relative time для due dates / clock skew | Отдельная ветка для `diffMs < 0` («через N мин» / абсолютная дата) |
| W11-A4-007 | P2 | frontend/src/hooks/use-websocket.ts:80-109 | `useRooms` экспортирован, по репо **0 вызовов**; deps `[rooms]` по reference | Мёртвый API; при будущем `useRooms(["channel:"+id])` без `useMemo` — churn subscribe/unsubscribe каждый render | Удалить или задокументировать + сравнивать rooms по join/содержимому |
| W11-A4-008 | P2 | frontend/src/lib/push-notifications.ts:70-72 | `dismissPushBanner` пишет timestamp, но `shouldShowPushBanner` лишь проверяет presence ключа — «Позже» навсегда | Soft-ask не возвращается; пользователь, отложивший раз, больше не видит баннер | TTL (напр. 7 дней) по timestamp или отдельный «never» vs «later» |
| W11-A4-009 | P2 | docs/architecture.md:177,197,204,209 | Якоря `main.go:575-605`, `:540`, TD-01 «630 LOC», TD-06 `main.go:190` | Wiring split сделал все line-refs мёртвыми; TD-01 про «раздутый main» закрыт фактом | Обновить evidence на `wiring_*.go`; TD-01 → closed/rephrase |
| W11-A4-010 | P2 | docs/BUSINESS.md:442 | Out of Scope: «Webcam в звонках — Phase 16 (запланировано)» | В коде уже есть `setCameraEnabled` (`CallControls.tsx`); competitive_analysis помечает webcams ✅ | Перенести в shipped / уточнить остаточный gap (если есть) |
| W11-A4-011 | P2 | docs/BUSINESS.md:438 | «SSO/LDAP/OAuth \| Запрещено до ADR» одним пунктом | TZ 2026-07-03: SSO/LDAP **в scope** после `ADR-AUTH-SSO`; OAuth соцсетей — по-прежнему out | Разделить: SSO/LDAP = pending ADR; social OAuth = forbidden |
| W11-A4-012 | P2 | docs/code_style_guide.md:17 | «Язык комментариев: **English** в Go/TS» | После RU-COMMENTS волна код/хуки/lib — на русском; правило противоречит факту и playbook | Сменить канон на RU (или «как в соседнем файле модуля») |
| W11-A4-013 | P2 | docs/code_style_guide.md:36-44,153 | Структура модуля: `*_test.go` рядом с handler; FE: `frontend/src/testframework/unit/` | Тесты: `backend/testframework/` и `testframework/stages/unit/…` (0 в `src/`) — §«Тесты — стиль» ниже уже верный | Выровнять верхние деревья под фактические пути |
| W11-A4-014 | P2 | docs/competitive_analysis_2026.md:150 | Differentiator: «Magic Link + Telegram (без SSO complexity)» | Telegram — только notification bind (ADR-AUTH-01), не auth; формулировка читается как login | «Magic Link + Telegram notify bind» |
| W11-A4-015 | P2 | docs/competitive_audit.md:127 | Won't Have: «SSO/LDAP/OAuth провайдеры (до v5 ADR)» | Scope SSO одобрен до v5; блокировка = принятие ADR, не milestone v5 | «SSO/LDAP — после ADR-AUTH-SSO; social OAuth — out» |
| W11-A4-016 | P3 | frontend/src/lib/format-api-error.ts:16-20 | `extractApiError`: любой object с полем `message` → `as ApiError` | Обычный `Error`/DOMException может попасть в ветку apiErr без `code`/`request_id` | Узкий type guard: `code` number или `'request_id' in error` |
| W11-A4-017 | P3 | docs/architecture.md:105 | FE routes «~60+ UI pages» | Замер 2026-07-27: **72** file routes | Заменить на 72 (или «72 `.tsx` routes») |

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 12 |
| P3 | 2 |
| **Σ** | **17** |

### Топ-3
1. **W11-A4-001** — unread poll не сбрасывает каналы, исчезнувшие из Redis hash → залипшие badge после read на другом клиенте.
2. **W11-A4-002 / W11-A4-003** — `architecture.md` описывает монолитный `main.go` (~820 LOC) и мёртвые line-refs; канон — thin main + `wiring_*.go`.
3. **W11-A4-004** — optimistic `markRead` без rollback при ошибке API усиливает рассинхрон unread.

## Чисто
Проверено: FE hooks (deps/leaks/error swallow), lib (validation/security/i18n/RBAC mirror), docs (drift vs код/TZ/AGENTS).

| Файл | Что проверено |
|------|----------------|
| profile-schema.ts | Zod min/max/trim; `buildProfileDisplayName`; safeParse → field errors; language enum ru/en |
| project-schema.ts | create/join schemas; key regex `^[A-Z0-9]+$`; path remap name→projectName |
| is-system-admin.ts | `admin` \| `founder` (ORG-04) — без лишних ролей |
| keyboard.ts | Enter/Space + preventDefault для non-button a11y |
| permissions.ts | 26 бит 0…25 совпадают с `backend/internal/rbac/permissions.go` iota; `hasPermission` корректна |
| sanitize-color.ts | Только `#RRGGBB` — защита от CSS injection |
| utils.ts | `cn` + `buildQueryString` (skip null/undefined) — без секретов/any |
| GLOSSARY.md | Префиксы ID, модули backend, infra/test/процесс — согласовано с деревом `internal/` |
| README.md | Индекс docs, roadmap-ссылки, CI versions Go 1.25/Node 22, команды hubtest — актуальны |

**Фокус-вердикт:** код lib в целом крепкий (RBAC mirror, sanitize, Zod onboarding); главный FE-дефект — **unread sync**; главный docs-дефект — **architecture.md post-wiring drift** + расхождения SSO/webcam/Telegram между BUSINESS/competitive и TZ.
