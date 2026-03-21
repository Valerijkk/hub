package tracker

import (
	"context"

	"github.com/google/uuid"
)

// TrackerProjectProvisioner реализует интерфейс project.TrackerProvisioner.
// Создаёт workflow по умолчанию и доску "Main Board" при создании проекта.
type TrackerProjectProvisioner struct {
	repo Repository
}

// NewTrackerProjectProvisioner создаёт провизионер для автоинициализации трекера.
func NewTrackerProjectProvisioner(repo Repository) *TrackerProjectProvisioner {
	return &TrackerProjectProvisioner{repo: repo}
}

// SeedDefaultWorkflow создаёт 6 стандартных статусов и переходы workflow для нового проекта.
func (p *TrackerProjectProvisioner) SeedDefaultWorkflow(ctx context.Context, projectID uuid.UUID) error {
	return p.repo.SeedDefaultWorkflow(ctx, projectID)
}

// CreateDefaultBoard создаёт доску "Main Board" по умолчанию для нового проекта.
func (p *TrackerProjectProvisioner) CreateDefaultBoard(ctx context.Context, projectID uuid.UUID) error {
	_, err := p.repo.CreateBoard(ctx, projectID, "Main Board", nil)
	return err
}
