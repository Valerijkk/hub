/**
 * Универсальный ответ с пагинацией от API.
 * List-issues и messenger — курсорная (`next_cursor`/`has_more`, NFR-14/15,
 * ADR-PAGINATION-CURSOR-ISSUES-2026). Offset-поля — для других доменов
 * (intranet/wiki/admin), не для `listIssues`.
 */
export interface PaginatedResponse<T> {
	/** Массив элементов текущей страницы. */
	items: T[];
	/** Курсор для следующей страницы (курсорная пагинация). */
	next_cursor?: string;
	/** Есть ли ещё данные после текущей страницы. */
	has_more?: boolean;
	/** Общее количество элементов (для отображения пагинации). */
	total: number;
	/** Смещение текущей страницы (offset-пагинация; не list-issues). */
	offset?: number;
	/** Количество элементов на странице. */
	limit?: number;
}
