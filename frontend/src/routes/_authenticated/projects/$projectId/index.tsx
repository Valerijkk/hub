import { createFileRoute, useNavigate, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  getProject,
  listMembers,
  archiveProject,
  deleteProject,
} from "@/api/projects";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { Dropzone } from "@/components/files/dropzone";

/**
 * Страница обзора проекта.
 * Отображает основную информацию, участников, файлы и настройки.
 */
function ProjectPage() {
  const { projectId } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: project, isLoading } = useQuery({
    queryKey: ["project", projectId],
    queryFn: () => getProject(projectId),
  });

  const { data: members } = useQuery({
    queryKey: ["project", projectId, "members"],
    queryFn: () => listMembers(projectId),
  });

  const archiveMutation = useMutation({
    mutationFn: () => archiveProject(projectId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Проект архивирован");
      void navigate({ to: "/projects" });
    },
    onError: () => toast.error("Не удалось архивировать проект"),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(projectId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Проект удален");
      void navigate({ to: "/projects" });
    },
    onError: () => toast.error("Не удалось удалить проект"),
  });

  if (isLoading) {
    return (
      <div className="animate-fade-in space-y-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-xl animate-shimmer" />
            <div className="space-y-2">
              <div className="h-6 w-48 rounded animate-shimmer" />
              <div className="h-4 w-24 rounded animate-shimmer" />
            </div>
          </div>
          <div className="h-9 w-32 rounded-lg animate-shimmer" />
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="rounded-xl border bg-card p-5 space-y-2">
              <div className="h-4 w-20 rounded animate-shimmer" />
              <div className="h-8 w-16 rounded animate-shimmer" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Проект не найден</p>
      </div>
    );
  }

  const isOwnerOrAdmin =
    project.my_role === "owner" || project.my_role === "admin";

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div>
            <h1 className="text-2xl font-bold">{project.name}</h1>
            <p className="text-sm text-muted-foreground">{project.key}</p>
          </div>
          <Badge variant={project.status === "active" ? "default" : "secondary"}>
            {project.status === "active" ? "Активный" : project.status}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" asChild>
            <Link to="/projects/$projectId/board" params={{ projectId }}>
              Доска
            </Link>
          </Button>
          <Button variant="outline" asChild>
            <Link to="/projects/$projectId/issues" params={{ projectId }}>
              Задачи
            </Link>
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Обзор</TabsTrigger>
          <TabsTrigger value="members">Участники</TabsTrigger>
          <TabsTrigger value="files">Файлы</TabsTrigger>
          {isOwnerOrAdmin && (
            <TabsTrigger value="settings">Настройки</TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="overview" className="mt-4 space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>О проекте</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                {project.description || "Описание не указано"}
              </p>
              <Separator className="my-4" />
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Участников:</span>{" "}
                  {project.member_count}
                </div>
                <div>
                  <span className="text-muted-foreground">Роль:</span>{" "}
                  {project.my_role ?? "Не указана"}
                </div>
                <div>
                  <span className="text-muted-foreground">Создан:</span>{" "}
                  {new Date(project.created_at).toLocaleDateString("ru-RU")}
                </div>
                <div>
                  <span className="text-muted-foreground">Обновлен:</span>{" "}
                  {new Date(project.updated_at).toLocaleDateString("ru-RU")}
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="members" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Участники</CardTitle>
              <CardDescription>
                {members?.length ?? 0} участников в проекте
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {members?.map((member) => (
                  <div
                    key={member.user_id}
                    className="flex items-center justify-between rounded-md border p-3"
                  >
                    <div>
                      <p className="font-medium">{member.name || member.email}</p>
                      <p className="text-sm text-muted-foreground">{member.email}</p>
                    </div>
                    <Badge variant="outline">{member.role}</Badge>
                  </div>
                ))}
                {(!members || members.length === 0) && (
                  <p className="text-sm text-muted-foreground">Нет участников</p>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="files" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Файлы</CardTitle>
              <CardDescription>
                Загрузите файлы через drag and drop или Ctrl+V
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Dropzone projectId={projectId} entityType="project" entityId={projectId} />
            </CardContent>
          </Card>
        </TabsContent>

        {isOwnerOrAdmin && (
          <TabsContent value="settings" className="mt-4 space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Настройки проекта</CardTitle>
                <CardDescription>Управление проектом</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <h3 className="mb-2 font-medium">Архивация</h3>
                  <p className="mb-2 text-sm text-muted-foreground">
                    Архивированный проект доступен только для чтения.
                  </p>
                  <Button
                    variant="outline"
                    onClick={() => archiveMutation.mutate()}
                    disabled={archiveMutation.isPending || project.status !== "active"}
                  >
                    Архивировать
                  </Button>
                </div>
                <Separator />
                <div>
                  <h3 className="mb-2 font-medium text-destructive">
                    Удаление
                  </h3>
                  <p className="mb-2 text-sm text-muted-foreground">
                    Проект будет перемещен в корзину на 30 дней.
                  </p>
                  <Button
                    variant="destructive"
                    onClick={() => {
                      if (window.confirm("Удалить проект?")) {
                        deleteMutation.mutate();
                      }
                    }}
                    disabled={deleteMutation.isPending}
                  >
                    Удалить
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/",
)({
  component: ProjectPage,
});
