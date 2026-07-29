import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as adminApi from "@/api/admin";
import { AuditLogViewer } from "@/components/admin/audit-log";

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

	it("renders alert audit entry with alertname", async () => {
		mockGetAuditLogs.mockResolvedValue({
			items: [
				{
					id: "alert-1",
					action: "alert.firing",
					entity_type: "alert",
					entity_id: "fp-abc",
					created_at: "2026-06-24T10:00:00Z",
					details: {
						labels: { alertname: "HubHighHTTP5xxRate", severity: "warning" },
					},
				},
			],
			total: 1,
			page: 1,
			per_page: 25,
		});

		render(<AuditLogViewer />);

		expect(await screen.findByText("alert.firing")).toBeInTheDocument();
		expect(screen.getByText(/HubHighHTTP5xxRate/)).toBeInTheDocument();
		expect(screen.getByText(/warning/)).toBeInTheDocument();
	});

	it("shows empty copy only after successful empty load", async () => {
		render(<AuditLogViewer />);
		expect(await screen.findByText("Записи не найдены")).toBeInTheDocument();
		expect(screen.queryByRole("alert")).not.toBeInTheDocument();
	});

	it("shows banner+retry on fetchLogs fail, not empty-as-fail (A3W9-1-004)", async () => {
		mockGetAuditLogs.mockRejectedValue(new Error("network"));
		render(<AuditLogViewer />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить журнал аудита/i,
			);
		});
		expect(screen.queryByText("Записи не найдены")).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
