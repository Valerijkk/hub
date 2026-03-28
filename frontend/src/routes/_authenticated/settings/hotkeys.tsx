import { createFileRoute } from "@tanstack/react-router";

/** Описание горячей клавиши. */
interface Hotkey {
	keys: string;
	description: string;
}

const HOTKEYS: Hotkey[] = [
	{ keys: "Ctrl + K", description: "Глобальный поиск" },
	{ keys: "Ctrl + N", description: "Новая задача" },
	{ keys: "Escape", description: "Закрыть диалог" },
	{ keys: "Enter", description: "Отправить сообщение" },
	{ keys: "Shift + Enter", description: "Новая строка в сообщении" },
];

/**
 * Страница справки по горячим клавишам.
 */
function HotkeysPage() {
	return (
		<div className="space-y-0">
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">
					Горячие клавиши
				</h2>
				<p className="text-[13px] text-t2 mb-5">
					Быстрый доступ к основным функциям
				</p>
			</div>

			<div className="overflow-x-auto">
				<table className="w-full border-collapse">
					<thead>
						<tr>
							<th className="px-[14px] py-[10px] text-[11px] font-bold text-t4 text-left uppercase tracking-[0.05em] border-b border-b2">
								Клавиши
							</th>
							<th className="px-[14px] py-[10px] text-[11px] font-bold text-t4 text-left uppercase tracking-[0.05em] border-b border-b2">
								Действие
							</th>
						</tr>
					</thead>
					<tbody>
						{HOTKEYS.map((hotkey) => (
							<tr key={hotkey.keys} className="hover:bg-[var(--hover)]">
								<td className="px-[14px] py-[10px] border-b border-b1">
									<kbd className="inline-flex items-center gap-1 px-[10px] py-[4px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] font-mono font-semibold text-t1">
										{hotkey.keys}
									</kbd>
								</td>
								<td className="px-[14px] py-[10px] border-b border-b1 text-[13px] text-t2">
									{hotkey.description}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/settings/hotkeys")({
	component: HotkeysPage,
});
