import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import {
	cleanup,
	fireEvent,
	render,
	screen,
	waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { BotMarkdown } from "@/components/messenger/bot-markdown";
import { SprintRetro } from "@/modules/tracker/sprint-retro";
import { useSprintStore } from "@/stores/sprint-store";
import { useTimerStore } from "@/stores/timer-store";

const root = resolve(__dirname, "../../../../../../frontend/src");

function readSrc(rel: string): string {
	return readFileSync(resolve(root, rel), "utf8");
}

/**
 * Unmount-only cleanup: clearTimeout must sit inside
 * `useEffect(() => () => { … }, [])`, not only in onCopy/handleSave.
 */
const unmountCopyTimerCleanup =
	/useEffect\(\s*\(\)\s*=>\s*\(\)\s*=>\s*\{\s*if\s*\(\s*copyTimerRef\.current\s*\)\s*clearTimeout\(\s*copyTimerRef\.current\s*\);?\s*\}\s*,\s*\[\s*\]/;

const unmountSavedTimerCleanup =
	/useEffect\(\s*\(\)\s*=>\s*\(\)\s*=>\s*\{\s*if\s*\(\s*savedTimerRef\.current\s*\)\s*clearTimeout\(\s*savedTimerRef\.current\s*\);?\s*\}\s*,\s*\[\s*\]/;

describe("P3-FE-05/16 timer cleanup + NaN (TASK-072)", () => {
	afterEach(() => {
		cleanup();
		vi.restoreAllMocks();
		vi.useRealTimers();
	});

	it("CallRoom InviteBlock clears copy timer on unmount", () => {
		const src = readSrc("components/calls/CallRoom.tsx");
		expect(src).toMatch(/copyTimerRef/);
		expect(src).toMatch(unmountCopyTimerCleanup);
		expect(src).not.toMatch(
			/setCopied\(true\);\s*setTimeout\(\(\)\s*=>\s*setCopied\(false\)/,
		);
	});

	it("bot-markdown CodeBlock clears copy timer on unmount", () => {
		const src = readSrc("components/messenger/bot-markdown.tsx");
		expect(src).toMatch(/copyTimerRef/);
		// Must be useEffect unmount cleanup — onCopy clearTimeout alone is NOT enough.
		expect(src).toMatch(unmountCopyTimerCleanup);
		expect(src).not.toMatch(
			/setCopied\(true\);\s*(?:toast\.[^;]+;\s*)?setTimeout\(\(\)\s*=>\s*setCopied\(false\)/,
		);
	});

	it("sprint-retro clears saved timer on unmount", () => {
		const src = readSrc("modules/tracker/sprint-retro.tsx");
		expect(src).toMatch(/savedTimerRef/);
		// Must be useEffect unmount cleanup — handleSave clearTimeout alone is NOT enough.
		expect(src).toMatch(unmountSavedTimerCleanup);
		expect(src).not.toMatch(
			/setSaved\(true\);\s*(?:toast\.[^;]+;\s*)?setTimeout\(\(\)\s*=>\s*setSaved\(false\)/,
		);
	});

	it("bot-markdown CodeBlock clearTimeout runs on unmount after copy", async () => {
		vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue(undefined);
		const clearSpy = vi.spyOn(globalThis, "clearTimeout");

		const { unmount } = render(
			<BotMarkdown text={"```js\nconst x = 1;\n```"} />,
		);
		fireEvent.click(screen.getByRole("button", { name: "Копировать код" }));
		await waitFor(() => {
			expect(navigator.clipboard.writeText).toHaveBeenCalled();
		});

		clearSpy.mockClear();
		unmount();
		expect(clearSpy).toHaveBeenCalled();
	});

	it("sprint-retro clearTimeout runs on unmount after save", async () => {
		const saveRetroNotes = vi.fn().mockResolvedValue(undefined);
		useSprintStore.setState({ saveRetroNotes });
		const clearSpy = vi.spyOn(globalThis, "clearTimeout");

		const { unmount } = render(
			<SprintRetro projectId="p1" sprintId="s1" />,
		);
		fireEvent.click(screen.getByRole("button", { name: /сохранить/i }));
		await waitFor(() => {
			expect(saveRetroNotes).toHaveBeenCalled();
		});

		clearSpy.mockClear();
		unmount();
		expect(clearSpy).toHaveBeenCalled();
	});

	beforeEach(() => {
		useTimerStore.setState({
			isRunning: false,
			activeIssueKey: null,
			activeProjectId: null,
			startedAt: null,
			elapsedSeconds: 0,
		});
		localStorage.removeItem("hub-timer");
	});

	it("tick with corrupt startedAt stops timer instead of NaN elapsed", () => {
		useTimerStore.setState({
			isRunning: true,
			activeIssueKey: "HUB-1",
			activeProjectId: "p1",
			startedAt: "bad",
			elapsedSeconds: 0,
		});
		useTimerStore.getState().tick();
		const state = useTimerStore.getState();
		expect(Number.isNaN(state.elapsedSeconds)).toBe(false);
		expect(state.isRunning).toBe(false);
		expect(state.startedAt).toBeNull();
		expect(state.elapsedSeconds).toBe(0);
	});

	it("tick with valid startedAt updates finite elapsed", () => {
		const started = new Date(Date.now() - 5000).toISOString();
		useTimerStore.setState({
			isRunning: true,
			activeIssueKey: "HUB-2",
			activeProjectId: "p1",
			startedAt: started,
			elapsedSeconds: 0,
		});
		useTimerStore.getState().tick();
		const { elapsedSeconds, isRunning } = useTimerStore.getState();
		expect(isRunning).toBe(true);
		expect(Number.isFinite(elapsedSeconds)).toBe(true);
		expect(elapsedSeconds).toBeGreaterThanOrEqual(4);
	});
});
