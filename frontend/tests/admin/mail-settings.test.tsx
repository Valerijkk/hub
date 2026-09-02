import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { MailSettingsPanel } from "@/components/admin/mail-settings";

const settingsFixture = {
	hub_mail_url: "https://mail.hub.example",
	from_alerts: "Hub Alerts <alerts@hub.example>",
	alerts_enabled: true,
	daily_limit: 20,
	resend_ready: true,
};

afterEach(() => {
	vi.restoreAllMocks();
});

describe("MailSettingsPanel", () => {
	it("keeps only settings controls and saves the trimmed numeric settings payload", async () => {
		const getSpy = vi.spyOn(api, "get").mockResolvedValue(settingsFixture);
		const putSpy = vi.spyOn(api, "put").mockResolvedValue(settingsFixture);
		const postSpy = vi.spyOn(api, "post").mockResolvedValue(undefined as never);

		render(<MailSettingsPanel />);

		const hubMailURL = await screen.findByLabelText(/url hub-mail/i);
		const fromAlerts = screen.getByLabelText(/from \(alerts\)/i);
		const dailyLimit = screen.getByLabelText(/(?:daily limit|лимит)/i);

		expect(getSpy).toHaveBeenCalledWith("/admin/mail/settings");
		expect(hubMailURL).toHaveValue(settingsFixture.hub_mail_url);
		expect(fromAlerts).toHaveValue(settingsFixture.from_alerts);
		expect(dailyLimit).toHaveValue(settingsFixture.daily_limit);
		expect(screen.getByRole("checkbox")).toBeChecked();
		expect(
			screen.getByRole("button", { name: /(?:save|сохранить)/i }),
		).toBeInTheDocument();

		expect(
			screen.queryByText(/(?:marketing broadcast|маркетинговая рассылка)/i),
		).not.toBeInTheDocument();
		expect(
			screen.queryByLabelText(/(?:subject|тема)/i),
		).not.toBeInTheDocument();
		expect(
			screen.queryByLabelText(/(?:html|body|текст)/i),
		).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: /(?:dry[- ]?run|тестов)/i }),
		).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: /(?:^send$|^отправить$)/i }),
		).not.toBeInTheDocument();

		fireEvent.change(hubMailURL, {
			target: { value: "  https://mail.example.test  " },
		});
		fireEvent.change(fromAlerts, {
			target: { value: "  Alerts <alerts@example.test>  " },
		});
		fireEvent.change(dailyLimit, { target: { value: "37" } });
		fireEvent.click(
			screen.getByRole("button", { name: /(?:save|сохранить)/i }),
		);

		await waitFor(() => {
			expect(putSpy).toHaveBeenCalledWith("/admin/mail/settings", {
				hub_mail_url: "https://mail.example.test",
				from_alerts: "Alerts <alerts@example.test>",
				alerts_enabled: true,
				daily_limit: 37,
			});
		});

		expect(postSpy).not.toHaveBeenCalled();
	});
});
