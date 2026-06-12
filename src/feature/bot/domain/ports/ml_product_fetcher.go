package ports

import (
	"context"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
)

type MLProductFetcher interface {
	FetchByCategory(ctx context.Context, category string) ([]entities.Product, error)
	FetchItemDescription(ctx context.Context, mlID string) (string, error)
}
