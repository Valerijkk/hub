import { expect, skipUnlessBackend, test } from "./fixtures";
import { fetchHealth, getApiUrl } from "./helpers/backend";

test.describe("API health", () => {
	test("GET /api/health returns status and service checks", async ({
		backendAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);

		const health = await fetchHealth();
		expect(health).not.toBeNull();
		expect(health?.status).toMatch(/^(ok|degraded)$/);
		expect(health?.services).toHaveProperty("postgres");
		expect(health?.services).toHaveProperty("redis");
		expect(health?.services).toHaveProperty("mongo");
	});

	test("health endpoint is reachable via frontend proxy", async ({
		request,
		backendAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);

		const res = await request.get("/api/health");
		expect([200, 503]).toContain(res.status());

		const body = (await res.json()) as { status: string };
		expect(body.status).toMatch(/^(ok|degraded)$/);
	});

	test("reports backend URL in skip message when unreachable", async ({
		backendAvailable,
	}) => {
		test.skip(backendAvailable, "sanity: backend is up for this run");
		expect(getApiUrl()).toContain("3000");
	});
});
