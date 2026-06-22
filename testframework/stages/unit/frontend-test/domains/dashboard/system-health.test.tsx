import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { AdminStats } from "@/api/dashboard";
import { SystemHealthWidget } from "@/components/dashboard/widgets/system-health";

describe("SystemHealthWidget", () => {
	it("shows empty state when no service data", () => {
		render(<SystemHealthWidget />);

		expect(screen.getByText("Нет данных о сервисах")).toBeInTheDocument();
	});

	it("renders service names and healthy status", () => {
		const data = {
			service_health: {
				postgres: "healthy",
				redis: "healthy",
			},
		} as unknown as AdminStats;

		render(<SystemHealthWidget data={data} />);

		expect(screen.getByText("postgres")).toBeInTheDocument();
		expect(screen.getByText("redis")).toBeInTheDocument();
		expect(screen.getAllByText("Работает")).toHaveLength(2);
		expect(screen.getByText("2/2 сервисов в норме")).toBeInTheDocument();
	});

	it("renders unhealthy service label", () => {
		const data = {
			service_health: {
				nats: "unhealthy",
			},
		} as unknown as AdminStats;

		render(<SystemHealthWidget data={data} />);

		expect(screen.getByText("nats")).toBeInTheDocument();
		expect(screen.getByText("Ошибка")).toBeInTheDocument();
		expect(screen.getByText("0/1 сервисов в норме")).toBeInTheDocument();
	});

	it("renders degraded service label", () => {
		const data = {
			service_health: {
				mongodb: "degraded",
			},
		} as unknown as AdminStats;

		render(<SystemHealthWidget data={data} />);

		expect(screen.getByText("mongodb")).toBeInTheDocument();
		expect(screen.getByText("Деградация")).toBeInTheDocument();
	});
});
