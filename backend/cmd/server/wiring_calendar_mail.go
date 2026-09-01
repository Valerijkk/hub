package main

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
	stalwartmod "github.com/valerii/hub/backend/internal/mail/stalwart"
)

var (
	_ calendarmod.EventClock         = serverUTCClock{}
	_ mailmod.MailboxRepositoryClock = serverUTCClock{}
	_ calendarmod.EventIDGenerator   = serverEventIDs{}
	_ calendarmod.EventSerializer    = serverITIPSerializer{}
)

type serverUTCClock struct{}

func (serverUTCClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

type serverEventIDs struct{}

func (serverEventIDs) NewEventID() uuid.UUID {
	return uuid.New()
}

type serverITIPSerializer struct{}

func (serverITIPSerializer) Serialize(message calendarmod.ITIPMessage) ([]byte, error) {
	return calendarmod.SerializeITIP(message)
}

func wireCalendarMail(d *serverDeps) {
	calendarRepo := calendarmod.NewEntRepository(d.db)
	calendarSvc := calendarmod.NewService(calendarRepo)
	calendarHandler := calendarmod.NewHandler(calendarSvc)
	calendarHandler.RegisterRoutes(d.api, d.requireAuth)

	eventRepo := calendarmod.NewEventEntRepository(d.db)
	eventSvc := calendarmod.NewEventService(eventRepo, serverUTCClock{}, serverEventIDs{}, serverITIPSerializer{})
	eventHandler := calendarmod.NewEventHandler(eventSvc)
	eventHandler.RegisterRoutes(d.api, d.requireAuth)
	meetingSvc := calendarmod.NewMeetingService(eventRepo, serverUTCClock{}, serverEventIDs{})
	meetingHandler := calendarmod.NewMeetingHandler(meetingSvc)
	meetingHandler.RegisterRoutes(d.api, d.requireAuth)

	// Legacy ScheduledCall rows remain the bounded availability source until the
	// CalendarEvent-to-call bridge has completed its production cutover.
	availabilityRepo := calendarmod.NewEntAvailabilityRepository(d.db, true)
	availabilitySvc := calendarmod.NewAvailabilityService(availabilityRepo)
	availabilityHandler := calendarmod.NewAvailabilityHandler(availabilitySvc)
	availabilityHandler.RegisterRoutes(d.api, d.requireAuth)

	mailboxRepo, err := mailmod.NewMailboxEntRepository(d.db, serverUTCClock{})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailDomainControl, err := mailmod.NewMailDomainControl(mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailDomainControlHandler, err := mailmod.NewMailDomainControlHandler(mailDomainControl, func(ctx fiber.Ctx, userID uuid.UUID) (bool, error) {
		return d.auditSvc.IsSystemAdmin(ctx.Context(), userID)
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailDomainControlHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	eventReadSvc := calendarmod.NewEventReadService(eventRepo, mailboxRepo)
	eventReadHandler := calendarmod.NewEventReadHandler(eventReadSvc)
	eventReadHandler.RegisterRoutes(d.api, d.requireAuth)
	mailboxSvc, err := mailmod.NewMailboxService(mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxScheduleParticipantsService, err := mailmod.NewMailboxScheduleParticipantsService(mailboxSvc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxScheduleParticipantsHandler, err := mailmod.NewMailboxScheduleParticipantsHandler(mailboxScheduleParticipantsService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxScheduleParticipantsHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSetupStatusHandler, err := mailmod.NewMailboxSetupStatusHandler(mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxSetupStatusHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxHandler, err := mailmod.NewMailboxHandler(mailboxSvc)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxProjectionService, err := mailmod.NewMailboxFolderProjectionService(mailboxSvc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxProjectionHandler, err := mailmod.NewMailboxFolderProjectionHandler(mailboxProjectionService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxProjectionHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxIdentityService, err := mailmod.NewMailboxIdentityService(mailboxSvc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxIdentityHandler, err := mailmod.NewMailboxIdentityHandler(mailboxIdentityService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxIdentityHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if wireQAMailCalendarRuntime(d, mailboxSvc, mailboxRepo, eventRepo, eventSvc) {
		return
	}
	if wireDevMailCalendarRuntime(d, mailboxSvc, mailboxRepo, eventRepo, eventSvc, eventReadSvc) {
		return
	}

	mailboxEngineConfig, err := mailmod.ParseMailboxEngineConfig(os.LookupEnv)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if !mailboxEngineConfig.Enabled {
		return
	}
	if mailboxEngineConfig.Status != mailmod.MailboxEngineConfigStatusConfigured ||
		mailboxEngineConfig.MailAuthMode != mailmod.MailboxEngineMailAuthModeOAuthAuthorizationCodePKCE ||
		mailboxEngineConfig.DomainVerificationMode != mailmod.MailboxEngineDomainVerificationModePublicDNS {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}

	oauthClientID, clientIDPresent := os.LookupEnv("MAILBOX_ENGINE_OAUTH_CLIENT_ID")
	oauthRedirectURI, redirectURIPresent := os.LookupEnv("MAILBOX_ENGINE_OAUTH_REDIRECT_URI")
	stalwartPublicURL, publicURLPresent := os.LookupEnv("STALWART_PUBLIC_URL")
	if !clientIDPresent || oauthClientID == "" || !redirectURIPresent || oauthRedirectURI == "" ||
		!publicURLPresent || stalwartPublicURL == "" {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}
	if oauthClientSecret, clientSecretPresent := os.LookupEnv("MAILBOX_ENGINE_OAUTH_CLIENT_SECRET"); !clientSecretPresent || oauthClientSecret == "" {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}
	registryAPIKey, registryAPIKeyPresent := os.LookupEnv("MAILBOX_ENGINE_REGISTRY_API_KEY")
	if !registryAPIKeyPresent || registryAPIKey == "" {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}

	mailboxS3PresignOrigin, presignOriginPresent := os.LookupEnv("MAILBOX_S3_PRESIGN_ORIGIN")
	mailboxClamAVAddress, clamAVAddressPresent := os.LookupEnv("MAILBOX_CLAMAV_INSTREAM_ADDRESS")
	if !presignOriginPresent || mailboxS3PresignOrigin == "" || !clamAVAddressPresent || mailboxClamAVAddress == "" {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}
	mailboxObjectStore, err := mailmod.NewMailboxS3ObjectStore(d.rootCtx, d.cfg.S3, mailboxS3PresignOrigin)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAttachmentScanner, err := mailmod.NewClamAVINSTREAMScanner(mailboxClamAVAddress)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAttachmentStagingService, err := mailmod.NewMailboxAttachmentStagingService(
		mailboxSvc,
		mailboxRepo,
		mailboxRepo,
		mailboxRepo,
		mailboxObjectStore,
		mailmod.MailboxAttachmentStagingConfig{
			Bucket:        d.cfg.S3.Bucket,
			PresignOrigin: mailboxS3PresignOrigin,
		},
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAttachmentAbandonService, err := mailmod.NewMailboxAttachmentAbandonService(mailboxSvc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAttachmentStagingHandler, err := mailmod.NewMailboxAttachmentStagingHandler(mailboxAttachmentStagingService, mailboxAttachmentAbandonService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	mailboxHTTPClient, err := newMailboxHTTPClient(os.LookupEnv)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	d.addCleanup(mailboxHTTPClient.CloseIdleConnections)
	managementRegistryAuthorizer, err := mailmod.NewStalwartRegistryAPIKeyAuthorizer(mailboxEngineConfig.Origin, func(ctx context.Context) ([]byte, error) {
		if ctx == nil {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		if err := ctx.Err(); err != nil {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		registryAPIKey, registryAPIKeyPresent := os.LookupEnv("MAILBOX_ENGINE_REGISTRY_API_KEY")
		if !registryAPIKeyPresent || registryAPIKey == "" {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		return append([]byte(nil), registryAPIKey...), nil
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	managementRegistryClient, err := mailmod.NewStalwartRegistryClient(mailboxEngineConfig.Origin, mailboxHTTPClient, managementRegistryAuthorizer)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	registryDomainController, err := stalwartmod.NewRegistryDomainController(managementRegistryClient)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	registryAccountClient, err := stalwartmod.NewRegistryAccountClient(managementRegistryClient)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	provisioningEngine, err := mailmod.NewMailboxRegistryProvisioningEngine(registryDomainController, registryAccountClient)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	publicDNSLookup, err := newMailboxPublicDNSLookup()
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	publicDNSVerifier, err := mailmod.NewPublicDNSVerifier(publicDNSLookup)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	publicDNSDomainProvisioner, err := mailmod.NewPublicDNSDomainProvisioner(provisioningEngine, publicDNSVerifier)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxDomainProvisioningWorker, err := mailmod.NewMailboxDomainProvisioningWorker(
		mailboxRepo,
		publicDNSDomainProvisioner,
		mailmod.NewMailboxDomainProvisioningWorkerTimer(),
		mailmod.DefaultMailboxDomainProvisioningWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAccountProvisioningWorker, err := mailmod.NewMailboxAccountProvisioningWorker(
		mailboxRepo,
		provisioningEngine,
		mailmod.NewMailboxAccountProvisioningWorkerTimer(),
		mailmod.DefaultMailboxAccountProvisioningWorkerOptions(),
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	managementPasswordClient, err := stalwartmod.NewRegistryPasswordClient(managementRegistryClient)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	managementPasswordService, err := mailmod.NewMailboxPasswordEnrollmentService(mailboxRepo, managementPasswordClient)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	oauthClientSecretProvider, err := mailmod.NewStalwartOAuthClientSecretProvider(oauthClientID, func(_ context.Context, _ string) ([]byte, error) {
		clientSecret, present := os.LookupEnv("MAILBOX_ENGINE_OAUTH_CLIENT_SECRET")
		if !present || clientSecret == "" {
			return nil, mailmod.ErrMailboxEngineConfigInvalid
		}
		return []byte(clientSecret), nil
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	oauthClientSecret, err := oauthClientSecretProvider.OAuthClientSecret(d.rootCtx, oauthClientID)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	clear(oauthClientSecret)

	oauthClient, err := stalwartmod.NewOAuthClient(stalwartmod.OAuthConfig{
		IssuerOrigin:         stalwartPublicURL,
		ResourceOrigin:       stalwartPublicURL,
		TransportOrigin:      mailboxEngineConfig.Origin,
		ClientID:             oauthClientID,
		ClientSecretProvider: oauthClientSecretProvider,
		RedirectURI:          oauthRedirectURI,
		HTTPClient:           mailboxHTTPClient,
		Random:               nil,
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	oauthStates, err := mailmod.NewMailboxOAuthRedisStateStore(d.rdb)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxRegistryAuthorizer := mailmod.NewStalwartMailboxRegistryAuthorizer()
	oauthAccountProof, err := mailmod.NewStalwartMailboxOAuthAccountProof(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxOAuthService, err := mailmod.NewMailboxOAuthService(mailboxRepo, oauthStates, oauthClient, oauthAccountProof, d.integrationCipher, serverUTCClock{}, stalwartPublicURL)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAuthorizers, err := mailmod.NewStalwartMailboxAuthorizerSource(stalwartPublicURL, mailboxRepo, mailboxOAuthService, d.integrationCipher, serverUTCClock{})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxClientFactory, err := mailmod.NewStalwartMailboxClientFactory(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer, mailboxAuthorizers)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	bootstrapOAuthStates, err := mailmod.NewMailboxBootstrapOAuthRedisStateStore(d.rdb)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	bootstrapIdentityProof, err := mailmod.NewStalwartMailboxBootstrapIdentityProof(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxBootstrapOAuthService, err := mailmod.NewMailboxBootstrapOAuthService(mailboxRepo, bootstrapOAuthStates, oauthClient, bootstrapIdentityProof, d.integrationCipher, serverUTCClock{}, stalwartPublicURL)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSetupHandler, err := mailmod.NewMailboxSetupHandler(managementPasswordService, mailboxBootstrapOAuthService, mailboxSvc, mailmod.MailboxSetupHandlerConfig{
		BindingCookieName:      "hub_mail_setup_oauth_binding",
		BindingCookieSecure:    !d.cfg.Server.DevMode,
		BindingCookieMaxAgeSec: 600,
		MaximumJSONBytes:       9216,
		MaximumQueryBytes:      2048,
		MaximumPasswordBytes:   256,
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxSetupHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxEngine, err := mailmod.NewStalwartEngine(mailboxClientFactory)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	calendarMailDeliveryDraftCreator, err := mailmod.NewCalendarMailDeliveryDraftCreator(mailboxSvc, mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSubmissionDispatcher, err := mailmod.NewMailboxSubmissionDispatcher(mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSubmissionReconciler, err := mailmod.NewMailboxSubmissionReconciler(mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSubmissionFinalizer, err := mailmod.NewMailboxSubmissionFinalizer(mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxDraftMaterializer, err := mailmod.NewMailboxEntDraftAttachmentMaterializer(mailboxRepo, mailboxObjectStore)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxDraftService, err := mailmod.NewMailboxDraftService(
		mailboxSvc,
		mailboxRepo,
		mailboxEngine,
		mailboxRepo,
		mailboxRepo,
		mailboxDraftMaterializer,
		mailboxRepo,
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	wireMailboxEngineRuntime(d, mailboxEngineRuntimeInputs{
		mailboxService:                  mailboxSvc,
		mailboxRepository:               mailboxRepo,
		eventRepository:                 eventRepo,
		eventService:                    eventSvc,
		syncEngine:                      mailboxEngine,
		messageReadEngine:               mailboxEngine,
		messageMutationEngine:           mailboxEngine,
		calendarPartEngine:              mailboxEngine,
		calendarMailDeliveryDraftCreator: calendarMailDeliveryDraftCreator,
		draftService:                    mailboxDraftService,
		submissionDispatcher:            mailboxSubmissionDispatcher,
		submissionReconciler:            mailboxSubmissionReconciler,
		submissionFinalizer:             mailboxSubmissionFinalizer,
	})

	mailboxOAuthHandler, err := mailmod.NewMailboxOAuthHandler(mailboxOAuthService, mailboxSvc, mailboxRepo, mailmod.MailboxOAuthHandlerConfig{
		BindingCookieName:      "hub_mail_oauth_binding",
		BindingCookieSecure:    !d.cfg.Server.DevMode,
		BindingCookieMaxAgeSec: 600,
		MaximumJSONBytes:       9216,
		MaximumQueryBytes:      2048,
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxOAuthHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxAttachmentStagingHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxObjectDeletionWorker, err := mailmod.NewMailboxObjectDeletionWorker(
		mailboxRepo,
		mailboxObjectStore,
		mailmod.NewMailboxObjectDeletionWorkerTimer(),
		mailmod.MailboxObjectDeletionWorkerConfig{
			Enabled:       true,
			ScanInterval:  5 * time.Second,
			StepTimeout:   30 * time.Second,
			LeaseDuration: 60 * time.Second,
			RetryMin:      5 * time.Second,
			RetryMax:      5 * time.Minute,
			BatchSize:     25,
		},
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAttachmentScanWorker, err := mailmod.NewMailboxAttachmentScanWorker(
		mailboxRepo,
		mailboxObjectStore,
		mailboxAttachmentScanner,
		mailmod.NewMailboxAttachmentScanWorkerTimer(),
		mailmod.MailboxAttachmentScanWorkerConfig{
			Enabled:       true,
			ScanInterval:  5 * time.Second,
			StepTimeout:   55 * time.Second,
			LeaseDuration: 60 * time.Second,
			RetryMin:      5 * time.Second,
			RetryMax:      5 * time.Minute,
			BatchSize:     8,
			Concurrency:   1,
		},
	)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	go func() {
		if err := mailboxObjectDeletionWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
	go func() {
		if err := mailboxAttachmentScanWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
	mailboxDomainProvisioningWorkerCtx, mailboxDomainProvisioningWorkerCancel := context.WithCancel(d.rootCtx)
	mailboxDomainProvisioningWorkerDone := make(chan struct{})
	go func() {
		defer close(mailboxDomainProvisioningWorkerDone)
		if err := mailboxDomainProvisioningWorker.Run(mailboxDomainProvisioningWorkerCtx); err != nil && mailboxDomainProvisioningWorkerCtx.Err() == nil {
			slog.Error("background worker stopped")
		}
	}()
	var mailboxDomainProvisioningWorkerCleanup sync.Once
	d.addCleanup(func() {
		mailboxDomainProvisioningWorkerCleanup.Do(func() {
			mailboxDomainProvisioningWorkerCancel()
			<-mailboxDomainProvisioningWorkerDone
		})
	})
	mailboxAccountProvisioningWorkerCtx, mailboxAccountProvisioningWorkerCancel := context.WithCancel(d.rootCtx)
	mailboxAccountProvisioningWorkerDone := make(chan struct{})
	go func() {
		defer close(mailboxAccountProvisioningWorkerDone)
		if err := mailboxAccountProvisioningWorker.Run(mailboxAccountProvisioningWorkerCtx); err != nil && mailboxAccountProvisioningWorkerCtx.Err() == nil {
			slog.Error("background worker stopped")
		}
	}()
	var mailboxAccountProvisioningWorkerCleanup sync.Once
	d.addCleanup(func() {
		mailboxAccountProvisioningWorkerCleanup.Do(func() {
			mailboxAccountProvisioningWorkerCancel()
			<-mailboxAccountProvisioningWorkerDone
		})
	})
}
