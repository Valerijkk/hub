import { createFileRoute, Outlet } from "@tanstack/react-router";
import { MailShell } from "@/modules/mail/mail-shell";

function MailRoute() {
	return (
		<MailShell>
			<Outlet />
		</MailShell>
	);
}

export const Route = createFileRoute("/_authenticated/mail")({
	component: MailRoute,
});
