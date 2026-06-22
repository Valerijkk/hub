import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import "@/i18n";

vi.mock("@/api/projects", () => ({
	listMembers: vi.fn(),
}));

import { listMembers, type MemberResponse } from "@/api/projects";
import { UserPicker } from "@/components/tracker/user-picker";

const members: MemberResponse[] = [
	{
		user_id: "u1",
		name: "Alice",
		email: "alice@hub.local",
		role: "member",
		joined_at: "2026-01-01T00:00:00Z",
	},
	{
		user_id: "u2",
		name: "Bob",
		email: "bob@hub.local",
		role: "member",
		joined_at: "2026-01-01T00:00:00Z",
	},
];

function renderPicker(
	props: Partial<React.ComponentProps<typeof UserPicker>> = {},
) {
	const qc = new QueryClient({
		defaultOptions: { queries: { retry: false } },
	});
	vi.mocked(listMembers).mockResolvedValue(members);
	const onChange = props.onChange ?? vi.fn();
	render(
		<QueryClientProvider client={qc}>
			<UserPicker projectId="p1" value="" onChange={onChange} {...props} />
		</QueryClientProvider>,
	);
	return { onChange };
}

describe("UserPicker", () => {
	it("renders combobox with placeholder", async () => {
		renderPicker();
		expect(await screen.findByRole("combobox")).toBeInTheDocument();
	});

	it("shows selected member name", async () => {
		renderPicker({ value: "u1" });
		expect(await screen.findByText("Alice")).toBeInTheDocument();
	});

	it("clears assignee when clear button clicked", async () => {
		const { onChange } = renderPicker({ value: "u1" });
		const clearBtn = await screen.findByRole("button", {
			name: /clear assignee|очистить/i,
		});
		fireEvent.click(clearBtn);
		expect(onChange).toHaveBeenCalledWith("");
	});
});
