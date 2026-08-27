import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import type { FormEvent, ReactNode } from "react";
import { useTranslation } from "react-i18next";
import {
	beginMailSetupOAuth,
	enrollMailSetupPassword,
	getMailSetupStatus,
	type MailSetupStage,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { rememberMailOAuthFlow } from "./mail-oauth-flow";
import { mailKeys } from "./mail-query-keys";

const PASSWORD_MIN_BYTES = 12;
const PASSWORD_MAX_BYTES = 256;
const OAUTH_URL_MAX_LENGTH = 2048;
const CONTROL_CHARACTERS = /[\u0000-\u001F\u007F-\u009F]/;
const ENCODED_CONTROL_CHARACTERS = /%(?:0[0-9A-F]|1[0-9A-F]|7F)/i;
const MAIL_SETUP_STAGES = new Set<MailSetupStage>([
	"preparing",
	"password_required",
	"oauth_required",
	"syncing",
	"ready",
	"action_required",
]);

export interface MailSetupPanelProps {
	onReady?: () => void;
}

function isMailSetupStage(value: unknown): value is MailSetupStage {
	return typeof value === "string" && MAIL_SETUP_STAGES.has(value as MailSetupStage);
}

function normalizePrimaryAddress(value: string | null): string | null {
	if (typeof value !== "string") return null;
	const normalized = value.trim().toLowerCase();
	const separator = normalized.indexOf("@");
	if (
		normalized.length === 0 ||
		normalized.length > 320 ||
		CONTROL_CHARACTERS.test(normalized) ||
		/\s/.test(normalized) ||
		separator <= 0 ||
		separator !== normalized.lastIndexOf("@") ||
		separator === normalized.length - 1
	) {
		return null;
	}
	return normalized;
}

function isSafeAuthorizationURL(value: string): boolean {
	if (
		value.length === 0 ||
		value.length > OAUTH_URL_MAX_LENGTH ||
		CONTROL_CHARACTERS.test(value) ||
		ENCODED_CONTROL_CHARACTERS.test(value)
	) {
		return false;
	}
	let decoded: string;
	try {
		decoded = decodeURIComponent(value);
	} catch {
		return false;
	}
	if (CONTROL_CHARACTERS.test(decoded)) return false;

	try {
		const url = new URL(value);
		const isLocalHTTP =
			url.protocol === "http:" && url.hostname.toLowerCase() === "localhost";
		return (
			url.href.length <= OAUTH_URL_MAX_LENGTH &&
			(url.protocol === "https:" || isLocalHTTP) &&
			url.username.length === 0 &&
			url.password.length === 0
		);
	} catch {
		return false;
	}
}

function SetupCard({
	title,
	description,
	address,
	children,
}: {
	title: string;
	description: string;
	address: string | null;
	children?: ReactNode;
}) {
	const { t } = useTranslation("mail");
	return (
		<Card>
			<CardHeader>
				<CardTitle>{title}</CardTitle>
				<CardDescription>{description}</CardDescription>
				{address ? (
					<p className="pt-2 text-sm font-medium text-t1">
						<span className="text-t3">{t("setup.address")}: </span>
						{address}
					</p>
				) : null}
			</CardHeader>
			{children ? <CardContent>{children}</CardContent> : null}
		</Card>
	);
}

export function MailSetupPanel({ onReady }: MailSetupPanelProps) {
	const { t } = useTranslation("mail");
	const [password, setPassword] = useState("");
	const [confirmation, setConfirmation] = useState("");
	const [isSubmittingPassword, setIsSubmittingPassword] = useState(false);
	const [isStartingOAuth, setIsStartingOAuth] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [notice, setNotice] = useState<string | null>(null);
	const readySignaledRef = useRef(false);
	const setupQuery = useQuery({
		queryKey: mailKeys.setup(),
		queryFn: getMailSetupStatus,
		retry: false,
		refetchInterval: (query) =>
			query.state.data?.stage === "syncing" ? 2_000 : false,
	});

	const setup = setupQuery.data;
	const stage = setup?.stage;
	const address = normalizePrimaryAddress(setup?.primary_address ?? null);
	const isBusy = isSubmittingPassword || isStartingOAuth;

	useEffect(() => {
		if (stage !== "ready" || readySignaledRef.current) return;
		readySignaledRef.current = true;
		onReady?.();
	}, [onReady, stage]);

	const clearSecrets = () => {
		setPassword("");
		setConfirmation("");
	};

	const submitPassword = async (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		setError(null);
		setNotice(null);

		const submittedPassword = password;
		const submittedConfirmation = confirmation;
		const passwordBytes = new TextEncoder().encode(submittedPassword);
		clearSecrets();
		if (
			submittedPassword !== submittedConfirmation ||
			passwordBytes.byteLength < PASSWORD_MIN_BYTES ||
			passwordBytes.byteLength > PASSWORD_MAX_BYTES
		) {
			passwordBytes.fill(0);
			setError(t("setup.password.validation"));
			return;
		}

		setIsSubmittingPassword(true);
		try {
			const response = await enrollMailSetupPassword(passwordBytes);
			if (
				response.status !== "enrolled" &&
				response.status !== "already_enrolled"
			) {
				throw new Error("Unexpected mail setup password response");
			}
			setNotice(t("setup.password.accepted"));
			await setupQuery.refetch();
		} catch {
			setError(t("setup.error"));
		} finally {
			passwordBytes.fill(0);
			clearSecrets();
			setIsSubmittingPassword(false);
		}
	};

	const startOAuth = async () => {
		setError(null);
		setNotice(null);
		setIsStartingOAuth(true);
		try {
			const response = await beginMailSetupOAuth();
			if (!isSafeAuthorizationURL(response.authorization_url)) {
				throw new Error("Unsafe mail setup authorization URL");
			}
			if (
				!rememberMailOAuthFlow(
					response.authorization_url,
					"setup",
					window.sessionStorage,
					Date.now(),
				)
			) {
				throw new Error("Unable to remember mail OAuth flow");
			}
			window.location.assign(response.authorization_url);
		} catch {
			setError(t("setup.error"));
		} finally {
			setIsStartingOAuth(false);
		}
	};

	if (setupQuery.isPending) {
		return (
			<SetupCard
				title={t("setup.preparing.title")}
				description={t("setup.preparing.description")}
				address={null}
			>
				<p role="status" aria-live="polite" aria-busy="true" className="text-sm text-t2">
					{t("setup.preparing.description")}
				</p>
			</SetupCard>
		);
	}

	if (setupQuery.isError || !setup || !isMailSetupStage(stage)) {
		return (
			<SetupCard
				title={t("setup.unavailable.title")}
				description={t("setup.unavailable.description")}
				address={null}
			>
				<Button
					variant="outline"
					disabled={isBusy || setupQuery.isFetching}
					onClick={() => void setupQuery.refetch()}
				>
					{t("retry")}
				</Button>
			</SetupCard>
		);
	}

	if (stage === "password_required") {
		return (
			<SetupCard
				title={t("setup.password.title")}
				description={t("setup.password.description")}
				address={address}
			>
				<form className="max-w-md space-y-4" onSubmit={submitPassword}>
					<div className="space-y-2">
						<Label htmlFor="mail-setup-password">
							{t("setup.password.passwordLabel")}
						</Label>
						<Input
							id="mail-setup-password"
							type="password"
							autoComplete="new-password"
							value={password}
							disabled={isBusy}
							onChange={(event) => setPassword(event.target.value)}
						/>
					</div>
					<div className="space-y-2">
						<Label htmlFor="mail-setup-password-confirmation">
							{t("setup.password.confirmLabel")}
						</Label>
						<Input
							id="mail-setup-password-confirmation"
							type="password"
							autoComplete="new-password"
							value={confirmation}
							disabled={isBusy}
							onChange={(event) => setConfirmation(event.target.value)}
						/>
					</div>
					{error ? <p role="alert" className="text-sm text-[var(--red)]">{error}</p> : null}
					{notice ? <p role="status" className="text-sm text-t2">{notice}</p> : null}
					<Button type="submit" disabled={isBusy} aria-busy={isSubmittingPassword}>
						{isSubmittingPassword
							? t("setup.password.submitting")
							: t("setup.password.submit")}
					</Button>
				</form>
			</SetupCard>
		);
	}

	if (stage === "oauth_required") {
		return (
			<SetupCard
				title={t("setup.oauth.title")}
				description={t("setup.oauth.description")}
				address={address}
			>
				{error ? <p role="alert" className="mb-4 text-sm text-[var(--red)]">{error}</p> : null}
				<Button onClick={() => void startOAuth()} disabled={isBusy} aria-busy={isStartingOAuth}>
					{isStartingOAuth ? t("setup.oauth.submitting") : t("setup.oauth.submit")}
				</Button>
			</SetupCard>
		);
	}

	if (stage === "syncing") {
		return (
			<SetupCard
				title={t("setup.syncing.title")}
				description={t("setup.syncing.description")}
				address={address}
			>
				<p role="status" aria-live="polite" aria-busy="true" className="text-sm text-t2">
					{t("setup.syncing.description")}
				</p>
			</SetupCard>
		);
	}

	if (stage === "ready") {
		return (
			<SetupCard
				title={t("setup.ready.title")}
				description={t("setup.ready.description")}
				address={address}
			/>
		);
	}

	if (stage === "action_required") {
		return (
			<SetupCard
				title={t("setup.actionRequired.title")}
				description={t("setup.actionRequired.description")}
				address={address}
			>
				<Button
					variant="outline"
					disabled={isBusy || setupQuery.isFetching}
					onClick={() => void setupQuery.refetch()}
				>
					{t("setup.actionRequired.retry")}
				</Button>
			</SetupCard>
		);
	}

	return (
		<SetupCard
			title={t("setup.preparing.title")}
			description={t("setup.preparing.description")}
			address={address}
		/>
	);
}
