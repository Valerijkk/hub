import { create } from "zustand";
import { searchGlobal, type SearchResult } from "@/api/search";

const RECENT_SEARCHES_KEY = "hub_recent_searches";
const MAX_RECENT = 10;

/** Состояние глобального поиска. */
interface SearchState {
  /** Открыта ли палитра команд. */
  isOpen: boolean;
  /** Текущий поисковый запрос. */
  query: string;
  /** Результаты поиска. */
  results: SearchResult[];
  /** Индикатор загрузки. */
  loading: boolean;
  /** Недавние поисковые запросы (из localStorage). */
  recentSearches: string[];

  /** Открыть/закрыть палитру. */
  setOpen: (open: boolean) => void;
  /** Выполнить поиск с debounce (вызывается из компонента). */
  search: (query: string) => Promise<void>;
  /** Очистить результаты. */
  clearResults: () => void;
  /** Добавить запрос в историю. */
  addRecentSearch: (query: string) => void;
  /** Загрузить недавние поиски из localStorage. */
  loadRecentSearches: () => void;
}

/** Загрузка недавних поисков из localStorage. */
function getRecentFromStorage(): string[] {
  try {
    const stored = localStorage.getItem(RECENT_SEARCHES_KEY);
    return stored ? (JSON.parse(stored) as string[]) : [];
  } catch {
    return [];
  }
}

/** Сохранение недавних поисков в localStorage. */
function saveRecentToStorage(searches: string[]) {
  try {
    localStorage.setItem(RECENT_SEARCHES_KEY, JSON.stringify(searches));
  } catch {
    // Игнорируем ошибки localStorage
  }
}

export const useSearchStore = create<SearchState>((set, get) => ({
  isOpen: false,
  query: "",
  results: [],
  loading: false,
  recentSearches: getRecentFromStorage(),

  setOpen: (open) => {
    set({ isOpen: open });
    if (!open) {
      set({ query: "", results: [], loading: false });
    }
  },

  search: async (query) => {
    set({ query, loading: true });
    if (!query.trim()) {
      set({ results: [], loading: false });
      return;
    }
    try {
      const response = await searchGlobal(query.trim());
      // Проверяем, что запрос не изменился за время ожидания
      if (get().query === query) {
        set({ results: response.results, loading: false });
      }
    } catch {
      if (get().query === query) {
        set({ results: [], loading: false });
      }
    }
  },

  clearResults: () => set({ results: [], query: "" }),

  addRecentSearch: (query) => {
    const trimmed = query.trim();
    if (!trimmed) return;
    const current = get().recentSearches.filter((s) => s !== trimmed);
    const updated = [trimmed, ...current].slice(0, MAX_RECENT);
    set({ recentSearches: updated });
    saveRecentToStorage(updated);
  },

  loadRecentSearches: () => {
    set({ recentSearches: getRecentFromStorage() });
  },
}));
