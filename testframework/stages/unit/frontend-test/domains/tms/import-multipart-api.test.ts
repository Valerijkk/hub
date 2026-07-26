import { beforeEach, describe, expect, it, vi } from "vitest";
import { importTestCases } from "@/api/tms";
import { useAuthStore } from "@/stores/auth-store";

describe("tms importTestCases (A11-002)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		useAuthStore.setState({ accessToken: "test-token" } as never);
	});

	it("sends FormData via fetch without JSON Content-Type", async () => {
		const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
			new Response(JSON.stringify({ created: 1, skipped: 0, errors: [] }), {
				status: 200,
				headers: { "Content-Type": "application/json" },
			}),
		);

		const file = new File(["a,b\n1,2"], "cases.csv", { type: "text/csv" });
		await importTestCases("proj-1", file, "csv");

		expect(fetchSpy).toHaveBeenCalledTimes(1);
		const [url, init] = fetchSpy.mock.calls[0]!;
		expect(url).toBe("/api/projects/proj-1/tms/import");
		expect(init?.method).toBe("POST");
		expect(init?.body).toBeInstanceOf(FormData);
		const headers = init?.headers as Record<string, string>;
		expect(headers.Authorization).toBe("Bearer test-token");
		expect(headers["Content-Type"]).toBeUndefined();
	});
});
