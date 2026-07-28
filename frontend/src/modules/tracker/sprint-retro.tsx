import {
	AlertTriangle,
	CheckCircle2,
	Lightbulb,
	Save,
	ThumbsUp,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { toastApiError } from "@/lib/format-api-error";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента ретроспективы спринта. */
interface SprintRetroProps {
	/** ID проекта. */
	projectId: string;
	/** ID спринта. */
	sprintId: string;
	/** Начальные данные ретроспективы. */
	initialNotes?: {
		went_well?: string[];
		to_improve?: string[];
		actions?: string[];
	};
}

/**
 * SprintRetro предоставляет форму ретроспективы спринта (SCRM-03).
 * Три секции: "Что прошло хорошо", "Что можно улучшить", "Действия".
 * Сохраняет как JSONB через PUT /sprints/:id/retro.
 */
export function SprintRetro({
	projectId,
	sprintId,
	initialNotes,
}: SprintRetroProps) {
	const { saveRetroNotes } = useSprintStore();

	const [wentWell, setWentWell] = useState(
		initialNotes?.went_well?.join("\n") ?? "",
	);
	const [toImprove, setToImprove] = useState(
		initialNotes?.to_improve?.join("\n") ?? "",
	);
	const [actions, setActions] = useState(
		initialNotes?.actions?.join("\n") ?? "",
	);
	const [saving, setSaving] = useState(false);
	const [saved, setSaved] = useState(false);
	const savedTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	useEffect(
		() => () => {
			if (savedTimerRef.current) clearTimeout(savedTimerRef.current);
		},
		[],
	);

	const handleSave = async () => {
		setSaving(true);
		try {
			await saveRetroNotes(projectId, sprintId, {
				went_well: wentWell
					.split("\n")
					.map((s) => s.trim())
					.filter(Boolean),
				to_improve: toImprove
					.split("\n")
					.map((s) => s.trim())
					.filter(Boolean),
				actions: actions
					.split("\n")
					.map((s) => s.trim())
					.filter(Boolean),
			});
			setSaved(true);
			toast.success("Ретроспектива сохранена");
			if (savedTimerRef.current) clearTimeout(savedTimerRef.current);
			savedTimerRef.current = setTimeout(() => setSaved(false), 2000);
		} catch (err) {
			toastApiError(err, "Не удалось сохранить ретроспективу");
		} finally {
			setSaving(false);
		}
	};

	const sections = [
		{
			key: "went_well",
			label: "Что прошло хорошо",
			icon: ThumbsUp,
			color: "text-emerald-400",
			bg: "from-emerald-500/10 to-emerald-500/5",
			borderColor: "border-emerald-500/20 focus-within:border-emerald-500/40",
			value: wentWell,
			onChange: setWentWell,
		},
		{
			key: "to_improve",
			label: "Что можно улучшить",
			icon: AlertTriangle,
			color: "text-amber-400",
			bg: "from-amber-500/10 to-amber-500/5",
			borderColor: "border-amber-500/20 focus-within:border-amber-500/40",
			value: toImprove,
			onChange: setToImprove,
		},
		{
			key: "actions",
			label: "Действия",
			icon: Lightbulb,
			color: "text-blue-400",
			bg: "from-blue-500/10 to-blue-500/5",
			borderColor: "border-blue-500/20 focus-within:border-blue-500/40",
			value: actions,
			onChange: setActions,
		},
	] as const;

	return (
		<div className="bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-6">
			<div className="flex items-center gap-2.5 mb-5">
				<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-500/10 to-violet-500/5">
					<Lightbulb className="h-4 w-4 text-violet-400" />
				</div>
				<h2 className="text-[14px] font-bold text-t1">Ретроспектива</h2>
			</div>

			<div className="grid grid-cols-1 gap-4 md:grid-cols-3">
				{sections.map((section) => {
					const SectionIcon = section.icon;
					return (
						<div key={section.key} className="space-y-2">
							<div className="flex items-center gap-2">
								<div
									className={`flex h-6 w-6 items-center justify-center rounded-md bg-gradient-to-br ${section.bg}`}
								>
									<SectionIcon className={`h-3 w-3 ${section.color}`} />
								</div>
								<label
									htmlFor={`retro-${section.key}`}
									className="text-[13px] font-medium text-t1"
								>
									{section.label}
								</label>
							</div>
							<Textarea
								id={`retro-${section.key}`}
								value={section.value}
								onChange={(e) => section.onChange(e.target.value)}
								placeholder="Каждый пункт с новой строки..."
								rows={5}
								className={`rounded-[var(--r3)] border bg-input text-[13px] text-t2 transition-colors ${section.borderColor}`}
							/>
						</div>
					);
				})}
			</div>

			<div className="mt-5 flex items-center gap-3">
				<Button
					onClick={handleSave}
					disabled={saving}
					className="btn-glow bg-gradient-to-r from-blue-600 to-blue-500 text-white shadow-md shadow-blue-500/20 hover:from-blue-500 hover:to-blue-400"
				>
					<Save className="mr-1.5 h-3.5 w-3.5" />
					{saving ? "Сохранение..." : "Сохранить"}
				</Button>
				{saved && (
					<span className="inline-flex items-center gap-1 text-sm text-emerald-400">
						<CheckCircle2 className="h-3.5 w-3.5" />
						Сохранено
					</span>
				)}
			</div>
		</div>
	);
}
