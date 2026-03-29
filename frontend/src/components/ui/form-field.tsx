/** Поле формы с лейблом -- переиспользуемый компонент. */
export function FormField({
	label,
	children,
	required,
	hint,
}: {
	label: string;
	children: React.ReactNode;
	required?: boolean;
	hint?: string;
}) {
	return (
		<div>
			<label className="block text-[12px] font-semibold text-t2 mb-[5px]">
				{label}
				{required && <span className="text-red ml-0.5">*</span>}
			</label>
			{children}
			{hint && <p className="text-[11px] text-t3 mt-1">{hint}</p>}
		</div>
	);
}
