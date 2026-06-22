import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ErrorBoundary } from "@/components/ui/error-boundary";

function BrokenChild(): never {
	throw new Error("render boom");
}

describe("ErrorBoundary", () => {
	it("renders children when there is no error", () => {
		render(
			<ErrorBoundary>
				<p>Healthy content</p>
			</ErrorBoundary>,
		);

		expect(screen.getByText("Healthy content")).toBeInTheDocument();
	});

	it("catches render errors and shows default fallback", () => {
		const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

		render(
			<ErrorBoundary>
				<BrokenChild />
			</ErrorBoundary>,
		);

		expect(screen.getByText("Что-то пошло не так")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Обновить" }),
		).toBeInTheDocument();

		consoleSpy.mockRestore();
	});

	it("renders custom fallback when provided", () => {
		const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

		render(
			<ErrorBoundary fallback={<div>Custom fallback</div>}>
				<BrokenChild />
			</ErrorBoundary>,
		);

		expect(screen.getByText("Custom fallback")).toBeInTheDocument();

		consoleSpy.mockRestore();
	});

	it("reload button triggers window.location.reload", () => {
		const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
		const reloadSpy = vi
			.spyOn(window.location, "reload")
			.mockImplementation(() => {});

		render(
			<ErrorBoundary>
				<BrokenChild />
			</ErrorBoundary>,
		);

		fireEvent.click(screen.getByRole("button", { name: "Обновить" }));

		expect(reloadSpy).toHaveBeenCalledOnce();

		reloadSpy.mockRestore();
		consoleSpy.mockRestore();
	});
});
