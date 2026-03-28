import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import { type ImportResult, importTestCases } from "@/api/tms";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";

/** Допустимые расширения файлов для импорта. */
const ACCEPTED_EXTENSIONS = ".csv,.json,.xml,.xlsx";

/** Определение формата по расширению файла. */
function detectFormat(filename: string): string {
	if (filename.endsWith(".csv")) return "csv";
	if (filename.endsWith(".json")) return "json";
	if (filename.endsWith(".xml")) return "xml";
	if (filename.endsWith(".xlsx")) return "excel";
	return "csv";
}

/**
 * Диалог импорта тест-кейсов.
 * Поддерживает загрузку файлов CSV, JSON, XML, XLSX.
 * Автоматически определяет формат по расширению.
 * Показывает результат импорта (создано/пропущено/ошибки).
 */
export function ImportDialog({
	projectId,
	open,
	onOpenChange,
}: {
	projectId: string;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}) {
	const queryClient = useQueryClient();
	const [file, setFile] = useState<File | null>(null);
	const [result, setResult] = useState<ImportResult | null>(null);

	const importMutation = useMutation({
		mutationFn: () => {
			if (!file) throw new Error("Файл не выбран");
			const format = detectFormat(file.name);
			return importTestCases(projectId, file, format);
		},
		onSuccess: (data) => {
			setResult(data);
			void queryClient.invalidateQueries({
				queryKey: ["tms-cases"],
			});
			toast.success(`Импортировано ${data.created} тест-кейсов`);
		},
		onError: () => toast.error("Ошибка импорта"),
	});

	const handleFileChange = useCallback(
		(e: React.ChangeEvent<HTMLInputElement>) => {
			const selected = e.target.files?.[0];
			if (selected) {
				setFile(selected);
				setResult(null);
			}
		},
		[],
	);

	const handleDrop = useCallback((e: React.DragEvent) => {
		e.preventDefault();
		const dropped = e.dataTransfer.files[0];
		if (dropped) {
			setFile(dropped);
			setResult(null);
		}
	}, []);

	const handleClose = useCallback(() => {
		setFile(null);
		setResult(null);
		onOpenChange(false);
	}, [onOpenChange]);

	return (
		<Dialog open={open} onOpenChange={handleClose}>
			<DialogContent className="max-w-lg">
				<DialogHeader>
					<DialogTitle>Импорт тест-кейсов</DialogTitle>
				</DialogHeader>

				{!result ? (
					<div className="space-y-4">
						{/* Зона перетаскивания */}
						<div
							className="flex min-h-[120px] cursor-pointer flex-col items-center justify-center rounded-md border-2 border-dashed p-6 transition-colors hover:border-primary"
							onDrop={handleDrop}
							onDragOver={(e) => e.preventDefault()}
							onClick={() =>
								document.getElementById("import-file-input")?.click()
							}
						>
							<input
								id="import-file-input"
								type="file"
								accept={ACCEPTED_EXTENSIONS}
								onChange={handleFileChange}
								className="hidden"
							/>
							{file ? (
								<div className="text-center">
									<div className="text-sm font-medium">{file.name}</div>
									<div className="text-xs text-t2">
										{(file.size / 1024).toFixed(1)} KB -{" "}
										{detectFormat(file.name).toUpperCase()}
									</div>
								</div>
							) : (
								<div className="text-center">
									<div className="text-sm text-t2">
										Перетащите файл или нажмите для выбора
									</div>
									<div className="mt-1 text-xs text-t2">
										Форматы: CSV, JSON, XML, XLSX
									</div>
								</div>
							)}
						</div>

						{/* Кнопка импорта */}
						<Button
							className="w-full"
							disabled={!file || importMutation.isPending}
							onClick={() => importMutation.mutate()}
						>
							{importMutation.isPending ? "Импорт..." : "Импортировать"}
						</Button>
					</div>
				) : (
					<div className="space-y-4">
						{/* Результат */}
						<div className="grid grid-cols-3 gap-3">
							<div className="rounded border p-3 text-center">
								<div className="text-lg font-bold text-green-600">
									{result.created}
								</div>
								<div className="text-xs text-t2">Создано</div>
							</div>
							<div className="rounded border p-3 text-center">
								<div className="text-lg font-bold text-yellow-600">
									{result.skipped}
								</div>
								<div className="text-xs text-t2">Пропущено</div>
							</div>
							<div className="rounded border p-3 text-center">
								<div className="text-lg font-bold text-red-600">
									{result.errors}
								</div>
								<div className="text-xs text-t2">Ошибок</div>
							</div>
						</div>

						{/* Ошибки валидации */}
						{result.validation_errors &&
							result.validation_errors.length > 0 && (
								<div className="max-h-40 overflow-y-auto rounded border p-2">
									<div className="mb-1 text-xs font-medium">
										Ошибки валидации:
									</div>
									{result.validation_errors.map((ve, i) => (
										<div key={i} className="text-xs text-red-600">
											Строка {ve.row}: {ve.field} - {ve.message}
										</div>
									))}
								</div>
							)}

						<Button variant="outline" className="w-full" onClick={handleClose}>
							Закрыть
						</Button>
					</div>
				)}
			</DialogContent>
		</Dialog>
	);
}
