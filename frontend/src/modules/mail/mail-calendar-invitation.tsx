import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { CalendarDays } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	getCalendarEvent,
	isCanonicalCalendarEventID,
	respondToCalendarOccurrence,
	respondToCalendarSeries,
	type CalendarEventRSVPChoice,
	type CalendarEventRSVPStatus,
} from "@/api/calendar";
import type { MailCalendarRSVPTarget } from "@/api/mail";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { mailKeys } from "./mail-query-keys";

const RSVP_ACTIONS: ReadonlyArray<{
	status: CalendarEventRSVPChoice;
	labelKey: string;
}> = [
	{ status: "accepted", labelKey: "reader.calendar.rsvp.actions.accepted" },
	{ status: "tentative", labelKey: "reader.calendar.rsvp.actions.tentative" },
	{ status: "declined", labelKey: "reader.calendar.rsvp.actions.declined" },
];

const CANONICAL_DATE_RECURRENCE_ID = /^\d{4}-\d{2}-\d{2}$/;
const CANONICAL_UTC_RECURRENCE_ID = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/;

function isCanonicalRecurrenceID(value: string): boolean {
	const expected = CANONICAL_DATE_RECURRENCE_ID.test(value)
		? `${value}T00:00:00.000Z`
		: CANONICAL_UTC_RECURRENCE_ID.test(value)
			? `${value.slice(0, -1)}.000Z`
			: null;
	if (expected === null) return false;

	const parsed = new Date(value);
	return Number.isFinite(parsed.getTime()) && parsed.toISOString() === expected;
}

function isCalendarRSVPStatus(
	value: unknown,
): value is CalendarEventRSVPStatus {
	return (
		value === "needs_action" ||
		value === "accepted" ||
		value === "tentative" ||
		value === "declined"
	);
}

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

function rsvpTargetForEvent(
	eventId: string,
	target: unknown,
): MailCalendarRSVPTarget | null {
	if (
		target === null ||
		typeof target !== "object" ||
		Array.isArray(target) ||
		!isCanonicalCalendarEventID(eventId)
	) {
		return null;
	}

	const candidate = target as {
		event_id?: unknown;
		scope?: unknown;
		recurrence_id?: unknown;
		attendee_version?: unknown;
		rsvp_status?: unknown;
	};
	if (
		typeof candidate.event_id !== "string" ||
		!isCanonicalCalendarEventID(candidate.event_id) ||
		candidate.event_id !== eventId ||
		typeof candidate.attendee_version !== "number" ||
		!Number.isSafeInteger(candidate.attendee_version) ||
		candidate.attendee_version <= 0 ||
		!isCalendarRSVPStatus(candidate.rsvp_status)
	) {
		return null;
	}
	if (candidate.scope === "series") {
		return candidate.recurrence_id === null
			? {
					event_id: candidate.event_id,
					scope: "series",
					recurrence_id: null,
					attendee_version: candidate.attendee_version,
					rsvp_status: candidate.rsvp_status,
				}
			: null;
	}
	if (
		candidate.scope !== "occurrence" ||
		typeof candidate.recurrence_id !== "string" ||
		!isCanonicalRecurrenceID(candidate.recurrence_id)
	) {
		return null;
	}
	return {
		event_id: candidate.event_id,
		scope: "occurrence",
		recurrence_id: candidate.recurrence_id,
		attendee_version: candidate.attendee_version,
		rsvp_status: candidate.rsvp_status,
	};
}

function rsvpTargetKey(target: MailCalendarRSVPTarget | null): string | null {
	if (target === null) return null;
	return `${target.event_id.toLowerCase()}:${target.scope}:${target.recurrence_id ?? ""}`;
}

function rsvpTargetStateKey(
	target: MailCalendarRSVPTarget | null,
): string | null {
	const targetKey = rsvpTargetKey(target);
	if (target === null || targetKey === null) return null;
	return `${targetKey}:${target.attendee_version}:${target.rsvp_status}`;
}

interface MailCalendarInvitationProps {
	eventId: string;
	rsvpTarget: MailCalendarRSVPTarget | null | undefined;
}

export function MailCalendarInvitation({
	eventId,
	rsvpTarget,
}: MailCalendarInvitationProps) {
	const { t } = useTranslation("mail");
	const queryClient = useQueryClient();
	const queryKey = calendarEventDetailKey(eventId);
	const target = rsvpTargetForEvent(eventId, rsvpTarget);
	const targetStateKey = rsvpTargetStateKey(target);
	const [localEffectiveRSVP, setLocalEffectiveRSVP] = useState<{
		targetStateKey: string;
		attendeeVersion: number;
		rsvpStatus: CalendarEventRSVPStatus;
	} | null>(null);
	useEffect(() => {
		setLocalEffectiveRSVP(null);
	}, [targetStateKey]);
	const effectiveRSVP =
		localEffectiveRSVP?.targetStateKey === targetStateKey
			? localEffectiveRSVP
			: target === null
				? null
				: {
						targetStateKey,
						attendeeVersion: target.attendee_version,
						rsvpStatus: target.rsvp_status,
					};
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
			target,
		}: {
			status: CalendarEventRSVPChoice;
			attendeeVersion: number;
			target: MailCalendarRSVPTarget;
			targetStateKey: string;
		}) => {
			const request = { status, attendee_version: attendeeVersion };
			const options = { idempotencyKey: globalThis.crypto.randomUUID() };
			if (target.scope === "series") {
				return respondToCalendarSeries(eventId, request, options);
			}
			return respondToCalendarOccurrence(
				eventId,
				target.recurrence_id,
				request,
				options,
			);
		},
		onSuccess: (attendee, submission) => {
			setLocalEffectiveRSVP({
				targetStateKey: submission.targetStateKey,
				attendeeVersion: attendee.version,
				rsvpStatus: attendee.rsvp_status,
			});
			invalidateCalendar();
			void queryClient.invalidateQueries({
				queryKey: mailKeys.messagesRoot(),
			});
		},
		onError: (error) => {
			if (apiStatus(error) === 409) {
				invalidateCalendar();
				void queryClient.invalidateQueries({
					queryKey: mailKeys.messagesRoot(),
				});
			}
		},
	});

	const event = eventQuery.data;
	const canRespond =
		target !== null &&
		event !== undefined &&
		event.can_rsvp &&
		event.status !== "cancelled";
	const rsvpBusy = rsvpMutation.isPending || eventQuery.isFetching;
	const rsvpError =
		rsvpMutation.isError && apiStatus(rsvpMutation.error) !== 409
			? t("reader.calendar.rsvp.failed")
			: null;

	const submitRSVP = (status: CalendarEventRSVPChoice) => {
		if (
			!canRespond ||
			!effectiveRSVP ||
			rsvpBusy ||
			target === null ||
			targetStateKey === null
		) {
			return;
		}
		rsvpMutation.mutate({
			status,
			attendeeVersion: effectiveRSVP.attendeeVersion,
			target,
			targetStateKey,
		});
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
						<Button
							asChild
							variant="outline"
							size="sm"
							className="min-h-11 shrink-0"
						>
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

					{effectiveRSVP ? (
						<div className="mt-3 flex flex-wrap items-center gap-2">
							<p className="text-xs text-t2">
								{t("reader.calendar.rsvp.current", {
									status: t(
										`reader.calendar.rsvp.status.${effectiveRSVP.rsvpStatus}`,
									),
								})}
							</p>
							<Badge variant={rsvpBadgeVariant(effectiveRSVP.rsvpStatus)}>
								{t(`reader.calendar.rsvp.status.${effectiveRSVP.rsvpStatus}`)}
							</Badge>
						</div>
					) : null}

					{canRespond && effectiveRSVP ? (
						<div className="mt-3 flex flex-wrap gap-2">
							{RSVP_ACTIONS.map((action) => (
								<Button
									key={action.status}
									type="button"
									variant={
										effectiveRSVP.rsvpStatus === action.status
											? "default"
											: "outline"
									}
									aria-pressed={effectiveRSVP.rsvpStatus === action.status}
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
