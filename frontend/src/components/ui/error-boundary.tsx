import type { ReactNode } from "react";
import { Component } from "react";
import { useTranslation } from "react-i18next";

import "@/i18n";

interface Props {
	children: ReactNode;
	fallback?: ReactNode;
	/** Смена ключа (например pathname) сбрасывает error state без hard reload. */
	resetKey?: string | number;
}

interface State {
	hasError: boolean;
}

function DefaultErrorFallback({ onSoftReset }: { onSoftReset: () => void }) {
	const { t } = useTranslation("common");
	return (
		<div
			role="alert"
			className="flex flex-col items-center justify-center gap-3 p-8 text-center"
		>
			<p className="text-[15px] font-semibold text-t1">
				{t("error.somethingWrong")}
			</p>
			<p className="text-[13px] text-t3 max-w-md">{t("error.renderFailed")}</p>
			<div className="flex gap-2">
				<button
					type="button"
					onClick={onSoftReset}
					className="rounded-[var(--r2)] bg-primary px-4 py-2 text-[13px] font-medium text-primary-foreground hover:bg-primary/90"
				>
					{t("error.tryAgain")}
				</button>
				<button
					type="button"
					onClick={() => window.location.reload()}
					className="rounded-[var(--r2)] border border-b1 px-4 py-2 text-[13px] text-t2 hover:text-t1"
				>
					{t("btn.refresh")}
				</button>
			</div>
		</div>
	);
}

/**
 * ErrorBoundary — перехватывает ошибки рендера дочерних компонентов (FE-H4).
 * Предотвращает белый экран при падении отдельного модуля.
 * Сброс: soft-retry или смена `resetKey` (навигация sidebar → Outlet).
 */
export class ErrorBoundary extends Component<Props, State> {
	state: State = { hasError: false };

	static getDerivedStateFromError(): State {
		return { hasError: true };
	}

	componentDidCatch(error: Error): void {
		console.error("UI error boundary:", error);
	}

	componentDidUpdate(prevProps: Props): void {
		if (
			this.state.hasError &&
			prevProps.resetKey !== this.props.resetKey &&
			this.props.resetKey !== undefined
		) {
			this.setState({ hasError: false });
		}
	}

	private softReset = (): void => {
		this.setState({ hasError: false });
	};

	render() {
		if (this.state.hasError) {
			return (
				this.props.fallback ?? (
					<DefaultErrorFallback onSoftReset={this.softReset} />
				)
			);
		}
		return this.props.children;
	}
}
