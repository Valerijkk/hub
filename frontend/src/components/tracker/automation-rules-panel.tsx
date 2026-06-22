import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import {
	type AutomationRule,
	createAutomationRule,
	deleteAutomationRule,
	listAutomationRules,
	updateAutomationRule,
} from "@/api/automation";
import { type Channel, getChannels } from "@/api/messenger";
import {
	listWorkflowStatuses,
	type WorkflowStatusResponse,
} from "@/api/tracker";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { toastApiError } from "@/lib/format-api-error";

interface AutomationRulesPanelProps {
	projectId: string;
}

/** PROD-01: правила workflow-автоматизации (макс. 3). */
export function AutomationRulesPanel({ projectId }: AutomationRulesPanelProps) {
	const queryClient = useQueryClient();
	const { data: rules = [], isLoading } = useQuery({
		queryKey: ["automation-rules", projectId],
		queryFn: () => listAutomationRules(projectId),
	});
	const { data: statuses = [] } = useQuery({
		queryKey: ["workflow-statuses", projectId],
		queryFn: () => listWorkflowStatuses(projectId),
	});
	const { data: channels = [] } = useQuery({
		queryKey: ["channels"],
		queryFn: () => getChannels(),
	});

	const [name, setName] = useState("");
	const [triggerType, setTriggerType] =
		useState<AutomationRule["trigger_type"]>("status_changed");
	const [statusId, setStatusId] = useState("");
	const [channelId, setChannelId] = useState("");

	const createMutation = useMutation({
		mutationFn: () =>
			createAutomationRule(projectId, {
				name: name.trim(),
				trigger_type: triggerType,
				trigger_config:
					triggerType === "status_changed" ? { to_status_id: statusId } : {},
				action_type: "post_to_channel",
				action_config: { channel_id: channelId },
			}),
		onSuccess: () => {
			toast.success("Правило создано");
			setName("");
			queryClient.invalidateQueries({
				queryKey: ["automation-rules", projectId],
			});
		},
		onError: (err) => toastApiError(err),
	});

	const toggleMutation = useMutation({
		mutationFn: (rule: AutomationRule) =>
			updateAutomationRule(projectId, rule.id, { enabled: !rule.enabled }),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: ["automation-rules", projectId],
			});
		},
		onError: (err) => toastApiError(err),
	});

	const deleteMutation = useMutation({
		mutationFn: (ruleId: string) => deleteAutomationRule(projectId, ruleId),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: ["automation-rules", projectId],
			});
		},
		onError: (err) => toastApiError(err),
	});

	const projectChannels = (channels as Channel[]).filter(
		(c) => c.project_id === projectId,
	);

	return (
		<div className="space-y-6 max-w-2xl">
			<Card>
				<CardHeader>
					<CardTitle className="text-base">Новое правило</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<p className="text-sm text-t3">
						До 3 правил на проект. Пример: при переходе в статус → сообщение в
						канал.
					</p>
					<div className="space-y-2">
						<Label htmlFor="rule-name">Название</Label>
						<Input
							id="rule-name"
							value={name}
							onChange={(e) => setName(e.target.value)}
							placeholder="Done → #general"
						/>
					</div>
					<div className="grid gap-4 sm:grid-cols-2">
						<div className="space-y-2">
							<Label>Триггер</Label>
							<Select
								value={triggerType}
								onValueChange={(v) =>
									setTriggerType(v as AutomationRule["trigger_type"])
								}
							>
								<SelectTrigger>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="status_changed">Смена статуса</SelectItem>
									<SelectItem value="issue_created">Создание задачи</SelectItem>
									<SelectItem value="issue_assigned">Назначение</SelectItem>
								</SelectContent>
							</Select>
						</div>
						{triggerType === "status_changed" && (
							<div className="space-y-2">
								<Label>Статус</Label>
								<Select value={statusId} onValueChange={setStatusId}>
									<SelectTrigger>
										<SelectValue placeholder="Выберите статус" />
									</SelectTrigger>
									<SelectContent>
										{(statuses as WorkflowStatusResponse[]).map((s) => (
											<SelectItem key={s.id} value={s.id}>
												{s.name}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</div>
						)}
					</div>
					<div className="space-y-2">
						<Label>Канал</Label>
						<Select value={channelId} onValueChange={setChannelId}>
							<SelectTrigger>
								<SelectValue placeholder="Куда отправить сообщение" />
							</SelectTrigger>
							<SelectContent>
								{projectChannels.map((ch) => (
									<SelectItem key={ch.id} value={ch.id}>
										#{ch.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
					<Button
						type="button"
						disabled={
							createMutation.isPending ||
							!name.trim() ||
							!channelId ||
							(triggerType === "status_changed" && !statusId) ||
							rules.length >= 3
						}
						onClick={() => createMutation.mutate()}
					>
						Создать правило
					</Button>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">
						Правила ({rules.length}/3)
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading && <p className="text-sm text-t3">Загрузка…</p>}
					{!isLoading && rules.length === 0 && (
						<p className="text-sm text-t3">Нет правил автоматизации</p>
					)}
					<ul className="space-y-3">
						{rules.map((rule) => (
							<li
								key={rule.id}
								className="flex items-center justify-between gap-3 rounded-lg border border-b1 p-3"
							>
								<div>
									<p className="font-medium text-sm">{rule.name}</p>
									<p className="text-xs text-t3">
										{rule.trigger_type} → {rule.action_type}
										{!rule.enabled ? " · выкл." : ""}
									</p>
								</div>
								<div className="flex gap-2">
									<Button
										type="button"
										variant="outline"
										size="sm"
										onClick={() => toggleMutation.mutate(rule)}
									>
										{rule.enabled ? "Выкл" : "Вкл"}
									</Button>
									<Button
										type="button"
										variant="destructive"
										size="sm"
										onClick={() => deleteMutation.mutate(rule.id)}
									>
										Удалить
									</Button>
								</div>
							</li>
						))}
					</ul>
				</CardContent>
			</Card>
		</div>
	);
}
