import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Check, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	type CalendarAvailabilityResponse,
	type CalendarMeetingCreateRequest,
	createCalendarMeeting,
	getCalendarAvailability,
} from "@/api/calendar";
import { getChannels } from "@/api/messenger";
import {
	listMembers,
	listProjects,
	type MemberResponse,
	type ProjectResponse,
} from "@/api/projects";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { useAuthStore } from "@/stores/auth-store";

// ---------------------------------------------------------------------------
// Пресеты длительности
// ---------------------------------------------------------------------------

const DURATION_OPTIONS = [
	{ label: "30 мин", value: "30" },
	{ label: "1 час", value: "60" },
	{ label: "1.5 часа", value: "90" },
	{ label: "2 часа", value: "120" },
] as const;

// ---------------------------------------------------------------------------
// Пресеты повторения
// ---------------------------------------------------------------------------

type RecurrencePreset = "" | "daily" | "weekly" | "workdays";

const RECURRENCE_OPTIONS: ReadonlyArray<{
	label: string;
	value: RecurrencePreset;
}> = [
	{ label: "Нет", value: "" },
	{ label: "Каждый день · 20", value: "daily" },
	{ label: "Каждую неделю · 12", value: "weekly" },
	{ label: "По будням · 12", value: "workdays" },
];

const MAX_AVAILABILITY_USERS = 50;
const MAX_INVITEE_COUNT = MAX_AVAILABILITY_USERS - 1;
const MAX_DESCRIPTION_BYTES = 2000;
const MAX_TIMEZONE_BYTES = 64;
const MAX_TIMEZONE_OFFSET_MS = 14 * 60 * 60 * 1000;
const WEEKDAY_CODES = ["SU", "MO", "TU", "WE", "TH", "FR", "SA"] as const;

interface WallDate {
	year: number;
	month: number;
	day: number;
	weekday: number;
}

interface WallTime extends WallDate {
	hours: number;
	minutes: number;
}

interface MeetingInterval {
	start: Date;
	end: Date;
	startsAt: string;
	endsAt: string;
	durationMinutes: number;
	weekday: number;
}

type AvailabilityAssessment = "available" | "busy" | "incomplete";

type IntlWithSupportedValuesOf = typeof Intl & {
	supportedValuesOf?: (key: "timeZone") => string[];
};

const timeZoneFormatters = new Map<string, Intl.DateTimeFormat>();

function utf8ByteLength(value: string): number {
	return new TextEncoder().encode(value).length;
}

function isSafeIanaTimeZone(value: string): boolean {
	if (
		!value ||
		value.trim() !== value ||
		value.toLowerCase() === "local" ||
		utf8ByteLength(value) > MAX_TIMEZONE_BYTES
	) {
		return false;
	}
	try {
		new Intl.DateTimeFormat("en-US", { timeZone: value }).format(0);
		return true;
	} catch {
		return false;
	}
}

function browserTimeZone(): string {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone ?? "";
	} catch {
		return "";
	}
}

function safeTimeZoneOptions(): string[] {
	const fallback = ["UTC", browserTimeZone()].filter(isSafeIanaTimeZone);
	const supportedValuesOf = (Intl as IntlWithSupportedValuesOf).supportedValuesOf;
	if (typeof supportedValuesOf !== "function") return fallback;
	try {
		return Array.from(
			new Set([...fallback, ...supportedValuesOf("timeZone")].filter(isSafeIanaTimeZone)),
		).sort((left, right) =>
			left === "UTC" ? -1 : right === "UTC" ? 1 : left.localeCompare(right),
		);
	} catch {
		return fallback;
	}
}

function defaultTimeZone(options: string[]): string {
	const browser = browserTimeZone();
	return options.includes(browser) ? browser : "UTC";
}

function parseWallDate(value: string): WallDate | null {
	const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
	if (!match) return null;
	const year = Number(match[1]);
	const month = Number(match[2]);
	const day = Number(match[3]);
	const instant = Date.UTC(year, month - 1, day);
	const parsed = new Date(instant);
	if (
		!Number.isFinite(instant) ||
		parsed.getUTCFullYear() !== year ||
		parsed.getUTCMonth() !== month - 1 ||
		parsed.getUTCDate() !== day
	) {
		return null;
	}
	return { year, month, day, weekday: parsed.getUTCDay() };
}

function timeZonePartsAt(instant: number, timeZone: string): WallTime | null {
	let formatter = timeZoneFormatters.get(timeZone);
	if (!formatter) {
		try {
			formatter = new Intl.DateTimeFormat("en-US", {
				calendar: "gregory",
				numberingSystem: "latn",
				timeZone,
				year: "numeric",
				month: "2-digit",
				day: "2-digit",
				hour: "2-digit",
				minute: "2-digit",
				hourCycle: "h23",
			});
			timeZoneFormatters.set(timeZone, formatter);
		} catch {
			return null;
		}
	}
	const values: Partial<Record<"year" | "month" | "day" | "hour" | "minute", number>> = {};
	for (const part of formatter.formatToParts(new Date(instant))) {
		if (part.type === "year" || part.type === "month" || part.type === "day" || part.type === "hour" || part.type === "minute") {
			values[part.type] = Number(part.value);
		}
	}
	const { year, month, day, hour, minute } = values;
	if (
		year === undefined ||
		month === undefined ||
		day === undefined ||
		hour === undefined ||
		minute === undefined
	) {
		return null;
	}
	const calendar = new Date(Date.UTC(year, month - 1, day));
	return {
		year,
		month,
		day,
		hours: hour,
		minutes: minute,
		weekday: calendar.getUTCDay(),
	};
}

function resolveWallTime(timeZone: string, wallTime: WallTime): Date | null {
	const target = Date.UTC(
		wallTime.year,
		wallTime.month - 1,
		wallTime.day,
		wallTime.hours,
		wallTime.minutes,
	);
	if (!Number.isFinite(target)) return null;
	let resolved: Date | null = null;
	for (
		let instant = target - MAX_TIMEZONE_OFFSET_MS;
		instant <= target + MAX_TIMEZONE_OFFSET_MS;
		instant += 60_000
	) {
		const parts = timeZonePartsAt(instant, timeZone);
		if (
			!parts ||
			parts.year !== wallTime.year ||
			parts.month !== wallTime.month ||
			parts.day !== wallTime.day ||
			parts.hours !== wallTime.hours ||
			parts.minutes !== wallTime.minutes
		) {
			continue;
		}
		if (resolved) return null;
		resolved = new Date(instant);
	}
	return resolved;
}

function parseMeetingInterval(
	dateValue: string,
	timeValue: string,
	durationValue: string,
	timeZone: string,
): MeetingInterval | null {
	const wallDate = parseWallDate(dateValue);
	const timeMatch = /^(\d{2}):(\d{2})$/.exec(timeValue);
	const durationMinutes = Number.parseInt(durationValue, 10);
	if (
		!wallDate ||
		!timeMatch ||
		!isSafeIanaTimeZone(timeZone) ||
		!DURATION_OPTIONS.some((option) => Number(option.value) === durationMinutes)
	) {
		return null;
	}
	const hours = Number(timeMatch[1]);
	const minutes = Number(timeMatch[2]);
	if (hours > 23 || minutes > 59) return null;
	const start = resolveWallTime(timeZone, { ...wallDate, hours, minutes });
	if (!start) return null;
	const end = new Date(start.getTime() + durationMinutes * 60_000);
	return {
		start,
		end,
		startsAt: start.toISOString(),
		endsAt: end.toISOString(),
		durationMinutes,
		weekday: wallDate.weekday,
	};
}

function formatWallDate(value: string): string | null {
	const wallDate = parseWallDate(value);
	if (!wallDate) return null;
	return new Intl.DateTimeFormat("ru-RU", {
		weekday: "short",
		day: "numeric",
		month: "long",
		timeZone: "UTC",
	}).format(new Date(Date.UTC(wallDate.year, wallDate.month - 1, wallDate.day)));
}

function recurrenceRule(
	preset: RecurrencePreset,
	interval: MeetingInterval,
): string | null {
	switch (preset) {
		case "":
			return "";
		case "daily":
			return "FREQ=DAILY;COUNT=20";
		case "weekly": {
			const weekday = WEEKDAY_CODES[interval.weekday];
			return weekday
				? `FREQ=WEEKLY;BYDAY=${weekday};COUNT=12`
				: null;
		}
		case "workdays": {
			const weekday = interval.weekday;
			return weekday === 0 || weekday === 6
				? null
				: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR;COUNT=12";
		}
	}
}

function availabilityInterval(
	value: { from: string; to: string },
	requestStart: number,
	requestEnd: number,
): [number, number] | null {
	const from = Date.parse(value.from);
	const to = Date.parse(value.to);
	if (
		!Number.isFinite(from) ||
		!Number.isFinite(to) ||
		from >= to ||
		from < requestStart ||
		to > requestEnd
	) {
		return null;
	}
	return [from, to];
}

function assessAvailability(
	response: CalendarAvailabilityResponse,
	requestedUserIds: string[],
	interval: MeetingInterval,
): AvailabilityAssessment {
	if (
		!Array.isArray(response.users) ||
		response.users.length !== requestedUserIds.length
	) {
		return "incomplete";
	}

	const requested = new Set(requestedUserIds);
	const returned = new Set<string>();
	const requestStart = interval.start.getTime();
	const requestEnd = interval.end.getTime();
	let hasBusyOverlap = false;

	for (const user of response.users) {
		if (
			!user ||
			typeof user.user_id !== "string" ||
			!requested.has(user.user_id) ||
			returned.has(user.user_id) ||
			!Array.isArray(user.busy) ||
			!Array.isArray(user.free)
		) {
			return "incomplete";
		}
		returned.add(user.user_id);

		const coverage: Array<[number, number]> = [];
		for (const busy of user.busy) {
			if (!busy) {
				return "incomplete";
			}
			const parsed = availabilityInterval(busy, requestStart, requestEnd);
			if (!parsed) {
				return "incomplete";
			}
			coverage.push(parsed);
			if (parsed[0] < requestEnd && parsed[1] > requestStart) {
				hasBusyOverlap = true;
			}
		}
		for (const free of user.free) {
			if (!free) {
				return "incomplete";
			}
			const parsed = availabilityInterval(free, requestStart, requestEnd);
			if (!parsed) {
				return "incomplete";
			}
			coverage.push(parsed);
		}

		coverage.sort((left, right) => left[0] - right[0] || left[1] - right[1]);
		let coveredUntil = requestStart;
		for (const [from, to] of coverage) {
			if (from > coveredUntil) {
				return "incomplete";
			}
			coveredUntil = Math.max(coveredUntil, to);
		}
		if (coveredUntil < requestEnd) {
			return "incomplete";
		}
	}

	return hasBusyOverlap ? "busy" : "available";
}

function newMeetingIdempotencyKey(): string {
	if (typeof globalThis.crypto?.randomUUID === "function") {
		return `meeting-${globalThis.crypto.randomUUID()}`;
	}
	const bytes = new Uint8Array(16);
	globalThis.crypto.getRandomValues(bytes);
	return `meeting-${Array.from(bytes, (value) =>
		value.toString(16).padStart(2, "0"),
	).join("")}`;
}

function meetingErrorCode(error: unknown): string {
	if (!error || typeof error !== "object") return "";
	const value = (error as Record<string, unknown>).error_code;
	return typeof value === "string" ? value : "";
}

// ---------------------------------------------------------------------------
// Пропсы компонента
// ---------------------------------------------------------------------------

interface CreateCallDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Предзаполнение даты (YYYY-MM-DD). */
	defaultDate?: string;
	/** Предзаполнение времени (HH:MM). */
	defaultTime?: string;
	/** Участники, предзаполненные вызывающим сценарием. */
	initialParticipantIds?: readonly string[];
	/** Заголовок, предзаполненный вызывающим сценарием. */
	initialTitle?: string;
}

// ---------------------------------------------------------------------------
// Компонент
// ---------------------------------------------------------------------------

/**
 * CreateCallDialog — создание календарной встречи со связанным звонком.
 * Перед отправкой приватно подтверждает доступность всех участников.
 */
export function CreateCallDialog({
	open,
	onOpenChange,
	defaultDate,
	defaultTime,
	initialParticipantIds,
	initialTitle,
}: CreateCallDialogProps) {
	const { t } = useTranslation("calls");
	const queryClient = useQueryClient();
	const currentUserId = useAuthStore((state) => state.user?.id ?? "");
	const timeZoneOptions = useMemo(() => safeTimeZoneOptions(), []);
	const defaultTimezone = useMemo(
		() => defaultTimeZone(timeZoneOptions),
		[timeZoneOptions],
	);
	const dialogSessionRef = useRef(0);
	const initialParticipantsSessionRef = useRef<number | null>(null);
	const [dialogSession, setDialogSession] = useState(0);
	const idempotencyRef = useRef<{ fingerprint: string; key: string } | null>(
		null,
	);
	const availabilityToastRef = useRef("");

	const [title, setTitle] = useState("");
	const [description, setDescription] = useState("");
	const [timezone, setTimezone] = useState(() => defaultTimezone);
	const [date, setDate] = useState("");
	const [time, setTime] = useState("10:00");
	const [duration, setDuration] = useState("60");
	const [recurrence, setRecurrence] = useState<RecurrencePreset>("");
	const [channelId, setChannelId] = useState("");
	const [projectId, setProjectId] = useState("");
	const [participantIds, setParticipantIds] = useState<string[]>([]);
	const [participantSearch, setParticipantSearch] = useState("");
	const normalizedInitialParticipantIds = useMemo(() => {
		const currentUserKey = currentUserId.trim().toLocaleLowerCase("und");
		const knownIDs = new Set<string>();
		for (const participantID of initialParticipantIds ?? []) {
			if (typeof participantID !== "string") continue;
			const normalized = participantID.trim().toLocaleLowerCase("und");
			if (!normalized || normalized === currentUserKey || knownIDs.has(normalized)) {
				continue;
			}
			knownIDs.add(normalized);
		}
		return Array.from(knownIDs);
	}, [currentUserId, initialParticipantIds]);
	const initialParticipantKey = normalizedInitialParticipantIds.join(",");

	const {
		data: channels,
		isError: channelsError,
		refetch: refetchChannels,
	} = useQuery({
		queryKey: ["messenger", "channels"],
		queryFn: getChannels,
	});

	const {
		data: projects,
		isError: projectsError,
		refetch: refetchProjects,
	} = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	const projectList = (projects ?? []) as ProjectResponse[];
	const channelList = useMemo(
		() => (channels ?? []).filter((channel) => channel.project_id === projectId),
		[channels, projectId],
	);
	const projectsEmpty = !projectsError && projectList.length === 0;
	const channelsEmpty =
		!!projectId && !channelsError && channelList.length === 0;

	// Если проект ещё не выбран — подставить первый из списка (можно сменить в UI).
	useEffect(() => {
		if (!projectId && projects && projects.length > 0) {
			setProjectId(projects[0]?.id ?? "");
		}
	}, [dialogSession, projectId, projects]);

	const {
		data: members,
		isSuccess: membersLoaded,
	} = useQuery({
		queryKey: ["project-members", projectId],
		queryFn: () => listMembers(projectId),
		enabled: !!projectId,
	});

	const meetingInterval = useMemo(
		() => parseMeetingInterval(date, time, duration, timezone),
		[date, time, duration, timezone],
	);
	const attendeeUserIds = useMemo(
		() =>
			Array.from(
				new Set(
					participantIds.filter(
						(userId) => userId.length > 0 && userId !== currentUserId,
					),
				),
			).sort(),
		[participantIds, currentUserId],
	);
	const availabilityUserIds = useMemo(
		() =>
			currentUserId
				? Array.from(new Set([currentUserId, ...attendeeUserIds])).sort()
				: [],
		[currentUserId, attendeeUserIds],
	);
	const availabilityQueryEnabled =
		open &&
		meetingInterval !== null &&
		attendeeUserIds.length > 0 &&
		attendeeUserIds.length <= MAX_INVITEE_COUNT &&
		availabilityUserIds.length === attendeeUserIds.length + 1;
	const availabilityRequest = useMemo(
		() =>
			availabilityQueryEnabled && meetingInterval
				? {
						user_ids: availabilityUserIds,
						from: meetingInterval.startsAt,
						to: meetingInterval.endsAt,
						timezone,
						slot_minutes: meetingInterval.durationMinutes,
					}
				: null,
		[
			availabilityQueryEnabled,
			availabilityUserIds,
			meetingInterval,
			timezone,
		],
	);
	const availabilityFingerprint = availabilityRequest
		? JSON.stringify(availabilityRequest)
		: "";
	const {
		data: availabilityData,
		isError: availabilityError,
		isFetching: availabilityFetching,
		isSuccess: availabilitySuccess,
		refetch: refetchAvailability,
	} = useQuery({
		queryKey: ["calendar-availability", dialogSession, availabilityRequest],
		queryFn: () => {
			if (!availabilityRequest) {
				throw new Error("calendar availability request is unavailable");
			}
			return getCalendarAvailability(availabilityRequest);
		},
		enabled: availabilityQueryEnabled,
		staleTime: 0,
		refetchOnWindowFocus: true,
	});

	const availabilityAssessment = useMemo(
		() =>
			availabilityData && meetingInterval
				? assessAvailability(
						availabilityData,
						availabilityUserIds,
						meetingInterval,
					)
				: "incomplete",
		[availabilityData, availabilityUserIds, meetingInterval],
	);

	useEffect(() => {
		if (
			!open ||
			dialogSession !== dialogSessionRef.current ||
			!availabilityQueryEnabled ||
			!availabilityFingerprint
		) {
			return;
		}
		const kind = availabilityError
			? "error"
			: availabilitySuccess && availabilityAssessment === "busy"
				? "busy"
				: availabilitySuccess && availabilityAssessment === "incomplete"
					? "incomplete"
					: "";
		if (!kind) {
			if (availabilitySuccess) availabilityToastRef.current = "";
			return;
		}
		const toastKey = `${dialogSession}:${kind}:${availabilityFingerprint}`;
		if (availabilityToastRef.current === toastKey) return;
		availabilityToastRef.current = toastKey;
		toast.error(
			kind === "busy"
				? "Выбранное время занято у одного или нескольких участников"
				: "Не удалось подтвердить доступность участников",
		);
	}, [
		availabilityAssessment,
		availabilityError,
		availabilityFingerprint,
		availabilityQueryEnabled,
		availabilitySuccess,
		dialogSession,
		open,
	]);

	const filteredMembers = useMemo(() => {
		const all = ((members ?? []) as MemberResponse[]).filter(
			(member) => member.user_id !== currentUserId,
		);
		if (!participantSearch.trim()) return all;
		const q = participantSearch.toLowerCase();
		return all.filter(
			(m) =>
				m.name.toLowerCase().includes(q) || m.email.toLowerCase().includes(q),
		);
	}, [currentUserId, members, participantSearch]);

	const toggleParticipant = (userId: string) => {
		if (!userId || userId === currentUserId) return;
		setParticipantIds((prev) => {
			if (prev.includes(userId)) {
				return prev.filter((id) => id !== userId);
			}
			if (prev.length >= MAX_INVITEE_COUNT) {
				toast.error("Для одной встречи можно выбрать не более 49 коллег");
				return prev;
			}
			return [...prev, userId];
		});
	};

	const removeParticipant = (userId: string) => {
		setParticipantIds((prev) => prev.filter((id) => id !== userId));
	};

	useEffect(() => {
		if (!currentUserId) return;
		setParticipantIds((previous) =>
			previous.includes(currentUserId)
				? previous.filter((userId) => userId !== currentUserId)
				: previous,
		);
	}, [currentUserId]);

	// Сбрасываем форму при открытии с переданными default-значениями
	useEffect(() => {
		const nextSession = dialogSessionRef.current + 1;
		dialogSessionRef.current = nextSession;
		setDialogSession(nextSession);
		initialParticipantsSessionRef.current = null;
		setParticipantIds([]);
		if (!open) return;

		setTitle(initialTitle ?? "");
		setDescription("");
		setDate(defaultDate ?? "");
		setTime(defaultTime ?? "10:00");
		setDuration("60");
		setRecurrence("");
		setTimezone(defaultTimezone);
		setChannelId("");
		setProjectId("");
		setParticipantSearch("");
		idempotencyRef.current = null;
		availabilityToastRef.current = "";
	}, [
		open,
		defaultDate,
		defaultTime,
		defaultTimezone,
		initialParticipantKey,
		initialTitle,
	]);

	useEffect(() => {
		if (
			!open ||
			!membersLoaded ||
			dialogSession !== dialogSessionRef.current ||
			initialParticipantsSessionRef.current === dialogSession
		) {
			return;
		}
		const availableMemberIDs = new Map(
			((members ?? []) as MemberResponse[]).flatMap((member) => {
				const normalized = member.user_id.trim().toLocaleLowerCase("und");
				return normalized ? [[normalized, member.user_id] as const] : [];
			}),
		);
		setParticipantIds(
			normalizedInitialParticipantIds
				.flatMap((userID) => {
					const availableUserID = availableMemberIDs.get(userID);
					return availableUserID ? [availableUserID] : [];
				})
				.slice(0, MAX_INVITEE_COUNT),
		);
		initialParticipantsSessionRef.current = dialogSession;
	}, [
		dialogSession,
		members,
		membersLoaded,
		normalizedInitialParticipantIds,
		open,
	]);

	const invalidateDialogSession = () => {
		const nextSession = dialogSessionRef.current + 1;
		dialogSessionRef.current = nextSession;
		setDialogSession(nextSession);
	};

	const handleDialogOpenChange = (nextOpen: boolean) => {
		if (!nextOpen) invalidateDialogSession();
		onOpenChange(nextOpen);
	};

	const mutation = useMutation({
		mutationFn: ({
			request,
			idempotencyKey,
		}: {
			request: CalendarMeetingCreateRequest;
			idempotencyKey: string;
			session: number;
		}) => createCalendarMeeting(request, { idempotencyKey }),
		onSuccess: (_data, variables) => {
			if (variables.session !== dialogSessionRef.current) return;
			idempotencyRef.current = null;
			void queryClient.invalidateQueries({ queryKey: ["calendar"] });
			void queryClient.invalidateQueries({
				queryKey: ["calendar-availability"],
			});
			void queryClient.invalidateQueries({ queryKey: ["calls", "scheduled"] });
			handleDialogOpenChange(false);
		},
		onError: (err, variables) => {
			if (variables.session !== dialogSessionRef.current) return;
			toast.error(
				meetingErrorCode(err) === "SLOT_UNAVAILABLE"
					? t("meetingSlotUnavailable")
					: "Не удалось создать встречу. Проверьте данные и попробуйте снова",
			);
		},
	});

	const recurrenceRRULE = useMemo(
		() => (meetingInterval ? recurrenceRule(recurrence, meetingInterval) : null),
		[meetingInterval, recurrence],
	);
	const availabilityReady =
		availabilityQueryEnabled &&
		availabilitySuccess &&
		!availabilityFetching &&
		availabilityAssessment === "available";
	const trimmedTitle = title.trim();
	const trimmedDescription = description.trim();
	const titleBytes = utf8ByteLength(trimmedTitle);
	const descriptionBytes = utf8ByteLength(trimmedDescription);
	const selectedProjectAvailable = projectList.some(
		(project) => project.id === projectId,
	);
	const selectedChannelAvailable = channelList.some(
		(channel) => channel.id === channelId,
	);
	const selectedAttendeesAvailable = attendeeUserIds.every((userId) =>
		((members ?? []) as MemberResponse[]).some(
			(member) => member.user_id === userId,
		),
	);
	const timeZoneValid =
		timeZoneOptions.includes(timezone) && isSafeIanaTimeZone(timezone);
	const isValid =
		trimmedTitle.length > 0 &&
		titleBytes <= 255 &&
		!!currentUserId &&
		!!meetingInterval &&
		selectedProjectAvailable &&
		selectedChannelAvailable &&
		selectedAttendeesAvailable &&
		timeZoneValid &&
		descriptionBytes <= MAX_DESCRIPTION_BYTES &&
		attendeeUserIds.length > 0 &&
		attendeeUserIds.length <= MAX_INVITEE_COUNT &&
		recurrenceRRULE !== null &&
		availabilityReady;

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!isValid || !meetingInterval || recurrenceRRULE === null) {
			if (!timeZoneValid) {
				toast.error(t("meetingTimezoneInvalid"));
			} else if (!meetingInterval) {
				toast.error(t("meetingWallTimeInvalid"));
			} else if (descriptionBytes > MAX_DESCRIPTION_BYTES) {
				toast.error(t("meetingDescriptionTooLong", { count: MAX_DESCRIPTION_BYTES }));
			} else if (!projectId) {
				toast.error("Выберите проект");
			} else if (!channelId) {
				toast.error("Выберите канал проекта");
			} else if (attendeeUserIds.length === 0) {
				toast.error("Выберите хотя бы одного коллегу");
			} else if (recurrenceRRULE === null) {
				toast.error("Для повтора по будням выберите будний день");
			} else if (
				availabilityError ||
				availabilityAssessment !== "available"
			) {
				toast.error("Не удалось подтвердить доступность участников");
			}
			return;
		}

		const request: CalendarMeetingCreateRequest = {
			title: trimmedTitle,
			description: trimmedDescription,
			project_id: projectId,
			channel_id: channelId,
			starts_at: meetingInterval.startsAt,
			ends_at: meetingInterval.endsAt,
			timezone,
			rrule: recurrenceRRULE,
			attendee_user_ids: attendeeUserIds,
		};
		const fingerprint = JSON.stringify(request);
		let idempotency = idempotencyRef.current;
		try {
			if (idempotency?.fingerprint !== fingerprint) {
				idempotency = {
					fingerprint,
					key: newMeetingIdempotencyKey(),
				};
				idempotencyRef.current = idempotency;
			}
		} catch {
			toast.error("Не удалось безопасно подготовить создание встречи");
			return;
		}
		if (!idempotency) {
			toast.error("Не удалось безопасно подготовить создание встречи");
			return;
		}
		mutation.mutate({
			request,
			idempotencyKey: idempotency.key,
			session: dialogSessionRef.current,
		});
	};

	// Форматируем выбранную дату для отображения в UI
	const dateDisplay = formatWallDate(date);

	return (
		<Dialog open={open} onOpenChange={handleDialogOpenChange}>
			<DialogContent className="max-h-[90vh] gap-0 overflow-y-auto p-0 sm:max-w-[420px]">
				<form onSubmit={handleSubmit}>
					<DialogHeader className="px-6 pt-6 pb-4">
						<DialogTitle className="text-lg">
							{t("createCallTitle", { defaultValue: "Новый звонок" })}
						</DialogTitle>
						{dateDisplay && (
							<p className="text-sm text-t2 capitalize mt-1">{dateDisplay}</p>
						)}
					</DialogHeader>

					<div className="px-6 space-y-5">
						{/* Название */}
						<div className="space-y-2">
							<Label htmlFor="call-title" className="text-sm font-medium">
								{t("callTitleLabel", { defaultValue: "Название" })}
							</Label>
							<Input
								id="call-title"
								placeholder={t("callTitlePlaceholder", {
									defaultValue: "Стендап, ретро, планирование...",
								})}
								value={title}
								onChange={(e) => setTitle(e.target.value)}
								maxLength={255}
								required
								autoFocus
								className="h-11"
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="call-description" className="text-sm font-medium">
								{t("meetingDescriptionLabel")}
							</Label>
							<textarea
								id="call-description"
								placeholder={t("meetingDescriptionPlaceholder")}
								value={description}
								onChange={(event) => setDescription(event.target.value)}
								maxLength={MAX_DESCRIPTION_BYTES}
								aria-describedby="call-description-limit"
								className="min-h-24 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm outline-none placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring"
							/>
							<p id="call-description-limit" className="text-xs text-t2">
								{t("meetingDescriptionLimit", { count: MAX_DESCRIPTION_BYTES })}
							</p>
						</div>

						{/* Проект — FE-MC-002: fail ≠ «нет проектов» */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">Проект</Label>
							{projectsError ? (
								<QueryErrorAlert
									compact
									message="Не удалось загрузить проекты"
									retry={() => void refetchProjects()}
								/>
							) : (
								<Select
									value={projectId}
									onValueChange={(v) => {
										setProjectId(v);
										setChannelId("");
										setParticipantIds([]);
									}}
								>
									<SelectTrigger
										className="h-11 rounded-xl"
										aria-label="Выбор проекта"
									>
										<SelectValue placeholder="Выберите проект" />
									</SelectTrigger>
									<SelectContent>
										{projectsEmpty ? (
											<SelectItem value="__none" disabled>
												Нет доступных проектов
											</SelectItem>
										) : (
											projectList.map((p) => (
												<SelectItem key={p.id} value={p.id}>
													{p.key ? `${p.key} — ${p.name}` : p.name}
												</SelectItem>
											))
										)}
									</SelectContent>
								</Select>
							)}
						</div>

						{/* Канал — FE-MC-002 / FE-MC-004 */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">Канал</Label>
							{channelsError ? (
								<QueryErrorAlert
									compact
									message="Не удалось загрузить каналы"
									retry={() => void refetchChannels()}
								/>
							) : (
								<Select value={channelId} onValueChange={setChannelId}>
									<SelectTrigger
										className="h-11 rounded-xl"
										aria-label="Выбор канала"
									>
										<SelectValue
											placeholder={
												channelsEmpty
													? "Нет доступных каналов"
													: "Выберите канал"
											}
										/>
									</SelectTrigger>
									<SelectContent>
										{channelsEmpty ? (
											<SelectItem value="__none" disabled>
												Нет доступных каналов
											</SelectItem>
										) : (
											channelList.map((ch) => (
												<SelectItem key={ch.id} value={ch.id}>
													{ch.name || ch.id.slice(0, 8)}
												</SelectItem>
											))
										)}
									</SelectContent>
								</Select>
							)}
						</div>

						{/* Участники */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callParticipantsLabel", { defaultValue: "Участники" })}
							</Label>

							{/* Чипы выбранных участников */}
							{participantIds.length > 0 && (
								<div className="flex flex-wrap gap-1.5">
									{participantIds.map((uid) => {
										const member = (members ?? []).find(
											(m: MemberResponse) => m.user_id === uid,
										);
										if (!member) return null;
										return (
											<span
												key={uid}
												className="inline-flex items-center gap-1 rounded-full border border-border bg-raised px-2.5 py-1 text-xs font-medium"
											>
												{member.avatar_url ? (
													<img
														src={member.avatar_url}
														alt=""
														className="h-4 w-4 rounded-full object-cover"
													/>
												) : (
													<span className="flex h-4 w-4 items-center justify-center rounded-full bg-primary/20 text-[10px] font-semibold text-primary">
														{member.name.charAt(0).toUpperCase()}
													</span>
												)}
												{member.name}
												<button
													type="button"
													onClick={() => removeParticipant(uid)}
													aria-label={`Удалить участника ${member.name}`}
													className="ml-0.5 rounded-full p-0.5 text-t2 hover:bg-destructive/20 hover:text-destructive transition-colors"
												>
													<X className="h-3 w-3" aria-hidden="true" />
												</button>
											</span>
										);
									})}
								</div>
							)}

							{/* Поле поиска участников */}
							<Input
								placeholder={t("searchParticipants", {
									defaultValue: "Поиск участников...",
								})}
								value={participantSearch}
								onChange={(e) => setParticipantSearch(e.target.value)}
								className="h-10"
							/>

							{/* Список участников */}
							<div className="max-h-32 overflow-y-auto rounded-lg border">
								{filteredMembers.length === 0 ? (
									<div className="px-3 py-2 text-xs text-t2">
										{t("noParticipantsFound", {
											defaultValue: "Участники не найдены",
										})}
									</div>
								) : (
									filteredMembers.map((member: MemberResponse) => {
										const isSelected = participantIds.includes(member.user_id);
										return (
											<button
												key={member.user_id}
												type="button"
												className={`flex w-full items-center gap-2.5 px-3 py-2 text-sm transition-colors hover:bg-raised/50 ${
													isSelected ? "bg-primary/5" : ""
												}`}
												onClick={() => toggleParticipant(member.user_id)}
											>
												{member.avatar_url ? (
													<img
														src={member.avatar_url}
														alt=""
														className="h-6 w-6 rounded-full object-cover"
													/>
												) : (
													<span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-xs font-semibold text-primary">
														{member.name.charAt(0).toUpperCase()}
													</span>
												)}
												<span className="flex-1 text-left truncate">
													{member.name}
												</span>
												{isSelected && (
													<Check className="h-4 w-4 text-primary shrink-0" />
												)}
											</button>
										);
									})
								)}
							</div>
						</div>

						{/* Время */}
						<div className="space-y-2">
							<Label htmlFor="call-time" className="text-sm font-medium">
								{t("callTimeLabel", { defaultValue: "Время начала" })}
							</Label>
							<div className="flex gap-3">
								<Input
									id="call-time"
									type="time"
									value={time}
									onChange={(e) => setTime(e.target.value)}
									required
									className="h-11 w-32"
								/>
								{/* Скрытый input даты — нужен для submit формы */}
								<Input
									type="date"
									value={date}
									onChange={(e) => setDate(e.target.value)}
									required
									className="h-11 flex-1"
								/>
							</div>
						</div>

						<div className="space-y-2">
							<Label htmlFor="call-timezone" className="text-sm font-medium">
								{t("meetingTimezoneLabel")}
							</Label>
							<Select value={timezone} onValueChange={setTimezone}>
								<SelectTrigger id="call-timezone" className="h-11 rounded-xl">
									<SelectValue placeholder={t("meetingTimezonePlaceholder")} />
								</SelectTrigger>
								<SelectContent>
									{timeZoneOptions.map((timeZoneOption) => (
										<SelectItem key={timeZoneOption} value={timeZoneOption}>
											{timeZoneOption}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						{/* Длительность — кнопки быстрого выбора */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callDurationLabel", { defaultValue: "Длительность" })}
							</Label>
							<div className="flex gap-2">
								{DURATION_OPTIONS.map((opt) => (
									<button
										key={opt.value}
										type="button"
										className={`flex-1 rounded-lg border px-3 py-2 text-sm font-medium transition-colors ${
											duration === opt.value
												? "border-primary bg-primary/10 text-primary"
												: "border-border bg-base text-t2 hover:border-primary/50 hover:text-t1"
										}`}
										onClick={() => setDuration(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>

						{/* Повторение */}
						<div className="space-y-2">
							<Label className="text-sm font-medium">
								{t("callRecurrenceLabel", { defaultValue: "Повтор" })}
							</Label>
							<div className="flex flex-wrap gap-2">
								{RECURRENCE_OPTIONS.map((opt) => (
									<button
										key={opt.value}
										type="button"
										className={`rounded-lg border px-3 py-2 text-sm font-medium transition-colors ${
											recurrence === opt.value
												? "border-primary bg-primary/10 text-primary"
												: "border-border bg-base text-t2 hover:border-primary/50 hover:text-t1"
										}`}
										onClick={() => setRecurrence(opt.value)}
									>
										{opt.label}
									</button>
								))}
							</div>
						</div>
					</div>

					{/* Приватный итог проверки доступности без деталей календаря */}
					{availabilityQueryEnabled && (
						<div className="px-6 mt-3">
							<div
								className={`flex items-start gap-2 rounded-lg border px-3 py-2.5 ${
									availabilitySuccess &&
									!availabilityFetching &&
									availabilityAssessment === "available"
										? "border-emerald-500/30 bg-emerald-500/10"
										: "border-amber-500/30 bg-amber-500/10"
								}`}
								role="status"
								aria-live="polite"
							>
								{availabilitySuccess &&
								!availabilityFetching &&
								availabilityAssessment === "available" ? (
									<Check
										className="h-4 w-4 shrink-0 text-emerald-500 mt-0.5"
										aria-hidden="true"
									/>
								) : (
									<AlertTriangle
										className="h-4 w-4 shrink-0 text-amber-500 mt-0.5"
										aria-hidden="true"
									/>
								)}
								<div className="flex-1 space-y-2">
									<p className="text-xs text-t2">
										{availabilityFetching
											? "Проверяем доступность участников…"
											: availabilityError ||
													availabilityAssessment === "incomplete"
												? "Не удалось подтвердить доступность участников"
												: availabilityAssessment === "busy"
													? "Выбранное время занято у одного или нескольких участников"
													: "Все участники свободны"}
									</p>
									{!availabilityFetching &&
										(availabilityError ||
											availabilityAssessment === "incomplete") && (
											<button
												type="button"
												className="text-xs font-medium text-primary hover:underline"
												onClick={() => void refetchAvailability()}
											>
												Проверить ещё раз
											</button>
										)}
								</div>
							</div>
						</div>
					)}

					{/* Футер диалога */}
					<div className="px-6 py-5 mt-2">
						<Button
							type="submit"
							disabled={!isValid || (mutation.isPending && mutation.variables?.session === dialogSession)}
							className="w-full h-11 text-base font-semibold"
							size="lg"
						>
							{mutation.isPending
								? t("creating", { defaultValue: "Создание..." })
								: t("create", { defaultValue: "Создать" })}
						</Button>
					</div>
				</form>
			</DialogContent>
		</Dialog>
	);
}
