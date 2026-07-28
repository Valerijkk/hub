//go:build hubtest

package calls

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/rbac"
)

// ExportToRecordingResponse конвертирует ent.Recording в RecordingResponse для hubtest.
func ExportToRecordingResponse(r *ent.Recording) *RecordingResponse {
	return toRecordingResponse(r)
}

// ExportToScheduledCallResponse конвертирует ent.ScheduledCall в ScheduledCallResponse для hubtest.
func ExportToScheduledCallResponse(sc *ent.ScheduledCall) *ScheduledCallResponse {
	return toScheduledCallResponse(sc)
}

// ExportToCallResponse конвертирует ent.Call + JWT в CallResponse для hubtest.
func ExportToCallResponse(c *ent.Call, token string) *CallResponse {
	return toCallResponse(c, token)
}

// ExportParseCallID извлекает callId из Fiber-контекста для hubtest.
func ExportParseCallID(c fiber.Ctx) (uuid.UUID, error) {
	return parseCallID(c)
}

// ExportVerifyWebhookSignature проверя подпись LiveKit webhook для hubtest.
func ExportVerifyWebhookSignature(h *Handler, authHeader string, body []byte) bool {
	return h.verifyWebhookSignature(authHeader, body)
}

// ExportFinalizeCreate выдаёт токен создателю и регистрирует участника для hubtest.
func ExportFinalizeCreate(ctx context.Context, room *CallRoomService, repo *EntRepository, callEnt *ent.Call, userID uuid.UUID, userName string) (*CallResponse, error) {
	return finalizeCreate(ctx, room, repo, callEnt, userID, userName)
}

// ExportResolveParticipant находит звонок и парсит identity webhook для hubtest.
func ExportResolveParticipant(ctx context.Context, repo *EntRepository, roomName, identity string) (*ent.Call, uuid.UUID, error) {
	return resolveParticipant(ctx, repo, roomName, identity)
}

// ExportDefaultMaxCallParticipants возвращает потолок участников комнаты для hubtest.
func ExportDefaultMaxCallParticipants() int {
	return defaultMaxCallParticipants
}

// ExportMaxActiveRecordingsPerCall возвращает лимит одновременных записей для hubtest.
func ExportMaxActiveRecordingsPerCall() int {
	return maxActiveRecordingsPerCall
}

// ExportNewCallServiceForTest создаёт CallService с заданными зависимостями для hubtest.
func ExportNewCallServiceForTest(repo *EntRepository, room *CallRoomService, pub *CallEventPublisher, chat *CallChatService) *CallService {
	return &CallService{repo: repo, room: room, pub: pub, chat: chat}
}

// ExportNewHandlerForTest создаёт Handler с полным набором зависимостей для hubtest.
func ExportNewHandlerForTest(service *CallService, recSvc *RecordingService, scheduledSvc *ScheduledCallService, cfg config.LiveKitConfig, rbacSvc *rbac.RBACService) *Handler {
	return &Handler{service: service, recSvc: recSvc, scheduledSvc: scheduledSvc, cfg: cfg, rbacSvc: rbacSvc}
}

// ExportNewHandlerWithCfg создаёт Handler только с LiveKit-конфигом для hubtest.
func ExportNewHandlerWithCfg(cfg config.LiveKitConfig) *Handler {
	return &Handler{cfg: cfg}
}

// ExportNewHandlerWithRecSvc создаёт Handler с RecordingService для hubtest.
func ExportNewHandlerWithRecSvc(recSvc *RecordingService) *Handler {
	return &Handler{recSvc: recSvc}
}

// ExportNewHandlerWithServiceAndScheduled создаёт Handler для scheduled-calls тестов hubtest.
func ExportNewHandlerWithServiceAndScheduled(service *CallService, scheduledSvc *ScheduledCallService, cfg config.LiveKitConfig) *Handler {
	return &Handler{service: service, scheduledSvc: scheduledSvc, cfg: cfg}
}

// ExportNewHandlerWithServiceAndRecSvc создаёт Handler для webhook/recording тестов hubtest.
func ExportNewHandlerWithServiceAndRecSvc(service *CallService, recSvc *RecordingService, cfg config.LiveKitConfig) *Handler {
	return &Handler{service: service, recSvc: recSvc, cfg: cfg}
}

// ExportNewHandlerWithScheduledSvc создаёт Handler только с ScheduledCallService для hubtest.
func ExportNewHandlerWithScheduledSvc(scheduledSvc *ScheduledCallService) *Handler {
	return &Handler{scheduledSvc: scheduledSvc}
}

// ExportNewHandlerWithRecSvcAndRBAC создаёт Handler с RBAC для recording-тестов hubtest.
func ExportNewHandlerWithRecSvcAndRBAC(recSvc *RecordingService, rbacSvc *rbac.RBACService) *Handler {
	return &Handler{recSvc: recSvc, rbacSvc: rbacSvc}
}

// ExportNewRecordingServiceForTest создаёт RecordingService без LiveKit egress для hubtest.
func ExportNewRecordingServiceForTest(repo *EntRepository, nc *natsclient.Client, s3Cfg config.S3Config) *RecordingService {
	return &RecordingService{repo: repo, nats: nc, s3Cfg: s3Cfg}
}

// ExportNewScheduledCallServiceForTest создаёт ScheduledCallService для hubtest.
func ExportNewScheduledCallServiceForTest(repo *EntRepository) *ScheduledCallService {
	return &ScheduledCallService{repo: repo}
}

// ExportCallServiceRepo возвращает EntRepository из CallService для hubtest.
func ExportCallServiceRepo(s *CallService) *EntRepository {
	return s.repo
}

// ExportCallServiceRoom возвращает CallRoomService из CallService для hubtest.
func ExportCallServiceRoom(s *CallService) *CallRoomService {
	return s.room
}

// ExportCallServicePub возвращает CallEventPublisher из CallService для hubtest.
func ExportCallServicePub(s *CallService) *CallEventPublisher {
	return s.pub
}

// ExportCallServiceChat возвращает CallChatService из CallService для hubtest.
func ExportCallServiceChat(s *CallService) *CallChatService {
	return s.chat
}

// ExportCallServiceSetRoom подменяет CallRoomService в CallService для hubtest.
func ExportCallServiceSetRoom(s *CallService, room *CallRoomService) {
	s.room = room
}

// ExportCallServiceSetChat подменяет CallChatService в CallService для hubtest.
func ExportCallServiceSetChat(s *CallService, chat *CallChatService) {
	s.chat = chat
}

// ExportHandlerService возвращает CallService из Handler для hubtest.
func ExportHandlerService(h *Handler) *CallService {
	return h.service
}

// ExportRecordingServiceRepo возвращает EntRepository из RecordingService для hubtest.
func ExportRecordingServiceRepo(s *RecordingService) *EntRepository {
	return s.repo
}

// ExportRecordingServicePresignClient возвращает S3 presign client для hubtest.
func ExportRecordingServicePresignClient(s *RecordingService) *s3.PresignClient {
	return s.presignClient
}

// ExportCallRoomServiceCfg возвращает LiveKit-конфиг из CallRoomService для hubtest.
func ExportCallRoomServiceCfg(room *CallRoomService) config.LiveKitConfig {
	return room.cfg
}

// ExportCallEventPublisherNats возвращает NATS client из CallEventPublisher для hubtest.
func ExportCallEventPublisherNats(p *CallEventPublisher) *natsclient.Client {
	return p.nats
}

// --- in-memory chat store for hubtest (migrated from chat_mem_test) ---

type hubtestMemChatStore struct {
	mu   sync.Mutex
	msgs []CallChatMessage
}

func (s *hubtestMemChatStore) ensureIndexes(_ context.Context) error { return nil }

func (s *hubtestMemChatStore) insert(_ context.Context, msg *CallChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgs = append(s.msgs, *msg)
	return nil
}

func (s *hubtestMemChatStore) listByCallID(_ context.Context, callID string, limit int) ([]CallChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var matched []CallChatMessage
	for _, m := range s.msgs {
		if m.CallID == callID {
			matched = append(matched, m)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	if len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nil
}

func (s *hubtestMemChatStore) deleteByCallID(_ context.Context, callID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.msgs[:0]
	for _, m := range s.msgs {
		if m.CallID != callID {
			kept = append(kept, m)
		}
	}
	s.msgs = kept
	return nil
}

type hubtestFailInsertChatStore struct{ hubtestMemChatStore }

func (s *hubtestFailInsertChatStore) insert(context.Context, *CallChatMessage) error {
	return errors.New("insert failed")
}

type hubtestFailDeleteChatStore struct{ hubtestMemChatStore }

func (s *hubtestFailDeleteChatStore) deleteByCallID(context.Context, string) error {
	return errors.New("delete failed")
}

// ExportNewMemChatRepository создаёт in-memory ChatRepository для hubtest.
func ExportNewMemChatRepository() *ChatRepository {
	return &ChatRepository{store: &hubtestMemChatStore{}}
}

// ExportNewChatRepositoryFailingInsert создаёт ChatRepository с ошибкой insert для hubtest.
func ExportNewChatRepositoryFailingInsert() *ChatRepository {
	return &ChatRepository{store: &hubtestFailInsertChatStore{}}
}

// ExportNewChatRepositoryFailingDelete создаёт ChatRepository с ошибкой delete для hubtest.
func ExportNewChatRepositoryFailingDelete() *ChatRepository {
	return &ChatRepository{store: &hubtestFailDeleteChatStore{}}
}

// ExportChatRepositoryInsertForTest напрямую вставляет сообщение в store для hubtest.
func ExportChatRepositoryInsertForTest(r *ChatRepository, ctx context.Context, msg *CallChatMessage) error {
	return r.store.insert(ctx, msg)
}

// ExportChannelIDStr возвращает строку channel ID для NATS subject hubtest.
func ExportChannelIDStr(id *uuid.UUID) string {
	return channelIDStr(id)
}

// ExportUserNameFromCtx извлекает userName из Fiber Locals для hubtest.
func ExportUserNameFromCtx(c fiber.Ctx) string {
	return userNameFromCtx(c)
}

// ExportGenerateGuestToken генерирует JWT для гостевого входа для hubtest.
func ExportGenerateGuestToken(h *Handler, roomName, identity, name string) (string, error) {
	return h.service.room.GenerateToken(roomName, identity, name)
}

type ExportMongoChatStore struct {
	inner *mongoChatStore
}

// ExportNewMongoChatStore создаёт mongo chat store для hubtest.
func ExportNewMongoChatStore(coll *mongo.Collection) *ExportMongoChatStore {
	return &ExportMongoChatStore{inner: newMongoChatStore(coll)}
}

// EnsureIndexes создаёт индексы MongoDB для hubtest.
func (s *ExportMongoChatStore) EnsureIndexes(ctx context.Context) error {
	return s.inner.ensureIndexes(ctx)
}

// Insert сохраняет сообщение для hubtest.
func (s *ExportMongoChatStore) Insert(ctx context.Context, msg *CallChatMessage) error {
	return s.inner.insert(ctx, msg)
}

// ListByCallID возвращает сообщения звонка для hubtest.
func (s *ExportMongoChatStore) ListByCallID(ctx context.Context, callID string, limit int) ([]CallChatMessage, error) {
	return s.inner.listByCallID(ctx, callID, limit)
}

// DeleteByCallID удаляет сообщения звонка для hubtest.
func (s *ExportMongoChatStore) DeleteByCallID(ctx context.Context, callID string) error {
	return s.inner.deleteByCallID(ctx, callID)
}
