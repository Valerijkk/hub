import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { CalendarDays } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	getCalendarEvent,
	respondToCalendarSeries,
	type CalendarEventRSVPChoice,
	type CalendarEventRSVPStatus,
} from "@/api/calendar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

const RSVP_ACTIONS: ReadonlyArray<{
	status: CalendarEventRSVPChoice;
	labelKey: string;
}> = [
	{ status: "accepted", labelKey: "reader.calendar.rsvp.actions.accepted" },
	{ status: "tentative", labelKey: "reader.calendar.rsvp.actions.tentative" },
	{ status: "declined", labelKey: "reader.calendar.rsvp.actions.declined" },
];

function calendarEventDetailKey(eventId: string) {
	return ["calendar", "event", eventId] as const;
}

function apiStatus(error: unknown): number | null {
	if (!error || typeof error !== "object") return null;
	const code = (error as { code?: unknown }).code;
	return typeof code === "number" ? code : null;
}

function rsvpBadgeVariant(status: CalendarEventRSVPStatus) {
	if (status === "accepted") return "success" as const;
	if (status === "declined") return "destructive" as const;
	if (status === "tentative") return "warning" as const;
	return "secondary" as const;
}

interface MailCalendarInvitationProps {
	eventId: string;
}

export function MailCalendarInvitation({
	eventId,
}: MailCalendarInvitationProps) {
	const { t } = useTranslation("mail");
	const queryClient = useQueryClient();
	const queryKey = calendarEventDetailKey(eventId);
	const eventQuery = useQuery({
		queryKey,
		queryFn: () => getCalendarEvent(eventId),
		retry: false,
		refetchOnWindowFocus: false,
	});
	const invalidateCalendar = () => {
		void queryClient.invalidateQueries({ queryKey, exact: true });
		void queryClient.invalidateQueries({ queryKey: ["calendar"] });
	};
	const rsvpMutation = useMutation({
		mutationFn: ({
			status,
			attendeeVersion,
		}: {
			status: CalendarEventRSVPChoice;
			attendeeVersion: number;
		}) =>
			respondToCalendarSeries(
				eventId,
				{ status, attendee_version: attendeeVersion },
				{ idempotencyKey: globalThis.crypto.randomUUID() },
			),
		onSuccess: invalidateCalendar,
		onError: (error) => {
			if (apiStatus(error) === 409) {
				invalidateCalendar();
			}
		},
	});

	const event = eventQuery.data;
	const viewer = event?.viewer_attendee;
	const canRespond =
		event !== undefined &&
		event.can_rsvp &&
		viewer !== null &&
		viewer !== undefined &&
		event.status !== "cancelled";
	const rsvpBusy = rsvpMutation.isPending || eventQuery.isFetching;
	const rsvpError =
		rsvpMutation.isError && apiStatus(rsvpMutation.error) !== 409
			? t("reader.calendar.rsvp.failed")
			: null;

	const submitRSVP = (status: CalendarEventRSVPChoice) => {
		if (!canRespond || !viewer || rsvpBusy) return;
		rsvpMutation.mutate({ status, attendeeVersion: viewer.version });
	};

	return (
		<Card className="hover:translate-y-0">
			<CardContent className="p-4">
				<section aria-labelledby={`mail-calendar-invitation-${eventId}`}>
					<div className="flex flex-wrap items-center justify-between gap-3">
						<div className="flex min-w-0 items-center gap-2">
							<CalendarDays
								aria-hidden="true"
								className="h-4 w-4 shrink-0 text-t3"
							/>
							<h4
								id={`mail-calendar-invitation-${eventId}`}
								className="text-sm font-semibold text-t1"
							>
								{t("reader.calendar.request")}
							</h4>
						</div>
						<Button asChild variant="outline" size="sm" className="min-h-11 shrink-0">
							<Link
								to="/calendar/events/$eventId"
								params={{ eventId: eventId.toLowerCase() }}
							>
								{t("reader.calendar.openEvent")}
							</Link>
						</Button>
					</div>

					{eventQuery.isError ? (
						<p role="alert" className="mt-3 text-xs text-t2">
							{t("reader.calendar.rsvp.unavailable")}
						</p>
					) : null}

					{viewer ? (
						<div className="mt-3 flex flex-wrap items-center gap-2">
							<p className="text-xs text-t2">
								{t("reader.calendar.rsvp.current", {
									status: t(
										`reader.calendar.rsvp.status.${viewer.rsvp_status}`,
									),
								})}
							</p>
							<Badge variant={rsvpBadgeVariant(viewer.rsvp_status)}>
								{t(`reader.calendar.rsvp.status.${viewer.rsvp_status}`)}
							</Badge>
						</div>
					) : null}

					{canRespond && viewer ? (
						<div className="mt-3 flex flex-wrap gap-2">
							{RSVP_ACTIONS.map((action) => (
								<Button
									key={action.status}
									type="button"
									variant={
										viewer.rsvp_status === action.status ? "default" : "outline"
									}
									aria-pressed={viewer.rsvp_status === action.status}
									disabled={rsvpBusy}
									onClick={() => submitRSVP(action.status)}
									className="min-h-11"
								>
									{t(action.labelKey)}
								</Button>
							))}
						</div>
					) : null}

					{rsvpMutation.isPending ? (
						<p className="mt-3 text-xs text-t3">
							{t("reader.calendar.rsvp.saving")}
						</p>
					) : null}
					{rsvpError ? (
						<p role="alert" className="mt-3 text-xs text-[var(--red)]">
							{rsvpError}
						</p>
					) : null}
				</section>
			</CardContent>
		</Card>
	);
}
