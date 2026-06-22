import { useEffect } from "react";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { useAIStore } from "@/stores/ai-store";

/** ModelPicker — выбор LLM из registry (Phase 17 W3). */
export function ModelPicker() {
	const models = useAIStore((s) => s.models);
	const selectedModel = useAIStore((s) => s.selectedModel);
	const setSelectedModel = useAIStore((s) => s.saveSelectedModel);
	const loadModels = useAIStore((s) => s.loadModels);

	useEffect(() => {
		void loadModels();
	}, [loadModels]);

	if (models.length === 0) {
		return (
			<span className="text-[11px] text-t3">Нет моделей — Admin → AI</span>
		);
	}

	return (
		<Select
			value={selectedModel ?? undefined}
			onValueChange={(v) => void setSelectedModel(v)}
		>
			<SelectTrigger className="h-8 text-xs w-[180px]">
				<SelectValue placeholder="Модель" />
			</SelectTrigger>
			<SelectContent>
				{models.map((m) => (
					<SelectItem key={m.id} value={m.id}>
						{m.provider_name}: {m.id}
						{m.ru_resident ? " · RU" : ""}
					</SelectItem>
				))}
			</SelectContent>
		</Select>
	);
}
