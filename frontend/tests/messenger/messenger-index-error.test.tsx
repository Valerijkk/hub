import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MessengerIndex } from "@/routes/_authenticated/messenger/index";
import { renderWithProviders } from "../support/render";

const refetch = vi.fn();

vi.mock("@/hooks/use-channels", () => ({
	useChannels: vi.fn(),
}));

vi.mock("@/components/messenger/channel-drawer-context", () => ({
	useChannelDrawer: () => ({ setOpen: vi.fn() }),
}));

vi.mock("@tanstack/react-router", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-router")>(
		"@tanstack/react-router",
	);
	return {
		...actual,
		createFileRoute: () => (opts: { component: React.ComponentType }) => ({
			...opts,
			useParams: () => ({}),
		}),
		useNavigate: () => vi.fn(),
	};
});

import { useChannels } from "@/hooks/use-channels";

describe("MessengerIndex error≠empty (A3W10-4-001 / TASK-127)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of welcoming empty when channels fail", () => {
		vi.mocked(useChannels).mockReturnValue({
			data: undefined,
			isLoading: false,
			isError: true,
			refetch,
		} as ReturnType<typeof useChannels>);

		renderWithProviders(<MessengerIndex />);

		expect(screen.getByRole("alert")).toHaveTextContent(
			/не удалось загрузить каналы/i,
		);
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText(/добро пожаловать в мессенджер/i),
		).not.toBeInTheDocument();
		expect(screen.queryByText(/у вас пока нет каналов/i)).not.toBeInTheDocument();
	});
});
