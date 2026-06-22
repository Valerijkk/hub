import { createFileRoute, Link } from "@tanstack/react-router";
import { StatusPage } from "@/modules/status/status-page";

function PublicStatusRoute() {
	return (
		<div>
			<StatusPage />
			<div className="pb-8 text-center">
				<Link to="/login" className="text-sm text-t3 hover:text-t2 underline">
					Войти в Hub
				</Link>
			</div>
		</div>
	);
}

export const Route = createFileRoute("/status")({
	component: PublicStatusRoute,
});
