import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const repoRoot = resolve(__dirname, "../../../../../../");

function readDoc(rel: string): string {
	return readFileSync(resolve(repoRoot, rel), "utf8");
}

describe("docs drift NF-11 / REST / webcam (A2W11 / TASK-123)", () => {
	it("A2W11-4-001: BUSINESS.md webcam wording is consistent (shipped, not Phase-16 blocker)", () => {
		const src = readDoc("docs/BUSINESS.md");
		// voice row: webcam shipped, not «Phase 16» as unimplemented
		expect(src).toMatch(/webcam \(UI shipped/i);
		expect(src).toMatch(/CallControls/);
		expect(src).toMatch(/setCameraEnabled/);
		// Phase 16+ residual must not list webcam as planned deliverable alongside Gantt
		const phase16 = src.match(/\|\s*\*\*Phase 16\+\*\*\s*\|([^|]+)\|/);
		expect(phase16?.[1] ?? "").toMatch(/shipped/i);
		expect(phase16?.[1] ?? "").not.toMatch(/^\s*Gantt,\s*webcam,/);
	});

	it("A2W11-5-003: NF-11 evidence points to wiring_fiber.go", () => {
		const src = readDoc("docs/security_guidelines.md");
		expect(src).toMatch(
			/\|\s*NF-11\s*\|[^|]*\|\s*✅\s*\*\*FIXED\*\*\s*\|\s*`cmd\/server\/wiring_fiber\.go`/,
		);
		expect(src).not.toMatch(
			/\|\s*NF-11\s*\|[^|]*\|\s*✅\s*\*\*FIXED\*\*\s*\|\s*`cmd\/server\/main\.go`/,
		);
	});

	it("A2W11-5-004: system_specification summary uses OpenAPI ~252 methodology", () => {
		const src = readDoc("docs/system_specification.md");
		// summary table row for API module
		expect(src).toMatch(/REST\s*\*\*~?252\*\*/);
		expect(src).not.toMatch(/\| API \| API \| REST \*\*286\*\* endpoints/);
	});
});
