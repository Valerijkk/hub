import { suggestComposerCompletion } from "@/api/ai";

const MIN_PREFIX_CHARS = 8;

export interface ComposerSuggestState {
	suggestion: string | null;
	suggestionFor: string | null;
	isLoading: boolean;
}

export function canSuggestComposer(prefix: string): boolean {
	return prefix.trim().length >= MIN_PREFIX_CHARS;
}

/** Вставляет кэшированную или свежую AI-подсказку в текст композера с корректным пробелом. */
export function applyComposerSuggestion(
	currentText: string,
	suggestion: string,
): string {
	const trimmed = currentText.trimEnd();
	const sep = trimmed.endsWith(" ") || trimmed === "" ? "" : " ";
	return trimmed + sep + suggestion;
}

export async function fetchComposerSuggestion(
	prefix: string,
	channelId?: string,
): Promise<string | null> {
	if (!canSuggestComposer(prefix)) {
		return null;
	}
	try {
		const resp = await suggestComposerCompletion({
			prefix: prefix.trim(),
			channel_id: channelId,
		});
		return resp.suggestion?.trim() || null;
	} catch {
		return null;
	}
}

export { MIN_PREFIX_CHARS };
