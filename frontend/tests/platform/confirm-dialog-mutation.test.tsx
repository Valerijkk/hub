import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

function Harness({
	onConfirm,
}: {
	onConfirm: () => void | Promise<void>;
}) {
	const [open, setOpen] = useState(true);
	return (
		<>
			<span data-testid="open-state">{open ? "open" : "closed"}</span>
			<ConfirmDialog
				open={open}
				onOpenChange={setOpen}
				title="Удалить?"
				description="Действие необратимо"
				confirmLabel="Удалить"
				onConfirm={onConfirm}
			/>
		</>
	);
}

describe("P2-FE-02 · ConfirmDialog close-on-reject (W9-A5-003)", () => {
	it("keeps dialog open when onConfirm rejects", async () => {
		const onConfirm = vi.fn().mockRejectedValue(new Error("boom"));

		render(<Harness onConfirm={onConfirm} />);

		expect(screen.getByTestId("open-state")).toHaveTextContent("open");
		expect(screen.getByRole("alertdialog")).toBeInTheDocument();

		fireEvent.click(screen.getByRole("button", { name: "Удалить" }));

		await waitFor(() => {
			expect(onConfirm).toHaveBeenCalledTimes(1);
		});

		// Mutation criterion: reject → dialog stays open (not closed in finally).
		expect(screen.getByTestId("open-state")).toHaveTextContent("open");
		expect(screen.getByRole("alertdialog")).toBeInTheDocument();
	});

	it("closes dialog when onConfirm resolves", async () => {
		const onConfirm = vi.fn().mockResolvedValue(undefined);

		render(<Harness onConfirm={onConfirm} />);

		fireEvent.click(screen.getByRole("button", { name: "Удалить" }));

		await waitFor(() => {
			expect(screen.getByTestId("open-state")).toHaveTextContent("closed");
		});
		expect(onConfirm).toHaveBeenCalledTimes(1);
	});
});
