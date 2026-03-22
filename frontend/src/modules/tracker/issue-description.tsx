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
					className="text-xl font-bold"
				/>
				<Button size="sm" onClick={save}>
					Сохранить
				</Button>
			</div>
		);
	}

	return (
		<h1
			className="cursor-pointer text-xl font-bold hover:bg-muted/50"
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
			<Label className="mb-2 block text-sm font-medium">Описание</Label>
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
					className="cursor-pointer rounded-md p-2 hover:bg-muted/50"
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
						<p className="text-sm text-muted-foreground">
							Нажмите, чтобы добавить описание
						</p>
					)}
				</div>
			)}
		</div>
	);
}
