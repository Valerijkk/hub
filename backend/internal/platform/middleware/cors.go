package middleware

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// CORS возвращает настроенный middleware для обработки CORS-заголовков.
// Источники задаются через переменную окружения CORS_ORIGINS (через запятую).
// По умолчанию разрешён только localhost dev-сервер.
//
// HI-11:
//   - значения тримятся; пустые/только-пробельные элементы игнорируются;
//   - "*" отвергается (нельзя сочетать с AllowCredentials=true);
//   - если в production (DEV_MODE != "true") после фильтрации не осталось
//     ни одного валидного origin — fail-start через os.Exit(1) с понятным
//     логом. Это аналог нашего обычного pattern'а startup-fatal в main.go,
//     но локализован тут, чтобы caller не должен был знать о валидации.
func CORS() fiber.Handler {
	origins, problems := parseCORSOrigins(os.Getenv("CORS_ORIGINS"))

	for _, p := range problems {
		slog.Warn("CORS: ignoring invalid origin", "reason", p)
	}

	if len(origins) == 0 {
		// Fallback to localhost defaults — same behaviour as before.
		origins = []string{"http://localhost:5173", "http://localhost:8080"}

		// In production no CORS_ORIGINS at all is almost certainly a deploy
		// misconfiguration. Fail loudly rather than silently expose only
		// localhost from a public domain.
		if !isDevMode() {
			slog.Error("CORS_ORIGINS не задан в production — небезопасная конфигурация",
				"hint", "set CORS_ORIGINS=https://yourdomain.example to allow your frontend")
			os.Exit(1)
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-API-Key", "X-Request-ID"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	})
}

// parseCORSOrigins разбирает значение CORS_ORIGINS и возвращает (валидные,
// problems). problems — человекочитаемые сообщения о том, что было откинуто.
// Вынесено в чистую функцию, чтобы покрывалось unit-тестами без os.Exit.
func parseCORSOrigins(env string) (origins []string, problems []string) {
	if env == "" {
		return nil, nil
	}
	for _, raw := range strings.Split(env, ",") {
		o := strings.TrimSpace(raw)
		if o == "" {
			problems = append(problems, "пустое значение")
			continue
		}
		if o == "*" {
			problems = append(problems, "wildcard '*' несовместим с AllowCredentials=true")
			continue
		}
		origins = append(origins, o)
	}
	return origins, problems
}

// isDevMode возвращает true, если переменная окружения DEV_MODE имеет
// значение "true" (любой регистр) или "1". Совпадает с тем, как
// config.getEnvBool парсит булевые значения.
func isDevMode() bool {
	v := strings.TrimSpace(os.Getenv("DEV_MODE"))
	switch strings.ToLower(v) {
	case "true", "1", "yes", "on":
		return true
	}
	return false
}
