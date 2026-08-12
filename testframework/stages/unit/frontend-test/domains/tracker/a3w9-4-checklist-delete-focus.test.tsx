import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Checklist } from "@/components/tracker/checklist";

describe("Checklist delete focus (A3W9-4-001 / TASK-175)", () => {
	it("delete control is revealed on group-focus-within / focus-visible without hover", () => {
		render(
			<Checklist
				items={[{ id: "c1", text: "Step 1", checked: false }]}
				onChange={vi.fn()}
			/>,
		);

		const del = screen.getByRole("button", { name: /удалить пункт/i });
		expect(del.className).toMatch(/group-focus-within:opacity-100/);
		expect(del.className).toMatch(/focus-visible:opacity-100/);
		expect(del.className).toMatch(/group-hover:opacity-100/);
		expect(del).toHaveAttribute("aria-label", "Удалить пункт");
	});
});
