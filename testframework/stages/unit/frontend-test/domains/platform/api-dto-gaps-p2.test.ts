import { describe, expect, expectTypeOf, it } from "vitest";
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { updateBase, updateRecord } from "@/api/bases";
import type { ActiveCallSession, GuestCallSession } from "@/api/calls";
import { buildMessageFromWsEvent } from "@/api/messenger";
import {
	createEnvironment,
	deleteEnvironment,
	listEnvironments,
} from "@/api/tms";
import { updateNews } from "@/api/intranet";
import { updateMemberRole } from "@/api/projects";
import { parsePlateValue } from "@/modules/tracker/issue-description";

const here = path.dirname(fileURLToPath(import.meta.url));
const frontendSrc = path.resolve(here, "../../../../../../frontend/src");

function readSrc(rel: string): string {
	return readFileSync(path.resolve(frontendSrc, rel), "utf8");
}

describe("API DTO gaps (P2-FE-08 / TASK-052)", () => {
	it("updateBase/updateRecord require version (mutation: omit → TS fail)", () => {
		expectTypeOf(updateBase).parameter(1).toMatchTypeOf<{
			version: number;
		}>();
		expectTypeOf(updateRecord).parameter(3).toMatchTypeOf<{
			version: number;
		}>();
		const src = readSrc("api/bases.ts");
		expect(src).toMatch(
			/updateBase\([\s\S]*?data:\s*\{[^}]*version:\s*number/,
		);
		expect(src).toMatch(
			/updateRecord\([\s\S]*?data:\s*\{[^}]*version:\s*number/,
		);
		expect(src).not.toMatch(
			/updateBase\([\s\S]*?data:\s*\{[^}]*version\?:\s*number/,
		);
	});

	it("tsc fixture: omit version is @ts-expect-error (mutation optional → unused)", () => {
		const tsconfig = path.join(here, "tsconfig.task052.json");
		const tscBin = path.resolve(
			here,
			"../../../../../../frontend/node_modules/typescript/bin/tsc",
		);
		execFileSync(process.execPath, [tscBin, "-p", tsconfig, "--pretty", "false"], {
			cwd: path.resolve(here, "../../../../../../frontend"),
			stdio: "pipe",
			encoding: "utf8",
		});
		expect(true).toBe(true);
	});

	it("exposes create/delete environment and updateNews/updateMemberRole", () => {
		expect(typeof listEnvironments).toBe("function");
		expect(typeof createEnvironment).toBe("function");
		expect(typeof deleteEnvironment).toBe("function");
		expect(typeof updateNews).toBe("function");
		expect(typeof updateMemberRole).toBe("function");

		const tms = readSrc("api/tms.ts");
		expect(tms).toMatch(/export function createEnvironment\b/);
		expect(tms).toMatch(/export function deleteEnvironment\b/);
		expect(tms).toMatch(
			/api\.post<EnvironmentResponse>\(\s*`\/projects\/\$\{projectId\}\/environments`/,
		);
		expect(tms).toMatch(
			/api\.delete<void>\(\s*`\/projects\/\$\{projectId\}\/environments\/\$\{envId\}`/,
		);

		const intranet = readSrc("api/intranet.ts");
		expect(intranet).toMatch(/export function updateNews\b/);
		expect(intranet).toMatch(/version:\s*number/);
		expect(intranet).toMatch(/api\.put<NewsPost>\(`\/intranet\/news\/\$\{id\}`/);

		const projects = readSrc("api/projects.ts");
		expect(projects).toMatch(/export function updateMemberRole\b/);
		expect(projects).toMatch(
			/`\/projects\/\$\{projectId\}\/members\/\$\{userId\}`/,
		);
	});

	it("buildMessageFromWsEvent rejects null/empty/missing sender", () => {
		expect(buildMessageFromWsEvent(null, "ch-1")).toBeNull();
		expect(buildMessageFromWsEvent({}, "ch-1")).toBeNull();
		expect(
			buildMessageFromWsEvent({ id: "m1", sender_id: "" }, "ch-1"),
		).toBeNull();
		const msg = buildMessageFromWsEvent(
			{ id: "m1", sender_id: "u1", content_text: "hi" },
			"ch-1",
		);
		expect(msg?.id).toBe("m1");
		expect(msg?.sender_id).toBe("u1");
		expect(msg?.channel_id).toBe("ch-1");

		const src = readSrc("api/messenger.ts");
		expect(src).toMatch(
			/export function buildMessageFromWsEvent\(\s*evt:\s*unknown/,
		);
		expect(src).toMatch(/if\s*\(!evt\s*\|\|\s*typeof evt !== "object"\)\s*return null/);
		expect(src).toMatch(/if\s*\(!id\)\s*return null/);
		expect(src).toMatch(/if\s*\(!sender_id\)\s*return null/);
	});

	it("GuestCallSession is assignable to ActiveCallSession without CallResponse cast", () => {
		const guest: GuestCallSession = {
			id: "c1",
			room_name: "room",
			token: "tok",
			status: "active",
		};
		const session: ActiveCallSession = guest;
		expect(session.id).toBe("c1");
		expect("channel_id" in session).toBe(false);

		const calls = readSrc("api/calls.ts");
		expect(calls).toMatch(/export interface GuestCallSession\b/);
		expect(calls).toMatch(
			/export type ActiveCallSession\s*=\s*CallResponse\s*\|\s*GuestCallSession/,
		);

		const guestRoute = readSrc("routes/call.$code.lazy.tsx");
		expect(guestRoute).not.toMatch(/as\s+import\(["']@\/api\/calls["']\)\.CallResponse/);
		expect(guestRoute).not.toMatch(/as\s+CallResponse\b/);
	});

	it("parsePlateValue accepts plate nodes and rejects garbage", () => {
		expect(parsePlateValue(null)).toBeUndefined();
		expect(parsePlateValue("x")).toBeUndefined();
		expect(parsePlateValue([{ type: "p", children: [{ text: "" }] }])).toEqual([
			{ type: "p", children: [{ text: "" }] },
		]);

		const src = readSrc("modules/tracker/issue-description.tsx");
		expect(src).toMatch(/plateValueSchema\.safeParse/);
		expect(src).not.toMatch(
			/export function parsePlateValue[\s\S]*?return raw as Value/,
		);
	});

	it("user-picker disables members query without projectId", () => {
		const src = readSrc("components/tracker/user-picker.tsx");
		expect(src).toMatch(/enabled:\s*Boolean\(projectId\)/);
	});

	it("case-filters updateFilter uses Partial filters not Record unknown", () => {
		const src = readSrc("modules/tms/components/case-filters.tsx");
		expect(src).toMatch(
			/updateFilter:\s*\(patch:\s*Partial<CaseFiltersProps\["filters"\]>\)/,
		);
		expect(src).not.toMatch(
			/updateFilter:\s*\(patch:\s*Record<string,\s*unknown>\)/,
		);
	});

	it("environment-comparison uses formatApiError not message cast", () => {
		const src = readSrc("modules/tms/components/environment-comparison.tsx");
		expect(src).toMatch(/formatApiError\(compError\)/);
		expect(src).not.toMatch(
			/\(compError\s+as\s+\{\s*message\?:\s*string\s*\}\)\.message/,
		);
	});
});
