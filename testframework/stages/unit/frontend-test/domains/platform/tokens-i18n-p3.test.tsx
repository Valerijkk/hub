import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
	CommandDialog,
	CommandInput,
} from "@/components/ui/command";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { Select, SelectTrigger } from "@/components/ui/select";
import i18n from "@/i18n";

function BrokenChild(): never {
	throw new Error("render boom");
}

const here = path.dirname(fileURLToPath(import.meta.url));
const frontendSrc = path.resolve(here, "../../../../../../frontend/src");

/**
 * P3-FE-03/04 — hub tokens + i18n in ui primitives.
 * Mutation: raw palette / hardcoded EN Close / RU-only EB → FAIL.
 */
describe("P3-FE-03/04 · tokens + i18n primitives (TASK-071)", () => {
	afterEach(async () => {
		await i18n.changeLanguage("ru");
	});

	it("SelectTrigger uses hub radius/border tokens (not shadcn rounded-md/border-input)", () => {
		const { container } = render(
			<Select>
				<SelectTrigger aria-label="Тест">
					<span>value</span>
				</SelectTrigger>
			</Select>,
		);
		const trigger = container.querySelector("button");
		expect(trigger?.className).toMatch(/rounded-\[var\(--r2\)\]/);
		expect(trigger?.className).toMatch(/border-b1/);
		expect(trigger?.className).not.toMatch(/\brounded-md\b/);
		expect(trigger?.className).not.toMatch(/border-input/);
	});

	it("messenger residual files have no raw Tailwind palette colors", () => {
		const files = [
			"components/messenger/voice-recorder.tsx",
			"components/messenger/poll-message.tsx",
			"components/messenger/channel-list.tsx",
			"components/messenger/thread-issue-links.tsx",
			"components/messenger/invite-link-dialog.tsx",
		];
		const palette =
			/\b(bg|text|hover:text|border)-(red|orange|yellow|green)-\d{2,3}\b/;
		for (const rel of files) {
			const src = readFileSync(path.resolve(frontendSrc, rel), "utf8");
			expect(src, rel).not.toMatch(palette);
		}
	});

	it("Dialog Close uses i18n btn.close (RU)", async () => {
		await i18n.changeLanguage("ru");
		render(
			<Dialog open>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Тест</DialogTitle>
					</DialogHeader>
				</DialogContent>
			</Dialog>,
		);
		expect(
			screen.getByRole("button", { name: "Закрыть" }),
		).toBeInTheDocument();
	});

	it("Dialog Close switches to EN Close via i18n", async () => {
		await i18n.changeLanguage("en");
		render(
			<Dialog open>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Test</DialogTitle>
					</DialogHeader>
				</DialogContent>
			</Dialog>,
		);
		expect(screen.getByRole("button", { name: "Close" })).toBeInTheDocument();
	});

	it("CommandDialog title follows locale", async () => {
		await i18n.changeLanguage("en");
		render(
			<CommandDialog open>
				<CommandInput placeholder="Search" />
			</CommandDialog>,
		);
		expect(screen.getByText("Global search")).toBeInTheDocument();
	});

	it("ErrorBoundary fallback uses i18n (EN locale)", async () => {
		const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
		await i18n.changeLanguage("en");
		render(
			<ErrorBoundary>
				<BrokenChild />
			</ErrorBoundary>,
		);
		expect(screen.getByText("Something went wrong")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Refresh" }),
		).toBeInTheDocument();
		consoleSpy.mockRestore();
	});
});
