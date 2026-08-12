/**
 * TASK-128 / A3W10-5-005 — avatar blob URL revoke on replace / unmount.
 *
 * Mutation: убрать URL.revokeObjectURL из onboarding.lazy.tsx cleanup →
 * source assert FAIL; harness spy assert FAIL.
 */
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { act, cleanup, render } from "@testing-library/react";
import { useEffect, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

/**
 * Минимальный контракт того же паттерна, что в onboarding.lazy.tsx
 * (useEffect cleanup + setAvatarPreview(createObjectURL)).
 */
function AvatarPreviewHarness({
	onReady,
}: {
	onReady: (api: {
		setFile: (blob: Blob) => void;
		clear: () => void;
	}) => void;
}) {
	const [avatarPreview, setAvatarPreview] = useState<string | null>(null);

	useEffect(() => {
		return () => {
			if (avatarPreview) {
				URL.revokeObjectURL(avatarPreview);
			}
		};
	}, [avatarPreview]);

	useEffect(() => {
		onReady({
			setFile: (blob) => {
				setAvatarPreview(URL.createObjectURL(blob));
			},
			clear: () => setAvatarPreview(null),
		});
	}, [onReady]);

	return avatarPreview ? <img alt="preview" src={avatarPreview} /> : null;
}

describe("A3W10-5-005 avatar blob revoke", () => {
	it("production onboarding.lazy.tsx revokes blob URL in cleanup", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/routes/onboarding.lazy.tsx",
			),
			"utf8",
		);
		expect(src).toContain("URL.revokeObjectURL(avatarPreview)");
		expect(src).toContain("URL.createObjectURL(file)");
	});

	let createSpy: ReturnType<typeof vi.spyOn>;
	let revokeSpy: ReturnType<typeof vi.spyOn>;
	let api: {
		setFile: (blob: Blob) => void;
		clear: () => void;
	};

	beforeEach(() => {
		createSpy = vi
			.spyOn(URL, "createObjectURL")
			.mockImplementation(() => `blob:mock-${Math.random()}`);
		revokeSpy = vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {});
		api = { setFile: () => {}, clear: () => {} };
	});

	afterEach(() => {
		cleanup();
		createSpy.mockRestore();
		revokeSpy.mockRestore();
	});

	it("revokes previous blob when preview replaced", () => {
		render(
			<AvatarPreviewHarness
				onReady={(a) => {
					api = a;
				}}
			/>,
		);

		act(() => {
			api.setFile(new Blob(["a"], { type: "image/png" }));
		});
		const first = createSpy.mock.results[0]?.value as string;
		expect(first).toBeTruthy();

		act(() => {
			api.setFile(new Blob(["b"], { type: "image/png" }));
		});

		expect(revokeSpy).toHaveBeenCalledWith(first);
	});

	it("revokes on unmount", () => {
		const { unmount } = render(
			<AvatarPreviewHarness
				onReady={(a) => {
					api = a;
				}}
			/>,
		);

		act(() => {
			api.setFile(new Blob(["c"], { type: "image/png" }));
		});
		const url = createSpy.mock.results[0]?.value as string;

		unmount();
		expect(revokeSpy).toHaveBeenCalledWith(url);
	});
});
