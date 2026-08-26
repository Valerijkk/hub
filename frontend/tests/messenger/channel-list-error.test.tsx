import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel } from "@/api/messenger";
import { ChannelList } from "@/components/messenger/channel-list";
import { queryError } from "../support/query-result";
import { renderWithProviders } from "../support/render";

const refetch = vi.fn();

vi.mock("@/hooks/use-channels", () => ({
	useChannels: vi.fn(),
	useDiscoverableChannels: () => ({
		data: [],
		isLoading: false,
		isFetching: false,
		isError: false,
	}),
}));

vi.mock("@/hooks/use-unread", () => ({
	useUnread: vi.fn(),
}));

vi.mock("@/hooks/use-presence", () => ({
	usePresence: () => ({
		getStatus: () => "offline",
		isError: false,
		isLoading: false,
		refetch: vi.fn(),
	}),
}));

vi.mock("@tanstack/react-router", () => ({
	useMatchRoute: () => () => false,
	useNavigate: () => vi.fn(),
	Link: ({ children, ...props }: { children: React.ReactNode }) => (
		<a {...props}>{children}</a>
	),
}));

import { useChannels } from "@/hooks/use-channels";

describe("ChannelList error state (A10-005)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of empty list when channels query fails", () => {
		vi.mocked(useChannels).mockReturnValue(
			queryError<Channel[]>(new Error("network"), refetch),
		);

		renderWithProviders(<ChannelList />);

		expect(screen.getByRole("alert")).toHaveTextContent(
			/не удалось загрузить каналы/i,
		);
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByText(/нет каналов/i)).not.toBeInTheDocument();
	});
});
