import { createFileRoute, Outlet } from "@tanstack/react-router";
import { MailShell } from "@/modules/mail/mail-shell";

const MAIL_THREAD_MAX_SEARCH_BYTES = 128;
const MAIL_QUERY_CONTROL_CHARACTERS = /[\u0000-\u001F\u007F]/u;

export interface MailRouteSearch {
	q?: string;
	unread?: true;
	starred?: true;
	has_attachments?: true;
}

export function normalizeMailRouteSearchQuery(
	value: unknown,
): string | undefined {
	if (typeof value !== "string" || MAIL_QUERY_CONTROL_CHARACTERS.test(value)) {
		return undefined;
	}

	const normalized = value
		.normalize("NFC")
		.toLocaleLowerCase("und")
		.normalize("NFC")
		.replace(/\s+/gu, " ")
		.trim();

	if (
		normalized.length === 0 ||
		MAIL_QUERY_CONTROL_CHARACTERS.test(normalized) ||
		/[%_\\]/.test(normalized) ||
		new TextEncoder().encode(normalized).length > MAIL_THREAD_MAX_SEARCH_BYTES
	) {
		return undefined;
	}

	return normalized;
}

function MailRoute() {
	return (
		<MailShell>
			<Outlet />
		</MailShell>
	);
}

export const Route = createFileRoute("/_authenticated/mail")({
	validateSearch: (search: Record<string, unknown>): MailRouteSearch => ({
		q: normalizeMailRouteSearchQuery(search.q),
		unread: search.unread === true ? true : undefined,
		starred: search.starred === true ? true : undefined,
		has_attachments: search.has_attachments === true ? true : undefined,
	}),
	component: MailRoute,
});
