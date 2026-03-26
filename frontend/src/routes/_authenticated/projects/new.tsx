import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { toast } from "sonner";
import { createProject } from "@/api/projects";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

/**
 * Страница создания нового проекта — центрированная карточка с формой.
 */
function generateKey(name: string): string {
	return name
		.toUpperCase()
		.replace(/[^A-Z0-9]/g, "")
		.slice(0, 10);
}

function NewProjectPage() {
	const navigate = useNavigate();
	const queryClient = useQueryClient();

	const [name, setName] = useState("");
	const [key, setKey] = useState("");
	const [description, setDescription] = useState("");
	const [keyManuallyEdited, setKeyManuallyEdited] = useState(false);

	const mutation = useMutation({
		mutationFn: createProject,
		onSuccess: (project) => {
			void queryClient.invalidateQueries({ queryKey: ["projects"] });
			toast.success("Проект создан");
			void navigate({
				to: "/projects/$projectId",
				params: { projectId: project.id },
			});
		},
		onError: (err: unknown) => {
			const message =
				err && typeof err === "object" && "message" in err
					? (err as { message: string }).message
					: "Не удалось создать проект";
			toast.error(message);
		},
	});

	function handleNameChange(value: string) {
		setName(value);
		if (!keyManuallyEdited) {
			setKey(generateKey(value));
		}
	}

	function handleKeyChange(value: string) {
		setKeyManuallyEdited(true);
		setKey(
			value
				.toUpperCase()
				.replace(/[^A-Z0-9]/g, "")
				.slice(0, 10),
		);
	}

	function handleSubmit(e: FormEvent) {
		e.preventDefault();
		if (!name.trim() || key.length < 2) {
			toast.error("Заполните название и ключ (2-10 символов)");
			return;
		}
		mutation.mutate({
			name: name.trim(),
			key,
			description: description.trim() || undefined,
		});
	}

	return (
		<div className="mx-auto max-w-lg animate-fade-in">
			{/* Навигация назад */}
			<button
				onClick={() => void navigate({ to: "/projects" })}
				className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
			>
				<svg
					className="h-4 w-4"
					viewBox="0 0 16 16"
					fill="none"
					stroke="currentColor"
					strokeWidth="2"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M10 12L6 8l4-4" />
				</svg>
				Назад к проектам
			</button>

			<Card className="rounded-xl border shadow-sm">
				<CardHeader className="pb-4">
					<CardTitle className="text-xl">Новый проект</CardTitle>
					<CardDescription>Создайте проект для вашей команды</CardDescription>
				</CardHeader>
				<CardContent>
					<form onSubmit={handleSubmit} className="space-y-5">
						<div className="space-y-2">
							<Label htmlFor="proj-name" className="text-sm font-medium">
								Название
							</Label>
							<Input
								id="proj-name"
								placeholder="Мой проект"
								value={name}
								onChange={(e) => handleNameChange(e.target.value)}
								autoFocus
								className="h-10"
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="proj-key" className="text-sm font-medium">
								Ключ проекта
							</Label>
							<Input
								id="proj-key"
								placeholder="PROJ"
								value={key}
								onChange={(e) => handleKeyChange(e.target.value)}
								maxLength={10}
								className="h-10 font-mono uppercase"
							/>
							<p className="text-xs text-muted-foreground">
								2-10 символов, A-Z и цифры. Используется в нумерации задач (
								<span className="font-mono">{key || "KEY"}-1</span>)
							</p>
						</div>
						<div className="space-y-2">
							<Label htmlFor="proj-desc" className="text-sm font-medium">
								Описание
								<span className="ml-1 text-muted-foreground font-normal">
									(необязательно)
								</span>
							</Label>
							<Textarea
								id="proj-desc"
								placeholder="О чем этот проект?"
								value={description}
								onChange={(e) => setDescription(e.target.value)}
								rows={3}
								className="resize-none"
							/>
						</div>

						{/* Подсказка */}
						<div className="rounded-lg border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-200">
							<p className="font-medium mb-1">После создания проекта вы сможете:</p>
							<ul className="list-disc pl-4 text-xs space-y-0.5">
								<li>Подключить <strong>GitHub / GitLab</strong> репозиторий в Настройки &rarr; Интеграции</li>
								<li>Настроить <strong>Webhook</strong> для автоматической привязки коммитов и PR/MR к задачам</li>
								<li>Подключить <strong>CI/CD</strong> для импорта результатов автотестов</li>
								<li>Добавить участников и настроить <strong>роли и права</strong></li>
							</ul>
						</div>

						{/* Разделитель */}
						<div className="border-t pt-5">
							<div className="flex gap-3 justify-end">
								<Button
									type="button"
									variant="outline"
									onClick={() => void navigate({ to: "/projects" })}
								>
									Отмена
								</Button>
								<Button
									type="submit"
									disabled={
										mutation.isPending || !name.trim() || key.length < 2
									}
									className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
								>
									{mutation.isPending ? (
										<span className="flex items-center gap-2">
											<svg
												className="h-4 w-4 animate-spin"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												strokeWidth="2"
											>
												<path d="M12 2v4m0 12v4m-7.07-3.93l2.83-2.83m8.49-8.49l2.83-2.83M2 12h4m12 0h4m-3.93 7.07l-2.83-2.83M7.76 7.76L4.93 4.93" />
											</svg>
											Создание...
										</span>
									) : (
										"Создать проект"
									)}
								</Button>
							</div>
						</div>
					</form>
				</CardContent>
			</Card>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/projects/new")({
	component: NewProjectPage,
});
