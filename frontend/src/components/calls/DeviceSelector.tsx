import { useRoomContext } from "@livekit/components-react";
import { Mic, Volume2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

/** Информация об аудиоустройстве. */
interface DeviceInfo {
	deviceId: string;
	label: string;
}

/**
 * Панель выбора аудиоустройств (микрофон и динамик).
 * Отображается внутри Popover в панели управления звонком.
 */
export function DeviceSelector() {
	const room = useRoomContext();

	const [audioInputs, setAudioInputs] = useState<DeviceInfo[]>([]);
	const [audioOutputs, setAudioOutputs] = useState<DeviceInfo[]>([]);
	const [selectedInput, setSelectedInput] = useState<string>("");
	const [selectedOutput, setSelectedOutput] = useState<string>("");

	/** Загрузить список устройств. */
	const loadDevices = useCallback(async () => {
		try {
			const inputs = await Room_getDevices("audioinput");
			const outputs = await Room_getDevices("audiooutput");
			setAudioInputs(inputs);
			setAudioOutputs(outputs);

			// Определить текущие активные устройства
			const activeInput = room.getActiveDevice("audioinput");
			const activeOutput = room.getActiveDevice("audiooutput");
			if (activeInput) setSelectedInput(activeInput);
			else if (inputs.length > 0) setSelectedInput(inputs[0]?.deviceId ?? "");
			if (activeOutput) setSelectedOutput(activeOutput);
			else if (outputs.length > 0) setSelectedOutput(outputs[0]?.deviceId ?? "");
		} catch {
			// Устройства недоступны (нет разрешения)
		}
	}, [room]);

	/**
	 * Обёртка над navigator.mediaDevices.enumerateDevices
	 * с фильтрацией по типу.
	 */
	async function Room_getDevices(
		kind: "audioinput" | "audiooutput",
	): Promise<DeviceInfo[]> {
		const devices = await navigator.mediaDevices.enumerateDevices();
		return devices
			.filter((d) => d.kind === kind && d.deviceId !== "")
			.map((d, i) => ({
				deviceId: d.deviceId,
				label: d.label || `${kind === "audioinput" ? "Микрофон" : "Динамик"} ${i + 1}`,
			}));
	}

	useEffect(() => {
		void loadDevices();

		// Обновлять список при подключении/отключении устройств
		const handler = () => void loadDevices();
		navigator.mediaDevices.addEventListener("devicechange", handler);
		return () => {
			navigator.mediaDevices.removeEventListener("devicechange", handler);
		};
	}, [loadDevices]);

	/** Переключить микрофон. */
	const handleInputChange = async (deviceId: string) => {
		setSelectedInput(deviceId);
		try {
			await room.switchActiveDevice("audioinput", deviceId);
		} catch {
			// Ошибка переключения устройства
		}
	};

	/** Переключить динамик. */
	const handleOutputChange = async (deviceId: string) => {
		setSelectedOutput(deviceId);
		try {
			await room.switchActiveDevice("audiooutput", deviceId);
		} catch {
			// Ошибка переключения устройства
		}
	};

	return (
		<div className="flex flex-col gap-4 p-1">
			<h3 className="text-sm font-semibold text-white/90">
				Настройки устройств
			</h3>

			{/* Микрофон */}
			<div className="flex flex-col gap-1.5">
				<label className="flex items-center gap-2 text-xs font-medium text-white/60">
					<Mic className="h-3.5 w-3.5" />
					Микрофон
				</label>
				<Select
					value={selectedInput || "none"}
					onValueChange={(v) => void handleInputChange(v === "none" ? "" : v)}
				>
					<SelectTrigger className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white/90">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{audioInputs.length === 0 && (
							<SelectItem value="none">Нет доступных устройств</SelectItem>
						)}
						{audioInputs.map((d) => (
							<SelectItem key={d.deviceId} value={d.deviceId}>
								{d.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>

			{/* Динамик */}
			<div className="flex flex-col gap-1.5">
				<label className="flex items-center gap-2 text-xs font-medium text-white/60">
					<Volume2 className="h-3.5 w-3.5" />
					Динамик
				</label>
				<Select
					value={selectedOutput || "none"}
					onValueChange={(v) => void handleOutputChange(v === "none" ? "" : v)}
				>
					<SelectTrigger className="w-full rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm text-white/90">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{audioOutputs.length === 0 && (
							<SelectItem value="none">Нет доступных устройств</SelectItem>
						)}
						{audioOutputs.map((d) => (
							<SelectItem key={d.deviceId} value={d.deviceId}>
								{d.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>
		</div>
	);
}
