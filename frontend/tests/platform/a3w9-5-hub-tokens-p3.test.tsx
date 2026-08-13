import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Avatar } from "@/components/ui/avatar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectSeparator,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import "@/i18n";

/**
 * A3W9-5 / TASK-176 — avatar online-dot, PopoverContent, SelectItem/Separator hub tokens.
 * Mutation: revert className to emerald/popover/accent/muted → asserts FAIL.
 */
describe("A3W9-5 · hub tokens avatar/popover/select (TASK-176)", () => {
	it("A3W9-5-001 Avatar online-dot uses var(--green), not emerald-500", () => {
		const { container } = render(<Avatar data-online="true" />);
		const dot = container.querySelector("[aria-hidden='true']");
		expect(dot?.className).toMatch(/bg-\[var\(--green\)\]/);
		expect(dot?.className).not.toMatch(/bg-emerald-500/);
	});

	it("A3W9-5-002 PopoverContent uses hub floating-surface tokens", () => {
		render(
			<Popover open>
				<PopoverTrigger asChild>
					<button type="button">Открыть</button>
				</PopoverTrigger>
				<PopoverContent>Контент</PopoverContent>
			</Popover>,
		);
		const content = screen.getByText("Контент");
		expect(content.className).toMatch(/rounded-\[var\(--r2\)\]/);
		expect(content.className).toMatch(/border-b1/);
		expect(content.className).toMatch(/\bbg-base\b/);
		expect(content.className).toMatch(/\btext-t1\b/);
		expect(content.className).not.toMatch(/bg-popover/);
		expect(content.className).not.toMatch(/\brounded-md\b/);
		expect(content.className).not.toMatch(/text-popover-foreground/);
	});

	it("A3W9-5-003 SelectItem/Separator align with Dropdown hub tokens", () => {
		const { container } = render(
			<Select open>
				<SelectTrigger aria-label="Тест">
					<SelectValue placeholder="Выбор" />
				</SelectTrigger>
				<SelectContent>
					<SelectItem value="a">Вариант A</SelectItem>
					<SelectSeparator data-testid="select-sep" />
					<SelectItem value="b">Вариант B</SelectItem>
				</SelectContent>
			</Select>,
		);
		const item = screen.getByRole("option", { name: "Вариант A" });
		expect(item.className).toMatch(/rounded-\[var\(--r1\)\]/);
		expect(item.className).toMatch(/focus:bg-\[var\(--hover\)\]/);
		expect(item.className).toMatch(/focus:text-t1/);
		expect(item.className).not.toMatch(/focus:bg-accent/);
		expect(item.className).not.toMatch(/\brounded-sm\b/);

		const sep =
			screen.queryByTestId("select-sep") ??
			container.ownerDocument.querySelector("[data-testid=select-sep]");
		expect(sep?.className).toMatch(/\bbg-b1\b/);
		expect(sep?.className).not.toMatch(/\bbg-muted\b/);
	});
});
