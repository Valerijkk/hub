import { useCallback, useEffect, useRef } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Command } from "cmdk";
import { Search, Clock, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { useSearchStore } from "@/stores/search-store";
import { SearchResultItem, groupLabels } from "./search-results";
import type { SearchResult } from "@/api/search";

/** Группировка результатов по типу сущности. */
function groupResults(
  results: SearchResult[]
): Record<string, SearchResult[]> {
  const groups: Record<string, SearchResult[]> = {};
  for (const r of results) {
    const arr = groups[r.type] ?? [];
    arr.push(r);
    groups[r.type] = arr;
  }
  return groups;
}

/** Палитра команд (Ctrl+K) для глобального поиска по всем сущностям. */
export function CommandPalette() {
  const navigate = useNavigate();
  const {
    isOpen,
    setOpen,
    query,
    results,
    loading,
    recentSearches,
    search,
    addRecentSearch,
  } = useSearchStore();

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  /** Обработка ввода с debounce 300мс. */
  const handleInputChange = useCallback(
    (value: string) => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
      debounceRef.current = setTimeout(() => {
        void search(value);
      }, 300);
    },
    [search]
  );

  /** Очистка таймера при размонтировании. */
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  /** Обработка выбора результата. */
  const handleSelect = useCallback(
    (result: SearchResult) => {
      addRecentSearch(query);
      setOpen(false);
      void navigate({ to: result.url });
    },
    [navigate, setOpen, addRecentSearch, query]
  );

  /** Обработка выбора из недавних. */
  const handleRecentSelect = useCallback(
    (recent: string) => {
      void search(recent);
    },
    [search]
  );

  const grouped = groupResults(results);
  const hasResults = results.length > 0;
  const showRecent = !query.trim() && recentSearches.length > 0;

  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      <DialogContent className="overflow-hidden p-0 sm:max-w-[550px]">
        <DialogTitle className="sr-only">
          Глобальный поиск
        </DialogTitle>
        <Command shouldFilter={false} className="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground">
          <div className="flex items-center border-b px-3">
            {loading ? (
              <Loader2 className="mr-2 h-4 w-4 shrink-0 animate-spin text-muted-foreground" />
            ) : (
              <Search className="mr-2 h-4 w-4 shrink-0 text-muted-foreground" />
            )}
            <Command.Input
              placeholder="Поиск задач, кейсов, пользователей..."
              className="flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
              onValueChange={handleInputChange}
            />
          </div>

          <Command.List className="max-h-[400px] overflow-y-auto p-2">
            {/* Пустое состояние при наличии запроса */}
            {query.trim() && !loading && !hasResults && (
              <Command.Empty className="py-6 text-center text-sm text-muted-foreground">
                Ничего не найдено
              </Command.Empty>
            )}

            {/* Недавние поиски (при пустом запросе) */}
            {showRecent && (
              <Command.Group heading="Недавние поиски">
                {recentSearches.map((recent) => (
                  <Command.Item
                    key={recent}
                    value={`recent-${recent}`}
                    onSelect={() => handleRecentSelect(recent)}
                    className="flex cursor-pointer items-center gap-2 rounded-md px-3 py-2 text-sm hover:bg-accent"
                  >
                    <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                    <span>{recent}</span>
                  </Command.Item>
                ))}
              </Command.Group>
            )}

            {/* Сгруппированные результаты */}
            {Object.entries(grouped).map(([type, items]) => (
              <Command.Group
                key={type}
                heading={groupLabels[type] ?? type}
              >
                {items.map((result) => (
                  <Command.Item
                    key={`${result.type}-${result.id}`}
                    value={`${result.type}-${result.id}`}
                    onSelect={() => handleSelect(result)}
                    className="cursor-pointer"
                  >
                    <SearchResultItem result={result} />
                  </Command.Item>
                ))}
              </Command.Group>
            ))}
          </Command.List>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
