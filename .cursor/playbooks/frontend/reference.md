# Frontend: справочник

> Часть плейбука. Канон и оглавление — [`../frontend.md`](../frontend.md).

Сквозной пример модуля и обновления экосистемы 2026.

---

## Приложение A. Сквозной пример новой фичи

Чтобы связать все разделы, разберём реализацию фичи «список меток проекта с добавлением» от API до маршрута. Это эталон порядка работы и слоёв для fe1/fe2.

### A.1. Шаг 1 — API-слой (`api/tracker.ts`)

Сначала фиксируем контракт: DTO и функции. Они уже есть в `tracker.ts`, используем как есть:

```typescript
export interface LabelResponse {
	id: string;
	name: string;
	color: string;
	group_name?: string;
}

export interface CreateLabelRequest {
	name: string;
	color: string;
	group_id?: string;
}

export function listLabels(projectId: string) {
	return api.get<LabelResponse[]>(`/projects/${projectId}/labels`);
}

export function createLabel(projectId: string, data: CreateLabelRequest) {
	return api.post<LabelResponse>(`/projects/${projectId}/labels`, data);
}

export function deleteLabel(projectId: string, labelId: string) {
	return api.delete<void>(`/projects/${projectId}/labels/${labelId}`);
}
```

**Что важно:** типы — источник истины для UI; функции — тонкие обёртки над `api`; ответы типизированы дженериком.

### A.2. Шаг 2 — Query-хук (`components/tracker/use-labels.ts` или `modules/tracker/hooks/`)

Оборачиваем API в Query-хуки. Server-state целиком в Query, ключи по конвенции.

```typescript
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	type CreateLabelRequest,
	createLabel,
	deleteLabel,
	listLabels,
} from "@/api/tracker";

/** Список меток проекта (редко меняется -> длинный staleTime). */
export function useLabels(projectId: string) {
	return useQuery({
		queryKey: ["labels", projectId],
		queryFn: () => listLabels(projectId),
		staleTime: 5 * 60_000,
		enabled: !!projectId,
	});
}

/** Создание метки с инвалидацией списка. */
export function useCreateLabel(projectId: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (data: CreateLabelRequest) => createLabel(projectId, data),
		onSuccess: () => {
			void qc.invalidateQueries({ queryKey: ["labels", projectId] });
		},
	});
}

/** Удаление метки с оптимистичным обновлением и откатом. */
export function useDeleteLabel(projectId: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (labelId: string) => deleteLabel(projectId, labelId),
		onMutate: async (labelId) => {
			await qc.cancelQueries({ queryKey: ["labels", projectId] });
			const previous = qc.getQueryData<import("@/api/tracker").LabelResponse[]>([
				"labels",
				projectId,
			]);
			qc.setQueryData<import("@/api/tracker").LabelResponse[]>(
				["labels", projectId],
				(old) => old?.filter((l) => l.id !== labelId) ?? old,
			);
			return { previous };
		},
		onError: (_err, _labelId, context) => {
			if (context?.previous) {
				qc.setQueryData(["labels", projectId], context.previous);
			}
		},
		onSettled: () => {
			void qc.invalidateQueries({ queryKey: ["labels", projectId] });
		},
	});
}
```

**Что важно:** `useLabels` — чтение; `useCreateLabel` — простая мутация с инвалидацией; `useDeleteLabel` — полный optimistic-цикл (cancel/snapshot/rollback/settled). UI не знает про `fetch` — только про хуки.

### A.3. Шаг 3 — Валидация формы (`components/tracker/label-form-schema.ts`)

```typescript
import { z } from "zod";

/** Схема формы метки: имя 1-40, цвет — hex. Сообщения — i18n-ключи. */
export const labelFormSchema = z.object({
	name: z.string().trim().min(1, "labels.nameRequired").max(40, "labels.nameMax"),
	color: z
		.string()
		.regex(/^#[0-9a-fA-F]{6}$/, "labels.colorInvalid"),
});

export type LabelFormValues = z.infer<typeof labelFormSchema>;
```

### A.4. Шаг 4 — Компонент формы (`components/tracker/label-form.tsx`)

```tsx
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { useCreateLabel } from "@/components/tracker/use-labels";
import {
	type LabelFormValues,
	labelFormSchema,
} from "@/components/tracker/label-form-schema";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toastApiError } from "@/lib/format-api-error";

/** Форма добавления метки проекта (RHF + Zod). */
export function LabelForm({ projectId }: { projectId: string }) {
	const { t } = useTranslation(["tracker", "common"]);
	const createMutation = useCreateLabel(projectId);

	const {
		register,
		handleSubmit,
		reset,
		formState: { errors, isValid, isSubmitting },
	} = useForm<LabelFormValues>({
		resolver: zodResolver(labelFormSchema),
		mode: "onChange",
		defaultValues: { name: "", color: "#5B9CF5" },
	});

	const onSubmit = (values: LabelFormValues) => {
		createMutation.mutate(values, {
			onSuccess: () => {
				toast.success(t("labels.created"));
				reset();
			},
			onError: (err) => toastApiError(err, t("labels.createError")),
		});
	};

	const isPending = isSubmitting || createMutation.isPending;

	return (
		<form onSubmit={(e) => void handleSubmit(onSubmit)(e)} className="space-y-4">
			<div className="space-y-2">
				<Label htmlFor="label-name">{t("labels.name")}</Label>
				<Input
					id="label-name"
					{...register("name")}
					aria-invalid={errors.name ? true : undefined}
					aria-describedby={errors.name ? "label-name-error" : undefined}
				/>
				{errors.name ? (
					<p id="label-name-error" role="alert" className="text-sm text-red-500">
						{t(errors.name.message as string)}
					</p>
				) : null}
			</div>

			<div className="space-y-2">
				<Label htmlFor="label-color">{t("labels.color")}</Label>
				<Input id="label-color" type="color" {...register("color")} className="h-9 w-16 p-1" />
				{errors.color ? (
					<p role="alert" className="text-sm text-red-500">
						{t(errors.color.message as string)}
					</p>
				) : null}
			</div>

			<Button type="submit" disabled={isPending || !isValid}>
				{isPending ? t("common:btn.saving") : t("labels.add")}
			</Button>
		</form>
	);
}
```

### A.5. Шаг 5 — Список с удалением (`components/tracker/label-list.tsx`)

```tsx
import { useTranslation } from "react-i18next";
import { useDeleteLabel, useLabels } from "@/components/tracker/use-labels";

/** Список меток проекта с удалением. */
export function LabelList({ projectId }: { projectId: string }) {
	const { t } = useTranslation("tracker");
	const { data: labels, isLoading, isError } = useLabels(projectId);
	const deleteMutation = useDeleteLabel(projectId);

	// Хуки выше — далее условный рендер
	if (isLoading) return <LabelListSkeleton />;
	if (isError) return <p className="text-sm text-t3">{t("labels.loadError")}</p>;
	if (!labels || labels.length === 0) {
		return <p className="text-sm text-t3">{t("labels.empty")}</p>;
	}

	return (
		<ul className="space-y-2">
			{labels.map((label) => (
				<li
					key={label.id}
					className="flex items-center justify-between rounded-[var(--r2)] border border-b1 bg-card px-3 py-2"
				>
					<span className="flex items-center gap-2 text-[13px] text-t1">
						<span
							aria-hidden="true"
							className="h-3 w-3 rounded-full"
							style={{ backgroundColor: label.color }}
						/>
						{label.name}
					</span>
					<button
						type="button"
						aria-label={t("labels.deleteAria", { name: label.name })}
						onClick={() => deleteMutation.mutate(label.id)}
						disabled={deleteMutation.isPending}
						className="text-t3 hover:text-red-500"
					>
						<svg aria-hidden="true" className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
							<path strokeLinecap="round" strokeLinejoin="round" d="M6 6l12 12M6 18L18 6" />
						</svg>
					</button>
				</li>
			))}
		</ul>
	);
}

function LabelListSkeleton() {
	return (
		<ul className="space-y-2">
			{[1, 2, 3].map((i) => (
				<li key={`label-sk-${i}`} className="h-10 rounded-[var(--r2)] animate-shimmer" />
			))}
		</ul>
	);
}
```

**Разбор по playbook:**
- Все хуки (`useLabels`, `useDeleteLabel`) — до условного рендера (раздел 3.2).
- `key={label.id}` — стабильный (3.3); скелетон — `label-sk-${i}`, не голый индекс.
- Цветовой индикатор — единственный оправданный `style` (рантайм-цвет из данных, не из дизайн-токена; раздел 8.4), при этом `aria-hidden`.
- Иконочная кнопка удаления имеет `aria-label` через i18n с интерполяцией (разделы 9.3, 14.3).
- Кнопка дизейблится на `isPending` (10.6).
- Строки через `t(...)` (14.1).

### A.6. Шаг 6 — Встраивание в маршрут

```tsx
import { createFileRoute } from "@tanstack/react-router";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { LabelForm } from "@/components/tracker/label-form";
import { LabelList } from "@/components/tracker/label-list";

function ProjectLabelsPage() {
	const { projectId } = Route.useParams();
	return (
		<ErrorBoundary>
			<section className="mx-auto max-w-lg space-y-6 p-6">
				<LabelForm projectId={projectId} />
				<LabelList projectId={projectId} />
			</section>
		</ErrorBoundary>
	);
}

export const Route = createFileRoute("/_authenticated/projects/$projectId/settings/labels/")({
	component: ProjectLabelsPage,
});
```

### A.7. Шаг 7 — Регрессия, тесты, docs-sync

Выполнить task-owned limited Docker job `frontend-unit` из `AGENTS.md` и
приложить source/image/resources/collection/duration/cleanup evidence.

- Genuine unit в `frontend/tests/tracker/` — рендер формы, ввод имени, сабмит (MSW изолирует HTTP boundary), проверка тоста; список — рендер, удаление, проверка оптимистичного исчезновения и отката при 500.
- Новый экран настроек -> при необходимости обновить `QUICKSTART.md`; `.planning/STATE.md` — session note.
- Атомарные коммиты: `feat(tracker): add project labels api hooks`, затем `feat(tracker): add label form and list ui`, затем `feat(tracker): add project labels settings route`.

Эта последовательность (API -> хуки -> схема -> компоненты -> маршрут -> регрессия/docs) — рекомендуемый порядок для любой фронтенд-фичи в Hub.

---
## Приложение B. Дельты 2026 на обсуждение

> Источник: `.planning/research/BEST-PRACTICES-2026-frontend-testing.md` (2026-07-03). Ниже — дельты medium/low со статусом «обсудить»: в playbook не внесены, решение — за team-lead/arch. Не применять без решения.

| # | Тема | Суть | Предлагаемое действие |
|---|------|------|----------------------|
| 9 | Simplified optimistic (Query v5) | Читать `mutation.variables` при `isPending` и рисовать оптимистичный элемент без записи в кэш (rollback не нужен) | Дополнить §5.5 выбором: variables-паттерн (одно место) vs кэш-паттерн (несколько мест) |
| 10 | Offline / persist для PWA | `persistQueryClient` + `PersistQueryClientProvider` (`gcTime ≥ maxAge`), `networkMode: 'offlineFirst'` | Оценить для PWA: минимум — задокументировать осознанный отказ; максимум — persist для справочников (channels, projects) |
| 11 | Строгие TS-флаги | `verbatimModuleSyntax`, `noUncheckedSideEffectImports`, `erasableSyntaxOnly` (дешёвые), `exactOptionalPropertyTypes` (шумный, последним) | Включать поэтапно в `frontend/tsconfig.json` с прогоном `tsc --noEmit` |
| 12 | TypeScript 5.9 | 5.8: `erasableSyntaxOnly`; 5.9: `strictInference` входит в `strict` | Плановый chore-апгрейд с прогоном; ожидать новые ошибки от strictInference |
| 16 | Perf-бюджеты в CI | Lighthouse CI assertions (blocking) + size-limit на PR (JS <300KB gzip, diff чанка) | Перевести lighthouse job из artifact-only в blocking; закрывает NFR-17. Учесть бюджет GitHub Actions |
| 17 | Vite 7 (rolldown) | Rust-бандлер, быстрее билд; миграция с v6 малоболезненная | Chore-апгрейд после стабилизации экосистемы плагинов (pwa, router plugin) |
| 20 | Biome-строгость | a11y-правила и `noExplicitAny` — `warn` → `error` (иначе gate декоративный); `noFloatingPromises`; GritQL-плагин для Hub-запретов | Поднять уровни в `frontend/biome.json`; плагин — nice-to-have |
| 21 | `ref` как проп (React 19) | `forwardRef` не нужен для новых компонентов | Правило для нового кода: без `forwardRef`; существующие Shadcn-обёртки не трогать |

---

*Playbook фронтенда Hub. Источники истины: `.cursor/rules/30-frontend.mdc`, `.cursor/rules/10-style.mdc`, `.cursor/rules/60-tz-constraints.mdc`, `docs/code_style_guide.md`. При расхождении — правила и ТЗ приоритетны.*
