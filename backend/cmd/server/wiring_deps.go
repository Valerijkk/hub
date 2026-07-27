package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valerii/hub/backend/ent"
	aimod "github.com/valerii/hub/backend/internal/ai"
	apigatewaymod "github.com/valerii/hub/backend/internal/apigateway"
	auditmod "github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/internal/auth"
	backupmod "github.com/valerii/hub/backend/internal/backup"
	callsmod "github.com/valerii/hub/backend/internal/calls"
	crmmod "github.com/valerii/hub/backend/internal/crm"
	dashboardmod "github.com/valerii/hub/backend/internal/dashboard"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	notifmod "github.com/valerii/hub/backend/internal/notification"
	"github.com/valerii/hub/backend/internal/platform/config"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/platform/outbox"
	projectmod "github.com/valerii/hub/backend/internal/project"
	"github.com/valerii/hub/backend/internal/rbac"
	searchmod "github.com/valerii/hub/backend/internal/search"
	sprintmod "github.com/valerii/hub/backend/internal/sprint"
	storagemod "github.com/valerii/hub/backend/internal/storage"
	tmsmod "github.com/valerii/hub/backend/internal/tms"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
	wsmod "github.com/valerii/hub/backend/internal/ws"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type serverDeps struct {
	cfg *config.Config
	ctx context.Context
	rootCtx    context.Context
	rootCancel context.CancelFunc

	integrationCipher *cryptopkg.Cipher
	traceShutdown     func(context.Context) error

	db    *ent.Client
	rawDB *sql.DB
	rdb   *redis.Client
	nc    *natsclient.Client

	outboxWorker *outbox.Worker
	auditRepo    auditmod.Repository
	auditSvc     *auditmod.AuditService

	app *fiber.App
	api fiber.Router

	authSvc     *auth.AuthService
	requireAuth fiber.Handler

	apigatewayRepo  *apigatewaymod.EntRepository
	apiKeySvc       *apigatewaymod.APIKeyService
	rbacCache       *rbac.PermissionCache
	outboxPublisher *outbox.Publisher
	projectRepo     *projectmod.EntRepository
	projectSvc      *projectmod.ProjectService
	rbacRepo        *rbac.EntRepository
	rbacSvc         *rbac.RBACService
	storageSvc      *storagemod.StorageService
	sprintSvc       *sprintmod.SprintService
	trackerRepo     *trackermod.EntRepository
	trackerSvc      *trackermod.TrackerService
	automationSvc   *trackermod.AutomationService
	mongoDB         *mongo.Database

	messengerEntRepo *messengermod.EntRepository
	messengerSvc     *messengermod.MessengerService
	tmsRepo          tmsmod.Repository
	tmsSvc           *tmsmod.TMSService
	searchSvc        *searchmod.SearchService

	notifSvc           *notifmod.NotificationService
	automationConsumer *trackermod.AutomationConsumer
	auditConsumer      *auditmod.AuditConsumer

	aiSvc     *aimod.Service
	aiHandler *aimod.Handler
	aiBot     *aimod.MessengerBot
	botUserID uuid.UUID

	dashSvc *dashboardmod.DashboardService
	crmSvc  *crmmod.Service

	callsRepo        *callsmod.EntRepository
	callsSvc         *callsmod.CallService
	scheduledCallSvc *callsmod.ScheduledCallService
	presenceSvc      *messengermod.PresenceService

	autoChannelSvc    *messengermod.AutoChannelService
	webhookDispatcher *apigatewaymod.WebhookDispatcher
	backupSvc         *backupmod.BackupService
	maintGate         *backupmod.MaintenanceGate

	retentionS3Client *s3.Client
	retentionS3Bucket string
	backupS3Client    *s3.Client
	backupS3Bucket    string

	wsHub    *wsmod.Hub
	wsBridge *wsmod.Bridge

	githubSvc *githubmod.GitHubService
	gitlabSvc *gitlabmod.GitLabService

	cleanups []func()
}

func (d *serverDeps) addCleanup(fn func()) {
	d.cleanups = append(d.cleanups, fn)
}

func (d *serverDeps) runCleanups() {
	for i := len(d.cleanups) - 1; i >= 0; i-- {
		d.cleanups[i]()
	}
}

func (d *serverDeps) fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}
