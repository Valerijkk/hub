import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

/**
 * Guard P2-FE-07 / TASK-051: useState(activePackId) must run before any early throw,
 * otherwise Rules of Hooks break when STICKER_PACKS is non-empty (always in prod).
 */
describe("sticker-picker hooks order (P2-FE-07 / TASK-051)", () => {
	it("declares activePackId useState before STICKER_PACKS empty throw", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/components/messenger/sticker-picker.tsx",
			),
			"utf8",
		);
		const fnStart = src.indexOf("export function StickerPicker");
		expect(fnStart).toBeGreaterThanOrEqual(0);
		const body = src.slice(fnStart);
		const useStateIdx = body.search(
			/useState\(\s*\(\)\s*=>\s*STICKER_PACKS\[0\]/,
		);
		const throwIdx = body.indexOf(
			'throw new Error("STICKER_PACKS must not be empty")',
		);
		expect(useStateIdx).toBeGreaterThanOrEqual(0);
		expect(throwIdx).toBeGreaterThanOrEqual(0);
		expect(useStateIdx).toBeLessThan(throwIdx);
	});
});
