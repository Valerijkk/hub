import { Inbox } from "lucide-react";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TrashManager } from "@/components/admin/trash-manager";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { renderWithProviders } from "../../render";

vi.mock("@/api/admin", () => ({
	getTrash: vi.fn().mockResolvedValue({ items: [], total: 0 }),
	restoreItem: vi.fn(),
}));

/**
 * P3-FE-02 — a11y structure residual.
 * Mutation: drop tablist/heading/aria-hidden/aria-required → role queries fail.
 */
describe("P3-FE-02 · a11y structure residual (TASK-070)", () => {
	it("TrashManager tabs use tablist/tab/aria-selected", async () => {
		renderWithProviders(<TrashManager />);
		await waitFor(() => {
			expect(screen.getByRole("tablist")).toBeInTheDocument();
		});
		const issuesTab = screen.getByRole("tab", { name: "Задачи" });
		expect(issuesTab).toHaveAttribute("aria-selected", "true");
	});

	it("RouteErrorBoundary fallback has role=alert", () => {
		const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
		function Boom(): never {
			throw new Error("route boom");
		}
		render(
			<RouteErrorBoundary module="Тест">
				<Boom />
			</RouteErrorBoundary>,
		);
		expect(screen.getByRole("alert")).toBeInTheDocument();
		consoleSpy.mockRestore();
	});

	it("CardTitle renders as heading", () => {
		render(
			<Card>
				<CardHeader>
					<CardTitle>Заголовок карточки</CardTitle>
				</CardHeader>
			</Card>,
		);
		expect(
			screen.getByRole("heading", { name: "Заголовок карточки" }),
		).toBeInTheDocument();
	});

	it("EmptyState decorative icon is aria-hidden", () => {
		const { container } = render(
			<EmptyState icon={Inbox} title="Пусто" description="Нет элементов" />,
		);
		const iconWrap = container.querySelector("[aria-hidden='true']");
		expect(iconWrap).toBeTruthy();
	});

	it("Avatar announces online status when data-online", async () => {
		const { default: i18n } = await import("@/i18n");
		await i18n.changeLanguage("ru");
		render(<Avatar data-online="true" />);
		expect(screen.getByText("В сети")).toBeInTheDocument();
	});

	it("Badge renders as span without focus ring classes", () => {
		const { container } = render(<Badge>Active</Badge>);
		const el = container.firstElementChild;
		expect(el?.tagName.toLowerCase()).toBe("span");
		expect(el?.className).not.toMatch(/focus:ring/);
	});

	it("FormField required sets aria-required on control", () => {
		renderWithProviders(
			<FormField label="Имя" htmlFor="name-field" required>
				<Input id="name-field" />
			</FormField>,
		);
		expect(screen.getByLabelText(/Имя/)).toHaveAttribute(
			"aria-required",
			"true",
		);
	});
});
