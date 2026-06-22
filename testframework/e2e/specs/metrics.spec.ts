/**
 * E2E-компонент Hub: metrics.spec.ts.
 */
import { expect, skipUnlessBackend, test } from "../fixtures";
import { getApiUrl } from "../helpers/backend";

// E2E-группа: Метрики Prometheus.
test.describe("Prometheus metrics", () => {
	// Проверяет: GET /metrics возвращает текст в формате exposition Prometheus.
	test("GET /metrics returns Prometheus text exposition", async ({
		backendAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);

		const base = getApiUrl().replace(/\/api\/?$/, "");
		const res = await fetch(`${base}/metrics`);
		expect(res.status).toBe(200);

		const contentType = res.headers.get("content-type") ?? "";
		expect(contentType).toMatch(/text\/plain/);

		const body = await res.text();
		expect(body).toContain("hub_operation_calls_total");
		expect(body).toContain("hub_http_requests_total");
	});
});
