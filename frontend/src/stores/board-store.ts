import { create } from "zustand";

/** Фильтры доски для Kanban-представления. */
interface BoardFilters {
	/** Фильтр по статусам. */
	status?: string[];
	/** Фильтр по приоритетам. */
	priority?: string[];
	/** Фильтр по исполнителю. */
	assignee?: string;
	/** Фильтр по типу задачи. */
	type?: string;
	/** Поисковый запрос. */
	search?: string;
	/** Показывать только мои задачи. */
	myTasks?: boolean;
}

/** Состояние стора доски: фильтры, активная доска, состояние перетаскивания. */
interface BoardState {
	/** ID активной доски. */
	activeBoardId: string | null;
	/** Текущая конфигурация фильтров. */
	filters: BoardFilters;
	/** Идёт ли перетаскивание карточки. */
	isDragging: boolean;
	/** ID задачи, которая перетаскивается. */
	draggingIssueId: string | null;

	/** Установить активную доску. */
	setActiveBoard: (boardId: string) => void;
	/** Обновить фильтры (частичное слияние). */
	setFilters: (filters: Partial<BoardFilters>) => void;
	/** Сбросить все фильтры. */
	resetFilters: () => void;
	/** Установить состояние перетаскивания. */
	setDragging: (isDragging: boolean, issueId?: string | null) => void;
}

const defaultFilters: BoardFilters = {};

/** Стор управления состоянием Kanban-доски. */
export const useBoardStore = create<BoardState>()((set) => ({
	activeBoardId: null,
	filters: defaultFilters,
	isDragging: false,
	draggingIssueId: null,

	setActiveBoard: (boardId) => set({ activeBoardId: boardId }),
	setFilters: (filters) =>
		set((state) => ({ filters: { ...state.filters, ...filters } })),
	resetFilters: () => set({ filters: defaultFilters }),
	setDragging: (isDragging, issueId = null) =>
		set({ isDragging, draggingIssueId: issueId }),
}));
