/**
 * Settings layout — left sidebar navigation + right content area.
 * Wraps all /settings/* child routes via <Outlet />.
 */
import {
	createFileRoute,
	Link,
	Outlet,
	useMatchRoute,
} from "@tanstack/react-router";
import {
	Bell,
	Keyboard,
	type LucideIcon,
	Palette,
	Puzzle,
	Shield,
	Trash2,
	User,
} from "lucide-react";
import { cn } from "@/lib/utils";

const navItems = [
	{
		to: "/settings/profile",
		icon: User,
		label: "\u041F\u0440\u043E\u0444\u0438\u043B\u044C",
	},
	{
		to: "/settings/appearance",
		icon: Palette,
		label: "\u0412\u043D\u0435\u0448\u043D\u0438\u0439 \u0432\u0438\u0434",
	},
	{
		to: "/settings/notifications",
		icon: Bell,
		label: "\u0423\u0432\u0435\u0434\u043E\u043C\u043B\u0435\u043D\u0438\u044F",
	},
	{
		to: "/settings/security",
		icon: Shield,
		label:
			"\u0411\u0435\u0437\u043E\u043F\u0430\u0441\u043D\u043E\u0441\u0442\u044C",
	},
	{
		to: "/settings/integrations",
		icon: Puzzle,
		label: "\u0418\u043D\u0442\u0435\u0433\u0440\u0430\u0446\u0438\u0438",
	},
	{
		to: "/settings/hotkeys",
		icon: Keyboard,
		label:
			"\u0413\u043E\u0440\u044F\u0447\u0438\u0435 \u043A\u043B\u0430\u0432\u0438\u0448\u0438",
	},
	{
		to: "/trash",
		icon: Trash2,
		label: "\u041A\u043E\u0440\u0437\u0438\u043D\u0430",
	},
] as const;

function SettingsLayout() {
	return (
		<div className="flex flex-1 overflow-hidden animate-[eu_400ms_var(--spring)] -mx-4 -my-4 md:-mx-6 md:-my-4">
			{/* Left sidebar nav */}
			<nav className="w-full md:w-60 md:min-w-60 border-r border-b1 py-4 px-3 overflow-y-auto">
				<div className="text-[10px] font-bold uppercase tracking-[0.08em] text-t2 px-[10px] pt-[14px] pb-[6px]">
					{"\u0410\u043A\u043A\u0430\u0443\u043D\u0442"}
				</div>
				{navItems.map((item) => (
					<NavItem key={item.to} {...item} />
				))}
			</nav>

			{/* Right content area */}
			<div className="flex-1 overflow-y-auto px-9 py-7 max-w-[720px]">
				<Outlet />
			</div>
		</div>
	);
}

function NavItem({
	to,
	icon: Icon,
	label,
	placeholder,
}: {
	to: string;
	icon: LucideIcon;
	label: string;
	placeholder?: boolean;
}) {
	const matchRoute = useMatchRoute();
	const isActive = matchRoute({ to, fuzzy: true });

	if (placeholder) {
		return (
			<div
				className="flex items-center gap-[10px] px-3 py-[9px] rounded-[var(--r2)] text-[14px] font-medium text-t3/50 cursor-not-allowed select-none mb-[1px]"
				title={`${label} (\u0441\u043A\u043E\u0440\u043E)`}
			>
				<Icon className="w-[18px] h-[18px] opacity-50 shrink-0" />
				<span className="truncate">{label}</span>
			</div>
		);
	}

	return (
		<Link
			to={to}
			className={cn(
				"flex items-center gap-[10px] px-3 py-[9px] rounded-[var(--r2)] text-[14px] font-medium text-t2 cursor-pointer transition-colors duration-180 mb-[1px] hover:bg-[var(--hover)] hover:text-t1",
				isActive && "bg-blue/[0.06] text-t1",
			)}
		>
			<Icon className="w-[18px] h-[18px] opacity-50 shrink-0" />
			<span className="truncate">{label}</span>
		</Link>
	);
}

export const Route = createFileRoute("/_authenticated/settings")({
	component: SettingsLayout,
});
