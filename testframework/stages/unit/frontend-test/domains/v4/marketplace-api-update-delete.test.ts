import { beforeEach, describe, expect, it, vi } from "vitest";
import { deleteListing, updateListing } from "@/api/marketplace";
import { api } from "@/api/client";

describe("marketplace listing update/delete API (A11-011)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("updateListing PUTs /marketplace/listings/:id", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({} as never);

		await updateListing("l-1", { title: "New", version: 1 });

		expect(putSpy).toHaveBeenCalledWith("/marketplace/listings/l-1", {
			title: "New",
			version: 1,
		});
	});

	it("deleteListing DELETEs /marketplace/listings/:id", async () => {
		const delSpy = vi.spyOn(api, "delete").mockResolvedValue(undefined);

		await deleteListing("l-1");

		expect(delSpy).toHaveBeenCalledWith("/marketplace/listings/l-1");
	});
});
