package domain

import (
	"context"
	"fmt"
	"time"
)

// AssetService es el "cerebro" que maneja la lógica de los activos
type AssetService struct {
	repo AssetRepository
}

// NewAssetService es el constructor que te faltaba
func NewAssetService(repo AssetRepository) *AssetService {
	return &AssetService{
		repo: repo,
	}
}

// ProcessMovement aplica reglas de negocio antes de guardar
func (s *AssetService) ProcessMovement(ctx context.Context, event Event) error {
	// 1. Validación de seguridad (Ciberseguridad básica: no confiar en el input)
	if event.AssetID == "" {
		return fmt.Errorf("el ID del activo es obligatorio")
	}

	// 2. Lógica de tiempo: si el sensor no mandó hora, usamos la del sistema
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 3. Mandamos al repositorio (PostgreSQL)
	return s.repo.UpdateLocation(ctx, event)
}
