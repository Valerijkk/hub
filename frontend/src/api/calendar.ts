import { api, type ApiRequestOptions } from "./client";

export type CalendarEventKind = "call" | "event" | "sprint" | "deadline" | "leave";

export interface CalendarEventDTO {
	id: string;
	type: CalendarEventKind;
	title: string;
	description?: string;
	starts_at: string;
	ends_at?: string;
	all_day: boolean;
	project_id?: string;
	project_key?: string;
	url?: string;
}

export interface CalendarEventsResponse {
	events: CalendarEventDTO[];
	from: string;
	to: string;
}

export interface CalendarAvailabilityRequest {
	user_ids: string[];
	from: string;
	to: string;
	timezone: string;
	slot_minutes: number;
}

export interface CalendarAvailabilityBusyInterval {
	from: string;
	to: string;
}

export interface CalendarAvailabilityFreeInterval {
	from: string;
	to: string;
}

export interface CalendarAvailabilityUserResponse {
	user_id: string;
	busy: CalendarAvailabilityBusyInterval[];
	free: CalendarAvailabilityFreeInterval[];
}

export interface CalendarAvailabilityResponse {
	users: CalendarAvailabilityUserResponse[];
}

export interface CalendarAvailabilityCheckRequest {
	attendee_user_ids: string[];
	from: string;
	to: string;
	timezone: string;
}

export interface CalendarAvailabilityCheckResponse {
	all_available: boolean;
}

export interface CalendarMeetingCreateRequest {
	title: string;
	description: string;
	project_id: string;
	channel_id: string;
	starts_at: string;
	ends_at: string;
	timezone: string;
	rrule: string;
	attendee_user_ids: string[];
}

export interface CalendarMeetingCreateResponse {
	event_id: string;
	scheduled_call_id: string;
	project_id: string;
	channel_id: string;
	organizer_id: string;
	attendee_user_ids: string[];
	title: string;
	description?: string;
	starts_at: string;
	ends_at: string;
	timezone: string;
	rrule?: string;
	status: "confirmed";
	event_version: number;
	scheduled_call_version: number;
}

export function listCalendarEvents(
	from: string,
	to: string,
): Promise<CalendarEventsResponse> {
	return api.get<CalendarEventsResponse>(
		`/calendar/events?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
	);
}

export function getCalendarAvailability(
	request: CalendarAvailabilityRequest,
): Promise<CalendarAvailabilityResponse> {
	return api.post<CalendarAvailabilityResponse>(
		"/calendar/v2/availability",
		request,
	);
}

export function checkCalendarAvailability(
	request: CalendarAvailabilityCheckRequest,
): Promise<CalendarAvailabilityCheckResponse> {
	return api.post<CalendarAvailabilityCheckResponse>(
		"/calendar/v2/availability/check",
		request,
	);
}

/** Скачивание .ics через api.getBlob (Bearer + one-retry refresh). */
export async function downloadCalendarExport(
	from: string,
	to: string,
): Promise<void> {
	const blob = await api.getBlob(
		`/calendar/export.ics?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
	);
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = "hub-calendar.ics";
	a.click();
	URL.revokeObjectURL(url);
}

export type CalendarEventAttendeeRole = "required" | "optional";
export type CalendarEventReadAttendeeRole =
	| CalendarEventAttendeeRole
	| "organizer";
export type CalendarEventRSVPStatus =
	| "needs_action"
	| "accepted"
	| "declined"
	| "tentative";
export type CalendarEventRSVPChoice = Exclude<
	CalendarEventRSVPStatus,
	"needs_action"
>;
export type CalendarEventReplyDeliveryStatus =
	| "not_required"
	| "pending"
	| "sent"
	| "failed";

export interface CalendarEventAttendeeRequest {
	user_id?: string | null;
	email: string;
	display_name?: string | null;
	role: CalendarEventAttendeeRole;
}

export interface CalendarEventAttendeeResponse {
	id: string;
	user_id: string | null;
	email: string;
	display_name: string | null;
	role: CalendarEventAttendeeRole;
	rsvp_status: CalendarEventRSVPStatus;
	recurrence_id: string | null;
	responded_at: string | null;
	version: number;
	reply_delivery_status: CalendarEventReplyDeliveryStatus;
}

export interface CalendarEventReadAttendee {
	id: string;
	user_id: string | null;
	email: string;
	role: CalendarEventReadAttendeeRole;
	rsvp_status: CalendarEventRSVPStatus;
	version: number;
	responded_at: string | null;
}

interface CalendarEventCreateRequestBase {
	title: string;
	description?: string;
	location?: string;
	timezone: string;
	rrule?: string | null;
	exdates?: string[];
	scheduled_call_id?: string | null;
	reminder_minutes?: number | null;
	attendees: CalendarEventAttendeeRequest[];
}

export type CalendarEventCreateRequest =
	| (CalendarEventCreateRequestBase & {
			all_day: false;
			starts_at: string;
			ends_at: string;
			start_date: null;
			end_date_exclusive: null;
		})
	| (CalendarEventCreateRequestBase & {
			all_day: true;
			starts_at: null;
			ends_at: null;
			start_date: string;
			end_date_exclusive: string;
		});

interface CalendarEventCreateResponseBase {
	id: string;
	uid: string;
	title: string;
	description?: string;
	location?: string;
	timezone: string;
	rrule?: string;
	exdates?: string[];
	sequence: number;
	source: "hub";
	status: "confirmed";
	organizer_id: string;
	organizer_email: string;
	scheduled_call_id: string | null;
	reminder_minutes: number | null;
	attendees: CalendarEventAttendeeResponse[];
	version: number;
}

export type CalendarEventCreateResponse =
	| (CalendarEventCreateResponseBase & {
			all_day: false;
			starts_at: string;
			ends_at: string;
			start_date: null;
			end_date_exclusive: null;
		})
	| (CalendarEventCreateResponseBase & {
			all_day: true;
			starts_at: null;
			ends_at: null;
			start_date: string;
			end_date_exclusive: string;
		});

interface CalendarEventUpdateRequestBase {
	title?: string;
	description?: string;
	location?: string;
	timezone?: string;
	rrule?: string | null;
	exdates?: string[];
	reminder_minutes?: number | null;
	attendees?: CalendarEventAttendeeRequest[];
}

type CalendarEventUpdateTemporalFields =
	| {
			all_day?: never;
			starts_at?: never;
			ends_at?: never;
			start_date?: never;
			end_date_exclusive?: never;
		}
	| {
			all_day?: never;
			starts_at: string | null;
			ends_at: string | null;
			start_date?: never;
			end_date_exclusive?: never;
		}
	| {
			all_day?: never;
			starts_at?: never;
			ends_at?: never;
			start_date: string | null;
			end_date_exclusive: string | null;
		}
	| {
			all_day?: never;
			starts_at: string | null;
			ends_at: string | null;
			start_date: string | null;
			end_date_exclusive: string | null;
		}
	| {
			all_day: false;
			starts_at: string;
			ends_at: string;
			start_date: null;
			end_date_exclusive: null;
		}
	| {
			all_day: true;
			starts_at: null;
			ends_at: null;
			start_date: string;
			end_date_exclusive: string;
		};

export type CalendarEventUpdateRequest = CalendarEventUpdateRequestBase &
	CalendarEventUpdateTemporalFields;

export type CalendarEventChangedField =
	| "title"
	| "description"
	| "location"
	| "starts_at"
	| "ends_at"
	| "start_date"
	| "end_date_exclusive"
	| "all_day"
	| "timezone"
	| "rrule"
	| "exdates"
	| "reminder_minutes"
	| "attendees";
export type CalendarEventStatus = "confirmed" | "tentative" | "cancelled";

interface CalendarEventDetailBase {
	id: string;
	title: string;
	description: string;
	location: string;
	timezone: string;
	rrule: string | null;
	exdates: string[];
	sequence: number;
	status: CalendarEventStatus;
	version: number;
	organizer_id: string | null;
	organizer_email: string;
	attendees: CalendarEventReadAttendee[];
	viewer_attendee: CalendarEventReadAttendee | null;
	can_rsvp: boolean;
	mail_message_id: string | null;
}

export type CalendarEventDetail = CalendarEventDetailBase &
	(
		| {
				all_day: false;
				starts_at: string;
				ends_at: string;
				start_date: null;
				end_date_exclusive: null;
		  }
		| {
				all_day: true;
				starts_at: null;
				ends_at: null;
				start_date: string;
				end_date_exclusive: string;
		  }
	);

export interface CalendarEventMutationResult {
	event_id: string;
	version: number;
	sequence: number;
	status: CalendarEventStatus;
	changed_fields: CalendarEventChangedField[];
}

export interface CalendarOccurrenceCancellation {
	event_id: string;
	recurrence_id: string;
	status: "cancelled";
	sequence: number;
	event_version: number;
}

export interface CalendarEventRSVPRequest {
	status: CalendarEventRSVPChoice;
	attendee_version: number;
}

export type CalendarIdempotencyRequestOptions = Omit<
	ApiRequestOptions,
	"idempotencyKey" | "ifMatch"
> & { idempotencyKey: string };

export type CalendarVersionRequestOptions = Omit<
	ApiRequestOptions,
	"idempotencyKey" | "ifMatch"
> & { ifMatch: number };

export type CalendarEventUpdateOptions = CalendarVersionRequestOptions & {
	notifyAttendees?: boolean;
};

function calendarEventPath(eventId: string): string {
	return `/calendar/v2/events/${encodeURIComponent(eventId)}`;
}

const CALENDAR_EVENT_ID_PATTERN =
	/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const CALENDAR_NIL_EVENT_ID = "00000000-0000-0000-0000-000000000000";

export function isCanonicalCalendarEventID(eventId: string): boolean {
	return (
		CALENDAR_EVENT_ID_PATTERN.test(eventId) &&
		eventId !== CALENDAR_NIL_EVENT_ID
	);
}

export function getCalendarEvent(eventId: string): Promise<CalendarEventDetail> {
	if (!isCanonicalCalendarEventID(eventId)) {
		return Promise.reject(new Error("Invalid calendar event ID"));
	}
	return api.get<CalendarEventDetail>(calendarEventPath(eventId));
}

export function createCalendarEvent(
	request: CalendarEventCreateRequest,
	options: CalendarIdempotencyRequestOptions,
): Promise<CalendarEventCreateResponse> {
	return api.post<CalendarEventCreateResponse>(
		"/calendar/v2/events",
		request,
		options,
	);
}

export function createCalendarMeeting(
	request: CalendarMeetingCreateRequest,
	options: CalendarIdempotencyRequestOptions,
): Promise<CalendarMeetingCreateResponse> {
	return api.post<CalendarMeetingCreateResponse>(
		"/calendar/v2/meetings",
		request,
		options,
	);
}

export function updateCalendarEvent(
	eventId: string,
	request: CalendarEventUpdateRequest,
	options: CalendarEventUpdateOptions,
): Promise<CalendarEventMutationResult> {
	const notifyAttendees =
		options.notifyAttendees === undefined
			? ""
			: `?notify_attendees=${options.notifyAttendees ? "true" : "false"}`;
	return api.patch<CalendarEventMutationResult>(
		`${calendarEventPath(eventId)}${notifyAttendees}`,
		request,
		options,
	);
}

export function cancelCalendarEvent(
	eventId: string,
	options: CalendarVersionRequestOptions,
): Promise<CalendarEventMutationResult> {
	return api.delete<CalendarEventMutationResult>(calendarEventPath(eventId), options);
}

export function respondToCalendarSeries(
	eventId: string,
	request: CalendarEventRSVPRequest,
	options: CalendarIdempotencyRequestOptions,
): Promise<CalendarEventAttendeeResponse> {
	return api.post<CalendarEventAttendeeResponse>(
		`${calendarEventPath(eventId)}/rsvp`,
		request,
		options,
	);
}

export function respondToCalendarOccurrence(
	eventId: string,
	recurrenceId: string,
	request: CalendarEventRSVPRequest,
	options: CalendarIdempotencyRequestOptions,
): Promise<CalendarEventAttendeeResponse> {
	return api.post<CalendarEventAttendeeResponse>(
		`${calendarEventPath(eventId)}/instances/${encodeURIComponent(recurrenceId)}/rsvp`,
		request,
		options,
	);
}

export function cancelCalendarOccurrence(
	eventId: string,
	recurrenceId: string,
	options: CalendarVersionRequestOptions,
): Promise<CalendarOccurrenceCancellation> {
	return api.delete<CalendarOccurrenceCancellation>(
		`${calendarEventPath(eventId)}/instances/${encodeURIComponent(recurrenceId)}`,
		options,
	);
}
