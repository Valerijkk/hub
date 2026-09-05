import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	cancelCalendarOccurrence,
	respondToCalendarOccurrence,
} from "@/api/calendar";
import { api } from "@/api/client";

vi.mock("@/api/client", () => ({
	api: {
		post: vi.fn(),
		delete: vi.fn(),
	},
}));

const eventID = "11111111-1111-4111-8111-111111111111";

const recurrenceCases = [
	{
		name: "a canonical date",
		recurrenceID: "2026-09-10",
		encodedRecurrenceID: "MjAyNi0wOS0xMA",
	},
	{
		name: "a UTC RFC3339 timestamp with seconds",
		recurrenceID: "2026-09-10T10:00:00Z",
		encodedRecurrenceID: "MjAyNi0wOS0xMFQxMDowMDowMFo",
	},
] as const;

describe("calendar occurrence API", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it.each(recurrenceCases)(
		"forwards RSVP and cancellation through base64url for $name",
		async ({ recurrenceID, encodedRecurrenceID }) => {
			const rsvpResponse = {} as never;
			const cancellationResponse = {} as never;
			const rsvpRequest = { status: "accepted" as const, attendee_version: 7 };
			const rsvpOptions = { idempotencyKey: "calendar-occurrence-rsvp" };
			const cancellationOptions = { ifMatch: 9 };
			const postSpy = vi
				.mocked(api.post)
				.mockResolvedValueOnce(rsvpResponse);
			const deleteSpy = vi
				.mocked(api.delete)
				.mockResolvedValueOnce(cancellationResponse);
			const occurrencePath = `/calendar/v2/events/${eventID}/instances/${encodedRecurrenceID}`;

			await expect(
				respondToCalendarOccurrence(
					eventID,
					recurrenceID,
					rsvpRequest,
					rsvpOptions,
				),
			).resolves.toBe(rsvpResponse);
			await expect(
				cancelCalendarOccurrence(
					eventID,
					recurrenceID,
					cancellationOptions,
				),
			).resolves.toBe(cancellationResponse);

			expect(postSpy).toHaveBeenCalledWith(
				`${occurrencePath}/rsvp`,
				rsvpRequest,
				rsvpOptions,
			);
			expect(deleteSpy).toHaveBeenCalledWith(
				occurrencePath,
				cancellationOptions,
			);

			const rsvpSegment = postSpy.mock.calls[0]?.[0]?.split("/")[6];
			const cancellationSegment = deleteSpy.mock.calls[0]?.[0]?.split("/")[6];
			expect(rsvpSegment).toBe(encodedRecurrenceID);
			expect(cancellationSegment).toBe(encodedRecurrenceID);
			expect(rsvpSegment).not.toMatch(/[+/=]/);
			expect(cancellationSegment).not.toMatch(/[+/=]/);
		},
	);
});
