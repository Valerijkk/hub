import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { type CreateScheduledCallData, createScheduledCall } from "@/api/calls";
import { getChannels } from "@/api/messenger";
import { listProjects } from "@/api/projects";
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

// ---------------------------------------------------------------------------
// Duration presets
// ---------------------------------------------------------------------------

const DURATION_OPTIONS = [
	{ label: "30 мин", value: "30" },
	{ label: "1 час", value: "60" },
	{ label: "1.5 часа", value: "90" },
	{ label: "2 часа", value: "120" },
] as const;

// ---------------------------------------------------------------------------
// Recurrence presets
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
// Props
// ---------------------------------------------------------------------------

interface CreateCallDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Pre-fill date (YYYY-MM-DD). */
	defaultDate?: string;
	/** Pre-fill time (HH:MM). */
	defaultTime?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

/**
 * CreateCallDialog -- simplified quick-create dialog for scheduling calls.
 * Pre-fills selected date, uses quick-select buttons for duration,
 * and clean minimal dark design inspired by Tinkoff Talk.
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
	const [recurrence, setRecurrence] = useState("");
	const [channelId, setChannelId] = useState("");

	const { data: channels } = useQuery({
		queryKey: ["messenger", "channels"],
		queryFn: getChannels,
	});

	const { data: projects } = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	const projectId = (projects as { id: string }[] | undefined)?.[0]?.id ?? "";

	// Reset form when dialog opens with defaults
	useEffect(() => {
		if (open) {
			setTitle("");
			setDate(defaultDate ?? "");
			setTime(defaultTime ?? "10:00");
			setDuration("60");
			setRecurrence("");
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
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim() || !date || !time || !channelId) return;

		const scheduledAt = new Date(`${date}T${time}:00`).toISOString();

		const data: CreateScheduledCallData = {
			title: title.trim(),
			scheduled_at: scheduledAt,
			channel_id: channelId,
		};

		if (recurrence) data.rrule = recurrence;

		mutation.mutate(data);
	};

	const isValid = title.trim().length > 0 && date && time && channelId;

	// Format the selected date for display
	const dateDisplay = date
		? new Date(date + "T00:00:00").toLocaleDateString("ru-RU", {
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
							<p className="text-sm text-muted-foreground capitalize mt-1">
								{dateDisplay}
							</p>
						)}
					</DialogHeader>

					<div className="px-6 space-y-5">
						{/* Title */}
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

						{/* Channel */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">Канал</Label>
							<Select value={channelId} onValueChange={setChannelId}>
								<SelectTrigger className="h-11 rounded-xl">
									<SelectValue placeholder="Выберите канал" />
								</SelectTrigger>
								<SelectContent>
									{(channels ?? []).map((ch) => (
										<SelectItem key={ch.id} value={ch.id}>
											{ch.name || ch.id.slice(0, 8)}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						{/* Time */}
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
								{/* Hidden date input -- always present for form submission */}
								<Input
									type="date"
									value={date}
									onChange={(e) => setDate(e.target.value)}
									required
									className="h-11 flex-1"
								/>
							</div>
						</div>

						{/* Duration -- quick select buttons */}
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
												: "border-border bg-background text-muted-foreground hover:border-primary/50 hover:text-foreground"
										}`}
										onClick={() => setDuration(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>

						{/* Recurrence */}
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
												: "border-border bg-background text-muted-foreground hover:border-primary/50 hover:text-foreground"
										}`}
										onClick={() => setRecurrence(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>
					</div>

					{/* Footer */}
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
