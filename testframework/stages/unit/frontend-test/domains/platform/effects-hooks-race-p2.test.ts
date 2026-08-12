import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

/** Guard: copy/toast timers must clear on unmount (P2-FE-07). */
describe("timer cleanup guards (P2-FE-07 / TASK-051)", () => {
	it("gif-picker clears debounceRef on unmount", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/components/messenger/gif-picker.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/clearTimeout\(debounceRef\.current\)/);
		expect(src).toMatch(/useEffect\(\s*\(\)\s*=>\s*\(\)\s*=>/);
	});

	it("resource-limits starts loading=true and clears toast timer", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/components/admin/resource-limits.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/useState\(true\)/);
		expect(src).toMatch(/toastTimerRef/);
		expect(src).toMatch(/clearTimeout\(toastTimerRef\.current\)/);
	});

	it("profile syncs form in useEffect, not during render", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/routes/_authenticated/settings/profile.lazy.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/useEffect\(\(\)\s*=>\s*\{[\s\S]*setForm\(/);
		expect(src).not.toMatch(
			/const \[initialized[\s\S]*if \(user && !initialized\) \{\s*setForm/,
		);
	});
});
