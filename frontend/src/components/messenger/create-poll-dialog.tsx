import { useCallback, useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { createPoll } from "@/api/messenger";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

/** Максимальное количество вариантов ответа. */
const MAX_OPTIONS = 10;
/** Минимальное количество вариантов ответа. */
const MIN_OPTIONS = 2;

/** Пропсы компонента CreatePollDialog. */
interface CreatePollDialogProps {
	/** ID канала для создания опроса. */
	channelId: string;
	/** Открыт ли диалог. */
	open: boolean;
	/** Callback изменения состояния открытия. */
	onOpenChange: (open: boolean) => void;
}

/**
 * CreatePollDialog — диалог создания опроса.
 * Форма: вопрос, 2-10 вариантов ответа, переключатели анонимного и множественного выбора.
 */
export function CreatePollDialog({
	channelId,
	open,
	onOpenChange,
}: CreatePollDialogProps) {
	const queryClient = useQueryClient();
	const [question, setQuestion] = useState("");
	const [options, setOptions] = useState(["", ""]);
	const [isAnonymous, setIsAnonymous] = useState(false);
	const [isMultiChoice, setIsMultiChoice] = useState(false);
	const [submitting, setSubmitting] = useState(false);

	/** Сбросить форму в начальное состояние. */
	const resetForm = useCallback(() => {
		setQuestion("");
		setOptions(["", ""]);
		setIsAnonymous(false);
		setIsMultiChoice(false);
	}, []);

	/** Обновить текст варианта по индексу. */
	const updateOption = useCallback((index: number, value: string) => {
		setOptions((prev) => prev.map((o, i) => (i === index ? value : o)));
	}, []);

	/** Добавить новый вариант. */
	const addOption = useCallback(() => {
		setOptions((prev) => {
			if (prev.length >= MAX_OPTIONS) return prev;
			return [...prev, ""];
		});
	}, []);

	/** Удалить вариант по индексу. */
	const removeOption = useCallback((index: number) => {
		setOptions((prev) => {
			if (prev.length <= MIN_OPTIONS) return prev;
			return prev.filter((_, i) => i !== index);
		});
	}, []);

	/** Отправить опрос. */
	const handleSubmit = useCallback(async () => {
		const trimmedQuestion = question.trim();
		if (!trimmedQuestion) {
			toast.error("Введите вопрос");
			return;
		}

		const validOptions = options
			.map((o) => o.trim())
			.filter((o) => o.length > 0);

		if (validOptions.length < MIN_OPTIONS) {
			toast.error("Необходимо минимум 2 варианта ответа");
			return;
		}

		setSubmitting(true);
		try {
			await createPoll(channelId, {
				question: trimmedQuestion,
				options: validOptions,
				is_anonymous: isAnonymous,
				is_multi_choice: isMultiChoice,
			});
			// Инвалидируем кэш сообщений, чтобы опрос появился в ленте.
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "messages", channelId],
			});
			resetForm();
			onOpenChange(false);
		} catch {
			toast.error("Не удалось создать опрос");
		} finally {
			setSubmitting(false);
		}
	}, [channelId, question, options, isAnonymous, isMultiChoice, queryClient, resetForm, onOpenChange]);

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle>Создать опрос</DialogTitle>
				</DialogHeader>

				<div className="space-y-4">
					{/* Вопрос */}
					<div className="space-y-1.5">
						<Label htmlFor="poll-question">Вопрос</Label>
						<Input
							id="poll-question"
							value={question}
							onChange={(e) => setQuestion(e.target.value)}
							placeholder="Введите вопрос..."
							maxLength={500}
							disabled={submitting}
						/>
					</div>

					{/* Варианты ответов */}
					<div className="space-y-1.5">
						<Label>Варианты ответа</Label>
						<div className="space-y-2">
							{options.map((option, index) => (
								<div key={index} className="flex items-center gap-2">
									<Input
										value={option}
										onChange={(e) => updateOption(index, e.target.value)}
										placeholder={`Вариант ${index + 1}`}
										maxLength={200}
										disabled={submitting}
									/>
									{options.length > MIN_OPTIONS && (
										<button
											type="button"
											onClick={() => removeOption(index)}
											disabled={submitting}
											className="shrink-0 h-8 w-8 flex items-center justify-center rounded-md text-t2 hover:text-destructive hover:bg-destructive/10 transition-colors"
											title="Удалить вариант"
										>
											<Trash2 className="h-4 w-4" />
										</button>
									)}
								</div>
							))}
						</div>
						{options.length < MAX_OPTIONS && (
							<button
								type="button"
								onClick={addOption}
								disabled={submitting}
								className="flex items-center gap-1.5 text-sm text-primary hover:text-primary/80 transition-colors mt-1"
							>
								<Plus className="h-4 w-4" />
								Добавить вариант
							</button>
						)}
					</div>

					{/* Переключатели */}
					<div className="space-y-2">
						<label className="flex items-center gap-2 cursor-pointer">
							<input
								type="checkbox"
								checked={isAnonymous}
								onChange={(e) => setIsAnonymous(e.target.checked)}
								disabled={submitting}
								className="rounded border-input"
							/>
							<span className="text-sm">Анонимное голосование</span>
						</label>
						<label className="flex items-center gap-2 cursor-pointer">
							<input
								type="checkbox"
								checked={isMultiChoice}
								onChange={(e) => setIsMultiChoice(e.target.checked)}
								disabled={submitting}
								className="rounded border-input"
							/>
							<span className="text-sm">Несколько вариантов ответа</span>
						</label>
					</div>

					{/* Кнопка создания */}
					<Button
						onClick={() => void handleSubmit()}
						disabled={submitting || !question.trim()}
						className="w-full"
					>
						{submitting ? "Создание..." : "Создать"}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
