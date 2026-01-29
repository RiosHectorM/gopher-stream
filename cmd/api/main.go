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
	"github.com/RiosHectorM/gopher-stream/internal/domain"
	"github.com/joho/godotenv"
)

func main() {
	// 1. CARGAR CONFIGURACIÓN
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ Usando variables de entorno del sistema")
	}

	// 2. CONTEXTO DE CIERRE (Graceful Shutdown)
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. CONEXIÓN A DB
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"),
	)

	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatalf("❌ Error crítico DB: %v", err)
	}

	// 4. INYECCIÓN DE DEPENDENCIAS (Orden correcto)
	// Primero el servicio (Lógica de negocio)
	service := domain.NewAssetService(repo)
	
	// Segundo el handler (Interfaz HTTP) que usa el servicio
	h := handler.NewAssetHandler(service)

	// 5. DEFINIR RUTAS
	http.HandleFunc("/tracking", h.UpdateLocation)

	// 6. LEVANTAR SERVIDOR EN GOROUTINE
	server := &http.Server{Addr: ":8080"}

	go func() {
		fmt.Println("🌐 Servidor HTTP escuchando en http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error en el servidor: %v", err)
		}
	}()

	fmt.Println("🚀 GopherStream API iniciada y lista.")
	fmt.Println("📡 Esperando eventos...")

	// 7. ESPERAR SEÑAL DE APAGADO
	<-appCtx.Done()

	// 8. CIERRE ELEGANTE
	fmt.Println("\n⚠️ Apagando servidor de forma segura...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error al cerrar el servidor: %v", err)
	}

	fmt.Println("🛑 API apagada correctamente.")
}