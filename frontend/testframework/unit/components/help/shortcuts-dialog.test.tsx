import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import "@/i18n";
import { ShortcutsDialog } from "@/components/help/shortcuts-dialog";

describe("ShortcutsDialog", () => {
	it("renders shortcut groups when open", () => {
		const onOpenChange = vi.fn();
		render(<ShortcutsDialog open onOpenChange={onOpenChange} />);

		expect(screen.getByRole("dialog")).toBeInTheDocument();
		expect(screen.getByRole("heading", { level: 2 })).toBeInTheDocument();
	});
});
