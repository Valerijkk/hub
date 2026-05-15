import { createFileRoute } from "@tanstack/react-router";

/**
 * Стаб маршрута /call/$code. Реальный компонент с LiveKit-зависимостями
 * вынесен в `call.$code.lazy.tsx` (FE-H5): livekit-client (~478KB) грузится
 * только когда пользователь реально открывает гостевую страницу звонка.
 */
export const Route = createFileRoute("/call/$code")({});
