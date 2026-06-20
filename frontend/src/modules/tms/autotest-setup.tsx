import { useCallback, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TriggerCIPanel } from "./trigger-ci-dialog";

// ---------------------------------------------------------------------------
// Autotest Integration Setup Page
// ---------------------------------------------------------------------------

interface Props {
	projectId: string;
}

/** Компонент блока кода с кнопкой копирования. */
function CodeBlock({ code, language }: { code: string; language?: string }) {
	const [copied, setCopied] = useState(false);

	const handleCopy = useCallback(async () => {
		await navigator.clipboard.writeText(code);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	}, [code]);

	return (
		<div className="group relative">
			<pre className="overflow-x-auto rounded-lg bg-gray-950 dark:bg-gray-900 p-4 text-sm text-gray-100">
				{language && (
					<span className="mb-2 block text-xs text-[var(--t3)]">
						{language}
					</span>
				)}
				<code>{code}</code>
			</pre>
			<button
				type="button"
				onClick={handleCopy}
				className="absolute right-2 top-2 rounded-md bg-gray-800 px-2 py-1 text-xs text-gray-300 opacity-0 transition-opacity hover:bg-gray-700 group-hover:opacity-100"
			>
				{copied ? "Скопировано!" : "Копировать"}
			</button>
		</div>
	);
}

/** Раскрываемая секция. */
function CollapsibleSection({
	title,
	defaultOpen = false,
	children,
}: {
	title: string;
	defaultOpen?: boolean;
	children: React.ReactNode;
}) {
	const [open, setOpen] = useState(defaultOpen);

	return (
		<div className="rounded-lg border border-[var(--b1)]">
			<button
				type="button"
				onClick={() => setOpen(!open)}
				className="flex w-full items-center justify-between p-4 text-left font-medium text-[var(--t1)] hover:bg-[var(--hover)]"
			>
				<span>{title}</span>
				<svg
					aria-hidden="true"
					className={`h-5 w-5 text-[var(--t3)] transition-transform ${open ? "rotate-180" : ""}`}
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M19 9l-7 7-7-7"
					/>
				</svg>
			</button>
			{open && (
				<div className="border-t p-4 border-[var(--b1)]">{children}</div>
			)}
		</div>
	);
}

/** Страница настройки интеграции автотестов. */
export function AutotestSetup({ projectId }: Props) {
	const webhookUrl = `\${YOUR_DOMAIN}/api/projects/${projectId}/tms/webhook`;
	const junitUrl = `\${YOUR_DOMAIN}/api/projects/${projectId}/tms/webhook/junit`;

	return (
		<div className="space-y-6 p-6">
			{/* Заголовок */}
			<div>
				<h2 className="text-xl font-semibold">Интеграция автотестов</h2>
				<p className="mt-1 text-sm text-t2">
					Настройте CI/CD пайплайн для автоматической загрузки результатов
					тестов
				</p>
			</div>

			{/* Запуск тестов из Hub */}
			<TriggerCIPanel projectId={projectId} />

			{/* 1. Webhook URL */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							1
						</span>
						Webhook URL
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					<p className="text-sm text-t2">
						Отправляйте результаты тестов на один из эндпоинтов:
					</p>

					<div className="space-y-2">
						<div className="flex items-center gap-2">
							<Badge variant="outline">Allure JSON</Badge>
							<span className="text-xs text-t2">POST</span>
						</div>
						<CodeBlock code={webhookUrl} />
					</div>

					<div className="space-y-2">
						<div className="flex items-center gap-2">
							<Badge variant="outline">JUnit XML</Badge>
							<span className="text-xs text-t2">POST</span>
						</div>
						<CodeBlock code={junitUrl} />
					</div>

					<p className="text-xs text-t2">
						Замените{" "}
						<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
							{"${YOUR_DOMAIN}"}
						</code>{" "}
						на адрес вашего Hub-инстанса (например,{" "}
						<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
							https://hub.example.com
						</code>
						).
					</p>
				</CardContent>
			</Card>

			{/* 2. API Token */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							2
						</span>
						API-токен для аутентификации
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<p className="text-sm text-t2">
						Для аутентификации CI/CD используйте API-ключ. Его можно создать в
						разделе <strong>Профиль &rarr; API-ключи</strong>.
					</p>

					<CollapsibleSection title="Как создать API-ключ" defaultOpen>
						<ol className="list-decimal pl-4 space-y-2 text-sm text-t2">
							<li>
								Перейдите в <strong>Профиль</strong> (аватар в левом нижнем
								углу)
							</li>
							<li>
								Откройте вкладку <strong>API-ключи</strong>
							</li>
							<li>
								Нажмите <strong>Создать ключ</strong>, укажите название
								(например,{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									ci-autotest
								</code>
								)
							</li>
							<li>Скопируйте ключ — он показывается только один раз</li>
							<li>Добавьте ключ в секреты CI/CD вашего репозитория</li>
						</ol>
					</CollapsibleSection>

					<p className="text-sm text-t2">Передавайте ключ одним из способов:</p>

					<div className="space-y-2">
						<p className="text-sm font-medium text-[var(--t1)]">
							Заголовок (рекомендуется):
						</p>
						<CodeBlock code="X-API-Key: your-api-key-here" />
					</div>

					<div className="space-y-2">
						<p className="text-sm font-medium text-[var(--t1)]">
							Query-параметр (альтернативный):
						</p>
						<CodeBlock code={`${webhookUrl}?token=your-api-key-here`} />
					</div>

					<CollapsibleSection title="Необходимые права доступа">
						<div className="space-y-3">
							<p className="text-sm text-t2">
								API-ключ наследует права пользователя, который его создал. Для
								работы автотестов требуются следующие права в проекте:
							</p>
							<div className="overflow-x-auto">
								<table className="w-full text-sm">
									<thead>
										<tr className="border-b border-[var(--b1)]">
											<th className="p-2 text-left font-medium text-[var(--t1)]">
												Действие
											</th>
											<th className="p-2 text-left font-medium text-[var(--t1)]">
												Минимальная роль
											</th>
											<th className="p-2 text-left font-medium text-[var(--t1)]">
												Описание
											</th>
										</tr>
									</thead>
									<tbody>
										<tr className="border-b border-[var(--b1)]">
											<td className="p-2">Загрузка результатов</td>
											<td className="p-2">
												<Badge variant="outline">Tester</Badge>
											</td>
											<td className="p-2 text-t2">
												Отправка результатов через webhook (Allure / JUnit)
											</td>
										</tr>
										<tr className="border-b border-[var(--b1)]">
											<td className="p-2">Создание прогонов</td>
											<td className="p-2">
												<Badge variant="outline">Tester</Badge>
											</td>
											<td className="p-2 text-t2">
												Автоматическое создание test run при импорте
											</td>
										</tr>
										<tr className="border-b border-[var(--b1)]">
											<td className="p-2">Запуск пайплайна</td>
											<td className="p-2">
												<Badge variant="outline">Developer</Badge>
											</td>
											<td className="p-2 text-t2">
												Триггер CI/CD через Hub (кнопка "Запуск тестов")
											</td>
										</tr>
										<tr className="border-[var(--b1)]">
											<td className="p-2">Управление интеграциями</td>
											<td className="p-2">
												<Badge variant="outline">Admin</Badge>
											</td>
											<td className="p-2 text-t2">
												Подключение/отключение GitLab-репозиториев
											</td>
										</tr>
									</tbody>
								</table>
							</div>
						</div>
					</CollapsibleSection>

					<CollapsibleSection title="Настройка секретов в CI/CD">
						<div className="space-y-3">
							<p className="text-sm font-medium text-[var(--t1)]">
								GitHub Actions:
							</p>
							<p className="text-sm text-t2">
								<strong>
									Settings &rarr; Secrets and variables &rarr; Actions
								</strong>{" "}
								&rarr; создайте секрет{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									HUB_API_KEY
								</code>{" "}
								и переменную{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									HUB_URL
								</code>
								.
							</p>
							<p className="text-sm font-medium text-[var(--t1)]">GitLab CI:</p>
							<p className="text-sm text-t2">
								<strong>Settings &rarr; CI/CD &rarr; Variables</strong> &rarr;
								создайте переменные{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									HUB_API_KEY
								</code>{" "}
								(masked) и{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									HUB_URL
								</code>
								.
							</p>
							<p className="text-sm font-medium text-[var(--t1)]">Jenkins:</p>
							<p className="text-sm text-t2">
								<strong>Manage Jenkins &rarr; Credentials</strong> &rarr;
								добавьте Secret text с ID{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									hub-api-key
								</code>
								.
							</p>
						</div>
					</CollapsibleSection>

					<div className="rounded-lg border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-900 dark:bg-yellow-950 dark:text-yellow-200">
						<strong>Важно:</strong> Храните API-ключ в секретах CI/CD (например,{" "}
						<code className="rounded bg-yellow-100 px-1 dark:bg-yellow-900">
							$HUB_API_KEY
						</code>
						). Никогда не фиксируйте его в репозитории. Регулярно ротируйте
						ключи — при увольнении сотрудника отзовите его ключи в{" "}
						<strong>Профиль &rarr; API-ключи</strong>.
					</div>
				</CardContent>
			</Card>

			{/* 3. Формат ID тест-кейсов */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							3
						</span>
						Формат ID тест-кейсов
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					<p className="text-sm text-t2">
						Для автоматической привязки результатов к тест-кейсам используйте
						паттерн{" "}
						<code className="rounded bg-gray-100 px-1 font-semibold dark:bg-gray-800">
							PROJ-TC-N
						</code>
						, где{" "}
						<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
							PROJ
						</code>{" "}
						— префикс проекта, а{" "}
						<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">N</code>{" "}
						— номер тест-кейса.
					</p>

					<p className="text-sm text-t2">
						Пример: <Badge variant="secondary">HUB-TC-42</Badge>
					</p>

					<p className="text-sm text-t2">
						Включите этот идентификатор в название теста, аннотацию Allure или
						поле{" "}
						<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
							automation_id
						</code>{" "}
						тест-кейса.
					</p>
				</CardContent>
			</Card>

			{/* 4. Примеры кода для фреймворков */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							4
						</span>
						Примеры интеграции для фреймворков
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					<CollapsibleSection
						title="Python (pytest + allure-pytest)"
						defaultOpen
					>
						<div className="space-y-3">
							<p className="text-sm text-t2">
								Используйте декоратор{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									@allure.id()
								</code>{" "}
								или метку{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									allure.label()
								</code>
								:
							</p>
							<CodeBlock
								language="python"
								code={`import allure

@allure.id("PROJ-TC-1")
@allure.title("Авторизация пользователя")
def test_user_login():
    """Тест входа в систему."""
    # ... тестовый код ...
    pass

# Альтернативный способ через label:
@allure.label("AS_ID", "PROJ-TC-2")
def test_user_registration():
    pass`}
							/>
							<p className="text-sm text-t2">Генерация результатов:</p>
							<CodeBlock
								language="bash"
								code="pytest --alluredir=allure-results tests/"
							/>
						</div>
					</CollapsibleSection>

					<CollapsibleSection title="JavaScript (Jest / Playwright)">
						<div className="space-y-3">
							<p className="text-sm text-t2">
								Включите ключ тест-кейса в название теста:
							</p>
							<CodeBlock
								language="typescript"
								code={`// Jest
describe("Auth", () => {
  test("PROJ-TC-1: should login user with valid credentials", async () => {
    // ... тестовый код ...
  });
});

// Playwright
test("PROJ-TC-2: should show error on invalid password", async ({ page }) => {
  await page.goto("/login");
  // ... тестовый код ...
});`}
							/>
							<p className="text-sm text-t2">
								Для Playwright с Allure используйте:
							</p>
							<CodeBlock
								language="typescript"
								code={`import { allure } from "allure-playwright";

test("Login test", async ({ page }) => {
  await allure.id("PROJ-TC-1");
  await allure.label("AS_ID", "PROJ-TC-1");
  // ... тестовый код ...
});`}
							/>
						</div>
					</CollapsibleSection>

					<CollapsibleSection title="Java (JUnit + allure-java)">
						<div className="space-y-3">
							<p className="text-sm text-t2">
								Используйте аннотацию{" "}
								<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
									@AllureId
								</code>
								:
							</p>
							<CodeBlock
								language="java"
								code={`import io.qameta.allure.AllureId;
import io.qameta.allure.TmsLink;
import org.junit.jupiter.api.Test;

public class AuthTests {

    @Test
    @AllureId("PROJ-TC-1")
    @TmsLink("PROJ-TC-1")
    void shouldLoginWithValidCredentials() {
        // ... тестовый код ...
    }

    @Test
    @AllureId("PROJ-TC-2")
    void shouldRejectInvalidPassword() {
        // ... тестовый код ...
    }
}`}
							/>
						</div>
					</CollapsibleSection>
				</CardContent>
			</Card>

			{/* 5. CI/CD примеры */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							5
						</span>
						Настройка CI/CD
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					<CollapsibleSection title="GitHub Actions" defaultOpen>
						<CodeBlock
							language="yaml"
							code={`name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run tests
        run: pytest --alluredir=allure-results tests/

      - name: Upload results to Hub
        if: always()
        run: |
          # Собираем JSON-результаты в массив
          echo "[" > results.json
          first=true
          for f in allure-results/*-result.json; do
            [ -f "$f" ] || continue
            if [ "$first" = true ]; then first=false; else echo "," >> results.json; fi
            cat "$f" >> results.json
          done
          echo "]" >> results.json

          curl -X POST \\
            "\${HUB_URL}/api/projects/${projectId}/tms/webhook" \\
            -H "X-API-Key: \${{ secrets.HUB_API_KEY }}" \\
            -H "Content-Type: application/json" \\
            -d @results.json
        env:
          HUB_URL: \${{ vars.HUB_URL }}`}
						/>

						<p className="mt-3 text-sm text-t2">
							Для JUnit XML (например, pytest --junitxml):
						</p>
						<CodeBlock
							language="yaml"
							code={`      - name: Upload JUnit results to Hub
        if: always()
        run: |
          curl -X POST \\
            "\${HUB_URL}/api/projects/${projectId}/tms/webhook/junit" \\
            -H "X-API-Key: \${{ secrets.HUB_API_KEY }}" \\
            -H "Content-Type: application/xml" \\
            --data-binary @test-results.xml
        env:
          HUB_URL: \${{ vars.HUB_URL }}`}
						/>
					</CollapsibleSection>

					<CollapsibleSection title="GitLab CI">
						<CodeBlock
							language="yaml"
							code={`stages:
  - test
  - report

test:
  stage: test
  script:
    - pytest --alluredir=allure-results --junitxml=junit-report.xml tests/
  artifacts:
    paths:
      - allure-results/
      - junit-report.xml
    when: always

upload_results:
  stage: report
  when: always
  needs: [test]
  script:
    # Вариант 1: Allure JSON
    - |
      echo "[" > results.json
      first=true
      for f in allure-results/*-result.json; do
        [ -f "$f" ] || continue
        if [ "$first" = true ]; then first=false; else echo "," >> results.json; fi
        cat "$f" >> results.json
      done
      echo "]" >> results.json

      curl -X POST \\
        "\${HUB_URL}/api/projects/${projectId}/tms/webhook" \\
        -H "X-API-Key: \${HUB_API_KEY}" \\
        -H "Content-Type: application/json" \\
        -d @results.json

    # Вариант 2: JUnit XML
    - |
      curl -X POST \\
        "\${HUB_URL}/api/projects/${projectId}/tms/webhook/junit" \\
        -H "X-API-Key: \${HUB_API_KEY}" \\
        -H "Content-Type: application/xml" \\
        --data-binary @junit-report.xml`}
						/>

						<p className="mt-3 text-sm text-t2">
							Добавьте переменные{" "}
							<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
								HUB_URL
							</code>{" "}
							и{" "}
							<code className="rounded bg-gray-100 dark:bg-gray-800 px-1">
								HUB_API_KEY
							</code>{" "}
							в<strong> Settings &rarr; CI/CD &rarr; Variables</strong>{" "}
							(masked).
						</p>
					</CollapsibleSection>
				</CardContent>
			</Card>

			{/* 6. Как работает матчинг */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							6
						</span>
						Как работает привязка результатов
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<p className="text-sm text-t2">
						Hub автоматически привязывает результаты автотестов к тест-кейсам,
						используя 6 стратегий в порядке приоритета:
					</p>

					<div className="overflow-x-auto">
						<table className="w-full text-sm">
							<thead>
								<tr className="border-b border-[var(--b1)]">
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										#
									</th>
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										Стратегия
									</th>
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										Источник данных
									</th>
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										Описание
									</th>
								</tr>
							</thead>
							<tbody>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<Badge className="bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">
											1
										</Badge>
									</td>
									<td className="p-2 font-medium">annotation</td>
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											@AllureId / @TmsLink / label AS_ID
										</code>
									</td>
									<td className="p-2 text-t2">
										Самый надёжный способ. Ищет по external_id тест-кейса.
									</td>
								</tr>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<Badge className="bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">
											2
										</Badge>
									</td>
									<td className="p-2 font-medium">automation_id</td>
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											label automation_id / @TmsLink
										</code>
									</td>
									<td className="p-2 text-t2">
										Поле automation_id тест-кейса, заданное QA-инженером.
									</td>
								</tr>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<Badge className="bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">
											3
										</Badge>
									</td>
									<td className="p-2 font-medium">case_key</td>
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											Паттерн PROJ-TC-N в имени
										</code>
									</td>
									<td className="p-2 text-t2">
										Регулярное выражение ищет ключ тест-кейса в названии теста.
									</td>
								</tr>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<Badge className="bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300">
											4
										</Badge>
									</td>
									<td className="p-2 font-medium">fullname</td>
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											fullName результата
										</code>
									</td>
									<td className="p-2 text-t2">
										Полное имя теста (пакет.класс.метод) сравнивается с
										full_name кейса.
									</td>
								</tr>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<Badge className="bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300">
											5
										</Badge>
									</td>
									<td className="p-2 font-medium">classmethod</td>
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											suite label + name
										</code>
									</td>
									<td className="p-2 text-t2">
										Совпадение по имени класса (label suite) и имени метода
										(name).
									</td>
								</tr>
								<tr className="border-[var(--b1)]">
									<td className="p-2">
										<Badge className="bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400">
											-
										</Badge>
									</td>
									<td className="p-2 font-medium text-t2">unmapped</td>
									<td className="p-2 text-t2">-</td>
									<td className="p-2 text-t2">
										Результат не привязан ни к одному тест-кейсу. Отображается в
										прогоне без ссылки на кейс.
									</td>
								</tr>
							</tbody>
						</table>
					</div>

					<div className="rounded-lg border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-200">
						<strong>Совет:</strong> Для максимальной надёжности используйте
						стратегию <strong>annotation</strong> — добавьте{" "}
						<code className="rounded bg-blue-100 px-1 dark:bg-blue-900">
							@allure.id("PROJ-TC-N")
						</code>{" "}
						или{" "}
						<code className="rounded bg-blue-100 px-1 dark:bg-blue-900">
							@AllureId("PROJ-TC-N")
						</code>{" "}
						к каждому тесту.
					</div>
				</CardContent>
			</Card>

			{/* 7. Запуск из Hub */}
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2 text-base text-[var(--t1)]">
						<span className="flex h-6 w-6 items-center justify-center rounded-full bg-[var(--blue)] text-xs text-white">
							7
						</span>
						Запуск тестов из Hub
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					<p className="text-sm text-t2">
						При запуске тестов через панель выше Hub передаёт переменные CI/CD в
						пайплайн:
					</p>

					<div className="overflow-x-auto">
						<table className="w-full text-sm">
							<thead>
								<tr className="border-b border-[var(--b1)]">
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										Переменная
									</th>
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										Пример
									</th>
									<th className="p-2 text-left font-medium text-[var(--t1)]">
										Описание
									</th>
								</tr>
							</thead>
							<tbody>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											HUB_TEST_CASE_IDS
										</code>
									</td>
									<td className="p-2 text-t2">
										<code className="text-xs">HUB-TC-1,HUB-TC-5,HUB-TC-12</code>
									</td>
									<td className="p-2 text-t2">
										Ключи выбранных тест-кейсов (через запятую)
									</td>
								</tr>
								<tr className="border-b border-[var(--b1)]">
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											HUB_RUN_SOURCE
										</code>
									</td>
									<td className="p-2 text-t2">
										<code className="text-xs">hub_trigger</code>
									</td>
									<td className="p-2 text-t2">
										Источник запуска (для фильтрации в пайплайне)
									</td>
								</tr>
								<tr className="border-[var(--b1)]">
									<td className="p-2">
										<code className="rounded bg-gray-100 px-1 text-xs dark:bg-gray-800">
											HUB_PROJECT_ID
										</code>
									</td>
									<td className="p-2 text-t2">
										<code className="text-xs">uuid</code>
									</td>
									<td className="p-2 text-t2">
										ID проекта в Hub (для обратной привязки)
									</td>
								</tr>
							</tbody>
						</table>
					</div>

					<CollapsibleSection title="Пример использования в pytest">
						<CodeBlock
							language="python"
							code={`import os
import pytest

# Получаем список кейсов из переменной CI
hub_case_ids = os.environ.get("HUB_TEST_CASE_IDS", "")

def filter_by_hub_ids(items):
    """Фильтрует тесты по HUB_TEST_CASE_IDS, если переменная задана."""
    if not hub_case_ids:
        return  # Запускаем все тесты

    allowed = set(hub_case_ids.split(","))
    selected = []
    deselected = []

    for item in items:
        # Ищем @allure.id() маркер или ID в имени теста
        markers = [m for m in item.iter_markers("allure_id")]
        test_id = markers[0].args[0] if markers else None

        if test_id and test_id in allowed:
            selected.append(item)
        else:
            deselected.append(item)

    items[:] = selected
    config = items[0].config if items else None
    if config and deselected:
        config.hook.pytest_deselected(items=deselected)

def pytest_collection_modifyitems(items):
    filter_by_hub_ids(items)`}
						/>
					</CollapsibleSection>

					<CollapsibleSection title="Пример использования в GitLab CI">
						<CodeBlock
							language="yaml"
							code={`test:
  stage: test
  script:
    - |
      if [ -n "$HUB_TEST_CASE_IDS" ]; then
        echo "Running selected cases: $HUB_TEST_CASE_IDS"
        pytest --alluredir=allure-results tests/ -k "$HUB_TEST_CASE_IDS"
      else
        echo "Running all tests"
        pytest --alluredir=allure-results tests/
      fi
  rules:
    # Запуск по триггеру из Hub
    - if: $HUB_RUN_SOURCE == "hub_trigger"
    # Обычный запуск по push
    - if: $CI_PIPELINE_SOURCE == "push"`}
						/>
					</CollapsibleSection>
				</CardContent>
			</Card>
		</div>
	);
}
