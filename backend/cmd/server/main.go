package main

import (
	_ "github.com/valerii/hub/backend/docs"
)

// @title Hub Platform API
// @version 1.0
// @description Единая платформа для команд разработки и тестирования: трекер задач, мессенджер, TMS, созвоны.
// @host localhost:3000
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header — Swagger annotation, не переводим.
// @name Authorization
// @securityDefinitions.apikey APIKeyAuth
// @in header — Swagger annotation, не переводим.
// @name X-API-Key
func main() {
	d := &serverDeps{}
	defer d.runCleanups()

	bootstrapServer(d)
	wireInfra(d)
	wireFiber(d)
	wireAuth(d)
	wireCore(d)
	wireMessenger(d)
	wireNotification(d)
	wireAuditMailAI(d)
	wireModules(d)
	wireIntegrations(d)
	wireWebSocket(d)
	runServer(d)
}
