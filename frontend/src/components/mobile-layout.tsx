/**
 * Утилитарные компоненты для адаптивной мобильной раскладки.
 *
 * Предоставляет хук useIsMobile и обёртки для контента,
 * которые переключают вид между desktop и mobile.
 */
import { useEffect, useState } from "react";

/** Точка перелома для мобильных устройств (Tailwind md) */
const MOBILE_BREAKPOINT = 768;

/**
 * Хук определения мобильного viewport (ширина < 768px).
 * Использует matchMedia для реактивного обновления.
 */
export function useIsMobile(): boolean {
  const [isMobile, setIsMobile] = useState(() => {
    if (typeof window === "undefined") return false;
    return window.innerWidth < MOBILE_BREAKPOINT;
  });

  useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);

    function handleChange(e: MediaQueryListEvent) {
      setIsMobile(e.matches);
    }

    mql.addEventListener("change", handleChange);
    setIsMobile(mql.matches);

    return () => mql.removeEventListener("change", handleChange);
  }, []);

  return isMobile;
}

/**
 * Обёртка, скрывающая контент на мобильных устройствах
 */
export function DesktopOnly({ children }: { children: React.ReactNode }) {
  return <div className="hidden md:block">{children}</div>;
}

/**
 * Обёртка, скрывающая контент на десктопе
 */
export function MobileOnly({ children }: { children: React.ReactNode }) {
  return <div className="block md:hidden">{children}</div>;
}

/**
 * Адаптивный контейнер: на мобильных -- 1 колонка, на десктопе -- grid
 */
export function ResponsiveGrid({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={`grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 ${className}`}>
      {children}
    </div>
  );
}
