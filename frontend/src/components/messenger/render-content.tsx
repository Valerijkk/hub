/** Рендерер Plate JSON для содержимого сообщений. */

/**
 * Whitelist of safe URL schemes/prefixes for <a href>. Blocks `javascript:`,
 * `data:`, `vbscript:`, and other dangerous schemes that can execute code.
 */
const SAFE_URL_RE = /^(https?:|mailto:|\/|#)/i;

/**
 * Sanitize a URL for use as an `<a href>` value. Returns `"#"` (safe no-op)
 * for anything that is not a string matching the safe-scheme whitelist.
 */
function sanitizeUrl(url: unknown): string {
	return typeof url === "string" && SAFE_URL_RE.test(url) ? url : "#";
}

/**
 * Safely converts any value to a React-renderable string.
 * Prevents "Objects are not valid as React child" errors.
 */
function safeString(value: unknown): string {
	if (value === null || value === undefined) return "";
	if (typeof value === "string") return value;
	if (typeof value === "number" || typeof value === "boolean") return String(value);
	// Prevent rendering objects/arrays as React children
	return "";
}

/**
 * Compute a stable React key for a Plate node based on its content.
 * Plate JSON nodes have no inherent ids, so we hash the serialized shape;
 * identical siblings (e.g. two empty paragraphs) get the same key — that's
 * acceptable here because re-render output is also identical for them.
 */
function nodeKey(node: Record<string, unknown>): string {
	try {
		const serialized = JSON.stringify(node);
		// FNV-1a 32-bit — small, fast, deterministic.
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

	// Text node.
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

	const children =
		(Array.isArray(node.children) ? node.children : []) as Record<string, unknown>[];
	const renderedChildren = children.map((child) => {
		if (!child || typeof child !== "object" || Array.isArray(child)) return null;
		// Plate nodes have no stable ids; use a JSON content hash so identical
		// re-renders reuse the same key but order-changes still re-key correctly.
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

	// Safety: if contentJson is somehow a string at runtime, render directly.
	if (typeof contentJson === "string") {
		return <p className="mb-1 leading-relaxed">{contentJson}</p>;
	}

	// Handle array-shaped content_json (some Plate editors emit arrays directly).
	if (Array.isArray(contentJson)) {
		const nodes = contentJson as unknown[];
		if (nodes.length === 0) return null;
		return (
			<div className="prose prose-sm dark:prose-invert max-w-none">
				{nodes.map((node) => {
					if (!node || typeof node !== "object" || Array.isArray(node)) return null;
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
				if (!node || typeof node !== "object" || Array.isArray(node)) return null;
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
