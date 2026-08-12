import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Button } from "@/components/ui/button";

/**
 * P2-FE-04 / W9-A5-002 — Button default type="button".
 * Mutation: remove default → bare Button inside <form> submits the form.
 */
describe("P2-FE-04 · Button default type (TASK-049)", () => {
	it("defaults to type=button and does not submit a surrounding form", () => {
		const onSubmit = vi.fn((e: React.FormEvent) => e.preventDefault());

		render(
			<form onSubmit={onSubmit}>
				<Button>Не сабмит</Button>
			</form>,
		);

		const btn = screen.getByRole("button", { name: "Не сабмит" });
		expect(btn).toHaveAttribute("type", "button");
		fireEvent.click(btn);
		expect(onSubmit).not.toHaveBeenCalled();
	});

	it("honours explicit type=submit for intentional form submit", () => {
		const onSubmit = vi.fn((e: React.FormEvent) => e.preventDefault());

		render(
			<form onSubmit={onSubmit}>
				<Button type="submit">Сохранить</Button>
			</form>,
		);

		const btn = screen.getByRole("button", { name: "Сохранить" });
		expect(btn).toHaveAttribute("type", "submit");
		fireEvent.click(btn);
		expect(onSubmit).toHaveBeenCalledTimes(1);
	});
});
