import { useQuery } from "@tanstack/react-query";
import type { PropsWithChildren, ReactNode } from "react";
import { useTranslation } from "react-i18next";
import {
	getMailAccountStatus,
	type MailAccountStatus,
} from "@/api/mail-account";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { mailQueryKeys } from "./mail-query-keys";

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
	const accountQuery = useQuery({
		queryKey: mailQueryKeys.account(),
		queryFn: getMailAccountStatus,
		retry: false,
	});

	const retryAccount = () => {
		void accountQuery.refetch();
	};

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
				<div className="mt-4">{children}</div>
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
		<main className="min-h-full bg-base px-4 py-6 sm:px-6 sm:py-8">
			<div className="mx-auto w-full max-w-xl">
				<h1 className="mb-6 text-2xl font-bold tracking-tight text-t1">
					{t("title")}
				</h1>
				<div aria-live="polite">{content}</div>
			</div>
		</main>
	);
}
