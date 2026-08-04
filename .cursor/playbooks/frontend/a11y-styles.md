# Frontend: доступность и стили

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

WCAG 2.2 AA, семантика, система стилей.

---

## 9. Доступность (a11y, WCAG 2.2 AA)

Hub целится в **WCAG 2.2 AA** *(2026: переход с 2.1; WCAG 2.2 — ISO/IEC 40500:2025, обратно совместим с 2.1)*. `fe-ver` гоняет axe и проверяет клавиатурную навигацию. A11y — не опция.

Новые для 2.2 критерии, критичные для Hub (детали и дизайн-требования — `.cursor/playbooks/design.md` §5.7): **2.5.7 Dragging Movements** (Kanban DnD обязан иметь альтернативу без drag — перемещение через меню/кнопки), **2.5.8 Target Size ≥24×24px** (иконочные кнопки), **2.4.11 Focus Not Obscured** (фокус не перекрыт sticky-панелями). Magic Link уже удовлетворяет 3.3.8 Accessible Authentication; 4.1.1 Parsing упразднён.

### 9.1. Семантика

Используй правильные элементы: `<button>` для действий, `<a>`/`<Link>` для навигации, заголовки `<h1>..<h3>` по иерархии, списки для списков.

<good-example>
Кнопка — это `<button>` с типом:

```tsx
<button type="button" onClick={handleClose}>Cancel</button>
```
</good-example>

<bad-example>
`<div onClick>` вместо кнопки — не фокусируется с клавиатуры, не читается скринридером:

```tsx
<div onClick={handleClose}>Cancel</div>
```
</bad-example>

### 9.2. `type="button"` обязателен

Любая кнопка внутри `<form>`, которая **не** отправляет форму, должна иметь `type="button"` — иначе по умолчанию `type="submit"` и она сабмитит форму. Кнопка отправки — `type="submit"`.

```tsx
<button type="button" onClick={resetFilters}>Сбросить</button>
<Button type="submit" disabled={isPending || !isValid}>Сохранить</Button>
```

### 9.3. `aria-label` на иконочных кнопках

Кнопка без текста (только иконка) обязана иметь `aria-label`, а декоративная svg — `aria-hidden="true"`.

<good-example>
```tsx
<button type="button" aria-label={t("issue.delete")} onClick={onDelete}>
	<svg aria-hidden="true" className="h-4 w-4" /* ... */ />
</button>
```
</good-example>

<bad-example>
Иконочная кнопка без метки — скринридер читает «button», пользователь не знает назначения:

```tsx
<button type="button" onClick={onDelete}>
	<TrashIcon />
</button>
```
</bad-example>

Также `aria-label` на элементах управления без видимой подписи (например, Select выбора доски):

```tsx
<SelectTrigger aria-label="Выбор доски" className="...">
```

(В проде текст `aria-label` должен идти через i18n — см. раздел 14.)

### 9.4. Focus ring и видимый фокус

Не убирай фокус-обводку без замены. Используй `focus-visible:` утилиты. Ошибочные поля — `aria-invalid` + связанное сообщение.

```tsx
<Input
	id="title"
	aria-invalid={errors.title ? true : undefined}
	aria-describedby={errors.title ? "title-error" : undefined}
	className={errors.title ? "border-red-500 focus-visible:ring-red-500" : undefined}
/>
{errors.title ? (
	<p id="title-error" role="alert" className="text-sm text-red-500">
		{fieldError("title")}
	</p>
) : null}
```

Это реальный паттерн из `IssueForm`: `aria-invalid`, `aria-describedby`, `role="alert"` на сообщении об ошибке.

<bad-example>
Снятие обводки без замены — клавиатурный пользователь теряет позицию:

```tsx
<button className="outline-none" /> // нет focus-visible замены
```
</bad-example>

### 9.5. Клавиатурная навигация (Tab / Enter / Escape)

- Всё интерактивное достижимо `Tab` (нативные элементы — из коробки; кастомные — `tabIndex={0}` + обработчики клавиш).
- `Enter`/`Space` активируют кнопки; для кастомных — обрабатывай `onKeyDown`.
- `Escape` закрывает модалы/поповеры — у Shadcn/Radix это встроено, **не отключай**.
- Модалки: фокус-trap и возврат фокуса — уже в Radix Dialog, не ломай.

```tsx
function handleKeyDown(e: React.KeyboardEvent) {
	if (e.key === "Enter" || e.key === " ") {
		e.preventDefault();
		onActivate();
	}
	if (e.key === "Escape") onClose();
}
```

### 9.6. Контраст 4.5:1

Текст должен иметь контраст ≥ 4.5:1 к фону (для крупного текста ≥ 3:1). Токены `--hub-text-*` подобраны под это — поэтому и нельзя хардкодить цвета: легко получить нечитаемую пару. Проверяй нестандартные сочетания (текст на градиенте, на цветном бейдже).

### 9.7. Skip link и лендмарки

Есть `components/layout/skip-link.tsx` — ссылка «к основному контенту». Основные регионы размечай лендмарками: `<nav>`, `<main>`, `<aside>`, `<header>`. Один `<h1>` на страницу.

### 9.8. axe-проверки

`fe-ver` запускает axe (в т.ч. в E2E на ключевых экранах — login axe PASS уже есть). Перед сдачей сам прогоняй axe DevTools на новом экране: 0 critical/serious violations.

### 9.9. Чек-лист a11y

- [ ] Семантические элементы (`button`/`a`/заголовки/списки), не `div` с onClick.
- [ ] `type="button"` на неотправляющих кнопках; `type="submit"` на отправке.
- [ ] Иконочные кнопки имеют `aria-label`; декоративные svg — `aria-hidden`.
- [ ] Видимый focus-visible; поля ошибок — `aria-invalid` + `aria-describedby` + `role="alert"`.
- [ ] Tab/Enter/Space/Escape работают; Radix-поведение (Esc, focus-trap) не сломано.
- [ ] Контраст ≥ 4.5:1 (через токены).
- [ ] Лендмарки и один `<h1>`; skip-link на месте.
- [ ] axe: 0 critical/serious на новом экране.
- [ ] WCAG 2.2 (2026): drag-действия имеют альтернативу без drag (2.5.7); интерактивные цели ≥24×24px (2.5.8); фокус не перекрыт sticky-контентом (2.4.11).

---
## 10. Формы и валидация

Стек форм в Hub: **react-hook-form** + **@hookform/resolvers/zod** + **zod**. Схема валидации — отдельным файлом рядом с формой. (2026) Для простейших одноактных форм без схемной валидации допустима Actions-триада React 19 (`useActionState`/`useFormStatus`) — см. раздел 3.9; сложные формы — только RHF + Zod.

### 10.1. Схема Zod (отдельный файл)

Ключи сообщений — **i18n-ключи**, а не готовые тексты. Форма переводит их через `t(...)`.

```typescript
import { z } from "zod";

export const issueFormSchema = z.object({
	title: z.string().trim().min(3, "form.titleMinLength").max(200, "form.titleMaxLength"),
	issue_type: z.enum(["bug", "task", "story", "epic", "sub_task", "improvement", "tech_debt", "spike", "release"]),
	priority: z.enum(["critical", "high", "medium", "low", "trivial"]),
	assignee_id: z.string().optional(),
	estimate_minutes: z
		.string()
		.optional()
		.refine((v) => !v || (!Number.isNaN(Number(v)) && Number(v) >= 0), "form.estimateInvalid"),
	story_points: z.string().optional(),
	due_date: z.string().optional(),
});

export type IssueFormValues = z.infer<typeof issueFormSchema>;
```

**Правила:**
- Тип значений выводим из схемы: `z.infer<typeof schema>` — не дублируем руками.
- Сообщения — i18n-ключи (`"form.titleMinLength"`).
- Кастомная логика — через `.refine`.

### 10.2. Подключение RHF + zodResolver

```tsx
const {
	register,
	control,
	handleSubmit,
	formState: { errors, isValid, isSubmitting },
} = useForm<IssueFormValues>({
	resolver: zodResolver(issueFormSchema),
	mode: "onChange", // валидация по мере ввода -> можно дизейблить submit по isValid
	defaultValues: {
		title: initialData?.title ?? "",
		issue_type: initialData?.issue_type ?? "task",
		priority: initialData?.priority ?? "medium",
		// ...
	},
});
```

### 10.3. Простые поля через `register`, сложные через `Controller`

- Нативные `<input>` — `register("field")`.
- Кастомные контролы (Shadcn Select, кастомный picker) — `<Controller>`, потому что у них нет нативного `ref`/`onChange` в формате RHF.

```tsx
{/* Нативный input */}
<Input id="title" {...register("title")} aria-invalid={errors.title ? true : undefined} />

{/* Кастомный Select через Controller */}
<Controller
	name="priority"
	control={control}
	render={({ field }) => (
		<Select value={field.value} onValueChange={field.onChange}>
			<SelectTrigger id="priority"><SelectValue /></SelectTrigger>
			<SelectContent>
				{PRIORITY_KEYS.map((p) => (
					<SelectItem key={p} value={p}>{t(`priorities.${p}`)}</SelectItem>
				))}
			</SelectContent>
		</Select>
	)}
/>
```

### 10.4. Отправка + связь с мутацией

`handleSubmit` валидирует и вызывает обработчик с типизированными значениями. Дальше — мутация Query (раздел 5).

```tsx
const onSubmit = (values: IssueFormValues) => {
	if (mode === "create") {
		createMutation.mutate({ title: values.title.trim(), /* ... */ });
	} else {
		updateMutation.mutate(values);
	}
};

<form onSubmit={(e) => void handleSubmit(onSubmit)(e)} className="space-y-6">
```

Промис `handleSubmit(...)()` помечаем `void` (Biome).

### 10.5. Отображение ошибок

Показываем ошибку под полем, переводя i18n-ключ из схемы:

```tsx
const fieldError = (key: keyof IssueFormValues) => {
	const err = errors[key];
	if (!err?.message) return null;
	return t(err.message as string);
};

{errors.title ? (
	<p id="title-error" role="alert" className="text-sm text-red-500">
		{fieldError("title")}
	</p>
) : null}
```

### 10.6. Состояния disabled / loading

Кнопка отправки дизейблится, пока идёт запрос **или** форма невалидна; показывает индикатор.

```tsx
const isPending = isSubmitting || createMutation.isPending || updateMutation.isPending;

<Button type="submit" disabled={isPending || !isValid} className="w-full">
	{isPending ? t("common:btn.saving") : mode === "create" ? t("form.submitCreate") : t("form.submitEdit")}
</Button>
```

**Правило:** запрещай двойной сабмит (дизейбл на `isPending`). Не полагайся только на `isValid` — сеть тоже может падать, обрабатывай `onError`.

### 10.7. Оптимистическая блокировка (version) в формах редактирования

ТЗ требует `version` в каждой изменяемой сущности. Форма редактирования **обязана** отправлять `version` из загруженных данных — бэк вернёт 409 при конфликте:

```tsx
return updateIssue(projectId, initialData.issue_key, {
	title: values.title.trim(),
	// ...
	version: initialData.version, // optimistic locking — обязательно
});
```

При 409 (`ApiError.code === 409`) показывай пользователю, что данные устарели, и предложи перезагрузить.

### 10.8. Чек-лист форм

- [ ] Схема Zod в отдельном файле; тип из `z.infer`; сообщения — i18n-ключи.
- [ ] RHF + `zodResolver`; `mode: "onChange"` где нужен live-контроль валидности.
- [ ] Нативные поля — `register`; кастомные контролы — `Controller`.
- [ ] Ошибки под полем + `aria-invalid`/`aria-describedby`/`role="alert"`.
- [ ] Submit дизейблится на `isPending`/`!isValid`; нет двойного сабмита; есть индикатор.
- [ ] Форма редактирования шлёт `version` (optimistic locking); 409 обрабатывается.
- [ ] Успех/ошибка -> `toast` (`toastApiError` для ошибок).

---
