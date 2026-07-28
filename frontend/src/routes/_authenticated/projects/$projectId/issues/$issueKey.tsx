import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { getProject } from "@/api/projects";
import {
	deleteIssue,
	getIssue,
	listWorkflowStatuses,
	updateIssue,
} from "@/api/tracker";
import { ActivityFeed } from "@/components/tracker/activity-feed";
import { Checklist } from "@/components/tracker/checklist";
import { CommentList } from "@/components/tracker/comment-list";
import { IssueAttachments } from "@/components/tracker/issue-attachments";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toastApiError } from "@/lib/format-api-error";
import { canMutateProject } from "@/lib/permissions";
import {
	IssueDescription,
	IssueTitle,
} from "@/modules/tracker/issue-description";
import { IssueHeader } from "@/modules/tracker/issue-header";
import { IssueMessageLinksPanel } from "@/modules/tracker/issue-message-links-panel";
import { IssueSidebarMeta } from "@/modules/tracker/issue-sidebar-meta";

/**
 * Страница детальной информации о задаче.
 * Двухколоночная раскладка с inline-редактированием, комментариями,
 * лентой активности и трекером времени.
 */
function IssueDetailPage() {
	const { projectId, issueKey } = Route.useParams();
	const navigate = useNavigate();
	const queryClient = useQueryClient();

	const [showDeleteDialog, setShowDeleteDialog] = useState(false);

	const {
		data: issue,
		isLoading,
		isError,
		refetch,
	} = useQuery({
		queryKey: ["issue", projectId, issueKey],
		queryFn: () => getIssue(projectId, issueKey),
	});

	const { data: project } = useQuery({
		queryKey: ["project", projectId],
		queryFn: () => getProject(projectId),
	});
	const canDelete = canMutateProject(project?.my_role);

	const { data: statuses } = useQuery({
		queryKey: ["workflow-statuses", projectId],
		queryFn: () => listWorkflowStatuses(projectId),
	});

	const updateMutation = useMutation({
		mutationFn: (fields: Record<string, unknown>) => {
			if (!issue) throw new Error("No issue data");
			return updateIssue(projectId, issueKey, {
				version: issue.version,
				...fields,
			} as Parameters<typeof updateIssue>[2]);
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["issue", projectId, issueKey],
			});
			void queryClient.invalidateQueries({
				queryKey: ["board", projectId],
			});
			toast.success("Задача обновлена");
		},
		onError: (err) => toastApiError(err, "Не удалось обновить задачу"),
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteIssue(projectId, issueKey),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["issues", projectId],
			});
			void queryClient.invalidateQueries({
				queryKey: ["board", projectId],
			});
			toast.success("Задача удалена");
			void navigate({
				to: "/projects/$projectId/issues",
				params: { projectId },
			});
		},
		onError: (err) => toastApiError(err, "Не удалось удалить задачу"),
	});

	if (isLoading) {
		return (
			<div className="animate-[eu_450ms_var(--spring)_both]">
				{/* Skeleton хлебных крошек */}
				<div className="mb-4 flex items-center gap-2">
					<div className="h-4 w-16 rounded bg-card/50 animate-shimmer" />
					<div className="h-4 w-4 rounded bg-card/50 animate-shimmer" />
					<div className="h-4 w-16 rounded bg-card/50 animate-shimmer" />
					<div className="h-4 w-4 rounded bg-card/50 animate-shimmer" />
					<div className="h-4 w-24 rounded bg-card/50 animate-shimmer" />
				</div>
				{/* Skeleton заголовка */}
				<div className="mb-6 flex items-center gap-3">
					<div className="h-8 w-8 rounded-lg bg-card/50 animate-shimmer" />
					<div className="h-6 w-32 rounded bg-card/50 animate-shimmer" />
				</div>
				{/* Skeleton двухколоночной раскладки */}
				<div className="grid gap-6 lg:grid-cols-[1fr_320px]">
					<div className="space-y-4">
						<div className="h-8 w-3/4 rounded bg-card/50 animate-shimmer" />
						<div className="h-24 w-full rounded-lg bg-card/50 animate-shimmer" />
						<div className="h-px bg-border" />
						<div className="h-10 w-64 rounded-lg bg-card/50 animate-shimmer" />
						<div className="space-y-3">
							<div className="h-16 w-full rounded-lg bg-card/50 animate-shimmer" />
							<div className="h-16 w-full rounded-lg bg-card/50 animate-shimmer" />
						</div>
					</div>
					<div className="space-y-3">
						<div className="h-10 w-full rounded-lg bg-card/50 animate-shimmer" />
						<div className="h-10 w-full rounded-lg bg-card/50 animate-shimmer" />
						<div className="h-10 w-full rounded-lg bg-card/50 animate-shimmer" />
						<div className="h-10 w-full rounded-lg bg-card/50 animate-shimmer" />
						<div className="h-10 w-full rounded-lg bg-card/50 animate-shimmer" />
					</div>
				</div>
			</div>
		);
	}

	if (isError) {
		return (
			<div
				role="alert"
				className="flex flex-col items-center justify-center gap-3 py-12 text-center"
			>
				<p className="text-[14px] font-semibold text-t1">
					Не удалось загрузить задачу
				</p>
				<button
					type="button"
					onClick={() => void refetch()}
					className="h-9 px-3 rounded-[var(--r2)] border border-b1 text-[13px] text-t2 hover:text-t1 hover:border-b2 transition-colors"
				>
					Попробовать снова
				</button>
			</div>
		);
	}

	if (!issue) {
		return (
			<div className="flex items-center justify-center py-12">
				<p className="text-t2">Задача не найдена</p>
			</div>
		);
	}

	return (
		<div className="animate-[eu_450ms_var(--spring)_both]">
			<AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Удалить задачу?</AlertDialogTitle>
						<AlertDialogDescription>
							Задача {issue.issue_key} будет перемещена в корзину на 30 дней.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Отмена</AlertDialogCancel>
						<AlertDialogAction
							onClick={() => {
								deleteMutation.mutate();
								setShowDeleteDialog(false);
							}}
						>
							Удалить
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>

			<IssueHeader
				projectId={projectId}
				issueKey={issue.issue_key}
				issueType={issue.issue_type}
				onDelete={() => setShowDeleteDialog(true)}
				isDeleting={deleteMutation.isPending}
				canDelete={canDelete}
			/>

			{/* Двухколоночная раскладка */}
			<div className="grid gap-6 lg:grid-cols-[1fr_320px]">
				{/* Левая колонка: основное */}
				<div className="space-y-6 overflow-y-auto pr-2">
					<IssueTitle
						title={issue.title}
						onSave={(newTitle) => updateMutation.mutate({ title: newTitle })}
					/>

					<IssueDescription
						descriptionJson={issue.description_json}
						descriptionText={issue.description_text}
						issueKey={issue.issue_key}
						issueTitle={issue.title}
						onSave={(value) =>
							updateMutation.mutate({
								description_json: value as Record<string, unknown> | undefined,
							})
						}
					/>

					<Separator />

					{/* Чеклист */}
					<Checklist
						items={issue.checklist ?? []}
						onChange={(checklist) => updateMutation.mutate({ checklist })}
						disabled={updateMutation.isPending}
					/>

					<Separator />

					{/* Вложения — список локально после upload (BE IssueResponse без attachments) */}
					<IssueAttachments
						projectId={projectId}
						issueId={issue.id}
						onAttachmentAdded={() =>
							queryClient.invalidateQueries({
								queryKey: ["issue", projectId, issueKey],
							})
						}
					/>

					<Separator />

					{/* Вкладки: Комментарии | Активность | Время */}
					<Tabs defaultValue="comments">
						<TabsList>
							<TabsTrigger value="comments">Комментарии</TabsTrigger>
							<TabsTrigger value="activity">Активность</TabsTrigger>
						</TabsList>

						<TabsContent value="comments" className="mt-4">
							<CommentList projectId={projectId} issueKey={issueKey} />
						</TabsContent>

						<TabsContent value="activity" className="mt-4">
							<ActivityFeed projectId={projectId} issueKey={issueKey} />
						</TabsContent>
					</Tabs>
				</div>

				{/* Правая колонка: свойства */}
				<div className="w-[320px] min-w-[320px] border-l border-b1 overflow-y-auto p-[22px] bg-void space-y-6">
					<IssueMessageLinksPanel projectId={projectId} issueKey={issueKey} />
					<IssueSidebarMeta
						projectId={projectId}
						issue={issue}
						statuses={statuses}
						onUpdateField={(fields) => updateMutation.mutate(fields)}
					/>
				</div>
			</div>
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/issues/$issueKey",
)({
	component: IssueDetailPage,
});
