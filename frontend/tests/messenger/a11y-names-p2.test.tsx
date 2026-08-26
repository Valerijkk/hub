import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { LanguageSwitcher } from "@/components/language-switcher";
import { PinnedMessages } from "@/components/messenger/pinned-messages";
import { ReactionPicker } from "@/components/messenger/reaction-picker";
import { SearchDialog } from "@/components/messenger/search-dialog";
import { TooltipProvider } from "@/components/ui/tooltip";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/messenger", () => ({
	getChannelPins: vi.fn().mockResolvedValue([]),
	searchMessages: vi.fn().mockResolvedValue({ messages: [], next_cursor: null }),
	togglePin: vi.fn(),
	toggleReaction: vi.fn(),
}));

vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

/**
 * P2-FE-03 — accessible names on high-traffic messenger/layout controls.
 * Mutation: remove aria-label → getByRole name empty / query fails.
 */
describe("P2-FE-03 · accessible names (TASK-049)", () => {
	it("PinnedMessages trigger has accessible name", () => {
		renderWithProviders(
			<PinnedMessages
				channelId="ch-1"
				capabilities={{ can_list_pins: true, can_pin: true }}
			/>
		);
		expect(
			screen.getByRole("button", { name: "Закреплённые сообщения" }),
		).toBeInTheDocument();
	});

	it("SearchDialog trigger has accessible name", () => {
		renderWithProviders(<SearchDialog channelId="ch-1" />);
		expect(
			screen.getByRole("button", { name: "Поиск сообщений" }),
		).toBeInTheDocument();
	});

	it("ReactionPicker emoji grid buttons expose aria-label when open", () => {
		render(
			<TooltipProvider>
				<ReactionPicker
					channelId="ch-1"
					messageId="m-1"
					canReact
					open
					onOpenChange={() => undefined}
				/>
			</TooltipProvider>,
		);

		expect(screen.getByRole("button", { name: "Реакция" })).toBeInTheDocument();
		// First popular emoji (thumbs up) — glyph as accessible name
		expect(screen.getByRole("button", { name: "👍" })).toBeInTheDocument();
	});

	it("LanguageSwitcher compact exposes aria-pressed and group label", () => {
		render(<LanguageSwitcher compact />);

		expect(screen.getByRole("group", { name: "Язык" })).toBeInTheDocument();
		const ru = screen.getByRole("button", { name: "Русский" });
		const en = screen.getByRole("button", { name: "English" });
		expect(ru).toHaveAttribute("aria-pressed");
		expect(en).toHaveAttribute("aria-pressed");
		// Exactly one language is pressed at a time.
		const pressed = [ru, en].filter(
			(el) => el.getAttribute("aria-pressed") === "true",
		);
		expect(pressed).toHaveLength(1);
	});
});
