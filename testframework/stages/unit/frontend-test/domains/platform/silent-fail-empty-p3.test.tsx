import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CallChatMessage } from "@/api/calls";
import { CallChat } from "@/components/calls/CallChat";
import { BYOKSettings } from "@/components/ai/byok-settings";
import { renderWithProviders } from "@testframework/render";

const root = resolve(__dirname, "../../../../../../frontend/src");

function readSrc(rel: string): string {
	return readFileSync(resolve(root, rel), "utf8");
}

const { toastApiError } = vi.hoisted(() => ({
	toastApiError: vi.fn(),
}));

vi.mock("@/lib/format-api-error", async () => {
	const actual = await vi.importActual<typeof import("@/lib/format-api-error")>(
		"@/lib/format-api-error",
	);
	return { ...actual, toastApiError };
});

vi.mock("@/api/ai-my", () => ({
	listMyAIProviders: vi.fn(),
	createMyAIProvider: vi.fn(),
}));

vi.mock("@/api/calls", () => ({
	listChatMessages: vi.fn(),
	sendChatMessage: vi.fn(),
}));

vi.mock("@/api/issue-message-links", () => ({
	listIssueMessageLinks: vi.fn(),
	deleteIssueMessageLink: vi.fn(),
}));

import { listMyAIProviders } from "@/api/ai-my";
import { listChatMessages } from "@/api/calls";
import {
	deleteIssueMessageLink,
	listIssueMessageLinks,
} from "@/api/issue-message-links";
import { IssueMessageLinksPanel } from "@/modules/tracker/issue-message-links-panel";

describe("P3-FE-06 silent fail → error UI (TASK-073)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("BYOK refresh calls toastApiError on reject (not silent empty)", async () => {
		vi.mocked(listMyAIProviders).mockRejectedValueOnce({
			code: 500,
			message: "boom",
		});
		renderWithProviders(<BYOKSettings />);
		await waitFor(() => {
			expect(toastApiError).toHaveBeenCalled();
		});
		expect(screen.getByRole("alert")).toBeInTheDocument();
		expect(screen.queryByText(/нет byok/i)).not.toBeInTheDocument();
	});

	it("CallChat shows error≠empty when history fails", async () => {
		vi.mocked(listChatMessages).mockRejectedValueOnce(new Error("network"));
		renderWithProviders(<CallChat callId="call-1" />);
		await waitFor(() => {
			expect(screen.getByRole("alert")).toBeInTheDocument();
		});
		expect(screen.queryByText(/нет сообщений/i)).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});

	it("CallChat shows empty only after successful empty load", async () => {
		vi.mocked(listChatMessages).mockResolvedValueOnce([] as CallChatMessage[]);
		renderWithProviders(<CallChat callId="call-2" />);
		await waitFor(() => {
			expect(screen.getByText(/нет сообщений/i)).toBeInTheDocument();
		});
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("issue-message-links shows alert on list reject", async () => {
		vi.mocked(listIssueMessageLinks).mockRejectedValueOnce(new Error("fail"));
		renderWithProviders(
			<IssueMessageLinksPanel projectId="p1" issueKey="HUB-1" />,
		);
		await waitFor(() => {
			expect(screen.getByRole("alert")).toBeInTheDocument();
		});
	});

	it("poll-message and bookmarks use toastApiError on mutate fail", () => {
		const poll = readSrc("components/messenger/poll-message.tsx");
		expect(poll).toContain("toastApiError");
		expect(poll).toContain("Не удалось проголосовать");
		expect(poll).toContain("Не удалось закрыть опрос");
		expect(poll).toMatch(/channelId/);

		const bookmarks = readSrc("components/messenger/bookmarks-panel.tsx");
		expect(bookmarks).toMatch(/toastApiError/);
		expect(bookmarks).toMatch(/QueryStateView/);
	});

	it("issue unlink wires onError toastApiError", () => {
		expect(typeof deleteIssueMessageLink).toBe("function");
		const src = readSrc("modules/tracker/issue-message-links-panel.tsx");
		expect(src).toMatch(/onError:[\s\S]*toastApiError/);
		expect(src).toMatch(/isError/);
	});
});
