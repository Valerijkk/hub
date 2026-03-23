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
				<h4 className="text-sm font-medium">Чеклист</h4>
				{total > 0 && (
					<span className="text-xs text-muted-foreground">
						{done}/{total} выполнено
					</span>
				)}
			</div>

			{/* Прогресс-бар */}
			{total > 0 && (
				<div className="h-1.5 w-full rounded-full bg-muted">
					<div
						className={cn(
							"h-1.5 rounded-full transition-all duration-300",
							pct === 100 ? "bg-green-500" : "bg-primary",
						)}
						style={{ width: `${pct}%` }}
					/>
				</div>
			)}

			{/* Список элементов */}
			<div className="space-y-1">
				{items.map((item) => (
					<div
						key={item.id}
						className="group flex items-center gap-2 rounded-md px-1 py-1 hover:bg-muted/50"
					>
						<input
							type="checkbox"
							checked={item.checked}
							onChange={() => handleToggle(item.id)}
							disabled={disabled}
							className="h-4 w-4 rounded border-border accent-primary cursor-pointer"
						/>
						<span
							className={cn(
								"flex-1 text-sm",
								item.checked && "line-through text-muted-foreground",
							)}
						>
							{item.text}
						</span>
						<button
							type="button"
							onClick={() => handleDelete(item.id)}
							disabled={disabled}
							className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity text-sm px-1"
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
