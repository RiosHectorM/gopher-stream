package handler

import (
	"net/http"
	"os"
)

// AuthMiddleware protege los endpoints con una API Key
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extraer la llave del Header "X-API-KEY"
		apiKey := r.Header.Get("X-API-KEY")

		// 2. Comparar con la que tenemos en nuestro .env
		// En producción, esto podría venir de una DB o Vault
		expectedKey := os.Getenv("APP_API_KEY")

		if apiKey == "" || apiKey != expectedKey {
			http.Error(w, "🚫 No autorizado: API Key inválida o ausente", http.StatusUnauthorized)
			return
		}

		// 3. Si todo está OK, pasar al siguiente handler
		next.ServeHTTP(w, r)
	}
}