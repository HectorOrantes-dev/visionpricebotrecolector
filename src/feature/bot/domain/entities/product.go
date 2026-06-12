package entities

import "time"

type Product struct {
	ID          string    `json:"id"`
	MLID        string    `json:"ml_id"`
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Precio      float64   `json:"precio"`
	Moneda      string    `json:"moneda"`
	Categoria   string    `json:"categoria"`
	CreatedAt   time.Time `json:"created_at"`
}
