/**
 * Settings layout — left sidebar navigation + right content area.
 * Wraps all /settings/* child routes via <Outlet />.
 */
import { Link, Outlet, createFileRoute, useMatchRoute } from "@tanstack/react-router";
import { cn } from "@/lib/utils";

const navItems = [
	{ to: "/settings/profile", icon: "\u{1F464}", label: "\u041F\u0440\u043E\u0444\u0438\u043B\u044C" },
	{ to: "/settings/appearance", icon: "\u{1F3A8}", label: "\u0412\u043D\u0435\u0448\u043D\u0438\u0439 \u0432\u0438\u0434" },
	{ to: "/settings/notifications", icon: "\u{1F514}", label: "\u0423\u0432\u0435\u0434\u043E\u043C\u043B\u0435\u043D\u0438\u044F" },
	{ to: "/settings/security" as string, icon: "\u{1F512}", label: "\u0411\u0435\u0437\u043E\u043F\u0430\u0441\u043D\u043E\u0441\u0442\u044C", placeholder: true },
	{ to: "/settings/integrations" as string, icon: "\u{1F517}", label: "\u0418\u043D\u0442\u0435\u0433\u0440\u0430\u0446\u0438\u0438", placeholder: true },
	{ to: "/settings/hotkeys" as string, icon: "\u2328\uFE0F", label: "\u0413\u043E\u0440\u044F\u0447\u0438\u0435 \u043A\u043B\u0430\u0432\u0438\u0448\u0438", placeholder: true },
	{ to: "/trash", icon: "\u{1F4E6}", label: "\u041A\u043E\u0440\u0437\u0438\u043D\u0430" },
] as const;

function SettingsLayout() {
	return (
		<div className="flex h-full overflow-hidden -mx-4 -my-4 md:-mx-6 md:-my-4">
			{/* Left sidebar nav */}
			<nav className="w-[220px] min-w-[220px] border-r border-white/[0.06] p-4 space-y-0.5 overflow-y-auto">
				<h2 className="px-2.5 pb-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
					{"\u041D\u0430\u0441\u0442\u0440\u043E\u0439\u043A\u0438"}
				</h2>
				{navItems.map((item) => (
					<NavItem key={item.to} {...item} />
				))}
			</nav>

			{/* Right content area */}
			<div className="flex-1 overflow-y-auto p-6">
				<Outlet />
			</div>
		</div>
	);
}

function NavItem({
	to,
	icon,
	label,
	placeholder,
}: {
	to: string;
	icon: string;
	label: string;
	placeholder?: boolean;
}) {
	const matchRoute = useMatchRoute();
	const isActive = matchRoute({ to, fuzzy: true });

	if (placeholder) {
		return (
			<div
				className="flex items-center gap-2.5 rounded-lg px-2.5 py-[7px] text-[13.5px] text-muted-foreground/50 cursor-not-allowed select-none"
				title={`${label} (\u0441\u043A\u043E\u0440\u043E)`}
			>
				<span className="text-base leading-none">{icon}</span>
				<span className="truncate">{label}</span>
			</div>
		);
	}

	return (
		<Link
			to={to}
			className={cn(
				"flex items-center gap-2.5 rounded-lg px-2.5 py-[7px] text-[13.5px] transition-colors",
				"text-[#979A9F] hover:bg-white/[0.05] hover:text-white",
				isActive && "bg-white/[0.07] text-white font-medium",
			)}
		>
			<span className="text-base leading-none">{icon}</span>
			<span className="truncate">{label}</span>
		</Link>
	);
}

export const Route = createFileRoute("/_authenticated/settings")({
	component: SettingsLayout,
});
