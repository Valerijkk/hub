import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as intranetApi from "@/api/intranet";
import { NewsFeed } from "@/modules/intranet/news-feed";
import { useAuthStore } from "@/stores/auth-store";

vi.mock("@/api/intranet");

function renderFeed() {
	const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={qc}>
			<NewsFeed />
		</QueryClientProvider>,
	);
}

describe("NewsFeed", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "u1",
				email: "a@hub.local",
				system_role: "member",
			} as never,
		});
		vi.mocked(intranetApi.listNews).mockResolvedValue({
			items: [
				{
					id: "n1",
					title: "All-hands",
					body: "Friday 15:00",
					pinned: true,
					published: true,
					version: 1,
					author_id: "admin",
					author_name: "Admin",
					created_at: "2026-06-24T10:00:00Z",
					updated_at: "2026-06-24T10:00:00Z",
				},
			],
			total: 1,
			page: 1,
			per_page: 50,
		});
	});

	it("renders news items from API", async () => {
		renderFeed();
		expect(await screen.findByText("All-hands")).toBeInTheDocument();
		expect(screen.getByText(/Friday 15:00/)).toBeInTheDocument();
	});

	it("shows admin compose form for admin users", async () => {
		useAuthStore.setState({
			user: {
				id: "admin",
				email: "admin@hub.local",
				system_role: "admin",
			} as never,
		});
		renderFeed();
		await waitFor(() => {
			expect(screen.getByText("Новая публикация")).toBeInTheDocument();
		});
	});
});
