import type { UseQueryResult } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { Button } from "@/components/ui/button";

/**
 * Минимальный интерфейс query-объекта, который понимает QueryStateView.
 * Поддерживает любой результат TanStack Query (useQuery, useInfiniteQuery)
 * или совместимую структуру, которую можно собрать вручную.
 */
export interface QueryLike<T> {
	isLoading: boolean;
	isError?: boolean;
	error?: unknown;
	data?: T;
	refetch?: () => undefined | Promise<unknown>;
}

/** Пропсы QueryStateView. */
interface QueryStateViewProps<T> {
	/** Результат useQuery / useInfiniteQuery либо совместимый объект. */
	query: UseQueryResult<T> | QueryLike<T>;
	/** Пользовательский UI на время загрузки. Если не указан -- shimmer-скелетон. */
	loadingFallback?: ReactNode;
	/** Пользовательский UI при ошибке. По умолчанию -- карточка с кнопкой "Попробовать снова". */
	errorFallback?: (error: unknown, retry: () => void) => ReactNode;
	/** Функция-проверка, является ли результат пустым. Если не указана -- проверка не делается. */
	isEmpty?: (data: T) => boolean;
	/** UI для пустого состояния. */
	emptyFallback?: ReactNode;
	/** Рендер успешно загруженных данных. */
	children: (data: T) => ReactNode;
}

/** Дефолтный скелетон загрузки на основе уже существующего animate-shimmer. */
function DefaultLoading() {
	return (
		<div className="space-y-3" aria-busy="true" aria-live="polite">
			<div className="h-4 w-3/4 rounded-md animate-shimmer" />
			<div className="h-4 w-1/2 rounded-md animate-shimmer" />
			<div className="h-4 w-2/3 rounded-md animate-shimmer" />
		</div>
	);
}

/** Пропсы компактного/полного UI ошибки загрузки (P2-FE-01 / TASK-047). */
export interface QueryErrorAlertProps {
	/** Текст ошибки; по умолчанию — общая формулировка. */
	message?: string;
	/** Повтор запроса. */
	retry: () => void;
	/** Компактный вариант для sidebar / inline-блоков. */
	compact?: boolean;
}

/**
 * Единый UI «ошибка ≠ empty»: alert + retry.
 * Используется QueryStateView и ручными fetch-панелями (admin settings и т.п.).
 */
export function QueryErrorAlert({
	message = "Не удалось загрузить данные. Попробовать снова?",
	retry,
	compact = false,
}: QueryErrorAlertProps) {
	if (compact) {
		return (
			<div role="alert" className="space-y-2 py-2 text-center">
				<p className="text-[13px] text-[var(--red)]">{message}</p>
				<button
					type="button"
					onClick={retry}
					className="text-[13px] font-semibold text-t2 hover:text-t1 hover:underline"
				>
					Попробовать снова
				</button>
			</div>
		);
	}

	return (
		<div
			role="alert"
			className="flex flex-col items-center justify-center gap-3 rounded-[var(--r4)] border border-b1 bg-card/80 p-6 text-center"
		>
			<div className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--red)]/10 text-[var(--red)]">
				<svg
					className="h-5 w-5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
					aria-hidden="true"
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M12 9v2m0 4h.01M10.29 3.86l-8.18 14.18A2 2 0 003.84 21h16.32a2 2 0 001.73-2.96L13.71 3.86a2 2 0 00-3.42 0z"
					/>
				</svg>
			</div>
			<p className="text-[14px] font-semibold text-t1">{message}</p>
			<Button type="button" variant="outline" size="sm" onClick={retry}>
				Попробовать снова
			</Button>
		</div>
	);
}

/**
 * QueryStateView -- унифицированная обёртка для отображения состояний запроса.
 *
 * Решает UX-проблему: при сбоях `useQuery` UI часто проваливался в пустое
 * состояние ("Нет проектов", "Задач не найдено"), смешивая отсутствие данных
 * с ошибкой сети. Эта обёртка явно разделяет loading / error / empty / data.
 *
 * @example Пример использования
 * ```tsx
 * const q = useQuery({...});
 * return (
 *   <QueryStateView
 *     query={q}
 *     isEmpty={(d) => d.length === 0}
 *     emptyFallback={<EmptyState/>}
 *   >
 *     {(data) => <List items={data} />}
 *   </QueryStateView>
 * );
 * ```
 */
export function QueryStateView<T>({
	query,
	loadingFallback,
	errorFallback,
	isEmpty,
	emptyFallback,
	children,
}: QueryStateViewProps<T>) {
	if (query.isLoading) {
		return <>{loadingFallback ?? <DefaultLoading />}</>;
	}

	if (query.isError) {
		const retry = () => {
			void query.refetch?.();
		};
		return (
			<>
				{errorFallback ? (
					errorFallback(query.error, retry)
				) : (
					<QueryErrorAlert retry={retry} />
				)}
			</>
		);
	}

	const data = query.data;
	if (data === undefined || data === null) {
		// Нет данных и нет ошибки -- считаем, что это loading-edge (например, enabled=false).
		return <>{loadingFallback ?? <DefaultLoading />}</>;
	}

	if (isEmpty?.(data) && emptyFallback !== undefined) {
		return <>{emptyFallback}</>;
	}

	return <>{children(data)}</>;
}
