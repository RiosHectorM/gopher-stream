package domain

import (
	"context"
	"fmt"
	"os"
	"time"
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

// El worker con lógica de reintento básica
func (s *AssetService) worker(id int) {
	fmt.Printf("👷 Worker %d iniciado\n", id)
	for event := range s.eventChan {
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			err := s.repo.UpdateLocation(context.Background(), event)
			if err == nil {
				break // Éxito, salimos del bucle de reintentos
			}

			fmt.Printf("❌ Worker %d: Error en intento %d para %s: %v\n", id, i+1, event.AssetID, err)
			if i < maxRetries-1 {
				time.Sleep(time.Second * 2)
			} else {
				fmt.Printf("⚠️ Worker %d: Agotado. Intentando guardar en DB DLQ...\n", id)
				errDLQ := s.repo.SaveToDLQ(context.Background(), event, "Agotados reintentos")

				if errDLQ != nil {
					// ¡PLAN C! Si la DB está caída del todo, guardamos en un archivo local
					fmt.Printf("🚨 ERROR CRÍTICO: DB inaccesible. Guardando en emergencia_log.json\n")

					linea := fmt.Sprintf(`{"time": "%s", "asset": "%s", "data": "%s", "error": "%v"}\n`,
						time.Now().Format(time.RFC3339), event.AssetID, event.Payload, errDLQ)

					// Escribimos en un archivo (append mode)
					f, _ := os.OpenFile("emergencia_dlq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					f.WriteString(linea)
					f.Close()
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
