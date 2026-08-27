import { fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { MailOAuthConnection } from "@/api/mail";
import * as mailApi from "@/api/mail";
import { MailConnectionPanel } from "@/modules/mail/mail-connection-panel";
import * as oauthFlow from "@/modules/mail/mail-oauth-flow";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/mail");
vi.mock("@/modules/mail/mail-oauth-flow", () => ({
	rememberMailOAuthFlow: vi.fn(),
}));
vi.mock("react-i18next", async (importOriginal) => ({
	...(await importOriginal<typeof import("react-i18next")>()),
	useTranslation: () => ({
		t: (key: string) => key,
	}),
}));

const NOW_MS = 1_800_000_000_000;
const STATE = "A".repeat(43);
const authorizationURL = "https://oauth.example.test/authorize?state=" + STATE;

function connection(
	status: MailOAuthConnection["status"],
): MailOAuthConnection {
	return {
		status,
		access_token_expires_at:
			status === "active" ? "2026-08-27T12:00:00Z" : null,
	};
}

function renderPanel() {
	return renderWithProviders(<MailConnectionPanel />);
}

describe("MailConnectionPanel", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.spyOn(Date, "now").mockReturnValue(NOW_MS);
		vi.mocked(mailApi.getMailOAuthConnectionStatus).mockResolvedValue(
			connection("reconnect_required"),
		);
		vi.mocked(mailApi.beginMailOAuth).mockResolvedValue({
			authorization_url: authorizationURL,
			expires_at: "2026-08-27T12:00:00Z",
		});
		vi.mocked(oauthFlow.rememberMailOAuthFlow).mockReturnValue(true);
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it.each([
		["not_connected", "connection.connect.title", "connection.connect.action"],
		["disconnected", "connection.connect.title", "connection.connect.action"],
		[
			"reconnect_required",
			"connection.reconnect.title",
			"connection.reconnect.action",
		],
	] as const)("maps %s to its compact connection control", async (status, title, action) => {
		vi.mocked(mailApi.getMailOAuthConnectionStatus).mockResolvedValue(
			connection(status),
		);

		renderPanel();

		expect(await screen.findByRole("heading", { name: title })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: action })).toBeInTheDocument();
	});

	it("renders nothing for an active connection", async () => {
		vi.mocked(mailApi.getMailOAuthConnectionStatus).mockResolvedValue(
			connection("active"),
		);

		renderPanel();

		await waitFor(() => {
			expect(mailApi.getMailOAuthConnectionStatus).toHaveBeenCalledTimes(1);
		});
		expect(screen.queryByRole("heading")).not.toBeInTheDocument();
	});

	it("uses normal OAuth, records the connection marker, then redirects", async () => {
		const assign = vi
			.spyOn(window.location, "assign")
			.mockImplementation(() => undefined);
		renderPanel();

		fireEvent.click(
			await screen.findByRole("button", {
				name: "connection.reconnect.action",
			}),
		);

		await waitFor(() => {
			expect(mailApi.beginMailOAuth).toHaveBeenCalledTimes(1);
		});
		expect(oauthFlow.rememberMailOAuthFlow).toHaveBeenCalledWith(
			authorizationURL,
			"connection",
			window.sessionStorage,
			NOW_MS,
		);
		expect(assign).toHaveBeenCalledWith(authorizationURL);
	});

	it("does not redirect and shows a generic error when marker persistence refuses", async () => {
		vi.mocked(oauthFlow.rememberMailOAuthFlow).mockReturnValue(false);
		const assign = vi
			.spyOn(window.location, "assign")
			.mockImplementation(() => undefined);
		renderPanel();

		fireEvent.click(
			await screen.findByRole("button", {
				name: "connection.reconnect.action",
			}),
		);

		expect(await screen.findByRole("alert")).toHaveTextContent(
			"connection.error",
		);
		expect(assign).not.toHaveBeenCalled();
		expect(document.body).not.toHaveTextContent("oauth.example.test");
	});

	it("disables the control while OAuth authorization is pending", async () => {
		let resolveAuthorization: ((value: {
			authorization_url: string;
			expires_at: string;
		}) => void) | null = null;
		vi.mocked(mailApi.beginMailOAuth).mockReturnValue(
			new Promise((resolve) => {
				resolveAuthorization = resolve;
			}),
		);
		const assign = vi
			.spyOn(window.location, "assign")
			.mockImplementation(() => undefined);
		renderPanel();

		const reconnect = await screen.findByRole("button", {
			name: "connection.reconnect.action",
		});
		fireEvent.click(reconnect);

		expect(
			screen.getByRole("button", { name: "connection.authorizing" }),
		).toBeDisabled();

		expect(resolveAuthorization).not.toBeNull();
		resolveAuthorization?.({
			authorization_url: authorizationURL,
			expires_at: "2026-08-27T12:00:00Z",
		});
		await waitFor(() => {
			expect(assign).toHaveBeenCalledWith(authorizationURL);
		});
	});

	it("shows only a retry status control when status retrieval fails", async () => {
		vi.mocked(mailApi.getMailOAuthConnectionStatus)
			.mockRejectedValueOnce(new Error("temporary status failure"))
			.mockResolvedValueOnce(connection("reconnect_required"));
		renderPanel();

		expect(
			await screen.findByRole("heading", {
				name: "connection.unavailable.title",
			}),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "connection.unavailable.retry" }),
		).toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "connection.reconnect.action" }),
		).not.toBeInTheDocument();

		fireEvent.click(
			screen.getByRole("button", { name: "connection.unavailable.retry" }),
		);

		expect(
			await screen.findByRole("button", {
				name: "connection.reconnect.action",
			}),
		).toBeInTheDocument();
		expect(mailApi.getMailOAuthConnectionStatus).toHaveBeenCalledTimes(2);
		expect(mailApi.beginMailOAuth).not.toHaveBeenCalled();
	});
});
