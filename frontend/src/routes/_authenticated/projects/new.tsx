import { useState, type FormEvent } from "react";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { createProject } from "@/api/projects";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function generateKey(name: string): string {
  return name
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 10);
}

function NewProjectPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [name, setName] = useState("");
  const [key, setKey] = useState("");
  const [description, setDescription] = useState("");
  const [keyManuallyEdited, setKeyManuallyEdited] = useState(false);

  const mutation = useMutation({
    mutationFn: createProject,
    onSuccess: (project) => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Проект создан");
      void navigate({ to: "/projects/$projectId", params: { projectId: project.id } });
    },
    onError: (err: unknown) => {
      const message =
        err && typeof err === "object" && "message" in err
          ? (err as { message: string }).message
          : "Не удалось создать проект";
      toast.error(message);
    },
  });

  function handleNameChange(value: string) {
    setName(value);
    if (!keyManuallyEdited) {
      setKey(generateKey(value));
    }
  }

  function handleKeyChange(value: string) {
    setKeyManuallyEdited(true);
    setKey(value.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 10));
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!name.trim() || key.length < 2) {
      toast.error("Заполните название и ключ (2-10 символов)");
      return;
    }
    mutation.mutate({ name: name.trim(), key, description: description.trim() || undefined });
  }

  return (
    <div className="mx-auto max-w-lg">
      <Card>
        <CardHeader>
          <CardTitle>Новый проект</CardTitle>
          <CardDescription>
            Создайте проект для команды
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="proj-name">Название</Label>
              <Input
                id="proj-name"
                placeholder="Мой проект"
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="proj-key">Ключ</Label>
              <Input
                id="proj-key"
                placeholder="PROJ"
                value={key}
                onChange={(e) => handleKeyChange(e.target.value)}
                maxLength={10}
              />
              <p className="text-xs text-muted-foreground">
                2-10 символов, A-Z и цифры. Используется в нумерации задач ({key || "KEY"}-1)
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="proj-desc">Описание</Label>
              <Textarea
                id="proj-desc"
                placeholder="Описание проекта (необязательно)"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
              />
            </div>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => void navigate({ to: "/projects" })}
              >
                Отмена
              </Button>
              <Button
                type="submit"
                disabled={mutation.isPending || !name.trim() || key.length < 2}
              >
                {mutation.isPending ? "Создание..." : "Создать"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/projects/new")({
  component: NewProjectPage,
});
