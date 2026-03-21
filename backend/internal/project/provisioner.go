package project

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// TrackerProvisioner — интерфейс для автоматической инициализации трекера при создании проекта.
type TrackerProvisioner interface {
	// SeedDefaultWorkflow создаёт 6 стандартных статусов и переходы workflow для нового проекта.
	SeedDefaultWorkflow(ctx context.Context, projectID uuid.UUID) error
	// CreateDefaultBoard создаёт доску "Main Board" по умолчанию для нового проекта.
	CreateDefaultBoard(ctx context.Context, projectID uuid.UUID) error
}

// SetTrackerProvisioner внедряет сервис автопровизии трекера в ProjectService.
// Вызывается из main.go после инициализации обоих сервисов.
func (s *ProjectService) SetTrackerProvisioner(tp TrackerProvisioner) {
	s.trackerProv = tp
}

// provisionProject выполняет автоматическую инициализацию ресурсов нового проекта:
// создаёт workflow по умолчанию и доску "Main Board".
func (s *ProjectService) provisionProject(ctx context.Context, projectID uuid.UUID) {
	if s.trackerProv == nil {
		return
	}

	if err := s.trackerProv.SeedDefaultWorkflow(ctx, projectID); err != nil {
		slog.Error("не удалось создать workflow по умолчанию",
			"project_id", projectID, "error", err)
	}

	if err := s.trackerProv.CreateDefaultBoard(ctx, projectID); err != nil {
		slog.Error("не удалось создать доску по умолчанию",
			"project_id", projectID, "error", err)
	}

	slog.Info("проект автопровизирован", "project_id", projectID)
}
