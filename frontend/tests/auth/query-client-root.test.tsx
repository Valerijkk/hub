import type { QueryClient } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

const capturedProviderClients = vi.hoisted(() => [] as QueryClient[]);

vi.mock("@tanstack/react-query", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@tanstack/react-query")>();
	return {
		...actual,
		QueryClientProvider: ({
			client,
			children,
		}: {
			client: QueryClient;
			children: ReactNode;
		}) => {
			capturedProviderClients.push(client);
			return <>{children}</>;
		},
	};
});

vi.mock("@tanstack/react-router", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@tanstack/react-router")>();
	return {
		...actual,
		Outlet: () => <div data-testid="root-outlet" />,
	};
});

vi.mock("@/components/theme-provider", () => ({
	ThemeProvider: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

vi.mock("sonner", () => ({
	Toaster: () => null,
}));

import { queryClient } from "@/lib/query-client";
import { Route } from "@/routes/__root";

describe("root query client ownership", () => {
	it("passes the exported application queryClient singleton to QueryClientProvider", () => {
		capturedProviderClients.length = 0;
		const RootLayout = Route.options.component;
		if (!RootLayout) throw new Error("root route must define a component");

		render(<RootLayout />);

		expect(capturedProviderClients).toEqual([queryClient]);
	});
});