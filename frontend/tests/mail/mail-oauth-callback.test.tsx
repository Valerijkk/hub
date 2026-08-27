import { StrictMode } from "react";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { MailAccountStatus, MailSetupStatus } from "@/api/mail";
import * as mailApi from "@/api/mail";
import { MailOAuthCallback } from "@/modules/mail/mail-oauth-callback";
import { rememberMailOAuthFlow } from "@/modules/mail/mail-oauth-flow";
import { mailKeys } from "@/modules/mail/mail-query-keys";
import { renderWithProviders } from "../support/render";

const navigate = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", async (importOriginal) => ({
	...(await importOriginal<typeof import("@tanstack/react-router")>()),
	useNavigate: () => navigate,
}));
vi.mock("@/api/mail");
vi.mock("react-i18next", async (importOriginal) => ({
	...(await importOriginal<typeof import("react-i18next")>()),
	useTranslation: () => ({
		t: (key: string) => key,
	}),
}));

const NOW_MS = 1_800_000_000_000;
const STATE = "A".repeat(43);
const OTHER_STATE = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE";
const CODE = "callback-code";

const activeAccount: MailAccountStatus = {
	status: "active",
	primary_address: "member@hub.local",
	retry_after_seconds: null,
	error_code: null,
	capabilities: {
		rich_content: false,
		rules: false,
		spam_training: false,
		remote_images: false,
	},
};

const oauthRequiredSetup: MailSetupStatus = {
	stage: "oauth_required",
	primary_address: "member@hub.local",
};

function callbackSearch(state = STATE, code = CODE): string {
	return `?${new URLSearchParams({ state, code }).toString()}`;
}

function setCallbackLocation(search: string): void {
	window.history.replaceState(
		{ callback: true },
		"",
		`/mail/oauth/callback${search}#provider-fragment`,
	);
}

function rememberFlow(flow: "setup" | "connection", state = STATE): void {
	expect(
		rememberMailOAuthFlow(
			`https://oauth.example.test/authorize?state=${state}`,
			flow,
			window.sessionStorage,
			NOW_MS,
		),
	).toBe(true);
}

function renderCallback(strict = false) {
	return renderWithProviders(
		strict ? (
			<StrictMode>
				<MailOAuthCallback />
			</StrictMode>
		) : (
			<MailOAuthCallback />
		),
	);
}

describe("MailOAuthCallback", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.spyOn(Date, "now").mockReturnValue(NOW_MS);
		window.sessionStorage.clear();
		setCallbackLocation(callbackSearch());
		vi.mocked(mailApi.getMailAccountStatus).mockResolvedValue(activeAccount);
		vi.mocked(mailApi.getMailSetupStatus).mockResolvedValue(oauthRequiredSetup);
		vi.mocked(mailApi.completeMailSetupOAuth).mockResolvedValue();
		vi.mocked(mailApi.completeMailOAuth).mockResolvedValue();
	});

	afterEach(() => {
		vi.restoreAllMocks();
		window.sessionStorage.clear();
		window.history.replaceState({}, "", "/mail");
	});

	it("uses only the setup callback endpoint for a setup marker", async () => {
		rememberFlow("setup");
		const { queryClient } = renderCallback();
		const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");

		await waitFor(() => {
			expect(mailApi.completeMailSetupOAuth).toHaveBeenCalledWith({
				state: STATE,
				code: CODE,
			});
		});

		expect(mailApi.completeMailOAuth).not.toHaveBeenCalled();
		expect(mailApi.getMailAccountStatus).not.toHaveBeenCalled();
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: mailKeys.setup(),
		});
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: mailKeys.mailbox(),
		});
		expect(navigate).toHaveBeenCalledWith({ to: "/mail", replace: true });
	});

	it("uses only the normal callback endpoint for an active connection marker", async () => {
		rememberFlow("connection");
		vi.mocked(mailApi.getMailSetupStatus).mockResolvedValue({
			...oauthRequiredSetup,
			stage: "ready",
		});

		renderCallback();

		await waitFor(() => {
			expect(mailApi.completeMailOAuth).toHaveBeenCalledWith({
				state: STATE,
				code: CODE,
			});
		});

		expect(mailApi.completeMailSetupOAuth).not.toHaveBeenCalled();
		expect(mailApi.getMailAccountStatus).toHaveBeenCalledTimes(1);
		expect(mailApi.getMailSetupStatus).toHaveBeenCalledTimes(1);
	});

	it.each([
		["a duplicate state", `?state=${STATE}&state=${OTHER_STATE}&code=${CODE}`],
		["a duplicate code", `?state=${STATE}&code=${CODE}&code=second`],
		["an empty code", `?state=${STATE}&code=`],
		["a malformed state", `?state=invalid&code=${CODE}`],
	])("does not call an API for %s", async (_description, search) => {
		rememberFlow("setup");
		setCallbackLocation(search);

		renderCallback();

		expect(await screen.findByRole("alert")).toHaveTextContent("callback.error");
		expect(mailApi.getMailSetupStatus).not.toHaveBeenCalled();
		expect(mailApi.getMailAccountStatus).not.toHaveBeenCalled();
		expect(mailApi.completeMailSetupOAuth).not.toHaveBeenCalled();
		expect(mailApi.completeMailOAuth).not.toHaveBeenCalled();
	});

	it.each([
		["a missing marker", null],
		["a mismatched marker", OTHER_STATE],
	])("rejects %s without calling an API", async (_description, markerState) => {
		if (markerState !== null) {
			rememberFlow("setup", markerState);
		}

		renderCallback();

		expect(await screen.findByRole("alert")).toHaveTextContent("callback.error");
		expect(mailApi.getMailSetupStatus).not.toHaveBeenCalled();
		expect(mailApi.getMailAccountStatus).not.toHaveBeenCalled();
		expect(mailApi.completeMailSetupOAuth).not.toHaveBeenCalled();
		expect(mailApi.completeMailOAuth).not.toHaveBeenCalled();
	});

	it("rejects a setup callback when the setup stage changed", async () => {
		rememberFlow("setup");
		vi.mocked(mailApi.getMailSetupStatus).mockResolvedValue({
			...oauthRequiredSetup,
			stage: "ready",
		});

		renderCallback();

		expect(await screen.findByRole("alert")).toHaveTextContent("callback.error");
		expect(mailApi.completeMailSetupOAuth).not.toHaveBeenCalled();
		expect(mailApi.completeMailOAuth).not.toHaveBeenCalled();
	});

	it("rejects a connection callback when its mailbox is no longer active", async () => {
		rememberFlow("connection");
		vi.mocked(mailApi.getMailAccountStatus).mockResolvedValue({
			...activeAccount,
			status: "provisioning",
		});
		vi.mocked(mailApi.getMailSetupStatus).mockResolvedValue({
			...oauthRequiredSetup,
			stage: "action_required",
		});

		renderCallback();

		expect(await screen.findByRole("alert")).toHaveTextContent("callback.error");
		expect(mailApi.completeMailSetupOAuth).not.toHaveBeenCalled();
		expect(mailApi.completeMailOAuth).not.toHaveBeenCalled();
	});

	it("scrubs callback query and fragment before the first asynchronous API call", async () => {
		rememberFlow("setup");
		const replaceState = vi.spyOn(window.history, "replaceState");
		vi.mocked(mailApi.getMailSetupStatus).mockImplementation(async () => {
			expect(window.location.search).toBe("");
			expect(window.location.hash).toBe("");
			return oauthRequiredSetup;
		});

		renderCallback();

		await waitFor(() => {
			expect(mailApi.getMailSetupStatus).toHaveBeenCalledTimes(1);
		});
		expect(replaceState).toHaveBeenCalledWith(
			{ callback: true },
			"",
			"/mail/oauth/callback",
		);
	});

	it("runs the callback only once in StrictMode", async () => {
		rememberFlow("setup");

		renderCallback(true);

		await waitFor(() => {
			expect(mailApi.completeMailSetupOAuth).toHaveBeenCalledTimes(1);
		});
		expect(mailApi.getMailSetupStatus).toHaveBeenCalledTimes(1);
	});

	it("shows a generic non-leaking error with a back control", async () => {
		rememberFlow("setup");
		vi.mocked(mailApi.completeMailSetupOAuth).mockRejectedValue(
			new Error(`provider rejected ${STATE} ${CODE}`),
		);

		renderCallback();

		expect(await screen.findByRole("alert")).toHaveTextContent("callback.error");
		const back = screen.getByRole("button", { name: "callback.back" });
		expect(back).toBeInTheDocument();
		expect(document.body).not.toHaveTextContent(STATE);
		expect(document.body).not.toHaveTextContent(CODE);
		expect(navigate).not.toHaveBeenCalled();

		fireEvent.click(back);
		expect(navigate).toHaveBeenCalledWith({ to: "/mail", replace: true });
	});
});
