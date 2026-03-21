// hub-cli — утилита командной строки для интеграции Hub с CI/CD.
// Позволяет загружать результаты тестов, создавать баги и обновлять задачи.
package main

import (
	"os"

	"github.com/valerii/hub/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
