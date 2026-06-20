import type { TestCaseResponse, TestStep } from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

/** Цвета BDD-ключевых слов для бейджей (Given, When, Then, And, But). */
const BDD_KEYWORD_COLORS: Record<string, string> = {
	Given: "text-blue bg-blue/[0.1] shadow-[0_0_8px_rgba(91,156,245,0.06)]",
	When: "text-amber bg-amber/[0.1] shadow-[0_0_8px_rgba(251,191,36,0.06)]",
	Then: "text-green bg-green/[0.1] shadow-[0_0_8px_rgba(52,211,153,0.06)]",
	And: "text-gray bg-gray/[0.1]",
	But: "text-red bg-red/[0.1]",
};

/** Пропсы редактора шагов в режиме редактирования. */
interface StepEditorEditProps {
	editing: true;
	editSteps: TestStep[];
	onAddStep: () => void;
	onRemoveStep: (index: number) => void;
	onUpdateStep: (
		index: number,
		field: "action" | "expected_result",
		value: string,
	) => void;
	onMoveStep: (index: number, direction: "up" | "down") => void;
}

/** Пропсы отображения шагов в режиме чтения. */
interface StepEditorReadProps {
	editing: false;
	tc: TestCaseResponse;
	onStartEditing: () => void;
}

export type StepEditorProps = StepEditorEditProps | StepEditorReadProps;

/** Редактор и отображение шагов тест-кейса (режим редактирования и чтения). */
export function StepEditor(props: StepEditorProps) {
	if (props.editing) {
		return (
			<div>
				<div className="flex items-center justify-between mb-3">
					<span className="text-sm font-semibold">Шаги теста</span>
					<Button
						variant="outline"
						size="sm"
						onClick={props.onAddStep}
						className="h-7 text-xs"
					>
						<svg
							aria-hidden="true"
							className="mr-1 h-3 w-3"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M12 4v16m8-8H4"
							/>
						</svg>
						Добавить шаг
					</Button>
				</div>
				{props.editSteps.length === 0 ? (
					<div className="text-center py-8 text-sm text-t2 border rounded-lg border-dashed">
						Нет шагов. Нажмите "Добавить шаг".
					</div>
				) : (
					<div className="space-y-2">
						{props.editSteps.map((step, idx) => (
							<div
								key={step._uid ?? `step-fallback-${step.order}`}
								className="flex gap-2 items-start rounded-lg border p-3 bg-card/50 hover:bg-card transition-colors"
							>
								<div className="flex flex-col items-center gap-0.5 shrink-0 mt-1">
									<button
										type="button"
										disabled={idx === 0}
										onClick={() => props.onMoveStep(idx, "up")}
										className={cn(
											"flex h-5 w-5 items-center justify-center rounded text-[10px] leading-none transition-colors",
											idx === 0
												? "text-t2/30 cursor-not-allowed"
												: "text-t2 hover:text-t1 hover:bg-raised",
										)}
										title="Переместить вверх"
									>
										▲
									</button>
									<span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary">
										{step.order}
									</span>
									<button
										type="button"
										disabled={idx === props.editSteps.length - 1}
										onClick={() => props.onMoveStep(idx, "down")}
										className={cn(
											"flex h-5 w-5 items-center justify-center rounded text-[10px] leading-none transition-colors",
											idx === props.editSteps.length - 1
												? "text-t2/30 cursor-not-allowed"
												: "text-t2 hover:text-t1 hover:bg-raised",
										)}
										title="Переместить вниз"
									>
										▼
									</button>
								</div>
								<div className="flex-1 space-y-2">
									<Input
										placeholder="Действие"
										value={step.action}
										onChange={(e) =>
											props.onUpdateStep(idx, "action", e.target.value)
										}
										className="text-sm"
									/>
									<Input
										placeholder="Ожидаемый результат"
										value={step.expected_result ?? ""}
										onChange={(e) =>
											props.onUpdateStep(idx, "expected_result", e.target.value)
										}
										className="text-sm"
									/>
								</div>
								<Button
									variant="ghost"
									size="sm"
									className="mt-1 h-6 w-6 p-0 text-t2 hover:text-destructive"
									onClick={() => props.onRemoveStep(idx)}
								>
									<svg
										aria-hidden="true"
										className="h-3.5 w-3.5"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={2}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M6 18L18 6M6 6l12 12"
										/>
									</svg>
								</Button>
							</div>
						))}
					</div>
				)}
			</div>
		);
	}

	// Read mode
	const { tc, onStartEditing } = props;

	if (tc.steps && tc.steps.length > 0) {
		return (
			<div>
				<table className="w-full border-collapse mt-[14px]">
					<thead>
						<tr>
							<th className="text-left px-3 py-[10px] text-[10px] font-bold text-t4 border-b border-b2 uppercase tracking-[0.06em] w-8">
								#
							</th>
							{tc.case_type === "bdd" && (
								<th className="text-left px-3 py-[10px] text-[10px] font-bold text-t4 border-b border-b2 uppercase tracking-[0.06em] w-24">
									Keyword
								</th>
							)}
							<th className="text-left px-3 py-[10px] text-[10px] font-bold text-t4 border-b border-b2 uppercase tracking-[0.06em]">
								Действие
							</th>
							<th className="text-left px-3 py-[10px] text-[10px] font-bold text-t4 border-b border-b2 uppercase tracking-[0.06em]">
								Ожидаемый результат
							</th>
						</tr>
					</thead>
					<tbody>
						{tc.steps.map((step) => (
							<tr
								key={step.order}
								className="transition-[background] duration-150 hover:bg-[var(--hover)]"
							>
								<td className="px-3 py-[14px] border-b border-b1 text-[13px] align-top leading-[1.65] font-bold text-t4 w-8">
									{step.order}
								</td>
								{tc.case_type === "bdd" && (
									<td className="px-3 py-[14px] border-b border-b1 text-[13px] align-top leading-[1.65]">
										{step.keyword && (
											<span
												className={cn(
													"inline-block px-[10px] py-[2px] rounded-[5px] text-[10px] font-bold tracking-[0.05em] mr-[6px]",
													BDD_KEYWORD_COLORS[step.keyword] ??
														"text-gray bg-gray/[0.1]",
												)}
											>
												{step.keyword}
											</span>
										)}
									</td>
								)}
								<td className="px-3 py-[14px] border-b border-b1 text-[13px] align-top leading-[1.65] text-t2">
									{step.action}
								</td>
								<td className="px-3 py-[14px] border-b border-b1 text-[13px] align-top leading-[1.65] text-t2">
									{step.expected_result || (
										<span className="italic text-t4">--</span>
									)}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		);
	}

	return (
		<div className="flex flex-col items-center justify-center py-12 text-center border rounded-xl border-dashed">
			<svg
				aria-hidden="true"
				className="h-10 w-10 text-t2/20 mb-3"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				strokeWidth={1.5}
			>
				<path
					strokeLinecap="round"
					strokeLinejoin="round"
					d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
				/>
			</svg>
			<p className="text-sm text-t2 mb-2">Шаги не заданы</p>
			<Button variant="outline" size="sm" onClick={onStartEditing}>
				Добавить шаги
			</Button>
		</div>
	);
}
