import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as callsApi from "@/api/calls";
import { RecordingsList } from "@/modules/calls/recordings-list";

vi.mock("@/api/calls");

/** A3W10-1-003: download opens with noopener,noreferrer. */
describe("RecordingsList download noopener (A3W10-1-003)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(callsApi.listProjectRecordings).mockResolvedValue({
			recordings: [
				{
					id: "rec-1",
					egress_id: "eg-1",
					project_id: "p1",
					call_id: "c1",
					started_at: "2026-07-29T10:00:00Z",
					duration_seconds: 90,
					status: "ready",
					file_size: 1024,
				},
			],
			total: 1,
		});
		vi.mocked(callsApi.getRecordingDownloadURL).mockResolvedValue({
			url: "https://cdn.example/rec.mp4",
			expires_at: "2026-07-29T12:00:00Z",
		});
	});

	it("passes noopener,noreferrer as third arg to window.open", async () => {
		const openSpy = vi.spyOn(window, "open").mockReturnValue(null);

		render(<RecordingsList projectId="p1" />);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: /скачать/i })).toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole("button", { name: /скачать/i }));

		await waitFor(() => {
			expect(openSpy).toHaveBeenCalledWith(
				"https://cdn.example/rec.mp4",
				"_blank",
				"noopener,noreferrer",
			);
		});

		openSpy.mockRestore();
	});
});
