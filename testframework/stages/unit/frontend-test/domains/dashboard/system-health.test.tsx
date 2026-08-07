import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { AdminStats } from "@/api/dashboard";
import { SystemHealthWidget } from "@/components/dashboard/widgets/system-health";

function adminStatsWithHealth(
	serviceHealth: AdminStats["service_health"],
): AdminStats {
	return {
		total_users: 0,
		active_users: 0,
		total_projects: 0,
		total_issues: 0,
		dau: 0,
		wau: 0,
		mau: 0,
		storage_used: 0,
		service_health: serviceHealth,
		top_projects: [],
	};
}

describe("SystemHealthWidget", () => {
	it("shows empty state when no service data", () => {
		render(<SystemHealthWidget />);

		expect(screen.getByText("Нет данных о сервисах")).toBeInTheDocument();
	});

	it("renders service names and healthy status", () => {
		const data = adminStatsWithHealth({
			postgres: "healthy",
			redis: "healthy",
		});

		render(<SystemHealthWidget data={data} />);

		expect(screen.getByText("postgres")).toBeInTheDocument();
		expect(screen.getByText("redis")).toBeInTheDocument();
		expect(screen.getAllByText("Работает")).toHaveLength(2);
		expect(screen.getByText("2/2 сервисов в норме")).toBeInTheDocument();
	});

	it("renders unhealthy service label", () => {
		const data = adminStatsWithHealth({ nats: "unhealthy" });

		render(<SystemHealthWidget data={data} />);

		expect(screen.getByText("nats")).toBeInTheDocument();
		expect(screen.getByText("Ошибка")).toBeInTheDocument();
		expect(screen.getByText("0/1 сервисов в норме")).toBeInTheDocument();
	});

	it("renders degraded service label", () => {
		const data = adminStatsWithHealth({ mongodb: "degraded" });

		render(<SystemHealthWidget data={data} />);

		expect(screen.getByText("mongodb")).toBeInTheDocument();
		expect(screen.getByText("Деградация")).toBeInTheDocument();
	});
});
