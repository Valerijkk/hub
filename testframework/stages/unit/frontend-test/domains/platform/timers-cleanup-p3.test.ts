import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { beforeEach, describe, expect, it } from "vitest";
import { useTimerStore } from "@/stores/timer-store";

const root = resolve(__dirname, "../../../../../../frontend/src");

function readSrc(rel: string): string {
	return readFileSync(resolve(root, rel), "utf8");
}

describe("P3-FE-05/16 timer cleanup + NaN (TASK-072)", () => {
	it("CallRoom InviteBlock clears copy timer on unmount", () => {
		const src = readSrc("components/calls/CallRoom.tsx");
		expect(src).toMatch(/copyTimerRef/);
		expect(src).toMatch(/clearTimeout\(copyTimerRef\.current\)/);
		expect(src).not.toMatch(
			/setCopied\(true\);\s*setTimeout\(\(\)\s*=>\s*setCopied\(false\)/,
		);
	});

	it("bot-markdown CodeBlock clears copy timer on unmount", () => {
		const src = readSrc("components/messenger/bot-markdown.tsx");
		expect(src).toMatch(/copyTimerRef/);
		expect(src).toMatch(/clearTimeout\(copyTimerRef\.current\)/);
	});

	it("sprint-retro clears saved timer on unmount", () => {
		const src = readSrc("modules/tracker/sprint-retro.tsx");
		expect(src).toMatch(/savedTimerRef/);
		expect(src).toMatch(/clearTimeout\(savedTimerRef\.current\)/);
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
