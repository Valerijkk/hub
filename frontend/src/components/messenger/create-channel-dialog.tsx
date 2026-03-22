import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import type { ChannelType } from "@/api/messenger";
import { createChannel } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

/** Диалог создания канала (вариант с кнопкой в заголовке). */
export function CreateChannelDialog() {
	const [open, setOpen] = useState(false);

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<button
					type="button"
					className="inline-flex h-7 w-7 items-center justify-center rounded-md text-sidebar-foreground/60 hover:bg-white/10 hover:text-sidebar-foreground transition-colors"
					title="Создать канал"
				>
					<svg
						className="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M12 4v16m8-8H4"
						/>
					</svg>
				</button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Создать канал</DialogTitle>
				</DialogHeader>
				<CreateChannelDialogContent />
			</DialogContent>
		</Dialog>
	);
}

/** Содержимое диалога создания канала (для кнопки нижней панели). */
export function CreateChannelDialogContent() {
	const [name, setName] = useState("");
	const [type, setType] = useState<ChannelType>("group");
	const [isLoading, setIsLoading] = useState(false);
	const queryClient = useQueryClient();

	const handleCreate = async () => {
		if (!name.trim()) return;
		setIsLoading(true);
		try {
			await createChannel({ name: name.trim(), channel_type: type });
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "channels"],
			});
			setName("");
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<div className="space-y-4 pt-2">
			<Input
				placeholder="Название канала"
				value={name}
				onChange={(e) => setName(e.target.value)}
			/>
			<div className="flex gap-2">
				{(["group", "public"] as const).map((t) => (
					<Button
						key={t}
						variant={type === t ? "default" : "outline"}
						size="sm"
						onClick={() => setType(t)}
					>
						{t === "group" ? "Групповой" : "Публичный"}
					</Button>
				))}
			</div>
			<Button
				className="w-full"
				onClick={handleCreate}
				disabled={!name.trim() || isLoading}
			>
				{isLoading ? "Создание..." : "Создать"}
			</Button>
		</div>
	);
}
