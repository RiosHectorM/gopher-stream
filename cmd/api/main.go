package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RiosHectorM/gopher-stream/internal/adapters/repository"
	"github.com/RiosHectorM/gopher-stream/internal/domain"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Cargar configuración segura
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	// 2. Contexto principal de la aplicación
	// Este contexto gobernará a todos los demás.
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Conexión a la DB con Timeout inicial
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"),
	)

	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatalf("❌ Error crítico: No se pudo conectar a la DB: %v", err)
	}

	// 4. Inyección de dependencias
	service := domain.NewAssetService(repo)

	_ = service
	fmt.Println("🚀 GopherStream API iniciada: Servicio de activos listo.")

	fmt.Println("📡 Esperando eventos de hardware / señales de interrupción...")

	// 5. El "Bucle" de espera o Servidor HTTP
	// Por ahora, simulamos que la app está viva hasta que reciba una señal de apagado.
	<-appCtx.Done()

	// 6. Graceful Shutdown
	fmt.Println("\n⚠️ Apagando servidor de forma segura...")

	// Damos 5 segundos para que los procesos pendientes terminen
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("🛑 API apagada. Todos los recursos liberados.")
}
