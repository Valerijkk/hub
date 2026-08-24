import { createContext, useContext } from "react";
import type { MailDraft } from "@/api/mail";

interface MailComposerContextValue {
	readonly isComposerOpen: boolean;
	openReplyDraft: (draft: MailDraft) => void;
}

export const MailComposerContext =
	createContext<MailComposerContextValue | null>(null);

export function useMailComposerContext(): MailComposerContextValue {
	const value = useContext(MailComposerContext);
	if (value === null) {
		throw new Error("Mail composer context is unavailable");
	}
	return value;
}
