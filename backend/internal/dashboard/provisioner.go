package dashboard

import (
	"context"

	"github.com/google/uuid"
)

// DashboardProjectProvisioner реализует интерфейс project.DashboardProvisioner.
// Применяет шаблон "dev" при создании нового проекта.
type DashboardProjectProvisioner struct {
	svc *DashboardService
}

// NewDashboardProjectProvisioner создаёт провизионер для автоинициализации дашборда.
func NewDashboardProjectProvisioner(svc *DashboardService) *DashboardProjectProvisioner {
	return &DashboardProjectProvisioner{svc: svc}
}

// CreateDefaultDashboard применяет шаблон "dev" для владельца нового проекта.
func (p *DashboardProjectProvisioner) CreateDefaultDashboard(ctx context.Context, ownerID, projectID uuid.UUID) error {
	return p.svc.ApplyTemplate(ctx, ownerID, projectID, "dev")
}
