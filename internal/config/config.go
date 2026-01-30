package config

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	DBConn string
}

func Load() (*Config, error) {
	// Intentamos cargar el .env pero no morimos si no está (útil para Docker/Prod)
	_ = godotenv.Load()

	// Validamos que las variables críticas existan
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	ssl := os.Getenv("DB_SSLMODE")

	if user == "" || pass == "" {
		return nil, fmt.Errorf("faltan variables de entorno críticas de la base de datos")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, dbPort, dbName, ssl,
	)

	return &Config{
		Port:   os.Getenv("PORT"), // Por defecto suele ser 8080
		DBConn: connStr,
	}, nil
}