/**
 * LexoRank helpers for Kanban card positioning (FE-P1-BOARD-DND / W10-A4-002).
 * Format mirrors Hub backend: `0|<base62body>:`.
 */

const RANK_ALPHABET =
	"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";

function stripRank(rank: string): string {
	return rank.replace(/^0\|/, "").replace(/:$/, "");
}

function wrapRank(body: string): string {
	return `0|${body}:`;
}

/** Increment last character of rank body (append-after). */
function rankAfter(rank: string): string {
	const body = stripRank(rank);
	if (!body) return "0|aaaaaa:";
	const last = body[body.length - 1] ?? "a";
	const idx = RANK_ALPHABET.indexOf(last);
	if (idx >= 0 && idx < RANK_ALPHABET.length - 1) {
		return wrapRank(body.slice(0, -1) + RANK_ALPHABET[idx + 1]);
	}
	return wrapRank(`${body}a`);
}

/** Rank strictly before given rank (prepend). */
function rankBefore(rank: string): string {
	const body = stripRank(rank);
	if (!body) return "0|aaaaaa:";
	const last = body[body.length - 1] ?? "a";
	const idx = RANK_ALPHABET.indexOf(last);
	if (idx > 0) {
		return wrapRank(body.slice(0, -1) + RANK_ALPHABET[idx - 1]);
	}
	return wrapRank(`0${body}`);
}

/** Midpoint between two rank bodies when possible. */
function rankBetween(prev: string, next: string): string {
	const a = stripRank(prev);
	const b = stripRank(next);
	if (a >= b) return rankAfter(prev);
	const maxLen = Math.max(a.length, b.length);
	const padA = a.padEnd(maxLen, "0");
	const padB = b.padEnd(maxLen, "0");
	let mid = "";
	for (let i = 0; i < maxLen; i++) {
		const ai = RANK_ALPHABET.indexOf(padA[i] ?? "0");
		const bi = RANK_ALPHABET.indexOf(padB[i] ?? "0");
		const ca = ai >= 0 ? ai : 0;
		const cb = bi >= 0 ? bi : 0;
		if (ca === cb) {
			mid += padA[i] ?? "0";
			continue;
		}
		if (cb - ca > 1) {
			mid += RANK_ALPHABET[Math.floor((ca + cb) / 2)] ?? "a";
			return wrapRank(mid);
		}
		mid += padA[i] ?? "0";
		return rankAfter(wrapRank(mid));
	}
	return rankAfter(prev);
}

/**
 * Compute insert rank for target column issues at index.
 * `issues` must already exclude the moved card.
 */
export function computeInsertRank(
	issues: Array<{ rank?: string }>,
	index: number,
): string {
	if (issues.length === 0) return "0|hzzzzz:";

	const clamped = Math.max(0, Math.min(index, issues.length));
	const prev = clamped > 0 ? issues[clamped - 1]?.rank : undefined;
	const next = clamped < issues.length ? issues[clamped]?.rank : undefined;

	if (!prev && !next) return "0|hzzzzz:";
	if (!prev && next) return rankBefore(next);
	if (prev && !next) return rankAfter(prev);
	if (prev && next) return rankBetween(prev, next);
	return "0|hzzzzz:";
}
