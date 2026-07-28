import { Check, Copy } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

interface BotMarkdownProps {
	text: string;
	className?: string;
}

type Block =
	| { kind: "paragraph"; lines: string[] }
	| { kind: "code"; lang: string; code: string };

/** Лёгкий markdown: абзацы и fenced code blocks. */
function parseBlocks(text: string): Block[] {
	const blocks: Block[] = [];
	const parts = text.split("```");
	for (let i = 0; i < parts.length; i++) {
		const chunk = parts[i] ?? "";
		if (i % 2 === 1) {
			const nl = chunk.indexOf("\n");
			const lang = nl >= 0 ? chunk.slice(0, nl).trim() : "";
			const code = nl >= 0 ? chunk.slice(nl + 1) : chunk;
			blocks.push({ kind: "code", lang, code: code.replace(/\n$/, "") });
		} else if (chunk.trim()) {
			blocks.push({
				kind: "paragraph",
				lines: chunk.split("\n").filter((l) => l.trim() !== ""),
			});
		}
	}
	if (blocks.length === 0 && text.trim()) {
		blocks.push({ kind: "paragraph", lines: [text] });
	}
	return blocks;
}

/** Стабильный key блоков markdown без index массива. */
function blockKey(block: Block): string {
	if (block.kind === "code") {
		return `code:${block.lang}:${block.code.length}:${block.code.slice(0, 48)}`;
	}
	return `p:${block.lines.join("\n").slice(0, 96)}`;
}

/** Стабильный key строки внутри абзаца. */
function lineKey(line: string, occurrence: number): string {
	return `line:${occurrence}:${line.slice(0, 64)}:${line.length}`;
}

/** Inline-разметка: **жирный**, *курсив*, `code`. */
function renderInline(line: string): React.ReactNode[] {
	const nodes: React.ReactNode[] = [];
	const re = /(\*\*[^*]+\*\*|\*[^*]+\*|_[^_]+_|`[^`]+`)/g;
	let last = 0;
	let key = 0;
	for (let match = re.exec(line); match !== null; match = re.exec(line)) {
		const m = match;
		if (m.index > last) {
			nodes.push(line.slice(last, m.index));
		}
		const token = m[0];
		if (token.startsWith("**")) {
			nodes.push(<strong key={key++}>{token.slice(2, -2)}</strong>);
		} else if (token.startsWith("*") || token.startsWith("_")) {
			nodes.push(<em key={key++}>{token.slice(1, -1)}</em>);
		} else if (token.startsWith("`")) {
			nodes.push(
				<code
					key={key++}
					className="rounded bg-raised px-1 py-0.5 text-xs font-mono"
				>
					{token.slice(1, -1)}
				</code>,
			);
		}
		last = m.index + token.length;
	}
	if (last < line.length) {
		nodes.push(line.slice(last));
	}
	return nodes;
}

function CodeBlock({ lang, code }: { lang: string; code: string }) {
	const [copied, setCopied] = useState(false);
	const copyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	useEffect(
		() => () => {
			if (copyTimerRef.current) clearTimeout(copyTimerRef.current);
		},
		[],
	);

	const onCopy = useCallback(async () => {
		try {
			await navigator.clipboard.writeText(code);
			setCopied(true);
			toast.success("Скопировано");
			if (copyTimerRef.current) clearTimeout(copyTimerRef.current);
			copyTimerRef.current = setTimeout(() => setCopied(false), 2000);
		} catch {
			toast.error("Не удалось скопировать");
		}
	}, [code]);

	return (
		<div className="relative my-2 rounded-lg border border-b1 bg-[var(--raised)] overflow-hidden">
			<div className="flex items-center justify-between px-3 py-1 border-b border-b1 text-[11px] text-t3">
				<span>{lang || "code"}</span>
				<button
					type="button"
					onClick={() => void onCopy()}
					className="inline-flex items-center gap-1 hover:text-t1 transition-colors"
					aria-label="Копировать код"
				>
					{copied ? (
						<Check className="h-3 w-3" />
					) : (
						<Copy className="h-3 w-3" />
					)}
					Копировать
				</button>
			</div>
			<pre className="overflow-x-auto p-3 text-xs font-mono leading-relaxed">
				<code>{code}</code>
			</pre>
		</div>
	);
}

/** Рендер markdown для ответов бота @hub-ai (M3-012, M3-013). */
export function BotMarkdown({ text, className }: BotMarkdownProps) {
	const blocks = parseBlocks(text);

	return (
		<div className={cn("space-y-1.5 text-sm leading-relaxed", className)}>
			{blocks.map((block) => {
				if (block.kind === "code") {
					return (
						<CodeBlock
							key={blockKey(block)}
							lang={block.lang}
							code={block.code}
						/>
					);
				}
				const lineCounts = new Map<string, number>();
				return (
					<div key={blockKey(block)} className="space-y-1">
						{block.lines.map((line) => {
							const occurrence = (lineCounts.get(line) ?? 0) + 1;
							lineCounts.set(line, occurrence);
							const key = lineKey(line, occurrence);
							if (line.startsWith("- ") || line.startsWith("* ")) {
								return (
									<ul key={key} className="list-disc pl-4">
										<li>{renderInline(line.slice(2))}</li>
									</ul>
								);
							}
							if (/^\d+\.\s/.test(line)) {
								const content = line.replace(/^\d+\.\s/, "");
								return (
									<ol key={key} className="list-decimal pl-4">
										<li>{renderInline(content)}</li>
									</ol>
								);
							}
							return <p key={key}>{renderInline(line)}</p>;
						})}
					</div>
				);
			})}
		</div>
	);
}

/** true, если content — markdown payload @hub-ai. */
export function isHubAIMarkdown(
	contentJson?: Record<string, unknown> | null,
): boolean {
	return contentJson?.hub_ai_markdown === true;
}
