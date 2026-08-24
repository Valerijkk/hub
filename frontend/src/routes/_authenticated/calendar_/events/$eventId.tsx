import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
	ArrowLeft,
	CalendarDays,
	Clock3,
	MapPin,
	Repeat2,
	UsersRound,
} from "lucide-react";
import {
	getCalendarEvent,
	isCanonicalCalendarEventID,
	respondToCalendarSeries,
	type CalendarEventDetail,
	type CalendarEventRSVPChoice,
	type CalendarEventRSVPStatus,
} from "@/api/calendar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

const RSVP_LABELS: Record<CalendarEventRSVPStatus, string> = {
	needs_action: "Ожидает ответа",
	accepted: "Принято",
	declined: "Отклонено",
	tentative: "Возможно",
};

const RSVP_ACTIONS: Array<{
	status: CalendarEventRSVPChoice;
	label: string;
}> = [
	{ status: "accepted", label: "Принять" },
	{ status: "tentative", label: "Возможно" },
	{ status: "declined", label: "Отклонить" },
];

function calendarEventDetailKey(eventId: string) {
	return ["calendar", "event", eventId] as const;
}

function apiStatus(error: unknown): number | null {
	if (!error || typeof error !== "object") return null;
	const code = (error as { code?: unknown }).code;
	return typeof code === "number" ? code : null;
}

function formatDateOnly(value: string): string {
	const parsed = new Date(`${value}T00:00:00Z`);
	if (Number.isNaN(parsed.getTime())) return "Дата недоступна";
	return new Intl.DateTimeFormat("ru-RU", {
		dateStyle: "long",
		timeZone: "UTC",
	}).format(parsed);
}

function previousDate(value: string): string {
	const parsed = new Date(`${value}T00:00:00Z`);
	parsed.setUTCDate(parsed.getUTCDate() - 1);
	return parsed.toISOString().slice(0, 10);
}

function formatDateTime(value: string, timeZone: string): string {
	const parsed = new Date(value);
	if (Number.isNaN(parsed.getTime())) return "Дата недоступна";
	const options: Intl.DateTimeFormatOptions = {
		dateStyle: "long",
		timeStyle: "short",
		timeZone,
	};
	try {
		return new Intl.DateTimeFormat("ru-RU", options).format(parsed);
	} catch {
		return new Intl.DateTimeFormat("ru-RU", {
			dateStyle: "long",
			timeStyle: "short",
		}).format(parsed);
	}
}

function eventTimeLabel(event: CalendarEventDetail): string {
	if (event.all_day) {
		const endDate = previousDate(event.end_date_exclusive);
		return endDate === event.start_date
			? formatDateOnly(event.start_date)
			: `${formatDateOnly(event.start_date)} — ${formatDateOnly(endDate)}`;
	}
	return `${formatDateTime(event.starts_at, event.timezone)} — ${formatDateTime(
		event.ends_at,
		event.timezone,
	)}`;
}

function eventStatusLabel(event: CalendarEventDetail): string {
	if (event.status === "cancelled") return "Отменено";
	if (event.status === "tentative") return "Предварительно";
	return "Подтверждено";
}

function attendeeRoleLabel(role: string): string {
	if (role === "organizer") return "Организатор";
	if (role === "optional") return "Необязательно";
	return "Участник";
}

function rsvpBadgeVariant(status: CalendarEventRSVPStatus) {
	if (status === "accepted") return "success" as const;
	if (status === "declined") return "destructive" as const;
	if (status === "tentative") return "warning" as const;
	return "secondary" as const;
}

function eventStatusBadgeVariant(status: CalendarEventDetail["status"]) {
	if (status === "cancelled") return "destructive" as const;
	if (status === "tentative") return "warning" as const;
	return "success" as const;
}

function CalendarEventLoading() {
	return (
		<div className="mx-auto max-w-3xl space-y-4" aria-busy="true">
			<div className="h-10 w-32 rounded-[var(--r2)] animate-shimmer" />
			<div className="h-56 rounded-[var(--r4)] animate-shimmer" />
			<div className="h-48 rounded-[var(--r4)] animate-shimmer" />
		</div>
	);
}

function CalendarEventUnavailable({
	notFound,
	onRetry,
}: {
	notFound: boolean;
	onRetry?: () => void;
}) {
	return (
		<div className="mx-auto flex max-w-xl flex-col items-center gap-4 py-16 text-center">
			<CalendarDays aria-hidden="true" className="h-10 w-10 text-t3" />
			<div>
				<h1 className="text-xl font-semibold text-t1">
					{notFound ? "Событие не найдено" : "Не удалось загрузить событие"}
				</h1>
				<p className="mt-2 text-sm text-t2">
					{notFound
						? "Оно удалено или у вас нет доступа."
						: "Проверьте подключение и попробуйте ещё раз."}
				</p>
			</div>
			<div className="flex flex-wrap justify-center gap-2">
				<Button asChild variant="outline">
					<Link to="/calendar">Вернуться в календарь</Link>
				</Button>
				{onRetry ? <Button onClick={onRetry}>Повторить</Button> : null}
			</div>
		</div>
	);
}

function CalendarEventDetailPage() {
	const { eventId } = Route.useParams();
	const queryClient = useQueryClient();
	const validEventId = isCanonicalCalendarEventID(eventId);
	const queryKey = calendarEventDetailKey(eventId);
	const eventQuery = useQuery({
		queryKey,
		queryFn: () => getCalendarEvent(eventId),
		enabled: validEventId,
	});
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
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey });
		},
		onError: (error) => {
			if (apiStatus(error) === 409) {
				void queryClient.invalidateQueries({ queryKey });
			}
		},
	});

	if (!validEventId) {
		return <CalendarEventUnavailable notFound />;
	}
	if (eventQuery.isLoading) {
		return <CalendarEventLoading />;
	}
	if (eventQuery.isError) {
		return (
			<CalendarEventUnavailable
				notFound={apiStatus(eventQuery.error) === 404}
				onRetry={
					apiStatus(eventQuery.error) === 404
						? undefined
						: () => void eventQuery.refetch()
				}
			/>
		);
	}
	if (!eventQuery.data) {
		return <CalendarEventUnavailable notFound />;
	}

	const event = eventQuery.data;
	const viewer = event.viewer_attendee;
	const rsvpBusy = rsvpMutation.isPending || eventQuery.isFetching;
	const rsvpError = rsvpMutation.isError
		? apiStatus(rsvpMutation.error) === 409
			? "Событие изменилось. Данные обновляются — повторите ответ."
			: "Не удалось сохранить ответ. Попробуйте ещё раз."
		: null;

	const submitRSVP = (status: CalendarEventRSVPChoice) => {
		if (!event.can_rsvp || !viewer || rsvpBusy) return;
		rsvpMutation.mutate({ status, attendeeVersion: viewer.version });
	};

	return (
		<div className="mx-auto max-w-3xl animate-[eu_450ms_var(--spring)_both]">
			<Button asChild variant="ghost" className="mb-4 min-h-11">
				<Link to="/calendar">
					<ArrowLeft aria-hidden="true" className="h-4 w-4" />
					К календарю
				</Link>
			</Button>

			<Card className="hover:translate-y-0">
				<CardHeader className="gap-3">
					<div className="flex flex-wrap items-center gap-2">
						<Badge variant={eventStatusBadgeVariant(event.status)}>
							{eventStatusLabel(event)}
						</Badge>
						{event.all_day ? <Badge variant="secondary">Весь день</Badge> : null}
					</div>
					<CardTitle className="text-2xl leading-tight text-t1">
						{event.title}
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-6">
					<div className="space-y-4 border-b border-b1 pb-6">
						<div className="flex items-start gap-3">
							<Clock3 aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-t3" />
							<div>
								<p className="text-sm font-medium text-t1">{eventTimeLabel(event)}</p>
								<p className="mt-1 text-xs text-t3">{event.timezone}</p>
							</div>
						</div>
						{event.location ? (
							<div className="flex items-start gap-3">
								<MapPin aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-t3" />
								<p className="break-words text-sm text-t1">{event.location}</p>
							</div>
						) : null}
						{event.rrule ? (
							<div className="flex items-start gap-3">
								<Repeat2 aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-t3" />
								<div className="min-w-0">
									<p className="text-sm text-t1">Повторяющееся событие</p>
									<p className="mt-1 break-all text-xs text-t3">{event.rrule}</p>
									{event.exdates.length > 0 ? (
										<p className="mt-1 text-xs text-t3">
											Исключений: {event.exdates.length}
										</p>
									) : null}
								</div>
							</div>
						) : null}
					</div>

					{event.description ? (
						<section aria-labelledby="calendar-event-description">
							<h2 id="calendar-event-description" className="text-sm font-semibold text-t1">
								Описание
							</h2>
							<p className="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-t2">
								{event.description}
							</p>
						</section>
					) : null}

					<section aria-labelledby="calendar-event-attendees">
						<div className="flex items-center gap-2">
							<UsersRound aria-hidden="true" className="h-5 w-5 text-t3" />
							<h2 id="calendar-event-attendees" className="text-sm font-semibold text-t1">
								Участники ({event.attendees.length})
							</h2>
						</div>
						<p className="mt-2 text-xs text-t3">
							Организатор: {event.organizer_email}
						</p>
						<ul className="mt-3 divide-y divide-b1 rounded-[var(--r2)] border border-b1 bg-base/50">
							{event.attendees.map((attendee) => (
								<li key={attendee.id} className="flex flex-wrap items-center gap-3 px-3 py-3">
									<div className="min-w-0 flex-1">
										<p className="truncate text-sm font-medium text-t1">{attendee.email}</p>
										<p className="mt-0.5 text-xs text-t3">{attendeeRoleLabel(attendee.role)}</p>
									</div>
									<Badge variant={rsvpBadgeVariant(attendee.rsvp_status)}>
										{RSVP_LABELS[attendee.rsvp_status]}
									</Badge>
								</li>
							))}
						</ul>
					</section>

					{event.can_rsvp && viewer ? (
						<section aria-labelledby="calendar-event-rsvp" className="rounded-[var(--r3)] bg-raised p-4">
							<h2 id="calendar-event-rsvp" className="text-sm font-semibold text-t1">
								Ваш ответ: {RSVP_LABELS[viewer.rsvp_status]}
							</h2>
							<div className="mt-3 flex flex-wrap gap-2">
								{RSVP_ACTIONS.map((action) => (
									<Button
										key={action.status}
										type="button"
										variant={viewer.rsvp_status === action.status ? "default" : "outline"}
										aria-pressed={viewer.rsvp_status === action.status}
										disabled={rsvpBusy}
										onClick={() => submitRSVP(action.status)}
									>
										{action.label}
									</Button>
								))}
							</div>
							{rsvpMutation.isPending ? (
								<p className="mt-3 text-xs text-t3">Сохраняем ответ…</p>
							) : null}
							{rsvpError ? (
								<p role="alert" className="mt-3 text-xs text-[var(--red)]">
									{rsvpError}
								</p>
							) : null}
						</section>
					) : null}
				</CardContent>
			</Card>
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/calendar_/events/$eventId",
)({
	component: CalendarEventDetailPage,
});
