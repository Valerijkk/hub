import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CallPanel } from "@/components/calls/CallPanel";
import { CallReactionBar } from "@/components/calls/CallReactions";
import { CreateChannelDialogContent } from "@/components/messenger/create-channel-dialog";
import { GifPicker } from "@/components/messenger/gif-picker";
import { StickerPicker } from "@/components/messenger/sticker-picker";
import { Checklist } from "@/components/tracker/checklist";
import { useCallStore } from "@/stores/call-store";
import { renderWithProviders } from "../../render";

vi.mock("@/components/calls/CallRoom", () => ({
	CallRoom: () => <div data-testid="call-room" />,
}));

vi.mock("@livekit/components-react", () => ({
	useRoomContext: () => ({
		localParticipant: {
			name: "Test",
			identity: "u1",
			publishData: vi.fn().mockResolvedValue(undefined),
		},
	}),
}));

vi.mock("@/api/messenger", () => ({
	createChannel: vi.fn(),
}));

vi.mock("@/hooks/use-invalidate", () => ({
	useInvalidateChannels: () => vi.fn(),
}));

/**
 * P3-FE-01 — residual accessible names.
 * Mutation: remove aria-label / Label → getByRole({ name }) fails.
 */
describe("P3-FE-01 · accessible names residual (TASK-069)", () => {
	beforeEach(() => {
		useCallStore.setState({
			isInCall: false,
			isConnecting: false,
			token: null,
			error: null,
		});
	});

	it("CallPanel fullscreen control exposes aria-label", () => {
		useCallStore.setState({
			isInCall: true,
			isConnecting: false,
			token: "tok",
			error: null,
		});
		render(<CallPanel />);
		expect(
			screen.getByRole("button", { name: "На весь экран" }),
		).toBeInTheDocument();
	});

	it("CallReactionBar emoji buttons expose aria-label", () => {
		render(<CallReactionBar onClose={() => undefined} />);
		expect(screen.getByRole("button", { name: "Like" })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Heart" })).toBeInTheDocument();
	});

	it("GifPicker search input has accessible name when open", () => {
		render(<GifPicker onSelect={() => undefined} />);
		fireEvent.click(screen.getByRole("button", { name: "Выбрать GIF" }));
		expect(
			screen.getByRole("textbox", { name: "Поиск GIF" }),
		).toBeInTheDocument();
	});

	it("StickerPicker pack tabs expose tablist and aria-label", () => {
		render(<StickerPicker onSelect={() => undefined} />);
		fireEvent.click(screen.getByRole("button", { name: "Выбрать стикер" }));
		expect(screen.getByRole("tablist")).toBeInTheDocument();
		expect(screen.getByRole("tab", { name: "Эмоции" })).toHaveAttribute(
			"aria-selected",
			"true",
		);
		expect(screen.getByRole("button", { name: "Круто" })).toBeInTheDocument();
	});

	it("CreateChannelDialogContent name field is labelled", () => {
		renderWithProviders(<CreateChannelDialogContent />);
		expect(screen.getByLabelText("Название канала")).toBeInTheDocument();
	});

	it("Checklist delete control has aria-label", () => {
		render(
			<Checklist
				items={[{ id: "1", text: "Пункт", checked: false }]}
				onChange={() => undefined}
			/>,
		);
		expect(
			screen.getByRole("button", { name: "Удалить пункт" }),
		).toBeInTheDocument();
	});
});
