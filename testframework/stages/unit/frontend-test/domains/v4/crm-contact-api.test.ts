import { beforeEach, describe, expect, it, vi } from "vitest";
import { deleteContact, updateContact } from "@/api/crm";
import { api } from "@/api/client";

describe("crm contact update/delete API (A11-009)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("updateContact PUTs /crm/contacts/:id with version", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({} as never);

		await updateContact("c-1", { name: "Ada", version: 2 });

		expect(putSpy).toHaveBeenCalledWith("/crm/contacts/c-1", {
			name: "Ada",
			version: 2,
		});
	});

	it("deleteContact DELETEs /crm/contacts/:id", async () => {
		const delSpy = vi.spyOn(api, "delete").mockResolvedValue(undefined);

		await deleteContact("c-1");

		expect(delSpy).toHaveBeenCalledWith("/crm/contacts/c-1");
	});
});
