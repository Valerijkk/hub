import { useRoomContext } from "@livekit/components-react";
import { Camera, Mic, Volume2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

/** Ключи localStorage для настроек звонка. */
const LS_KEY_MIC_OFF = "hub:call:joinMicOff";
const LS_KEY_CAM_OFF = "hub:call:joinCamOff";
const LS_KEY_NOISE_SUPPRESSION = "hub:call:noiseSuppression";

/** Информация об устройстве ввода/вывода. */
interface DeviceInfo {
	deviceId: string;
	label: string;
}

/** Пропсы компонента CallSettingsModal. */
interface CallSettingsModalProps {
	/** Модальное окно открыто. */
	open: boolean;
	/** Коллбэк при закрытии. */
	onOpenChange: (open: boolean) => void;
}

/**
 * Возвращает сохранённые настройки звонка из localStorage.
 * Используется при подключении к звонку для определения
 * начального состояния микрофона и камеры.
 */
export function getCallPreferences() {
	return {
		joinMicOff: localStorage.getItem(LS_KEY_MIC_OFF) === "true",
		joinCamOff: localStorage.getItem(LS_KEY_CAM_OFF) === "true",
		noiseSuppression: localStorage.getItem(LS_KEY_NOISE_SUPPRESSION) !== "false",
	};
}

/**
 * Получить список устройств определённого типа через MediaDevices API.
 * Возвращает пустой массив если нет разрешений.
 */
async function enumerateDevices(
	kind: "audioinput" | "audiooutput" | "videoinput",
): Promise<DeviceInfo[]> {
	try {
		const devices = await navigator.mediaDevices.enumerateDevices();
		const labels: Record<string, string> = {
			audioinput: "Микрофон",
			audiooutput: "Динамик",
			videoinput: "Камера",
		};
		return devices
			.filter((d) => d.kind === kind && d.deviceId !== "")
			.map((d, i) => ({
				deviceId: d.deviceId,
				label: d.label || `${labels[kind]} ${i + 1}`,
			}));
	} catch {
		return [];
	}
}

/**
 * Модальное окно настроек звонка в стиле Telemost.
 * Две вкладки: Звук (динамик, микрофон, тогглы) и Видео (камера, тоггл).
 * Настройки сохраняются в localStorage и применяются к текущей комнате LiveKit.
 */
export function CallSettingsModal({
	open,
	onOpenChange,
}: CallSettingsModalProps) {
	const room = useRoomContext();

	// Списки устройств
	const [audioInputs, setAudioInputs] = useState<DeviceInfo[]>([]);
	const [audioOutputs, setAudioOutputs] = useState<DeviceInfo[]>([]);
	const [videoInputs, setVideoInputs] = useState<DeviceInfo[]>([]);

	// Выбранные устройства
	const [selectedInput, setSelectedInput] = useState("");
	const [selectedOutput, setSelectedOutput] = useState("");
	const [selectedCamera, setSelectedCamera] = useState("");

	// Настройки
	const [joinMicOff, setJoinMicOff] = useState(
		() => localStorage.getItem(LS_KEY_MIC_OFF) === "true",
	);
	const [joinCamOff, setJoinCamOff] = useState(
		() => localStorage.getItem(LS_KEY_CAM_OFF) === "true",
	);
	const [noiseSuppression, setNoiseSuppression] = useState(
		() => localStorage.getItem(LS_KEY_NOISE_SUPPRESSION) !== "false",
	);

	/** Загрузить все доступные устройства и определить активные. */
	const loadDevices = useCallback(async () => {
		const [inputs, outputs, cameras] = await Promise.all([
			enumerateDevices("audioinput"),
			enumerateDevices("audiooutput"),
			enumerateDevices("videoinput"),
		]);

		setAudioInputs(inputs);
		setAudioOutputs(outputs);
		setVideoInputs(cameras);

		// Определяем активные устройства из LiveKit room
		const activeIn = room.getActiveDevice("audioinput");
		const activeOut = room.getActiveDevice("audiooutput");
		const activeCam = room.getActiveDevice("videoinput");

		setSelectedInput(activeIn || inputs[0]?.deviceId || "");
		setSelectedOutput(activeOut || outputs[0]?.deviceId || "");
		setSelectedCamera(activeCam || cameras[0]?.deviceId || "");
	}, [room]);

	useEffect(() => {
		if (!open) return;
		void loadDevices();

		const handler = () => void loadDevices();
		navigator.mediaDevices.addEventListener("devicechange", handler);
		return () => {
			navigator.mediaDevices.removeEventListener("devicechange", handler);
		};
	}, [open, loadDevices]);

	/** Переключить аудио-вход (микрофон). */
	const handleInputChange = async (deviceId: string) => {
		setSelectedInput(deviceId);
		try {
			await room.switchActiveDevice("audioinput", deviceId);
		} catch {
			// Устройство может быть недоступно
		}
	};

	/** Переключить аудио-выход (динамик). */
	const handleOutputChange = async (deviceId: string) => {
		setSelectedOutput(deviceId);
		try {
			await room.switchActiveDevice("audiooutput", deviceId);
		} catch {
			// Устройство может быть недоступно
		}
	};

	/** Переключить камеру. */
	const handleCameraChange = async (deviceId: string) => {
		setSelectedCamera(deviceId);
		try {
			await room.switchActiveDevice("videoinput", deviceId);
		} catch {
			// Устройство может быть недоступно
		}
	};

	/** Сохранить тоггл "подключаться с выключенным микрофоном". */
	const handleJoinMicOffToggle = () => {
		const next = !joinMicOff;
		setJoinMicOff(next);
		localStorage.setItem(LS_KEY_MIC_OFF, String(next));
	};

	/** Сохранить тоггл "подключаться с выключенной камерой". */
	const handleJoinCamOffToggle = () => {
		const next = !joinCamOff;
		setJoinCamOff(next);
		localStorage.setItem(LS_KEY_CAM_OFF, String(next));
	};

	/** Сохранить тоггл "шумоподавление" (косметический). */
	const handleNoiseSuppressionToggle = () => {
		const next = !noiseSuppression;
		setNoiseSuppression(next);
		localStorage.setItem(LS_KEY_NOISE_SUPPRESSION, String(next));
	};

	/** Стили для Radix Select trigger в тёмной теме настроек. */
	const selectTriggerClasses =
		"w-full rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-white/90 transition-all hover:bg-white/10 focus:border-blue-500/40 focus:ring-1 focus:ring-blue-500/30 h-auto";

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="border-white/10 bg-[#1a1f2e] text-white sm:max-w-lg p-0 gap-0">
				<DialogHeader className="px-6 pt-6 pb-2">
					<DialogTitle className="text-lg font-semibold">
						Настройки
					</DialogTitle>
				</DialogHeader>

				<Tabs defaultValue="audio" className="px-6 pb-6">
					<TabsList className="grid w-full grid-cols-2 bg-white/5 border border-white/10 rounded-lg">
						<TabsTrigger
							value="audio"
							className="data-[state=active]:bg-white/10 data-[state=active]:text-white text-white/60 rounded-md"
						>
							<Volume2 className="h-4 w-4 mr-2" />
							Звук
						</TabsTrigger>
						<TabsTrigger
							value="video"
							className="data-[state=active]:bg-white/10 data-[state=active]:text-white text-white/60 rounded-md"
						>
							<Camera className="h-4 w-4 mr-2" />
							Видео
						</TabsTrigger>
					</TabsList>

					{/* Вкладка "Звук" */}
					<TabsContent value="audio" className="mt-4 space-y-5">
						{/* Динамик */}
						<div className="space-y-2">
							<label className="flex items-center gap-2 text-xs font-medium text-white/60 uppercase tracking-wider">
								<Volume2 className="h-3.5 w-3.5" />
								Динамик
							</label>
							<Select
								value={selectedOutput}
								onValueChange={(v) => void handleOutputChange(v)}
							>
								<SelectTrigger className={selectTriggerClasses}>
									<SelectValue
										placeholder="Нет доступных устройств"
									/>
								</SelectTrigger>
								<SelectContent>
									{audioOutputs.length === 0 ? (
										<SelectItem value="__none" disabled>
											Нет доступных устройств
										</SelectItem>
									) : (
										audioOutputs.map((d) => (
											<SelectItem key={d.deviceId} value={d.deviceId}>
												{d.label}
											</SelectItem>
										))
									)}
								</SelectContent>
							</Select>
						</div>

						{/* Микрофон */}
						<div className="space-y-2">
							<label className="flex items-center gap-2 text-xs font-medium text-white/60 uppercase tracking-wider">
								<Mic className="h-3.5 w-3.5" />
								Микрофон
							</label>
							<Select
								value={selectedInput}
								onValueChange={(v) => void handleInputChange(v)}
							>
								<SelectTrigger className={selectTriggerClasses}>
									<SelectValue
										placeholder="Нет доступных устройств"
									/>
								</SelectTrigger>
								<SelectContent>
									{audioInputs.length === 0 ? (
										<SelectItem value="__none" disabled>
											Нет доступных устройств
										</SelectItem>
									) : (
										audioInputs.map((d) => (
											<SelectItem key={d.deviceId} value={d.deviceId}>
												{d.label}
											</SelectItem>
										))
									)}
								</SelectContent>
							</Select>
						</div>

						{/* Разделитель */}
						<div className="border-t border-white/5" />

						{/* Тоггл: подключаться с выключенным микрофоном */}
						<ToggleRow
							label="Подключаться с выключенным микрофоном"
							description="Микрофон будет выключен при входе в звонок"
							checked={joinMicOff}
							onChange={handleJoinMicOffToggle}
						/>

						{/* Тоггл: шумоподавление */}
						<ToggleRow
							label="Шумоподавление"
							description="Автоматическое подавление фонового шума"
							checked={noiseSuppression}
							onChange={handleNoiseSuppressionToggle}
						/>
					</TabsContent>

					{/* Вкладка "Видео" */}
					<TabsContent value="video" className="mt-4 space-y-5">
						{/* Камера */}
						<div className="space-y-2">
							<label className="flex items-center gap-2 text-xs font-medium text-white/60 uppercase tracking-wider">
								<Camera className="h-3.5 w-3.5" />
								Камера
							</label>
							<Select
								value={selectedCamera}
								onValueChange={(v) => void handleCameraChange(v)}
							>
								<SelectTrigger className={selectTriggerClasses}>
									<SelectValue
										placeholder="Нет доступных устройств"
									/>
								</SelectTrigger>
								<SelectContent>
									{videoInputs.length === 0 ? (
										<SelectItem value="__none" disabled>
											Нет доступных устройств
										</SelectItem>
									) : (
										videoInputs.map((d) => (
											<SelectItem key={d.deviceId} value={d.deviceId}>
												{d.label}
											</SelectItem>
										))
									)}
								</SelectContent>
							</Select>
						</div>

						{/* Разделитель */}
						<div className="border-t border-white/5" />

						{/* Тоггл: подключаться с выключенной камерой */}
						<ToggleRow
							label="Подключаться с выключенной камерой"
							description="Камера будет выключена при входе в звонок"
							checked={joinCamOff}
							onChange={handleJoinCamOffToggle}
						/>
					</TabsContent>
				</Tabs>
			</DialogContent>
		</Dialog>
	);
}

/** Пропсы строки с тогглом. */
interface ToggleRowProps {
	/** Основной текст. */
	label: string;
	/** Описание под основным текстом. */
	description: string;
	/** Текущее состояние тоггла. */
	checked: boolean;
	/** Коллбэк при переключении. */
	onChange: () => void;
}

/**
 * Строка с тогглом: текст слева, переключатель справа.
 * Кастомный switch без зависимости от @radix-ui/react-switch.
 */
function ToggleRow({ label, description, checked, onChange }: ToggleRowProps) {
	return (
		<button
			type="button"
			onClick={onChange}
			className="flex w-full items-center justify-between gap-4 rounded-lg p-2 -mx-2 hover:bg-white/5 transition-colors text-left"
		>
			<div className="min-w-0">
				<p className="text-sm font-medium text-white/90">{label}</p>
				<p className="text-xs text-white/40 mt-0.5">{description}</p>
			</div>
			{/* Кастомный switch */}
			<div
				className={cn(
					"relative inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors",
					checked ? "bg-blue-600" : "bg-white/20",
				)}
			>
				<span
					className={cn(
						"pointer-events-none block h-4 w-4 rounded-full bg-white shadow-lg transition-transform",
						checked ? "translate-x-5" : "translate-x-1",
					)}
				/>
			</div>
		</button>
	);
}
