import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { listProjects } from "@/api/projects";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

/**
 * Страница списка проектов — сетка карточек, пустое состояние,
 * скелетон-загрузка, CTA для создания нового проекта.
 */
function ProjectsPage() {
  const { data: projects, isLoading } = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
  });

  return (
    <div className="mx-auto max-w-5xl animate-fade-in">
      {/* Заголовок страницы */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Проекты</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Управляйте проектами вашей команды
          </p>
        </div>
        <Button
          asChild
          className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
        >
          <Link to="/projects/new">
            <svg className="mr-2 h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M8 2v12M2 8h12" /></svg>
            Создать проект
          </Link>
        </Button>
      </div>

      {/* Скелетон-загрузка */}
      {isLoading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border bg-card p-5">
              <div className="space-y-3">
                <div className="h-5 w-2/3 rounded-md animate-shimmer" />
                <div className="h-4 w-1/4 rounded-md animate-shimmer" />
                <div className="h-4 w-full rounded-md animate-shimmer" />
                <div className="h-3 w-1/3 rounded-md animate-shimmer" />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Пустое состояние */}
      {!isLoading && projects?.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed bg-card/50 py-16 px-6">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 mb-4">
            <svg className="h-6 w-6 text-primary" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><rect x="2" y="2" width="5" height="5" rx="1" /><rect x="9" y="2" width="5" height="5" rx="1" /><rect x="2" y="9" width="5" height="5" rx="1" /><rect x="9" y="9" width="5" height="5" rx="1" /></svg>
          </div>
          <h3 className="text-lg font-medium mb-1">Нет проектов</h3>
          <p className="mb-6 text-sm text-muted-foreground text-center max-w-sm">
            Создайте первый проект для вашей команды, чтобы начать работу с задачами, тестами и общением.
          </p>
          <Button
            asChild
            className="bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
          >
            <Link to="/projects/new">Создать первый проект</Link>
          </Button>
        </div>
      )}

      {/* Сетка проектов */}
      {projects && projects.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <Link
              key={project.id}
              to="/projects/$projectId"
              params={{ projectId: project.id }}
              className="group block"
            >
              <Card className="h-full rounded-xl border bg-card transition-all duration-150 hover:shadow-md hover:border-primary/20 hover:-translate-y-px">
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <CardTitle className="text-base truncate">{project.name}</CardTitle>
                      <CardDescription className="mt-0.5">
                        <span className="inline-flex items-center rounded-md bg-muted px-1.5 py-0.5 text-xs font-mono font-medium text-muted-foreground">
                          {project.key}
                        </span>
                      </CardDescription>
                    </div>
                    <Badge
                      variant={project.status === "active" ? "default" : "secondary"}
                      className={project.status === "active" ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-0" : ""}
                    >
                      {project.status === "active" ? "Активный" : project.status}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  {project.description && (
                    <p className="mb-3 line-clamp-2 text-sm text-muted-foreground leading-relaxed">
                      {project.description}
                    </p>
                  )}
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <svg className="h-3.5 w-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"><circle cx="8" cy="5" r="3" /><path d="M2 14c0-3 2.7-5 6-5s6 2 6 5" /></svg>
                    <span>
                      {project.member_count}{" "}
                      {project.member_count === 1 ? "участник" : "участников"}
                    </span>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/projects/")({
  component: ProjectsPage,
});
