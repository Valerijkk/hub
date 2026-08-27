import { createFileRoute } from "@tanstack/react-router";
import { MailOAuthCallback } from "@/modules/mail/mail-oauth-callback";

function MailOAuthCallbackRoute() {
	return <MailOAuthCallback />;
}

export const Route = createFileRoute("/_authenticated/mail_/oauth/callback")({
	component: MailOAuthCallbackRoute,
});
