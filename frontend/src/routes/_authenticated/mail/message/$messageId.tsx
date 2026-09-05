import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { getMailMessage } from "@/api/mail";
import { Button } from "@/components/ui/button";
import { isValidMailEntityID } from "@/modules/mail/mail-thread-list";
import { mailKeys } from "@/modules/mail/mail-query-keys";
import { useAuthStore } from "@/stores/auth-store";

export const Route = createFileRoute("/_authenticated/mail/message/$messageId")(
	{
		component: DirectMailMessageRoute,
	},
);

function canonicalID(value: unknown): string | null {
	if (typeof value !== "string" || !isValidMailEntityID(value)) {
		return null;
	}

	return value.toLowerCase();
}

function DirectMailMessageRoute() {
	const { messageId } = Route.useParams();
	const actorID = useAuthStore((state) => state.user?.id ?? null);
	const canonicalActorID = actorID === null ? null : canonicalID(actorID);
	const canonicalMessageID = canonicalID(messageId);
	const navigate = useNavigate();
	const { t } = useTranslation("mail");

	const directMessageQuery = useQuery({
		queryKey: mailKeys.directMessage(canonicalActorID, canonicalMessageID),
		enabled: canonicalActorID !== null && canonicalMessageID !== null,
		queryFn: ({ signal }) => {
			if (canonicalMessageID === null) {
				return Promise.reject(new Error("Invalid direct mail message query"));
			}
			return getMailMessage(canonicalMessageID, { signal });
		},
		retry: false,
		staleTime: 0,
		gcTime: 0,
		refetchOnWindowFocus: false,
	});

	const responseMessageID =
		directMessageQuery.data === undefined
			? null
			: canonicalID(directMessageQuery.data.id);
	const responseThreadID =
		directMessageQuery.data === undefined
			? null
			: canonicalID(directMessageQuery.data.thread_id);
	const hasResponseMismatch =
		directMessageQuery.data !== undefined &&
		(responseMessageID !== canonicalMessageID || responseThreadID === null);

	useEffect(() => {
		if (
			canonicalActorID === null ||
			canonicalMessageID === null ||
			responseMessageID !== canonicalMessageID ||
			responseThreadID === null
		) {
			return;
		}

		void navigate({
			to: "/mail/$folderKey/thread/$threadId",
			params: {
				folderKey: "all",
				threadId: responseThreadID,
			},
			replace: true,
		});
	}, [
		canonicalActorID,
		canonicalMessageID,
		navigate,
		responseMessageID,
		responseThreadID,
	]);

	if (canonicalMessageID === null) {
		return (
			<DirectMessageErrorPanel
				description={t("reader.direct.invalidDescription")}
				title={t("reader.direct.invalidTitle")}
			/>
		);
	}

	if (actorID !== null && canonicalActorID === null) {
		return (
			<DirectMessageErrorPanel
				description={t("reader.direct.unavailableDescription")}
				title={t("reader.direct.unavailableTitle")}
			/>
		);
	}

	if (actorID === null) {
		return <DirectMessageOpeningPanel text={t("reader.direct.opening")} />;
	}

	if (directMessageQuery.isError || hasResponseMismatch) {
		return (
			<DirectMessageErrorPanel
				description={t("reader.direct.unavailableDescription")}
				onRetry={() => {
					void directMessageQuery.refetch();
				}}
				retryDisabled={directMessageQuery.isFetching}
				retryLabel={t("reader.direct.retry")}
				title={t("reader.direct.unavailableTitle")}
			/>
		);
	}

	return <DirectMessageOpeningPanel text={t("reader.direct.opening")} />;
}

function DirectMessageOpeningPanel({ text }: { readonly text: string }) {
	return (
		<div className="mx-auto w-full max-w-xl p-4 sm:p-6">
			<section
				aria-busy="true"
				aria-labelledby="mail-direct-opening-title"
				className="rounded-[var(--r3)] border border-b1 bg-card p-6"
				role="status"
			>
				<h2
					className="text-lg font-semibold text-t1"
					id="mail-direct-opening-title"
				>
					{text}
				</h2>
			</section>
		</div>
	);
}

interface DirectMessageErrorPanelProps {
	readonly description: string;
	readonly onRetry?: () => void;
	readonly retryDisabled?: boolean;
	readonly retryLabel?: string;
	readonly title: string;
}

function DirectMessageErrorPanel({
	description,
	onRetry,
	retryDisabled = false,
	retryLabel,
	title,
}: DirectMessageErrorPanelProps) {
	const { t } = useTranslation("mail");

	return (
		<div className="mx-auto w-full max-w-xl p-4 sm:p-6">
			<section
				aria-labelledby="mail-direct-error-title"
				className="rounded-[var(--r3)] border border-b1 bg-card p-6"
				role="alert"
			>
				<h2
					className="text-lg font-semibold text-t1"
					id="mail-direct-error-title"
				>
					{title}
				</h2>
				<p className="mt-2 text-sm leading-6 text-t2">{description}</p>
				<div className="mt-5 flex flex-wrap gap-3">
					{onRetry !== undefined && retryLabel !== undefined ? (
						<Button
							className="min-h-11"
							disabled={retryDisabled}
							onClick={onRetry}
							type="button"
							variant="outline"
						>
							{retryLabel}
						</Button>
					) : null}
					<Button asChild className="min-h-11" variant="outline">
						<Link to="/mail">{t("reader.direct.back")}</Link>
					</Button>
				</div>
			</section>
		</div>
	);
}
