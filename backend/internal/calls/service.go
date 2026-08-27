// Package calls реализует голосовые созвоны через LiveKit: создание, присоединение, запись, планирование.
package calls

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// inviteCodeAlphabet — base62; 12 символов ≈ 71 bit (A2W3-5-003; MaxLen(12) без Ent-миграции).
const inviteCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// inviteCodeLength — длина invite_code (совпадает с Call.invite_code MaxLen).
const inviteCodeLength = 12

// generateInviteCode — crypto/rand base62, не усечённый UUID-hex (~48 bit).
func generateInviteCode() (string, error) {
	b := make([]byte, inviteCodeLength)
	max := big.NewInt(int64(len(inviteCodeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("генерация invite_code: %w", err)
		}
		b[i] = inviteCodeAlphabet[n.Int64()]
	}
	return string(b), nil
}

// ErrCallForbidden — HI-03: 403 для CreateCall/JoinCall/EndCall / A2-002 без утечки причины.
var ErrCallForbidden = errors.New("forbidden")

// ErrCallNotFound — звонок отсутствует (RaiseHand и др. без утечки деталей).
var ErrCallNotFound = errors.New("call not found")

// ErrInvalidInput — клиентский ввод вне контракта (MaxLen, nil project_id) → 400.
var ErrInvalidInput = errors.New("invalid input")

// defaultMaxCallParticipants — потолок участников комнаты LiveKit.
const defaultMaxCallParticipants = 50

// maxCallTitleRunes — Call.title MaxLen(255) в Ent; rune-cap до записи (A3W3-A5-004).
const maxCallTitleRunes = 255

// maxGuestNameRunes — разумный cap имени гостя до JWT (A3W3-A5-004).
const maxGuestNameRunes = 100

// CallService — тонкий orchestrator модуля звонков (SC-14-2).
// Делегирует конференции → ConferenceProvider, NATS → CallEventPublisher,
// чат → CallChatService, БД → EntRepository.
type CallService struct {
	repo     *EntRepository
	provider ConferenceProvider
	pub      *CallEventPublisher
	chat     *CallChatService

	// createCall — hubtest override CreateCall DB (A3W3-A5-001 mutation).
	createCall func(ctx context.Context, roomName string, channelID, createdBy uuid.UUID) (*ent.Call, error)
	// createInstantCall — hubtest override CreateInstantCall DB (A3W3-A5-001 mutation).
	createInstantCall func(ctx context.Context, roomName string, createdBy uuid.UUID, inviteCode, title string) (*ent.Call, error)
}

// NewCallService preserves the LiveKit constructor boundary while server wiring migrates to the provider factory.
func NewCallService(cfg config.LiveKitConfig, repo *EntRepository, nc *natsclient.Client, chatRepo *ChatRepository) *CallService {
	return NewCallServiceWithProvider(NewLiveKitConferenceProvider(cfg), repo, nc, chatRepo)
}

// NewCallServiceWithProvider constructs the calls orchestrator with the supplied conference provider.
func NewCallServiceWithProvider(provider ConferenceProvider, repo *EntRepository, nc *natsclient.Client, chatRepo *ChatRepository) *CallService {
	pub := NewCallEventPublisher(nc)
	return &CallService{
		repo:     repo,
		provider: provider,
		pub:      pub,
		chat:     NewCallChatService(repo, chatRepo, pub),
	}
}

// GenerateJoinCredential returns a provider-neutral credential for joining a conference room.
func (s *CallService) GenerateJoinCredential(roomName, userID, userName string) (JoinCredential, error) {
	credential, err := s.provider.IssueJoinCredential(roomName, userID, userName)
	if err != nil {
		return JoinCredential{}, err
	}
	return credential, nil
}

// GenerateToken returns a credential token for joining a conference room.
func (s *CallService) GenerateToken(roomName, userID, userName string) (string, error) {
	credential, err := s.GenerateJoinCredential(roomName, userID, userName)
	if err != nil {
		return "", err
	}
	return credential.Token, nil
}

// CreateCall создаёт звонок в канале (room + БД + участник + событие старта).
// A2-002: IsChannelMember before provisionRoom; else ErrCallForbidden.
func (s *CallService) CreateCall(ctx context.Context, channelID, userID uuid.UUID, userName string) (*CallResponse, error) {
	isMember, mErr := s.repo.IsChannelMember(ctx, channelID, userID)
	if mErr != nil {
		return nil, fmt.Errorf("channel membership check: %w", mErr)
	}
	if !isMember {
		return nil, ErrCallForbidden
	}
	roomName, err := provisionRoom(ctx, s.provider)
	if err != nil {
		return nil, err
	}
	callEnt, err := s.persistChannelCall(ctx, roomName, channelID, userID)
	if err != nil {
		// A3W3-A5-001: DB-fail после provisionRoom → DeleteRoom (как invite-fail / rollbackCreate).
		s.deleteRoomBestEffort(ctx, roomName, "CreateCall: не удалось удалить orphan-комнату после ошибки БД")
		return nil, fmt.Errorf("не удалось сохранить звонок: %w", err)
	}
	resp, err := finalizeCreate(ctx, s.provider, s.repo, callEnt, userID, userName)
	if err != nil {
		return nil, err
	}
	if channelID != uuid.Nil {
		s.pub.Started(callEnt.ChannelID, callEnt.ID, roomName, userID)
	}
	return resp, nil
}

// CreateInstantCall создаёт быстрый звонок без канала с invite-кодом.
func (s *CallService) CreateInstantCall(ctx context.Context, userID uuid.UUID, userName, title string) (*CallResponse, error) {
	if title == "" {
		title = userName + " — быстрый звонок"
	}
	// A3W3-A5-004: rune-cap до Ent MaxLen(255) → ErrInvalidInput/400, не 500.
	if utf8.RuneCountInString(title) > maxCallTitleRunes {
		return nil, ErrInvalidInput
	}
	roomName, err := provisionRoom(ctx, s.provider)
	if err != nil {
		return nil, err
	}
	// A2W3-5-003: crypto/rand base62×12 (~71 bit). Полный UUID — после MaxLen-миграции.
	inviteCode, err := generateInviteCode()
	if err != nil {
		s.deleteRoomBestEffort(ctx, roomName, "CreateInstantCall: не удалось удалить комнату после ошибки invite")
		return nil, err
	}
	callEnt, err := s.persistInstantCall(ctx, roomName, userID, inviteCode, title)
	if err != nil {
		// A3W3-A5-001: DB-fail после provisionRoom → DeleteRoom.
		s.deleteRoomBestEffort(ctx, roomName, "CreateInstantCall: не удалось удалить orphan-комнату после ошибки БД")
		return nil, fmt.Errorf("не удалось сохранить звонок: %w", err)
	}
	resp, err := finalizeCreate(ctx, s.provider, s.repo, callEnt, userID, userName)
	if err != nil {
		return nil, err
	}
	resp.InviteCode = inviteCode
	return resp, nil
}

// persistChannelCall пишет канальный звонок (hubtest может подменить createCall).
func (s *CallService) persistChannelCall(ctx context.Context, roomName string, channelID, createdBy uuid.UUID) (*ent.Call, error) {
	if s.createCall != nil {
		return s.createCall(ctx, roomName, channelID, createdBy)
	}
	return s.repo.CreateCall(ctx, roomName, channelID, createdBy)
}

// persistInstantCall пишет instant-звонок (hubtest может подменить createInstantCall).
func (s *CallService) persistInstantCall(ctx context.Context, roomName string, createdBy uuid.UUID, inviteCode, title string) (*ent.Call, error) {
	if s.createInstantCall != nil {
		return s.createInstantCall(ctx, roomName, createdBy, inviteCode, title)
	}
	return s.repo.CreateInstantCall(ctx, roomName, createdBy, inviteCode, title)
}

// deleteRoomBestEffort — best-effort DeleteRoom после провала post-provision (A3W3-A5-001).
func (s *CallService) deleteRoomBestEffort(ctx context.Context, roomName, logMsg string) {
	if delErr := s.provider.Delete(ctx, roomName); delErr != nil {
		slog.Error(logMsg, "room", roomName, "error", delErr)
	}
}

// JoinByInviteCode присоединяет пользователя к звонку по invite-коду.
func (s *CallService) JoinByInviteCode(ctx context.Context, code string, userID uuid.UUID, userName string) (*CallResponse, error) {
	callEnt, err := s.repo.GetCallByInviteCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}
	if err := s.repo.AddParticipant(ctx, callEnt.ID, userID); err != nil {
		return nil, fmt.Errorf("не удалось добавить участника: %w", err)
	}
	token, err := s.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}
	resp := toCallResponse(callEnt, token)
	resp.InviteCode = code
	return resp, nil
}

// GuestJoin — гостевой вход без аккаунта (W3-A5-007: слой Service, не handler→repo).
// A3W3-A5-004: guest_name rune-cap до JWT.
func (s *CallService) GuestJoin(ctx context.Context, inviteCode, guestName string) (*GuestJoinResponse, error) {
	if utf8.RuneCountInString(guestName) > maxGuestNameRunes {
		return nil, ErrInvalidInput
	}
	callEnt, err := s.repo.GetCallByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, ErrCallNotFound
	}
	if callEnt.Status != "active" {
		return nil, ErrCallNotFound
	}
	credential, err := s.provider.IssueGuestCredential(callEnt.RoomName, guestName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать гостевой токен: %w", err)
	}
	return &GuestJoinResponse{
		ID:         callEnt.ID,
		RoomName:   callEnt.RoomName,
		Token:      credential.Join.Token,
		LivekitURL: credential.Join.ServiceURL,
	}, nil
}

// JoinCall присоединяет к существующему звонку. HI-03: канальные — членство в канале;
// instant — только создатель (гости приходят через invite-код).
func (s *CallService) JoinCall(ctx context.Context, callID, userID uuid.UUID, userName string) (*CallResponse, error) {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}
	if callEnt.ChannelID != nil {
		isMember, mErr := s.repo.IsChannelMember(ctx, *callEnt.ChannelID, userID)
		if mErr != nil {
			return nil, fmt.Errorf("ошибка проверки членства в канале: %w", mErr)
		}
		if !isMember {
			return nil, ErrCallForbidden
		}
	} else if callEnt.CreatedBy != userID {
		return nil, ErrCallForbidden
	}
	if err := s.repo.AddParticipant(ctx, callID, userID); err != nil {
		return nil, fmt.Errorf("не удалось добавить участника: %w", err)
	}
	token, err := s.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}
	s.pub.Joined(callEnt.ChannelID, callID, userID, userName)
	return toCallResponse(callEnt, token), nil
}

// LeaveCall обрабатывает выход пользователя из звонка.
func (s *CallService) LeaveCall(ctx context.Context, callID, userID uuid.UUID) error {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("звонок не найден: %w", err)
	}
	if err := s.repo.RemoveParticipant(ctx, callID, userID, time.Now()); err != nil {
		return fmt.Errorf("не удалось удалить участника: %w", err)
	}
	s.pub.Left(callEnt.ChannelID, callID, userID)
	return nil
}

// EndCall завершает звонок. HI-03: только создатель или владелец/админ канала.
func (s *CallService) EndCall(ctx context.Context, callID, userID uuid.UUID) error {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.CreatedBy != userID {
		if callEnt.ChannelID == nil {
			return ErrCallForbidden
		}
		role, roleErr := s.repo.GetChannelMemberRole(ctx, *callEnt.ChannelID, userID)
		if roleErr != nil {
			return fmt.Errorf("ошибка проверки роли в канале: %w", roleErr)
		}
		if role != "owner" && role != "admin" {
			return ErrCallForbidden
		}
	}
	if delErr := s.provider.Delete(ctx, callEnt.RoomName); delErr != nil {
		slog.Error("не удалось удалить комнату LiveKit", "room", callEnt.RoomName, "error", delErr)
	}
	if err := s.repo.UpdateCallEnded(ctx, callID, time.Now(), callEnt.ParticipantCount); err != nil {
		return fmt.Errorf("не удалось обновить статус звонка: %w", err)
	}
	s.pub.Ended(callEnt.ChannelID, callID, callEnt.ParticipantCount)
	return nil
}

// GetActiveCallByChannel — активный звонок в канале или nil.
// A2-008: только член канала видит метаданные активного звонка.
func (s *CallService) GetActiveCallByChannel(ctx context.Context, channelID, userID uuid.UUID) (*CallResponse, error) {
	isMember, mErr := s.repo.IsChannelMember(ctx, channelID, userID)
	if mErr != nil {
		return nil, fmt.Errorf("проверка членства в канале: %w", mErr)
	}
	if !isMember {
		return nil, ErrCallForbidden
	}

	callEnt, err := s.repo.GetActiveCallByChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска активного звонка: %w", err)
	}
	if callEnt == nil {
		return nil, nil
	}
	return toCallResponse(callEnt, ""), nil
}

// RaiseHand — A2-006: поднять/опустить руку только для активного участника звонка.
// Раньше handler брал call по ID и звал LiveKit без authz — любой user с callId
// мог дергать UpdateParticipantMetadata для своего identity.
func (s *CallService) RaiseHand(ctx context.Context, callID, userID uuid.UUID, raised bool) error {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCallNotFound
		}
		return fmt.Errorf("get call: %w", err)
	}
	ok, pErr := s.repo.IsActiveParticipant(ctx, callID, userID)
	if pErr != nil {
		return fmt.Errorf("check participant: %w", pErr)
	}
	if !ok {
		return ErrCallForbidden
	}
	return s.SetRaiseHand(ctx, callEnt.RoomName, userID.String(), raised)
}

// SetRaiseHand обновляет provider-specific состояние участника ("поднять руку").
func (s *CallService) SetRaiseHand(ctx context.Context, roomName, identity string, raised bool) error {
	meta := fmt.Sprintf(`{"hand_raised":%v,"raised_at":"%s"}`, raised, time.Now().Format(time.RFC3339))
	return s.provider.UpdateParticipantState(ctx, roomName, identity, meta)
}

// OnRoomStarted — webhook room_started от LiveKit.
func (s *CallService) OnRoomStarted(ctx context.Context, roomName string) error {
	return s.repo.UpdateCallStarted(ctx, roomName, time.Now())
}

// OnRoomFinished — webhook room_finished: завершает звонок и чистит чат (CALL-06).
func (s *CallService) OnRoomFinished(ctx context.Context, roomName string) error {
	callEnt, err := s.repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return fmt.Errorf("звонок не найден по имени комнаты: %w", err)
	}
	if err := s.repo.UpdateCallEnded(ctx, callEnt.ID, time.Now(), callEnt.ParticipantCount); err != nil {
		return err
	}
	s.chat.DeleteCallChat(ctx, callEnt.ID.String())
	s.pub.Ended(callEnt.ChannelID, callEnt.ID, callEnt.ParticipantCount)
	return nil
}

// SendChatMessage делегирует в CallChatService.
func (s *CallService) SendChatMessage(ctx context.Context, callID string, userID, userName, text string, ephemeral bool) (*CallChatMessage, error) {
	return s.chat.SendMessage(ctx, callID, userID, userName, text, ephemeral)
}

// ListChatMessages делегирует в CallChatService (A2-003: userID для authz).
func (s *CallService) ListChatMessages(ctx context.Context, callID, userID string, limit int) ([]CallChatMessage, error) {
	return s.chat.ListMessages(ctx, callID, userID, limit)
}

// OnParticipantJoined — webhook participant_joined от LiveKit.
func (s *CallService) OnParticipantJoined(ctx context.Context, roomName, identity string) error {
	callEnt, userID, err := resolveParticipant(ctx, s.repo, roomName, identity)
	if err != nil || callEnt == nil {
		return err
	}
	return s.repo.AddParticipant(ctx, callEnt.ID, userID)
}

// OnParticipantLeft — webhook participant_left от LiveKit.
func (s *CallService) OnParticipantLeft(ctx context.Context, roomName, identity string) error {
	callEnt, userID, err := resolveParticipant(ctx, s.repo, roomName, identity)
	if err != nil || callEnt == nil {
		return err
	}
	return s.repo.RemoveParticipant(ctx, callEnt.ID, userID, time.Now())
}
