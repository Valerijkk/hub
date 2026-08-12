import { fireEvent, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CallSummary, PersonalData } from "@/api/dashboard";
import { UpcomingCallsWidget } from "@/components/dashboard/widgets/upcoming-calls";
import { useCallStore } from "@/stores/call-store";
import { renderWithProviders } from "@testframework/render";

function personalDataWithCall(call: CallSummary): PersonalData {
	return {
		my_tasks: [],
		unread_count: 0,
		upcoming_calls: [call],
		recent_runs: [],
		deadlines: [],
		time_today: 0,
	};
}

describe("UpcomingCallsWidget join (W9-A2-001)", () => {
	beforeEach(() => {
		useCallStore.getState().reset();
	});

	it("calls joinCall(call.id) on Войти, not startCall", () => {
		const joinCall = vi.fn(async () => undefined);
		const startCall = vi.fn(async () => undefined);
		useCallStore.setState({ joinCall, startCall });

		renderWithProviders(
			<UpcomingCallsWidget
				data={personalDataWithCall({
					id: "call-xyz",
					title: "Стендап",
					scheduled_at: new Date().toISOString(),
					channel_id: "channel-dev",
					channel_name: "dev",
				})}
			/>,
		);

		fireEvent.click(screen.getByRole("button", { name: /Войти/i }));

		expect(joinCall).toHaveBeenCalledWith("call-xyz");
		expect(startCall).not.toHaveBeenCalled();
	});

	it("joins even when channel_id is missing from dashboard DTO", () => {
		const joinCall = vi.fn(async () => undefined);
		useCallStore.setState({ joinCall, startCall: vi.fn() });

		renderWithProviders(
			<UpcomingCallsWidget
				data={personalDataWithCall({
					id: "call-no-channel",
					title: "Ретро",
					scheduled_at: new Date().toISOString(),
					channel_name: "retro",
				})}
			/>,
		);

		fireEvent.click(screen.getByRole("button", { name: /Войти/i }));
		expect(joinCall).toHaveBeenCalledWith("call-no-channel");
	});
});
