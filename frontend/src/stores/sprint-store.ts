import { create } from "zustand";
import type {
	BurndownPoint,
	CreateSprintRequest,
	PokerSessionResponse,
	RetroNotesRequest,
	SprintResponse,
	VelocityPoint,
} from "@/api/sprint";
import * as sprintApi from "@/api/sprint";

/** Состояние стора спринтов. */
interface SprintState {
	/** Список спринтов проекта. */
	sprints: SprintResponse[];
	/** Активный спринт (если есть). */
	activeSprint: SprintResponse | null;
	/** Данные burndown текущего спринта. */
	burndown: BurndownPoint[];
	/** Данные velocity проекта. */
	velocity: VelocityPoint[];
	/** Текущая сессия Planning Poker. */
	pokerSession: PokerSessionResponse | null;
	/** Флаг загрузки. */
	loading: boolean;
	/** Ошибка загрузки списка (не путать с пустым списком). */
	error: string | null;

	/** Загрузить спринты проекта. */
	fetchSprints: (projectId: string) => Promise<void>;
	/** Создать новый спринт. */
	createSprint: (
		projectId: string,
		data: CreateSprintRequest,
	) => Promise<SprintResponse>;
	/** Запустить спринт. */
	startSprint: (projectId: string, sprintId: string) => Promise<void>;
	/** Завершить спринт. */
	completeSprint: (projectId: string, sprintId: string) => Promise<void>;
	/** Добавить задачи в спринт. */
	moveIssueToSprint: (
		projectId: string,
		sprintId: string,
		issueIds: string[],
	) => Promise<void>;
	/** Загрузить данные burndown. */
	fetchBurndown: (projectId: string, sprintId: string) => Promise<void>;
	/** Загрузить данные velocity. */
	fetchVelocity: (projectId: string) => Promise<void>;
	/** Сохранить заметки ретроспективы. */
	saveRetroNotes: (
		projectId: string,
		sprintId: string,
		data: RetroNotesRequest,
	) => Promise<void>;
	/** Установить сессию Planning Poker. */
	setPokerSession: (session: PokerSessionResponse | null) => void;
}

/** Стор управления спринтами. */
export const useSprintStore = create<SprintState>()((set, get) => ({
	sprints: [],
	activeSprint: null,
	burndown: [],
	velocity: [],
	pokerSession: null,
	loading: false,
	error: null,

	fetchSprints: async (projectId) => {
		set({ loading: true, error: null });
		try {
			const sprints = await sprintApi.listSprints(projectId);
			const activeSprint = sprints.find((s) => s.status === "active") ?? null;
			set({ sprints, activeSprint, loading: false, error: null });
		} catch {
			set({
				loading: false,
				error: "Не удалось загрузить спринты",
			});
		}
	},

	createSprint: async (projectId, data) => {
		const sprint = await sprintApi.createSprint(projectId, data);
		set({ sprints: [sprint, ...get().sprints] });
		return sprint;
	},

	startSprint: async (projectId, sprintId) => {
		const updated = await sprintApi.startSprint(projectId, sprintId);
		const sprints = get().sprints.map((s) => (s.id === sprintId ? updated : s));
		set({ sprints, activeSprint: updated });
	},

	completeSprint: async (projectId, sprintId) => {
		const updated = await sprintApi.completeSprint(projectId, sprintId);
		const sprints = get().sprints.map((s) => (s.id === sprintId ? updated : s));
		set({ sprints, activeSprint: null });
	},

	moveIssueToSprint: async (projectId, sprintId, issueIds) => {
		await sprintApi.addIssuesToSprint(projectId, sprintId, {
			issue_ids: issueIds,
		});
		await get().fetchSprints(projectId);
	},

	fetchBurndown: async (projectId, sprintId) => {
		const burndown = await sprintApi.getBurndown(projectId, sprintId);
		set({ burndown });
	},

	fetchVelocity: async (projectId) => {
		const velocity = await sprintApi.getVelocity(projectId);
		set({ velocity });
	},

	saveRetroNotes: async (projectId, sprintId, data) => {
		const updated = await sprintApi.updateRetroNotes(projectId, sprintId, data);
		const sprints = get().sprints.map((s) => (s.id === sprintId ? updated : s));
		set({ sprints });
	},

	setPokerSession: (session) => set({ pokerSession: session }),
}));
