/**
 * E2E-компонент Hub: health.spec.ts.
 */
import { expect, skipUnlessBackend, test } from "../fixtures";
import { fetchHealth, getApiUrl } from "../helpers/backend";

// E2E-группа: API здоровья.
test.describe("API health", () => {
	// Проверяет: GET /api/health возвращает статус и проверки сервисов.
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

	// Проверяет: endpoint health доступен через прокси фронтенда.
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

	// Проверяет: URL backend в сообщении пропуска, когда сервис недоступен.
	test("reports backend URL in skip message when unreachable", async ({
		backendAvailable,
	}) => {
		test.skip(backendAvailable, "sanity: backend is up for this run");
		expect(getApiUrl()).toContain("3000");
	});
});
