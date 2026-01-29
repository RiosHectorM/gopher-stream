package domain

import "time"

// Asset representa el recurso que estamos trackeando (un paquete, un móvil, etc.)
type Asset struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`   // ej: "biological_transport"
	Status     string    `json:"status"` // ej: "in_transit", "delivered"
	LastUpdate time.Time `json:"last_update"`
}

// Event es la señal que llega desde un sensor o GPS
type Event struct {
	AssetID   string    `json:"asset_id"`
	Lat       float64   `json:"lat"`
	Long      float64   `json:"long"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"` // Info extra del hardware (sensores)
}

// AssetRepository es el "contrato" (puerto). No nos importa si es Postgres o Mongo.
type AssetRepository interface {
	UpdateLocation(event Event) error
	GetByID(id string) (Asset, error)
}
