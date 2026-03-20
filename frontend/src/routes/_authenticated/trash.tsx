import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { listProjects, restoreProject } from "@/api/projects";
import type { ProjectResponse } from "@/api/projects";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function TrashPage() {
  const queryClient = useQueryClient();

  // Fetch all projects and filter deleted ones (status === "deleted")
  const { data: allProjects, isLoading } = useQuery({
    queryKey: ["projects", "all"],
    queryFn: listProjects,
  });

  const deletedProjects = allProjects?.filter(
    (p: ProjectResponse) => p.status === "deleted" || p.status === "archived"
  );

  const restoreMutation = useMutation({
    mutationFn: restoreProject,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Проект восстановлен");
    },
    onError: () => toast.error("Не удалось восстановить проект"),
  });

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Корзина</h1>
        <p className="text-sm text-muted-foreground">
          Удаленные проекты хранятся 30 дней. После этого удаляются безвозвратно.
        </p>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      )}

      {!isLoading && (!deletedProjects || deletedProjects.length === 0) && (
        <Card className="py-12 text-center">
          <CardContent>
            <p className="text-muted-foreground">Корзина пуста</p>
          </CardContent>
        </Card>
      )}

      {deletedProjects && deletedProjects.length > 0 && (
        <div className="space-y-3">
          {deletedProjects.map((project: ProjectResponse) => (
            <Card key={project.id}>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle className="text-lg">{project.name}</CardTitle>
                    <CardDescription>{project.key}</CardDescription>
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => restoreMutation.mutate(project.id)}
                    disabled={restoreMutation.isPending}
                  >
                    Восстановить
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-muted-foreground">
                  Удален:{" "}
                  {new Date(project.updated_at).toLocaleDateString("ru-RU")}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/trash")({
  component: TrashPage,
});
