package domain

import (
	"context"
	"fmt"
	"os"
	"time"
	"log/slog"
)

type AssetService struct {
	repo      AssetRepository
	eventChan chan Event
}

func NewAssetService(repo AssetRepository) *AssetService {
	s := &AssetService{
		repo:      repo,
		eventChan: make(chan Event, 100), // Buffer para 100 eventos
	}

	// Lanzamos 3 workers para procesar en paralelo
	for i := 1; i <= 3; i++ {
		go s.worker(i)
	}

	return s
}

func (s *AssetService) worker(id int) {
	slog.Info("👷 Worker iniciado", "worker_id", id)

	for event := range s.eventChan {
		maxRetries := 3
		success := false

		for i := 0; i < maxRetries; i++ {
			err := s.repo.UpdateLocation(context.Background(), event)
			if err == nil {
				slog.Info("✅ Ubicación actualizada", "worker_id", id, "asset_id", event.AssetID)
				success = true
				break
			}

			slog.Warn("❌ Error en intento de guardado",
				"worker_id", id,
				"attempt", i+1,
				"asset_id", event.AssetID,
				"error", err,
			)

			if i < maxRetries-1 {
				time.Sleep(time.Second * 2)
			}
		}

		// Si después de los reintentos no hubo éxito, vamos a la DLQ
		if !success {
			slog.Error("⚠️ Agotados reintentos. Intentando persistir en DB DLQ...", "asset_id", event.AssetID)
			
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			errDLQ := s.repo.SaveToDLQ(ctx, event, "Agotados reintentos")
			cancel()

			if errDLQ != nil {
				slog.Error("🚨 FALLÓ DB DLQ. Escribiendo en archivo de emergencia...", 
					"error", errDLQ, 
					"asset_id", event.AssetID,
				)

				f, err := os.OpenFile("emergencia_dlq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					slog.Error("❌ ERROR FATAL: No se pudo crear el archivo", "error", err)
				} else {
					// Guardamos la línea en un formato más limpio para la futura recuperación
					linea := fmt.Sprintf("TIME:%s|ID:%s|PAYLOAD:%s|ERR:%v\n",
						time.Now().Format(time.RFC3339), event.AssetID, event.Payload, errDLQ)
					
					_, _ = f.WriteString(linea)
					f.Close()
					slog.Info("💾 Datos salvados localmente", "archivo", "emergencia_dlq.log")
				}
			}
		}
	}
}

func (s *AssetService) ProcessMovement(ctx context.Context, event Event) error {
	// Aquí podrías validar datos (ej: lat/long válidas)
	if event.AssetID == "" {
		return fmt.Errorf("asset_id es obligatorio")
	}

	// Mandamos al canal y liberamos el Handler inmediatamente
	s.eventChan <- event
	return nil
}
