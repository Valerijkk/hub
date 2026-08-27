import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
	beginMailOAuth,
	getMailOAuthConnectionStatus,
	type MailOAuthConnectionStatus,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { rememberMailOAuthFlow } from "./mail-oauth-flow";
import { mailKeys } from "./mail-query-keys";

interface ConnectionCopy {
	title: string;
	description: string;
	action: string;
}

function connectionCopy(
	status: Extract<
		MailOAuthConnectionStatus,
		"not_connected" | "reconnect_required" | "disconnected"
	>,
): ConnectionCopy {
	return status === "reconnect_required"
		? {
			title: "connection.reconnect.title",
			description: "connection.reconnect.description",
			action: "connection.reconnect.action",
		}
		: {
			title: "connection.connect.title",
			description: "connection.connect.description",
			action: "connection.connect.action",
		};
}

/**
 * Small active-mailbox control for starting an OAuth connect or reconnect.
 * The callback route uses the marker this component writes to select exactly
 * one completion endpoint; the marker contains no URL, authorization code or
 * credentials.
 */
export function MailConnectionPanel() {
	const { t } = useTranslation("mail");
	const [isAuthorizing, setIsAuthorizing] = useState(false);
	const [actionError, setActionError] = useState(false);
	const connectionQuery = useQuery({
		queryKey: mailKeys.oauthConnection(),
		queryFn: getMailOAuthConnectionStatus,
		retry: false,
	});

	const beginConnection = async () => {
		setActionError(false);
		setIsAuthorizing(true);
		try {
			const response = await beginMailOAuth();
			if (
				!rememberMailOAuthFlow(
					response.authorization_url,
					"connection",
					window.sessionStorage,
					Date.now(),
				)
			) {
				throw new Error("Unable to remember mail OAuth connection flow");
			}
			window.location.assign(response.authorization_url);
		} catch {
			setActionError(true);
		} finally {
			setIsAuthorizing(false);
		}
	};

	if (connectionQuery.isPending) {
		return null;
	}

	if (connectionQuery.isError || !connectionQuery.data) {
		return (
			<section
				className="mb-4"
				aria-labelledby="mail-connection-unavailable-title"
			>
				<Card>
					<CardHeader>
						<CardTitle id="mail-connection-unavailable-title">
							{t("connection.unavailable.title")}
						</CardTitle>
						<CardDescription>
							{t("connection.unavailable.description")}
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Button
							variant="outline"
							onClick={() => void connectionQuery.refetch()}
							disabled={connectionQuery.isFetching}
						>
							{t("connection.unavailable.retry")}
						</Button>
					</CardContent>
				</Card>
			</section>
		);
	}

	if (connectionQuery.data.status === "active") {
		return null;
	}

	const copy = connectionCopy(connectionQuery.data.status);
	const titleID = "mail-connection-title";
	return (
		<section className="mb-4" aria-labelledby={titleID}>
			<Card>
				<CardHeader>
					<CardTitle id={titleID}>{t(copy.title)}</CardTitle>
					<CardDescription>{t(copy.description)}</CardDescription>
				</CardHeader>
				<CardContent className="space-y-3">
					{actionError ? (
						<p role="alert" className="text-sm text-[var(--red)]">
							{t("connection.error")}
						</p>
					) : null}
					<Button onClick={() => void beginConnection()} disabled={isAuthorizing}>
						{isAuthorizing ? t("connection.authorizing") : t(copy.action)}
					</Button>
				</CardContent>
			</Card>
		</section>
	);
}
