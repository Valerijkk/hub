package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

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
	eventReadSvc := calendarmod.NewEventReadService(eventRepo)
	eventReadHandler := calendarmod.NewEventReadHandler(eventReadSvc)
	eventReadHandler.RegisterRoutes(d.api, d.requireAuth)
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
	mailboxSvc, err := mailmod.NewMailboxService(mailboxRepo)
	if err != nil {
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

	mailboxEngineConfig, err := mailmod.ParseMailboxEngineConfig(os.LookupEnv)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if !mailboxEngineConfig.Enabled {
		return
	}
	if mailboxEngineConfig.Status != mailmod.MailboxEngineConfigStatusConfigured ||
		mailboxEngineConfig.MailAuthMode != mailmod.MailboxEngineMailAuthModeOAuthAuthorizationCodePKCE {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}

	oauthClientID, clientIDPresent := os.LookupEnv("MAILBOX_ENGINE_OAUTH_CLIENT_ID")
	oauthRedirectURI, redirectURIPresent := os.LookupEnv("MAILBOX_ENGINE_OAUTH_REDIRECT_URI")
	stalwartPublicURL, publicURLPresent := os.LookupEnv("STALWART_PUBLIC_URL")
	if !clientIDPresent || oauthClientID == "" || !redirectURIPresent || oauthRedirectURI == "" ||
		!publicURLPresent || stalwartPublicURL != mailboxEngineConfig.Origin {
		d.fatal("failed to wire calendar and mail routes", mailmod.ErrMailboxEngineConfigInvalid)
	}
	if oauthClientSecret, clientSecretPresent := os.LookupEnv("MAILBOX_ENGINE_OAUTH_CLIENT_SECRET"); !clientSecretPresent || oauthClientSecret == "" {
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
		mailboxRepo,
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
	mailboxAttachmentAbandonService, err := mailmod.NewMailboxAttachmentAbandonService(mailboxRepo, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAttachmentStagingHandler, err := mailmod.NewMailboxAttachmentStagingHandler(mailboxAttachmentStagingService, mailboxAttachmentAbandonService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	mailboxHTTPClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			MaxIdleConns:          8,
			MaxIdleConnsPerHost:   4,
			MaxConnsPerHost:       8,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
		Timeout: 15 * time.Second,
	}
	d.addCleanup(mailboxHTTPClient.CloseIdleConnections)

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
		IssuerOrigin:         mailboxEngineConfig.Origin,
		ResourceOrigin:       mailboxEngineConfig.Origin,
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
	registryAuthorizer := mailmod.NewStalwartMailboxRegistryAuthorizer()
	oauthAccountProof, err := mailmod.NewStalwartMailboxOAuthAccountProof(mailboxEngineConfig.Origin, mailboxHTTPClient, registryAuthorizer)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxOAuthService, err := mailmod.NewMailboxOAuthService(mailboxRepo, oauthStates, oauthClient, oauthAccountProof, d.integrationCipher, serverUTCClock{}, mailboxEngineConfig.Origin)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxAuthorizers, err := mailmod.NewStalwartMailboxAuthorizerSource(mailboxEngineConfig.Origin, mailboxRepo, mailboxOAuthService, d.integrationCipher, serverUTCClock{})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxClientFactory, err := mailmod.NewStalwartMailboxClientFactory(mailboxEngineConfig.Origin, mailboxHTTPClient, registryAuthorizer, mailboxAuthorizers)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxEngine, err := mailmod.NewStalwartEngine(mailboxClientFactory)
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
	mailboxDraftHandler, err := mailmod.NewMailboxDraftCreateHandler(mailboxDraftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxDraftHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxReplyDraftService, err := mailmod.NewMailboxReplyDraftService(mailboxSvc, mailboxRepo, mailboxEngine, mailboxDraftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxReplyDraftHandler, err := mailmod.NewMailboxReplyDraftHandler(mailboxReplyDraftService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxReplyDraftHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

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
	mailboxMessageService, err := mailmod.NewMailboxMessageService(mailboxSvc, mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxMessageHandler, err := mailmod.NewMailboxMessageHandler(mailboxMessageService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxMessageHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxMessageMutationService, err := mailmod.NewMailboxMessageMutationService(mailboxSvc, mailboxRepo, mailboxEngine)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxMessageMutationHandler, err := mailmod.NewMailboxMessageMutationHandler(mailboxMessageMutationService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxMessageMutationHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxAttachmentStagingHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}

	mailboxSyncService, err := mailmod.NewMailboxSyncService(mailboxEngine, mailboxRepo, mailmod.MailboxSyncServiceConfig{
		Now:           serverUTCClock{}.Now,
		NewLeaseID:    uuid.New,
		NewEventID:    uuid.New,
		Jitter:        func(time.Duration) time.Duration { return 0 },
		LeaseDuration: 30 * time.Second,
		PollInterval:  5 * time.Second,
		RetryInterval: 15 * time.Second,
		PageSize:      100,
	})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSyncWorker, err := mailmod.NewMailboxSyncWorker(mailboxRepo, mailboxSyncService, mailmod.NewMailboxSyncWorkerTimer(), mailmod.MailboxSyncWorkerConfig{
		Enabled:      true,
		ScanInterval: 2 * time.Second,
		StepTimeout:  30 * time.Second,
		RetryMin:     5 * time.Second,
		RetryMax:     time.Minute,
		BatchSize:    25,
		Concurrency:  2,
	})
	if err != nil {
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
		if err := mailboxSyncWorker.Run(d.rootCtx); err != nil && d.rootCtx.Err() == nil {
			slog.Error("background worker stopped", "error", err)
		}
	}()
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
}
