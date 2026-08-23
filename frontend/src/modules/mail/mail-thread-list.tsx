import { useInfiniteQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { getMailThreads, type MailThreadSummary } from "@/api/mail";
import { Button } from "@/components/ui/button";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { cn } from "@/lib/utils";
import {
	normalizeMailRouteSearchQuery,
	type MailRouteSearch,
	Route as MailRoute,
} from "@/routes/_authenticated/mail";
import { mailKeys } from "./mail-query-keys";
import { MailThreadRow } from "./mail-thread-row";

const MAIL_THREAD_PAGE_SIZE = 40;
const MAIL_UUID_PATTERN =
	/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const MAIL_NIL_UUID = "00000000-0000-0000-0000-000000000000";

interface MailThreadListProps {
	folderKey: string;
	selectedThreadId?: string;
	className?: string;
}

export function isValidMailEntityID(value: string): boolean {
	return MAIL_UUID_PATTERN.test(value) && value.toLowerCase() !== MAIL_NIL_UUID;
}

export function isValidMailFolderKey(value: string): boolean {
	return value === "all" || isValidMailEntityID(value);
}

function uniqueThreads(pages: { items: MailThreadSummary[] }[] | undefined) {
	const known = new Set<string>();
	return (pages ?? []).flatMap((page) =>
		page.items.filter((thread) => {
			if (known.has(thread.id)) return false;
			known.add(thread.id);
			return true;
		}),
	);
}

function worstProjectionState(
	pages: { projection_state: string }[] | undefined,
): "bootstrapping" | "rebuilding" | "degraded" | null {
	const states = pages?.map((page) => page.projection_state) ?? [];
	if (states.includes("degraded")) return "degraded";
	if (states.includes("rebuilding")) return "rebuilding";
	if (states.includes("bootstrapping")) return "bootstrapping";
	return null;
}

function ThreadListSkeleton() {
	const { t } = useTranslation("mail");
	return (
		<div className="divide-y divide-b1" aria-busy="true" aria-live="polite">
			<p className="sr-only">{t("threadList.loading")}</p>
			{[1, 2, 3, 4, 5].map((index) => (
				<div key={index} className="space-y-2 px-3 py-3">
					<div className="h-4 w-3/4 rounded animate-shimmer" />
					<div className="h-3 w-2/5 rounded animate-shimmer" />
					<div className="h-3 w-full rounded animate-shimmer" />
				</div>
			))}
		</div>
	);
}

export function MailThreadList({
	folderKey,
	selectedThreadId,
	className,
}: MailThreadListProps) {
	const { t } = useTranslation("mail");
	const navigate = useNavigate();
	const mailSearch = MailRoute.useSearch();
	const [searchInput, setSearchInput] = useState(mailSearch.q ?? "");
	const debouncedSearch = useDebouncedValue(searchInput, 300);
	const mailSearchRef = useRef(mailSearch);
	const folderID = folderKey === "all" ? undefined : folderKey.toLowerCase();
	const hasActiveFilters = Boolean(
		mailSearch.q ||
		mailSearch.unread ||
		mailSearch.starred ||
		mailSearch.has_attachments,
	);

	const updateSearch = useCallback(
		(nextSearch: MailRouteSearch) => {
			void navigate({ search: nextSearch, replace: true });
		},
		[navigate],
	);

	const applySearch = useCallback(
		(value: string) => {
			const q = normalizeMailRouteSearchQuery(value);
			setSearchInput(q ?? "");
			if (q === mailSearch.q && value === (mailSearch.q ?? "")) return;
			updateSearch({ ...mailSearch, q });
		},
		[mailSearch, updateSearch],
	);

	useEffect(() => {
		mailSearchRef.current = mailSearch;
	}, [mailSearch]);

	useEffect(() => {
		setSearchInput(mailSearch.q ?? "");
	}, [mailSearch.q]);

	useEffect(() => {
		const currentSearch = mailSearchRef.current;
		const q = normalizeMailRouteSearchQuery(debouncedSearch);
		if (q === currentSearch.q && debouncedSearch === (currentSearch.q ?? "")) {
			return;
		}
		setSearchInput(q ?? "");
		updateSearch({ ...currentSearch, q });
	}, [debouncedSearch, updateSearch]);

	const threadQueryParams = useMemo(
		() => ({
			folder_id: folderID,
			search: mailSearch.q,
			unread: mailSearch.unread ? true : undefined,
			starred: mailSearch.starred ? true : undefined,
			has_attachments: mailSearch.has_attachments ? true : undefined,
			limit: MAIL_THREAD_PAGE_SIZE,
		}),
		[
			folderID,
			mailSearch.has_attachments,
			mailSearch.q,
			mailSearch.starred,
			mailSearch.unread,
		],
	);
	const threadsQuery = useInfiniteQuery({
		queryKey: mailKeys.threads(threadQueryParams),
		queryFn: ({ pageParam }) =>
			getMailThreads({
				...threadQueryParams,
				cursor: pageParam,
			}),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (
			lastPage,
			_allPages,
			_lastPageParam,
			allPageParams,
		) => {
			const nextCursor = lastPage.next_cursor ?? undefined;
			return nextCursor && !allPageParams.includes(nextCursor)
				? nextCursor
				: undefined;
		},
		retry: false,
		refetchOnWindowFocus: false,
	});

	const threads = useMemo(
		() => uniqueThreads(threadsQuery.data?.pages),
		[threadsQuery.data?.pages],
	);
	const projectionState = worstProjectionState(threadsQuery.data?.pages);
	const hasLoadedPages = (threadsQuery.data?.pages.length ?? 0) > 0;

	let content;
	if (threadsQuery.isPending) {
		content = <ThreadListSkeleton />;
	} else if ((threadsQuery.isError || !threadsQuery.data) && !hasLoadedPages) {
		content = (
			<div className="space-y-3 px-4 py-6" role="alert">
				<p className="text-sm text-t2">{t("threadList.unavailable")}</p>
				<Button
					size="sm"
					variant="outline"
					onClick={() => void threadsQuery.refetch()}
				>
					{t("retry")}
				</Button>
			</div>
		);
	} else if (threadsQuery.isError && threads.length === 0) {
		content = (
			<div className="space-y-3 bg-[var(--amber)]/10 px-4 py-6" role="alert">
				<p className="text-sm text-t2">{t("threadList.partialUnavailable")}</p>
				<p className="text-xs leading-5 text-t3">
					{t("threadList.partialEmpty")}
				</p>
				<Button
					size="sm"
					variant="outline"
					onClick={() => void threadsQuery.refetch()}
				>
					{t("retry")}
				</Button>
			</div>
		);
	} else if (threads.length === 0) {
		content = (
			<div className="px-4 py-10 text-center">
				<p className="text-sm font-medium text-t2">
					{t(
						hasActiveFilters
							? "threadList.filteredEmpty"
							: "threadList.empty",
					)}
				</p>
				<p className="mt-1 text-xs leading-5 text-t3">
					{t(
						hasActiveFilters
							? "threadList.filteredEmptyDescription"
							: "threadList.emptyDescription",
					)}
				</p>
			</div>
		);
	} else {
		content = (
			<>
				{threadsQuery.isError ? (
					<div className="flex items-center justify-between gap-3 border-b border-b1 bg-[var(--amber)]/10 px-3 py-2" role="alert">
						<p className="text-xs text-t2">{t("threadList.partialUnavailable")}</p>
						<Button
							size="sm"
							variant="ghost"
							className="shrink-0"
							onClick={() => void threadsQuery.refetch()}
						>
							{t("retry")}
						</Button>
					</div>
				) : null}
				<ul aria-label={t("threadList.listLabel")} className="divide-y divide-b1">
					{threads.map((thread) => (
						<MailThreadRow
							key={thread.id}
							thread={thread}
							folderKey={folderKey}
							selected={selectedThreadId?.toLowerCase() === thread.id.toLowerCase()}
						/>
					))}
				</ul>
				<div className="border-t border-b1 px-3 py-3 text-center">
					{threadsQuery.hasNextPage ? (
						<Button
							size="sm"
							variant="ghost"
							disabled={threadsQuery.isFetchingNextPage}
							onClick={() => void threadsQuery.fetchNextPage()}
						>
							{threadsQuery.isFetchingNextPage
								? t("threadList.loadingMore")
								: t("threadList.loadMore")}
						</Button>
					) : (
						<p className="text-xs text-t3">{t("threadList.end")}</p>
					)}
				</div>
			</>
		);
	}

	return (
		<section
			aria-labelledby="mail-thread-list-heading"
			className={cn(
				"min-w-0 flex-1 bg-card md:max-w-[23rem] md:flex-none md:border-r md:border-b1",
				className,
			)}
		>
			<header className="border-b border-b1 px-4 py-3">
				<div className="flex min-h-8 items-center justify-between gap-3">
					<div>
						<h2 id="mail-thread-list-heading" className="text-sm font-semibold text-t1">
							{t(folderKey === "all" ? "threadList.title" : "threadList.folderTitle")}
						</h2>
						{projectionState ? (
							<p className="mt-0.5 text-xs text-t3">
								{t(`threadList.${projectionState}`)}
							</p>
						) : null}
					</div>
					{threads.length > 0 ? (
						<span className="text-xs tabular-nums text-t3">
							{t("threadList.count", { count: threads.length })}
						</span>
					) : null}
				</div>
				<div className="mt-3 flex flex-wrap items-center gap-2">
					<form
						className="min-w-0 flex-1 basis-44"
						onSubmit={(event) => {
							event.preventDefault();
							applySearch(searchInput);
						}}
					>
						<label htmlFor="mail-thread-search" className="sr-only">
							{t("threadList.searchLabel")}
						</label>
						<input
							id="mail-thread-search"
							type="search"
							aria-label={t("threadList.searchLabel")}
							value={searchInput}
							onChange={(event) => setSearchInput(event.target.value)}
							placeholder={t("threadList.searchPlaceholder")}
							className="min-h-11 w-full rounded-[var(--r2)] border border-b1 bg-base px-3 text-sm text-t1 placeholder:text-t3 outline-none transition-colors focus-visible:border-[var(--accent)] focus-visible:ring-2 focus-visible:ring-[var(--accent)]/30"
						/>
					</form>
					<button
						type="button"
						aria-pressed={Boolean(mailSearch.unread)}
						onClick={() =>
							updateSearch({
								...mailSearch,
								unread: mailSearch.unread ? undefined : true,
							})
						}
						className={cn(
							"min-h-11 rounded-[var(--r2)] border px-3 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]/30",
							mailSearch.unread
								? "border-[var(--accent)] bg-[var(--accent)]/10 text-t1"
								: "border-b1 bg-base text-t2 hover:text-t1",
						)}
					>
						{t("threadList.filterUnread")}
					</button>
					<button
						type="button"
						aria-pressed={Boolean(mailSearch.starred)}
						onClick={() =>
							updateSearch({
								...mailSearch,
								starred: mailSearch.starred ? undefined : true,
							})
						}
						className={cn(
							"min-h-11 rounded-[var(--r2)] border px-3 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]/30",
							mailSearch.starred
								? "border-[var(--accent)] bg-[var(--accent)]/10 text-t1"
								: "border-b1 bg-base text-t2 hover:text-t1",
						)}
					>
						{t("threadList.filterStarred")}
					</button>
					<button
						type="button"
						aria-pressed={Boolean(mailSearch.has_attachments)}
						onClick={() =>
							updateSearch({
								...mailSearch,
								has_attachments: mailSearch.has_attachments
									? undefined
									: true,
							})
						}
						className={cn(
							"min-h-11 rounded-[var(--r2)] border px-3 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]/30",
							mailSearch.has_attachments
								? "border-[var(--accent)] bg-[var(--accent)]/10 text-t1"
								: "border-b1 bg-base text-t2 hover:text-t1",
						)}
					>
						{t("threadList.filterHasAttachments")}
					</button>
					{hasActiveFilters ? (
						<button
							type="button"
							onClick={() => {
								setSearchInput("");
								updateSearch({});
							}}
							className="min-h-11 rounded-[var(--r2)] px-3 text-sm font-medium text-t2 transition-colors hover:text-t1 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]/30"
						>
							{t("threadList.clearFilters")}
						</button>
					) : null}
				</div>
			</header>
			{content}
		</section>
	);
}
