import { beforeEach, describe, expect, it, vi } from "vitest";
import { deleteBase, updateBase } from "@/api/bases";
import { api } from "@/api/client";

describe("bases update/delete API (A11-010)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("updateBase PUTs /bases/:baseId", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({} as never);

		await updateBase("b-1", { name: "Renamed", version: 3 });

		expect(putSpy).toHaveBeenCalledWith("/bases/b-1", {
			name: "Renamed",
			version: 3,
		});
	});

	it("deleteBase DELETEs /bases/:baseId", async () => {
		const delSpy = vi.spyOn(api, "delete").mockResolvedValue(undefined);

		await deleteBase("b-1");

		expect(delSpy).toHaveBeenCalledWith("/bases/b-1");
	});
});
