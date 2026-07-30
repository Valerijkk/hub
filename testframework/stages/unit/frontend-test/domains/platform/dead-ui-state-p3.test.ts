import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { useTmsStore } from "@/stores/tms-store";

const fe = (...parts: string[]) =>
	resolve(__dirname, "../../../../../../frontend/src", ...parts);

describe("TASK-075 P3-FE-08 dead UI / unused state", () => {
	it("issues list has no dead statusFilter state", () => {
		const src = readFileSync(
			fe("routes/_authenticated/projects/$projectId/issues/index.tsx"),
			"utf8",
		);
		expect(src).not.toMatch(/statusFilter/);
		expect(src).not.toMatch(/setStatusFilter/);
	});

	it("settings NavItem has no unused placeholder branch", () => {
		const src = readFileSync(fe("routes/_authenticated/settings.tsx"), "utf8");
		expect(src).not.toMatch(/placeholder\?:/);
		expect(src).not.toMatch(/if \(placeholder\)/);
	});

	it("issue-attachments empty copy does not promise DnD", () => {
		const src = readFileSync(
			fe("components/tracker/issue-attachments.tsx"),
			"utf8",
		);
		expect(src).not.toMatch(/Перетащите/);
		expect(src).toMatch(/Добавить файл/);
	});

	it("issue-sidebar-meta assignee/reporter are not cursor-pointer", () => {
		const src = readFileSync(
			fe("modules/tracker/issue-sidebar-meta.tsx"),
			"utf8",
		);
		// Slice people sections only — Status/Priority above legitimately use cursor-pointer.
		// Match className=…cursor-pointer so comment text cannot self-match (W10-A3-019).
		const peopleStart = src.indexOf("{/* Исполнитель");
		const peopleEnd = src.indexOf("{/* Метки");
		expect(peopleStart).toBeGreaterThanOrEqual(0);
		expect(peopleEnd).toBeGreaterThan(peopleStart);
		const people = src.slice(peopleStart, peopleEnd);
		expect(people).toMatch(/Исполнитель/);
		expect(people).toMatch(/Автор/);
		expect(people).not.toMatch(/className="[^"]*cursor-pointer/);
	});

	it("run-execution-panel has no dead stepResults state", () => {
		const src = readFileSync(
			fe("modules/tms/components/run-execution-panel.tsx"),
			"utf8",
		);
		expect(src).not.toMatch(/stepResults/);
	});

	it("mention-autocomplete Esc sets suppressed", () => {
		const src = readFileSync(
			fe("components/messenger/mention-autocomplete.tsx"),
			"utf8",
		);
		expect(src).toMatch(/setSuppressed\(true\)/);
		expect(src).toMatch(/!suppressed/);
	});

	it("profile avatar opens file input; no dead department state", () => {
		const src = readFileSync(
			fe("routes/_authenticated/settings/profile.lazy.tsx"),
			"utf8",
		);
		// Avatar circle must wire onClick→fileInput (W10-A5-016); button-only click is insufficient.
		expect(src).toMatch(
			/aria-label="Загрузить фото профиля"\s*\n\s*onClick=\{\(\) => fileInputRef\.current\?\.click\(\)\}/,
		);
		expect(src).not.toMatch(/department:/);
		expect(src).toMatch(/const statusDuration = "1d"/);
	});

	it("tms-store is filters-only (no server-cache fields)", () => {
		const state = useTmsStore.getState();
		expect(state).toHaveProperty("filters");
		expect(state).toHaveProperty("updateFilter");
		expect(state).toHaveProperty("resetFilters");
		expect(state).not.toHaveProperty("testCases");
		expect(state).not.toHaveProperty("setTestCases");
		expect(state).not.toHaveProperty("suites");
		expect(state).not.toHaveProperty("testRuns");
	});

	it("tms cases bulk bar has delete action", () => {
		const src = readFileSync(
			fe("routes/_authenticated/projects/$projectId/tms/cases.lazy.tsx"),
			"utf8",
		);
		expect(src).toMatch(/bulkDeleteMutation/);
		expect(src).toMatch(/deleteTestCase/);
		expect(src).toMatch(/Удалить/);
	});
});
