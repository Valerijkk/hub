import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { CI_CASES_FETCH_LIMIT } from "@/modules/tms/trigger-ci-dialog";

describe("shell/tracker UX (P2-FE-13/14 / TASK-055)", () => {
	it("sidebar-user does not claim fake Online", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/components/layout/sidebar-user.tsx",
			),
			"utf8",
		);
		expect(src).not.toMatch(/>Онлайн</);
		expect(src).not.toMatch(/bg-emerald-400/);
	});

	it("issue soft-delete copy mentions корзина 30 дней", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/routes/_authenticated/projects/$projectId/issues/$issueKey.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/корзину на 30 дней/);
		expect(src).not.toMatch(/нельзя\s*\n?\s*отменить/);
	});

	it("flaky queryKey uses tms-flaky prefix", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/modules/tms/components/flaky-tests-report.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/\["tms-flaky", projectId\]/);
		expect(src).not.toMatch(/\["flaky-tests"/);
	});

	it("CI cases fetch limit is explicit (500)", () => {
		expect(CI_CASES_FETCH_LIMIT).toBe(500);
	});

	it("app-shell closes mobile menu on pathname change", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/components/layout/app-shell.tsx",
			),
			"utf8",
		);
		expect(src).toMatch(/setMobileMenuOpen\(false\)/);
		expect(src).toMatch(/pathname/);
		expect(src).toMatch(/Escape/);
	});
});
