import { useTranslation } from "react-i18next";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";

interface ShortcutsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

/** FB-01: модалка горячих клавиш по `?`. */
export function ShortcutsDialog({ open, onOpenChange }: ShortcutsDialogProps) {
	const { t } = useTranslation("help");

	const groups = [
		{
			key: "navigation",
			items: t("shortcuts.navigation", { returnObjects: true }) as string[],
		},
		{
			key: "tracker",
			items: t("shortcuts.tracker", { returnObjects: true }) as string[],
		},
		{
			key: "messenger",
			items: t("shortcuts.messenger", { returnObjects: true }) as string[],
		},
	] as const;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle>{t("shortcuts.title")}</DialogTitle>
				</DialogHeader>
				<div className="space-y-4 max-h-[60vh] overflow-y-auto">
					{groups.map(({ key, items }) => (
						<div key={key}>
							<h3 className="text-xs font-bold uppercase tracking-wider text-t3 mb-2">
								{t(`shortcuts.groups.${key}`)}
							</h3>
							<ul className="space-y-1.5 text-sm text-t2">
								{items.map((line) => (
									<li key={line} className="font-mono text-[13px]">
										{line}
									</li>
								))}
							</ul>
						</div>
					))}
				</div>
				<p className="text-xs text-t3">{t("shortcuts.hint")}</p>
			</DialogContent>
		</Dialog>
	);
}
