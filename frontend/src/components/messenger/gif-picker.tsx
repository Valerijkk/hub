import { useCallback, useEffect, useRef, useState } from "react";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Типы Giphy API
// ---------------------------------------------------------------------------

/** Объект GIF из ответа Giphy API. */
interface GiphyGif {
	id: string;
	title: string;
	images: {
		/** Фиксированная ширина для превью в сетке. */
		fixed_width: {
			url: string;
			width: string;
			height: string;
		};
		/** Оригинальный GIF для отправки. */
		original: {
			url: string;
			width: string;
			height: string;
		};
		/** Маленькое неподвижное превью для lazy-загрузки. */
		fixed_width_still: {
			url: string;
		};
	};
}

/** Ответ Giphy API (trending / search). */
interface GiphyResponse {
	data: GiphyGif[];
}

// ---------------------------------------------------------------------------
// Константы
// ---------------------------------------------------------------------------

/** Публичный бета-ключ Giphy для разработки. */
const GIPHY_API_KEY = "GlVGYHkr3WSBnllca54iNt0yFbjz7L65";

/** Базовый URL Giphy API. */
const GIPHY_BASE = "https://api.giphy.com/v1/gifs";

/** Количество GIF на один запрос. */
const LIMIT = 20;

/** Задержка дебаунса поиска (мс). */
const SEARCH_DEBOUNCE_MS = 400;

// ---------------------------------------------------------------------------
// Хук для загрузки GIF с Giphy API
// ---------------------------------------------------------------------------

/**
 * useGiphySearch -- загружает trending или результаты поиска GIF.
 * Включает дебаунс поиска и обработку ошибок.
 */
function useGiphySearch() {
	const [gifs, setGifs] = useState<GiphyGif[]>([]);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const abortRef = useRef<AbortController | null>(null);

	/** Загрузить trending GIFs. */
	const fetchTrending = useCallback(async () => {
		abortRef.current?.abort();
		const controller = new AbortController();
		abortRef.current = controller;

		setLoading(true);
		setError(null);

		try {
			const url = `${GIPHY_BASE}/trending?api_key=${GIPHY_API_KEY}&limit=${LIMIT}&rating=g`;
			const res = await fetch(url, { signal: controller.signal });
			if (!res.ok) throw new Error(`Giphy API: ${res.status}`);
			const json = (await res.json()) as GiphyResponse;
			setGifs(json.data);
		} catch (err) {
			if (err instanceof DOMException && err.name === "AbortError") return;
			setError("Не удалось загрузить GIF");
		} finally {
			setLoading(false);
		}
	}, []);

	/** Поиск GIF по запросу. */
	const searchGifs = useCallback(async (query: string) => {
		abortRef.current?.abort();
		const controller = new AbortController();
		abortRef.current = controller;

		setLoading(true);
		setError(null);

		try {
			const encoded = encodeURIComponent(query);
			const url = `${GIPHY_BASE}/search?api_key=${GIPHY_API_KEY}&q=${encoded}&limit=${LIMIT}&rating=g`;
			const res = await fetch(url, { signal: controller.signal });
			if (!res.ok) throw new Error(`Giphy API: ${res.status}`);
			const json = (await res.json()) as GiphyResponse;
			setGifs(json.data);
		} catch (err) {
			if (err instanceof DOMException && err.name === "AbortError") return;
			setError("Не удалось выполнить поиск");
		} finally {
			setLoading(false);
		}
	}, []);

	/** Отменить текущий запрос при размонтировании. */
	useEffect(() => {
		return () => abortRef.current?.abort();
	}, []);

	return { gifs, loading, error, fetchTrending, searchGifs };
}

// ---------------------------------------------------------------------------
// Компонент GifPicker
// ---------------------------------------------------------------------------

/** Пропсы GifPicker. */
interface GifPickerProps {
	/** Колбэк при выборе GIF -- передаёт URL оригинального GIF. */
	onSelect: (gifUrl: string) => void;
}

/**
 * GifPicker -- поповер для поиска и выбора GIF через Giphy API.
 * Отображает сетку из 2 колонок с ленивой загрузкой изображений.
 * Внизу -- обязательная атрибуция "Powered by GIPHY".
 */
export function GifPicker({ onSelect }: GifPickerProps) {
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const { gifs, loading, error, fetchTrending, searchGifs } = useGiphySearch();
	const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);

	/** При открытии поповера -- загружаем trending. */
	const handleOpenChange = useCallback(
		(isOpen: boolean) => {
			setOpen(isOpen);
			if (isOpen) {
				setQuery("");
				void fetchTrending();
				// Фокус на поле поиска после открытия.
				requestAnimationFrame(() => inputRef.current?.focus());
			}
		},
		[fetchTrending],
	);

	/** Обработка ввода в поле поиска с дебаунсом. */
	const handleSearchInput = useCallback(
		(value: string) => {
			setQuery(value);

			// Сбрасываем предыдущий таймер.
			if (debounceRef.current) {
				clearTimeout(debounceRef.current);
			}

			const trimmed = value.trim();
			if (!trimmed) {
				// Пустой запрос -- показываем trending.
				void fetchTrending();
				return;
			}

			// Дебаунс поиска.
			debounceRef.current = setTimeout(() => {
				void searchGifs(trimmed);
			}, SEARCH_DEBOUNCE_MS);
		},
		[fetchTrending, searchGifs],
	);

	/** Обработка выбора GIF. */
	const handleGifClick = useCallback(
		(gif: GiphyGif) => {
			// Отправляем URL оригинального GIF.
			onSelect(gif.images.original.url);
			setOpen(false);
			setQuery("");
		},
		[onSelect],
	);

	return (
		<Popover open={open} onOpenChange={handleOpenChange}>
			<PopoverTrigger asChild>
				<button
					type="button"
					className="flex h-11 w-9 shrink-0 items-center justify-center text-t2/60 hover:text-t2 transition-colors"
					title="GIF"
					aria-label="Выбрать GIF"
				>
					{/* Иконка плёнки (Film) */}
					<svg
						aria-hidden="true"
						className="h-5 w-5"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
						/>
					</svg>
				</button>
			</PopoverTrigger>

			<PopoverContent
				className="w-80 p-0 overflow-hidden"
				align="end"
				side="top"
				sideOffset={8}
			>
				<div className="flex flex-col max-h-[420px]">
					{/* Поле поиска */}
					<div className="p-2 border-b border-border/50">
						<input
							ref={inputRef}
							type="text"
							value={query}
							onChange={(e) => handleSearchInput(e.target.value)}
							placeholder="Поиск GIF..."
							className={cn(
								"w-full rounded-lg border border-input bg-base px-3 py-2 text-sm",
								"placeholder:text-t2/50",
								"focus:outline-none focus:ring-2 focus:ring-ring/50",
							)}
						/>
					</div>

					{/* Контент: сетка GIF / загрузка / ошибка */}
					<div className="flex-1 overflow-y-auto p-2 min-h-[200px]">
						{loading && gifs.length === 0 && (
							<div className="flex items-center justify-center h-[200px]">
								<div className="flex flex-col items-center gap-2 text-t2">
									<svg
										aria-hidden="true"
										className="h-6 w-6 animate-spin"
										fill="none"
										viewBox="0 0 24 24"
									>
										<circle
											className="opacity-25"
											cx="12"
											cy="12"
											r="10"
											stroke="currentColor"
											strokeWidth="4"
										/>
										<path
											className="opacity-75"
											fill="currentColor"
											d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
										/>
									</svg>
									<span className="text-xs">Загрузка...</span>
								</div>
							</div>
						)}

						{error && (
							<div className="flex items-center justify-center h-[200px]">
								<p className="text-xs text-destructive">{error}</p>
							</div>
						)}

						{!loading && !error && gifs.length === 0 && (
							<div className="flex items-center justify-center h-[200px]">
								<p className="text-xs text-t2">Ничего не найдено</p>
							</div>
						)}

						{gifs.length > 0 && (
							<div className="columns-2 gap-1.5 space-y-1.5">
								{gifs.map((gif) => (
									<GifThumbnail
										key={gif.id}
										gif={gif}
										onClick={handleGifClick}
									/>
								))}
							</div>
						)}
					</div>

					{/* Атрибуция Giphy (обязательная по ToS) */}
					<div className="flex items-center justify-center py-1.5 border-t border-border/50 bg-raised/30">
						<span className="text-[10px] text-t2/70 font-medium tracking-wide">
							Powered by GIPHY
						</span>
					</div>
				</div>
			</PopoverContent>
		</Popover>
	);
}

// ---------------------------------------------------------------------------
// Компонент миниатюры GIF
// ---------------------------------------------------------------------------

/** Пропсы GifThumbnail. */
interface GifThumbnailProps {
	gif: GiphyGif;
	onClick: (gif: GiphyGif) => void;
}

/**
 * GifThumbnail -- одна миниатюра GIF в сетке.
 * Использует IntersectionObserver для ленивой загрузки.
 */
function GifThumbnail({ gif, onClick }: GifThumbnailProps) {
	const imgRef = useRef<HTMLImageElement>(null);
	const [loaded, setLoaded] = useState(false);
	const [inView, setInView] = useState(false);

	// Ленивая загрузка через IntersectionObserver.
	useEffect(() => {
		const el = imgRef.current;
		if (!el) return;

		const observer = new IntersectionObserver(
			([entry]) => {
				if (entry?.isIntersecting) {
					setInView(true);
					observer.disconnect();
				}
			},
			{ rootMargin: "100px" },
		);

		observer.observe(el);
		return () => observer.disconnect();
	}, []);

	// Вычисляем aspect ratio для плейсхолдера.
	const w = Number.parseInt(gif.images.fixed_width.width, 10) || 200;
	const h = Number.parseInt(gif.images.fixed_width.height, 10) || 200;
	const aspectRatio = `${w}/${h}`;

	return (
		<button
			type="button"
			className={cn(
				"block w-full rounded-md overflow-hidden cursor-pointer",
				"transition-[transform,filter,opacity] duration-150",
				"hover:ring-2 hover:ring-primary/50 hover:scale-[1.02]",
				"break-inside-avoid",
			)}
			onClick={() => onClick(gif)}
			title={gif.title || "GIF"}
		>
			<img
				ref={imgRef}
				src={inView ? gif.images.fixed_width.url : undefined}
				alt={gif.title || "GIF"}
				className={cn(
					"w-full object-cover transition-opacity duration-200",
					loaded ? "opacity-100" : "opacity-0",
				)}
				style={{ aspectRatio }}
				loading="lazy"
				onLoad={() => setLoaded(true)}
			/>
			{/* Плейсхолдер пока изображение не загружено */}
			{!loaded && (
				<div
					className="w-full bg-raised/50 animate-pulse rounded-md"
					style={{ aspectRatio }}
				/>
			)}
		</button>
	);
}
