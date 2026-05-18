import { createContext, useContext } from "react";

/**
 * Контекст мобильного выдвижного списка каналов.
 *
 * Лейаут мессенджера (`/_authenticated/messenger`) на мобилке прячет sidebar и
 * выставляет наружу способ его открыть. Дочерний `ChatView` использует это
 * чтобы показать гамбургер в шапке. Контекст молча no-op'ит, если провайдер
 * отсутствует (например, при изолированном рендере для тестов).
 */
export interface ChannelDrawerContextValue {
	/** Открыт ли drawer сейчас. */
	open: boolean;
	/** Установить состояние drawer'а. */
	setOpen: (open: boolean) => void;
}

const noopContext: ChannelDrawerContextValue = {
	open: false,
	setOpen: () => {},
};

export const ChannelDrawerContext =
	createContext<ChannelDrawerContextValue>(noopContext);

/** Хук доступа к контексту drawer'а каналов. */
export function useChannelDrawer(): ChannelDrawerContextValue {
	return useContext(ChannelDrawerContext);
}
