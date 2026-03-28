import { createFileRoute } from "@tanstack/react-router";
import { Component, useEffect, type ErrorInfo, type ReactNode } from "react";
import { ChatView } from "@/components/messenger/chat-view";
import { useMessengerStore } from "@/stores/messenger-store";

/** Error boundary для чата — ловит React #185 и показывает fallback. */
class ChatErrorBoundary extends Component<
	{ children: ReactNode },
	{ error: Error | null }
> {
	state: { error: Error | null } = { error: null };

	static getDerivedStateFromError(error: Error) {
		return { error };
	}

	componentDidCatch(error: Error, info: ErrorInfo) {
		console.error("ChatView crashed:", error.message, info.componentStack);
	}

	render() {
		if (this.state.error) {
			return (
				<div className="flex flex-1 items-center justify-center p-8">
					<div className="text-center max-w-sm">
						<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-red-500/10">
							<svg
								className="h-8 w-8 text-red-500"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={1.5}
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
								/>
							</svg>
						</div>
						<h3 className="text-lg font-semibold mb-2">
							Ошибка загрузки чата
						</h3>
						<p className="text-sm text-t2 mb-4">
							{this.state.error.message}
						</p>
						<button
							type="button"
							className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
							onClick={() => this.setState({ error: null })}
						>
							Попробовать снова
						</button>
					</div>
				</div>
			);
		}
		return this.props.children;
	}
}

function ChannelPage() {
	const { channelId } = Route.useParams();
	const setActiveChannel = useMessengerStore((s) => s.setActiveChannel);
	const resetUnread = useMessengerStore((s) => s.resetUnread);

	useEffect(() => {
		setActiveChannel(channelId);
		resetUnread(channelId);
		return () => {
			setActiveChannel(null);
		};
	}, [channelId, setActiveChannel, resetUnread]);

	return (
		<div className="flex h-full flex-col overflow-hidden">
			<ChatErrorBoundary>
				<ChatView channelId={channelId} />
			</ChatErrorBoundary>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/messenger/$channelId")({
	component: ChannelPage,
});
