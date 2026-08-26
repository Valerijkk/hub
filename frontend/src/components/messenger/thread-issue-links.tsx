import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import {
	createIssueMessageLink,
	deleteIssueMessageLink,
	listThreadIssueLinks,
} from "@/api/issue-message-links";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

interface ThreadIssueLinksProps {
	projectId: string;
	channelId: string;
	messageId: string;
}

export const THREAD_ISSUE_LINKS_QUERY_KEY = ["thread-issue-links"] as const;

export function threadIssueLinksQueryKey(
	projectId: string,
	channelId: string,
	messageId: string,
	userId: string,
) {
	return [
		...THREAD_ISSUE_LINKS_QUERY_KEY,
		projectId,
		channelId,
		messageId,
		userId,
	] as const;
}

/** PROD-03: связанные задачи в заголовке треда. */
export function ThreadIssueLinks({
	projectId,
	channelId,
	messageId,
}: ThreadIssueLinksProps) {
	const queryClient = useQueryClient();
	const [issueKey, setIssueKey] = useState("");
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const hasAccess = Boolean(userId) && !channelRevoked;
	const queryKey = threadIssueLinksQueryKey(
		projectId,
		channelId,
		messageId,
		userId ?? "__no-user__",
	);
	const canMutateNow = useCallback(() => {
		return (
			Boolean(userId) &&
			useAuthStore.getState().user?.id === userId &&
			!useMessengerStore.getState().revokedChannelIds.has(channelId)
		);
	}, [channelId, userId]);

	const { data: links = [] } = useQuery({
		queryKey,
		queryFn: () => listThreadIssueLinks(projectId, channelId, messageId),
		enabled: hasAccess,
	});

	const linkMutation = useMutation({
		mutationFn: () => {
			if (!canMutateNow()) {
				throw new Error("thread issue links access unavailable");
			}
			return createIssueMessageLink(projectId, issueKey.trim().toUpperCase(), {
				channel_id: channelId,
				message_id: messageId,
				mirror_to_thread: true,
			});
		},
		onSuccess: () => {
			if (!canMutateNow()) return;
			toast.success("Задача привязана к треду");
			setIssueKey("");
			void queryClient.invalidateQueries({
				queryKey,
				exact: true,
			});
		},
		onError: (err) => {
			if (canMutateNow()) toastApiError(err);
		},
	});

	const unlinkMutation = useMutation({
		mutationFn: (link: { issue_key?: string; id: string }) => {
			if (!canMutateNow()) {
				throw new Error("thread issue links access unavailable");
			}
			if (!link.issue_key) throw new Error("missing issue key");
			return deleteIssueMessageLink(projectId, link.issue_key, link.id);
		},
		onSuccess: () => {
			if (!canMutateNow()) return;
			void queryClient.invalidateQueries({
				queryKey,
				exact: true,
			});
		},
		onError: (err) => {
			if (canMutateNow()) toastApiError(err);
		},
	});

	if (!hasAccess) return null;

	return (
		<div className="border-b bg-raised/20 px-4 py-2 space-y-2">
			<div className="flex flex-wrap items-center gap-2">
				<span className="text-xs font-medium text-t3">Задачи:</span>
				{links.length === 0 && (
					<span className="text-xs text-t4">нет привязок</span>
				)}
				{links.map((link) => (
					<span key={link.id} className="inline-flex items-center gap-1">
						<Link
							to="/projects/$projectId/issues/$issueKey"
							params={{
								projectId,
								issueKey: link.issue_key ?? link.issue_id,
							}}
							className="rounded bg-primary/10 px-2 py-0.5 text-xs font-mono text-primary hover:underline"
						>
							{link.issue_key}
						</Link>
						<button
							type="button"
							className="text-t4 hover:text-[var(--red)] text-xs"
							onClick={() => {
								if (canMutateNow()) unlinkMutation.mutate(link);
							}}
							aria-label={`Отвязать ${link.issue_key}`}
						>
							×
						</button>
					</span>
				))}
			</div>
			<div className="flex gap-2">
				<Input
					value={issueKey}
					onChange={(e) => setIssueKey(e.target.value)}
					placeholder="HUB-42"
					className="h-8 text-xs max-w-[140px]"
				/>
				<Button
					type="button"
					size="sm"
					variant="outline"
					className="h-8 text-xs"
					disabled={!issueKey.trim() || linkMutation.isPending}
					onClick={() => {
						if (canMutateNow()) linkMutation.mutate();
					}}
				>
					Привязать
				</Button>
			</div>
		</div>
	);
}
