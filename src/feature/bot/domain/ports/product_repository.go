package ports

import (
	"context"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type ProductRepository interface {
	Upsert(ctx context.Context, product *entities.Product) error
	SaveSnapshot(ctx context.Context, snapshot *entities.PriceSnapshot) error
	ListByCategory(ctx context.Context, category string) ([]entities.Product, error)
}
