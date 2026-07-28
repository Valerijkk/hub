import { Plus, Trash2 } from "lucide-react";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import { createPoll } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useInvalidateMessages } from "@/hooks/use-invalidate";
import i18n from "@/i18n";
import { toastApiError } from "@/lib/format-api-error";

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

/** Вариант ответа со стабильным идентификатором для React-ключей. */
interface PollOption {
	id: string;
	text: string;
}

/** Стабильный id варианта опроса (UUID с fallback без crypto). */
function makeOptionId(): string {
	if (
		typeof crypto !== "undefined" &&
		typeof crypto.randomUUID === "function"
	) {
		return crypto.randomUUID();
	}
	return `opt-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

/** Создать массив пустых вариантов с уникальными id. */
function makeInitialOptions(): PollOption[] {
	return [
		{ id: makeOptionId(), text: "" },
		{ id: makeOptionId(), text: "" },
	];
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
	const invalidateMessages = useInvalidateMessages(channelId);
	const [question, setQuestion] = useState("");
	const [options, setOptions] = useState<PollOption[]>(() =>
		makeInitialOptions(),
	);
	const [isAnonymous, setIsAnonymous] = useState(false);
	const [isMultiChoice, setIsMultiChoice] = useState(false);
	const [submitting, setSubmitting] = useState(false);

	/** Сбросить форму в начальное состояние. */
	const resetForm = useCallback(() => {
		setQuestion("");
		setOptions(makeInitialOptions());
		setIsAnonymous(false);
		setIsMultiChoice(false);
	}, []);

	/** Обновить текст варианта по id. */
	const updateOption = useCallback((id: string, value: string) => {
		setOptions((prev) =>
			prev.map((o) => (o.id === id ? { ...o, text: value } : o)),
		);
	}, []);

	/** Добавить новый вариант. */
	const addOption = useCallback(() => {
		setOptions((prev) => {
			if (prev.length >= MAX_OPTIONS) return prev;
			return [...prev, { id: makeOptionId(), text: "" }];
		});
	}, []);

	/** Удалить вариант по id. */
	const removeOption = useCallback((id: string) => {
		setOptions((prev) => {
			if (prev.length <= MIN_OPTIONS) return prev;
			return prev.filter((o) => o.id !== id);
		});
	}, []);

	/** Отправить опрос. */
	const handleSubmit = useCallback(async () => {
		const trimmedQuestion = question.trim();
		if (!trimmedQuestion) {
			toast.error(i18n.t("common:validation.enterQuestion"));
			return;
		}

		const validOptions = options
			.map((o) => o.text.trim())
			.filter((o) => o.length > 0);

		if (validOptions.length < MIN_OPTIONS) {
			toast.error(i18n.t("common:validation.pollMinOptions"));
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
			invalidateMessages();
			resetForm();
			onOpenChange(false);
		} catch (err) {
			toastApiError(err, "Не удалось создать опрос");
		} finally {
			setSubmitting(false);
		}
	}, [
		channelId,
		question,
		options,
		isAnonymous,
		isMultiChoice,
		invalidateMessages,
		resetForm,
		onOpenChange,
	]);

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
								// Стабильный id сохраняет focus при добавлении/удалении вариантов опроса.
								<div key={option.id} className="flex items-center gap-2">
									<Input
										value={option.text}
										onChange={(e) => updateOption(option.id, e.target.value)}
										placeholder={`Вариант ${index + 1}`}
										maxLength={200}
										disabled={submitting}
									/>
									{options.length > MIN_OPTIONS && (
										<button
											type="button"
											onClick={() => removeOption(option.id)}
											disabled={submitting}
											className="shrink-0 h-8 w-8 flex items-center justify-center rounded-md text-t2 hover:text-destructive hover:bg-destructive/10 transition-colors"
											title="Удалить вариант"
											aria-label="Удалить вариант"
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
