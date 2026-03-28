/**
 * Редактируемое описание задачи с Plate editor.
 */

import type { Value } from "@udecode/plate";
import { useState } from "react";
import { PlateEditor } from "@/components/tracker/plate-editor";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function IssueTitle({
	title,
	onSave,
}: {
	title: string;
	onSave: (newTitle: string) => void;
}) {
	const [isEditing, setIsEditing] = useState(false);
	const [editValue, setEditValue] = useState("");

	const save = () => {
		if (editValue.trim() && editValue !== title) {
			onSave(editValue.trim());
		}
		setIsEditing(false);
	};

	if (isEditing) {
		return (
			<div className="flex gap-2">
				<Input
					value={editValue}
					onChange={(e) => setEditValue(e.target.value)}
					onKeyDown={(e) => {
						if (e.key === "Enter") save();
						if (e.key === "Escape") setIsEditing(false);
					}}
					autoFocus
					className="text-[24px] font-extrabold tracking-[-0.03em] leading-[1.3] border-b-2 border-transparent focus:border-blue bg-transparent w-full text-t1 outline-none transition-[border-color] duration-250 mb-6"
				/>
				<Button size="sm" onClick={save}>
					Сохранить
				</Button>
			</div>
		);
	}

	return (
		<h1
			className="cursor-pointer text-[24px] font-extrabold tracking-[-0.03em] leading-[1.3] border-b-2 border-transparent hover:border-blue bg-transparent w-full text-t1 transition-[border-color] duration-250 mb-6"
			onClick={() => {
				setEditValue(title);
				setIsEditing(true);
			}}
		>
			{title}
		</h1>
	);
}

export function IssueDescription({
	descriptionJson,
	onSave,
}: {
	descriptionJson: unknown;
	onSave: (value: Value | undefined) => void;
}) {
	const [isEditing, setIsEditing] = useState(false);
	const [editValue, setEditValue] = useState<Value | undefined>();

	const save = () => {
		onSave(editValue);
		setIsEditing(false);
	};

	return (
		<div>
			<Label className="mb-2 block text-[13px] font-bold uppercase tracking-[0.06em] text-t3">Описание</Label>
			{isEditing ? (
				<div className="space-y-2">
					<PlateEditor
						value={editValue}
						onChange={setEditValue}
						placeholder="Описание задачи..."
					/>
					<div className="flex gap-2">
						<Button size="sm" onClick={save}>
							Сохранить
						</Button>
						<Button
							variant="ghost"
							size="sm"
							onClick={() => setIsEditing(false)}
						>
							Отмена
						</Button>
					</div>
				</div>
			) : (
				<div
					className="cursor-pointer rounded-md p-2 hover:bg-[var(--hover)] text-[14px] leading-[1.75] text-t2"
					onClick={() => {
						setEditValue(descriptionJson as Value | undefined);
						setIsEditing(true);
					}}
				>
					{descriptionJson ? (
						<PlateEditor
							value={descriptionJson as unknown as Value}
							readOnly
							className="border-none p-0"
						/>
					) : (
						<p className="text-[14px] text-t3">
							Нажмите, чтобы добавить описание
						</p>
					)}
				</div>
			)}
		</div>
	);
}
