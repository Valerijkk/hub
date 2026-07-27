import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	applyDashboardTemplate,
	listDashboardTemplates,
} from "@/api/dashboard";
import { api } from "@/api/client";

describe("dashboard templates API (A11-012)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("listDashboardTemplates GETs /dashboard/templates", async () => {
		const getSpy = vi.spyOn(api, "get").mockResolvedValue({
			scrum: [],
		});

		const result = await listDashboardTemplates();

		expect(getSpy).toHaveBeenCalledWith("/dashboard/templates");
		expect(result).toEqual({ scrum: [] });
	});

	it("applyDashboardTemplate POSTs project_id + template", async () => {
		const postSpy = vi.spyOn(api, "post").mockResolvedValue(undefined);

		await applyDashboardTemplate("proj-1", "scrum");

		expect(postSpy).toHaveBeenCalledWith("/dashboard/templates/apply", {
			project_id: "proj-1",
			template: "scrum",
		});
	});
});
