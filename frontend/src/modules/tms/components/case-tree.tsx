/**
 * Дерево тест-сьютов (папок) для навигации по тест-кейсам.
 */
import { useState } from "react";
import type { SuiteResponse } from "@/api/tms";
import { cn } from "@/lib/utils";

/**
 * Рекурсивный компонент дерева папок.
 */
/** Рекурсивный элемент дерева сьютов с раскрытием/свёртыванием и выбором. */
export function SuiteTreeItem({
	suite,
	selectedId,
	onSelect,
	depth = 0,
}: {
	suite: SuiteResponse;
	selectedId?: string;
	onSelect: (id: string | undefined) => void;
	depth?: number;
}) {
	const [expanded, setExpanded] = useState(depth < 2);
	const hasChildren = suite.children && suite.children.length > 0;

	return (
		<div>
			<button
				className={cn(
					"group flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-all duration-150",
					"hover:bg-muted/80",
					selectedId === suite.id && "bg-primary/10 text-primary font-medium",
				)}
				style={{ paddingLeft: `${depth * 16 + 8}px` }}
				onClick={() => onSelect(selectedId === suite.id ? undefined : suite.id)}
			>
				{hasChildren ? (
					<span
						className={cn(
							"mr-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded text-[10px] transition-colors",
							"text-muted-foreground hover:bg-muted-foreground/10",
						)}
						onClick={(e) => {
							e.stopPropagation();
							setExpanded(!expanded);
						}}
					>
						<svg
							className={cn(
								"h-3 w-3 transition-transform duration-200",
								expanded && "rotate-90",
							)}
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M9 5l7 7-7 7"
							/>
						</svg>
					</span>
				) : (
					<span className="mr-0.5 w-4 shrink-0" />
				)}
				<svg
					className={cn(
						"h-3.5 w-3.5 shrink-0 transition-colors",
						selectedId === suite.id ? "text-primary" : "text-muted-foreground",
					)}
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
					/>
				</svg>
				<span className="truncate flex-1">{suite.name}</span>
			</button>
			{expanded && hasChildren && (
				<div className="animate-fade-in">
					{suite.children?.map((child) => (
						<SuiteTreeItem
							key={child.id}
							suite={child}
							selectedId={selectedId}
							onSelect={onSelect}
							depth={depth + 1}
						/>
					))}
				</div>
			)}
		</div>
	);
}

/**
 * Боковая панель с деревом сьютов.
 */
/** Боковая панель с деревом тест-сьютов и кнопкой "Все кейсы". */
export function CaseTreePanel({
	suites,
	selectedSuiteId,
	onSelectSuite,
	total,
}: {
	suites?: SuiteResponse[];
	selectedSuiteId?: string;
	onSelectSuite: (id: string | undefined) => void;
	total: number;
}) {
	return (
		<div className="w-60 shrink-0 flex flex-col rounded-xl border bg-card/50 shadow-card overflow-hidden">
			<div className="flex items-center justify-between px-3 py-2.5 border-b bg-muted/30">
				<span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
					Тест-сьюты
				</span>
			</div>
			<div className="flex-1 overflow-y-auto p-1.5">
				<button
					className={cn(
						"mb-1 flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-all duration-150",
						"hover:bg-muted/80",
						!selectedSuiteId && "bg-primary/10 text-primary font-medium",
					)}
					onClick={() => onSelectSuite(undefined)}
				>
					<svg
						className="h-3.5 w-3.5 shrink-0"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
						/>
					</svg>
					<span className="flex-1">Все кейсы</span>
					<span className="text-[10px] text-muted-foreground tabular-nums">
						{total}
					</span>
				</button>
				{suites?.map((suite) => (
					<SuiteTreeItem
						key={suite.id}
						suite={suite}
						selectedId={selectedSuiteId}
						onSelect={onSelectSuite}
					/>
				))}
			</div>
		</div>
	);
}
