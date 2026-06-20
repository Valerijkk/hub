import type { ReactNode } from "react";
import { Component } from "react";

interface Props {
	children: ReactNode;
	fallback?: ReactNode;
}

interface State {
	hasError: boolean;
}

/**
 * ErrorBoundary — перехватывает ошибки рендера дочерних компонентов (FE-H4).
 * Предотвращает белый экран при падении отдельного модуля.
 */
export class ErrorBoundary extends Component<Props, State> {
	state: State = { hasError: false };

	static getDerivedStateFromError(): State {
		return { hasError: true };
	}

	componentDidCatch(error: Error): void {
		console.error("UI error boundary:", error);
	}

	render() {
		if (this.state.hasError) {
			return (
				this.props.fallback ?? (
					<div className="flex flex-col items-center justify-center gap-3 p-8 text-center">
						<p className="text-[15px] font-semibold text-t1">
							Что-то пошло не так
						</p>
						<p className="text-[13px] text-t3 max-w-md">
							Произошла ошибка при отображении раздела. Попробуйте обновить
							страницу.
						</p>
						<button
							type="button"
							onClick={() => window.location.reload()}
							className="rounded-[var(--r2)] border border-b1 px-4 py-2 text-[13px] text-t2 hover:text-t1"
						>
							Обновить
						</button>
					</div>
				)
			);
		}
		return this.props.children;
	}
}
