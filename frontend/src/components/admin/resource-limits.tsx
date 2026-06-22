import { useEffect, useState } from "react";
import { getLimits, type ResourceLimit, updateLimit } from "@/api/admin";
import { toastApiError } from "@/lib/format-api-error";

/** Описания лимитов на русском. */
const LIMIT_LABELS: Record<string, string> = {
	"limit.file.max_size_mb": "Максимальный размер файла (МБ)",
	"limit.user.max_projects": "Максимум проектов на пользователя",
	"limit.project.max_storage_mb": "Максимум хранилища на проект (МБ)",
};

/**
 * ResourceLimits -- страница администратора для настройки лимитов ресурсов.
 * Отображает форму с числовыми полями и кнопкой сохранения.
 */
export function ResourceLimits() {
	const [limits, setLimits] = useState<ResourceLimit[]>([]);
	const [values, setValues] = useState<Record<string, string>>({});
	const [loading, setLoading] = useState(false);
	const [saving, setSaving] = useState(false);
	const [toast, setToast] = useState("");

	useEffect(() => {
		setLoading(true);
		getLimits()
			.then((data) => {
				setLimits(data);
				const vals: Record<string, string> = {};
				for (const l of data) {
					vals[l.key] = l.value;
				}
				setValues(vals);
			})
			.catch((err) => toastApiError(err, "Не удалось загрузить лимиты"))
			.finally(() => setLoading(false));
	}, []);

	/** Сохранить все изменённые лимиты. */
	const handleSave = async () => {
		setSaving(true);
		try {
			for (const limit of limits) {
				if (values[limit.key] !== limit.value) {
					await updateLimit(limit.key, values[limit.key] ?? limit.value);
				}
			}
			setToast("Лимиты сохранены");
			setTimeout(() => setToast(""), 3000);
			// Обновляем
			const data = await getLimits();
			setLimits(data);
		} catch (err) {
			toastApiError(err, "Ошибка сохранения лимитов");
			setToast("Ошибка сохранения");
			setTimeout(() => setToast(""), 3000);
		} finally {
			setSaving(false);
		}
	};

	if (loading) {
		return (
			<div className="py-8 text-center text-t3 text-[13px]">Загрузка...</div>
		);
	}

	return (
		<div>
			<h3 className="text-[16px] font-bold mb-[18px]">Лимиты ресурсов</h3>

			<div className="max-w-[480px]">
				{limits.map((limit) => {
					const inputId = `limit-${limit.key.replace(/\./g, "-")}`;
					return (
						<div key={limit.key} className="mb-4">
							<label
								htmlFor={inputId}
								className="block text-[13px] font-semibold text-t2 mb-[6px]"
							>
								{LIMIT_LABELS[limit.key] ?? limit.key}
							</label>
							<input
								id={inputId}
								type="number"
								min={0}
								value={values[limit.key] ?? ""}
								onChange={(e) =>
									setValues((v) => ({ ...v, [limit.key]: e.target.value }))
								}
								className="w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none"
							/>
						</div>
					);
				})}
			</div>

			<div className="flex items-center gap-3 mt-2">
				<button
					type="button"
					onClick={handleSave}
					disabled={saving}
					className="inline-flex items-center gap-[7px] px-5 py-[9px] rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_28px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96] disabled:opacity-50"
				>
					{saving ? "Сохранение..." : "Сохранить"}
				</button>
				{toast && <span className="text-[13px] text-green">{toast}</span>}
			</div>
		</div>
	);
}
