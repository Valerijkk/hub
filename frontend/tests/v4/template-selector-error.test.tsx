import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { TemplateSelector } from "@/modules/settings/project-templates";

/** A3W10-1-002: TemplateSelector fail ≠ empty grid. */
describe("TemplateSelector error≠empty (A3W10-1-002)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows QueryErrorAlert when /project-templates fails", async () => {
		vi.spyOn(api, "get").mockRejectedValue(new Error("network"));

		render(<TemplateSelector onSelect={vi.fn()} />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить шаблоны/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByRole("button", { name: /системный/i })).not.toBeInTheDocument();
	});

	it("renders template cards when load succeeds", async () => {
		vi.spyOn(api, "get").mockResolvedValue([
			{
				id: "t1",
				name: "Empty",
				template_type: "empty",
				config: {},
				is_system: true,
				created_at: "2026-01-01T00:00:00Z",
			},
		]);

		render(<TemplateSelector onSelect={vi.fn()} />);

		await waitFor(() => {
			expect(screen.getByText("Empty")).toBeInTheDocument();
		});
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});
});
