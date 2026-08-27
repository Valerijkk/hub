import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	completeMailOAuth,
	completeMailSetupOAuth,
	getMailAccountStatus,
	getMailSetupStatus,
} from "@/api/mail";
import { Button } from "@/components/ui/button";
import {
	consumeMailOAuthFlow,
	isCanonicalMailOAuthState,
} from "./mail-oauth-flow";
import { mailKeys } from "./mail-query-keys";

const CALLBACK_CODE_MAX_BYTES = 8_192;
const CALLBACK_CODE_UNSAFE_CHARACTERS = /[\u0000-\u0020\u007F]/u;

interface MailOAuthCallbackValues {
	state: string;
	code: string;
}

function parseMailOAuthCallback(
	search: string,
): MailOAuthCallbackValues | null {
	const parameters = new URLSearchParams(search);
	const states = parameters.getAll("state");
	const codes = parameters.getAll("code");
	if (states.length !== 1 || codes.length !== 1) return null;

	const state = states[0];
	const code = codes[0];
	const codeBytes = new TextEncoder().encode(code);
	if (
		!isCanonicalMailOAuthState(state) ||
		CALLBACK_CODE_UNSAFE_CHARACTERS.test(code) ||
		codeBytes.byteLength < 1 ||
		codeBytes.byteLength > CALLBACK_CODE_MAX_BYTES
	) {
		return null;
	}

	return { state, code };
}

export function MailOAuthCallback() {
	const { t } = useTranslation("mail");
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const startedRef = useRef(false);
	const [failed, setFailed] = useState(false);

	useEffect(() => {
		if (startedRef.current) return;
		startedRef.current = true;

		const search = window.location.search;
		window.history.replaceState(
			window.history.state,
			"",
			window.location.pathname,
		);

		const callback = parseMailOAuthCallback(search);
		if (!callback) {
			setFailed(true);
			return;
		}

		const flow = consumeMailOAuthFlow(
			callback.state,
			window.sessionStorage,
			Date.now(),
		);
		if (!flow) {
			setFailed(true);
			return;
		}

		void (async () => {
			try {
				if (flow === "setup") {
					const setup = await getMailSetupStatus();
					if (setup.stage !== "oauth_required") {
						throw new Error("Mail setup OAuth callback is no longer expected");
					}
					await completeMailSetupOAuth(callback);
				} else {
					const [account, setup] = await Promise.all([
						getMailAccountStatus(),
						getMailSetupStatus(),
					]);
					if (
						account.status !== "active" ||
						(setup.stage !== "ready" && setup.stage !== "action_required")
					) {
						throw new Error("Mail OAuth callback connection state is invalid");
					}
					await completeMailOAuth(callback);
				}

				await Promise.all([
					queryClient.invalidateQueries({ queryKey: mailKeys.setup() }),
					queryClient.invalidateQueries({ queryKey: mailKeys.mailbox() }),
				]);
				await navigate({ to: "/mail", replace: true });
			} catch {
				setFailed(true);
			}
		})();
	}, [navigate, queryClient]);

	if (failed) {
		return (
			<div className="mx-auto flex min-h-[20rem] max-w-md flex-col items-center justify-center gap-4 px-6 text-center">
				<div>
					<h1 className="text-xl font-semibold text-t1">
						{t("callback.errorTitle")}
					</h1>
					<p role="alert" className="mt-2 text-sm leading-6 text-t2">
						{t("callback.error")}
					</p>
				</div>
				<Button
					variant="outline"
					onClick={() => void navigate({ to: "/mail", replace: true })}
				>
					{t("callback.back")}
				</Button>
			</div>
		);
	}

	return (
		<div className="mx-auto flex min-h-[20rem] max-w-md items-center justify-center px-6 text-center">
			<p role="status" aria-live="polite" aria-busy="true" className="text-sm text-t2">
				{t("callback.loading")}
			</p>
		</div>
	);
}
