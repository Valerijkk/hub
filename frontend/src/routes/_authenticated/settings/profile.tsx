import { createFileRoute } from "@tanstack/react-router";

/**
 * Стаб маршрута /settings/profile. Реальный компонент (~438 LOC) — в
 * `profile.lazy.tsx`. Code-split TanStack Router исключает страницу профиля
 * из main-бандла.
 */
export const Route = createFileRoute("/_authenticated/settings/profile")({});
