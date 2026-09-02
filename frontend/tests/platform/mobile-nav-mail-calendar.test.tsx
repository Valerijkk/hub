import { fireEvent, render, screen } from "@testing-library/react";
import type { ComponentProps, ReactNode } from "react";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import { MobileNav } from "@/components/mobile-nav";

type RouterLinkProps = Omit<ComponentProps<"a">, "href"> & {
	children: ReactNode;
	to: string;
	params?: {
		folderKey?: string;
	};
};

type MatchOptions = {
	fuzzy?: boolean;
	to: string;
};

const { currentPath } = vi.hoisted(() => ({
	currentPath: { value: "/calendar/events/event-1" },
}));

vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, to, params, ...props }: RouterLinkProps) => (
		<a
			href={
				params?.folderKey === undefined
					? to
					: to.replace("$folderKey", encodeURIComponent(params.folderKey))
			}
			{...props}
		>
			{children}
		</a>
	),
	useMatchRoute: () => ({ to, fuzzy }: MatchOptions) => {
		const path = currentPath.value;
		return fuzzy ? path.startsWith(to) : path === to;
	},
}));

describe("MobileNav Mail and Calendar access", () => {
	beforeEach(async () => {
		currentPath.value = "/calendar/events/event-1";
		await i18n.changeLanguage("en");
	});

	afterAll(async () => {
		await i18n.changeLanguage("ru");
	});

	it("opens first-class Mail and Calendar links, closes after selection, and keeps More active on their routes", () => {
		const { rerender } = render(<MobileNav />);
		const more = screen.getByRole("button", { name: "More" });

		expect(more).toHaveClass("text-primary");

		fireEvent.click(more);

		expect(screen.getByRole("link", { name: "Mail" })).toHaveAttribute(
			"href",
			"/mail/inbox",
		);
		expect(screen.getByRole("link", { name: "Calendar" })).toHaveAttribute(
			"href",
			"/calendar",
		);

		fireEvent.click(screen.getByRole("link", { name: "Mail" }));

		expect(
			screen.queryByRole("link", { name: "Calendar" }),
		).not.toBeInTheDocument();
		expect(more).toHaveClass("text-primary");

		currentPath.value = "/mail/inbox";
		rerender(<MobileNav />);

		expect(screen.getByRole("button", { name: "More" })).toHaveClass(
			"text-primary",
		);
	});
});
