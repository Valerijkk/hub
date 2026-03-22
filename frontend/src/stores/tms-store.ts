import { create } from "zustand";
import type {
	SuiteResponse,
	TestCasePriority,
	TestCaseResponse,
	TestCaseStatus,
	TestRunDetailResponse,
	TestRunResponse,
} from "@/api/tms";

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

/** Состояние стора TMS модуля. */
interface TmsState {
	/** Список тест-кейсов. */
	testCases: TestCaseResponse[];
	/** Выбранный тест-кейс. */
	selectedTestCase: TestCaseResponse | null;
	/** Дерево папок. */
	suites: SuiteResponse[];
	/** Выбранная папка. */
	selectedSuite: SuiteResponse | null;
	/** Список прогонов. */
	testRuns: TestRunResponse[];
	/** Выбранный прогон (с результатами). */
	selectedRun: TestRunDetailResponse | null;
	/** Текущие фильтры. */
	filters: TmsFilters;
	/** Общее количество тест-кейсов (для пагинации). */
	totalTestCases: number;

	/** Установить список тест-кейсов. */
	setTestCases: (cases: TestCaseResponse[], total: number) => void;
	/** Установить выбранный тест-кейс. */
	setSelectedTestCase: (tc: TestCaseResponse | null) => void;
	/** Установить дерево папок. */
	setSuites: (suites: SuiteResponse[]) => void;
	/** Установить выбранную папку. */
	setSelectedSuite: (suite: SuiteResponse | null) => void;
	/** Установить список прогонов. */
	setTestRuns: (runs: TestRunResponse[]) => void;
	/** Установить выбранный прогон. */
	setSelectedRun: (run: TestRunDetailResponse | null) => void;
	/** Обновить фильтры (частичное слияние). */
	updateFilter: (filter: Partial<TmsFilters>) => void;
	/** Сбросить все фильтры. */
	resetFilters: () => void;
}

const defaultFilters: TmsFilters = {};

/** Стор управления состоянием модуля TMS. */
export const useTmsStore = create<TmsState>()((set) => ({
	testCases: [],
	selectedTestCase: null,
	suites: [],
	selectedSuite: null,
	testRuns: [],
	selectedRun: null,
	filters: defaultFilters,
	totalTestCases: 0,

	setTestCases: (cases, total) =>
		set({ testCases: cases, totalTestCases: total }),
	setSelectedTestCase: (tc) => set({ selectedTestCase: tc }),
	setSuites: (suites) => set({ suites }),
	setSelectedSuite: (suite) => set({ selectedSuite: suite }),
	setTestRuns: (runs) => set({ testRuns: runs }),
	setSelectedRun: (run) => set({ selectedRun: run }),
	updateFilter: (filter) =>
		set((state) => ({ filters: { ...state.filters, ...filter } })),
	resetFilters: () => set({ filters: defaultFilters }),
}));
