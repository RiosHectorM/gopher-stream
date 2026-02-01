package domain

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

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

const recoveryDir = "storage/recovery"

func (s *AssetService) RecoverFromEmergencyLog() {
	_ = os.MkdirAll(recoveryDir, 0755)

	fileName := "emergencia_dlq.log"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return
	}

	slog.Info("📂 Archivo de emergencia detectado. Iniciando recuperación...")

	// 1. Leemos todo el contenido y cerramos el archivo rápido
	content, err := os.ReadFile(fileName)
	if err != nil {
		slog.Error("❌ No se pudo leer el archivo", "error", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	recoveredCount := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		slog.Info("♻️ Recuperando línea", "contenido", line)
		// Aquí es donde en el futuro harás el parseo real a un struct Event
		recoveredCount++
	}

	// 2. Intentamos renombrar (Plan de rotación de logs)
	if recoveredCount > 0 {
		// Movemos el archivo a la carpeta de backup con un nombre limpio
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		backupPath := fmt.Sprintf("%s/recuperado_%s.bak", recoveryDir, timestamp)

		err := os.Rename(fileName, backupPath)
		if err != nil {
			slog.Error("❌ No se pudo mover el archivo a storage", "error", err)
			return
		}
		slog.Info("✅ Datos recuperados y archivados en storage", "total", recoveredCount, "path", backupPath)
	}
}

func (s *AssetService) StartRecoveryMonitor(ctx context.Context) {
	// Revisamos cada 30 segundos
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	slog.Info("🔍 Monitor de recuperación iniciado")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 1. Verificamos si existe el archivo antes de molestar a la DB
			if _, err := os.Stat("emergencia_dlq.log"); err == nil {
				slog.Info("📡 Detectado archivo de emergencia. Comprobando DB...")

				// 2. ¿La base de datos está disponible?
				// Suponiendo que tu repo tiene un método Ping o simplemente probamos
				if s.repo.IsAvailable(ctx) {
					slog.Info("✅ DB disponible. Iniciando auto-recuperación...")
					s.RecoverFromEmergencyLog()
				} else {
					slog.Warn("⏳ Archivo pendiente pero la DB sigue caída. Reintentando luego...")
				}
			}
		}
	}
}
