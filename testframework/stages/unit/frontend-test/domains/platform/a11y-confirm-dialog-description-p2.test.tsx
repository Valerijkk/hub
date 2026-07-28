import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

/**
 * P2-FE-04 / W9-A5-004 — ConfirmDialog always exposes Description.
 * Mutation: drop Description when description omitted → no alertdialog description.
 */
describe("P2-FE-04 · ConfirmDialog Description (TASK-049)", () => {
	it("renders AlertDialogDescription when description is omitted", () => {
		render(
			<ConfirmDialog
				open
				onOpenChange={() => undefined}
				title="Удалить?"
				onConfirm={() => undefined}
			/>,
		);

		const dialog = screen.getByRole("alertdialog");
		expect(dialog).toBeInTheDocument();
		// Radix Description always present (sr-only fallback uses title).
		expect(dialog.getAttribute("aria-describedby")).toBeTruthy();
		expect(document.querySelector(".sr-only")).toHaveTextContent("Удалить?");
	});

	it("renders provided description text", () => {
		render(
			<ConfirmDialog
				open
				onOpenChange={() => undefined}
				title="Удалить?"
				description="Действие необратимо"
				onConfirm={() => undefined}
			/>,
		);

		expect(screen.getByText("Действие необратимо")).toBeInTheDocument();
	});
});
