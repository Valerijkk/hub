import { useMemo } from "react";

const HTML_NAMESPACE = "http://www.w3.org/1999/xhtml";
const MAX_UTF8_BYTES = 8_388_608;
const MAX_NODE_COUNT = 100_000;
const MAX_DEPTH = 64;
const ALLOWED_TAGS = new Set([
	"p",
	"br",
	"hr",
	"strong",
	"b",
	"em",
	"i",
	"u",
	"s",
	"del",
	"code",
	"pre",
	"blockquote",
	"ul",
	"ol",
	"li",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"div",
	"span",
]);
const FORBIDDEN_TEXT =
	/[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F-\u009F\u061C\u200E\u200F\u202A-\u202E\u2066-\u2069]/u;
const BCP47_LIKE_LANGUAGE =
	/^[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8}){0,3}$/;
const RTL_PRIMARY_LANGUAGES = new Set([
	"ar",
	"arc",
	"ckb",
	"dv",
	"fa",
	"he",
	"ks",
	"nqo",
	"ps",
	"sd",
	"ug",
	"ur",
	"yi",
]);
const CONTENT_SECURITY_POLICY =
	"default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'; frame-src 'none'; img-src 'none'; media-src 'none'; object-src 'none'; script-src 'none'; connect-src 'none'; font-src 'none'; style-src 'unsafe-inline'";
const RICH_DOCUMENT_STYLE = [
	":root{color-scheme:light dark}",
	"html{background:#fff;color:#1d1d1f}",
	'body{margin:0;padding:1rem;background:#fff;color:#1d1d1f;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;font-size:.9375rem;line-height:1.7;overflow-wrap:anywhere}',
	"main{max-width:100%}",
	"pre,code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}",
	"pre{overflow:auto;padding:.75rem;border-radius:.5rem;background:#f2f2f7;white-space:pre-wrap;overflow-wrap:anywhere}",
	"blockquote{margin:1rem 0;padding-inline-start:1rem;border-inline-start:.25rem solid #c7c7cc;color:#515154}",
	"hr{border:0;border-top:1px solid #d1d1d6;margin:1.25rem 0}",
	"@media (prefers-color-scheme:dark){html,body{background:#1c1c1e;color:#f5f5f7}pre{background:#2c2c2e}blockquote{border-inline-start-color:#636366;color:#d1d1d6}hr{border-top-color:#48484a}}",
].join("");

export interface MailSafeContentProps {
	readonly sanitizedHTML: string | null;
	readonly plainBody: string;
	readonly title: string;
	readonly plainUnavailable: string;
	readonly richUnavailable: string;
	readonly language: string;
}

function utf8ByteLength(value: string): number | null {
	if (typeof TextEncoder === "undefined") {
		return null;
	}

	return new TextEncoder().encode(value).byteLength;
}

function serializeValidatedFragment(source: string): string | null {
	if (typeof document === "undefined" || FORBIDDEN_TEXT.test(source)) {
		return null;
	}

	const template = document.createElement("template");
	template.innerHTML = source;
	let nodeCount = 0;
	let valid = true;

	const visit = (node: Node, depth: number): void => {
		if (!valid) {
			return;
		}

		nodeCount += 1;
		if (nodeCount > MAX_NODE_COUNT || depth > MAX_DEPTH) {
			valid = false;
			return;
		}

		if (node.nodeType === 3) {
			if (FORBIDDEN_TEXT.test(node.nodeValue ?? "")) {
				valid = false;
			}
			return;
		}

		if (node.nodeType !== 1) {
			valid = false;
			return;
		}

		const element = node as Element;
		if (
			element.namespaceURI !== HTML_NAMESPACE ||
			element.localName !== element.localName.toLowerCase() ||
			!ALLOWED_TAGS.has(element.localName) ||
			element.attributes.length !== 0
		) {
			valid = false;
			return;
		}

		for (const child of Array.from(element.childNodes)) {
			visit(child, depth + 1);
		}
	};

	for (const child of Array.from(template.content.childNodes)) {
		visit(child, 1);
	}

	return valid ? template.innerHTML : null;
}

function toSafeHTML(source: string): string | null {
	const sourceBytes = utf8ByteLength(source);
	if (sourceBytes === null || sourceBytes > MAX_UTF8_BYTES) {
		return null;
	}

	const serialized = serializeValidatedFragment(source);
	if (serialized === null) {
		return null;
	}

	const serializedBytes = utf8ByteLength(serialized);
	if (serializedBytes === null || serializedBytes > MAX_UTF8_BYTES) {
		return null;
	}

	const reparsed = serializeValidatedFragment(serialized);
	return reparsed === serialized ? serialized : null;
}

function getDocumentMetadata(language: string): {
	readonly language: string;
	readonly direction: "ltr" | "rtl";
} {
	const candidate = language.trim();
	const documentLanguage = BCP47_LIKE_LANGUAGE.test(candidate)
		? candidate.toLowerCase()
		: "en";
	const primaryLanguage = documentLanguage.split("-", 1)[0] ?? "en";

	return {
		language: documentLanguage,
		direction: RTL_PRIMARY_LANGUAGES.has(primaryLanguage) ? "rtl" : "ltr",
	};
}

function createSafeDocument(
	fragment: string,
	language: string,
	direction: "ltr" | "rtl",
): string {
	return [
		'<!doctype html><html lang="',
		language,
		'" dir="',
		direction,
		'"><head><meta http-equiv="Content-Security-Policy" content="',
		CONTENT_SECURITY_POLICY,
		'"><meta charset="utf-8"><meta name="referrer" content="no-referrer"><meta name="viewport" content="width=device-width, initial-scale=1"><style>',
		RICH_DOCUMENT_STYLE,
		"</style></head><body><main>",
		fragment,
		"</main></body></html>",
	].join("");
}

export function MailSafeContent({
	sanitizedHTML,
	plainBody,
	title,
	plainUnavailable,
	richUnavailable,
	language,
}: MailSafeContentProps) {
	const hasRichSource =
		sanitizedHTML !== null && sanitizedHTML.trim().length > 0;
	const safeHTML = useMemo(
		() =>
			hasRichSource && sanitizedHTML !== null
				? toSafeHTML(sanitizedHTML)
				: null,
		[hasRichSource, sanitizedHTML],
	);
	const documentMetadata = useMemo(
		() => getDocumentMetadata(language),
		[language],
	);
	const sourceDocument = useMemo(
		() =>
			safeHTML === null
				? null
				: createSafeDocument(
						safeHTML,
						documentMetadata.language,
						documentMetadata.direction,
					),
		[documentMetadata.direction, documentMetadata.language, safeHTML],
	);

	return (
		<div className="rounded-[var(--r2)] bg-base/60 p-4">
			{sourceDocument !== null ? (
				<iframe
					className="h-[min(65vh,44rem)] min-h-[20rem] w-full rounded-[var(--r2)] border border-b1 bg-card focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--blue)] focus-visible:ring-offset-2"
					loading="lazy"
					referrerPolicy="no-referrer"
					sandbox=""
					srcDoc={sourceDocument}
					title={title}
				/>
			) : (
				<>
					{hasRichSource ? (
						<p className="mb-2 text-xs leading-5 text-t3">
							{richUnavailable}
						</p>
					) : null}
					{plainBody.length > 0 ? (
						<div className="whitespace-pre-wrap break-words text-sm leading-7 text-t1">
							{plainBody}
						</div>
					) : (
						<p className="text-sm text-t3">{plainUnavailable}</p>
					)}
				</>
			)}
		</div>
	);
}
