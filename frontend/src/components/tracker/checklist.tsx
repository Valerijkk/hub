/**
 * Компонент чеклиста задачи.
 * Позволяет добавлять, удалять и переключать элементы чеклиста.
 * Показывает индикатор прогресса выполнения.
 */
import { useState } from "react";
import type { ChecklistItem } from "@/api/tracker";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

/** Пропсы компонента Checklist. */
interface ChecklistProps {
	/** Текущие элементы чеклиста. */
	items: ChecklistItem[];
	/** Колбэк при изменении чеклиста (добавление, удаление, переключение). */
	onChange: (items: ChecklistItem[]) => void;
	/** Блокировка редактирования (например, при сохранении). */
	disabled?: boolean;
}

/** Генерирует уникальный ID для элемента чеклиста. */
function generateId(): string {
	return `cl_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

/**
 * Интерактивный чеклист с добавлением, удалением и переключением элементов.
 * Отображает прогресс: "N/M выполнено" с прогресс-баром.
 */
export function Checklist({ items, onChange, disabled }: ChecklistProps) {
	const [newItemText, setNewItemText] = useState("");

	const total = items.length;
	const done = items.filter((i) => i.checked).length;
	const pct = total > 0 ? Math.round((done / total) * 100) : 0;

	/** Добавить новый элемент в конец списка. */
	const handleAdd = () => {
		const text = newItemText.trim();
		if (!text) return;
		onChange([...items, { id: generateId(), text, checked: false }]);
		setNewItemText("");
	};

	/** Обработка нажатия Enter в поле ввода. */
	const handleKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === "Enter") {
			e.preventDefault();
			handleAdd();
		}
	};

	/** Переключить состояние элемента. */
	const handleToggle = (id: string) => {
		onChange(
			items.map((item) =>
				item.id === id ? { ...item, checked: !item.checked } : item,
			),
		);
	};

	/** Удалить элемент из списка. */
	const handleDelete = (id: string) => {
		onChange(items.filter((item) => item.id !== id));
	};

	return (
		<div className="space-y-3">
			{/* Заголовок с прогрессом */}
			<div className="flex items-center justify-between">
				<h4 className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">Чеклист</h4>
				{total > 0 && (
					<div className="flex items-center gap-[10px]">
						<div className="h-1 bg-white/[0.04] rounded-sm overflow-hidden max-w-[200px] w-[200px]">
							<div
								className="h-full bg-green rounded-sm shadow-[0_0_8px_var(--gG)] transition-all duration-300"
								style={{ width: `${pct}%` }}
							/>
						</div>
						<span className="text-[12px] text-t3 font-semibold">
							{done}/{total}
						</span>
					</div>
				)}
			</div>

			{/* Список элементов */}
			<div>
				{items.map((item) => (
					<div
						key={item.id}
						className="group flex items-center gap-3 py-[10px] border-b border-b1 text-[14px]"
					>
						<div
							onClick={() => !disabled && handleToggle(item.id)}
							className={cn(
								"w-5 h-5 border-2 border-b3 rounded-[5px] shrink-0 flex items-center justify-center cursor-pointer transition-all duration-200 hover:border-green",
								item.checked && "bg-green border-green shadow-[0_0_8px_var(--gG)]",
							)}
						>
							{item.checked && <span className="text-white text-[12px] font-extrabold">{"\u2713"}</span>}
						</div>
						<span
							className={cn(
								"flex-1 text-[14px]",
								item.checked && "line-through text-t3",
							)}
						>
							{item.text}
						</span>
						<button
							type="button"
							onClick={() => handleDelete(item.id)}
							disabled={disabled}
							className="opacity-0 group-hover:opacity-100 text-t2 hover:text-destructive transition-opacity text-sm px-1"
							title="Удалить"
						>
							&times;
						</button>
					</div>
				))}
			</div>

			{/* Добавление нового элемента */}
			<div className="flex items-center gap-2">
				<Input
					value={newItemText}
					onChange={(e) => setNewItemText(e.target.value)}
					onKeyDown={handleKeyDown}
					placeholder="Добавить пункт..."
					disabled={disabled}
					className="text-sm"
				/>
				<Button
					type="button"
					variant="outline"
					size="sm"
					onClick={handleAdd}
					disabled={disabled || !newItemText.trim()}
				>
					Добавить
				</Button>
			</div>
		</div>
	);
}
