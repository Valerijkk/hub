import { useQuery } from "@tanstack/react-query";
import { Check, ChevronsUpDown, UserRound, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { listMembers, type MemberResponse } from "@/api/projects";
import { Button } from "@/components/ui/button";
import {
	Command,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "@/components/ui/command";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

interface UserPickerProps {
	projectId: string;
	value: string;
	onChange: (userId: string) => void;
	placeholder?: string;
	disabled?: boolean;
}

function memberLabel(m: MemberResponse): string {
	return m.name?.trim() || m.email || m.user_id.slice(0, 8);
}

/** Autocomplete выбора исполнителя из участников проекта — UX-P0-02. */
export function UserPicker({
	projectId,
	value,
	onChange,
	placeholder,
	disabled,
}: UserPickerProps) {
	const { t } = useTranslation("tracker");
	const [open, setOpen] = useState(false);

	const {
		data: members = [],
		isLoading,
		isError,
		refetch,
	} = useQuery({
		queryKey: ["project-members", projectId],
		queryFn: () => listMembers(projectId),
		staleTime: 60_000,
	});

	const selected = useMemo(
		() => members.find((m) => m.user_id === value),
		[members, value],
	);

	const label = placeholder ?? t("form.assigneePlaceholder");

	return (
		<div className="flex gap-2">
			<Popover open={open} onOpenChange={setOpen}>
				<PopoverTrigger asChild>
					<Button
						type="button"
						variant="outline"
						role="combobox"
						aria-expanded={open}
						disabled={disabled || isLoading}
						className="w-full justify-between font-normal"
					>
						<span className="flex items-center gap-2 truncate">
							<UserRound className="h-4 w-4 shrink-0 text-t3" aria-hidden />
							{selected ? memberLabel(selected) : label}
						</span>
						<ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" />
					</Button>
				</PopoverTrigger>
				<PopoverContent
					className="w-[var(--radix-popover-trigger-width)] p-0"
					align="start"
				>
					<Command>
						<CommandInput
							placeholder={t("userPicker.searchPlaceholder")}
							aria-label={t("userPicker.searchAria")}
						/>
						<CommandList>
							{isError ? (
								<div
									role="alert"
									className="px-3 py-6 text-center space-y-2"
								>
									<p className="text-sm text-[var(--red)]">
										{t("userPicker.loadError")}
									</p>
									<button
										type="button"
										className="text-sm font-semibold text-t2 hover:text-t1 hover:underline"
										onClick={() => void refetch()}
									>
										{t("userPicker.retry")}
									</button>
								</div>
							) : (
								<>
									<CommandEmpty>
										{isLoading
											? t("userPicker.loading")
											: t("userPicker.empty")}
									</CommandEmpty>
									<CommandGroup>
										{members.map((m) => (
											<CommandItem
												key={m.user_id}
												value={`${memberLabel(m)} ${m.email}`}
												onSelect={() => {
													onChange(m.user_id);
													setOpen(false);
												}}
											>
												<Check
													className={cn(
														"mr-2 h-4 w-4",
														value === m.user_id ? "opacity-100" : "opacity-0",
													)}
												/>
												<div className="flex min-w-0 flex-col">
													<span className="truncate">{memberLabel(m)}</span>
													<span className="truncate text-xs text-t2">
														{m.email}
													</span>
												</div>
											</CommandItem>
										))}
									</CommandGroup>
								</>
							)}
						</CommandList>
					</Command>
				</PopoverContent>
			</Popover>
			{value ? (
				<Button
					type="button"
					variant="ghost"
					size="icon"
					aria-label={t("userPicker.clearAssignee")}
					onClick={() => onChange("")}
				>
					<X className="h-4 w-4" />
				</Button>
			) : null}
		</div>
	);
}
