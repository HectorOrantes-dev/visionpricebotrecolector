package entities

import "time"

type PriceSnapshot struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Precio    float64   `json:"precio"`
	Moneda    string    `json:"moneda"`
	FetchedAt time.Time `json:"fetched_at"`
}
