import { act, fireEvent, render, renderHook, screen } from "@testing-library/react";
import type { ClipboardEvent, DragEvent } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useFileUpload } from "@/hooks/use-file-upload";

vi.mock("@/api/files", () => ({
	uploadFile: vi.fn().mockResolvedValue("file-id-1"),
}));

import { uploadFile } from "@/api/files";

function clipboardDataWithFile(file: File) {
	const item = {
		kind: "file" as const,
		type: file.type,
		getAsFile: () => file,
	};
	return {
		items: [item],
		files: [file],
	};
}

describe("useFileUpload drag/paste (P2-FE-11)", () => {
	beforeEach(() => {
		vi.mocked(uploadFile).mockClear();
	});

	it("N dragover after one enter: leave clears isDragging (dragCounter not inflated)", () => {
		const { result } = renderHook(() =>
			useFileUpload({
				projectId: "p1",
				entityType: "issue",
				entityId: "i1",
			}),
		);

		act(() => {
			result.current.dropZoneProps.onDragEnter({
				preventDefault: vi.fn(),
				stopPropagation: vi.fn(),
			} as unknown as DragEvent);
		});
		expect(result.current.isDragging).toBe(true);

		// Раньше каждый dragover инкрементировал counter → leave не обнулял.
		act(() => {
			for (let i = 0; i < 20; i++) {
				result.current.dropZoneProps.onDragOver({
					preventDefault: vi.fn(),
					stopPropagation: vi.fn(),
				} as unknown as DragEvent);
			}
		});
		expect(result.current.isDragging).toBe(true);

		act(() => {
			result.current.dropZoneProps.onDragLeave({
				preventDefault: vi.fn(),
				stopPropagation: vi.fn(),
			} as unknown as DragEvent);
		});
		expect(result.current.isDragging).toBe(false);
	});

	it("paste on drop zone uploads; document paste does not (no global hijack)", async () => {
		const file = new File(["img"], "paste.png", { type: "image/png" });
		const data = clipboardDataWithFile(file);

		function Zone() {
			const { dropZoneProps, isUploading } = useFileUpload({
				projectId: "p1",
				entityType: "issue",
				entityId: "i1",
			});
			return (
				<div data-testid="drop-zone" {...dropZoneProps}>
					{isUploading ? "uploading" : "idle"}
				</div>
			);
		}

		render(
			<>
				<input data-testid="other-input" />
				<Zone />
			</>,
		);

		const zone = screen.getByTestId("drop-zone");
		const input = screen.getByTestId("other-input");

		// Paste в input/document — не должен дергать upload Dropzone.
		fireEvent.paste(input, { clipboardData: data });
		fireEvent.paste(document, { clipboardData: data });
		expect(uploadFile).not.toHaveBeenCalled();

		// Paste на зоне — upload.
		await act(async () => {
			fireEvent.paste(zone, { clipboardData: data });
		});

		expect(uploadFile).toHaveBeenCalledTimes(1);
		expect(uploadFile).toHaveBeenCalledWith(file, "p1", "issue", "i1");
	});

	it("onPaste handler extracts files from clipboardData", async () => {
		const file = new File(["x"], "x.png", { type: "image/png" });
		const { result } = renderHook(() =>
			useFileUpload({
				projectId: "p1",
				entityType: "issue",
				entityId: "i1",
			}),
		);

		await act(async () => {
			result.current.dropZoneProps.onPaste({
				clipboardData: clipboardDataWithFile(file),
				preventDefault: vi.fn(),
			} as unknown as ClipboardEvent);
		});

		expect(uploadFile).toHaveBeenCalledWith(file, "p1", "issue", "i1");
	});
});
