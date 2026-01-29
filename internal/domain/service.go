package domain

import (
	"context"
	"fmt"
	"time"
)

// AssetService es el cerebro. Contiene el repositorio (la DB)
// pero no sabe que es Postgres, solo sabe que puede "UpdateLocation".
type AssetService struct {
	repo AssetRepository
}

// NewAssetService es el constructor que usamos en el main.go
func NewAssetService(repo AssetRepository) *AssetService {
	return &AssetService{
		repo: repo,
	}
}

// ProcessMovement es la "Lógica de Negocio" propiamente dicha.
func (s *AssetService) ProcessMovement(ctx context.Context, event Event) error {

	// 1. REGLA DE NEGOCIO: El ID del activo no puede ser vacío
	if event.AssetID == "" {
		return fmt.Errorf("el ID del activo es obligatorio para el tracking")
	}

	// 2. REGLA DE NEGOCIO: Si el sensor no mandó Timestamp, usamos la hora del servidor
	// Esto es clave para el tracking de logística
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 3. REGLA DE NEGOCIO: Validar que las coordenadas sean reales
	if event.Lat == 0 && event.Long == 0 {
		return fmt.Errorf("coordenadas inválidas: (0,0) no es una posición de tracking permitida")
	}

	// 4. PERSISTENCIA: Si todo está OK, le pedimos al repositorio que lo guarde
	// El service no sabe SQL, solo sabe mandar la orden.
	err := s.repo.UpdateLocation(ctx, event)
	if err != nil {
		return fmt.Errorf("error al persistir el movimiento: %w", err)
	}

	return nil
}
