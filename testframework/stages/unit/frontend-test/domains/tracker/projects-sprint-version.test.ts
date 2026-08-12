import { beforeEach, describe, expect, expectTypeOf, it, vi } from "vitest";
import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { api } from "@/api/client";
import { updateProject, type ProjectResponse } from "@/api/projects";
import { updateSprint, type SprintResponse } from "@/api/sprint";
import { formatApiError } from "@/lib/format-api-error";

const here = path.dirname(fileURLToPath(import.meta.url));

describe("projects/sprint API version (TASK-130 / A3W11-2)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("updateProject requires version (mutation: omit → TS fail)", () => {
		expectTypeOf(updateProject).parameter(1).toMatchTypeOf<{
			version: number;
		}>();
	});

	it("updateSprint requires version (mutation: omit → TS fail)", () => {
		expectTypeOf(updateSprint).parameter(2).toMatchTypeOf<{
			version: number;
		}>();
	});

	it("ProjectResponse and SprintResponse expose version", () => {
		expectTypeOf<ProjectResponse>().toHaveProperty("version");
		expectTypeOf<ProjectResponse["version"]>().toEqualTypeOf<number>();
		expectTypeOf<SprintResponse>().toHaveProperty("version");
		expectTypeOf<SprintResponse["version"]>().toEqualTypeOf<number>();
	});

	it("tsc fixture: omit version is @ts-expect-error (mutation optional → unused)", () => {
		const tsconfig = path.join(here, "tsconfig.task130.json");
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

	it("updateProject PUT body includes version from caller", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({
			id: "p1",
			name: "Hub",
			key: "HUB",
			status: "active",
			member_count: 1,
			version: 2,
			created_at: "2026-01-01T00:00:00Z",
			updated_at: "2026-01-02T00:00:00Z",
		} satisfies ProjectResponse);

		await updateProject("proj-1", {
			name: "Hub",
			description: "desc",
			version: 1,
		});

		expect(putSpy).toHaveBeenCalledWith("/projects/proj-1", {
			name: "Hub",
			description: "desc",
			version: 1,
		});
	});

	it("updateSprint PUT body includes version from caller", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({
			id: "s1",
			name: "Sprint 1",
			status: "planning",
			start_date: "2026-07-01",
			end_date: "2026-07-14",
			issue_count: 0,
			version: 2,
			created_at: "2026-01-01T00:00:00Z",
			updated_at: "2026-01-02T00:00:00Z",
		} satisfies SprintResponse);

		await updateSprint("proj-1", "s1", { name: "Sprint 1", version: 1 });

		expect(putSpy).toHaveBeenCalledWith("/projects/proj-1/sprints/s1", {
			name: "Sprint 1",
			version: 1,
		});
	});

	it("stale optimistic lock: formatApiError surfaces 409 message", () => {
		expect(
			formatApiError({
				message: "version conflict",
				code: 409,
				request_id: "req-stale",
			}),
		).toBe("version conflict");
	});
});
