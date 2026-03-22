import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { type CreateScheduledCallData, createScheduledCall } from "@/api/calls";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

// ---------------------------------------------------------------------------
// Recurrence presets
// ---------------------------------------------------------------------------

const RECURRENCE_OPTIONS = [
	{ label: "Без повторения", value: "" },
	{ label: "Ежедневно", value: "FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0" },
	{
		label: "Еженедельно",
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
 * CreateCallDialog -- modal dialog for creating a new scheduled call.
 * Uses shadcn Dialog, Input, Label, Button components.
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
	const [description, setDescription] = useState("");
	const [date, setDate] = useState("");
	const [time, setTime] = useState("10:00");
	const [duration, setDuration] = useState("60");
	const [recurrence, setRecurrence] = useState("");

	// Reset form when dialog opens with defaults
	useEffect(() => {
		if (open) {
			setTitle("");
			setDescription("");
			setDate(defaultDate ?? "");
			setTime(defaultTime ?? "10:00");
			setDuration("60");
			setRecurrence("");
		}
	}, [open, defaultDate, defaultTime]);

	const mutation = useMutation({
		mutationFn: (data: CreateScheduledCallData) => {
			// Use a placeholder projectId -- the actual project context
			// will be resolved by the route/layout that provides it.
			// For now we use "current" which the API proxy resolves.
			return createScheduledCall("current", data);
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["calls", "scheduled"] });
			onOpenChange(false);
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim() || !date || !time) return;

		const scheduledAt = new Date(`${date}T${time}:00`).toISOString();

		const data: CreateScheduledCallData = {
			title: title.trim(),
			scheduled_at: scheduledAt,
			channel_id: "", // Will be set by server or channel context
		};

		if (description.trim()) data.description = description.trim();
		if (recurrence) data.rrule = recurrence;

		mutation.mutate(data);
	};

	const isValid = title.trim().length > 0 && date && time;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<form onSubmit={handleSubmit}>
					<DialogHeader>
						<DialogTitle>
							{t("createCallTitle", { defaultValue: "Запланировать звонок" })}
						</DialogTitle>
						<DialogDescription>
							{t("createCallDesc", {
								defaultValue:
									"Создайте новый запланированный звонок или встречу",
							})}
						</DialogDescription>
					</DialogHeader>

					<div className="mt-4 space-y-4">
						{/* Title */}
						<div className="space-y-2">
							<Label htmlFor="call-title">
								{t("callTitleLabel", { defaultValue: "Название" })}
							</Label>
							<Input
								id="call-title"
								placeholder={t("callTitlePlaceholder", {
									defaultValue: "Ежедневный стендап",
								})}
								value={title}
								onChange={(e) => setTitle(e.target.value)}
								maxLength={255}
								required
							/>
						</div>

						{/* Description */}
						<div className="space-y-2">
							<Label htmlFor="call-desc">
								{t("callDescLabel", { defaultValue: "Описание" })}
							</Label>
							<Input
								id="call-desc"
								placeholder={t("callDescPlaceholder", {
									defaultValue: "Необязательно",
								})}
								value={description}
								onChange={(e) => setDescription(e.target.value)}
								maxLength={2000}
							/>
						</div>

						{/* Date & Time row */}
						<div className="grid grid-cols-2 gap-3">
							<div className="space-y-2">
								<Label htmlFor="call-date">
									{t("callDateLabel", { defaultValue: "Дата" })}
								</Label>
								<Input
									id="call-date"
									type="date"
									value={date}
									onChange={(e) => setDate(e.target.value)}
									required
								/>
							</div>
							<div className="space-y-2">
								<Label htmlFor="call-time">
									{t("callTimeLabel", { defaultValue: "Время" })}
								</Label>
								<Input
									id="call-time"
									type="time"
									value={time}
									onChange={(e) => setTime(e.target.value)}
									required
								/>
							</div>
						</div>

						{/* Duration */}
						<div className="space-y-2">
							<Label htmlFor="call-duration">
								{t("callDurationLabel", { defaultValue: "Длительность (мин)" })}
							</Label>
							<select
								id="call-duration"
								value={duration}
								onChange={(e) => setDuration(e.target.value)}
								className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
							>
								<option value="15">15 min</option>
								<option value="30">30 min</option>
								<option value="45">45 min</option>
								<option value="60">1 hour</option>
								<option value="90">1.5 hours</option>
								<option value="120">2 hours</option>
							</select>
						</div>

						{/* Recurrence */}
						<div className="space-y-2">
							<Label htmlFor="call-recurrence">
								{t("callRecurrenceLabel", { defaultValue: "Повторение" })}
							</Label>
							<select
								id="call-recurrence"
								value={recurrence}
								onChange={(e) => setRecurrence(e.target.value)}
								className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
							>
								{RECURRENCE_OPTIONS.map((opt) => (
									<option key={opt.value} value={opt.value}>
										{opt.label}
									</option>
								))}
							</select>
						</div>
					</div>

					<DialogFooter className="mt-6">
						<Button
							type="button"
							variant="outline"
							onClick={() => onOpenChange(false)}
						>
							{t("cancel", { defaultValue: "Отмена" })}
						</Button>
						<Button type="submit" disabled={!isValid || mutation.isPending}>
							{mutation.isPending
								? t("creating", { defaultValue: "Создание..." })
								: t("create", { defaultValue: "Создать" })}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
