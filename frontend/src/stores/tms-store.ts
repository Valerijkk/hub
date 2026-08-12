import { create } from "zustand";
import type { TestCasePriority, TestCaseStatus } from "@/api/tms";

/** Фильтры для списка тест-кейсов. */
interface TmsFilters {
	/** ID папки для фильтрации. */
	suiteId?: string;
	/** Статус тест-кейса. */
	status?: TestCaseStatus;
	/** Приоритет тест-кейса. */
	priority?: TestCasePriority;
	/** Поисковый запрос. */
	search?: string;
}

/** Client-state стора TMS: фильтры UI (server-state — TanStack Query). */
interface TmsState {
	/** Текущие фильтры. */
	filters: TmsFilters;
	/** Обновить фильтры (частичное слияние). */
	updateFilter: (filter: Partial<TmsFilters>) => void;
	/** Сбросить все фильтры. */
	resetFilters: () => void;
}

const defaultFilters: TmsFilters = {};

/** Стор UI-фильтров модуля TMS. */
export const useTmsStore = create<TmsState>()((set) => ({
	filters: defaultFilters,
	updateFilter: (filter) =>
		set((state) => ({ filters: { ...state.filters, ...filter } })),
	resetFilters: () => set({ filters: defaultFilters }),
}));
