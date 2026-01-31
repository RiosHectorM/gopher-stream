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
				fmt.Println("⚠️ Agotados reintentos. Intentando DB...")
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // Timeout corto para que no cuelgue
				errDLQ := s.repo.SaveToDLQ(ctx, event, "Agotados reintentos")
				cancel()

				if errDLQ != nil {
					fmt.Printf("🚨 FALLÓ DB DLQ: %v. Escribiendo en archivo...\n", errDLQ)

					// Usamos una ruta más explícita para estar seguros
					f, err := os.OpenFile("emergencia_dlq.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Printf("❌ ERROR FATAL: Ni siquiera pude crear el archivo: %v\n", err)
					} else {
						linea := fmt.Sprintf("TIME: %s | ASSET: %s | DATA: %s | ERR: %v\n",
							time.Now().Format(time.RFC3339), event.AssetID, event.Payload, errDLQ)

						_, _ = f.WriteString(linea)
						f.Close()
						fmt.Println("✅ Datos salvados en emergencia_dlq.log")
					}
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
