/**
 * Individual test case table row component.
 */
import { Link } from "@tanstack/react-router";
import type { TestCaseResponse } from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { formatRelativeTime } from "@/lib/format-time";
import { cn } from "@/lib/utils";

/** Color config for test case priorities. */
export const PRIORITY_CONFIG: Record<
	string,
	{ dot: string; bg: string; text: string; label: string }
> = {
	critical: {
		dot: "bg-red",
		bg: "bg-red/[0.08]",
		text: "text-red",
		label: "Critical",
	},
	high: {
		dot: "bg-orange",
		bg: "bg-orange/[0.08]",
		text: "text-orange",
		label: "High",
	},
	medium: {
		dot: "bg-amber",
		bg: "bg-amber/[0.08]",
		text: "text-amber",
		label: "Medium",
	},
	low: {
		dot: "bg-blue",
		bg: "bg-blue/[0.08]",
		text: "text-blue",
		label: "Low",
	},
	trivial: {
		dot: "bg-gray",
		bg: "bg-gray/[0.08]",
		text: "text-gray",
		label: "Trivial",
	},
};

/** Test case status config. */
export const STATUS_CONFIG: Record<
	string,
	{ label: string; bg: string; text: string }
> = {
	draft: {
		label: "Черновик",
		bg: "bg-gray/[0.08]",
		text: "text-gray",
	},
	active: {
		label: "Активный",
		bg: "bg-green/[0.08]",
		text: "text-green",
	},
	deprecated: {
		label: "Устаревший",
		bg: "bg-red/[0.08]",
		text: "text-red",
	},
};

interface CaseTableRowProps {
	tc: TestCaseResponse;
	projectId: string;
	index: number;
	isSelected: boolean;
	onToggleSelect: (id: string) => void;
}

export function CaseTableRow({
	tc,
	projectId,
	index,
	isSelected,
	onToggleSelect,
}: CaseTableRowProps) {
	const priority = (PRIORITY_CONFIG[tc.priority] ??
		PRIORITY_CONFIG.medium)!;
	const statusCfg = (STATUS_CONFIG[tc.status] ??
		STATUS_CONFIG.draft)!;
	const hasAutotest = !!(
		tc.full_name ||
		tc.class_name ||
		tc.external_id
	);

	return (
		<tr
			className={cn(
				"group cursor-pointer",
				isSelected && "bg-purple/[0.03]",
			)}
			style={{
				animation: `enterUp 350ms var(--spring) both`,
				animationDelay: `${index * 40}ms`,
			}}
		>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)]">
				<input
					type="checkbox"
					className="rounded border-b2"
					checked={isSelected}
					onChange={() => onToggleSelect(tc.id)}
				/>
			</td>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)] group-first:group-hover:shadow-[-3px_0_0_var(--purple)_inset]">
				<Link
					to="/projects/$projectId/tms/cases/$caseKey"
					params={{ projectId, caseKey: tc.case_key }}
					className="font-mono text-[12px] font-medium text-t3 hover:underline whitespace-nowrap"
				>
					{tc.case_key}
				</Link>
			</td>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)] max-w-md">
				<Link
					to="/projects/$projectId/tms/cases/$caseKey"
					params={{ projectId, caseKey: tc.case_key }}
					className="text-[13.5px] font-medium text-t1 transition-colors duration-150 group-hover:text-white line-clamp-1"
				>
					{tc.title}
				</Link>
			</td>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)]">
				<span className="inline-flex items-center gap-[7px] text-[12px] font-medium text-t2">
					<span
						className={cn(
							"w-2 h-2 rounded-[3px] transition-[box-shadow,transform] duration-200 group-hover:shadow-[0_0_8px_currentColor] group-hover:scale-[1.2]",
						)}
						style={{
							background: `var(--${tc.priority === "critical" ? "red" : tc.priority === "high" ? "orange" : tc.priority === "medium" ? "amber" : tc.priority === "low" ? "blue" : "gray"})`,
							color: `var(--${tc.priority === "critical" ? "red" : tc.priority === "high" ? "orange" : tc.priority === "medium" ? "amber" : tc.priority === "low" ? "blue" : "gray"})`,
						}}
					/>
					{priority.label}
				</span>
			</td>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)]">
				<Badge
					variant="secondary"
					className={cn(
						"text-[11px] font-semibold",
						statusCfg.bg,
						statusCfg.text,
						"border",
						tc.status === "active" ? "border-green/[0.15]" : tc.status === "deprecated" ? "border-red/[0.15]" : "border-gray/[0.15]",
					)}
				>
					{statusCfg.label}
				</Badge>
			</td>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-[180ms] group-hover:bg-[var(--hover)]">
				{hasAutotest ? (
					<span
						className="text-[16px] opacity-[0.55] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 group-hover:opacity-100 group-hover:scale-[1.15] inline-block"
						title="Автоматизирован"
					>
						🤖
					</span>
				) : (
					<span
						className="text-[16px] opacity-[0.25] inline-block"
						title="Ручной"
					>
						📝
					</span>
				)}
			</td>
			<td className="px-5 h-14 border-b border-b1 align-middle transition-colors duration-[180ms] group-hover:bg-[var(--hover)] text-right">
				<span className="text-[11px] text-t3 whitespace-nowrap">
					{formatRelativeTime(tc.updated_at)}
				</span>
			</td>
		</tr>
	);
}
