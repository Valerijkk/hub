import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	resolveMailSchedulingParticipants,
	type MailAddress,
	type MailThreadMessageSummary,
} from "@/api/mail";
import { CreateCallDialog } from "@/modules/calls/create-call-dialog";
import { useAuthStore } from "@/stores/auth-store";
import { Button } from "@/components/ui/button";

type MailScheduleNotice = "empty" | "unresolved" | "failed";

interface MailScheduleActionProps {
	message: Pick<
		MailThreadMessageSummary,
		"id" | "from" | "to" | "cc" | "subject"
	>;
}

function normalizeAddress(value: string): string | null {
	const normalized = value.trim().normalize("NFC").toLocaleLowerCase("und");
	return normalized.length > 0 ? normalized : null;
}

function collectSchedulingAddresses(
	message: MailScheduleActionProps["message"],
	currentUserEmail: string | undefined,
): string[] {
	const currentUserAddress = currentUserEmail
		? normalizeAddress(currentUserEmail)
		: null;
	const knownAddresses = new Set<string>();
	const addresses: MailAddress[] = [message.from, ...message.to, ...message.cc];

	return addresses.flatMap((address) => {
		const normalized = normalizeAddress(address.email);
		if (
			normalized === null ||
			normalized === currentUserAddress ||
			knownAddresses.has(normalized)
		) {
			return [];
		}
		knownAddresses.add(normalized);
		return [normalized];
	});
}

function resolvedParticipantIDs(
	participants: Awaited<
		ReturnType<typeof resolveMailSchedulingParticipants>
	>["participants"],
	currentUserID: string | undefined,
): string[] {
	const currentUserKey = currentUserID?.trim().toLocaleLowerCase("und");
	const knownIDs = new Set<string>();

	return participants.flatMap((participant) => {
		if (typeof participant.user_id !== "string") return [];
		const normalized = participant.user_id.trim().toLocaleLowerCase("und");
		if (
			normalized.length === 0 ||
			normalized === currentUserKey ||
			knownIDs.has(normalized)
		) {
			return [];
		}
		knownIDs.add(normalized);
		return [normalized];
	});
}

export function MailScheduleAction({ message }: MailScheduleActionProps) {
	const { t } = useTranslation("mail");
	const currentUser = useAuthStore((state) => state.user);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [initialParticipantIDs, setInitialParticipantIDs] = useState<string[]>([]);
	const [initialTitle, setInitialTitle] = useState("");
	const [isResolving, setIsResolving] = useState(false);
	const [notice, setNotice] = useState<MailScheduleNotice | null>(null);
	const addresses = useMemo(
		() => collectSchedulingAddresses(message, currentUser?.email),
		[currentUser?.email, message],
	);

	const schedule = async (): Promise<void> => {
		if (isResolving) return;
		setNotice(null);
		setDialogOpen(false);
		if (addresses.length === 0) {
			setNotice("empty");
			return;
		}

		setIsResolving(true);
		try {
			const result = await resolveMailSchedulingParticipants(message.id, {
				addresses,
			});
			const participantIDs = resolvedParticipantIDs(
				Array.isArray(result.participants) ? result.participants : [],
				currentUser?.id,
			);
			if (participantIDs.length === 0) {
				setNotice("unresolved");
				return;
			}
			setInitialParticipantIDs(participantIDs);
			setInitialTitle(message.subject);
			setDialogOpen(true);
		} catch {
			setNotice("failed");
		} finally {
			setIsResolving(false);
		}
	};

	return (
		<section className="mt-4" aria-label={t("reader.schedule.label")}>
			<Button
				type="button"
				variant="outline"
				size="sm"
				className="min-h-11"
				disabled={isResolving}
				aria-busy={isResolving}
				onClick={() => {
					void schedule();
				}}
			>
				{isResolving
					? t("reader.schedule.resolving")
					: t("reader.schedule.action")}
			</Button>
			{notice ? (
				<p className="mt-2 text-xs text-t2" role="alert">
					{t(`reader.schedule.${notice}`)}
				</p>
			) : null}
			<CreateCallDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				initialParticipantIds={initialParticipantIDs}
				initialTitle={initialTitle}
			/>
		</section>
	);
}
