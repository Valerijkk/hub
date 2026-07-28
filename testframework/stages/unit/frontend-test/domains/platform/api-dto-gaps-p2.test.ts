import { describe, expect, expectTypeOf, it } from "vitest";
import { execFileSync } from "node:child_process";
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

describe("API DTO gaps (P2-FE-08 / TASK-052)", () => {
	it("updateBase/updateRecord require version (mutation: omit → TS fail)", () => {
		expectTypeOf(updateBase).parameter(1).toMatchTypeOf<{
			version: number;
		}>();
		expectTypeOf(updateRecord).parameter(3).toMatchTypeOf<{
			version: number;
		}>();
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
	});

	it("parsePlateValue accepts plate nodes and rejects garbage", () => {
		expect(parsePlateValue(null)).toBeUndefined();
		expect(parsePlateValue("x")).toBeUndefined();
		expect(parsePlateValue([{ type: "p", children: [{ text: "" }] }])).toEqual([
			{ type: "p", children: [{ text: "" }] },
		]);
	});
});
