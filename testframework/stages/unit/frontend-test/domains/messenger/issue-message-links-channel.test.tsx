import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { IssueMessageLinksPanel } from "@/modules/tracker/issue-message-links-panel";
import { renderWithProviders } from "@testframework/render";

vi.mock("@/api/issue-message-links", () => ({
	listIssueMessageLinks: vi.fn(async () => [
		{
			id: "link-1",
			issue_id: "iss-1",
			channel_id: "channel-abc",
			message_id: "message-xyz-123456",
			mirror_to_thread: false,
			created_by_id: "u-1",
			created_at: "2026-07-28T12:00:00Z",
		},
	]),
	deleteIssueMessageLink: vi.fn(),
}));

vi.mock("@tanstack/react-router", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@tanstack/react-router")>();
	return {
		...actual,
		Link: ({
			to,
			params,
			children,
			...rest
		}: {
			to: string;
			params?: { channelId?: string };
			children: React.ReactNode;
			className?: string;
		}) => (
			<a
				href={
					params?.channelId
						? `/messenger/${params.channelId}`
						: String(to)
				}
				data-to={to}
				data-channel={params?.channelId}
				{...rest}
			>
				{children}
			</a>
		),
	};
});

describe("FE-P1-MESSENGER · issue→messenger channelId (W10-A2-002)", () => {
	it("links to /messenger/$channelId with channel params", async () => {
		renderWithProviders(
			<IssueMessageLinksPanel projectId="proj-1" issueKey="HUB-1" />,
		);

		const link = await screen.findByRole("link");
		expect(link.getAttribute("data-to")).toBe("/messenger/$channelId");
		expect(link.getAttribute("data-channel")).toBe("channel-abc");
		expect(link.getAttribute("href")).toBe("/messenger/channel-abc");
	});
});
