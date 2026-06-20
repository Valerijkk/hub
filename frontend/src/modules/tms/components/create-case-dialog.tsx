/**
 * Диалог создания нового тест-кейса.
 */

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import { toastApiError } from "@/lib/format-api-error";
import {
	createTestCase,
	type SuiteResponse,
	type TestCaseResponse,
	type TestCaseType,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

/** Пропсы диалога создания тест-кейса. */
interface CreateCaseDialogProps {
	projectId: string;
	suites?: SuiteResponse[];
	open: boolean;
	onOpenChange: (open: boolean) => void;
	trigger: React.ReactNode;
}

/** Диалог создания нового тест-кейса с выбором типа, приоритета и папки. */
export function CreateCaseDialog({
	projectId,
	suites,
	open,
	onOpenChange,
	trigger,
}: CreateCaseDialogProps) {
	const queryClient = useQueryClient();
	const [newTitle, setNewTitle] = useState("");
	const [newType, setNewType] = useState<TestCaseType>("classic");
	const [newPriority, setNewPriority] = useState("medium");
	const [newSuiteId, setNewSuiteId] = useState("");
	const [newDescription, setNewDescription] = useState("");

	const createMutation = useMutation({
		mutationFn: () =>
			createTestCase(projectId, {
				title: newTitle,
				case_type: newType,
				priority: newPriority as TestCaseResponse["priority"],
				suite_id: newSuiteId || undefined,
				description: newDescription || undefined,
			}),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-cases", projectId],
			});
			onOpenChange(false);
			setNewTitle("");
			setNewDescription("");
			toast.success("Тест-кейс создан");
		},
		onError: (err) => toastApiError(err, "Ошибка создания тест-кейса"),
	});

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogTrigger asChild>{trigger}</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Новый тест-кейс</DialogTitle>
				</DialogHeader>
				<div className="space-y-4">
					<Input
						placeholder="Название тест-кейса"
						value={newTitle}
						onChange={(e) => setNewTitle(e.target.value)}
					/>
					<div className="grid grid-cols-2 gap-3">
						<Select
							value={newType}
							onValueChange={(v) => setNewType(v as TestCaseType)}
						>
							<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="classic">Classic</SelectItem>
								<SelectItem value="bdd">BDD</SelectItem>
							</SelectContent>
						</Select>
						<Select
							value={newPriority}
							onValueChange={(v) => setNewPriority(v)}
						>
							<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="critical">Критический</SelectItem>
								<SelectItem value="high">Высокий</SelectItem>
								<SelectItem value="medium">Средний</SelectItem>
								<SelectItem value="low">Низкий</SelectItem>
								<SelectItem value="trivial">Тривиальный</SelectItem>
							</SelectContent>
						</Select>
					</div>
					{suites && suites.length > 0 && (
						<Select
							value={newSuiteId || "none"}
							onValueChange={(v) => setNewSuiteId(v === "none" ? "" : v)}
						>
							<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">Без папки</SelectItem>
								{suites.map((s) => (
									<SelectItem key={s.id} value={s.id}>
										{s.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					)}
					<textarea
						className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
						placeholder="Описание (необязательно)"
						rows={3}
						value={newDescription}
						onChange={(e) => setNewDescription(e.target.value)}
					/>
					<Button
						className="w-full"
						disabled={!newTitle.trim() || createMutation.isPending}
						onClick={() => createMutation.mutate()}
					>
						{createMutation.isPending ? "Создание..." : "Создать"}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
