import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { SendMessageRequest } from "@/api/messenger";
import { MessageComposer } from "@/components/messenger/message-composer";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { Toaster, toast } from "sonner";
import { renderWithProviders } from "../support/render";

const UUIDS = [
	"10000000-0000-4000-8000-000000000001",
	"10000000-0000-4000-8000-000000000002",
	"10000000-0000-4000-8000-000000000003",
] as const;

const user = {
	id: "composer-toast-user", email: "composer-toast-user@example.test",
	timezone: "UTC", language: "ru", system_role: "employee", onboarding_completed: true,
};
const sendMessage = vi.fn();
const sendTypingIndicator = vi.fn();
const uploadFile = vi.fn();
const invalidateMessages = vi.fn();
const randomUUID = vi.fn<() => string>();

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		sendMessage: (...args: unknown[]) => sendMessage(...args),
		sendTypingIndicator: (...args: unknown[]) => sendTypingIndicator(...args),
	};
});
vi.mock("@/api/files", () => ({
	uploadFile: (...args: unknown[]) => uploadFile(...args),
}));
vi.mock("@/hooks/use-invalidate", () => ({
	useInvalidateMessages: () => invalidateMessages,
}));
vi.mock("@/components/ui/popover", () => ({
	Popover: ({ children }: { children: ReactNode }) => <>{children}</>,
	PopoverContent: ({ children }: { children: ReactNode }) => <>{children}</>,
	PopoverTrigger: ({ children }: { children: ReactNode }) => <>{children}</>,
}));
vi.mock("@/components/messenger/mention-autocomplete", () => ({
	MentionAutocomplete: () => null,
}));
vi.mock("@/components/messenger/create-poll-dialog", () => ({
	CreatePollDialog: () => null,
}));
vi.mock("@/components/messenger/gif-picker", () => ({
	GifPicker: () => null,
}));
vi.mock("@/components/messenger/sticker-picker", () => ({
	buildStickerContentJson: vi.fn(),
	StickerPicker: () => null,
}));
vi.mock("@/components/messenger/voice-recorder", () => ({
	VoiceRecorder: () => null,
}));

interface Deferred<T> {
	promise: Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
}
function deferred<T>(): Deferred<T> {
	let resolve!: (value: T) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise;
		reject = rejectPromise;
	});
	return { promise, resolve, reject };
}
function composerTree(threadId: string) {
	return (
		<>
			<MessageComposer canSend channelId="channel-toast-retry" threadId={threadId} />
			<Toaster closeButton position="top-right" richColors />
		</>
	);
}
function renderComposer(threadId = "retry-scope") {
	return renderWithProviders(composerTree(threadId));
}
function composerInput() {
	return screen.getByRole("textbox");
}
function sendButton() {
	return screen.getByRole("button", { name: "Отправить сообщение" });
}
function sendText(value: string) {
	fireEvent.change(composerInput(), { target: { value } });
	fireEvent.click(sendButton());
}
async function retryControl() {
	return screen.findByRole("button", { name: "Повторить" });
}
async function cancelControl() {
	return screen.findByRole("button", { name: "Отменить отправку" });
}
async function expectRecoveryControlsGone() {
	await waitFor(() => {
		expect(screen.queryByRole("button", { name: "Повторить" })).not.toBeInTheDocument();
		expect(screen.queryByRole("button", { name: "Отменить отправку" })).not.toBeInTheDocument();
	});
}

describe("FE-P149-MESSENGER · real Sonner retry recovery controls", () => {
	let uuidIndex = 0;
	beforeEach(() => {
		sendMessage.mockReset();
		sendTypingIndicator.mockReset();
		uploadFile.mockReset();
		invalidateMessages.mockReset();
		randomUUID.mockReset();
		uuidIndex = 0;
		randomUUID.mockImplementation(() => UUIDS[uuidIndex++] ?? UUIDS.at(-1)!);
		uploadFile.mockResolvedValue("unused-attachment-id");
		useAuthStore.setState({
			accessToken: null, refreshToken: null, user, isAuthenticated: true,
		});
		useMessengerStore.getState().reset();
		vi.stubGlobal("crypto", { randomUUID });
	});
	afterEach(() => {
		act(() => {
			for (const pendingToast of toast.getToasts()) toast.dismiss(pendingToast.id);
		});
		cleanup();
		useAuthStore.setState({
			accessToken: null, refreshToken: null, user: null, isAuthenticated: false,
		});
		useMessengerStore.getState().reset();
		vi.unstubAllGlobals();
	});

	it("keeps Retry available through a deferred rejection and recovers with a later click", async () => {
		const firstRetry = deferred<unknown>();
		sendMessage
			.mockRejectedValueOnce(new Error("initial failed"))
			.mockReturnValueOnce(firstRetry.promise)
			.mockResolvedValueOnce({ id: "recovered" });
		renderComposer();
		sendText("retry once");
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const [, initialPayload] = sendMessage.mock.calls[0] as [string, SendMessageRequest];
		const retry = await retryControl();
		act(() => {
			fireEvent.click(retry);
			fireEvent.click(retry);
		});
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(2));
		const [, retriedPayload] = sendMessage.mock.calls[1] as [string, SendMessageRequest];
		expect(retriedPayload).toEqual(initialPayload);
		expect(retriedPayload.client_nonce).toBe(initialPayload.client_nonce);
		await act(async () => {
			firstRetry.reject(new Error("retry failed"));
			await firstRetry.promise.catch(() => undefined);
		});
		expect(await retryControl()).toBeInTheDocument();
		fireEvent.click(await retryControl());
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(3));
		await waitFor(() => {
			expect(invalidateMessages).toHaveBeenCalledTimes(1);
			expect(invalidateMessages).toHaveBeenCalledWith("retry-scope");
		});
		await expectRecoveryControlsGone();
	});

	it("uses native Cancel to remove a non-dismissible failed-send toast without retrying", async () => {
		sendMessage.mockRejectedValueOnce(new Error("cancelled failure"));
		renderComposer();
		sendText("cancel without retry");
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		fireEvent.click(await cancelControl());
		await expectRecoveryControlsGone();
		expect(sendMessage).toHaveBeenCalledTimes(1);
		expect(invalidateMessages).not.toHaveBeenCalled();
	});

	async function assertCancelledRetryDoesNotRetarget(outcome: "reject" | "resolve") {
		const oldRetry = deferred<unknown>();
		const newerSend = deferred<unknown>();
		sendMessage
			.mockRejectedValueOnce(new Error("old initial failed"))
			.mockReturnValueOnce(oldRetry.promise)
			.mockReturnValueOnce(newerSend.promise);
		const { rerender } = renderComposer("old-scope");
		sendText("old send");
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		fireEvent.click(await retryControl());
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(2));
		rerender(composerTree("newer-scope"));
		sendText("newer send");
		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(3));
		fireEvent.click(await cancelControl());
		await expectRecoveryControlsGone();
		await act(async () => {
			newerSend.resolve({ id: "newer" });
			await newerSend.promise;
		});
		await waitFor(() => {
			expect(invalidateMessages).toHaveBeenCalledTimes(1);
			expect(invalidateMessages).toHaveBeenCalledWith("newer-scope");
		});
		if (outcome === "reject") {
			await act(async () => {
				oldRetry.reject(new Error("late old retry failure"));
				await oldRetry.promise.catch(() => undefined);
			});
		} else {
			await act(async () => {
				oldRetry.resolve({ id: "late old retry success" });
				await oldRetry.promise;
			});
		}
		await expectRecoveryControlsGone();
		expect(invalidateMessages).toHaveBeenCalledTimes(1);
		expect(invalidateMessages).toHaveBeenCalledWith("newer-scope");
	}

	it("does not resurrect a cancelled retry after its late rejection", async () => {
		await assertCancelledRetryDoesNotRetarget("reject");
	});
	it("does not invalidate a newer scope after a cancelled retry later succeeds", async () => {
		await assertCancelledRetryDoesNotRetarget("resolve");
	});
});
