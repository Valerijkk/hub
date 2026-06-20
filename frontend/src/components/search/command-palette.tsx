import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
	AlertTriangle,
	CircleDot,
	FlaskConical,
	FolderKanban,
	Hash,
	HelpCircle,
	LayoutDashboard,
	MessageSquare,
	Paperclip,
	Plus,
	SearchX,
	User,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { searchGlobal, type SearchResult } from "@/api/search";
import { Button } from "@/components/ui/button";
import {
	CommandDialog,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "@/components/ui/command";
import { EmptyState } from "@/components/ui/empty-state";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useCommandStore } from "@/stores/command-store";

const GROUP_ORDER: SearchResult["type"][] = [
	"issue",
	"project",
	"test_case",
	"channel",
	"message",
	"user",
	"file",
];

const TYPE_ICON = {
	issue: CircleDot,
	project: FolderKanban,
	test_case: FlaskConical,
	channel: Hash,
	message: MessageSquare,
	user: User,
	file: Paperclip,
} as const;

function CommandResultSkeleton() {
	return (
		<div className="flex items-center gap-3 px-3 py-2" aria-hidden="true">
			<div className="h-4 w-4 shrink-0 animate-pulse rounded bg-[var(--hover)]" />
			<div className="h-3 flex-1 animate-pulse rounded bg-[var(--hover)]" />
		</div>
	);
}

/** Глобальный поиск Ctrl+K — US-M07 / UX-P0-01. */
export function CommandPalette() {
	const { t } = useTranslation("common");
	const open = useCommandStore((s) => s.open);
	const setOpen = useCommandStore((s) => s.setOpen);
	const navigate = useNavigate();
	const [query, setQuery] = useState("");

	const debounced = useDebouncedValue(query.trim(), 300);
	const enabled = debounced.length >= 2;

	const { data, isFetching, isError, refetch } = useQuery({
		queryKey: ["global-search", debounced],
		queryFn: () => searchGlobal(debounced, undefined, 20),
		enabled,
		staleTime: 30_000,
		placeholderData: keepPreviousData,
	});

	useEffect(() => {
		if (!open) setQuery("");
	}, [open]);

	const handleSelect = useCallback(
		(result: SearchResult) => {
			setOpen(false);
			setQuery("");
			if (result.type === "file") {
				window.open(result.url, "_blank", "noopener,noreferrer");
				return;
			}
			void navigate({ to: result.url });
		},
		[navigate, setOpen],
	);

	const go = useCallback(
		(path: string) => {
			setOpen(false);
			setQuery("");
			void navigate({ to: path });
		},
		[navigate, setOpen],
	);

	const ACTION_ITEMS = [
		{ icon: FolderKanban, label: t("command.goProjects"), path: "/projects" },
		{ icon: MessageSquare, label: t("command.goMessenger"), path: "/messenger" },
		{ icon: LayoutDashboard, label: t("command.goDashboard"), path: "/dashboard" },
		{ icon: Plus, label: t("command.createProject"), path: "/projects/new" },
		{ icon: HelpCircle, label: t("command.goHelp"), path: "/help" },
	] as const;

	const results = data?.results ?? [];
	const grouped = GROUP_ORDER.map((type) => ({
		type,
		items: results.filter((r) => r.type === type),
	})).filter((g) => g.items.length > 0);

	return (
		<CommandDialog open={open} onOpenChange={setOpen}>
			<CommandInput
				value={query}
				onValueChange={setQuery}
				placeholder={t("search.placeholderExtended")}
				aria-label={t("search.globalSearch")}
			/>
			<CommandList aria-busy={isFetching}>
				{!enabled && (
					<>
						<CommandGroup heading={t("search.groups.actions")}>
							{ACTION_ITEMS.map(({ icon: Icon, label, path }) => (
								<CommandItem key={path} value={path} onSelect={() => go(path)}>
									<Icon className="h-4 w-4 shrink-0 text-t3" aria-hidden="true" />
									<span>{label}</span>
								</CommandItem>
							))}
						</CommandGroup>
						<CommandEmpty>{t("search.startTyping")}</CommandEmpty>
					</>
				)}

				{enabled && isError && (
					<div className="px-2 py-4">
						<EmptyState
							icon={AlertTriangle}
							title={t("search.error")}
							description={t("search.errorHint")}
						/>
						<div className="flex justify-center pb-2">
							<Button type="button" variant="outline" size="sm" onClick={() => void refetch()}>
								{t("btn.retry")}
							</Button>
						</div>
					</div>
				)}

				{enabled && isFetching && results.length === 0 && (
					<div role="status" aria-live="polite">
						<span className="sr-only">{t("search.loading")}</span>
						{Array.from({ length: 5 }, (_, i) => (
							<CommandResultSkeleton key={`sk-${i}`} />
						))}
					</div>
				)}

				{enabled && !isFetching && !isError && results.length === 0 && (
					<CommandEmpty>
						<EmptyState
							icon={SearchX}
							title={t("search.noResults")}
							description={t("search.noResultsHint")}
						/>
					</CommandEmpty>
				)}

				{grouped.map(({ type, items }) => {
					const Icon = TYPE_ICON[type];
					return (
						<CommandGroup key={type} heading={t(`search.groups.${type}`)}>
							{items.map((result) => (
								<CommandItem
									key={result.id}
									value={result.id}
									onSelect={() => handleSelect(result)}
								>
									<Icon className="h-4 w-4 shrink-0 text-t3" aria-hidden="true" />
									<div className="min-w-0 flex-1">
										<p className="truncate text-t1">{result.title}</p>
										{result.subtitle && (
											<p className="truncate text-xs text-t3">{result.subtitle}</p>
										)}
									</div>
								</CommandItem>
							))}
						</CommandGroup>
					);
				})}
			</CommandList>
		</CommandDialog>
	);
}
