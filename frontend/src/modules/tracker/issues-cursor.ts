import type { IssueResponse } from "@/api/tracker";
import type { PaginatedResponse } from "@/api/types";

/**
 * Склеивает страницы infinite-query в плоский список задач.
 * A10-002: loadMore раньше клал *текущую* страницу в allIssues и двигал cursor —
 * новая страница попадала в кэш Query, но не в display до следующего клика.
 */
export function flattenIssuePages(
	pages: PaginatedResponse<IssueResponse>[] | undefined,
): IssueResponse[] {
	if (!pages?.length) return [];
	return pages.flatMap((page) => page.items ?? []);
}

/** Курсор следующей страницы из ответа listIssues. */
export function nextIssuesCursor(
	lastPage: PaginatedResponse<IssueResponse>,
): string | undefined {
	if (lastPage.has_more && lastPage.next_cursor) {
		return lastPage.next_cursor;
	}
	return undefined;
}
