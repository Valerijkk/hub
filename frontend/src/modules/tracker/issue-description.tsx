/**
 * Редактируемое описание задачи с Plate editor.
 */

import type { Value } from "@udecode/plate";
import { Sparkles } from "lucide-react";
import { useState } from "react";
import { PlateEditor } from "@/components/tracker/plate-editor";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { activateOnEnterOrSpace } from "@/lib/keyboard";
import { useAIStore } from "@/stores/ai-store";

/** Узкий type-guard Plate Value — без unsafe cast (P2-FE-08 / W10-A2-018). */
export function parsePlateValue(raw: unknown): Value | undefined {
	if (!Array.isArray(raw) || raw.length === 0) return undefined;
	const first = raw[0];
	if (!first || typeof first !== "object") return undefined;
	if (!("children" in first) || !Array.isArray(first.children)) return undefined;
	return raw as Value;
}

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
		<button
			type="button"
			className="cursor-pointer text-[24px] font-extrabold tracking-[-0.03em] leading-[1.3] border-b-2 border-transparent hover:border-blue bg-transparent w-full text-t1 transition-[border-color] duration-250 mb-6 text-left"
			onClick={() => {
				setEditValue(title);
				setIsEditing(true);
			}}
			onKeyDown={activateOnEnterOrSpace(() => {
				setEditValue(title);
				setIsEditing(true);
			})}
		>
			{title}
		</button>
	);
}

export function IssueDescription({
	descriptionJson,
	descriptionText,
	issueKey,
	issueTitle,
	onSave,
}: {
	descriptionJson: unknown;
	descriptionText?: string;
	issueKey?: string;
	issueTitle?: string;
	onSave: (value: Value | undefined) => void;
}) {
	const [isEditing, setIsEditing] = useState(false);
	const [editValue, setEditValue] = useState<Value | undefined>();
	const openWithPrefill = useAIStore((s) => s.openWithPrefill);
	const loadModels = useAIStore((s) => s.loadModels);

	const askAI = () => {
		void loadModels();
		const key = issueKey ?? "задача";
		const title = issueTitle ?? "";
		const body = (descriptionText ?? "").trim() || "(описание пустое)";
		openWithPrefill(
			`Помоги улучшить описание задачи ${key}${title ? ` «${title}»` : ""}:\n\n${body}\n\nПредложи улучшенную версию в Markdown.`,
		);
	};

	const save = () => {
		onSave(editValue);
		setIsEditing(false);
	};

	const plateValue = parsePlateValue(descriptionJson);

	return (
		<div>
			<div className="mb-2 flex items-center justify-between gap-2">
				<Label className="text-[13px] font-bold uppercase tracking-[0.06em] text-t3">
					Описание
				</Label>
				<Button
					type="button"
					variant="outline"
					size="sm"
					className="h-7 gap-1 text-xs"
					onClick={askAI}
				>
					<Sparkles className="h-3.5 w-3.5 text-purple-500" />
					Спросить AI
				</Button>
			</div>
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
				<button
					type="button"
					className="cursor-pointer rounded-md p-2 hover:bg-[var(--hover)] text-[14px] leading-[1.75] text-t2 w-full text-left"
					onClick={() => {
						setEditValue(plateValue);
						setIsEditing(true);
					}}
					onKeyDown={activateOnEnterOrSpace(() => {
						setEditValue(plateValue);
						setIsEditing(true);
					})}
				>
					{plateValue ? (
						<PlateEditor
							value={plateValue}
							readOnly
							className="border-none p-0"
						/>
					) : (
						<p className="text-[14px] text-t3">
							Нажмите, чтобы добавить описание
						</p>
					)}
				</button>
			)}
		</div>
	);
}
