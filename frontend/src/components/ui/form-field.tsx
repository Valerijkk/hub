import { Children, cloneElement, isValidElement, useId } from "react";

/** Поле формы с лейблом — переиспользуемый компонент. */
export function FormField({
	label,
	htmlFor,
	children,
	required,
	hint,
}: {
	label: string;
	htmlFor: string;
	children: React.ReactNode;
	required?: boolean;
	hint?: string;
}) {
	const autoId = useId();
	const hintId = hint ? `${htmlFor || autoId}-hint` : undefined;

	const describedChildren =
		hintId != null
			? Children.map(children, (child) => {
					if (!isValidElement<{ "aria-describedby"?: string }>(child)) {
						return child;
					}
					const existing = child.props["aria-describedby"];
					return cloneElement(child, {
						"aria-describedby": existing ? `${existing} ${hintId}` : hintId,
						...(required ? { "aria-required": true } : {}),
					});
				})
			: required
				? Children.map(children, (child) => {
						if (!isValidElement(child)) return child;
						return cloneElement(
							child as React.ReactElement<{ "aria-required"?: boolean }>,
							{ "aria-required": true },
						);
					})
				: children;

	return (
		<div>
			<label
				htmlFor={htmlFor}
				className="block text-[12px] font-semibold text-t2 mb-[5px]"
			>
				{label}
				{required && (
					<span className="text-red ml-0.5">
						*<span className="sr-only"> (обязательно)</span>
					</span>
				)}
			</label>
			{describedChildren}
			{hint && (
				<p id={hintId} className="text-[11px] text-t3 mt-1">
					{hint}
				</p>
			)}
		</div>
	);
}
