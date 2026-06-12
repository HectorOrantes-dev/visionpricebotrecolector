package domain

import (
	"context"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type ProductRepository interface {
	Upsert(ctx context.Context, product *entities.Product) error
	SaveSnapshot(ctx context.Context, snapshot *entities.PriceSnapshot) error
	ListByCategory(ctx context.Context, category string) ([]entities.Product, error)
}

type MLProductFetcher interface {
	FetchByCategory(ctx context.Context, category string) ([]entities.Product, error)
	FetchItemDescription(ctx context.Context, mlID string) (string, error)
}
