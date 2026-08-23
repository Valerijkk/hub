import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as onecApi from "@/api/admin-onec";
import { OnecSettingsPanel } from "@/components/admin/onec-settings";

vi.mock("@/api/admin-onec");

describe("OnecSettingsPanel", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(onecApi.getOnecSettings).mockResolvedValue({
			enabled: false,
			webhook_url: "",
			bearer_token_set: false,
			employee_id_field: "email",
		});
	});

	it("renders 1C connector settings form", async () => {
		render(<OnecSettingsPanel />);

		await waitFor(() => {
			expect(onecApi.getOnecSettings).toHaveBeenCalled();
		});
		expect(screen.getByText(/Коннектор 1C/i)).toBeInTheDocument();
		expect(screen.getByLabelText(/Webhook URL/i)).toBeInTheDocument();
	});

	it("shows error + retry instead of empty form when load fails (W9-A1-013)", async () => {
		vi.mocked(onecApi.getOnecSettings).mockRejectedValue(new Error("network"));
		render(<OnecSettingsPanel />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить настройки 1c/i,
			);
		});
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(screen.queryByLabelText(/Webhook URL/i)).not.toBeInTheDocument();
	});
});
