import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeAll, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/tracker", () => ({
	createIssue: vi.fn(),
	updateIssue: vi.fn(),
}));

vi.mock("@/components/tracker/plate-editor", () => ({
	PlateEditor: () => <div data-testid="plate-editor" />,
}));

vi.mock("@/components/tracker/user-picker", () => ({
	UserPicker: ({
		onChange,
		value,
	}: {
		onChange: (v: string) => void;
		value: string;
	}) => (
		<input
			data-testid="user-picker"
			value={value}
			onChange={(e) => onChange(e.target.value)}
		/>
	),
}));

import { createIssue } from "@/api/tracker";
import { IssueForm } from "@/components/tracker/issue-form";

function renderForm(mode: "create" | "edit" = "create") {
	const onSuccess = vi.fn();
	renderWithProviders(
		<IssueForm projectId="proj-1" mode={mode} onSuccess={onSuccess} />,
	);
	return { onSuccess };
}

beforeAll(async () => {
	await i18n.loadNamespaces("tracker");
});

describe("IssueForm", () => {
	it("shows validation error when title is too short", async () => {
		renderForm();
		const title = screen.getByLabelText(/название|title/i);
		fireEvent.change(title, { target: { value: "ab" } });
		fireEvent.blur(title);
		await waitFor(() => {
			expect(screen.getByText(/минимум 3|at least 3/i)).toBeInTheDocument();
		});
		expect(
			screen.getByRole("button", { name: /создать|create/i }),
		).toBeDisabled();
	});

	it("submits valid create form", async () => {
		vi.mocked(createIssue).mockResolvedValue({
			id: "iss-1",
			issue_key: "PROJ-1",
			title: "Valid issue title",
			issue_type: "task",
			priority: "medium",
			version: 1,
		} as never);

		const { onSuccess } = renderForm();
		const title = screen.getByLabelText(/название|title/i);
		fireEvent.change(title, { target: { value: "Valid issue title" } });

		const submit = screen.getByRole("button", { name: /создать|create/i });
		await waitFor(() => expect(submit).not.toBeDisabled());
		fireEvent.click(submit);

		await waitFor(() => {
			expect(createIssue).toHaveBeenCalledWith(
				"proj-1",
				expect.objectContaining({ title: "Valid issue title" }),
			);
		});
		expect(onSuccess).toHaveBeenCalled();
	});
});
