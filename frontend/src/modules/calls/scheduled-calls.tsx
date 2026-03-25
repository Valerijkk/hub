import { useCallback, useEffect, useState } from "react";

import type {
	CreateScheduledCallData,
	ScheduledCallResponse,
} from "@/api/calls";
import * as callsApi from "@/api/calls";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

// ---------------------------------------------------------------------------
// ScheduledCalls -- запланированные и повторяющиеся звонки (CALL-02)
// ---------------------------------------------------------------------------

/** Пропсы компонента ScheduledCalls. */
interface ScheduledCallsProps {
	/** ID канала для загрузки запланированных звонков. */
	channelId: string;
	/** ID проекта для создания запланированных звонков. */
	projectId: string;
}

/** Предустановленные варианты повторения. */
const RECURRENCE_OPTIONS = [
	{ label: "Без повторения", value: "" },
	{
		label: "Ежедневно",
		value: "FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
	},
	{
		label: "Еженедельно",
		value: "FREQ=WEEKLY;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
	},
	{
		label: "По будням",
		value: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
	},
] as const;

/**
 * ScheduledCalls отображает список запланированных звонков в канале
 * и форму для создания нового запланированного звонка с поддержкой повторений.
 */
export function ScheduledCalls({ channelId, projectId }: ScheduledCallsProps) {
	const [calls, setCalls] = useState<ScheduledCallResponse[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [showForm, setShowForm] = useState(false);

	/** Загрузить запланированные звонки. */
	const loadCalls = useCallback(async () => {
		setIsLoading(true);
		try {
			const data = await callsApi.listScheduledCalls(channelId);
			setCalls(data.scheduled_calls ?? []);
		} catch {
			// Ошибка загрузки
		} finally {
			setIsLoading(false);
		}
	}, [channelId]);

	useEffect(() => {
		loadCalls();
	}, [loadCalls]);

	/** Удалить запланированный звонок. */
	const handleDelete = useCallback(
		async (id: string) => {
			try {
				await callsApi.deleteScheduledCall(id);
				await loadCalls();
			} catch {
				// Ошибка удаления
			}
		},
		[loadCalls],
	);

	/** Получить человекочитаемое описание RRULE. */
	const getRecurrenceLabel = (rrule?: string) => {
		if (!rrule) return "Одноразовый";
		if (rrule.includes("BYDAY=MO,TU,WE,TH,FR")) return "По будням";
		if (rrule.includes("FREQ=DAILY")) return "Ежедневно";
		if (rrule.includes("FREQ=WEEKLY")) return "Еженедельно";
		return "Повторяющийся";
	};

	return (
		<div className="flex flex-col gap-3">
			{/* Заголовок с кнопкой создания */}
			<div className="flex items-center justify-between">
				<h3 className="flex items-center gap-2 text-sm font-medium">
					<CalendarIcon />
					Запланированные звонки
				</h3>
				<button
					onClick={() => setShowForm(!showForm)}
					className="rounded-md bg-primary px-3 py-1 text-xs text-primary-foreground transition-colors hover:bg-primary/90"
				>
					{showForm ? "Отмена" : "Запланировать"}
				</button>
			</div>

			{/* Форма создания */}
			{showForm && (
				<CreateScheduledCallForm
					channelId={channelId}
					projectId={projectId}
					onCreated={() => {
						setShowForm(false);
						loadCalls();
					}}
				/>
			)}

			{/* Список */}
			{isLoading ? (
				<div className="py-4 text-center text-sm text-muted-foreground">
					Загрузка...
				</div>
			) : calls.length === 0 ? (
				<div className="py-4 text-center text-sm text-muted-foreground">
					Нет запланированных звонков
				</div>
			) : (
				<div className="flex flex-col gap-2">
					{calls.map((sc) => (
						<div
							key={sc.id}
							className="flex items-center justify-between rounded-lg border border-border p-3"
						>
							<div className="flex flex-col gap-0.5">
								<span className="text-sm font-medium">{sc.title}</span>
								<span className="text-xs text-muted-foreground">
									{new Date(sc.scheduled_at).toLocaleString("ru-RU")} --{" "}
									{getRecurrenceLabel(sc.rrule)}
								</span>
								{sc.description && (
									<span className="text-xs text-muted-foreground">
										{sc.description}
									</span>
								)}
							</div>
							<button
								onClick={() => handleDelete(sc.id)}
								className="rounded-md border border-border px-2 py-1 text-xs text-muted-foreground transition-colors hover:border-destructive hover:text-destructive"
								title="Удалить"
								aria-label={`Удалить запланированный звонок "${sc.title}"`}
							>
								Удалить
							</button>
						</div>
					))}
				</div>
			)}
		</div>
	);
}

// ---------------------------------------------------------------------------
// CreateScheduledCallForm -- форма создания запланированного звонка
// ---------------------------------------------------------------------------

interface CreateScheduledCallFormProps {
	channelId: string;
	projectId: string;
	onCreated: () => void;
}

function CreateScheduledCallForm({
	channelId,
	projectId,
	onCreated,
}: CreateScheduledCallFormProps) {
	const [title, setTitle] = useState("");
	const [description, setDescription] = useState("");
	const [scheduledAt, setScheduledAt] = useState("");
	const [recurrence, setRecurrence] = useState("");
	const [customRRule, setCustomRRule] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim() || !scheduledAt) return;

		setIsSubmitting(true);
		try {
			const rrule = recurrence === "custom" ? customRRule : recurrence;

			const data: CreateScheduledCallData = {
				title: title.trim(),
				scheduled_at: new Date(scheduledAt).toISOString(),
				channel_id: channelId,
			};
			if (description.trim()) data.description = description.trim();
			if (rrule) data.rrule = rrule;

			await callsApi.createScheduledCall(projectId, data);
			onCreated();
		} catch {
			// Ошибка создания
		} finally {
			setIsSubmitting(false);
		}
	};

	return (
		<form
			onSubmit={handleSubmit}
			className="flex flex-col gap-3 rounded-lg border border-border p-3"
		>
			<input
				type="text"
				placeholder="Название звонка"
				value={title}
				onChange={(e) => setTitle(e.target.value)}
				required
				maxLength={255}
				className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
				aria-label="Название запланированного звонка"
			/>

			<textarea
				placeholder="Описание (необязательно)"
				value={description}
				onChange={(e) => setDescription(e.target.value)}
				maxLength={2000}
				rows={2}
				className="resize-none rounded-md border border-border bg-background px-3 py-1.5 text-sm"
				aria-label="Описание запланированного звонка"
			/>

			<input
				type="datetime-local"
				value={scheduledAt}
				onChange={(e) => setScheduledAt(e.target.value)}
				required
				className="rounded-md border border-border bg-background px-3 py-1.5 text-sm"
				aria-label="Дата и время звонка"
			/>

			<Select
				value={recurrence || "no_repeat"}
				onValueChange={(v) => setRecurrence(v === "no_repeat" ? "" : v)}
			>
				<SelectTrigger className="rounded-xl border border-border bg-background px-3 py-1.5 text-sm" aria-label="Повторение">
					<SelectValue />
				</SelectTrigger>
				<SelectContent>
					{RECURRENCE_OPTIONS.map((opt) => (
						<SelectItem key={opt.value || "no_repeat"} value={opt.value || "no_repeat"}>
							{opt.label}
						</SelectItem>
					))}
					<SelectItem value="custom">Custom RRULE</SelectItem>
				</SelectContent>
			</Select>

			{recurrence === "custom" && (
				<input
					type="text"
					placeholder="FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0"
					value={customRRule}
					onChange={(e) => setCustomRRule(e.target.value)}
					className="rounded-md border border-border bg-background px-3 py-1.5 text-sm font-mono"
					aria-label="Custom RRULE"
				/>
			)}

			<button
				type="submit"
				disabled={isSubmitting || !title.trim() || !scheduledAt}
				className="rounded-md bg-primary px-4 py-1.5 text-sm text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
			>
				{isSubmitting ? "Создание..." : "Создать"}
			</button>
		</form>
	);
}

// ---------------------------------------------------------------------------
// CalendarIcon -- SVG иконка календаря
// ---------------------------------------------------------------------------

function CalendarIcon() {
	return (
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			strokeWidth="2"
			strokeLinecap="round"
			strokeLinejoin="round"
			aria-hidden="true"
		>
			<rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
			<line x1="16" y1="2" x2="16" y2="6" />
			<line x1="8" y1="2" x2="8" y2="6" />
			<line x1="3" y1="10" x2="21" y2="10" />
		</svg>
	);
}
