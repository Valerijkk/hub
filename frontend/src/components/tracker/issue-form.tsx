import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { Value } from "@udecode/plate";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { toastApiError } from "@/lib/format-api-error";
import {
	type CreateIssueRequest,
	createIssue,
	type IssuePriority,
	type IssueResponse,
	type IssueType,
	updateIssue,
} from "@/api/tracker";
import { PlateEditor } from "@/components/tracker/plate-editor";
import { UserPicker } from "@/components/tracker/user-picker";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

const ISSUE_TYPE_KEYS: IssueType[] = [
	"bug",
	"task",
	"story",
	"epic",
	"sub_task",
	"improvement",
	"tech_debt",
	"spike",
	"release",
];

const PRIORITY_KEYS: IssuePriority[] = [
	"critical",
	"high",
	"medium",
	"low",
	"trivial",
];

function issueTypeLabel(type: IssueType, t: (k: string) => string): string {
	if (type === "sub_task") return t("types.subtask");
	if (type === "tech_debt") return t("types.techDebt");
	return t(`types.${type}`);
}

/** Форма создания и редактирования задачи. */
export function IssueForm({
	projectId,
	onSuccess,
	initialData,
	mode,
}: {
	projectId: string;
	onSuccess?: (issue: IssueResponse) => void;
	initialData?: IssueResponse;
	mode: "create" | "edit";
}) {
	const { t } = useTranslation(["tracker", "common"]);
	const queryClient = useQueryClient();

	const [title, setTitle] = useState(initialData?.title ?? "");
	const [issueType, setIssueType] = useState<IssueType>(
		initialData?.issue_type ?? "task",
	);
	const [priority, setPriority] = useState<IssuePriority>(
		initialData?.priority ?? "medium",
	);
	const [description, setDescription] = useState<Value | undefined>(
		initialData?.description_json as Value | undefined,
	);
	const [assigneeId, setAssigneeId] = useState(initialData?.assignee?.id ?? "");
	const [estimateMinutes, setEstimateMinutes] = useState<string>(
		initialData?.estimate_minutes?.toString() ?? "",
	);
	const [storyPoints, setStoryPoints] = useState<string>(
		initialData?.story_points?.toString() ?? "",
	);
	const [dueDate, setDueDate] = useState(initialData?.due_date ?? "");
	const [titleError, setTitleError] = useState<string | null>(null);

	const createMutation = useMutation({
		mutationFn: (data: CreateIssueRequest) => createIssue(projectId, data),
		onSuccess: (result) => {
			void queryClient.invalidateQueries({ queryKey: ["board", projectId] });
			void queryClient.invalidateQueries({ queryKey: ["issues", projectId] });
			toast.success(t("form.createSuccess"));
			onSuccess?.(result);
		},
		onError: (err) => toastApiError(err, t("form.createError")),
	});

	const updateMutation = useMutation({
		mutationFn: () => {
			if (!initialData) throw new Error("No initial data for edit");
			return updateIssue(projectId, initialData.issue_key, {
				title,
				description_json: description as Record<string, unknown> | undefined,
				priority,
				assignee_id: assigneeId || undefined,
				estimate_minutes: estimateMinutes ? Number(estimateMinutes) : undefined,
				story_points: storyPoints ? Number(storyPoints) : undefined,
				due_date: dueDate || undefined,
				version: initialData.version,
			});
		},
		onSuccess: (result) => {
			void queryClient.invalidateQueries({
				queryKey: ["issue", projectId, initialData?.issue_key],
			});
			void queryClient.invalidateQueries({ queryKey: ["board", projectId] });
			toast.success(t("form.updateSuccess"));
			onSuccess?.(result);
		},
		onError: (err) => toastApiError(err, t("form.updateError")),
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim()) {
			setTitleError(t("form.titleRequired"));
			return;
		}
		setTitleError(null);

		if (mode === "create") {
			createMutation.mutate({
				title: title.trim(),
				issue_type: issueType,
				priority,
				description_json: description as Record<string, unknown> | undefined,
				assignee_id: assigneeId || undefined,
				estimate_minutes: estimateMinutes ? Number(estimateMinutes) : undefined,
				story_points: storyPoints ? Number(storyPoints) : undefined,
				due_date: dueDate || undefined,
			});
		} else {
			updateMutation.mutate();
		}
	};

	const isPending = createMutation.isPending || updateMutation.isPending;

	return (
		<form onSubmit={handleSubmit} className="space-y-6">
			<div className="space-y-2">
				<Label htmlFor="title">{t("form.title")}</Label>
				<Input
					id="title"
					value={title}
					onChange={(e) => {
						setTitle(e.target.value);
						if (titleError && e.target.value.trim()) setTitleError(null);
					}}
					placeholder={t("form.titlePlaceholder")}
					required
					aria-invalid={titleError ? true : undefined}
					aria-describedby={titleError ? "title-error" : undefined}
					className={titleError ? "border-red-500 focus-visible:ring-red-500" : undefined}
				/>
				{titleError ? (
					<p id="title-error" className="text-sm text-red-500" role="alert">
						{titleError}
					</p>
				) : null}
			</div>

			<div className="grid grid-cols-2 gap-4">
				<div className="space-y-2">
					<Label htmlFor="issue_type">{t("issueType")}</Label>
					<Select
						value={issueType}
						onValueChange={(v) => setIssueType(v as IssueType)}
					>
						<SelectTrigger className="w-full rounded-xl border bg-base px-3 py-2 text-sm">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{ISSUE_TYPE_KEYS.map((type) => (
								<SelectItem key={type} value={type}>
									{issueTypeLabel(type, t)}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</div>
				<div className="space-y-2">
					<Label htmlFor="priority">{t("priority")}</Label>
					<Select
						value={priority}
						onValueChange={(v) => setPriority(v as IssuePriority)}
					>
						<SelectTrigger className="w-full rounded-xl border bg-base px-3 py-2 text-sm">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{PRIORITY_KEYS.map((p) => (
								<SelectItem key={p} value={p}>
									{t(`priorities.${p}`)}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</div>
			</div>

			<div className="space-y-2">
				<Label>{t("description")}</Label>
				<PlateEditor
					value={description}
					onChange={setDescription}
					placeholder={t("form.descriptionPlaceholder")}
				/>
			</div>

			<div className="space-y-2">
				<Label htmlFor="assignee">{t("assignee")}</Label>
				<UserPicker
					projectId={projectId}
					value={assigneeId}
					onChange={setAssigneeId}
					placeholder={t("form.assigneePlaceholder")}
				/>
			</div>

			<div className="grid grid-cols-3 gap-4">
				<div className="space-y-2">
					<Label htmlFor="estimate">{t("form.estimateMinutes")}</Label>
					<Input
						id="estimate"
						type="number"
						min={0}
						value={estimateMinutes}
						onChange={(e) => setEstimateMinutes(e.target.value)}
						placeholder="0"
					/>
				</div>
				<div className="space-y-2">
					<Label htmlFor="story_points">{t("storyPoints")}</Label>
					<Select
						value={storyPoints || "none"}
						onValueChange={(v) => setStoryPoints(v === "none" ? "" : v)}
					>
						<SelectTrigger className="w-full rounded-xl border bg-base px-3 py-2 text-sm">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="none">{t("form.storyPointsNone")}</SelectItem>
							{[1, 2, 3, 5, 8, 13, 21].map((sp) => (
								<SelectItem key={sp} value={String(sp)}>
									{sp}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</div>
				<div className="space-y-2">
					<Label htmlFor="due_date">{t("dueDate")}</Label>
					<Input
						id="due_date"
						type="date"
						value={dueDate}
						onChange={(e) => setDueDate(e.target.value)}
					/>
				</div>
			</div>

			<Button type="submit" disabled={isPending} className="w-full">
				{isPending
					? t("common:btn.saving")
					: mode === "create"
						? t("form.submitCreate")
						: t("form.submitEdit")}
			</Button>
		</form>
	);
}
