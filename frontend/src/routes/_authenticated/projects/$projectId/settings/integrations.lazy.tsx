import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createLazyFileRoute } from "@tanstack/react-router";
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
							<div className="h-10 w-full rounded bg-raised" />
						</div>
					)}

					{!isLoading && (!integrations || integrations.length === 0) && (
						<p className="text-sm text-t2">
							Нет подключённых GitHub-репозиториев
						</p>
					)}

					{integrations && integrations.length > 0 && (
						<div className="space-y-2">
							{integrations.map(
								(ig: GitHubIntegrationResponse) => (
									<div
										key={ig.id}
										className="flex items-center justify-between bg-gradient-to-br from-card to-raised border border-b1 rounded-[var(--r3)] px-4 py-3 hover:border-b2 transition-colors duration-200"
									>
										<div className="flex items-center gap-2">
											<GitHubIcon className="h-5 w-5 text-t2" />
											<span className="text-sm font-medium">
												{ig.repo_owner}/{ig.repo_name}
											</span>
										</div>
										<Button
											variant="ghost"
											size="sm"
											className="text-red hover:text-red/80"
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
					<div className="space-y-2">
						<Label>Access Token</Label>
						<Input
							type="password"
							value={accessToken}
							onChange={(e) => setAccessToken(e.target.value)}
							placeholder="ghp_..."
						/>
						<details className="group">
							<summary className="cursor-pointer text-[12px] text-blue hover:underline">
								Как создать токен в GitHub?
							</summary>
							<div className="mt-2 rounded-lg border border-blue-200 bg-blue-50 p-3 text-xs text-blue-800 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-200 space-y-2">
								<ol className="list-decimal pl-4 space-y-1">
									<li>Перейдите в <strong>GitHub &rarr; Settings &rarr; Developer settings &rarr; Personal access tokens &rarr; Fine-grained tokens</strong></li>
									<li>Нажмите <strong>Generate new token</strong></li>
									<li>Укажите название (например, <code className="rounded bg-blue-100 px-1 dark:bg-blue-900">hub-integration</code>)</li>
									<li>В <strong>Repository access</strong> выберите нужный репозиторий</li>
									<li>В <strong>Permissions</strong> выдайте:
										<ul className="list-disc pl-4 mt-1">
											<li><strong>Contents</strong>: Read (для веток и коммитов)</li>
											<li><strong>Pull requests</strong>: Read &amp; Write (для линковки PR)</li>
											<li><strong>Webhooks</strong>: Read &amp; Write (для подписки на события)</li>
											<li><strong>Actions</strong>: Read (для статуса пайплайнов)</li>
											<li><strong>Metadata</strong>: Read (обязательно)</li>
										</ul>
									</li>
									<li>Нажмите <strong>Generate token</strong> и скопируйте</li>
								</ol>
								<p className="mt-1 font-medium">Classic token (альтернатива): скоупы <code className="rounded bg-blue-100 px-1 dark:bg-blue-900">repo</code>, <code className="rounded bg-blue-100 px-1 dark:bg-blue-900">admin:repo_hook</code></p>
							</div>
						</details>
					</div>
					<div className="space-y-2">
						<Label>Webhook Secret</Label>
						<Input
							type="password"
							value={webhookSecret}
							onChange={(e) => setWebhookSecret(e.target.value)}
							placeholder="Секрет для валидации webhook"
						/>
						<p className="text-xs text-t2">
							Настройте Webhook в <strong>GitHub &rarr; Settings &rarr; Webhooks</strong>.
							URL: <code className="rounded-[var(--r1)] bg-input px-1">{"${HUB_URL}"}/api/github/webhook</code>.
							Укажите тот же секрет. Events: Push, Pull requests, Workflow runs.
						</p>
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
							<div className="h-10 w-full rounded bg-raised" />
						</div>
					)}

					{!isLoading && (!integrations || integrations.length === 0) && (
						<p className="text-sm text-t2">
							Нет подключённых GitLab-репозиториев
						</p>
					)}

					{integrations && integrations.length > 0 && (
						<div className="space-y-2">
							{integrations.map(
								(ig: GitLabIntegrationResponse) => (
									<div
										key={ig.id}
										className="flex items-center justify-between bg-gradient-to-br from-card to-raised border border-b1 rounded-[var(--r3)] px-4 py-3 hover:border-b2 transition-colors duration-200"
									>
										<div className="flex items-center gap-2">
											<GitLabIcon className="h-5 w-5 text-orange-600 dark:text-orange-400" />
											<div className="flex flex-col">
												<span className="text-sm font-medium">
													{ig.project_path}
												</span>
												<span className="text-xs text-t2">
													{ig.gitlab_url}
												</span>
											</div>
										</div>
										<Button
											variant="ghost"
											size="sm"
											className="text-red hover:text-red/80"
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
						<p className="mt-1 text-xs text-t2">
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
					<div className="space-y-2">
						<Label>Access Token</Label>
						<Input
							type="password"
							value={accessToken}
							onChange={(e) => setAccessToken(e.target.value)}
							placeholder="glpat-..."
						/>
						<details className="group">
							<summary className="cursor-pointer text-[12px] text-blue hover:underline">
								Как создать токен в GitLab?
							</summary>
							<div className="mt-2 rounded-lg border border-orange-200 bg-orange-50 p-3 text-xs text-orange-800 dark:border-orange-900 dark:bg-orange-950 dark:text-orange-200 space-y-2">
								<p className="font-medium">Вариант 1: Project Access Token (рекомендуется)</p>
								<ol className="list-decimal pl-4 space-y-1">
									<li>Перейдите в <strong>GitLab &rarr; Проект &rarr; Settings &rarr; Access Tokens</strong></li>
									<li>Нажмите <strong>Add new token</strong></li>
									<li>Укажите название (например, <code className="rounded bg-orange-100 px-1 dark:bg-orange-900">hub-integration</code>)</li>
									<li>Роль: <strong>Developer</strong> или выше</li>
									<li>Выберите скоупы:
										<ul className="list-disc pl-4 mt-1">
											<li><strong>api</strong> — полный доступ к API (пайплайны, ветки, MR)</li>
											<li><strong>read_repository</strong> — чтение кода и коммитов</li>
										</ul>
									</li>
									<li>Нажмите <strong>Create project access token</strong></li>
								</ol>
								<p className="mt-2 font-medium">Вариант 2: Personal Access Token</p>
								<ol className="list-decimal pl-4 space-y-1">
									<li>Перейдите в <strong>GitLab &rarr; User Settings &rarr; Access Tokens</strong></li>
									<li>Создайте токен со скоупами: <code className="rounded bg-orange-100 px-1 dark:bg-orange-900">api</code>, <code className="rounded bg-orange-100 px-1 dark:bg-orange-900">read_repository</code></li>
								</ol>
								<p className="mt-1 text-orange-600 dark:text-orange-300">Для self-hosted GitLab убедитесь, что URL инстанса указан корректно выше.</p>
							</div>
						</details>
					</div>
					<div className="space-y-2">
						<Label>Webhook Secret (опционально)</Label>
						<Input
							type="password"
							value={webhookSecret}
							onChange={(e) => setWebhookSecret(e.target.value)}
							placeholder="Секрет для валидации webhook"
						/>
						<p className="text-xs text-t2">
							Настройте Webhook в <strong>GitLab &rarr; Settings &rarr; Webhooks</strong>.
							URL: <code className="rounded-[var(--r1)] bg-input px-1">{"${HUB_URL}"}/api/gitlab/webhook</code>.
							Укажите токен в поле <strong>Secret token</strong>. Triggers: Push, Merge request, Pipeline.
						</p>
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
		<div className="mx-auto max-w-3xl space-y-6 p-6 animate-[eu_450ms_var(--spring)_both]">
			<div className="mb-9">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] text-t1">Интеграции</h2>
				<p className="text-[13px] text-t2 mt-1">
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

export const Route = createLazyFileRoute(
	"/_authenticated/projects/$projectId/settings/integrations",
)({
	component: IntegrationsSettingsPage,
});
