import { beforeEach, describe, expect, it, vi } from "vitest";
import { joinByInvite } from "@/api/projects";
import { api } from "@/api/client";

describe("projects API joinByInvite (A11-001)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("posts invite code to /invites/join (not /projects/join)", async () => {
		const postSpy = vi.spyOn(api, "post").mockResolvedValue(undefined);

		await joinByInvite("invite-code-abc");

		expect(postSpy).toHaveBeenCalledWith("/invites/join", {
			code: "invite-code-abc",
		});
		expect(postSpy).not.toHaveBeenCalledWith(
			"/projects/join",
			expect.anything(),
		);
	});
});
