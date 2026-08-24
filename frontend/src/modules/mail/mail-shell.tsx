import { useQuery } from "@tanstack/react-query";
import { PenLine } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { PropsWithChildren, ReactNode } from "react";
import { useTranslation } from "react-i18next";
import {
	getMailAccountStatus,
	type MailAccountStatus,
	type MailDraft,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { MailSidebar } from "./mail-sidebar";
import { MailComposer } from "./mail-composer";
import { MailComposerContext } from "./mail-composer-context";
import { mailKeys } from "./mail-query-keys";

type MailShellProps = PropsWithChildren;

interface MailStatePanelProps {
	title: string;
	description: string;
	retryLabel?: string;
	onRetry?: () => void;
}

function MailStatePanel({
	title,
	description,
	retryLabel,
	onRetry,
}: MailStatePanelProps) {
	return (
		<Card>
			<CardHeader>
				<CardTitle>{title}</CardTitle>
				<CardDescription>{description}</CardDescription>
			</CardHeader>
			{retryLabel && onRetry ? (
				<CardContent>
					<Button size="lg" variant="outline" onClick={onRetry}>
						{retryLabel}
					</Button>
				</CardContent>
			) : null}
		</Card>
	);
}

function accountStateCopy(status: MailAccountStatus["status"]): {
	title: string;
	description: string;
	canRetry: boolean;
} {
	switch (status) {
		case "not_configured":
			return {
				title: "notConfigured.title",
				description: "notConfigured.description",
				canRetry: false,
			};
		case "provisioning":
			return {
				title: "provisioning.title",
				description: "provisioning.description",
				canRetry: true,
			};
		case "suspended":
			return {
				title: "suspended.title",
				description: "suspended.description",
				canRetry: true,
			};
		case "error":
			return {
				title: "error.title",
				description: "error.description",
				canRetry: true,
			};
		case "deleting":
			return {
				title: "deleting.title",
				description: "deleting.description",
				canRetry: true,
			};
		case "active":
			return {
				title: "active.title",
				description: "active.description",
				canRetry: false,
			};
	}
}

export function MailShell({ children }: MailShellProps) {
	const { t } = useTranslation("mail");
	const [composerOpen, setComposerOpen] = useState(false);
	const [composerDraft, setComposerDraft] = useState<MailDraft | null>(null);
	const accountQuery = useQuery({
		queryKey: mailKeys.mailbox(),
		queryFn: getMailAccountStatus,
		retry: false,
	});
	const stateResultRef = useRef<HTMLDivElement>(null);
	const retryIntentRef = useRef(false);
	const retryControlVisible =
		!accountQuery.isPending &&
		(accountQuery.isError ||
			!accountQuery.data ||
			(accountQuery.data.status !== "active" &&
				accountStateCopy(accountQuery.data.status).canRetry));
	const composerAvailable =
		!accountQuery.isPending &&
		!accountQuery.isError &&
		accountQuery.data?.status === "active";

	const retryAccount = () => {
		retryIntentRef.current = true;
		void accountQuery.refetch();
	};
	const openNewComposer = useCallback(() => {
		setComposerDraft(null);
		setComposerOpen(true);
	}, []);
	const openReplyDraft = useCallback((draft: MailDraft) => {
		setComposerDraft(draft);
		setComposerOpen(true);
	}, []);
	const handleComposerOpenChange = useCallback((nextOpen: boolean) => {
		setComposerOpen(nextOpen);
		if (!nextOpen) {
			setComposerDraft(null);
		}
	}, []);
	const composerContext = useMemo(
		() => ({
			isComposerOpen: composerOpen,
			openReplyDraft,
		}),
		[composerOpen, openReplyDraft],
	);

	useEffect(() => {
		if (!retryIntentRef.current) return;

		if (!retryControlVisible) {
			stateResultRef.current?.focus();
			retryIntentRef.current = false;
			return;
		}

		if (accountQuery.fetchStatus === "idle") {
			retryIntentRef.current = false;
		}
	}, [accountQuery.fetchStatus, retryControlVisible]);

	let content: ReactNode;
	if (accountQuery.isPending) {
		content = (
			<Card>
				<CardContent className="p-6" aria-busy="true">
					<p role="status" aria-live="polite" className="text-sm text-t2">
						{t("loading")}
					</p>
				</CardContent>
			</Card>
		);
	} else if (accountQuery.isError || !accountQuery.data) {
		content = (
			<MailStatePanel
				title={t("unavailable.title")}
				description={t("unavailable.description")}
				retryLabel={t("retry")}
				onRetry={retryAccount}
			/>
		);
	} else if (accountQuery.data.status === "active") {
		const state = accountStateCopy(accountQuery.data.status);
		content = (
			<>
				<MailStatePanel
					title={t(state.title)}
					description={t(state.description)}
				/>
				<div className="mt-4 overflow-hidden rounded-[var(--r4)] border border-b1 bg-card md:flex">
					<MailSidebar />
					<section className="min-w-0 flex-1 p-4 sm:p-6">{children}</section>
				</div>
			</>
		);
	} else {
		const state = accountStateCopy(accountQuery.data.status);
		content = (
			<MailStatePanel
				title={t(state.title)}
				description={t(state.description)}
				retryLabel={state.canRetry ? t("retry") : undefined}
				onRetry={state.canRetry ? retryAccount : undefined}
			/>
		);
	}

	return (
		<MailComposerContext.Provider value={composerContext}>
			<main className="min-h-full bg-base px-4 py-6 sm:px-6 sm:py-8">
			<div className="mx-auto w-full max-w-7xl">
				<div className="mb-6 flex items-center justify-between gap-4">
					<h1 className="text-2xl font-bold tracking-tight text-t1">
						{t("title")}
					</h1>
					{composerAvailable ? (
						<Button onClick={openNewComposer}>
							<PenLine aria-hidden="true" />
							{t("compose.open")}
						</Button>
					) : null}
				</div>
				<div ref={stateResultRef} tabIndex={-1} aria-live="polite">
					{content}
				</div>
				{composerAvailable ? (
					<MailComposer
						key={composerDraft?.id ?? "new"}
						open={composerOpen}
						bootstrapDraft={composerDraft}
						onOpenChange={handleComposerOpenChange}
					/>
				) : null}
			</div>
			</main>
		</MailComposerContext.Provider>
	);
}
