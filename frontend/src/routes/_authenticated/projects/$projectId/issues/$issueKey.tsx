import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import {
	deleteIssue,
	getIssue,
	listWorkflowStatuses,
	updateIssue,
} from "@/api/tracker";
import { ActivityFeed } from "@/components/tracker/activity-feed";
import { Checklist } from "@/components/tracker/checklist";
import { CommentList } from "@/components/tracker/comment-list";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
	IssueDescription,
	IssueTitle,
} from "@/modules/tracker/issue-description";
import { IssueHeader } from "@/modules/tracker/issue-header";
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

	const { data: issue, isLoading } = useQuery({
		queryKey: ["issue", projectId, issueKey],
		queryFn: () => getIssue(projectId, issueKey),
	});

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
		onError: () => toast.error("Не удалось обновить задачу"),
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
		onError: () => toast.error("Не удалось удалить задачу"),
	});

	if (isLoading) {
		return (
			<div className="animate-fade-in">
				{/* Breadcrumb skeleton */}
				<div className="mb-4 flex items-center gap-2">
					<div className="h-4 w-16 rounded animate-shimmer" />
					<div className="h-4 w-4 rounded animate-shimmer" />
					<div className="h-4 w-16 rounded animate-shimmer" />
					<div className="h-4 w-4 rounded animate-shimmer" />
					<div className="h-4 w-24 rounded animate-shimmer" />
				</div>
				{/* Header skeleton */}
				<div className="mb-6 flex items-center gap-3">
					<div className="h-8 w-8 rounded-lg animate-shimmer" />
					<div className="h-6 w-32 rounded animate-shimmer" />
				</div>
				{/* Two-column skeleton */}
				<div className="grid gap-6 lg:grid-cols-[1fr_320px]">
					<div className="space-y-4">
						<div className="h-8 w-3/4 rounded animate-shimmer" />
						<div className="h-24 w-full rounded-lg animate-shimmer" />
						<div className="h-px bg-border" />
						<div className="h-10 w-64 rounded-lg animate-shimmer" />
						<div className="space-y-3">
							<div className="h-16 w-full rounded-lg animate-shimmer" />
							<div className="h-16 w-full rounded-lg animate-shimmer" />
						</div>
					</div>
					<div className="space-y-3">
						<div className="h-10 w-full rounded-lg animate-shimmer" />
						<div className="h-10 w-full rounded-lg animate-shimmer" />
						<div className="h-10 w-full rounded-lg animate-shimmer" />
						<div className="h-10 w-full rounded-lg animate-shimmer" />
						<div className="h-10 w-full rounded-lg animate-shimmer" />
					</div>
				</div>
			</div>
		);
	}

	if (!issue) {
		return (
			<div className="flex items-center justify-center py-12">
				<p className="text-muted-foreground">Задача не найдена</p>
			</div>
		);
	}

	return (
		<div className="animate-fade-in">
			<IssueHeader
				projectId={projectId}
				issueKey={issue.issue_key}
				issueType={issue.issue_type}
				onDelete={() => {
					if (window.confirm("Удалить задачу?")) {
						deleteMutation.mutate();
					}
				}}
				isDeleting={deleteMutation.isPending}
			/>

			{/* Двухколоночная раскладка */}
			<div className="grid gap-6 lg:grid-cols-[1fr_320px]">
				{/* Левая колонка: основное */}
				<div className="space-y-6">
					<IssueTitle
						title={issue.title}
						onSave={(newTitle) => updateMutation.mutate({ title: newTitle })}
					/>

					<IssueDescription
						descriptionJson={issue.description_json}
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
				<IssueSidebarMeta
					projectId={projectId}
					issue={issue}
					statuses={statuses}
					onUpdateField={(fields) => updateMutation.mutate(fields)}
				/>
			</div>
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/issues/$issueKey",
)({
	component: IssueDetailPage,
});
