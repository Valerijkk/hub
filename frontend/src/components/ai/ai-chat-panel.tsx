import { useQuery } from "@tanstack/react-query";
import { KeyRound, Sparkles, X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
	type ChatCompletionResponse,
	type ChatMessage,
	chatCompletion,
	getAIUsageToday,
	type StreamStartResponse,
} from "@/api/ai";
import { BotMarkdown } from "@/components/messenger/bot-markdown";
import { Button } from "@/components/ui/button";
import { toastApiError } from "@/lib/format-api-error";
import { cn } from "@/lib/utils";
import { useAIStore } from "@/stores/ai-store";
import { useWsStore } from "@/stores/ws-store";
import { BYOKSettings } from "./byok-settings";
import { ModelPicker } from "./model-picker";

interface PanelMessage {
	id: string;
	role: "user" | "assistant";
	content: string;
}

function isStreamStart(
	r: ChatCompletionResponse | StreamStartResponse,
): r is StreamStartResponse {
	return "status" in r && r.status === "streaming";
}

/** AI side panel — Ctrl+Shift+A (Phase 17 W3, WS-стриминг). */
export function AIChatPanel() {
	const open = useAIStore((s) => s.panelOpen);
	const setPanelOpen = useAIStore((s) => s.setPanelOpen);
	const selectedModel = useAIStore((s) => s.selectedModel);
	const consumePrefill = useAIStore((s) => s.consumePrefill);
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);
	const [input, setInput] = useState("");
	const [loading, setLoading] = useState(false);
	const [showBYOK, setShowBYOK] = useState(false);
	const [messages, setMessages] = useState<PanelMessage[]>([]);
	const scrollRef = useRef<HTMLDivElement>(null);
	const streamHandlerRef = useRef<
		((event: string, payload: unknown) => void) | null
	>(null);

	const { data: usage } = useQuery({
		queryKey: ["ai", "usage", "today"],
		queryFn: getAIUsageToday,
		enabled: open,
		staleTime: 30_000,
	});

	useEffect(() => {
		if (!open) return;
		const prefill = consumePrefill();
		if (prefill) {
			setInput(prefill);
		}
	}, [open, consumePrefill]);

	useEffect(() => {
		return () => {
			if (streamHandlerRef.current) {
				removeListener(streamHandlerRef.current);
			}
		};
	}, [removeListener]);

	const appendAssistantDelta = useCallback((delta: string) => {
		setMessages((prev) => {
			const copy = [...prev];
			const last = copy[copy.length - 1];
			if (last?.role === "assistant") {
				copy[copy.length - 1] = {
					id: last.id,
					role: "assistant",
					content: last.content + delta,
				};
			}
			return copy;
		});
	}, []);

	const send = useCallback(async () => {
		const text = input.trim();
		if (!text || !selectedModel || loading) return;
		setInput("");
		const userMsg: PanelMessage = {
			id: crypto.randomUUID(),
			role: "user",
			content: text,
		};
		const history: PanelMessage[] = [...messages, userMsg];
		setMessages([
			...history,
			{ id: crypto.randomUUID(), role: "assistant", content: "" },
		]);
		setLoading(true);

		const apiMessages: ChatMessage[] = history.map((m) => ({
			role: m.role,
			content: m.content,
		}));

		const streamId = crypto.randomUUID();

		const handler = (event: string, payload: unknown) => {
			const p = payload as {
				stream_id?: string;
				content?: string;
				message?: string;
			};
			if (p.stream_id !== streamId) return;

			if (event === "ai.stream.delta" && p.content) {
				appendAssistantDelta(p.content);
			}
			if (event === "ai.stream.done") {
				setLoading(false);
				if (streamHandlerRef.current) {
					removeListener(streamHandlerRef.current);
					streamHandlerRef.current = null;
				}
			}
			if (event === "ai.stream.error") {
				setLoading(false);
				toastApiError(new Error(p.message ?? "stream error"));
				if (streamHandlerRef.current) {
					removeListener(streamHandlerRef.current);
					streamHandlerRef.current = null;
				}
			}
		};

		streamHandlerRef.current = handler;
		addListener(handler);

		try {
			const resp = await chatCompletion({
				model: selectedModel,
				messages: apiMessages,
				stream: true,
				stream_id: streamId,
			});

			if (!isStreamStart(resp)) {
				const full =
					(
						resp as ChatCompletionResponse
					).choices[0]?.message?.content?.trim() ?? "";
				setMessages((prev) => {
					const copy = [...prev];
					copy[copy.length - 1] = {
						id: copy[copy.length - 1]?.id ?? crypto.randomUUID(),
						role: "assistant",
						content: full,
					};
					return copy;
				});
				setLoading(false);
				if (streamHandlerRef.current) {
					removeListener(streamHandlerRef.current);
					streamHandlerRef.current = null;
				}
			}
		} catch (err) {
			setLoading(false);
			if (streamHandlerRef.current) {
				removeListener(streamHandlerRef.current);
				streamHandlerRef.current = null;
			}
			setMessages((prev) => prev.slice(0, -1));
			toastApiError(err, "AI запрос не удался");
		}

		requestAnimationFrame(() => {
			scrollRef.current?.scrollTo({
				top: scrollRef.current.scrollHeight,
				behavior: "smooth",
			});
		});
	}, [
		input,
		selectedModel,
		loading,
		messages,
		addListener,
		removeListener,
		appendAssistantDelta,
	]);

	if (!open) return null;

	return (
		<div
			className={cn(
				"fixed inset-y-0 right-0 z-[60] w-full max-w-md flex flex-col",
				"border-l border-b1 bg-base shadow-[var(--sh3)] animate-[eu_300ms_var(--spring)]",
			)}
			role="dialog"
			aria-label="AI Chat"
		>
			<div className="flex items-center gap-2 border-b border-b1 px-4 py-3">
				<Sparkles className="h-4 w-4 text-purple-500" />
				<div className="flex-1 min-w-0">
					<span className="text-sm font-semibold block">Hub AI</span>
					{usage && (
						<span className="text-[11px] text-t3 truncate block">
							Сегодня: {usage.request_count}/{usage.daily_limit} запросов · ~
							{usage.tokens_in + usage.tokens_out} токенов
						</span>
					)}
				</div>
				<ModelPicker />
				<Button
					type="button"
					variant="ghost"
					size="icon"
					className="h-8 w-8"
					onClick={() => setShowBYOK((v) => !v)}
					aria-label="BYOK keys"
				>
					<KeyRound className="h-4 w-4" />
				</Button>
				<Button
					type="button"
					variant="ghost"
					size="icon"
					className="h-8 w-8"
					onClick={() => setPanelOpen(false)}
					aria-label="Закрыть"
				>
					<X className="h-4 w-4" />
				</Button>
			</div>

			{showBYOK && (
				<div className="border-b border-b1 p-3 max-h-48 overflow-y-auto">
					<BYOKSettings onChanged={() => useAIStore.getState().loadModels()} />
				</div>
			)}

			<div ref={scrollRef} className="flex-1 overflow-y-auto p-4 space-y-4">
				{messages.length === 0 && (
					<p className="text-[13px] text-t3 text-center py-8">
						Спросите про код, тесты или задачи команды
					</p>
				)}
				{messages.map((m) => (
					<div
						key={m.id}
						className={cn(
							"rounded-lg px-3 py-2 text-sm",
							m.role === "user"
								? "bg-primary/10 ml-8"
								: "bg-raised mr-4 border border-b1",
						)}
					>
						{m.role === "assistant" ? (
							m.content ? (
								<BotMarkdown text={m.content} />
							) : loading ? (
								<span className="text-t3 animate-pulse">...</span>
							) : null
						) : (
							<p className="whitespace-pre-wrap">{m.content}</p>
						)}
					</div>
				))}
			</div>

			<div className="border-t border-b1 p-3 flex gap-2">
				<textarea
					value={input}
					onChange={(e) => setInput(e.target.value)}
					onKeyDown={(e) => {
						if (e.key === "Enter" && !e.shiftKey) {
							e.preventDefault();
							void send();
						}
					}}
					placeholder={selectedModel ? "Сообщение..." : "Выберите модель"}
					disabled={!selectedModel || loading}
					rows={2}
					className="flex-1 resize-none rounded-lg border border-b1 bg-[var(--input)] px-3 py-2 text-[13px]"
				/>
				<Button
					type="button"
					disabled={!selectedModel || loading || !input.trim()}
					onClick={() => void send()}
				>
					Отправить
				</Button>
			</div>
		</div>
	);
}
