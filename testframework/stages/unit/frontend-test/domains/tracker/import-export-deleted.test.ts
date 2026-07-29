import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

describe("FE · ImportExportToolbar removed (A3W10-2-004)", () => {
	it("import-export.tsx module is deleted (dead export cleanup)", () => {
		const here = dirname(fileURLToPath(import.meta.url));
		const modulePath = resolve(
			here,
			"../../../../../../frontend/src/modules/tracker/import-export.tsx",
		);
		expect(existsSync(modulePath)).toBe(false);

		// Mutation: вернуть файл import-export.tsx → existsSync true → FAIL.
	});
});
