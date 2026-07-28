import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, expectTypeOf, it } from "vitest";
import { plateDocToContentJson, type PlateDoc } from "@/lib/plate-doc";
import { formatApiError } from "@/lib/format-api-error";
import {
	isAdminStats,
	isPersonalData,
	isProjectStats,
	type DashboardWidgetData,
} from "@/components/dashboard/widget-data";
import { formatEnvConfigValue } from "@/lib/format-env-config";
import { listMarketplaceListings, listMyListings } from "@/api/marketplace";
import type { AdminAIProvider, MyAIProvider } from "@/api/ai-types";
import type { MarketplaceListing } from "@/api/marketplace";

const root = resolve(__dirname, "../../../../../../frontend/src");

function readSrc(rel: string): string {
	return readFileSync(resolve(root, rel), "utf8");
}

describe("P3-FE-07 type casts (TASK-074)", () => {
	it("plateDocToContentJson requires PlateDoc fields (omit type → TS fail)", () => {
		const doc: PlateDoc = {
			type: "doc",
			children: [{ type: "p", children: [{ text: "hi" }] }],
		};
		const json = plateDocToContentJson(doc);
		expect(json.type).toBe("doc");
		expect(Array.isArray(json.children)).toBe(true);
		expectTypeOf(plateDocToContentJson).parameter(0).toMatchTypeOf<PlateDoc>();
	});

	it("message-composer has no as unknown as Record double-cast", () => {
		const src = readSrc("components/messenger/message-composer.tsx");
		expect(src).not.toMatch(/as unknown as Record<string, unknown>/);
		expect(src).toMatch(/plateDocToContentJson/);
	});

	it("extractApiError does not treat plain Error as ApiError", () => {
		expect(formatApiError(new Error("Network failed"))).toBe("Network failed");
		expect(
			formatApiError({ message: "API boom", code: 500, request_id: "r1" }),
		).toBe("API boom");
		const src = readSrc("lib/format-api-error.ts");
		expect(src).toMatch(/typeof candidate\.code === "number"/);
		expect(src).toMatch(/request_id/);
	});

	it("marketplace list calls use explicit generics", () => {
		expectTypeOf(listMarketplaceListings).returns.toEqualTypeOf<
			Promise<{ items: MarketplaceListing[]; total: number }>
		>();
		expectTypeOf(listMyListings).returns.toEqualTypeOf<
			Promise<{ items: MarketplaceListing[] }>
		>();
		const src = readSrc("api/marketplace.ts");
		expect(src).toMatch(/api\.get<\{ items: MarketplaceListing/);
	});

	it("AdminAIProvider extends base; MyAIProvider is distinct shape", () => {
		const admin: AdminAIProvider = {
			id: "1",
			name: "a",
			provider_type: "openai",
			base_url: "https://x",
			models: [],
			enabled: true,
			ru_resident: false,
			scope: "org",
			has_api_key: true,
			priority: 1,
			max_rpm: 60,
			created_at: "",
			updated_at: "",
		};
		const mine: MyAIProvider = {
			id: "2",
			name: "b",
			provider_type: "openai",
			base_url: "https://y",
			models: [],
			enabled: true,
			ru_resident: false,
			scope: "user",
			has_api_key: false,
		};
		expect(admin.priority).toBe(1);
		expect("priority" in mine).toBe(false);
	});

	it("widget data guards narrow DashboardWidgetData", () => {
		const personal = {
			my_tasks: [],
			unread_count: 0,
			upcoming_calls: [],
			recent_runs: [],
			deadlines: [],
			time_today: 0,
		} satisfies DashboardWidgetData;
		expect(isPersonalData(personal)).toBe(true);
		expect(isProjectStats(personal)).toBe(false);
		expect(isAdminStats(personal)).toBe(false);
	});

	it("formatEnvConfigValue avoids [object Object]", () => {
		expect(formatEnvConfigValue("chrome")).toBe("chrome");
		expect(formatEnvConfigValue({ browser: "ff" })).toBe('{"browser":"ff"}');
		expect(formatEnvConfigValue(null)).toBe("");
	});

	it("sticker/render avoid bare children as unknown[] cast", () => {
		const sticker = readSrc("components/messenger/sticker-picker.tsx");
		expect(sticker).not.toMatch(/children as unknown\[\]/);
		expect(sticker).toMatch(/Array\.isArray\(children\)/);

		const render = readSrc("components/messenger/render-content.tsx");
		expect(render).toMatch(/const nodes: unknown\[\] = contentJson/);
	});
});
