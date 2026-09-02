import { fireEvent, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type {
	CalendarEventAttendeeResponse,
	CalendarEventDetail,
} from "@/api/calendar";
import * as calendarApi from "@/api/calendar";
import { MailCalendarInvitation } from "@/modules/mail/mail-calendar-invitation";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/calendar");
vi.mock("@tanstack/react-router", () => ({
	Link: ({ children }: { children: ReactNode }) => (
		<a href="/calendar/events/11111111-1111-4111-8111-111111111111">
			{children}
		</a>
	),
}));
vi.mock("react-i18next", async (importOriginal) => ({
	...(await importOriginal<typeof import("react-i18next")>()),
	useTranslation: () => ({
		t: (key: string, options?: { status?: string }) => {
			const messages: Record<string, string> = {
				"reader.calendar.request": "Calendar invitation",
				"reader.calendar.openEvent": "Open event",
				"reader.calendar.rsvp.unavailable": "Invitation details are unavailable.",
				"reader.calendar.rsvp.failed": "Could not save your response. Try again.",
				"reader.calendar.rsvp.saving": "Saving response…",
				"reader.calendar.rsvp.actions.accepted": "Accept",
				"reader.calendar.rsvp.actions.tentative": "Tentative",
				"reader.calendar.rsvp.actions.declined": "Decline",
				"reader.calendar.rsvp.status.needs_action": "Needs action",
				"reader.calendar.rsvp.status.accepted": "Accepted",
				"reader.calendar.rsvp.status.tentative": "Tentative",
				"reader.calendar.rsvp.status.declined": "Declined",
			};
			if (key === "reader.calendar.rsvp.current") {
				return `Your response: ${options?.status ?? ""}`;
			}
			return messages[key] ?? key;
		},
	}),
}));

const eventID = "11111111-1111-4111-8111-111111111111";

const invitation: CalendarEventDetail = {
	id: eventID,
	title: "Project kickoff",
	description: "",
	location: "",
	timezone: "Europe/Moscow",
	rrule: null,
	exdates: [],
	sequence: 1,
	status: "confirmed",
	version: 1,
	organizer_id: "22222222-2222-4222-8222-222222222222",
	organizer_email: "organizer@hub.test",
	attendees: [
		{
			id: "33333333-3333-4333-8333-333333333333",
			user_id: "44444444-4444-4444-8444-444444444444",
			email: "viewer@hub.test",
			role: "required",
			rsvp_status: "needs_action",
			version: 7,
			responded_at: null,
		},
	],
	viewer_attendee: {
		id: "33333333-3333-4333-8333-333333333333",
		user_id: "44444444-4444-4444-8444-444444444444",
		email: "viewer@hub.test",
		role: "required",
		rsvp_status: "needs_action",
		version: 7,
		responded_at: null,
	},
	can_rsvp: true,
	mail_message_id: null,
	all_day: false,
	starts_at: "2026-09-02T10:00:00Z",
	ends_at: "2026-09-02T11:00:00Z",
	start_date: null,
	end_date_exclusive: null,
};

const rsvpSuppressionCases: ReadonlyArray<{
	name: string;
	event: CalendarEventDetail;
}> = [
	{
		name: "the server disallows RSVP",
		event: { ...invitation, can_rsvp: false },
	},
	{
		name: "the event is cancelled",
		event: { ...invitation, status: "cancelled" },
	},
	{
		name: "the viewer is not an attendee",
		event: { ...invitation, viewer_attendee: null },
	},
];

function attendeeResponse(): CalendarEventAttendeeResponse {
	return {
		id: invitation.viewer_attendee?.id ?? "",
		user_id: invitation.viewer_attendee?.user_id ?? null,
		email: invitation.viewer_attendee?.email ?? "",
		display_name: null,
		role: "required",
		rsvp_status: "accepted",
		recurrence_id: null,
		responded_at: "2026-09-01T12:00:00Z",
		version: 8,
		reply_delivery_status: "pending",
	};
}

function deferred<T>() {
	let resolve!: (value: T) => void;
	let reject!: (error: unknown) => void;
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise;
		reject = rejectPromise;
	});
	return { promise, resolve, reject };
}

describe("MailCalendarInvitation", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.stubGlobal("crypto", { randomUUID: vi.fn(() => "rsvp-request-id") });
		vi.mocked(calendarApi.getCalendarEvent).mockResolvedValue(invitation);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it("submits the selected RSVP with the attendee version and disables every choice while saving", async () => {
		const response = deferred<CalendarEventAttendeeResponse>();
		vi.mocked(calendarApi.respondToCalendarSeries).mockReturnValue(response.promise);

		const { queryClient } = renderWithProviders(
			<MailCalendarInvitation eventId={eventID} />,
		);
		const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");

		expect(await screen.findByText("Your response: Needs action")).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "Open event" })).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Accept" }));

		await waitFor(() => {
			expect(calendarApi.respondToCalendarSeries).toHaveBeenCalledWith(
				eventID,
				{ status: "accepted", attendee_version: 7 },
				{ idempotencyKey: "rsvp-request-id" },
			);
		});
		expect(screen.getByRole("button", { name: "Accept" })).toBeDisabled();
		expect(screen.getByRole("button", { name: "Tentative" })).toBeDisabled();
		expect(screen.getByRole("button", { name: "Decline" })).toBeDisabled();

		response.resolve(attendeeResponse());

		await waitFor(() => {
			expect(invalidateQueries).toHaveBeenCalledWith({
				queryKey: ["calendar", "event", eventID],
				exact: true,
			});
		});
		expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ["calendar"] });
	});

	it.each(rsvpSuppressionCases)(
		"suppresses RSVP controls when $name while retaining the event link",
		async ({ event }) => {
			vi.mocked(calendarApi.getCalendarEvent).mockResolvedValue(event);

			const { queryClient } = renderWithProviders(
				<MailCalendarInvitation eventId={eventID} />,
			);

			await waitFor(() => {
				expect(
					queryClient.getQueryState(["calendar", "event", eventID])?.status,
				).toBe("success");
			});
			expect(
				screen.getByRole("link", { name: "Open event" }),
			).toBeInTheDocument();
			expect(
				screen.queryByRole("button", { name: "Accept" }),
			).not.toBeInTheDocument();
			expect(
				screen.queryByRole("button", { name: "Tentative" }),
			).not.toBeInTheDocument();
			expect(
				screen.queryByRole("button", { name: "Decline" }),
			).not.toBeInTheDocument();
		},
	);

	it("keeps the event link available when invitation details cannot be loaded", async () => {
		vi.mocked(calendarApi.getCalendarEvent).mockRejectedValue(
			new Error("calendar transport unavailable"),
		);

		renderWithProviders(<MailCalendarInvitation eventId={eventID} />);

		expect(screen.getByRole("link", { name: "Open event" })).toBeInTheDocument();
		expect(await screen.findByRole("alert")).toHaveTextContent(
			"Invitation details are unavailable.",
		);
		expect(screen.queryByRole("button", { name: "Accept" })).not.toBeInTheDocument();
	});

	it("shows a retryable generic error without changing the displayed RSVP after a non-conflict failure", async () => {
		vi.mocked(calendarApi.respondToCalendarSeries).mockRejectedValue(
			new Error("response delivery failed"),
		);

		renderWithProviders(<MailCalendarInvitation eventId={eventID} />);

		fireEvent.click(await screen.findByRole("button", { name: "Decline" }));

		expect(await screen.findByRole("alert")).toHaveTextContent(
			"Could not save your response. Try again.",
		);
		expect(screen.getByText("Your response: Needs action")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Decline" })).toBeEnabled();

		fireEvent.click(screen.getByRole("button", { name: "Decline" }));
		await waitFor(() => {
			expect(calendarApi.respondToCalendarSeries).toHaveBeenCalledTimes(2);
		});
	});

	it("refreshes the calendar cache on a conflict without presenting a stale mutation failure", async () => {
		vi.mocked(calendarApi.respondToCalendarSeries).mockRejectedValue({ code: 409 });

		const { queryClient } = renderWithProviders(
			<MailCalendarInvitation eventId={eventID} />,
		);
		const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");

		fireEvent.click(await screen.findByRole("button", { name: "Tentative" }));

		await waitFor(() => {
			expect(invalidateQueries).toHaveBeenCalledWith({
				queryKey: ["calendar", "event", eventID],
				exact: true,
			});
		});
		expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ["calendar"] });
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
