import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { MailAccountStatus } from "@/api/mail";
import * as mailApi from "@/api/mail";
import { MailShell } from "@/modules/mail/mail-shell";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/mail");
vi.mock("react-i18next", async (importOriginal) => ({
	...(await importOriginal<typeof import("react-i18next")>()),
	useTranslation: () => ({
		t: (key: string) => key,
	}),
}));
vi.mock("@/modules/mail/mail-setup-panel", () => ({
	MailSetupPanel: ({ onReady }: { onReady?: () => void }) => (
		<button type="button" onClick={onReady}>
			mock-mail-setup
		</button>
	),
}));
vi.mock("@/modules/mail/mail-sidebar", () => ({
	MailSidebar: () => <aside data-testid="mail-sidebar">sidebar</aside>,
}));
vi.mock("@/modules/mail/mail-composer", () => ({
	MailComposer: () => <div data-testid="mail-composer">composer</div>,
}));

const provisioningAccount: MailAccountStatus = {
	status: "provisioning",
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

const activeAccount: MailAccountStatus = {
	...provisioningAccount,
	status: "active",
};

function renderMailShell() {
	return renderWithProviders(
		<MailShell>
			<div data-testid="mail-content">mail content</div>
		</MailShell>,
	);
}

describe("MailShell", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("uses the setup panel while provisioning and refetches the account on ready", async () => {
		vi.mocked(mailApi.getMailAccountStatus).mockResolvedValue(
			provisioningAccount,
		);

		renderMailShell();

		const setup = await screen.findByRole("button", {
			name: "mock-mail-setup",
		});
		expect(screen.queryByTestId("mail-sidebar")).not.toBeInTheDocument();
		expect(screen.queryByTestId("mail-composer")).not.toBeInTheDocument();
		expect(screen.queryByTestId("mail-content")).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "compose.open" }),
		).not.toBeInTheDocument();
		expect(mailApi.getMailAccountStatus).toHaveBeenCalledTimes(1);

		fireEvent.click(setup);

		await waitFor(() => {
			expect(mailApi.getMailAccountStatus).toHaveBeenCalledTimes(2);
		});
	});

	it("keeps the compose control and mail content available for active accounts", async () => {
		vi.mocked(mailApi.getMailAccountStatus).mockResolvedValue(activeAccount);

		renderMailShell();

		expect(await screen.findByTestId("mail-sidebar")).toBeInTheDocument();
		expect(screen.getByTestId("mail-composer")).toBeInTheDocument();
		expect(screen.getByTestId("mail-content")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "compose.open" }),
		).toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "mock-mail-setup" }),
		).not.toBeInTheDocument();
	});
});
