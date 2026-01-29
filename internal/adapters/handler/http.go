package handler

import (
	"encoding/json"
	"net/http"
	"github.com/RiosHectorM/gopher-stream/internal/domain"
)

// AssetHandler traduce las peticiones HTTP para el servicio
type AssetHandler struct {
	service *domain.AssetService
}

func NewAssetHandler(s *domain.AssetService) *AssetHandler {
	return &AssetHandler{service: s}
}

func (h *AssetHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	// 1. Solo aceptamos POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decodificar el JSON que llega del sensor/frontend
	var event domain.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3. Pasar al servicio usando el contexto de la petición
	if err := h.service.ProcessMovement(r.Context(), event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Respuesta exitosa
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Evento recibido correctamente"))
}