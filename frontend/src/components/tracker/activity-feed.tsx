import { useQuery } from "@tanstack/react-query";
import {
	type ActivityLogResponse,
	listActivities,
	listWorkflowStatuses,
} from "@/api/tracker";
import { listMembers } from "@/api/projects";
import { cn } from "@/lib/utils";

interface ActivityFeedProps {
	projectId: string;
	issueKey: string;
}

export function ActivityFeed({ projectId, issueKey }: ActivityFeedProps) {
	const { data: activities, isLoading } = useQuery({
		queryKey: ["activities", projectId, issueKey],
		queryFn: () => listActivities(projectId, issueKey),
	});

	const { data: statuses } = useQuery({
		queryKey: ["workflow-statuses", projectId],
		queryFn: () => listWorkflowStatuses(projectId),
	});

	const { data: members } = useQuery({
		queryKey: ["project-members", projectId],
		queryFn: () => listMembers(projectId),
	});

	// Маппинг UUID статуса → имя
	const statusMap = new Map<string, string>();
	if (statuses) {
		for (const s of statuses) {
			statusMap.set(s.id, s.name);
		}
	}

	// Маппинг UUID пользователя → имя
	const userMap = new Map<string, string>();
	if (members) {
		for (const m of members) {
			userMap.set(m.user_id, m.name);
		}
	}

	if (isLoading) {
		return (
			<p className="text-sm text-muted-foreground">Загрузка активности...</p>
		);
	}

	if (!activities || activities.length === 0) {
		return <p className="text-sm text-muted-foreground">Нет активности</p>;
	}

	return (
		<div className="space-y-3">
			{activities.map((activity) => (
				<ActivityItem
					key={activity.id}
					activity={activity}
					statusMap={statusMap}
					userMap={userMap}
				/>
			))}
		</div>
	);
}

function ActivityItem({
	activity,
	statusMap,
	userMap,
}: {
	activity: ActivityLogResponse;
	statusMap: Map<string, string>;
	userMap: Map<string, string>;
}) {
	const timeAgo = getTimeAgo(activity.created_at);

	/** Резолвит значение — если это UUID статуса/пользователя, возвращает имя. */
	const resolveValue = (field: string, value: unknown): string => {
		const str = String(value ?? "—");
		if (field === "status" && statusMap.has(str)) {
			return statusMap.get(str)!;
		}
		if (field === "assignee" && userMap.has(str)) {
			return userMap.get(str)!;
		}
		return str;
	};

	/** Color dot based on action type. */
	const dotColor =
		activity.action === "created"
			? "bg-green"
			: activity.action === "status_changed"
				? "bg-blue"
				: activity.action === "commented"
					? "bg-purple"
					: "bg-t3";

	return (
		<div className="flex gap-3 py-3 border-b border-b1 text-[13px]">
			<div className={cn("w-[9px] h-[9px] rounded-full mt-[5px] shrink-0", dotColor)} />

			<div className="flex-1 text-t2">
				<div className="flex items-center gap-2">
					<span className="text-t1 font-semibold">{activity.user.name}</span>
					<span className="text-t2">
						{getActionLabel(activity.action)}
					</span>
				</div>

				{(activity.changes ?? []).length > 0 && (
					<div className="mt-1 space-y-1">
						{(activity.changes ?? []).map((change, idx) => (
							<div key={idx} className="flex items-center gap-2 text-[13px]">
								<span className="font-medium text-t3">
									{getFieldLabel(change.field)}:
								</span>
								{change.field === "status" ? (
									<>
										<span className="line-through text-t4">
											{resolveValue(change.field, change.old_value)}
										</span>
										<span className="text-t4">{"\u2192"}</span>
										<span className="text-t1 font-semibold">
											{resolveValue(change.field, change.new_value)}
										</span>
									</>
								) : (
									<span>
										<span className="line-through text-t4">
											{resolveValue(change.field, change.old_value)}
										</span>
										{" \u2192 "}
										<span className="text-t1 font-semibold">
											{resolveValue(change.field, change.new_value)}
										</span>
									</span>
								)}
							</div>
						))}
					</div>
				)}
				<div className="text-[10.5px] text-t4 mt-[3px]">{timeAgo}</div>
			</div>
		</div>
	);
}

function getActionLabel(action: string): string {
	const labels: Record<string, string> = {
		created: "создал задачу",
		updated: "обновил задачу",
		status_changed: "изменил статус",
		commented: "добавил комментарий",
		assigned: "назначил исполнителя",
		priority_changed: "изменил приоритет",
		deleted: "удалил",
	};
	return labels[action] ?? action;
}

function getFieldLabel(field: string): string {
	const labels: Record<string, string> = {
		status: "Статус",
		priority: "Приоритет",
		assignee: "Исполнитель",
		title: "Название",
		description: "Описание",
		due_date: "Срок",
		labels: "Метки",
	};
	return labels[field] ?? field;
}

function getTimeAgo(dateStr: string): string {
	const now = Date.now();
	const then = new Date(dateStr).getTime();
	const diff = Math.floor((now - then) / 1000);

	if (diff < 60) return "только что";
	if (diff < 3600) return `${Math.floor(diff / 60)} мин назад`;
	if (diff < 86400) return `${Math.floor(diff / 3600)} ч назад`;
	if (diff < 2592000) return `${Math.floor(diff / 86400)} дн назад`;
	return new Date(dateStr).toLocaleDateString("ru-RU");
}
