import type { BoardFilters } from "@/stores/board-store";

/**
 * Клиентская фильтрация карточек канбана (поиск, priority, type, assignee, myTasks).
 * Вынесено для unit-теста P2-FE-06: чип «Мои задачи» обязан скрывать чужие.
 */
export function filterBoardIssue(
	issue: {
		title: string;
		issue_key: string;
		priority: string;
		issue_type: string;
		assignee?: { id: string } | null;
	},
	filters: BoardFilters,
	currentUserId?: string | null,
): boolean {
	if (filters.search) {
		const q = filters.search.toLowerCase();
		if (
			!issue.title.toLowerCase().includes(q) &&
			!issue.issue_key.toLowerCase().includes(q)
		) {
			return false;
		}
	}
	if (filters.priority?.length && !filters.priority.includes(issue.priority)) {
		return false;
	}
	if (filters.type && issue.issue_type !== filters.type) {
		return false;
	}
	if (filters.assignee && issue.assignee?.id !== filters.assignee) {
		return false;
	}
	if (filters.myTasks && currentUserId) {
		if (issue.assignee?.id !== currentUserId) {
			return false;
		}
	}
	return true;
}
