package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RiosHectorM/gopher-stream/internal/adapters/handler"
	"github.com/RiosHectorM/gopher-stream/internal/adapters/repository"
	"github.com/RiosHectorM/gopher-stream/internal/config"
	"github.com/RiosHectorM/gopher-stream/internal/domain"
)

func main() {
	// 1. Configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	// 2. Infraestructura (DB)
	repo, err := repository.NewPostgresRepository(cfg.DBConn)
	if err != nil {
		log.Fatalf("❌ Error crítico DB: %v", err)
	}

	// 3. Dominio e Interfaz (Inyección)
	service := domain.NewAssetService(repo)
	h := handler.NewAssetHandler(service)

	// 4. Servidor
	mux := http.NewServeMux()
	mux.HandleFunc("/tracking", h.UpdateLocation)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// 5. Orquestación (Graceful Shutdown)
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("🌐 Servidor escuchando en puerto %s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error: %v", err)
		}
	}()

	<-appCtx.Done()

	// Cierre elegante
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error al apagar: %v", err)
	}
	fmt.Println("🛑 API apagada correctamente.")
}
