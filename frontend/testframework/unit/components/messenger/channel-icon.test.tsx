import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ChannelIcon } from "@/components/messenger/channel-icon";

describe("ChannelIcon", () => {
	it("renders hash for public channels", () => {
		render(<ChannelIcon type="public" />);

		expect(screen.getByText("#")).toBeInTheDocument();
	});

	it("renders svg icon for dm channels", () => {
		const { container } = render(<ChannelIcon type="dm" />);

		expect(container.querySelector("svg")).toBeInTheDocument();
		expect(screen.queryByText("#")).not.toBeInTheDocument();
	});

	it("renders svg icon for group channels", () => {
		const { container } = render(<ChannelIcon type="group" />);

		expect(container.querySelector("svg")).toBeInTheDocument();
	});

	it("renders svg icon for private channels", () => {
		const { container } = render(<ChannelIcon type="private" />);

		expect(container.querySelector("svg")).toBeInTheDocument();
		expect(screen.queryByText("#")).not.toBeInTheDocument();
	});
});
