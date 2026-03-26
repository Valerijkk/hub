import { create } from "zustand";
import { persist } from "zustand/middleware";

/**
 * Состояние стора боковой панели.
 * Персистится в localStorage под ключом "hub-sidebar".
 */
interface SidebarState {
	/** Свёрнута ли боковая панель (только иконки). */
	collapsed: boolean;
	/** Переключить состояние свёрнутости. */
	toggle: () => void;
	/** Развернуть боковую панель. */
	expand: () => void;
}

/** Стор боковой панели навигации с персистом. */
export const useSidebarStore = create<SidebarState>()(
	persist(
		(set) => ({
			collapsed: false,
			toggle: () => set((state) => ({ collapsed: !state.collapsed })),
			expand: () => set({ collapsed: false }),
		}),
		{ name: "hub-sidebar" },
	),
);
