# Frontend: производительность и локализация

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

Оптимизация рендера и интернационализация.

---

## 11. Производительность

### 11.0. Метрики и бюджеты Core Web Vitals (2026)

Производительность меряется метриками, а не ощущениями. **INP (Interaction to Next Paint) ≤200ms** заменил FID и является самым проваливаемым CWV-метриком 2026 года; Lighthouse-score как прокси CWV — признанный антипаттерн (лабораторный скор ≠ полевые данные).

Целевые бюджеты Hub (с буфером к порогам Google):

| Метрика | Цель Hub | Порог Google «good» |
|---------|----------|---------------------|
| INP | < 150 ms | ≤ 200 ms |
| LCP | < 2.0 s | ≤ 2.5 s |
| CLS | < 0.05 | ≤ 0.1 |
| JS основного маршрута | < 300 KB gzip | — |

Практика:
- Длинные обработчики (drag на доске, фильтрация больших списков) бей на куски: long task >50ms блокирует отклик — используй `scheduler.yield()` / `startTransition` между фазами.
- `content-visibility: auto` для тяжёлых оффскрин-секций.
- Полевые метрики важнее лабораторных: RUM через библиотеку `web-vitals` (p75 INP/LCP/CLS) — план в observability-стек; алерт при p75 INP >200ms.
- Lighthouse CI сейчас artifact-only (NFR-17); перевод в blocking assertions — на обсуждении (Приложение B).

### 11.1. Lazy-загрузка тяжёлых зависимостей (обязательно)

Тяжёлые библиотеки не должны попадать в основной бандл. Грузим их через динамический `import()` только когда нужны.

**Обязательно lazy:** LiveKit (звонки), чарты (dashboard), редакторы (Plate editor), крупные модули (calls, dashboard).

```tsx
const CallsRoom = lazy(() => import("@/modules/calls/calls-room"));
const Dashboard = lazy(() => import("@/routes/dashboard"));
```

Всегда с `<Suspense fallback>` и Error Boundary (LiveKit может упасть на несовместимом браузере).

### 11.2. Виртуализация длинных списков

Списки на сотни+ элементов (сообщения, большие таблицы задач) виртуализируем (рендерим только видимое). Для чата это сочетается с infinite scroll (`useInfiniteQuery`). Используй виртуализацию (например `@tanstack/react-virtual`), когда список стабильно > ~100 строк и есть тормоза при скролле.

<bad-example>
Рендер 5000 сообщений разом — фриз UI, гигантский DOM:

```tsx
{messages.map((m) => <MessageRow key={m.id} message={m} />)} // 5000 узлов
```
</bad-example>

<good-example>
Виртуализация: в DOM только видимые строки + буфер. `key` — по `m.id`.
</good-example>

### 11.3. Debounce / throttle

Частые события (ввод в поиск, resize, скролл) throttle/debounce. Канон — `useDebouncedValue`:

```tsx
export function useDebouncedValue<T>(value: T, delayMs = 300): T {
	const [debounced, setDebounced] = useState(value);
	useEffect(() => {
		const id = setTimeout(() => setDebounced(value), delayMs);
		return () => clearTimeout(id); // критично: чистим предыдущий таймер
	}, [value, delayMs]);
	return debounced;
}
```

Применение: дебаунсим строку поиска, затем кладём в `queryKey`, чтобы не слать запрос на каждый символ:

```tsx
const debouncedSearch = useDebouncedValue(search, 300);
const { data } = useQuery({
	queryKey: ["search", debouncedSearch],
	queryFn: ({ signal }) => searchApi.search(debouncedSearch, { signal }),
	enabled: debouncedSearch.length >= 2,
});
```

### 11.4. Разбиение бандла

- Route-level split: тяжёлые маршруты — `.lazy.tsx`/`lazy()`.
- Vendor-чанки: крупные библиотеки выносятся сборкой автоматически, но не тяни тяжёлое в общий модуль без нужды.
- Не импортируй весь пакет ради одной функции (`import { thing } from "lib"`, не `import * as lib`).

### 11.5. Мемоизация (по делу)

См. раздел 3.7. Кратко: `useMemo` для дорогих вычислений (фильтрация больших массивов), `useCallback` для колбэков в deps/мемо-дочерних, `memo` для дочернего, который часто получает те же пропсы, но перерисовывается из-за родителя.

### 11.6. Избегание лишних эффектов и ререндеров

- Не клади в `useEffect` то, что можно вычислить при рендере (производные значения — просто переменные или `useMemo`).
- Узкие Zustand-селекторы (раздел 6.2) вместо подписки на весь стор.
- Не создавай новые объекты/массивы в пропсах мемоизированных дочерних без мемоизации (иначе `memo` бесполезен).

<bad-example>
Эффект для производного значения — лишний рендер и рассинхрон:

```tsx
const [fullName, setFullName] = useState("");
useEffect(() => setFullName(`${first} ${last}`), [first, last]); // не нужно
```
</bad-example>

<good-example>
```tsx
const fullName = `${first} ${last}`; // просто вычисляем при рендере
```
</good-example>

### 11.7. Skeletons вместо спиннеров

Для восприятия скорости используем skeleton-заглушки (см. `boardLoadingFallback`), а не голый спиннер на весь экран. Skeleton отражает будущий layout — меньше «прыжков».

### 11.8. Чек-лист производительности

- [ ] (2026) Бюджеты CWV соблюдены: INP <150ms, LCP <2.0s, CLS <0.05; долгие обработчики разбиты (`scheduler.yield()`/`startTransition`).
- [ ] Тяжёлые зависимости (LiveKit, charts, editor) — lazy + Suspense + Error Boundary.
- [ ] Длинные списки виртуализированы (> ~100 строк) со стабильными ключами.
- [ ] Поиск/resize/скролл — debounce/throttle с очисткой таймеров.
- [ ] Нет `import * as` ради одной функции; тяжёлое не в общем модуле.
- [ ] Мемоизация обоснована; производные значения не через эффект.
- [ ] Узкие селекторы Zustand.
- [ ] Loading через skeleton, отражающий layout.

---
## 14. Интернационализация (i18n)

Стек: **i18next** + **react-i18next** + **i18next-browser-languagedetector**. Языки: `ru` (fallback) и `en`. Namespaces: `common`, `tracker`, `messenger`, `tms`, `calls`, `dashboard`, `settings`, `help`, `onboarding`.

### 14.1. Никакого хардкода user-visible строк

Любой текст, который видит пользователь, идёт через `t(...)`. Хардкод-строки в новом коде — замечание `fe-ver`.

<good-example>
```tsx
const { t } = useTranslation(["tracker", "common"]);
<Label htmlFor="title">{t("form.title")}</Label>
<Button type="submit">{t("form.submitCreate")}</Button>
```
</good-example>

<bad-example>
```tsx
<Label htmlFor="title">Название задачи</Label> // хардкод -> непереводимо
```
</bad-example>

> Примечание: в исторических участках Hub встречаются русские строки прямо в JSX (например фильтры доски). Это техдолг. **Новый** код пишем через i18n; при существенной правке legacy-экрана — выносим строки в ключи.

### 14.2. Ключи и namespaces

- Ключ — по смыслу и месту: `form.title`, `priorities.high`, `kanban.moveError`.
- Namespace выбираем по домену: `useTranslation("tracker")` или список `["tracker", "common"]`.
- Кросс-namespace ключ — с префиксом: `t("common:btn.saving")`.

```tsx
const { t } = useTranslation(["tracker", "common"]);
t("form.createSuccess");      // из tracker
t("common:btn.saving");       // из common
```

Файлы переводов: `src/i18n/locales/ru/<ns>.json` и `.../en/<ns>.json`. Новый ключ добавляется **в оба** языка (ru и en) синхронно — иначе один язык покажет ключ вместо текста.

### 14.3. Интерполяция и плюрализация

Интерполяция:

```json
{ "issues.count": "Задач: {{count}}" }
```

```tsx
t("issues.count", { count: total });
```

Плюрализация (i18next поддерживает формы по `count`):

```json
{
	"comments_zero": "Нет комментариев",
	"comments_one": "{{count}} комментарий",
	"comments_few": "{{count}} комментария",
	"comments_many": "{{count}} комментариев"
}
```

```tsx
t("comments", { count: comments.length });
```

**Правило:** не собирай множественное число конкатенацией (`count + " задач"`) — используй плюрал-формы i18next.

### 14.4. Форматирование дат в TZ пользователя

ТЗ (жёстко): **время в БД всегда UTC**; фронтенд конвертирует в локальный timezone пользователя. Бэк отдаёт ISO/RFC3339 в UTC — на фронте форматируем в локаль.

```tsx
// Дата из API (UTC ISO) -> локальный формат пользователя
const formatted = new Intl.DateTimeFormat(i18n.language, {
	dateStyle: "medium",
	timeStyle: "short",
}).format(new Date(issue.created_at));
```

**Правила дат:**
- Никогда не показывай сырой UTC-ISO пользователю.
- Не форматируй дату руками через конкатенацию — используй `Intl.DateTimeFormat`/общий хелпер (`lib/`), учитывающий локаль (`i18n.language`).
- Относительное время («5 минут назад») — через локале-зависимый форматтер (`Intl.RelativeTimeFormat`).
- Отправляя дату на бэк — отправляй в UTC/ISO (`toISOString()`), не в локальном виде.

### 14.5. Переключение языка

Есть `components/language-switcher.tsx`; выбор кэшируется в localStorage (`i18nextLng`) детектором. При смене языка перерисовываются переведённые компоненты — не кэшируй перевод в `useState`.

### 14.6. Чек-лист i18n

- [ ] Нет хардкода user-visible строк; всё через `t(...)`.
- [ ] Ключ добавлен в **оба** языка (ru + en), нужный namespace.
- [ ] Интерполяция через `{{var}}`; множественное — через плюрал-формы, не конкатенацию.
- [ ] Даты: UTC из бэка -> `Intl.DateTimeFormat` в локали пользователя; на бэк — ISO/UTC.
- [ ] Строки в `aria-label`/placeholder/тостах — тоже через i18n.

---
