import { useCallback, useEffect, useRef, useState } from "react";
import {
	type APIKeyInfo,
	createAPIKey,
	listAPIKeys,
	revokeAPIKey,
} from "@/api/apikeys";
import { QueryErrorAlert } from "@/components/query-state-view";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { toastApiError } from "@/lib/format-api-error";

/** Страница управления API-ключами в настройках профиля. */
export function APIKeysPage() {
	const [keys, setKeys] = useState<APIKeyInfo[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);
	const [showCreate, setShowCreate] = useState(false);
	const [newKeyName, setNewKeyName] = useState("");
	const [createdKey, setCreatedKey] = useState<string | null>(null);
	const [copied, setCopied] = useState(false);
	const [revokeDialogId, setRevokeDialogId] = useState<string | null>(null);
	const copyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	useEffect(
		() => () => {
			if (copyTimerRef.current) clearTimeout(copyTimerRef.current);
		},
		[],
	);

	const fetchKeys = useCallback(async () => {
		try {
			setLoading(true);
			const data = await listAPIKeys();
			setKeys(data ?? []);
			setLoadError(false);
		} catch (err) {
			setLoadError(true);
			setKeys([]);
			toastApiError(err, "Не удалось загрузить API-ключи");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void fetchKeys();
	}, [fetchKeys]);

	const handleCreate = async () => {
		if (!newKeyName.trim()) return;
		try {
			const resp = await createAPIKey({ name: newKeyName.trim() });
			setCreatedKey(resp.key);
			setNewKeyName("");
			setShowCreate(false);
			void fetchKeys();
		} catch (err) {
			toastApiError(err, "Не удалось создать API-ключ");
		}
	};

	const handleRevoke = async (id: string) => {
		try {
			await revokeAPIKey(id);
			void fetchKeys();
		} catch (err) {
			toastApiError(err, "Не удалось отозвать API-ключ");
		}
	};

	const handleCopy = () => {
		if (createdKey) {
			void navigator.clipboard.writeText(createdKey);
			setCopied(true);
			if (copyTimerRef.current) clearTimeout(copyTimerRef.current);
			copyTimerRef.current = setTimeout(() => setCopied(false), 2000);
		}
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em]">
					API-ключи
				</h2>
				<button
					type="button"
					onClick={() => setShowCreate(true)}
					className="settings-btn-save inline-flex items-center gap-[7px] px-5 py-[9px] rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_28px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
				>
					Создать ключ
				</button>
			</div>

			{createdKey && (
				<div className="rounded-[var(--r3)] border border-green/[0.15] bg-green/[0.06] p-4">
					<p className="mb-2 text-[13px] font-semibold text-green">
						API-ключ создан. Скопируйте его сейчас — он больше не будет показан.
					</p>
					<div className="flex items-center gap-2">
						<code className="flex-1 rounded-[var(--r2)] bg-[var(--input)] p-2 text-[13px] font-mono text-t1">
							{createdKey}
						</code>
						<button
							type="button"
							onClick={handleCopy}
							className="rounded-[var(--r2)] bg-green px-3 py-2 text-[13px] font-bold text-white hover:bg-[var(--greenS)]"
						>
							{copied ? "Скопировано" : "Копировать"}
						</button>
					</div>
					<button
						type="button"
						onClick={() => setCreatedKey(null)}
						className="mt-2 text-[12px] text-green underline"
					>
						Закрыть
					</button>
				</div>
			)}

			{showCreate && (
				<div className="rounded-[var(--r3)] border border-b1 p-4 bg-[var(--card)]">
					<h3 className="mb-3 text-[14px] font-semibold">Новый API-ключ</h3>
					<input
						type="text"
						value={newKeyName}
						onChange={(e) => setNewKeyName(e.target.value)}
						placeholder="Имя ключа (например, CI Pipeline)"
						className="mb-3 w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none focus:border-[rgba(91,156,245,0.3)] placeholder:text-t3"
					/>
					<div className="flex gap-2">
						<button
							type="button"
							onClick={handleCreate}
							className="settings-btn-save inline-flex items-center px-4 py-2 rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)]"
						>
							Создать
						</button>
						<button
							type="button"
							onClick={() => setShowCreate(false)}
							className="rounded-[var(--r2)] border border-b1 px-4 py-2 text-[13px] text-t2 hover:text-t1 hover:border-b2 transition-colors duration-180"
						>
							Отмена
						</button>
					</div>
				</div>
			)}

			{loading ? (
				<p className="text-t3 text-[13px]">Загрузка...</p>
			) : loadError ? (
				<QueryErrorAlert
					message="Не удалось загрузить API-ключи"
					retry={() => void fetchKeys()}
				/>
			) : keys.length === 0 ? (
				<p className="text-t3 text-[13px]">Нет API-ключей</p>
			) : (
				<table className="w-full border-collapse">
					<thead>
						<tr>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
								Имя
							</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
								Префикс
							</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
								Использование
							</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
								Создан
							</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
								Действия
							</th>
						</tr>
					</thead>
					<tbody>
						{keys.map((key) => (
							<tr key={key.id} className="hover:bg-[var(--hover)]">
								<td className="px-4 py-[14px] border-b border-b1 text-[13px]">
									{key.name}
								</td>
								<td className="px-4 py-[14px] border-b border-b1 text-[13px]">
									<code className="font-mono text-[12px] text-t4">
										{key.prefix}...
									</code>
								</td>
								<td className="px-4 py-[14px] border-b border-b1 text-[12px] text-t3">
									{key.last_used_at
										? new Date(key.last_used_at).toLocaleDateString("ru-RU")
										: "Не использовался"}
								</td>
								<td className="px-4 py-[14px] border-b border-b1 text-[12px] text-t3">
									{new Date(key.created_at).toLocaleDateString("ru-RU")}
								</td>
								<td className="px-4 py-[14px] border-b border-b1">
									<button
										type="button"
										onClick={() => setRevokeDialogId(key.id)}
										className="text-[12px] font-bold text-red hover:underline"
									>
										Отозвать
									</button>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}

			<ConfirmDialog
				open={revokeDialogId !== null}
				onOpenChange={(open) => {
					if (!open) setRevokeDialogId(null);
				}}
				title="Отозвать API-ключ?"
				description="Действие необратимо. Все интеграции, использующие этот ключ, перестанут работать."
				confirmLabel="Отозвать"
				destructive
				onConfirm={async () => {
					if (revokeDialogId) {
						await handleRevoke(revokeDialogId);
					}
				}}
			/>
		</div>
	);
}
