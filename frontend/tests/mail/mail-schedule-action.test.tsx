import { fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { UserResponse } from "@/api/auth";
import type { MailThreadMessageSummary } from "@/api/mail";
import * as mailApi from "@/api/mail";
import * as messengerApi from "@/api/messenger";
import * as projectsApi from "@/api/projects";
import { MailScheduleAction } from "@/modules/mail/mail-schedule-action";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/mail");
vi.mock("@/api/messenger");
vi.mock("@/api/projects");
vi.mock("react-i18next", async (importOriginal) => ({
	...(await importOriginal<typeof import("react-i18next")>()),
	useTranslation: () => ({
		t: (key: string) => key,
	}),
}));

const currentUser: UserResponse = {
	id: "self-user",
	email: "me@hub.test",
	name: "Me",
	timezone: "Europe/Moscow",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

const message: MailThreadMessageSummary = {
	id: "11111111-1111-4111-8111-111111111111",
	thread_id: "22222222-2222-4222-8222-222222222222",
	from: { email: "Sender@external.test", name: "Sender" },
	to: [
		{ email: "ME@hub.test", name: "Me" },
		{ email: "Colleague@hub.test", name: "Colleague" },
	],
	cc: [
		{ email: "colleague@hub.test", name: "Colleague" },
		{ email: "Copy@hub.test", name: "Copy" },
	],
	subject: "Project kickoff",
	snippet: "Agenda",
	sent_at: "2026-09-01T10:00:00Z",
	received_at: "2026-09-01T10:01:00Z",
	priority: "normal",
	size_bytes: 42,
	has_attachments: false,
	is_draft: false,
	folder_ids: [],
	seen: true,
	flagged: false,
	answered: false,
	junk: false,
	calendar_method: null,
};

describe("MailScheduleAction", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({ user: currentUser });
		vi.mocked(projectsApi.listProjects).mockResolvedValue([
			{
				id: "project-1",
				name: "Project",
				key: "PRJ",
				status: "active",
				member_count: 3,
				version: 1,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			},
		]);
		vi.mocked(messengerApi.getChannels).mockResolvedValue([
			{
				id: "channel-1",
				channel_type: "public",
				name: "Project channel",
				is_readonly: false,
				project_id: "project-1",
				created_by: "self-user",
				is_starred: false,
				created_at: "2026-01-01T00:00:00Z",
				updated_at: "2026-01-01T00:00:00Z",
			},
		]);
		vi.mocked(projectsApi.listMembers).mockResolvedValue([
			{
				user_id: "member-1",
				email: "colleague@hub.test",
				name: "Colleague",
				role: "developer",
				joined_at: "2026-01-01T00:00:00Z",
			},
			{
				user_id: "member-2",
				email: "copy@hub.test",
				name: "Copy",
				role: "tester",
				joined_at: "2026-01-01T00:00:00Z",
			},
		]);
	});

	afterEach(() => {
		useAuthStore.setState({ user: null });
	});

	it("opens the existing call dialog with unique resolved teammates from an email", async () => {
		vi.mocked(mailApi.resolveMailSchedulingParticipants).mockResolvedValue({
			participants: [
				{ email: "sender@external.test", user_id: null },
				{ email: "colleague@hub.test", user_id: "member-1" },
				{ email: "copy@hub.test", user_id: "member-2" },
				{ email: "unknown@hub.test", user_id: "missing-member" },
				{ email: "duplicate@hub.test", user_id: "member-1" },
			],
		});

		renderWithProviders(<MailScheduleAction message={message} />);
		fireEvent.click(
			screen.getByRole("button", { name: "reader.schedule.action" }),
		);

		await waitFor(() => {
			expect(mailApi.resolveMailSchedulingParticipants).toHaveBeenCalledWith(
				message.id,
				{
					addresses: [
						"sender@external.test",
						"colleague@hub.test",
						"copy@hub.test",
					],
				},
			);
		});

		expect(
			await screen.findByRole("button", {
				name: "\u0423\u0434\u0430\u043b\u0438\u0442\u044c \u0443\u0447\u0430\u0441\u0442\u043d\u0438\u043a\u0430 Colleague",
			}),
		).toBeInTheDocument();
		expect(
			screen.getByRole("button", {
				name: "\u0423\u0434\u0430\u043b\u0438\u0442\u044c \u0443\u0447\u0430\u0441\u0442\u043d\u0438\u043a\u0430 Copy",
			}),
		).toBeInTheDocument();
		expect(
			screen.queryByRole("button", {
				name: "\u0423\u0434\u0430\u043b\u0438\u0442\u044c \u0443\u0447\u0430\u0441\u0442\u043d\u0438\u043a\u0430 missing-member",
			}),
		).not.toBeInTheDocument();
		expect(screen.getByLabelText("callTitleLabel")).toHaveValue(
			"Project kickoff",
		);
	});

	it("keeps the call dialog closed when the resolver finds no teammate", async () => {
		vi.mocked(mailApi.resolveMailSchedulingParticipants).mockResolvedValue({
			participants: [
				{ email: "sender@external.test", user_id: null },
			],
		});

		renderWithProviders(<MailScheduleAction message={message} />);
		fireEvent.click(
			screen.getByRole("button", { name: "reader.schedule.action" }),
		);

		expect(await screen.findByRole("alert")).toHaveTextContent(
			"reader.schedule.unresolved",
		);
		expect(screen.queryByLabelText("callTitleLabel")).not.toBeInTheDocument();
	});

	it("keeps the call dialog closed when the resolver returns only the current user", async () => {
		vi.mocked(mailApi.resolveMailSchedulingParticipants).mockResolvedValue({
			participants: [
				{ email: "me@hub.test", user_id: currentUser.id },
			],
		});

		renderWithProviders(<MailScheduleAction message={message} />);
		fireEvent.click(
			screen.getByRole("button", { name: "reader.schedule.action" }),
		);

		expect(await screen.findByRole("alert")).toHaveTextContent(
			/^reader\.schedule\.unresolved$/,
		);
		expect(screen.queryByLabelText("callTitleLabel")).not.toBeInTheDocument();
	});

	it("shows a generic failure without exposing resolver error details", async () => {
		const privateEmail = "private-contact@external.test";
		const resolverError = new Error(`Resolver failed for ${privateEmail}`);
		vi.mocked(mailApi.resolveMailSchedulingParticipants).mockRejectedValue(
			resolverError,
		);

		renderWithProviders(<MailScheduleAction message={message} />);
		fireEvent.click(
			screen.getByRole("button", { name: "reader.schedule.action" }),
		);

		const alert = await screen.findByRole("alert");
		expect(alert).toHaveTextContent(/^reader\.schedule\.failed$/);
		expect(alert).not.toHaveTextContent(privateEmail);
		expect(document.body).not.toHaveTextContent(resolverError.message);
		expect(screen.queryByLabelText("callTitleLabel")).not.toBeInTheDocument();
	});
});
