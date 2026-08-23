import { describe, expect, it, vi } from "vitest";
import { queryError } from "./query-result";

describe("query result fixtures", () => {
	it("creates a settled error result and preserves a custom refetch", () => {
		const error = new Error("fixture failure");
		const refetch = vi.fn();

		const result = queryError(error, refetch);

		expect(result.isError).toBe(true);
		expect(result.isLoading).toBe(false);
		expect(result.isPending).toBe(false);
		expect(result.error).toBe(error);
		expect(result.refetch).toBe(refetch);
	});
});
