import { useTranslation } from "react-i18next";

/** Skip link для NFR-16 / UX-P0-05. */
export function SkipLink() {
	const { t } = useTranslation("common");
	return (
		<a
			href="#main-content"
			className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[100] focus:rounded-lg focus:bg-amber-400 focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-black focus:shadow-lg focus:outline-none"
		>
			{t("a11y.skipToContent")}
		</a>
	);
}
