import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import type { ChannelInvite } from "@/api/messenger";
import {
	createInviteLink,
	getChannelInvites,
	revokeInvite,
} from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Константы
// ---------------------------------------------------------------------------

/** Варианты срока действия пригласительной ссылки. */
const EXPIRATION_OPTIONS = [
	{ value: "1h", label: "1 час" },
	{ value: "1d", label: "1 день" },
	{ value: "7d", label: "7 дней" },
	{ value: "30d", label: "30 дней" },
	{ value: "", label: "Бессрочно" },
] as const;

/** Варианты лимита использований. */
const USAGE_LIMIT_OPTIONS = [
	{ value: 0, label: "Без лимита" },
	{ value: 1, label: "1" },
	{ value: 5, label: "5" },
	{ value: 10, label: "10" },
	{ value: 25, label: "25" },
	{ value: 50, label: "50" },
	{ value: 100, label: "100" },
] as const;

// ---------------------------------------------------------------------------
// Утилиты
// ---------------------------------------------------------------------------

/** Форматирует дату истечения для отображения. */
function formatExpiration(expiresAt?: string): string {
	if (!expiresAt) return "Бессрочно";
	const date = new Date(expiresAt);
	if (date < new Date()) return "Истекла";
	return date.toLocaleString("ru-RU", {
		day: "numeric",
		month: "short",
		hour: "2-digit",
		minute: "2-digit",
	});
}

/** Форматирует лимит использований. */
function formatUsage(invite: ChannelInvite): string {
	if (invite.max_uses === 0) {
		return `${invite.use_count} исп.`;
	}
	return `${invite.use_count} / ${invite.max_uses}`;
}

/** Копирует текст в буфер обмена. */
async function copyToClipboard(text: string): Promise<boolean> {
	try {
		await navigator.clipboard.writeText(text);
		return true;
	} catch {
		return false;
	}
}

/** Формирует полный URL приглашения. */
function getFullInviteUrl(token: string): string {
	return `${window.location.origin}/invite/${token}`;
}

// ---------------------------------------------------------------------------
// Компонент формы создания ссылки
// ---------------------------------------------------------------------------

/** Форма создания новой пригласительной ссылки. */
function CreateInviteForm({
	channelId,
	onCreated,
}: {
	channelId: string;
	onCreated: () => void;
}) {
	const [expiresIn, setExpiresIn] = useState<string>("7d");
	const [maxUses, setMaxUses] = useState<number>(0);
	const [isCreating, setIsCreating] = useState(false);
	const [copiedToken, setCopiedToken] = useState<string | null>(null);

	const handleCreate = useCallback(async () => {
		setIsCreating(true);
		try {
			const invite = await createInviteLink(channelId, {
				max_uses: maxUses,
				expires_in: expiresIn,
			});
			// Копируем ссылку в буфер.
			const url = getFullInviteUrl(invite.token);
			const copied = await copyToClipboard(url);
			if (copied) {
				setCopiedToken(invite.token);
				setTimeout(() => setCopiedToken(null), 2000);
			}
			onCreated();
		} finally {
			setIsCreating(false);
		}
	}, [channelId, expiresIn, maxUses, onCreated]);

	return (
		<div className="space-y-4">
			{/* Срок действия */}
			<div>
				<span className="text-xs font-medium text-t2 mb-1.5 block">
					Срок действия
				</span>
				<div className="flex flex-wrap gap-1.5">
					{EXPIRATION_OPTIONS.map((opt) => (
						<button
							key={opt.value}
							type="button"
							className={cn(
								"rounded-md px-3 py-1.5 text-xs font-medium transition-colors",
								expiresIn === opt.value
									? "bg-primary text-primary-foreground"
									: "bg-raised text-t2 hover:bg-raised/80",
							)}
							onClick={() => setExpiresIn(opt.value)}
						>
							{opt.label}
						</button>
					))}
				</div>
			</div>

			{/* Лимит использований */}
			<div>
				<span className="text-xs font-medium text-t2 mb-1.5 block">
					Лимит использований
				</span>
				<div className="flex flex-wrap gap-1.5">
					{USAGE_LIMIT_OPTIONS.map((opt) => (
						<button
							key={opt.value}
							type="button"
							className={cn(
								"rounded-md px-3 py-1.5 text-xs font-medium transition-colors",
								maxUses === opt.value
									? "bg-primary text-primary-foreground"
									: "bg-raised text-t2 hover:bg-raised/80",
							)}
							onClick={() => setMaxUses(opt.value)}
						>
							{opt.label}
						</button>
					))}
				</div>
			</div>

			{/* Кнопка создания */}
			<Button
				onClick={() => void handleCreate()}
				disabled={isCreating}
				className="w-full"
			>
				{isCreating
					? "Создание..."
					: copiedToken
						? "Ссылка скопирована!"
						: "Создать ссылку"}
			</Button>
		</div>
	);
}

// ---------------------------------------------------------------------------
// Компонент списка существующих ссылок
// ---------------------------------------------------------------------------

/** Элемент списка пригласительных ссылок. */
function InviteItem({
	invite,
	channelId,
	onRevoked,
}: {
	invite: ChannelInvite;
	channelId: string;
	onRevoked: () => void;
}) {
	const [copied, setCopied] = useState(false);
	const [revoking, setRevoking] = useState(false);

	const handleCopy = useCallback(async () => {
		const url = getFullInviteUrl(invite.token);
		const ok = await copyToClipboard(url);
		if (ok) {
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		}
	}, [invite.token]);

	const handleRevoke = useCallback(async () => {
		setRevoking(true);
		try {
			await revokeInvite(channelId, invite.id);
			onRevoked();
		} finally {
			setRevoking(false);
		}
	}, [channelId, invite.id, onRevoked]);

	const isExpired =
		invite.expires_at != null && new Date(invite.expires_at) < new Date();
	const isExhausted =
		invite.max_uses > 0 && invite.use_count >= invite.max_uses;

	return (
		<div
			className={cn(
				"flex items-center justify-between rounded-md border px-3 py-2",
				(isExpired || isExhausted) && "opacity-50",
			)}
		>
			<div className="min-w-0 flex-1">
				<div className="flex items-center gap-2">
					<code className="text-xs font-mono text-t1 truncate">
						/invite/{invite.token}
					</code>
					{isExpired && (
						<span className="shrink-0 rounded bg-destructive/10 px-1.5 py-0.5 text-[10px] text-destructive">
							Истекла
						</span>
					)}
					{isExhausted && !isExpired && (
						<span className="shrink-0 rounded bg-orange-500/10 px-1.5 py-0.5 text-[10px] text-orange-600">
							Исчерпана
						</span>
					)}
				</div>
				<div className="mt-0.5 flex items-center gap-3 text-[11px] text-t2">
					<span>{formatUsage(invite)}</span>
					<span>{formatExpiration(invite.expires_at)}</span>
				</div>
			</div>

			<div className="flex items-center gap-1 shrink-0 ml-2">
				{/* Кнопка копирования */}
				<button
					type="button"
					className="inline-flex h-7 w-7 items-center justify-center rounded-md hover:bg-raised transition-colors"
					title="Копировать ссылку"
					aria-label="Копировать ссылку"
					onClick={() => void handleCopy()}
				>
					{copied ? (
						<svg
							aria-hidden="true"
							className="h-3.5 w-3.5 text-green-500"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M5 13l4 4L19 7"
							/>
						</svg>
					) : (
						<svg
							aria-hidden="true"
							className="h-3.5 w-3.5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
							/>
						</svg>
					)}
				</button>

				{/* Кнопка отзыва */}
				<button
					type="button"
					className="inline-flex h-7 w-7 items-center justify-center rounded-md hover:bg-destructive/10 text-t2 hover:text-destructive transition-colors"
					title="Отозвать"
					aria-label="Отозвать"
					disabled={revoking}
					onClick={() => void handleRevoke()}
				>
					<svg
						aria-hidden="true"
						className="h-3.5 w-3.5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
						/>
					</svg>
				</button>
			</div>
		</div>
	);
}

/** Список существующих пригласительных ссылок канала. */
function InviteList({ channelId }: { channelId: string }) {
	const queryClient = useQueryClient();

	const { data: invites, isLoading } = useQuery({
		queryKey: ["messenger", "channel-invites", channelId],
		queryFn: () => getChannelInvites(channelId),
	});

	const handleRefresh = useCallback(() => {
		void queryClient.invalidateQueries({
			queryKey: ["messenger", "channel-invites", channelId],
		});
	}, [queryClient, channelId]);

	if (isLoading) {
		return (
			<div className="space-y-2">
				{["invite-sk-1", "invite-sk-2"].map((sk) => (
					<div key={sk} className="h-14 rounded-md animate-shimmer" />
				))}
			</div>
		);
	}

	const activeInvites = invites?.filter((inv) => !inv.is_revoked) ?? [];

	if (activeInvites.length === 0) {
		return (
			<p className="text-xs text-t2 text-center py-4">Нет активных ссылок</p>
		);
	}

	return (
		<div className="space-y-2">
			{activeInvites.map((invite) => (
				<InviteItem
					key={invite.id}
					invite={invite}
					channelId={channelId}
					onRevoked={handleRefresh}
				/>
			))}
		</div>
	);
}

// ---------------------------------------------------------------------------
// Основной компонент диалога
// ---------------------------------------------------------------------------

/** Контент диалога пригласительных ссылок. */
function InviteLinkDialogContent({ channelId }: { channelId: string }) {
	const queryClient = useQueryClient();

	const handleCreated = useCallback(() => {
		void queryClient.invalidateQueries({
			queryKey: ["messenger", "channel-invites", channelId],
		});
	}, [queryClient, channelId]);

	return (
		<div className="space-y-6">
			{/* Форма создания */}
			<CreateInviteForm channelId={channelId} onCreated={handleCreated} />

			{/* Разделитель */}
			<div className="relative">
				<div className="absolute inset-0 flex items-center">
					<div className="w-full border-t" />
				</div>
				<div className="relative flex justify-center text-xs">
					<span className="bg-base px-2 text-t2">Активные ссылки</span>
				</div>
			</div>

			{/* Список существующих ссылок */}
			<InviteList channelId={channelId} />
		</div>
	);
}

/**
 * InviteLinkDialog — диалог управления пригласительными ссылками канала.
 * Позволяет создавать новые ссылки с настройками срока действия и лимита использований,
 * копировать и отзывать существующие.
 */
export function InviteLinkDialog({ channelId }: { channelId: string }) {
	const [open, setOpen] = useState(false);

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<button
					type="button"
					className="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-xs font-medium text-t2 hover:bg-raised hover:text-t1 transition-colors"
					title="Пригласить в канал"
					aria-label="Пригласить в канал"
				>
					<svg
						aria-hidden="true"
						className="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
						/>
					</svg>
					<span className="hidden sm:inline">Пригласить</span>
				</button>
			</DialogTrigger>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle>Пригласительные ссылки</DialogTitle>
				</DialogHeader>
				<InviteLinkDialogContent channelId={channelId} />
			</DialogContent>
		</Dialog>
	);
}
