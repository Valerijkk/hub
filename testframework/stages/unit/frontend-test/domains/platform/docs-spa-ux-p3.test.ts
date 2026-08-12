import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const root = (...parts: string[]) =>
	resolve(__dirname, "../../../../../../", ...parts);

describe("TASK-078 P3-FE-13..15 docs / SPA / preview URLs", () => {
	it("architecture.md pins route count (not ~60+)", () => {
		const src = readFileSync(root("docs/architecture.md"), "utf8");
		expect(src).not.toMatch(/~60\+/);
		expect(src).toMatch(/\*\*7[0-9]\*\*\s*`\.tsx`\s*routes/);
	});

	it("stack_and_tools.md pins 12 stores and current motion/pwa", () => {
		const src = readFileSync(root("docs/stack_and_tools.md"), "utf8");
		expect(src).toMatch(/\*\*12\*\*\s*stores/);
		expect(src).not.toMatch(/11\+/);
		expect(src).toMatch(/motion.*\*\*\^12\.40\.0\*\*/);
		expect(src).toMatch(/vite-plugin-pwa.*\*\*\^1\.3\.0\*\*/);
	});

	it("API.md magic-link curl uses Caddy :8080", () => {
		const src = readFileSync(root("docs/wiki/API.md"), "utf8");
		expect(src).toMatch(/localhost:8080\/api\/auth\/magic-link/);
		expect(src).toMatch(/Caddy/);
	});

	it("security settings uses TanStack Link, not <a href>", () => {
		const src = readFileSync(
			root("frontend/src/routes/_authenticated/settings/security.tsx"),
			"utf8",
		);
		expect(src).toMatch(/<Link\s*\n?\s*to="\/settings\/profile"/);
		expect(src).not.toMatch(/href="\/settings\/profile"/);
	});

	it("trash column is Удалено (date), not Удалил", () => {
		const src = readFileSync(
			root("frontend/src/routes/_authenticated/trash.tsx"),
			"utf8",
		);
		expect(src).toMatch(/Удалено/);
		expect(src).not.toMatch(/>\s*Удалил\s*</);
	});

	it("run-results-table has correct UTF-8 Исполнитель", () => {
		const src = readFileSync(
			root("frontend/src/modules/tms/components/run-results-table.tsx"),
			"utf8",
		);
		expect(src).toMatch(/Исполнитель/);
		expect(src).not.toMatch(/Испол��итель/);
	});

	it("message-bubble extracts URLs from displayText and contentText", () => {
		const src = readFileSync(
			root("frontend/src/modules/messenger/message-bubble.tsx"),
			"utf8",
		);
		expect(src).toMatch(/extractUrls\(displayText\)/);
		expect(src).toMatch(/extractUrls\(contentText\)/);
	});
});
