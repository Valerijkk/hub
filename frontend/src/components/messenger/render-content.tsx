/** Рендерер Plate JSON для содержимого сообщений. */

/**
 * Whitelist безопасных схем для <a href>. Блокирует `javascript:`,
 * `data:`, `vbscript:` и другие опасные схемы.
 */
const SAFE_URL_RE = /^(https?:|mailto:|\/|#)/i;

/**
 * Санитизация URL для href. Небезопасное → `"#"` (no-op).
 * Если строка не проходит whitelist схем.
 */
function sanitizeUrl(url: unknown): string {
	return typeof url === "string" && SAFE_URL_RE.test(url) ? url : "#";
}

/**
 * Безопасное приведение к строке для React children.
 * Предотвращает «Objects are not valid as React child».
 */
function safeString(value: unknown): string {
	if (value === null || value === undefined) return "";
	if (typeof value === "string") return value;
	if (typeof value === "number" || typeof value === "boolean")
		return String(value);
	// Объекты/массивы нельзя рендерить как children
	return "";
}

/**
 * Compute a stable React key for a Plate node based on its content.
 * Plate JSON nodes have no inherent ids, so we hash the serialized shape;
 * Одинаковые siblings могут иметь один key —
 * допустимо: output при re-render тоже идентичен.
 */
function nodeKey(node: Record<string, unknown>): string {
	try {
		const serialized = JSON.stringify(node);
		// FNV-1a 32-bit — быстрый детерминированный hash для key.
		let hash = 0x811c9dc5;
		for (let i = 0; i < serialized.length; i++) {
			hash ^= serialized.charCodeAt(i);
			hash = Math.imul(hash, 0x01000193);
		}
		const type = typeof node.type === "string" ? node.type : "node";
		return `${type}-${(hash >>> 0).toString(36)}`;
	} catch {
		const type = typeof node.type === "string" ? node.type : "node";
		return `${type}-fallback`;
	}
}

/** Рендерит один узел Plate JSON (текст, параграф, заголовок, список, цитату, ссылку, код). */
export function RenderNode({ node }: { node: Record<string, unknown> }) {
	if (!node || typeof node !== "object" || Array.isArray(node)) return null;

	// Текстовый узел Plate.
	if ("text" in node) {
		const text = safeString(node.text);
		if (!text) return null;
		let el: React.ReactNode = text;
		if (node.bold) el = <strong>{el}</strong>;
		if (node.italic) el = <em>{el}</em>;
		if (node.code)
			el = (
				<code className="rounded bg-raised px-1.5 py-0.5 text-xs font-mono">
					{el}
				</code>
			);
		if (node.strikethrough) el = <s>{el}</s>;
		return <>{el}</>;
	}

	const children = (
		Array.isArray(node.children) ? node.children : []
	) as Record<string, unknown>[];
	const renderedChildren = children.map((child) => {
		if (!child || typeof child !== "object" || Array.isArray(child))
			return null;
		// У узлов Plate нет стабильных id; JSON-хеш содержимого даёт одинаковый
		// key при re-render; смена порядка — корректный re-key.
		return (
			<RenderNode
				key={nodeKey(child as Record<string, unknown>)}
				node={child as Record<string, unknown>}
			/>
		);
	});

	switch (node.type) {
		case "sticker": {
			const emoji = typeof node.emoji === "string" ? node.emoji : "";
			const label = typeof node.label === "string" ? node.label : "";
			return (
				<span
					className="inline-block text-7xl leading-none select-none"
					title={label}
					role="img"
					aria-label={label}
				>
					{emoji}
				</span>
			);
		}
		case "p":
			return <p className="mb-1 leading-relaxed">{renderedChildren}</p>;
		case "h1":
			return <h1 className="text-lg font-bold">{renderedChildren}</h1>;
		case "h2":
			return <h2 className="text-base font-bold">{renderedChildren}</h2>;
		case "ul":
			return <ul className="list-disc pl-4">{renderedChildren}</ul>;
		case "ol":
			return <ol className="list-decimal pl-4">{renderedChildren}</ol>;
		case "li":
			return <li>{renderedChildren}</li>;
		case "blockquote":
			return (
				<blockquote className="border-l-2 border-primary/30 pl-3 italic text-t2">
					{renderedChildren}
				</blockquote>
			);
		case "a":
			return (
				<a
					href={sanitizeUrl(node.url)}
					className="text-primary underline decoration-primary/30 hover:decoration-primary"
					target="_blank"
					rel="noopener noreferrer"
				>
					{renderedChildren}
				</a>
			);
		case "code_block":
			return (
				<pre className="rounded-lg bg-raised p-3 text-xs font-mono overflow-x-auto">
					<code>{renderedChildren}</code>
				</pre>
			);
		default:
			return <>{renderedChildren}</>;
	}
}

/** Рендерит содержимое сообщения из Plate JSON (упрощённый вариант). */
export function RenderContent({
	contentJson,
}: {
	contentJson?: Record<string, unknown> | null;
}) {
	if (!contentJson || typeof contentJson !== "object") return null;

	// Если contentJson строка в runtime — рендерим как plain text.
	if (typeof contentJson === "string") {
		return <p className="mb-1 leading-relaxed">{contentJson}</p>;
	}

	// content_json может быть массивом напрямую (некоторые редакторы Plate так отдают).
	if (Array.isArray(contentJson)) {
		const nodes = contentJson as unknown[];
		if (nodes.length === 0) return null;
		return (
			<div className="prose prose-sm dark:prose-invert max-w-none">
				{nodes.map((node) => {
					if (!node || typeof node !== "object" || Array.isArray(node))
						return null;
					return (
						<RenderNode
							key={nodeKey(node as Record<string, unknown>)}
							node={node as Record<string, unknown>}
						/>
					);
				})}
			</div>
		);
	}

	const children = (contentJson as { children?: unknown[] }).children;
	if (!Array.isArray(children)) {
		// Fallback: если нет children, попробуем отрендерить как текст.
		const text = (contentJson as { text?: unknown }).text;
		if (typeof text === "string") return <p>{text}</p>;
		return null;
	}

	return (
		<div className="prose prose-sm dark:prose-invert max-w-none">
			{children.map((node) => {
				if (!node || typeof node !== "object" || Array.isArray(node))
					return null;
				return (
					<RenderNode
						key={nodeKey(node as Record<string, unknown>)}
						node={node as Record<string, unknown>}
					/>
				);
			})}
		</div>
	);
}
