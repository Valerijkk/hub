import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// JQL-поля, операторы и значения
// ---------------------------------------------------------------------------

/** Описание JQL-поля: имя и допустимые операторы. */
interface JQLFieldDef {
	name: string;
	label: string;
	operators: string[];
	values?: string[];
}

/** Список поддерживаемых полей JQL с допустимыми операторами и значениями. */
const JQL_FIELDS: JQLFieldDef[] = [
	{
		name: "status",
		label: "Статус",
		operators: ["=", "!=", "IN", "NOT IN"],
		values: ["To Do", "In Progress", "Done", "Backlog", "Review"],
	},
	{
		name: "priority",
		label: "Приоритет",
		operators: ["=", "!=", "IN", "NOT IN"],
		values: ["low", "medium", "high", "critical"],
	},
	{
		name: "assignee",
		label: "Исполнитель",
		operators: ["=", "!="],
		values: ["currentUser()"],
	},
	{
		name: "type",
		label: "Тип",
		operators: ["=", "!=", "IN", "NOT IN"],
		values: ["task", "bug", "story", "epic"],
	},
	{
		name: "created",
		label: "Дата создания",
		operators: ["=", ">", "<", ">=", "<="],
	},
	{
		name: "updated",
		label: "Дата обновления",
		operators: ["=", ">", "<", ">=", "<="],
	},
];

/** Ключевые слова JQL для соединения условий и сортировки. */
const JQL_KEYWORDS = ["AND", "OR", "ORDER BY"];

/** Определяет контекст текущей позиции ввода в JQL-строке. */
type SuggestionContext =
	| { type: "field" }
	| { type: "operator"; fieldName: string }
	| { type: "value"; fieldName: string }
	| { type: "keyword" }
	| { type: "none" };

function getJqlContext(query: string): SuggestionContext {
	const trimmed = query.trimEnd();
	if (!trimmed) return { type: "field" };

	// Разбиваем на токены с учётом кавычек
	const tokens = tokenize(trimmed);
	if (tokens.length === 0) return { type: "field" };

	const lastToken = tokens[tokens.length - 1];
	if (!lastToken) return { type: "field" };
	const prevTokens = tokens.slice(0, -1);

	// Проверяем, закончился ли ввод пробелом (следующий токен ожидается)
	const endsWithSpace = query.endsWith(" ");

	if (endsWithSpace) {
		// Смотрим на последний полный токен для определения контекста
		const upperLast = lastToken.toUpperCase();

		// После AND/OR — ожидаем поле
		if (upperLast === "AND" || upperLast === "OR") {
			return { type: "field" };
		}

		// Если последний токен — оператор, ожидаем значение
		if (isOperator(lastToken)) {
			const fieldName = findFieldBeforeOperator(tokens, tokens.length - 1);
			if (fieldName) return { type: "value", fieldName };
		}

		// Если последний — значение или кавычка закрылась, ожидаем ключевое слово
		if (isFieldName(lastToken)) {
			return { type: "operator", fieldName: lastToken };
		}

		// После значения ожидаем AND/OR/ORDER BY
		return { type: "keyword" };
	}

	// Вводим частичный токен
	const upperLast = lastToken.toUpperCase();

	// Если это начало или после ключевого слова — подсказываем поля
	if (prevTokens.length === 0) {
		return { type: "field" };
	}

	const prevToken = prevTokens[prevTokens.length - 1];
	if (!prevToken) return { type: "field" };
	const upperPrev = prevToken.toUpperCase();

	if (upperPrev === "AND" || upperPrev === "OR") {
		return { type: "field" };
	}

	if (isFieldName(prevToken)) {
		return { type: "operator", fieldName: prevToken };
	}

	if (isOperator(prevToken)) {
		const fieldName = findFieldBeforeOperator(
			prevTokens,
			prevTokens.length - 1,
		);
		if (fieldName) return { type: "value", fieldName };
	}

	// Частичное ключевое слово
	if (
		"AND".startsWith(upperLast) ||
		"OR".startsWith(upperLast) ||
		"ORDER".startsWith(upperLast)
	) {
		return { type: "keyword" };
	}

	return { type: "field" };
}

function tokenize(input: string): string[] {
	const tokens: string[] = [];
	let pos = 0;
	while (pos < input.length) {
		// Пропускаем пробелы
		while (pos < input.length && input[pos] === " ") pos++;
		if (pos >= input.length) break;

		// Строка в кавычках
		const currentChar = input[pos];
		if (currentChar === '"' || currentChar === "'") {
			const quote = currentChar;
			const start = pos;
			pos++;
			while (pos < input.length && input[pos] !== quote) pos++;
			if (pos < input.length) pos++;
			tokens.push(input.slice(start, pos));
			continue;
		}

		// Специальные операторы
		if (pos + 1 < input.length) {
			const twoChar = input.slice(pos, pos + 2);
			if (twoChar === ">=" || twoChar === "<=" || twoChar === "!=") {
				tokens.push(twoChar);
				pos += 2;
				continue;
			}
		}
		const opChar = input[pos];
		if (opChar === "=" || opChar === ">" || opChar === "<") {
			tokens.push(opChar);
			pos++;
			continue;
		}

		// Обычное слово
		const start = pos;
		while (
			pos < input.length &&
			input[pos] !== " " &&
			input[pos] !== "=" &&
			input[pos] !== ">" &&
			input[pos] !== "<" &&
			input[pos] !== "!"
		) {
			pos++;
		}
		if (pos > start) {
			tokens.push(input.slice(start, pos));
		}
	}
	return tokens;
}

function isOperator(token: string): boolean {
	const ops = ["=", "!=", ">", "<", ">=", "<=", "IN", "NOT", "NOT IN"];
	return ops.includes(token.toUpperCase());
}

function isFieldName(token: string): boolean {
	return JQL_FIELDS.some((f) => f.name === token.toLowerCase());
}

function findFieldBeforeOperator(
	tokens: string[],
	opIdx: number,
): string | undefined {
	for (let i = opIdx - 1; i >= 0; i--) {
		const t = tokens[i];
		if (!t) break;
		if (isFieldName(t)) return t.toLowerCase();
		if (t.toUpperCase() === "NOT") continue; // NOT IN — составной оператор JQL — составной оператор JQL
		break;
	}
	return undefined;
}

/** Получить неполный токен, который пользователь печатает. */
function getPartialTokenFor(query: string): string {
	if (query.endsWith(" ")) return "";
	const tokens = tokenize(query);
	return tokens[tokens.length - 1] ?? "";
}

// ---------------------------------------------------------------------------
// Пропсы компонента JQLEditor
// ---------------------------------------------------------------------------

interface JQLEditorProps {
	/** Текущее значение запроса. */
	value: string;
	/** Callback при изменении JQL. */
	onChange: (value: string) => void;
	/** Callback по Enter (выполнить JQL). */
	onSubmit?: (query: string) => void;
	/** Placeholder поля ввода. */
	placeholder?: string;
	/** Дополнительные CSS-классы для корневого элемента. */
	className?: string;
}

/**
 * JQL-редактор с автодополнением.
 * Подсказывает имена полей, операторы и значения при вводе JQL-запроса.
 */
export function JQLEditor({
	value,
	onChange,
	onSubmit,
	placeholder = 'JQL-запрос, например: status = "In Progress" AND priority = high',
	className,
}: JQLEditorProps) {
	const [showSuggestions, setShowSuggestions] = useState(false);
	const [selectedIndex, setSelectedIndex] = useState(0);
	const inputRef = useRef<HTMLInputElement>(null);
	const containerRef = useRef<HTMLDivElement>(null);

	/** Вычисляем подсказки на основе текущего ввода. */
	const suggestions = useMemo(() => {
		const ctx = getJqlContext(value);
		const partialToken = getPartialTokenFor(value);
		const lower = partialToken.toLowerCase();

		switch (ctx.type) {
			case "field":
				return JQL_FIELDS.filter((f) => !lower || f.name.startsWith(lower)).map(
					(f) => ({ label: f.name, description: f.label }),
				);

			case "operator": {
				const field = JQL_FIELDS.find((f) => f.name === ctx.fieldName);
				if (!field) return [];
				return field.operators
					.filter((op) => !lower || op.toLowerCase().startsWith(lower))
					.map((op) => ({ label: op, description: "" }));
			}

			case "value": {
				const field = JQL_FIELDS.find((f) => f.name === ctx.fieldName);
				if (!field?.values) return [];
				return field.values
					.filter((v) => !lower || v.toLowerCase().startsWith(lower))
					.map((v) => ({
						label: v.includes(" ") ? `"${v}"` : v,
						description: "",
					}));
			}

			case "keyword":
				return JQL_KEYWORDS.filter(
					(kw) => !lower || kw.toLowerCase().startsWith(lower),
				).map((kw) => ({ label: kw, description: "" }));

			default:
				return [];
		}
	}, [value]);

	/** Подставить выбранную подсказку. */
	const applySuggestion = useCallback(
		(suggestion: string) => {
			const partial = getPartialTokenFor(value);
			const base = partial
				? value.slice(0, value.length - partial.length)
				: value;
			const newValue = `${base + suggestion} `;
			onChange(newValue);
			setShowSuggestions(false);
			setSelectedIndex(0);
			inputRef.current?.focus();
		},
		[value, onChange],
	);

	/** Обработчик клавиш. */
	const handleKeyDown = useCallback(
		(e: React.KeyboardEvent) => {
			if (!showSuggestions || suggestions.length === 0) {
				if (e.key === "Enter") {
					e.preventDefault();
					onSubmit?.(value);
				}
				return;
			}

			switch (e.key) {
				case "ArrowDown":
					e.preventDefault();
					setSelectedIndex((prev) =>
						prev < suggestions.length - 1 ? prev + 1 : 0,
					);
					break;
				case "ArrowUp":
					e.preventDefault();
					setSelectedIndex((prev) =>
						prev > 0 ? prev - 1 : suggestions.length - 1,
					);
					break;
				case "Tab":
				case "Enter":
					e.preventDefault();
					if (suggestions[selectedIndex]) {
						applySuggestion(suggestions[selectedIndex].label);
					}
					break;
				case "Escape":
					e.preventDefault();
					setShowSuggestions(false);
					break;
			}
		},
		[
			showSuggestions,
			suggestions,
			selectedIndex,
			applySuggestion,
			onSubmit,
			value,
		],
	);

	/** Показывать подсказки при фокусе/вводе. */
	const handleChange = useCallback(
		(e: React.ChangeEvent<HTMLInputElement>) => {
			onChange(e.target.value);
			setShowSuggestions(true);
			setSelectedIndex(0);
		},
		[onChange],
	);

	/** Скрывать подсказки при клике снаружи. */
	useEffect(() => {
		const handleClickOutside = (e: MouseEvent) => {
			if (
				containerRef.current &&
				!containerRef.current.contains(e.target as Node)
			) {
				setShowSuggestions(false);
			}
		};
		document.addEventListener("mousedown", handleClickOutside);
		return () => document.removeEventListener("mousedown", handleClickOutside);
	}, []);

	return (
		<div ref={containerRef} className={cn("relative", className)}>
			<div className="relative">
				<svg
					aria-hidden="true"
					className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-t3/60"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
					/>
				</svg>
				<Input
					ref={inputRef}
					value={value}
					onChange={handleChange}
					onKeyDown={handleKeyDown}
					onFocus={() => setShowSuggestions(true)}
					placeholder={placeholder}
					className="pl-9 pr-3 font-mono text-sm"
					spellCheck={false}
					autoComplete="off"
				/>
			</div>

			{/* Выпадающий список подсказок */}
			{showSuggestions && suggestions.length > 0 && (
				<div className="absolute z-50 mt-1 w-full bg-raised border border-b2 rounded-[var(--r2)] shadow-[var(--sh2)] overflow-hidden">
					<div className="max-h-60 overflow-y-auto p-1">
						{suggestions.map((item, idx) => (
							<button
								key={item.label}
								type="button"
								className={cn(
									"flex w-full items-center gap-3 rounded-[var(--r1)] px-3 py-2 text-left text-[13px] transition-colors duration-150",
									idx === selectedIndex
										? "bg-blue/10 text-blue"
										: "hover:bg-[var(--hover)] text-t1",
								)}
								onMouseEnter={() => setSelectedIndex(idx)}
								onMouseDown={(e) => {
									e.preventDefault();
									applySuggestion(item.label);
								}}
							>
								<span className="font-mono font-medium">{item.label}</span>
								{item.description && (
									<span className="text-[12px] text-t3">
										{item.description}
									</span>
								)}
							</button>
						))}
					</div>
					<div className="border-t border-b1 px-3 py-1.5 text-[10px] text-t3">
						<kbd className="rounded border px-1 py-0.5 text-[10px]">Tab</kbd>{" "}
						или{" "}
						<kbd className="rounded border px-1 py-0.5 text-[10px]">Enter</kbd>{" "}
						-- подставить,{" "}
						<kbd className="rounded border px-1 py-0.5 text-[10px]">Esc</kbd> --
						скрыть
					</div>
				</div>
			)}
		</div>
	);
}
