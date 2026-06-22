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

/** Дефолтный UI ошибки -- карточка с понятным текстом и кнопкой повтора. */
function DefaultError({ retry }: { error: unknown; retry: () => void }) {
	return (
		<div
			role="alert"
			className="flex flex-col items-center justify-center gap-3 rounded-[var(--r4)] border border-b1 bg-card/80 p-6 text-center"
		>
			<div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/10 text-red-500">
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
			<p className="text-[14px] font-semibold text-t1">
				Не удалось загрузить данные. Попробовать снова?
			</p>
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
					<DefaultError error={query.error} retry={retry} />
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
