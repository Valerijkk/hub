import { useQuery } from "@tanstack/react-query";
import { Link, useRouterState } from "@tanstack/react-router";
import { Folder, Inbox } from "lucide-react";
import { useTranslation } from "react-i18next";
import { getMailFolders, type MailFolder } from "@/api/mail";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { mailKeys } from "./mail-query-keys";

interface FolderNode {
	folder: MailFolder;
	children: FolderNode[];
}

function compareFolders(left: MailFolder, right: MailFolder) {
	return (
		left.sort_order - right.sort_order ||
		left.name.localeCompare(right.name) ||
		left.id.localeCompare(right.id)
	);
}

function buildFolderTree(folders: MailFolder[]): FolderNode[] {
	const byID = new Map(folders.map((folder) => [folder.id, folder]));
	const children = new Map<string | null, MailFolder[]>();
	for (const folder of folders) {
		const parentID =
			folder.parent_id && byID.has(folder.parent_id) ? folder.parent_id : null;
		const group = children.get(parentID) ?? [];
		group.push(folder);
		children.set(parentID, group);
	}

	const visited = new Set<string>();
	const visit = (parentID: string | null): FolderNode[] =>
		(children.get(parentID) ?? [])
			.sort(compareFolders)
			.flatMap((folder) => {
				if (visited.has(folder.id)) return [];
				visited.add(folder.id);
				return [{ folder, children: visit(folder.id) }];
			});

	const roots = visit(null);
	const unlinked = [...byID.values()]
		.filter((folder) => !visited.has(folder.id))
		.sort(compareFolders)
		.map((folder) => {
			visited.add(folder.id);
			return { folder, children: visit(folder.id) };
		});
	return [...roots, ...unlinked];
}

function FolderTree({
	nodes,
	activeFolderKey,
	depth = 0,
}: {
	nodes: FolderNode[];
	activeFolderKey: string;
	depth?: number;
}) {
	const { t } = useTranslation("mail");
	return (
		<ul className={depth === 0 ? "space-y-0.5" : "mt-0.5 space-y-0.5"}>
			{nodes.map(({ folder, children }) => (
				<li key={folder.id}>
					<Link
						to="/mail/$folderKey"
						params={{ folderKey: folder.id }}
						aria-current={activeFolderKey === folder.id ? "page" : undefined}
						className={cn(
							"flex min-h-10 items-center gap-2 rounded-[var(--r2)] px-2 text-sm text-t2 outline-none transition-colors hover:bg-[var(--hover)] focus-visible:ring-2 focus-visible:ring-[var(--blue)]",
							activeFolderKey === folder.id && "bg-[var(--blue)]/10 text-t1",
						)}
						style={{ paddingInlineStart: String(8 + depth * 12) + "px" }}
					>
						<Folder aria-hidden="true" className="h-4 w-4 shrink-0 text-t3" />
						<span className="min-w-0 flex-1 truncate text-t1">{folder.name}</span>
						{folder.unread_count > 0 ? (
							<span
								className="shrink-0 rounded-full bg-[var(--blue)]/10 px-1.5 py-0.5 text-xs font-semibold tabular-nums text-[var(--blue)]"
								aria-label={t("sidebar.unreadCount", { count: folder.unread_count })}
							>
								{folder.unread_count}
							</span>
						) : null}
					</Link>
					{children.length > 0 ? (
						<FolderTree
							nodes={children}
							activeFolderKey={activeFolderKey}
							depth={depth + 1}
						/>
					) : null}
				</li>
			))}
		</ul>
	);
}

function FolderState({ state }: { state: string }) {
	const { t } = useTranslation("mail");
	if (state === "bootstrapping") return <p className="px-2 text-xs text-t3">{t("sidebar.bootstrapping")}</p>;
	if (state === "rebuilding") return <p className="px-2 text-xs text-t3">{t("sidebar.rebuilding")}</p>;
	if (state === "degraded") return <p className="px-2 text-xs text-[var(--amber)]">{t("sidebar.degraded")}</p>;
	return null;
}

export function MailSidebar() {
	const { t } = useTranslation("mail");
	const pathname = useRouterState({ select: (state) => state.location.pathname });
	const segments = pathname.split("/").filter(Boolean);
	const activeFolderKey =
		segments[0] === "mail" ? (segments[1] ?? "all").toLowerCase() : "all";
	const threadOpen = segments[0] === "mail" && segments[2] === "thread" && Boolean(segments[3]);
	const foldersQuery = useQuery({
		queryKey: mailKeys.folders(),
		queryFn: getMailFolders,
		retry: false,
	});

	let content;
	if (foldersQuery.isPending) {
		content = (
			<div className="space-y-2 px-2 py-1" aria-busy="true" aria-live="polite">
				<p className="text-sm text-t3">{t("sidebar.loading")}</p>
				<div className="h-8 w-full rounded-[var(--r2)] animate-shimmer" />
				<div className="h-8 w-4/5 rounded-[var(--r2)] animate-shimmer" />
			</div>
		);
	} else if (foldersQuery.isError || !foldersQuery.data) {
		content = (
			<div className="space-y-3 px-2 py-1" role="alert">
				<p className="text-sm text-t2">{t("sidebar.unavailable")}</p>
				<Button variant="outline" size="sm" className="min-h-11" onClick={() => void foldersQuery.refetch()}>
					{t("retry")}
				</Button>
			</div>
		);
	} else if (foldersQuery.data.items.length === 0) {
		content = (
			<div className="space-y-1 px-2 py-1">
				<p className="text-sm text-t2">{t("sidebar.empty")}</p>
				<p className="text-xs leading-5 text-t3">{t("sidebar.emptyDescription")}</p>
				<FolderState state={foldersQuery.data.projection_state} />
			</div>
		);
	} else {
		content = (
			<div className="space-y-3">
				<FolderTree
					nodes={buildFolderTree(foldersQuery.data.items)}
					activeFolderKey={activeFolderKey}
				/>
				<FolderState state={foldersQuery.data.projection_state} />
			</div>
		);
	}

	return (
		<aside
			className={cn(
				"w-full shrink-0 border-b border-b1 bg-raised/40 px-3 py-4 md:w-64 md:border-r md:border-b-0",
				threadOpen && "hidden md:block",
			)}
			aria-labelledby="mail-folders-heading"
		>
			<nav aria-label={t("sidebar.navigation")}>
				<h2 id="mail-folders-heading" className="mb-3 px-2 text-xs font-semibold uppercase tracking-[0.08em] text-t3">{t("sidebar.title")}</h2>
				<Link
					to="/mail/$folderKey"
					params={{ folderKey: "all" }}
					aria-current={activeFolderKey === "all" ? "page" : undefined}
					className={cn(
						"mb-2 flex min-h-10 items-center gap-2 rounded-[var(--r2)] px-2 text-sm text-t2 outline-none transition-colors hover:bg-[var(--hover)] focus-visible:ring-2 focus-visible:ring-[var(--blue)]",
						activeFolderKey === "all" && "bg-[var(--blue)]/10 text-t1",
					)}
				>
					<Inbox aria-hidden="true" className="h-4 w-4 shrink-0 text-t3" />
					<span className="min-w-0 flex-1 truncate text-t1">{t("sidebar.allMail")}</span>
				</Link>
				{content}
			</nav>
		</aside>
	);
}
