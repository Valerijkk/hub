package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	accountingmod "github.com/valerii/hub/backend/internal/accounting"
	basesmod "github.com/valerii/hub/backend/internal/bases"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	crmmod "github.com/valerii/hub/backend/internal/crm"
	dashboardmod "github.com/valerii/hub/backend/internal/dashboard"
	hrmod "github.com/valerii/hub/backend/internal/hr"
	intranetmod "github.com/valerii/hub/backend/internal/intranet"
	marketplacemod "github.com/valerii/hub/backend/internal/marketplace"
	onecmod "github.com/valerii/hub/backend/internal/integrations/onec"
	tmsmod "github.com/valerii/hub/backend/internal/tms"
)

func wireModules(d *serverDeps) {
	api := d.api
	requireAuth := d.requireAuth
	rbacSvc := d.rbacSvc

	// Dashboard routes (дашборды: персональный, проектный, админский)
	dashSvc := dashboardmod.NewDashboardService(dashboardmod.NewEntRepository(d.db), d.rdb, d.mongoDB)
	dashSvc.SetMessengerUnreadSummaryProvider(d.messengerSvc)
	dashHandler := dashboardmod.NewHandler(dashSvc)
	dashHandler.RegisterRoutes(api, requireAuth, rbacSvc)
	d.dashSvc = dashSvc

	intranetRepo := intranetmod.NewEntRepository(d.db)
	intranetNotifier := &intranetNotifier{svc: d.notifSvc}
	intranetSvc := intranetmod.NewService(intranetRepo, intranetNotifier, intranetNotifier)
	intranetHandler := intranetmod.NewHandler(intranetSvc)
	intranetHandler.RegisterRoutes(api, requireAuth)

	calendarRepo := calendarmod.NewEntRepository(d.db)
	calendarSvc := calendarmod.NewService(calendarRepo)
	calendarHandler := calendarmod.NewHandler(calendarSvc)
	calendarHandler.RegisterRoutes(api, requireAuth)

	crmRepo := crmmod.NewEntRepository(d.db)
	crmSvc := crmmod.NewService(crmRepo)
	crmHandler := crmmod.NewHandler(crmSvc)
	crmHandler.RegisterRoutes(api, requireAuth)
	d.crmSvc = crmSvc

	hrRepo := hrmod.NewEntRepository(d.db)
	onecStore := onecmod.NewEntSettingsStore(d.db)
	onecSvc := onecmod.NewService(onecStore, onecmod.NewEntUserLookup(d.db), d.integrationCipher)
	onecAdminHandler := onecmod.NewAdminHandler(onecSvc, func(c fiber.Ctx, userID uuid.UUID) (bool, error) {
		return d.auditSvc.IsSystemAdmin(c.Context(), userID)
	})
	onecAdminHandler.RegisterRoutes(api, requireAuth)

	hrSvc := hrmod.NewService(hrRepo, &hrNotifierAdapter{svc: d.notifSvc}, onecSvc)
	hrHandler := hrmod.NewHandler(hrSvc)
	hrHandler.RegisterRoutes(api, requireAuth)

	mpRepo := marketplacemod.NewEntRepository(d.db)
	mpSvc := marketplacemod.NewService(
		mpRepo,
		&marketplaceNotifierAdapter{svc: d.notifSvc},
		&marketplaceChannelBroadcaster{messenger: d.messengerSvc, repo: d.messengerEntRepo},
	)
	mpHandler := marketplacemod.NewHandler(mpSvc)
	mpHandler.RegisterRoutes(api, requireAuth)

	basesRepo := basesmod.NewEntRepository(d.db)
	basesSvc := basesmod.NewService(basesRepo)
	basesHandler := basesmod.NewHandler(basesSvc)
	basesHandler.RegisterRoutes(api, requireAuth)

	acctRepo := accountingmod.NewEntRepository(d.db)
	acctSvc := accountingmod.NewService(acctRepo)
	crmSvc.SetDealInvoiceCreator(acctSvc)
	acctHandler := accountingmod.NewHandler(acctSvc)
	acctHandler.RegisterRoutes(api, requireAuth)

	// Автопровизия TMS: создаём папку "Общие тесты" при создании проекта
	tmsProv := tmsmod.NewTMSProjectProvisioner(d.tmsRepo)
	d.projectSvc.SetTMSProvisioner(tmsProv)

	// Автопровизия дашборда: применяем шаблон "dev" при создании проекта
	dashProv := dashboardmod.NewDashboardProjectProvisioner(dashSvc)
	d.projectSvc.SetDashboardProvisioner(dashProv)
}
