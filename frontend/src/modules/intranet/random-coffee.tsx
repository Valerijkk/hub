import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Coffee } from "lucide-react";
import { toast } from "sonner";
import {
	getCoffeeEnrollment,
	getMyCoffeeMatch,
	listCoffeeRounds,
	runCoffeeMatching,
	setCoffeeEnrollment,
} from "@/api/intranet";
import { Button } from "@/components/ui/button";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";

/** V4-05: Random Coffee — opt-in + weekly matching. */
export function RandomCoffeePanel() {
	const user = useAuthStore((s) => s.user);
	const isAdmin = isSystemAdmin(user?.system_role);
	const queryClient = useQueryClient();

	const { data: enrollment } = useQuery({
		queryKey: ["intranet", "coffee", "enrollment"],
		queryFn: getCoffeeEnrollment,
	});

	const { data: match } = useQuery({
		queryKey: ["intranet", "coffee", "match"],
		queryFn: getMyCoffeeMatch,
		enabled: enrollment?.is_active === true,
	});

	const { data: roundsData } = useQuery({
		queryKey: ["intranet", "coffee", "rounds"],
		queryFn: listCoffeeRounds,
		enabled: isAdmin,
	});
	const rounds = roundsData?.items ?? [];

	const enrollMutation = useMutation({
		mutationFn: (active: boolean) => setCoffeeEnrollment(active),
		onSuccess: (_, active) => {
			void queryClient.invalidateQueries({ queryKey: ["intranet", "coffee"] });
			toast.success(active ? "Вы в Random Coffee" : "Участие отключено");
		},
		onError: () => toast.error("Не удалось изменить участие в Random Coffee"),
	});

	const runMutation = useMutation({
		mutationFn: runCoffeeMatching,
		onSuccess: (data) => {
			void queryClient.invalidateQueries({ queryKey: ["intranet", "coffee"] });
			toast.success(
				`Раунд: ${data.pair_count} пар${data.unpaired_count ? `, ${data.unpaired_count} без пары` : ""}`,
			);
		},
		onError: () => toast.error("Не удалось запустить matching"),
	});

	const isActive = enrollment?.is_active === true;

	return (
		<div className="space-y-6">
			<section className="rounded-xl border border-b1 bg-[var(--hover)]/10 p-4 space-y-4">
				<div className="flex items-center gap-2">
					<Coffee className="h-5 w-5 text-amber-400" aria-hidden="true" />
					<h2 className="text-lg font-semibold">Random Coffee</h2>
				</div>
				<p className="text-sm text-t2">
					Еженедельный random matching — познакомьтесь с коллегами из других
					отделов за чашкой кофе.
				</p>
				<label className="flex items-center gap-3 text-sm cursor-pointer">
					<input
						type="checkbox"
						checked={isActive}
						onChange={(e) => enrollMutation.mutate(e.target.checked)}
						disabled={enrollMutation.isPending}
					/>
					Участвовать в следующих раундах
				</label>
			</section>

			{isActive && match?.partner && (
				<section className="rounded-xl border border-amber-500/30 bg-amber-500/[0.05] p-4 space-y-2">
					<p className="text-sm font-medium text-amber-400">
						Ваш напарник на этой неделе
					</p>
					<p className="text-lg font-semibold">{match.partner.display_name}</p>
					{match.partner.department && (
						<p className="text-sm text-t2">{match.partner.department}</p>
					)}
					{match.round_at && (
						<p className="text-xs text-t3">
							Раунд: {new Date(match.round_at).toLocaleString("ru-RU")}
						</p>
					)}
				</section>
			)}

			{isActive && match && !match.partner && match.unpaired && (
				<p className="text-sm text-t3">
					В последнем раунде не хватило пары — вы в резерве на следующий раз.
				</p>
			)}

			{isAdmin && (
				<section className="space-y-3 rounded-xl border border-b1 p-4">
					<p className="text-sm font-medium">Admin: запуск раунда</p>
					<Button
						type="button"
						onClick={() => runMutation.mutate()}
						disabled={runMutation.isPending}
					>
						Сформировать пары
					</Button>
					{rounds.length > 0 && (
						<ul className="text-sm text-t2 space-y-1">
							{rounds.slice(0, 3).map((r) => (
								<li key={r.id}>
									{new Date(r.created_at).toLocaleDateString("ru-RU")} —{" "}
									{r.pair_count} пар
								</li>
							))}
						</ul>
					)}
				</section>
			)}
		</div>
	);
}
