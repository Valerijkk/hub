import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import {
	deleteIssueMessageLink,
	listIssueMessageLinks,
} from "@/api/issue-message-links";

interface IssueMessageLinksPanelProps {
	projectId: string;
	issueKey: string;
}

/** Связанные треды мессенджера в сайдбаре задачи (PROD-03). */
export function IssueMessageLinksPanel({
	projectId,
	issueKey,
}: IssueMessageLinksPanelProps) {
	const queryClient = useQueryClient();
	const { data: links = [] } = useQuery({
		queryKey: ["issue-message-links", projectId, issueKey],
		queryFn: () => listIssueMessageLinks(projectId, issueKey),
	});

	const unlinkMutation = useMutation({
		mutationFn: (linkId: string) =>
			deleteIssueMessageLink(projectId, issueKey, linkId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["issue-message-links", projectId, issueKey],
			});
		},
	});

	if (links.length === 0) {
		return null;
	}

	return (
		<div className="space-y-2">
			<h3 className="text-sm font-medium text-t2">Треды</h3>
			<ul className="space-y-1">
				{links.map((link) => (
					<li
						key={link.id}
						className="flex items-center justify-between gap-2 text-xs"
					>
						<Link
							to="/messenger/$channelId"
							params={{ channelId: link.channel_id }}
							className="text-primary hover:underline truncate"
						>
							Канал · тред …{link.message_id.slice(-6)}
						</Link>
						<button
							type="button"
							className="text-t4 hover:text-red-500 shrink-0"
							onClick={() => unlinkMutation.mutate(link.id)}
						>
							Отвязать
						</button>
					</li>
				))}
			</ul>
		</div>
	);
}
