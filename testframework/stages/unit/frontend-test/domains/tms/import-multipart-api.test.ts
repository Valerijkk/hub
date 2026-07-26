import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { importTestCases } from "@/api/tms";
import { useAuthStore } from "@/stores/auth-store";

describe("tms importTestCases (A11-002)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		useAuthStore.setState({ accessToken: "test-token" } as never);
	});

	it("sends FormData via api.post without JSON Content-Type", async () => {
		const fetchSpy = vi.spyOn(globalThis, "fetch").mockImplementation(async () =>
			new Response(JSON.stringify({ created: 1, skipped: 0, errors: [] }), {
				status: 200,
				headers: { "Content-Type": "application/json" },
			}),
		);

		const file = new File(["a,b\n1,2"], "cases.csv", { type: "text/csv" });
		const formDataProbe = new FormData();
		formDataProbe.append("probe", "1");

		// Прямой контракт client: FormData доходит до fetch неизменной.
		await api.post("/projects/proj-1/tms/import", formDataProbe);
		expect(fetchSpy).toHaveBeenCalledTimes(1);
		const [, probeInit] = fetchSpy.mock.calls[0]!;
		expect(probeInit?.body).toBe(formDataProbe);
		expect((probeInit?.headers as Record<string, string>)["Content-Type"]).toBeUndefined();

		fetchSpy.mockClear();
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
