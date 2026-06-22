import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { Value } from "@udecode/plate";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	type CreateIssueRequest,
	createIssue,
	type IssuePriority,
	type IssueResponse,
	type IssueType,
	updateIssue,
} from "@/api/tracker";
import {
	type IssueFormValues,
	issueFormSchema,
} from "@/components/tracker/issue-form-schema";
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
import { toastApiError } from "@/lib/format-api-error";

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

/** Форма создания и редактирования задачи (RHF + Zod, FB-11). */
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
	const [description, setDescription] = useState<Value | undefined>(
		initialData?.description_json as Value | undefined,
	);

	const {
		register,
		control,
		handleSubmit,
		formState: { errors, isValid, isSubmitting },
	} = useForm<IssueFormValues>({
		resolver: zodResolver(issueFormSchema),
		mode: "onChange",
		defaultValues: {
			title: initialData?.title ?? "",
			issue_type: initialData?.issue_type ?? "task",
			priority: initialData?.priority ?? "medium",
			assignee_id: initialData?.assignee?.id ?? "",
			estimate_minutes: initialData?.estimate_minutes?.toString() ?? "",
			story_points: initialData?.story_points?.toString() ?? "",
			due_date: initialData?.due_date ?? "",
		},
	});

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
		mutationFn: (values: IssueFormValues) => {
			if (!initialData) throw new Error("No initial data for edit");
			return updateIssue(projectId, initialData.issue_key, {
				title: values.title.trim(),
				description_json: description as Record<string, unknown> | undefined,
				priority: values.priority,
				assignee_id: values.assignee_id || undefined,
				estimate_minutes: values.estimate_minutes
					? Number(values.estimate_minutes)
					: undefined,
				story_points: values.story_points
					? Number(values.story_points)
					: undefined,
				due_date: values.due_date || undefined,
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

	const onSubmit = (values: IssueFormValues) => {
		if (mode === "create") {
			createMutation.mutate({
				title: values.title.trim(),
				issue_type: values.issue_type,
				priority: values.priority,
				description_json: description as Record<string, unknown> | undefined,
				assignee_id: values.assignee_id || undefined,
				estimate_minutes: values.estimate_minutes
					? Number(values.estimate_minutes)
					: undefined,
				story_points: values.story_points
					? Number(values.story_points)
					: undefined,
				due_date: values.due_date || undefined,
			});
		} else {
			updateMutation.mutate(values);
		}
	};

	const fieldError = (key: keyof IssueFormValues) => {
		const err = errors[key];
		if (!err?.message) return null;
		return t(err.message as string);
	};

	const isPending =
		isSubmitting || createMutation.isPending || updateMutation.isPending;

	return (
		<form
			onSubmit={(e) => void handleSubmit(onSubmit)(e)}
			className="space-y-6"
		>
			<div className="space-y-2">
				<Label htmlFor="title">{t("form.title")}</Label>
				<Input
					id="title"
					{...register("title")}
					placeholder={t("form.titlePlaceholder")}
					aria-invalid={errors.title ? true : undefined}
					aria-describedby={errors.title ? "title-error" : undefined}
					className={
						errors.title
							? "border-red-500 focus-visible:ring-red-500"
							: undefined
					}
				/>
				{errors.title ? (
					<p id="title-error" className="text-sm text-red-500" role="alert">
						{fieldError("title")}
					</p>
				) : null}
			</div>

			<div className="grid grid-cols-2 gap-4">
				<div className="space-y-2">
					<Label htmlFor="issue_type">{t("issueType")}</Label>
					<Controller
						name="issue_type"
						control={control}
						render={({ field }) => (
							<Select value={field.value} onValueChange={field.onChange}>
								<SelectTrigger
									id="issue_type"
									className="w-full rounded-xl border bg-base px-3 py-2 text-sm"
								>
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
						)}
					/>
				</div>
				<div className="space-y-2">
					<Label htmlFor="priority">{t("priority")}</Label>
					<Controller
						name="priority"
						control={control}
						render={({ field }) => (
							<Select value={field.value} onValueChange={field.onChange}>
								<SelectTrigger
									id="priority"
									className="w-full rounded-xl border bg-base px-3 py-2 text-sm"
								>
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
						)}
					/>
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
				<Controller
					name="assignee_id"
					control={control}
					render={({ field }) => (
						<UserPicker
							projectId={projectId}
							value={field.value ?? ""}
							onChange={field.onChange}
							placeholder={t("form.assigneePlaceholder")}
						/>
					)}
				/>
			</div>

			<div className="grid grid-cols-3 gap-4">
				<div className="space-y-2">
					<Label htmlFor="estimate">{t("form.estimateMinutes")}</Label>
					<Input
						id="estimate"
						type="number"
						min={0}
						{...register("estimate_minutes")}
						placeholder="0"
					/>
					{errors.estimate_minutes ? (
						<p className="text-sm text-red-500">
							{fieldError("estimate_minutes")}
						</p>
					) : null}
				</div>
				<div className="space-y-2">
					<Label htmlFor="story_points">{t("storyPoints")}</Label>
					<Controller
						name="story_points"
						control={control}
						render={({ field }) => (
							<Select
								value={field.value || "none"}
								onValueChange={(v) => field.onChange(v === "none" ? "" : v)}
							>
								<SelectTrigger
									id="story_points"
									className="w-full rounded-xl border bg-base px-3 py-2 text-sm"
								>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="none">
										{t("form.storyPointsNone")}
									</SelectItem>
									{[1, 2, 3, 5, 8, 13, 21].map((sp) => (
										<SelectItem key={sp} value={String(sp)}>
											{sp}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						)}
					/>
				</div>
				<div className="space-y-2">
					<Label htmlFor="due_date">{t("dueDate")}</Label>
					<Input id="due_date" type="date" {...register("due_date")} />
				</div>
			</div>

			<Button type="submit" disabled={isPending || !isValid} className="w-full">
				{isPending
					? t("common:btn.saving")
					: mode === "create"
						? t("form.submitCreate")
						: t("form.submitEdit")}
			</Button>
		</form>
	);
}
