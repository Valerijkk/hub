/**
 * Стор дашбордов: управление данными виджетов и раскладками.
 * Поддерживает дебаунс сохранения раскладки при перетаскивании.
 */
import { create } from "zustand";
import type {
	AdminStats,
	LayoutItem,
	PersonalData,
	ProjectStats,
} from "@/api/dashboard";
import {
	getAdminDashboard,
	getLayout,
	getPersonalDashboard,
	getProjectDashboard,
	saveLayout,
} from "@/api/dashboard";

interface DashboardState {
	/** Данные персонального дашборда. */
	personalData: PersonalData | null;
	/** Данные проектного дашборда. */
	projectData: ProjectStats | null;
	/** Данные админского дашборда. */
	adminData: AdminStats | null;
	/** Раскладки по типу дашборда. */
	layouts: Record<string, LayoutItem[]>;
	/** Флаг загрузки. */
	loading: boolean;
	/** Ошибка загрузки данных. */
	error: string | null;

	/** Загрузить данные персонального дашборда. */
	fetchPersonal: () => Promise<void>;
	/** Загрузить данные проектного дашборда. */
	fetchProject: (projectId: string) => Promise<void>;
	/** Загрузить данные админского дашборда. */
	fetchAdmin: () => Promise<void>;
	/** Загрузить раскладку. */
	loadLayout: (type: string, projectId?: string) => Promise<void>;
	/** Сохранить раскладку с дебаунсом. */
	saveLayoutDebounced: (
		type: string,
		items: LayoutItem[],
		projectId?: string,
	) => void;
	/** Установить раскладку локально. */
	setLayout: (type: string, items: LayoutItem[]) => void;
}

let saveTimeout: ReturnType<typeof setTimeout> | null = null;

export const useDashboardStore = create<DashboardState>()((set, get) => ({
	personalData: null,
	projectData: null,
	adminData: null,
	layouts: {},
	loading: false,
	error: null,

	fetchPersonal: async () => {
		set({ loading: true, error: null });
		try {
			const data = await getPersonalDashboard();
			set({ personalData: data });
		} catch (err) {
			console.error("Ошибка загрузки персонального дашборда", err);
			set({ error: "Не удалось загрузить данные дашборда" });
		} finally {
			set({ loading: false });
		}
	},

	fetchProject: async (projectId: string) => {
		set({ loading: true, error: null });
		try {
			const data = await getProjectDashboard(projectId);
			set({ projectData: data });
		} catch (err) {
			console.error("Ошибка загрузки проектного дашборда", err);
			set({ error: "Не удалось загрузить данные дашборда" });
		} finally {
			set({ loading: false });
		}
	},

	fetchAdmin: async () => {
		set({ loading: true, error: null });
		try {
			const data = await getAdminDashboard();
			set({ adminData: data });
		} catch (err) {
			console.error("Ошибка загрузки админского дашборда", err);
			set({ error: "Не удалось загрузить данные дашборда" });
		} finally {
			set({ loading: false });
		}
	},

	loadLayout: async (type: string, projectId?: string) => {
		try {
			const layout = await getLayout(type, projectId);
			set((state) => ({
				layouts: { ...state.layouts, [type]: layout.items },
			}));
		} catch (err) {
			console.error("Ошибка загрузки раскладки", err);
			// error≠empty: fail API ≠ «пустая раскладка / default успех».
			set({ error: "Не удалось загрузить раскладку" });
			throw err;
		}
	},

	saveLayoutDebounced: (
		type: string,
		items: LayoutItem[],
		projectId?: string,
	) => {
		const previous = get().layouts[type];
		// Обновляем локально немедленно
		set((state) => ({
			layouts: { ...state.layouts, [type]: items },
			error: null,
		}));

		// Дебаунс сохранения на сервер
		if (saveTimeout) clearTimeout(saveTimeout);
		saveTimeout = setTimeout(() => {
			saveLayout(type, items, projectId).catch((err) => {
				console.error("Ошибка сохранения раскладки", err);
				set((state) => ({
					error: "Не удалось сохранить раскладку дашборда",
					layouts:
						previous !== undefined
							? { ...state.layouts, [type]: previous }
							: state.layouts,
				}));
			});
		}, 500);
	},

	setLayout: (type: string, items: LayoutItem[]) => {
		set((state) => ({
			layouts: { ...state.layouts, [type]: items },
		}));
	},
}));
