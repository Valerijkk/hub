import { beforeEach, describe, expect, it, vi } from "vitest";
import { exportTestCases, exportTestRunReport } from "@/api/tms";
import { useAuthStore } from "@/stores/auth-store";

function readBlobText(blob: Blob): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.addEventListener("load", () => {
			if (typeof reader.result === "string") {
				resolve(reader.result);
				return;
			}
			reject(new TypeError("Expected Blob text result"));
		});
		reader.addEventListener("error", () => {
			reject(reader.error ?? new Error("Failed to read Blob"));
		});
		reader.readAsText(blob);
	});
}

describe("tms export blob API (A11-007)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		useAuthStore.setState({ accessToken: "test-token" } as never);
	});

	it("exportTestCases returns CSV response as Blob", async () => {
		const csv = "a,b\n1,2\n";
		const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
			new Response(csv, {
				status: 200,
				headers: { "Content-Type": "text/csv" },
			}),
		);

		const blob = await exportTestCases("proj-1", "csv");

		expect(fetchSpy).toHaveBeenCalledWith(
			"/api/projects/proj-1/tms/export?format=csv",
			expect.objectContaining({ method: "GET" }),
		);
		const headers = fetchSpy.mock.calls[0]?.[1]?.headers as Record<
			string,
			string
		>;
		expect(headers.Authorization).toBe("Bearer test-token");
		expect(blob).toBeInstanceOf(Blob);
		expect(await readBlobText(blob)).toBe(csv);
	});

	it("exportTestRunReport uses blob transport for run export", async () => {
		const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
			new Response("ok", {
				status: 200,
				headers: { "Content-Type": "text/csv" },
			}),
		);

		const blob = await exportTestRunReport("proj-1", "run-9", "csv");

		expect(fetchSpy).toHaveBeenCalledWith(
			"/api/projects/proj-1/tms/runs/run-9/export?format=csv",
			expect.objectContaining({ method: "GET" }),
		);
		expect(blob).toBeInstanceOf(Blob);
	});
});
