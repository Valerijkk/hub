import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const fe = (...parts: string[]) =>
	resolve(__dirname, "../../../../../../frontend/src", ...parts);

describe("TASK-076 P3-FE-09 perf / stale clock / Set hoist", () => {
	it("pipeline refetchInterval only when queued/in_progress", () => {
		const src = readFileSync(fe("modules/tracker/pipeline-status.tsx"), "utf8");
		expect(src).toMatch(/refetchInterval:\s*\(query\)/);
		expect(src).toMatch(/status === "queued"/);
		expect(src).toMatch(/status === "in_progress"/);
		expect(src).not.toMatch(/refetchInterval:\s*30_000/);
	});

	it("dashboard clock updates via interval, not frozen useMemo", () => {
		const src = readFileSync(
			fe("routes/_authenticated/dashboard.lazy.tsx"),
			"utf8",
		);
		expect(src).toMatch(/setCurrentTime\(formatTime\(\)\)/);
		expect(src).toMatch(/setInterval/);
		expect(src).not.toMatch(/useMemo\(\(\)\s*=>\s*formatTime\(\),\s*\[\]\)/);
	});

	it("run add-cases dialog hoists existingIds Set outside map", () => {
		const src = readFileSync(
			fe("routes/_authenticated/projects/$projectId/tms/runs_.$runId.tsx"),
			"utf8",
		);
		expect(src).toMatch(/const existingIds = new Set/);
		// Set must not be created inside .map callback body as first statement pattern
		const mapIdx = src.indexOf("(casesData?.items ?? []).map");
		expect(mapIdx).toBeGreaterThan(-1);
		const beforeMap = src.slice(0, mapIdx);
		expect(beforeMap).toMatch(/const existingIds = new Set/);
	});
});
