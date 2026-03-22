import { createRootRoute, Outlet } from "@tanstack/react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { ThemeProvider } from "@/components/theme-provider";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
});

function RootLayout() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <Outlet />
        <Toaster position="top-right" richColors closeButton />
      </ThemeProvider>
    </QueryClientProvider>
  );
}

function RootErrorBoundary() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="rounded-lg border bg-card p-8 text-center shadow-sm">
        <h1 className="mb-2 text-2xl font-semibold text-foreground">
          Произошла ошибка
        </h1>
        <p className="mb-4 text-muted-foreground">
          Что-то пошло не так. Попробуйте обновить страницу.
        </p>
        <button
          type="button"
          onClick={() => window.location.reload()}
          className="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Обновить
        </button>
      </div>
    </div>
  );
}

function NotFoundComponent() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="text-center animate-fade-in max-w-md px-6">
        {/* Fun illustration */}
        <div className="relative mx-auto mb-8 flex h-32 w-32 items-center justify-center">
          <div className="absolute inset-0 rounded-full bg-gradient-to-br from-blue-500/10 to-purple-500/10 animate-pulse" />
          <div className="relative">
            <svg className="h-20 w-20 text-muted-foreground/30" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <path d="M8 15s1.5-2 4-2 4 2 4 2" />
              <line x1="9" y1="9" x2="9.01" y2="9" strokeWidth="2" />
              <line x1="15" y1="9" x2="15.01" y2="9" strokeWidth="2" />
            </svg>
          </div>
          <div className="absolute -top-2 -right-2 h-4 w-4 rounded-full bg-blue-500/20" />
          <div className="absolute -bottom-1 -left-3 h-3 w-3 rounded-full bg-purple-500/20" />
        </div>
        <h1 className="mb-2 text-7xl font-extrabold tracking-tighter">
          <span className="gradient-text">404</span>
        </h1>
        <p className="mb-2 text-xl font-semibold text-foreground">
          Страница не найдена
        </p>
        <p className="mb-8 text-sm text-muted-foreground leading-relaxed">
          Похоже, эта страница решила взять выходной. Попробуйте вернуться на главную.
        </p>
        <a
          href="/"
          className="inline-flex h-10 items-center justify-center rounded-lg bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 px-6 py-2 text-sm font-medium text-white shadow-md shadow-blue-500/25 hover-pulse transition-all"
        >
          <svg className="mr-2 h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M8 2L2 8l6 6M2 8h12" /></svg>
          На главную
        </a>
      </div>
    </div>
  );
}

export const Route = createRootRoute({
  component: RootLayout,
  errorComponent: RootErrorBoundary,
  notFoundComponent: NotFoundComponent,
});
