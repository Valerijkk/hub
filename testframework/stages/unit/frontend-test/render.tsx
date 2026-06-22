import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, type RenderOptions } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

import "@/i18n";

/** TestRenderer options (Factory + DI, DRY for Vitest). */
export interface TestRenderOptions extends Omit<RenderOptions, "wrapper"> {
	queryClient?: QueryClient;
	wrapper?: (children: ReactNode) => ReactNode;
}

function createTestQueryClient(): QueryClient {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

/**
 * renderWithProviders — единая точка рендера (SOLID: SRP, DIP).
 * Не Singleton: каждый тест получает свой QueryClient.
 */
export function renderWithProviders(
	ui: ReactElement,
	options: TestRenderOptions = {},
) {
	const {
		queryClient = createTestQueryClient(),
		wrapper: ExtraWrapper,
		...renderOptions
	} = options;

	function Wrapper({ children }: { children: ReactNode }) {
		const inner = ExtraWrapper ? ExtraWrapper(children) : children;
		return (
			<QueryClientProvider client={queryClient}>{inner}</QueryClientProvider>
		);
	}

	return {
		queryClient,
		...render(ui, { wrapper: Wrapper, ...renderOptions }),
	};
}

export { createTestQueryClient };
