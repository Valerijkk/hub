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
				"reader.calendar.rsvp.unavailable":
					"Invitation details are unavailable.",
				"reader.calendar.rsvp.failed":
					"Could not save your response. Try again.",
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
const occurrenceID = "2026-09-09T10:00:00Z";
const seriesRSVPTarget = {
	event_id: eventID,
	scope: "series" as const,
	recurrence_id: null,
	attendee_version: 7,
	rsvp_status: "needs_action" as const,
};
const occurrenceRSVPTarget = {
	event_id: eventID,
	scope: "occurrence" as const,
	recurrence_id: occurrenceID,
	attendee_version: 12,
	rsvp_status: "tentative" as const,
};

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

function occurrenceAttendeeResponse(): CalendarEventAttendeeResponse {
	return {
		...attendeeResponse(),
		recurrence_id: occurrenceID,
		version: 12,
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
		vi.mocked(calendarApi.isCanonicalCalendarEventID).mockImplementation(
			(value) =>
				/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(
					value,
				) && value !== "00000000-0000-0000-0000-000000000000",
		);
		vi.mocked(calendarApi.getCalendarEvent).mockResolvedValue(invitation);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it("uses the series endpoint for an authoritative series target and disables every choice while saving", async () => {
		const response = deferred<CalendarEventAttendeeResponse>();
		vi.mocked(calendarApi.respondToCalendarSeries).mockReturnValue(
			response.promise,
		);

		const { queryClient } = renderWithProviders(
			<MailCalendarInvitation
				eventId={eventID}
				rsvpTarget={seriesRSVPTarget}
			/>,
		);
		const invalidateQueries = vi.spyOn(queryClient, "invalidateQueries");

		expect(
			await screen.findByText("Your response: Needs action"),
		).toBeInTheDocument();
		expect(
			screen.getByRole("link", { name: "Open event" }),
		).toBeInTheDocument();

		fireEvent.click(await screen.findByRole("button", { name: "Accept" }));

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
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: ["mail", "message"],
		});
	});

	it("uses the occurrence endpoint and keeps its returned attendee state local to that occurrence", async () => {
		vi.stubGlobal("crypto", {
			randomUUID: vi
				.fn()
				.mockReturnValueOnce("occurrence-rsvp-request-id")
				.mockReturnValueOnce("occurrence-follow-up-request-id"),
		});
		vi.mocked(calendarApi.respondToCalendarOccurrence)
			.mockResolvedValueOnce({
				...occurrenceAttendeeResponse(),
				version: 13,
			})
			.mockResolvedValueOnce({
				...occurrenceAttendeeResponse(),
				rsvp_status: "declined",
				version: 14,
			});

		renderWithProviders(
			<MailCalendarInvitation
				eventId={eventID}
				rsvpTarget={occurrenceRSVPTarget}
			/>,
		);

		fireEvent.click(await screen.findByRole("button", { name: "Accept" }));

		await waitFor(() => {
			expect(calendarApi.respondToCalendarOccurrence).toHaveBeenCalledWith(
				eventID,
				occurrenceID,
				{ status: "accepted", attendee_version: 12 },
				{ idempotencyKey: "occurrence-rsvp-request-id" },
			);
		});
		expect(calendarApi.respondToCalendarSeries).not.toHaveBeenCalled();
		expect(
			await screen.findByText("Your response: Accepted"),
		).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Decline" }));
		await waitFor(() => {
			expect(calendarApi.respondToCalendarOccurrence).toHaveBeenLastCalledWith(
				eventID,
				occurrenceID,
				{ status: "declined", attendee_version: 13 },
				{ idempotencyKey: "occurrence-follow-up-request-id" },
			);
		});
	});

	it("keeps a legacy or non-actionable invitation navigable without exposing RSVP controls", async () => {
		renderWithProviders(
			<MailCalendarInvitation eventId={eventID} rsvpTarget={null} />,
		);

		expect(
			await screen.findByRole("link", { name: "Open event" }),
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
		expect(calendarApi.respondToCalendarSeries).not.toHaveBeenCalled();
		expect(calendarApi.respondToCalendarOccurrence).not.toHaveBeenCalled();
	});

	it("keeps an invitation with an undefined RSVP target navigable without exposing RSVP controls", async () => {
		renderWithProviders(
			<MailCalendarInvitation eventId={eventID} rsvpTarget={undefined} />,
		);

		expect(
			await screen.findByRole("link", { name: "Open event" }),
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
		expect(calendarApi.respondToCalendarSeries).not.toHaveBeenCalled();
		expect(calendarApi.respondToCalendarOccurrence).not.toHaveBeenCalled();
	});

	it.each(
		rsvpSuppressionCases,
	)("suppresses RSVP controls when $name while retaining the event link", async ({
		event,
	}) => {
		vi.mocked(calendarApi.getCalendarEvent).mockResolvedValue(event);

		const { queryClient } = renderWithProviders(
			<MailCalendarInvitation
				eventId={eventID}
				rsvpTarget={seriesRSVPTarget}
			/>,
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
	});

	it("keeps the event link available when invitation details cannot be loaded", async () => {
		vi.mocked(calendarApi.getCalendarEvent).mockRejectedValue(
			new Error("calendar transport unavailable"),
		);

		renderWithProviders(
			<MailCalendarInvitation
				eventId={eventID}
				rsvpTarget={seriesRSVPTarget}
			/>,
		);

		expect(
			screen.getByRole("link", { name: "Open event" }),
		).toBeInTheDocument();
		expect(await screen.findByRole("alert")).toHaveTextContent(
			"Invitation details are unavailable.",
		);
		expect(
			screen.queryByRole("button", { name: "Accept" }),
		).not.toBeInTheDocument();
	});

	it("shows a retryable generic error without changing the displayed RSVP after a non-conflict failure", async () => {
		vi.mocked(calendarApi.respondToCalendarSeries).mockRejectedValue(
			new Error("response delivery failed"),
		);

		renderWithProviders(
			<MailCalendarInvitation
				eventId={eventID}
				rsvpTarget={seriesRSVPTarget}
			/>,
		);

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
		vi.mocked(calendarApi.respondToCalendarSeries).mockRejectedValue({
			code: 409,
		});

		const { queryClient } = renderWithProviders(
			<MailCalendarInvitation
				eventId={eventID}
				rsvpTarget={seriesRSVPTarget}
			/>,
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
		expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ["mail", "message"] });
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
