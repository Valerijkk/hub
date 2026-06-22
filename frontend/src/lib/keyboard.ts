import type { KeyboardEvent } from "react";

/** Enter/Space activation for non-button interactive elements. */
export function activateOnEnterOrSpace(onActivate: () => void) {
	return (event: KeyboardEvent<HTMLElement>) => {
		if (event.key === "Enter" || event.key === " ") {
			event.preventDefault();
			onActivate();
		}
	};
}
