import { QueryClient } from "@tanstack/react-query";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
	getChannelPins,
	togglePin,
	type PinnedMessage,
} from "@/api/messenger";
import { PinnedMessages } from "@/components/messenger/pinned-messages";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/messenger", () => ({
	getChannelPins: vi.fn(),
	togglePin: vi.fn(),
}));

const channelId = "pins-query-state-channel";
const pinCapabilities = { can_list_pins: true, can_pin: true } as const;
const user = {
	id: "pins-query-state-user",
	email: "pins-query-state@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

function createQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

function openPinnedMessages() {
	fireEvent.click(
		screen.getByRole("button", { name: "Закреплённые сообщения" }),
	);
}

describe("Phase 151 pinned-message query states", () => {
	const queryClients: QueryClient[] = [];

	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(getChannelPins).mockReset();
		vi.mocked(togglePin).mockReset().mockResolvedValue({ pinned: false });
		useAuthStore.setState({ user });
	});

	afterEach(() => {
		for (const queryClient of queryClients.splice(0)) {
			queryClient.clear();
		}
		useAuthStore.setState({ user: null });
	});

	it("shows loading instead of empty while the authenticated pins query is pending", async () => {
		const pending = new Promise<PinnedMessage[]>(() => undefined);
		vi.mocked(getChannelPins).mockReturnValue(pending);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();

		expect(
			await screen.findByRole("dialog", { name: "Закреплённые сообщения" }),
		).toBeInTheDocument();
		await waitFor(() => expect(getChannelPins).toHaveBeenCalledTimes(1));
		expect(await screen.findByRole("status")).toHaveTextContent(
			"Загрузка закреплённых сообщений...",
		);
		expect(
			screen.queryByText("Нет закреплённых сообщений"),
		).not.toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("shows an alert and retries instead of empty when the authenticated pins query fails", async () => {
		vi.mocked(getChannelPins)
			.mockRejectedValueOnce(new Error("network"))
			.mockResolvedValueOnce([]);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();

		const alert = await screen.findByRole("alert");
		expect(alert).toHaveTextContent("Не удалось загрузить закреплённые сообщения");
		expect(
			screen.queryByText("Нет закреплённых сообщений"),
		).not.toBeInTheDocument();
		fireEvent.click(screen.getByRole("button", { name: "Попробовать снова" }));

		await waitFor(() => expect(getChannelPins).toHaveBeenCalledTimes(2));
		expect(
			await screen.findByText("Нет закреплённых сообщений"),
		).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("shows empty only for an authenticated successful empty pins response", async () => {
		vi.mocked(getChannelPins).mockResolvedValueOnce([]);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);

		renderWithProviders(
			<PinnedMessages channelId={channelId} capabilities={pinCapabilities} />,
			{ queryClient },
		);
		openPinnedMessages();

		await waitFor(() => expect(getChannelPins).toHaveBeenCalledTimes(1));
		expect(
			await screen.findByText("Нет закреплённых сообщений"),
		).toBeInTheDocument();
		expect(screen.queryByRole("status")).not.toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "Открепить" }),
		).not.toBeInTheDocument();
	});
});
