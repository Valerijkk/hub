import { Component, type ErrorInfo, type ReactNode } from "react";

/**
 * Универсальный Error Boundary для рантайм-ошибок внутри маршрутов (FE-H4).
 *
 * Раньше единственный ErrorBoundary жил в `ChatView` (мессенджер), поэтому
 * любой неуправляемый throw в дашборде, звонках, трекере или TMS приводил
 * к белому экрану всего приложения. Этот boundary локализует падение в
 * пределах раздела и предлагает пользователю явный retry.
 *
 * Пропс `module` используется в сообщении и логе для быстрой идентификации
 * упавшего раздела.
 */
interface RouteErrorBoundaryProps {
	/** Дочерние элементы — содержимое маршрута. */
	children: ReactNode;
	/** Человеко-читаемое имя модуля (например, "Звонки", "Дашборд"). */
	module: string;
	/** Опциональный fallback-узел вместо стандартного. */
	fallback?: (error: Error, reset: () => void) => ReactNode;
}

interface RouteErrorBoundaryState {
	error: Error | null;
}

export class RouteErrorBoundary extends Component<
	RouteErrorBoundaryProps,
	RouteErrorBoundaryState
> {
	state: RouteErrorBoundaryState = { error: null };

	static getDerivedStateFromError(error: Error): RouteErrorBoundaryState {
		return { error };
	}

	componentDidCatch(error: Error, info: ErrorInfo): void {
		// Лог в консоль — для разработки. В продакшене сюда стоит подключить
		// внешний сборщик (Sentry/etc.), но это вне скоупа FE-H4.
		console.error(
			`[${this.props.module}] route crashed:`,
			error.message,
			info.componentStack,
		);
	}

	private reset = (): void => {
		this.setState({ error: null });
	};

	render(): ReactNode {
		const { error } = this.state;
		if (!error) return this.props.children;

		if (this.props.fallback) {
			return this.props.fallback(error, this.reset);
		}

		return (
			<div
				role="alert"
				aria-live="assertive"
				className="flex flex-1 items-center justify-center p-8"
			>
				<div className="max-w-sm text-center">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-[var(--red)]/10">
						<svg
							className="h-8 w-8 text-[var(--red)]"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={1.5}
							aria-hidden="true"
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
							/>
						</svg>
					</div>
					<h3 className="mb-2 text-lg font-semibold">
						Ошибка в разделе «{this.props.module}»
					</h3>
					<p className="mb-4 text-sm text-t2">{error.message}</p>
					<div className="flex justify-center gap-2">
						<button
							type="button"
							className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
							onClick={this.reset}
						>
							Попробовать снова
						</button>
						<button
							type="button"
							className="rounded-lg border px-4 py-2 text-sm font-medium hover:bg-raised/50"
							onClick={() => window.location.reload()}
						>
							Перезагрузить страницу
						</button>
					</div>
				</div>
			</div>
		);
	}
}
