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

// TMSProvisioner — интерфейс для автоматического создания тест-сьюта при создании проекта.
type TMSProvisioner interface {
	// CreateDefaultSuite создаёт папку "Общие тесты" по умолчанию для нового проекта.
	CreateDefaultSuite(ctx context.Context, projectID uuid.UUID) error
}

// DashboardProvisioner — интерфейс для автоматического создания раскладки дашборда при создании проекта.
type DashboardProvisioner interface {
	// CreateDefaultDashboard применяет шаблон "dev" для владельца нового проекта.
	CreateDefaultDashboard(ctx context.Context, ownerID, projectID uuid.UUID) error
}

// SetTrackerProvisioner внедряет сервис автопровизии трекера в ProjectService.
// Вызывается из main.go после инициализации обоих сервисов.
func (s *ProjectService) SetTrackerProvisioner(tp TrackerProvisioner) {
	s.trackerProv = tp
}

// SetTMSProvisioner внедряет сервис автопровизии TMS в ProjectService.
// Вызывается из main.go после инициализации TMS-сервиса.
func (s *ProjectService) SetTMSProvisioner(tp TMSProvisioner) {
	s.tmsProv = tp
}

// SetDashboardProvisioner внедряет сервис автопровизии дашборда в ProjectService.
// Вызывается из main.go после инициализации Dashboard-сервиса.
func (s *ProjectService) SetDashboardProvisioner(dp DashboardProvisioner) {
	s.dashProv = dp
}

// provisionProject выполняет автоматическую инициализацию ресурсов нового проекта:
// создаёт workflow по умолчанию, доску "Main Board", тест-сьют и дашборд.
func (s *ProjectService) provisionProject(ctx context.Context, projectID uuid.UUID, ownerID uuid.UUID) {
	if s.trackerProv != nil {
		if err := s.trackerProv.SeedDefaultWorkflow(ctx, projectID); err != nil {
			slog.Error("не удалось создать workflow по умолчанию",
				"project_id", projectID, "error", err)
		}

		if err := s.trackerProv.CreateDefaultBoard(ctx, projectID); err != nil {
			slog.Error("не удалось создать доску по умолчанию",
				"project_id", projectID, "error", err)
		}
	}

	if s.tmsProv != nil {
		if err := s.tmsProv.CreateDefaultSuite(ctx, projectID); err != nil {
			slog.Error("не удалось создать тест-сьют по умолчанию",
				"project_id", projectID, "error", err)
		}
	}

	if s.dashProv != nil {
		if err := s.dashProv.CreateDefaultDashboard(ctx, ownerID, projectID); err != nil {
			slog.Error("не удалось создать дашборд по умолчанию",
				"project_id", projectID, "error", err)
		}
	}

	slog.Info("проект автопровизирован", "project_id", projectID)
}
