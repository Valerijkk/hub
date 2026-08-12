import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Check, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	type CreateScheduledCallData,
	checkCallConflicts,
	createScheduledCall,
} from "@/api/calls";
import { getChannels } from "@/api/messenger";
import {
	listMembers,
	listProjects,
	type MemberResponse,
	type ProjectResponse,
} from "@/api/projects";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { toastApiError } from "@/lib/format-api-error";

// ---------------------------------------------------------------------------
// Пресеты длительности
// ---------------------------------------------------------------------------

const DURATION_OPTIONS = [
	{ label: "30 мин", value: "30" },
	{ label: "1 час", value: "60" },
	{ label: "1.5 часа", value: "90" },
	{ label: "2 часа", value: "120" },
] as const;

// ---------------------------------------------------------------------------
// Пресеты повторения
// ---------------------------------------------------------------------------

const RECURRENCE_OPTIONS = [
	{ label: "Нет", value: "" },
	{ label: "Каждый день", value: "FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0" },
	{
		label: "Каждую неделю",
		value: "FREQ=WEEKLY;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
	},
	{
		label: "По будням",
		value: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;BYHOUR=10;BYMINUTE=0;BYSECOND=0",
	},
] as const;

// ---------------------------------------------------------------------------
// Пропсы компонента
// ---------------------------------------------------------------------------

interface CreateCallDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Предзаполнение даты (YYYY-MM-DD). */
	defaultDate?: string;
	/** Предзаполнение времени (HH:MM). */
	defaultTime?: string;
}

// ---------------------------------------------------------------------------
// Компонент
// ---------------------------------------------------------------------------

/**
 * CreateCallDialog — быстрое создание запланированного звонка.
 * Предзаполняет дату, длительность — кнопками быстрого выбора,
 * минималистичный тёмный UI в духе Tinkoff Talk.
 */
export function CreateCallDialog({
	open,
	onOpenChange,
	defaultDate,
	defaultTime,
}: CreateCallDialogProps) {
	const { t } = useTranslation("calls");
	const queryClient = useQueryClient();

	const [title, setTitle] = useState("");
	const [date, setDate] = useState("");
	const [time, setTime] = useState("10:00");
	const [duration, setDuration] = useState("60");
	const [reminder, setReminder] = useState("15");
	const [recurrence, setRecurrence] = useState("");
	const [channelId, setChannelId] = useState("");
	const [projectId, setProjectId] = useState("");
	const [participantIds, setParticipantIds] = useState<string[]>([]);
	const [participantSearch, setParticipantSearch] = useState("");

	const {
		data: channels,
		isError: channelsError,
		refetch: refetchChannels,
	} = useQuery({
		queryKey: ["messenger", "channels"],
		queryFn: getChannels,
	});

	const {
		data: projects,
		isError: projectsError,
		refetch: refetchProjects,
	} = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	const projectList = (projects ?? []) as ProjectResponse[];
	const channelList = channels ?? [];
	const projectsEmpty = !projectsError && projectList.length === 0;
	const channelsEmpty = !channelsError && channelList.length === 0;

	// Если проект ещё не выбран — подставить первый из списка (можно сменить в UI).
	useEffect(() => {
		if (!projectId && projects && projects.length > 0) {
			setProjectId(projects[0]?.id ?? "");
		}
	}, [projectId, projects]);

	const { data: members } = useQuery({
		queryKey: ["project-members", projectId],
		queryFn: () => listMembers(projectId),
		enabled: !!projectId,
	});

	// --- Проверка конфликтов расписания ---
	const conflictQueryEnabled =
		!!date && !!time && participantIds.length > 0 && open;

	const scheduledAtISO = useMemo(() => {
		if (!date || !time) return "";
		try {
			return new Date(`${date}T${time}:00`).toISOString();
		} catch {
			return "";
		}
	}, [date, time]);

	const { data: conflictData } = useQuery({
		queryKey: ["calls", "conflicts", scheduledAtISO, duration, participantIds],
		queryFn: () =>
			checkCallConflicts(scheduledAtISO, Number(duration), participantIds),
		enabled: conflictQueryEnabled && !!scheduledAtISO,
		staleTime: 10_000,
	});

	const hasConflict = conflictData?.has_conflict ?? false;

	const filteredMembers = useMemo(() => {
		const all = (members ?? []) as MemberResponse[];
		if (!participantSearch.trim()) return all;
		const q = participantSearch.toLowerCase();
		return all.filter(
			(m) =>
				m.name.toLowerCase().includes(q) || m.email.toLowerCase().includes(q),
		);
	}, [members, participantSearch]);

	const toggleParticipant = (userId: string) => {
		setParticipantIds((prev) =>
			prev.includes(userId)
				? prev.filter((id) => id !== userId)
				: [...prev, userId],
		);
	};

	const removeParticipant = (userId: string) => {
		setParticipantIds((prev) => prev.filter((id) => id !== userId));
	};

	// Сбрасываем форму при открытии с переданными default-значениями
	useEffect(() => {
		if (open) {
			setTitle("");
			setDate(defaultDate ?? "");
			setTime(defaultTime ?? "10:00");
			setDuration("60");
			setReminder("15");
			setRecurrence("");
			setChannelId("");
			setProjectId("");
			setParticipantIds([]);
			setParticipantSearch("");
		}
	}, [open, defaultDate, defaultTime]);

	const mutation = useMutation({
		mutationFn: (data: CreateScheduledCallData) => {
			return createScheduledCall(projectId, data);
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["calls", "scheduled"] });
			onOpenChange(false);
		},
		onError: (err) => {
			toastApiError(err, "Не удалось создать звонок");
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim() || !date || !time || !channelId || !projectId) {
			if (!projectId) {
				toast.error("Выберите проект");
			}
			return;
		}

		const scheduledAt = new Date(`${date}T${time}:00`).toISOString();

		const data: CreateScheduledCallData = {
			title: title.trim(),
			scheduled_at: scheduledAt,
			duration_minutes: Number.parseInt(duration, 10) || 60,
			channel_id: channelId,
		};

		if (recurrence) data.rrule = recurrence;
		if (participantIds.length > 0) data.participant_ids = participantIds;
		if (reminder && reminder !== "0")
			data.reminder_minutes = Number.parseInt(reminder, 10);

		mutation.mutate(data);
	};

	const isValid =
		title.trim().length > 0 &&
		!!date &&
		!!time &&
		!!channelId &&
		!!projectId &&
		!hasConflict;

	// Форматируем выбранную дату для отображения в UI
	const dateDisplay = date
		? new Date(`${date}T00:00:00`).toLocaleDateString("ru-RU", {
				weekday: "short",
				day: "numeric",
				month: "long",
			})
		: null;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[420px] gap-0 p-0 overflow-hidden">
				<form onSubmit={handleSubmit}>
					<DialogHeader className="px-6 pt-6 pb-4">
						<DialogTitle className="text-lg">
							{t("createCallTitle", { defaultValue: "Новый звонок" })}
						</DialogTitle>
						{dateDisplay && (
							<p className="text-sm text-t2 capitalize mt-1">{dateDisplay}</p>
						)}
					</DialogHeader>

					<div className="px-6 space-y-5">
						{/* Название */}
						<div className="space-y-2">
							<Label htmlFor="call-title" className="text-sm font-medium">
								{t("callTitleLabel", { defaultValue: "Название" })}
							</Label>
							<Input
								id="call-title"
								placeholder={t("callTitlePlaceholder", {
									defaultValue: "Стендап, ретро, планирование...",
								})}
								value={title}
								onChange={(e) => setTitle(e.target.value)}
								maxLength={255}
								required
								autoFocus
								className="h-11"
							/>
						</div>

						{/* Проект — FE-MC-002: fail ≠ «нет проектов» */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">Проект</Label>
							{projectsError ? (
								<QueryErrorAlert
									compact
									message="Не удалось загрузить проекты"
									retry={() => void refetchProjects()}
								/>
							) : (
								<Select
									value={projectId}
									onValueChange={(v) => {
										setProjectId(v);
										setParticipantIds([]);
									}}
								>
									<SelectTrigger
										className="h-11 rounded-xl"
										aria-label="Выбор проекта"
									>
										<SelectValue placeholder="Выберите проект" />
									</SelectTrigger>
									<SelectContent>
										{projectsEmpty ? (
											<SelectItem value="__none" disabled>
												Нет доступных проектов
											</SelectItem>
										) : (
											projectList.map((p) => (
												<SelectItem key={p.id} value={p.id}>
													{p.key ? `${p.key} — ${p.name}` : p.name}
												</SelectItem>
											))
										)}
									</SelectContent>
								</Select>
							)}
						</div>

						{/* Канал — FE-MC-002 / FE-MC-004 */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">Канал</Label>
							{channelsError ? (
								<QueryErrorAlert
									compact
									message="Не удалось загрузить каналы"
									retry={() => void refetchChannels()}
								/>
							) : (
								<Select value={channelId} onValueChange={setChannelId}>
									<SelectTrigger
										className="h-11 rounded-xl"
										aria-label="Выбор канала"
									>
										<SelectValue
											placeholder={
												channelsEmpty
													? "Нет доступных каналов"
													: "Выберите канал"
											}
										/>
									</SelectTrigger>
									<SelectContent>
										{channelsEmpty ? (
											<SelectItem value="__none" disabled>
												Нет доступных каналов
											</SelectItem>
										) : (
											channelList.map((ch) => (
												<SelectItem key={ch.id} value={ch.id}>
													{ch.name || ch.id.slice(0, 8)}
												</SelectItem>
											))
										)}
									</SelectContent>
								</Select>
							)}
						</div>

						{/* Участники */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callParticipantsLabel", { defaultValue: "Участники" })}
							</Label>

							{/* Чипы выбранных участников */}
							{participantIds.length > 0 && (
								<div className="flex flex-wrap gap-1.5">
									{participantIds.map((uid) => {
										const member = (members ?? []).find(
											(m: MemberResponse) => m.user_id === uid,
										);
										if (!member) return null;
										return (
											<span
												key={uid}
												className="inline-flex items-center gap-1 rounded-full border border-border bg-raised px-2.5 py-1 text-xs font-medium"
											>
												{member.avatar_url ? (
													<img
														src={member.avatar_url}
														alt=""
														className="h-4 w-4 rounded-full object-cover"
													/>
												) : (
													<span className="flex h-4 w-4 items-center justify-center rounded-full bg-primary/20 text-[10px] font-semibold text-primary">
														{member.name.charAt(0).toUpperCase()}
													</span>
												)}
												{member.name}
												<button
													type="button"
													onClick={() => removeParticipant(uid)}
													aria-label={`Удалить участника ${member.name}`}
													className="ml-0.5 rounded-full p-0.5 text-t2 hover:bg-destructive/20 hover:text-destructive transition-colors"
												>
													<X className="h-3 w-3" aria-hidden="true" />
												</button>
											</span>
										);
									})}
								</div>
							)}

							{/* Поле поиска участников */}
							<Input
								placeholder={t("searchParticipants", {
									defaultValue: "Поиск участников...",
								})}
								value={participantSearch}
								onChange={(e) => setParticipantSearch(e.target.value)}
								className="h-10"
							/>

							{/* Список участников */}
							<div className="max-h-32 overflow-y-auto rounded-lg border">
								{filteredMembers.length === 0 ? (
									<div className="px-3 py-2 text-xs text-t2">
										{t("noParticipantsFound", {
											defaultValue: "Участники не найдены",
										})}
									</div>
								) : (
									filteredMembers.map((member: MemberResponse) => {
										const isSelected = participantIds.includes(member.user_id);
										return (
											<button
												key={member.user_id}
												type="button"
												className={`flex w-full items-center gap-2.5 px-3 py-2 text-sm transition-colors hover:bg-raised/50 ${
													isSelected ? "bg-primary/5" : ""
												}`}
												onClick={() => toggleParticipant(member.user_id)}
											>
												{member.avatar_url ? (
													<img
														src={member.avatar_url}
														alt=""
														className="h-6 w-6 rounded-full object-cover"
													/>
												) : (
													<span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-xs font-semibold text-primary">
														{member.name.charAt(0).toUpperCase()}
													</span>
												)}
												<span className="flex-1 text-left truncate">
													{member.name}
												</span>
												{isSelected && (
													<Check className="h-4 w-4 text-primary shrink-0" />
												)}
											</button>
										);
									})
								)}
							</div>
						</div>

						{/* Время */}
						<div className="space-y-2">
							<Label htmlFor="call-time" className="text-sm font-medium">
								{t("callTimeLabel", { defaultValue: "Время начала" })}
							</Label>
							<div className="flex gap-3">
								<Input
									id="call-time"
									type="time"
									value={time}
									onChange={(e) => setTime(e.target.value)}
									required
									className="h-11 w-32"
								/>
								{/* Скрытый input даты — нужен для submit формы */}
								<Input
									type="date"
									value={date}
									onChange={(e) => setDate(e.target.value)}
									required
									className="h-11 flex-1"
								/>
							</div>
						</div>

						{/* Длительность — кнопки быстрого выбора */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callDurationLabel", { defaultValue: "Длительность" })}
							</Label>
							<div className="flex gap-2">
								{DURATION_OPTIONS.map((opt) => (
									<button
										key={opt.value}
										type="button"
										className={`flex-1 rounded-lg border px-3 py-2 text-sm font-medium transition-colors ${
											duration === opt.value
												? "border-primary bg-primary/10 text-primary"
												: "border-border bg-base text-t2 hover:border-primary/50 hover:text-t1"
										}`}
										onClick={() => setDuration(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>

						{/* Напоминание */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callReminderLabel", { defaultValue: "Напоминание" })}
							</Label>
							<div className="flex gap-2">
								{[
									{ label: "Нет", value: "0" },
									{ label: "5 мин", value: "5" },
									{ label: "15 мин", value: "15" },
									{ label: "30 мин", value: "30" },
									{ label: "1 час", value: "60" },
								].map((opt) => (
									<button
										key={opt.value}
										type="button"
										className={`flex-1 rounded-lg border px-2 py-2 text-xs font-medium transition-colors ${
											reminder === opt.value
												? "border-primary bg-primary/10 text-primary"
												: "border-border bg-base text-t2 hover:border-primary/50 hover:text-t1"
										}`}
										onClick={() => setReminder(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>

						{/* Повторение */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callRecurrenceLabel", { defaultValue: "Повтор" })}
							</Label>
							<div className="flex flex-wrap gap-2">
								{RECURRENCE_OPTIONS.map((opt) => (
									<button
										key={opt.value}
										type="button"
										className={`rounded-lg border px-3 py-2 text-sm font-medium transition-colors ${
											recurrence === opt.value
												? "border-primary bg-primary/10 text-primary"
												: "border-border bg-base text-t2 hover:border-primary/50 hover:text-t1"
										}`}
										onClick={() => setRecurrence(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>
					</div>

					{/* Предупреждение о конфликте */}
					{hasConflict && conflictData && (
						<div className="px-6 mt-3">
							<div className="flex items-start gap-2 rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2.5">
								<AlertTriangle className="h-4 w-4 shrink-0 text-amber-500 mt-0.5" />
								<div className="space-y-1">
									{conflictData.conflicts.map((c) => {
										const member = (members ?? []).find(
											(m: MemberResponse) => m.user_id === c.participant_id,
										);
										const name = member?.name ?? c.participant_id.slice(0, 8);
										return (
											<p
												key={`${c.participant_id}-${c.call_title}`}
												className="text-xs text-amber-200"
											>
												У участника <span className="font-medium">{name}</span>{" "}
												уже есть звонок в это время:{" "}
												<span className="font-medium">{c.call_title}</span>
											</p>
										);
									})}
								</div>
							</div>
						</div>
					)}

					{/* Футер диалога */}
					<div className="px-6 py-5 mt-2">
						<Button
							type="submit"
							disabled={!isValid || mutation.isPending}
							className="w-full h-11 text-base font-semibold"
							size="lg"
						>
							{mutation.isPending
								? t("creating", { defaultValue: "Создание..." })
								: t("create", { defaultValue: "Создать" })}
						</Button>
					</div>
				</form>
			</DialogContent>
		</Dialog>
	);
}
