package calls

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// authorizeCallRecordingAccess — A2-004: запись/остановка только для члена канала
// звонка (или создателя instant-звонка). Без этой проверки любой authenticated
// user мог стартовать egress / стопать чужой recording по UUID.
func (s *RecordingService) authorizeCallRecordingAccess(ctx context.Context, callEnt *ent.Call, userID uuid.UUID) error {
	if callEnt.ChannelID == nil {
		if callEnt.CreatedBy != userID {
			return ErrCallForbidden
		}
		return nil
	}
	isMember, err := s.repo.IsChannelMember(ctx, *callEnt.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("ошибка проверки членства в канале: %w", err)
	}
	if !isMember {
		return ErrCallForbidden
	}
	return nil
}

// maxActiveRecordingsPerCall определяет максимальное количество одновременных
// записей на один звонок. Ограничение защищает от перегрузки LiveKit Egress.
const maxActiveRecordingsPerCall = 2

// downloadURLTTL — время жизни presigned URL для скачивания записи.
const downloadURLTTL = 1 * time.Hour

// RecordingService реализует бизнес-логику записи звонков через LiveKit Egress.
// Управляет жизненным циклом записи: запуск Egress, остановка, обработка
// завершения и предоставление presigned URL для скачивания.
type RecordingService struct {
	egressClient  *lksdk.EgressClient
	s3Cfg         config.S3Config
	presignClient *s3.PresignClient
	repo          *EntRepository
	nats          *natsclient.Client
	// createRecording — опциональный override для hubtest (orphan egress compensation).
	createRecording func(ctx context.Context, egressID string, callID, projectID, userID uuid.UUID, filePath string) (*ent.Recording, error)
}

// NewRecordingService создаёт новый сервис записи звонков.
// Инициализирует EgressClient для управления записью через LiveKit
// и presign S3-клиент для генерации временных ссылок на скачивание.
// Если presign-клиент создать не удалось — логируем и продолжаем без него
// (GetRecordingDownloadURL вернёт ошибку, остальные операции продолжат работать).
func NewRecordingService(lkCfg config.LiveKitConfig, s3Cfg config.S3Config, repo *EntRepository, nc *natsclient.Client) *RecordingService {
	egressClient := lksdk.NewEgressClient(lkCfg.URL, lkCfg.APIKey, lkCfg.APISecret)

	var presignClient *s3.PresignClient
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(s3Cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s3Cfg.AccessKey, s3Cfg.SecretKey, ""),
		),
	)
	if err != nil {
		slog.Error("recording service: failed to load AWS config for presign", "error", err)
	} else {
		protocol := "http"
		if s3Cfg.UseSSL {
			protocol = "https"
		}
		endpoint := s3Cfg.Endpoint
		// Если эндпоинт не содержит схему, добавляем её — иначе SDK интерпретирует
		// его как путь и presigned URL будет некорректным.
		if endpoint != "" && len(endpoint) > 0 && endpoint[0] != 'h' {
			endpoint = fmt.Sprintf("%s://%s", protocol, endpoint)
		}
		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if endpoint != "" {
				o.BaseEndpoint = aws.String(endpoint)
			}
			o.UsePathStyle = true // Требуется для SeaweedFS
		})
		presignClient = s3.NewPresignClient(s3Client)
	}

	return &RecordingService{
		egressClient:  egressClient,
		s3Cfg:         s3Cfg,
		presignClient: presignClient,
		repo:          repo,
		nats:          nc,
	}
}

// StartRecording запускает запись звонка через LiveKit Egress.
// A2-004: только член канала звонка. Далее — лимит активных записей,
// RoomCompositeEgress с S3 output, метаданные в БД и NATS-событие.
func (s *RecordingService) StartRecording(ctx context.Context, callID, userID uuid.UUID) (*RecordingResponse, error) {
	// Получаем звонок для проверки и данных комнаты
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}

	// Получаем project_id через канал
	if callEnt.ChannelID == nil {
		return nil, fmt.Errorf("instant-звонки пока не поддерживают запись")
	}
	if err := s.authorizeCallRecordingAccess(ctx, callEnt, userID); err != nil {
		return nil, err
	}
	projectID, err := s.repo.GetProjectIDByChannel(ctx, *callEnt.ChannelID)
	if err != nil {
		// A3W3-A5-003: канал без project_id → NotFound/явная ошибка → 400, не FK 500.
		if ent.IsNotFound(err) || errors.Is(err, ErrInvalidInput) {
			return nil, ErrInvalidInput
		}
		return nil, fmt.Errorf("не удалось получить проект канала: %w", err)
	}
	if projectID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	// Проверяем лимит активных записей
	activeCount, err := s.repo.CountActiveRecordings(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки активных записей: %w", err)
	}
	if activeCount >= maxActiveRecordingsPerCall {
		return nil, fmt.Errorf("превышен лимит активных записей (%d)", maxActiveRecordingsPerCall)
	}

	// Формируем путь для S3: recordings/{projectID}/{callID}.mp4
	filePath := fmt.Sprintf("recordings/%s/%s.mp4", projectID.String(), callID.String())

	// Определяем S3 endpoint protocol
	protocol := "http"
	if s.s3Cfg.UseSSL {
		protocol = "https"
	}

	// Создаём Egress запрос
	egressReq := &livekit.RoomCompositeEgressRequest{
		RoomName: callEnt.RoomName,
		AudioOnly: true,
		Output: &livekit.RoomCompositeEgressRequest_File{
			File: &livekit.EncodedFileOutput{
				FileType: livekit.EncodedFileType_MP4,
				Filepath: filePath,
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey: s.s3Cfg.AccessKey,
						Secret:    s.s3Cfg.SecretKey,
						Region:    s.s3Cfg.Region,
						Bucket:    s.s3Cfg.Bucket,
						Endpoint:  fmt.Sprintf("%s://%s", protocol, s.s3Cfg.Endpoint),
						ForcePathStyle: true,
					},
				},
			},
		},
	}

	// Запускаем Egress
	egressInfo, err := s.egressClient.StartRoomCompositeEgress(ctx, egressReq)
	if err != nil {
		return nil, fmt.Errorf("не удалось запустить запись: %w", err)
	}

	// Сохраняем запись в БД. A2W3-5-001: при ошибке — StopEgress, иначе orphan egress в LiveKit/S3.
	rec, err := s.createRecordingRow(ctx, egressInfo.EgressId, callID, projectID, userID, filePath)
	if err != nil {
		s.compensateOrphanEgress(ctx, egressInfo.EgressId)
		return nil, fmt.Errorf("не удалось сохранить запись: %w", err)
	}

	// Публикуем NATS событие для уведомления участников
	s.nats.PublishJSON("calls.recording.started", map[string]interface{}{
		"recording_id": rec.ID,
		"call_id":      callID,
		"started_by":   userID,
		"channel_id":   callEnt.ChannelID,
	})

	return toRecordingResponse(rec), nil
}

// createRecordingRow пишет метаданные записи. В hubtest можно подменить через
// ExportRecordingServiceSetCreateRecording (мутация A2W3-5-001).
func (s *RecordingService) createRecordingRow(ctx context.Context, egressID string, callID, projectID, userID uuid.UUID, filePath string) (*ent.Recording, error) {
	if s.createRecording != nil {
		return s.createRecording(ctx, egressID, callID, projectID, userID, filePath)
	}
	return s.repo.CreateRecording(ctx, egressID, callID, projectID, userID, filePath)
}

// compensateOrphanEgress останавливает LiveKit egress после провала CreateRecording (A2W3-5-001).
func (s *RecordingService) compensateOrphanEgress(ctx context.Context, egressID string) {
	if s.egressClient == nil || egressID == "" {
		return
	}
	if _, stopErr := s.egressClient.StopEgress(ctx, &livekit.StopEgressRequest{EgressId: egressID}); stopErr != nil {
		slog.Error("recording: не удалось остановить orphan egress после ошибки CreateRecording",
			"egress_id", egressID, "error", stopErr)
	}
}

// StopRecording останавливает запись звонка через LiveKit Egress.
// A2-004: callID из path должен совпадать с rec.CallID; доступ — член канала
// (тот же authz, что Start) + userID в сервисе. Статус → processing + StopEgress.
func (s *RecordingService) StopRecording(ctx context.Context, callID, recordingID, userID uuid.UUID) error {
	rec, err := s.repo.GetRecording(ctx, recordingID)
	if err != nil {
		return fmt.Errorf("запись не найдена: %w", err)
	}
	if rec.CallID != callID {
		return ErrCallForbidden
	}

	callEnt, err := s.repo.GetCall(ctx, rec.CallID)
	if err != nil {
		return fmt.Errorf("звонок не найден: %w", err)
	}
	if authErr := s.authorizeCallRecordingAccess(ctx, callEnt, userID); authErr != nil {
		return authErr
	}

	if rec.Status != recording.StatusRecording {
		return fmt.Errorf("запись не активна (текущий статус: %s)", rec.Status)
	}

	// Останавливаем Egress
	_, err = s.egressClient.StopEgress(ctx, &livekit.StopEgressRequest{
		EgressId: rec.EgressID,
	})
	if err != nil {
		return fmt.Errorf("не удалось остановить запись: %w", err)
	}

	// Обновляем статус
	if err := s.repo.UpdateRecordingStatus(ctx, recordingID, recording.StatusProcessing); err != nil {
		return fmt.Errorf("не удалось обновить статус записи: %w", err)
	}

	s.nats.PublishJSON("calls.recording.stopped", map[string]interface{}{
		"recording_id": recordingID,
		"call_id":      rec.CallID,
		"stopped_by":   userID,
	})

	return nil
}

// OnEgressEnded обрабатывает webhook egress_ended от LiveKit.
// Обновляет метаданные записи (статус, размер, длительность) и публикует
// NATS-событие для отправки ссылки на запись в канал мессенджера.
func (s *RecordingService) OnEgressEnded(ctx context.Context, egressID string, fileSize int64, duration float64) error {
	rec, err := s.repo.GetRecordingByEgressID(ctx, egressID)
	if err != nil {
		return fmt.Errorf("запись не найдена по egress_id: %w", err)
	}

	now := time.Now()
	durationSec := int(duration)

	if err := s.repo.UpdateRecordingCompleted(ctx, rec.ID, fileSize, durationSec, now); err != nil {
		return fmt.Errorf("не удалось обновить запись: %w", err)
	}

	s.nats.PublishJSON("calls.recording.ready", map[string]interface{}{
		"recording_id": rec.ID,
		"call_id":      rec.CallID,
		"project_id":   rec.ProjectID,
		"file_size":    fileSize,
		"duration":     durationSec,
	})

	return nil
}

// ListProjectRecordings возвращает список записей проекта с пагинацией.
// Записи отсортированы по дате создания (новые первые).
func (s *RecordingService) ListProjectRecordings(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*RecordingResponse, int, error) {
	recordings, total, err := s.repo.ListRecordingsByProject(ctx, projectID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("не удалось получить записи: %w", err)
	}

	result := make([]*RecordingResponse, len(recordings))
	for i, r := range recordings {
		result[i] = toRecordingResponse(r)
	}

	return result, total, nil
}

// RecordingProjectID возвращает ID проекта, которому принадлежит запись.
// Используется обработчиком для проверки прав доступа до выдачи presigned URL.
func (s *RecordingService) RecordingProjectID(ctx context.Context, recordingID uuid.UUID) (uuid.UUID, error) {
	rec, err := s.repo.GetRecording(ctx, recordingID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("запись не найдена: %w", err)
	}
	return rec.ProjectID, nil
}

// GetRecordingDownloadURL генерирует presigned S3 URL для скачивания записи.
// URL действителен 1 час. Авторизацию должен выполнить обработчик до вызова
// (через RecordingProjectID + rbac.GetEffectivePermissions).
func (s *RecordingService) GetRecordingDownloadURL(ctx context.Context, recordingID uuid.UUID) (*RecordingDownloadResponse, error) {
	rec, err := s.repo.GetRecording(ctx, recordingID)
	if err != nil {
		return nil, fmt.Errorf("запись не найдена: %w", err)
	}

	if rec.Status != recording.StatusReady {
		return nil, fmt.Errorf("запись ещё не готова (статус: %s)", rec.Status)
	}

	if s.presignClient == nil {
		return nil, fmt.Errorf("presign-клиент не инициализирован")
	}

	expiresAt := time.Now().Add(downloadURLTTL)
	presignResult, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.s3Cfg.Bucket),
		Key:    aws.String(rec.FilePath),
	}, s3.WithPresignExpires(downloadURLTTL))
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать presigned URL: %w", err)
	}

	return &RecordingDownloadResponse{
		URL:       presignResult.URL,
		ExpiresAt: expiresAt,
	}, nil
}

// toRecordingResponse конвертирует Ent Recording в DTO ответа.
func toRecordingResponse(r *ent.Recording) *RecordingResponse {
	resp := &RecordingResponse{
		ID:              r.ID,
		EgressID:        r.EgressID,
		Status:          string(r.Status),
		FilePath:        r.FilePath,
		FileSize:        r.FileSize,
		DurationSeconds: r.DurationSeconds,
		StartedAt:       r.StartedAt,
		CallID:          r.CallID,
		ProjectID:       r.ProjectID,
		StartedBy:       r.StartedBy,
		CreatedAt:       r.CreatedAt,
	}
	if r.EndedAt != nil {
		resp.EndedAt = r.EndedAt
	}
	return resp
}
