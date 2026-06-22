import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import {
	type CrmDeal,
	type DealStage,
	formatAmount,
	getPipeline,
	STAGE_LABELS,
	updateDeal,
} from "@/api/crm";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const STAGE_COLORS: Record<DealStage, string> = {
	lead: "border-slate-400/40 bg-slate-400/10",
	qualified: "border-blue-400/40 bg-blue-400/10",
	proposal: "border-violet-400/40 bg-violet-400/10",
	negotiation: "border-amber-400/40 bg-amber-400/10",
	won: "border-emerald-400/40 bg-emerald-400/10",
	lost: "border-rose-400/40 bg-rose-400/10",
};

type DealsPipelineProps = {
	onCreateDeal: () => void;
};

/** MOD-05: Kanban pipeline сделок. */
export function DealsPipeline({ onCreateDeal }: DealsPipelineProps) {
	const navigate = useNavigate();
	const queryClient = useQueryClient();

	const { data, isLoading, isError } = useQuery({
		queryKey: ["crm", "pipeline"],
		queryFn: getPipeline,
	});

	const moveMutation = useMutation({
		mutationFn: ({ deal, stage }: { deal: CrmDeal; stage: DealStage }) =>
			updateDeal(deal.id, { stage, version: deal.version }),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["crm", "pipeline"] });
		},
		onError: () => toast.error("Не удалось переместить сделку"),
	});

	const handleDrop = (deal: CrmDeal, stage: DealStage) => {
		if (deal.stage === stage) return;
		moveMutation.mutate({ deal, stage });
	};

	if (isLoading) {
		return (
			<p className="text-sm text-t3" role="status">
				Загрузка pipeline…
			</p>
		);
	}
	if (isError || !data) {
		return (
			<p className="text-sm text-red-400" role="alert">
				Не удалось загрузить сделки
			</p>
		);
	}

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between gap-2">
				<p className="text-sm text-t2">
					{data.stages.reduce((n, s) => n + s.deals.length, 0)} сделок в
					pipeline
				</p>
				<Button type="button" size="sm" onClick={onCreateDeal}>
					Новая сделка
				</Button>
			</div>

			<div className="flex gap-3 overflow-x-auto pb-2">
				{data.stages.map((column) => (
					<section
						key={column.stage}
						aria-label={STAGE_LABELS[column.stage]}
						className="min-w-[260px] flex-shrink-0 rounded-xl border border-b1 bg-[var(--hover)]/10 p-3"
						onDragOver={(e) => e.preventDefault()}
						onDrop={(e) => {
							e.preventDefault();
							const raw = e.dataTransfer.getData("application/x-hub-deal");
							if (!raw) return;
							try {
								const deal = JSON.parse(raw) as CrmDeal;
								handleDrop(deal, column.stage);
							} catch {
								/* WS-payload некорректен — не обновляем pipeline */
							}
						}}
					>
						<h2 className="mb-3 flex items-center justify-between text-xs font-semibold uppercase tracking-wide text-t2">
							<span>{STAGE_LABELS[column.stage]}</span>
							<span className="font-normal text-t3">{column.deals.length}</span>
						</h2>
						<ul className="space-y-2">
							{column.deals.map((deal) => (
								<li key={deal.id}>
									<article
										draggable
										onDragStart={(e) => {
											e.dataTransfer.setData(
												"application/x-hub-deal",
												JSON.stringify(deal),
											);
										}}
										className={cn(
											"cursor-grab rounded-lg border p-3 active:cursor-grabbing",
											STAGE_COLORS[deal.stage],
										)}
									>
										<p className="font-medium leading-snug">{deal.title}</p>
										{deal.amount_cents != null && (
											<p className="mt-1 text-xs text-t2">
												{formatAmount(deal.amount_cents, deal.currency)}
											</p>
										)}
										<div className="mt-2 flex flex-wrap gap-1 text-[10px] text-t3">
											{deal.contact_name && <span>{deal.contact_name}</span>}
											{deal.project_key && (
												<button
													type="button"
													className="underline hover:text-t1"
													onClick={() =>
														void navigate({
															to: "/projects/$projectId",
															params: { projectId: deal.project_id as string },
														})
													}
												>
													{deal.project_key}
												</button>
											)}
										</div>
									</article>
								</li>
							))}
						</ul>
					</section>
				))}
			</div>
		</div>
	);
}
