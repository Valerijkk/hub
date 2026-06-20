import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as adminApi from "@/api/admin";
import { AuditLogViewer } from "./audit-log";

const mockGetAuditLogs = vi.spyOn(adminApi, "getAuditLogs");

describe("AuditLogViewer", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		mockGetAuditLogs.mockResolvedValue({
			items: [],
			total: 0,
			page: 1,
			per_page: 25,
		});
	});

	it("renders audit log filters", async () => {
		render(<AuditLogViewer />);

		await waitFor(() => {
			expect(mockGetAuditLogs).toHaveBeenCalled();
		});

		const dateInputs = document.querySelectorAll('input[type="date"]');
		expect(dateInputs).toHaveLength(2);
		expect(screen.getByPlaceholderText("User ID")).toBeInTheDocument();
		expect(screen.getByPlaceholderText("IP")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Применить" }),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: "Экспорт в CSV" }),
		).toBeInTheDocument();
	});

	it("loads audit entries from API", async () => {
		mockGetAuditLogs.mockResolvedValue({
			items: [
				{
					id: "entry-1",
					action: "login",
					entity_type: "session",
					created_at: "2026-06-17T10:00:00Z",
				},
			],
			total: 1,
			page: 1,
			per_page: 25,
		});

		render(<AuditLogViewer />);

		expect(await screen.findByText("login")).toBeInTheDocument();
	});

	it("shows slow-requests filter", async () => {
		render(<AuditLogViewer />);

		await waitFor(() => {
			expect(mockGetAuditLogs).toHaveBeenCalled();
		});

		expect(screen.getByText(/Медленные/)).toBeInTheDocument();
	});

	it("toggling slow checkbox calls getAuditLogs with slow=1", async () => {
		render(<AuditLogViewer />);

		await waitFor(() => {
			expect(mockGetAuditLogs).toHaveBeenCalled();
		});

		mockGetAuditLogs.mockClear();

		const checkbox = screen.getByRole("checkbox", { name: /Медленные/ });
		fireEvent.click(checkbox);

		await waitFor(() => {
			expect(mockGetAuditLogs).toHaveBeenCalledWith(
				expect.objectContaining({ slow: true }),
			);
		});
	});
});
