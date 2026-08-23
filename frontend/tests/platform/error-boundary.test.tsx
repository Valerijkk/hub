import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import i18n from "@/i18n";

function BrokenChild(): never {
	throw new Error("render boom");
}

describe("ErrorBoundary", () => {
	beforeEach(async () => {
		await i18n.changeLanguage("ru");
	});

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

		expect(
			screen.getByText(i18n.t("error.somethingWrong")),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: i18n.t("btn.refresh") }),
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

});
