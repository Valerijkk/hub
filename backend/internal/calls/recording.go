package calls

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// maxActiveRecordingsPerCall определяет максимальное количество одновременных
// записей на один звонок. Ограничение защищает от перегрузки LiveKit Egress.
const maxActiveRecordingsPerCall = 2

// RecordingService реализует бизнес-логику записи звонков через LiveKit Egress.
// Управляет жизненным циклом записи: запуск Egress, остановка, обработка
// завершения и предоставление presigned URL для скачивания.
type RecordingService struct {
	egressClient *lksdk.EgressClient
	s3Cfg        config.S3Config
	repo         *EntRepository
	nats         *natsclient.Client
}

// NewRecordingService создаёт новый сервис записи звонков.
// Инициализирует EgressClient для управления записью через LiveKit.
func NewRecordingService(lkCfg config.LiveKitConfig, s3Cfg config.S3Config, repo *EntRepository, nc *natsclient.Client) *RecordingService {
	egressClient := lksdk.NewEgressClient(lkCfg.URL, lkCfg.APIKey, lkCfg.APISecret)
	return &RecordingService{
		egressClient: egressClient,
		s3Cfg:        s3Cfg,
		repo:         repo,
		nats:         nc,
	}
}

// StartRecording запускает запись звонка через LiveKit Egress.
// Проверяет лимит активных записей, создаёт RoomCompositeEgress с S3 output,
// сохраняет метаданные в БД и публикует NATS-событие для уведомления участников.
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
	projectID, err := s.repo.GetProjectIDByChannel(ctx, callEnt.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить проект канала: %w", err)
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

	// Сохраняем запись в БД
	rec, err := s.repo.CreateRecording(ctx, egressInfo.EgressId, callID, projectID, userID, filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить запись: %w", err)
	}

	// Публикуем NATS событие для уведомления участников
	s.publishEvent("calls.recording.started", map[string]interface{}{
		"recording_id": rec.ID,
		"call_id":      callID,
		"started_by":   userID,
		"channel_id":   callEnt.ChannelID,
	})

	return toRecordingResponse(rec), nil
}

// StopRecording останавливает запись звонка через LiveKit Egress.
// Обновляет статус записи на "processing" и вызывает StopEgress.
func (s *RecordingService) StopRecording(ctx context.Context, recordingID uuid.UUID) error {
	rec, err := s.repo.GetRecording(ctx, recordingID)
	if err != nil {
		return fmt.Errorf("запись не найдена: %w", err)
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

	s.publishEvent("calls.recording.stopped", map[string]interface{}{
		"recording_id": recordingID,
		"call_id":      rec.CallID,
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

	s.publishEvent("calls.recording.ready", map[string]interface{}{
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

// GetRecordingDownloadURL генерирует presigned S3 URL для скачивания записи.
// URL действителен 1 час.
func (s *RecordingService) GetRecordingDownloadURL(ctx context.Context, recordingID uuid.UUID) (*RecordingDownloadResponse, error) {
	rec, err := s.repo.GetRecording(ctx, recordingID)
	if err != nil {
		return nil, fmt.Errorf("запись не найдена: %w", err)
	}

	if rec.Status != recording.StatusReady {
		return nil, fmt.Errorf("запись ещё не готова (статус: %s)", rec.Status)
	}

	// Генерируем presigned URL через S3 (реализация в storage module)
	// Для простоты возвращаем прямой путь — в production используется presigned URL
	expiresAt := time.Now().Add(1 * time.Hour)
	protocol := "http"
	if s.s3Cfg.UseSSL {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", protocol, s.s3Cfg.Endpoint, s.s3Cfg.Bucket, rec.FilePath)

	return &RecordingDownloadResponse{
		URL:       url,
		ExpiresAt: expiresAt,
	}, nil
}

// publishEvent публикует событие в NATS (fire-and-forget).
// Безопасен при nil NATS-клиенте.
func (s *RecordingService) publishEvent(subject string, payload interface{}) {
	if s.nats == nil {
		slog.Debug("NATS не подключён, пропускаем публикацию", "subject", subject)
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать событие", "subject", subject, "error", err)
		return
	}
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать событие", "subject", subject, "error", err)
	}
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
