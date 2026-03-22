import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import {
	createLabel,
	createLabelGroup,
	deleteLabel,
	type LabelResponse,
	listLabelGroups,
	listLabels,
} from "@/api/tracker";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

/** Предустановленные цвета Tailwind для меток. */
const PRESET_LABEL_COLORS = [
	"#ef4444",
	"#f97316",
	"#eab308",
	"#22c55e",
	"#06b6d4",
	"#3b82f6",
	"#8b5cf6",
	"#ec4899",
	"#6b7280",
	"#14b8a6",
];

/**
 * Страница управления метками проекта.
 * Отображает метки, сгруппированные по группам, с CRUD-операциями.
 * Позволяет создавать метки с цветами и группы меток.
 */
function LabelsSettingsPage() {
	const { projectId } = Route.useParams();
	const queryClient = useQueryClient();

	// Метки
	const { data: labels, isLoading: labelsLoading } = useQuery({
		queryKey: ["labels", projectId],
		queryFn: () => listLabels(projectId),
	});

	// Группы
	const { data: groups } = useQuery({
		queryKey: ["label-groups", projectId],
		queryFn: () => listLabelGroups(projectId),
	});

	// Форма новой метки
	const [labelName, setLabelName] = useState("");
	const [labelColor, setLabelColor] = useState(PRESET_LABEL_COLORS[0]!);
	const [labelGroupId, setLabelGroupId] = useState("");

	// Форма новой группы
	const [groupName, setGroupName] = useState("");

	const createLabelMutation = useMutation({
		mutationFn: () =>
			createLabel(projectId, {
				name: labelName.trim(),
				color: labelColor,
				group_id: labelGroupId || undefined,
			}),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["labels", projectId],
			});
			void queryClient.invalidateQueries({
				queryKey: ["label-groups", projectId],
			});
			setLabelName("");
			toast.success("Метка создана");
		},
		onError: () => toast.error("Не удалось создать метку"),
	});

	const deleteLabelMutation = useMutation({
		mutationFn: (labelId: string) => deleteLabel(projectId, labelId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["labels", projectId],
			});
			void queryClient.invalidateQueries({
				queryKey: ["label-groups", projectId],
			});
			toast.success("Метка удалена");
		},
		onError: () => toast.error("Не удалось удалить метку"),
	});

	const createGroupMutation = useMutation({
		mutationFn: () => createLabelGroup(projectId, { name: groupName.trim() }),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["label-groups", projectId],
			});
			setGroupName("");
			toast.success("Группа создана");
		},
		onError: () => toast.error("Не удалось создать группу"),
	});

	const handleCreateLabel = (e: React.FormEvent) => {
		e.preventDefault();
		if (!labelName.trim()) {
			toast.error("Введите название метки");
			return;
		}
		createLabelMutation.mutate();
	};

	const handleCreateGroup = (e: React.FormEvent) => {
		e.preventDefault();
		if (!groupName.trim()) {
			toast.error("Введите название группы");
			return;
		}
		createGroupMutation.mutate();
	};

	// Разделяем метки на сгруппированные и без группы
	const ungroupedLabels = labels?.filter((l) => !l.group_name) ?? [];

	return (
		<div className="mx-auto max-w-3xl space-y-6">
			<div className="flex items-center justify-between">
				<h1 className="text-2xl font-bold">Метки</h1>
				<Button variant="outline" asChild>
					<Link to="/projects/$projectId" params={{ projectId }}>
						Назад
					</Link>
				</Button>
			</div>

			{/* Метки без группы */}
			<Card>
				<CardHeader>
					<CardTitle>Без группы</CardTitle>
				</CardHeader>
				<CardContent>
					{labelsLoading ? (
						<p className="text-sm text-muted-foreground">Загрузка...</p>
					) : ungroupedLabels.length > 0 ? (
						<div className="space-y-2">
							{ungroupedLabels.map((label) => (
								<LabelRow
									key={label.id}
									label={label}
									onDelete={() => {
										if (window.confirm("Удалить метку?")) {
											deleteLabelMutation.mutate(label.id);
										}
									}}
								/>
							))}
						</div>
					) : (
						<p className="text-sm text-muted-foreground">
							Нет меток без группы
						</p>
					)}
				</CardContent>
			</Card>

			{/* Группы меток */}
			{groups?.map((group) => (
				<Card key={group.id}>
					<CardHeader>
						<CardTitle>{group.name}</CardTitle>
					</CardHeader>
					<CardContent>
						{group.labels.length > 0 ? (
							<div className="space-y-2">
								{group.labels.map((label) => (
									<LabelRow
										key={label.id}
										label={label}
										onDelete={() => {
											if (window.confirm("Удалить метку?")) {
												deleteLabelMutation.mutate(label.id);
											}
										}}
									/>
								))}
							</div>
						) : (
							<p className="text-sm text-muted-foreground">
								Нет меток в группе
							</p>
						)}
					</CardContent>
				</Card>
			))}

			<Separator />

			{/* Добавить метку */}
			<Card>
				<CardHeader>
					<CardTitle>Добавить метку</CardTitle>
				</CardHeader>
				<CardContent>
					<form onSubmit={handleCreateLabel} className="space-y-3">
						<div className="grid grid-cols-2 gap-3">
							<div className="space-y-1">
								<Label htmlFor="label_name">Название</Label>
								<Input
									id="label_name"
									value={labelName}
									onChange={(e) => setLabelName(e.target.value)}
									placeholder="Название метки"
								/>
							</div>
							<div className="space-y-1">
								<Label htmlFor="label_group">Группа (опционально)</Label>
								<select
									id="label_group"
									className="w-full rounded-md border bg-background px-3 py-2 text-sm"
									value={labelGroupId}
									onChange={(e) => setLabelGroupId(e.target.value)}
								>
									<option value="">Без группы</option>
									{groups?.map((g) => (
										<option key={g.id} value={g.id}>
											{g.name}
										</option>
									))}
								</select>
							</div>
						</div>
						<div className="space-y-1">
							<Label>Цвет</Label>
							<div className="flex gap-1">
								{PRESET_LABEL_COLORS.map((color) => (
									<button
										key={color}
										type="button"
										className={`h-6 w-6 rounded-full border-2 ${labelColor === color ? "border-foreground" : "border-transparent"}`}
										style={{ backgroundColor: color }}
										onClick={() => setLabelColor(color)}
									/>
								))}
							</div>
						</div>
						<div className="flex items-center gap-3">
							<Badge
								variant="outline"
								style={{ borderColor: labelColor, color: labelColor }}
							>
								{labelName || "Превью"}
							</Badge>
							<Button
								type="submit"
								size="sm"
								disabled={createLabelMutation.isPending}
							>
								{createLabelMutation.isPending
									? "Создание..."
									: "Создать метку"}
							</Button>
						</div>
					</form>
				</CardContent>
			</Card>

			{/* Добавить группу */}
			<Card>
				<CardHeader>
					<CardTitle>Добавить группу</CardTitle>
				</CardHeader>
				<CardContent>
					<form onSubmit={handleCreateGroup} className="flex gap-3">
						<Input
							value={groupName}
							onChange={(e) => setGroupName(e.target.value)}
							placeholder="Название группы"
							className="flex-1"
						/>
						<Button
							type="submit"
							size="sm"
							disabled={createGroupMutation.isPending}
						>
							{createGroupMutation.isPending ? "Создание..." : "Создать группу"}
						</Button>
					</form>
				</CardContent>
			</Card>
		</div>
	);
}

/** Пропсы компонента LabelRow. */
interface LabelRowProps {
	/** Данные метки. */
	label: LabelResponse;
	/** Колбэк удаления метки. */
	onDelete: () => void;
}

/** Строка метки в списке с цветовой индикацией и кнопкой удаления. */
function LabelRow({ label, onDelete }: LabelRowProps) {
	return (
		<div className="flex items-center justify-between rounded-md border p-2">
			<div className="flex items-center gap-2">
				<div
					className="h-4 w-4 rounded-full"
					style={{ backgroundColor: label.color }}
				/>
				<span className="text-sm font-medium">{label.name}</span>
			</div>
			<Button
				variant="ghost"
				size="sm"
				className="h-6 px-2 text-xs text-muted-foreground hover:text-destructive"
				onClick={onDelete}
			>
				Удалить
			</Button>
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/settings/labels",
)({
	component: LabelsSettingsPage,
});
