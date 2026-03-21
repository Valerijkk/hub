package tms

import (
	"context"

	"github.com/google/uuid"
)

// TMSProjectProvisioner реализует интерфейс project.TMSProvisioner.
// Создаёт тестовую папку по умолчанию при создании нового проекта.
type TMSProjectProvisioner struct {
	repo Repository
}

// NewTMSProjectProvisioner создаёт провизионер для автоинициализации TMS.
func NewTMSProjectProvisioner(repo Repository) *TMSProjectProvisioner {
	return &TMSProjectProvisioner{repo: repo}
}

// CreateDefaultSuite создаёт папку "Общие тесты" по умолчанию для нового проекта.
func (p *TMSProjectProvisioner) CreateDefaultSuite(ctx context.Context, projectID uuid.UUID) error {
	_, err := p.repo.CreateSuite(ctx, projectID, CreateSuiteInput{
		Name:      "Общие тесты",
		SortOrder: 0,
	})
	return err
}
