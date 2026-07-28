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
	const isSelected = selectedId === suite.id;

	return (
		<div>
			<div
				className={cn(
					"group relative flex w-full items-center gap-2 rounded-[var(--r2)] px-[10px] py-2 text-[13px] font-medium text-t2",
					"hover:bg-[var(--hover)] hover:text-t1",
					isSelected &&
						"bg-purple/[0.05] text-white shadow-[0_0_0_1px_rgba(167,139,250,0.08)_inset]",
				)}
				style={{ paddingLeft: `${depth * 26 + 10}px` }}
			>
				{isSelected && (
					<span className="absolute left-0 top-0 bottom-0 w-10 bg-gradient-to-r from-purple/[0.06] to-transparent rounded-[var(--r2)] pointer-events-none" />
				)}
				{hasChildren ? (
					<button
						type="button"
						className={cn(
							"w-4 h-4 text-[8px] text-t4 transition-transform duration-[250ms] shrink-0 flex items-center justify-center border-0 bg-transparent p-0 cursor-pointer",
							expanded && "rotate-90",
						)}
						aria-label={expanded ? "Свернуть сьют" : "Развернуть сьют"}
						aria-expanded={expanded}
						onClick={() => setExpanded(!expanded)}
					>
						<svg
							aria-hidden="true"
							className="h-[10px] w-[10px]"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={3}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M9 5l7 7-7 7"
							/>
						</svg>
					</button>
				) : (
					<span className="w-4 shrink-0" aria-hidden="true" />
				)}
				<button
					type="button"
					className="flex min-w-0 flex-1 items-center gap-2 border-0 bg-transparent p-0 text-left text-[13px] font-medium text-inherit cursor-pointer"
					onClick={() => onSelect(isSelected ? undefined : suite.id)}
				>
					{hasChildren ? (
						<span
							className={cn(
								"shrink-0 flex items-center justify-center w-[22px] h-[22px] rounded-[7px] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
								"group-hover:scale-110 group-hover:shadow-[0_0_10px_rgba(167,139,250,0.1)]",
							)}
							style={{
								background: isSelected
									? "rgba(167,139,250,0.1)"
									: "rgba(255,255,255,0.035)",
								color: isSelected ? "var(--purple)" : "var(--text-3)",
							}}
						>
							<svg
								aria-hidden="true"
								className="h-3 w-3"
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
						</span>
					) : (
						<span
							className="w-[7px] h-[7px] rounded-full shrink-0 ml-[1px] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 group-hover:shadow-[0_0_8px_currentColor] group-hover:scale-[1.3]"
							style={{
								background: isSelected ? "var(--purple)" : "var(--text-3)",
								color: isSelected ? "var(--purple)" : "var(--text-3)",
							}}
						/>
					)}
					<span className="truncate">{suite.name}</span>
				</button>
			</div>
			{expanded && hasChildren && (
				<div>
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
	onCreateSuite,
	total,
}: {
	suites?: SuiteResponse[];
	selectedSuiteId?: string;
	onSelectSuite: (id: string | undefined) => void;
	/** Если не передан — кнопка «+» скрыта (нет мёртвого no-op). */
	onCreateSuite?: () => void;
	total: number;
}) {
	return (
		<div className="w-full md:w-[290px] md:min-w-[290px] border-r border-b1 flex flex-col bg-void animate-[enterRight_400ms_var(--spring)_both]">
			<div className="px-[18px] py-4 border-b border-b1 flex items-center justify-between">
				<span className="text-[10.5px] font-bold uppercase tracking-[0.08em] text-t4">
					Тест-сьюты
				</span>
				{onCreateSuite ? (
					<button
						type="button"
						className="w-7 h-7 flex items-center justify-center rounded-[var(--r2)] text-t3 text-[18px] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 hover:bg-[var(--active)] hover:text-t1 hover:scale-110"
						aria-label="Создать сьют"
						onClick={onCreateSuite}
					>
						+
					</button>
				) : null}
			</div>
			<div className="flex-1 overflow-y-auto p-2">
				<button
					type="button"
					className={cn(
						"group flex w-full items-center gap-2 rounded-[var(--r2)] px-[10px] py-2 text-left text-[13px] font-medium cursor-pointer transition-colors duration-[180ms] text-t2 relative mb-1",
						"hover:bg-[var(--hover)] hover:text-t1",
						!selectedSuiteId &&
							"bg-purple/[0.05] text-white shadow-[0_0_0_1px_rgba(167,139,250,0.08)_inset]",
					)}
					onClick={() => onSelectSuite(undefined)}
				>
					{!selectedSuiteId && (
						<span className="absolute left-0 top-0 bottom-0 w-10 bg-gradient-to-r from-purple/[0.06] to-transparent rounded-[var(--r2)] pointer-events-none" />
					)}
					<svg
						aria-hidden="true"
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
					<span className="text-[10.5px] font-semibold text-t4 bg-white/[0.035] px-2 py-[1px] rounded-full transition-colors duration-200 group-hover:bg-white/[0.06] group-hover:text-t3">
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
