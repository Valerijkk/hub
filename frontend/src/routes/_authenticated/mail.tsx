import {
	createFileRoute,
	Outlet,
	useNavigate,
	useRouterState,
} from "@tanstack/react-router";
import { useEffect } from "react";
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

function isCanonicalMailRouteSearch(
	search: Record<string, unknown>,
	canonical: MailRouteSearch,
): boolean {
	return (
		Object.keys(search).length === Object.keys(canonical).length &&
		Object.entries(canonical).every(
			([key, value]) => search[key] === value,
		)
	);
}

function MailRoute() {
	const navigate = useNavigate();
	const rawSearch = useRouterState({
		select: (state) => state.location.search,
	});
	const mailSearch = Route.useSearch();

	useEffect(() => {
		if (isCanonicalMailRouteSearch(rawSearch, mailSearch)) return;
		void navigate({ search: mailSearch, replace: true });
	}, [mailSearch, navigate, rawSearch]);

	return (
		<MailShell>
			<Outlet />
		</MailShell>
	);
}

export const Route = createFileRoute("/_authenticated/mail")({
	validateSearch: (search: Record<string, unknown>): MailRouteSearch => {
		const q = normalizeMailRouteSearchQuery(search.q);
		return {
			...(q === undefined ? {} : { q }),
			...(search.unread === true ? { unread: true } : {}),
			...(search.starred === true ? { starred: true } : {}),
			...(search.has_attachments === true ? { has_attachments: true } : {}),
		};
	},
	component: MailRoute,
});
