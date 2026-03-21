import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** Объединяет CSS-классы с помощью clsx и tailwind-merge для корректного слияния Tailwind-утилит. */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
