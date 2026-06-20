import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import type { Result } from "axe-core";

const CRITICAL_IMPACTS = new Set(["critical", "serious"]);

export function getCriticalViolations(violations: Result[]): Result[] {
	return violations.filter(
		(violation) => violation.impact && CRITICAL_IMPACTS.has(violation.impact),
	);
}

export function formatViolations(violations: Result[]): string {
	if (violations.length === 0) {
		return "";
	}

	return violations
		.map((violation) => {
			const targets = violation.nodes
				.map((node) => node.target.join(", "))
				.join("\n    ");
			return `[${violation.impact}] ${violation.id}: ${violation.help}\n    ${targets}`;
		})
		.join("\n\n");
}

export async function scanAccessibility(
	page: Page,
	options?: { include?: string },
): Promise<Result[]> {
	let builder = new AxeBuilder({ page });
	if (options?.include) {
		builder = builder.include(options.include);
	}

	const results = await builder.analyze();
	return getCriticalViolations(results.violations);
}
