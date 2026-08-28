package main

import devmod "github.com/valerii/hub/backend/internal/dev"

func wireDev(d *serverDeps) {
	devHandler := devmod.NewHandler(d.db, d.authSvc, d.mailCalendarDemoSeeder)
	devHandler.RegisterRoutes(d.api, d.cfg.Server.DevMode)
}
