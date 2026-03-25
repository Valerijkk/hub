import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import {
	createIntegration as createGitHubIntegration,
	deleteIntegration as deleteGitHubIntegration,
	listIntegrations as listGitHubIntegrations,
	type IntegrationResponse as GitHubIntegrationResponse,
} from "@/api/github";
import {
	connectGitLab,
	disconnectGitLab,
	listGitLabIntegrations,
	type GitLabIntegrationResponse,
} from "@/api/gitlab";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

// ---------------------------------------------------------------------------
// GitHub секция
// ---------------------------------------------------------------------------

function GitHubSection({ projectId }: { projectId: string }) {
	const queryClient = useQueryClient();
	const [repoOwner, setRepoOwner] = useState("");
	const [repoName, setRepoName] = useState("");
	const [accessToken, setAccessToken] = useState("");
	const [webhookSecret, setWebhookSecret] = useState("");

	const { data: integrations, isLoading } = useQuery({
		queryKey: ["github-integrations", projectId],
		queryFn: () => listGitHubIntegrations(projectId),
	});

	const createMutation = useMutation({
		mutationFn: () =>
			createGitHubIntegration(projectId, {
				repo_owner: repoOwner,
				repo_name: repoName,
				access_token: accessToken,
				webhook_secret: webhookSecret,
			}),
		onSuccess: () => {
			toast.success("GitHub-репозиторий подключён");
			queryClient.invalidateQueries({
				queryKey: ["github-integrations", projectId],
			});
			setRepoOwner("");
			setRepoName("");
			setAccessToken("");
			setWebhookSecret("");
		},
		onError: () => {
			toast.error("Ошибка подключения GitHub-репозитория");
		},
	});

	const deleteMutation = useMutation({
		mutationFn: (integrationId: string) =>
			deleteGitHubIntegration(projectId, integrationId),
		onSuccess: () => {
			toast.success("Интеграция удалена");
			queryClient.invalidateQueries({
				queryKey: ["github-integrations", projectId],
			});
		},
		onError: () => {
			toast.error("Ошибка удаления интеграции");
		},
	});

	return (
		<div className="space-y-6">
			{/* Подключённые репозитории */}
			<Card>
				<CardHeader>
					<CardTitle className="text-base">
						Подключённые репозитории
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading && (
						<div className="animate-pulse space-y-2">
							<div className="h-10 w-full rounded bg-gray-200 dark:bg-gray-700" />
						</div>
					)}

					{!isLoading && (!integrations || integrations.length === 0) && (
						<p className="text-sm text-gray-500 dark:text-gray-400">
							Нет подключённых GitHub-репозиториев
						</p>
					)}

					{integrations && integrations.length > 0 && (
						<div className="space-y-2">
							{integrations.map(
								(ig: GitHubIntegrationResponse) => (
									<div
										key={ig.id}
										className="flex items-center justify-between rounded-lg border px-4 py-3 dark:border-gray-700"
									>
										<div className="flex items-center gap-2">
											<GitHubIcon className="h-5 w-5 text-gray-600 dark:text-gray-400" />
											<span className="text-sm font-medium">
												{ig.repo_owner}/{ig.repo_name}
											</span>
										</div>
										<Button
											variant="ghost"
											size="sm"
											className="text-red-600 hover:text-red-700 dark:text-red-400"
											onClick={() =>
												deleteMutation.mutate(ig.id)
											}
											disabled={deleteMutation.isPending}
										>
											Отключить
										</Button>
									</div>
								),
							)}
						</div>
					)}
				</CardContent>
			</Card>

			{/* Форма подключения */}
			<Card>
				<CardHeader>
					<CardTitle className="text-base">
						Подключить репозиторий
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="grid grid-cols-2 gap-4">
						<div>
							<Label>Владелец (owner)</Label>
							<Input
								value={repoOwner}
								onChange={(e) => setRepoOwner(e.target.value)}
								placeholder="например, my-org"
							/>
						</div>
						<div>
							<Label>Имя репозитория</Label>
							<Input
								value={repoName}
								onChange={(e) => setRepoName(e.target.value)}
								placeholder="например, my-repo"
							/>
						</div>
					</div>
					<div>
						<Label>Access Token</Label>
						<Input
							type="password"
							value={accessToken}
							onChange={(e) => setAccessToken(e.target.value)}
							placeholder="ghp_..."
						/>
					</div>
					<div>
						<Label>Webhook Secret</Label>
						<Input
							type="password"
							value={webhookSecret}
							onChange={(e) => setWebhookSecret(e.target.value)}
							placeholder="Секрет для валидации webhook"
						/>
					</div>
					<Button
						onClick={() => createMutation.mutate()}
						disabled={
							createMutation.isPending ||
							!repoOwner ||
							!repoName ||
							!accessToken ||
							!webhookSecret
						}
					>
						{createMutation.isPending
							? "Подключение..."
							: "Подключить"}
					</Button>
				</CardContent>
			</Card>
		</div>
	);
}

// ---------------------------------------------------------------------------
// GitLab секция
// ---------------------------------------------------------------------------

function GitLabSection({ projectId }: { projectId: string }) {
	const queryClient = useQueryClient();
	const [gitlabUrl, setGitlabUrl] = useState("https://gitlab.com");
	const [projectPath, setProjectPath] = useState("");
	const [accessToken, setAccessToken] = useState("");
	const [webhookSecret, setWebhookSecret] = useState("");

	const { data: integrations, isLoading } = useQuery({
		queryKey: ["gitlab-integrations", projectId],
		queryFn: () => listGitLabIntegrations(projectId),
	});

	const createMutation = useMutation({
		mutationFn: () =>
			connectGitLab(projectId, {
				gitlab_url: gitlabUrl,
				project_path: projectPath,
				access_token: accessToken,
				webhook_secret: webhookSecret || undefined,
			}),
		onSuccess: () => {
			toast.success("GitLab-репозиторий подключён");
			queryClient.invalidateQueries({
				queryKey: ["gitlab-integrations", projectId],
			});
			setProjectPath("");
			setAccessToken("");
			setWebhookSecret("");
		},
		onError: () => {
			toast.error("Ошибка подключения GitLab-репозитория");
		},
	});

	const deleteMutation = useMutation({
		mutationFn: (integrationId: string) =>
			disconnectGitLab(projectId, integrationId),
		onSuccess: () => {
			toast.success("Интеграция удалена");
			queryClient.invalidateQueries({
				queryKey: ["gitlab-integrations", projectId],
			});
		},
		onError: () => {
			toast.error("Ошибка удаления интеграции");
		},
	});

	return (
		<div className="space-y-6">
			{/* Подключённые репозитории */}
			<Card>
				<CardHeader>
					<CardTitle className="text-base">
						Подключённые репозитории
					</CardTitle>
				</CardHeader>
				<CardContent>
					{isLoading && (
						<div className="animate-pulse space-y-2">
							<div className="h-10 w-full rounded bg-gray-200 dark:bg-gray-700" />
						</div>
					)}

					{!isLoading && (!integrations || integrations.length === 0) && (
						<p className="text-sm text-gray-500 dark:text-gray-400">
							Нет подключённых GitLab-репозиториев
						</p>
					)}

					{integrations && integrations.length > 0 && (
						<div className="space-y-2">
							{integrations.map(
								(ig: GitLabIntegrationResponse) => (
									<div
										key={ig.id}
										className="flex items-center justify-between rounded-lg border px-4 py-3 dark:border-gray-700"
									>
										<div className="flex items-center gap-2">
											<GitLabIcon className="h-5 w-5 text-orange-600 dark:text-orange-400" />
											<div className="flex flex-col">
												<span className="text-sm font-medium">
													{ig.project_path}
												</span>
												<span className="text-xs text-gray-500 dark:text-gray-400">
													{ig.gitlab_url}
												</span>
											</div>
										</div>
										<Button
											variant="ghost"
											size="sm"
											className="text-red-600 hover:text-red-700 dark:text-red-400"
											onClick={() =>
												deleteMutation.mutate(ig.id)
											}
											disabled={deleteMutation.isPending}
										>
											Отключить
										</Button>
									</div>
								),
							)}
						</div>
					)}
				</CardContent>
			</Card>

			{/* Форма подключения */}
			<Card>
				<CardHeader>
					<CardTitle className="text-base">
						Подключить репозиторий
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<div>
						<Label>URL GitLab</Label>
						<Input
							value={gitlabUrl}
							onChange={(e) => setGitlabUrl(e.target.value)}
							placeholder="https://gitlab.com"
						/>
						<p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
							Для self-hosted GitLab укажите URL вашего инстанса
						</p>
					</div>
					<div>
						<Label>Путь проекта (Project Path)</Label>
						<Input
							value={projectPath}
							onChange={(e) => setProjectPath(e.target.value)}
							placeholder="например, my-group/my-repo"
						/>
					</div>
					<div>
						<Label>Access Token</Label>
						<Input
							type="password"
							value={accessToken}
							onChange={(e) => setAccessToken(e.target.value)}
							placeholder="glpat-..."
						/>
					</div>
					<div>
						<Label>Webhook Secret (опционально)</Label>
						<Input
							type="password"
							value={webhookSecret}
							onChange={(e) => setWebhookSecret(e.target.value)}
							placeholder="Секрет для валидации webhook"
						/>
					</div>
					<Button
						onClick={() => createMutation.mutate()}
						disabled={
							createMutation.isPending ||
							!projectPath ||
							!accessToken
						}
					>
						{createMutation.isPending
							? "Подключение..."
							: "Подключить"}
					</Button>
				</CardContent>
			</Card>
		</div>
	);
}

// ---------------------------------------------------------------------------
// Иконки
// ---------------------------------------------------------------------------

function GitHubIcon({ className }: { className?: string }) {
	return (
		<svg
			className={className}
			viewBox="0 0 16 16"
			fill="currentColor"
		>
			<path d="M8 0c4.42 0 8 3.58 8 8a8.013 8.013 0 0 1-5.45 7.59c-.4.08-.55-.17-.55-.38 0-.27.01-1.13.01-2.2 0-.75-.25-1.23-.54-1.48 1.78-.2 3.65-.88 3.65-3.95 0-.88-.31-1.59-.82-2.15.08-.2.36-1.02-.08-2.12 0 0-.67-.22-2.2.82-.64-.18-1.32-.27-2-.27-.68 0-1.36.09-2 .27-1.53-1.03-2.2-.82-2.2-.82-.44 1.1-.16 1.92-.08 2.12-.51.56-.82 1.28-.82 2.15 0 3.06 1.86 3.75 3.64 3.95-.23.2-.44.55-.51 1.07-.46.21-1.61.55-2.33-.66-.15-.24-.6-.83-1.23-.82-.67.01-.27.38.01.53.34.19.73.9.82 1.13.16.45.68 1.31 2.69.94 0 .67.01 1.3.01 1.49 0 .21-.15.45-.55.38A7.995 7.995 0 0 1 0 8c0-4.42 3.58-8 8-8Z" />
		</svg>
	);
}

function GitLabIcon({ className }: { className?: string }) {
	return (
		<svg
			className={className}
			viewBox="0 0 16 16"
			fill="currentColor"
		>
			<path d="M8 14.932l2.358-7.26H5.642L8 14.932z" />
			<path
				d="M8 14.932L5.642 7.672H1.21L8 14.932z"
				opacity="0.7"
			/>
			<path
				d="M1.21 7.672L.124 11.013a.737.737 0 0 0 .268.824L8 14.932 1.21 7.672z"
				opacity="0.5"
			/>
			<path
				d="M1.21 7.672h4.432L3.598 1.296a.37.37 0 0 0-.703 0L1.21 7.672z"
			/>
			<path
				d="M8 14.932l2.358-7.26h4.432L8 14.932z"
				opacity="0.7"
			/>
			<path
				d="M14.79 7.672l1.086 3.341a.737.737 0 0 1-.268.824L8 14.932l6.79-7.26z"
				opacity="0.5"
			/>
			<path
				d="M14.79 7.672H10.358l2.044-6.376a.37.37 0 0 1 .703 0l1.685 6.376z"
			/>
		</svg>
	);
}

// ---------------------------------------------------------------------------
// Основная страница
// ---------------------------------------------------------------------------

/**
 * Страница настроек интеграций проекта.
 * Вкладки GitHub и GitLab для управления подключёнными репозиториями.
 */
function IntegrationsSettingsPage() {
	const { projectId } = Route.useParams();

	return (
		<div className="mx-auto max-w-3xl space-y-6 p-6">
			<div>
				<h2 className="text-xl font-semibold">Интеграции</h2>
				<p className="text-sm text-gray-500 dark:text-gray-400">
					Подключение репозиториев GitHub и GitLab для автоматической
					привязки коммитов, PR/MR и пайплайнов к задачам
				</p>
			</div>

			<Tabs defaultValue="github">
				<TabsList>
					<TabsTrigger value="github" className="gap-1.5">
						<GitHubIcon className="h-4 w-4" />
						GitHub
					</TabsTrigger>
					<TabsTrigger value="gitlab" className="gap-1.5">
						<GitLabIcon className="h-4 w-4" />
						GitLab
					</TabsTrigger>
				</TabsList>
				<TabsContent value="github" className="mt-4">
					<GitHubSection projectId={projectId} />
				</TabsContent>
				<TabsContent value="gitlab" className="mt-4">
					<GitLabSection projectId={projectId} />
				</TabsContent>
			</Tabs>
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/settings/integrations",
)({
	component: IntegrationsSettingsPage,
});
