import { useCallback, useEffect, useState } from "react";
import {
	type APIKeyInfo,
	createAPIKey,
	listAPIKeys,
	revokeAPIKey,
} from "@/api/apikeys";

/** Страница управления API-ключами в настройках профиля. */
export function APIKeysPage() {
	const [keys, setKeys] = useState<APIKeyInfo[]>([]);
	const [loading, setLoading] = useState(true);
	const [showCreate, setShowCreate] = useState(false);
	const [newKeyName, setNewKeyName] = useState("");
	const [createdKey, setCreatedKey] = useState<string | null>(null);
	const [copied, setCopied] = useState(false);

	const fetchKeys = useCallback(async () => {
		try {
			setLoading(true);
			const data = await listAPIKeys();
			setKeys(data ?? []);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchKeys();
	}, [fetchKeys]);

	const handleCreate = async () => {
		if (!newKeyName.trim()) return;
		const resp = await createAPIKey({ name: newKeyName.trim() });
		setCreatedKey(resp.key);
		setNewKeyName("");
		setShowCreate(false);
		fetchKeys();
	};

	const handleRevoke = async (id: string) => {
		if (!confirm("Отозвать этот API-ключ? Действие необратимо.")) return;
		await revokeAPIKey(id);
		fetchKeys();
	};

	const handleCopy = () => {
		if (createdKey) {
			navigator.clipboard.writeText(createdKey);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		}
	};

	return (
		<div className="space-y-6 p-6">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">API-ключи</h2>
				<button
					onClick={() => setShowCreate(true)}
					className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
				>
					Создать ключ
				</button>
			</div>

			{createdKey && (
				<div className="rounded-md border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-950">
					<p className="mb-2 text-sm font-medium text-green-800 dark:text-green-200">
						API-ключ создан. Скопируйте его сейчас — он больше не будет показан.
					</p>
					<div className="flex items-center gap-2">
						<code className="flex-1 rounded bg-white p-2 text-sm dark:bg-gray-900">
							{createdKey}
						</code>
						<button
							onClick={handleCopy}
							className="rounded-md bg-green-600 px-3 py-2 text-sm text-white hover:bg-green-700"
						>
							{copied ? "Скопировано" : "Копировать"}
						</button>
					</div>
					<button
						onClick={() => setCreatedKey(null)}
						className="mt-2 text-sm text-green-600 underline"
					>
						Закрыть
					</button>
				</div>
			)}

			{showCreate && (
				<div className="rounded-md border p-4 dark:border-gray-700">
					<h3 className="mb-3 font-medium">Новый API-ключ</h3>
					<input
						type="text"
						value={newKeyName}
						onChange={(e) => setNewKeyName(e.target.value)}
						placeholder="Имя ключа (например, CI Pipeline)"
						className="mb-3 w-full rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
					/>
					<div className="flex gap-2">
						<button
							onClick={handleCreate}
							className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
						>
							Создать
						</button>
						<button
							onClick={() => setShowCreate(false)}
							className="rounded-md border px-4 py-2 text-sm"
						>
							Отмена
						</button>
					</div>
				</div>
			)}

			{loading ? (
				<p className="text-gray-500">Загрузка...</p>
			) : keys.length === 0 ? (
				<p className="text-gray-500">Нет API-ключей</p>
			) : (
				<table className="w-full text-sm">
					<thead>
						<tr className="border-b text-left dark:border-gray-700">
							<th className="p-2">Имя</th>
							<th className="p-2">Префикс</th>
							<th className="p-2">Последнее использование</th>
							<th className="p-2">Создан</th>
							<th className="p-2">Действия</th>
						</tr>
					</thead>
					<tbody>
						{keys.map((key) => (
							<tr key={key.id} className="border-b dark:border-gray-700">
								<td className="p-2">{key.name}</td>
								<td className="p-2">
									<code>{key.prefix}...</code>
								</td>
								<td className="p-2">
									{key.last_used_at
										? new Date(key.last_used_at).toLocaleDateString("ru-RU")
										: "Не использовался"}
								</td>
								<td className="p-2">
									{new Date(key.created_at).toLocaleDateString("ru-RU")}
								</td>
								<td className="p-2">
									<button
										onClick={() => handleRevoke(key.id)}
										className="text-red-600 hover:underline"
									>
										Отозвать
									</button>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
		</div>
	);
}
