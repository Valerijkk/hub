import { createFileRoute } from "@tanstack/react-router";

/**
 * Стаб маршрута /onboarding. Реальный компонент (~543 LOC) вынесен в
 * `onboarding.lazy.tsx` — TanStack Router code-split: страница онбординга
 * грузится только при первом заходе нового пользователя.
 */
export const Route = createFileRoute("/onboarding")({});
