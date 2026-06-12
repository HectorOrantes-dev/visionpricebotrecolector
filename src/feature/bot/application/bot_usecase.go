package application

import (
	"context"
	"fmt"
	"time"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/domain/entities"
	"github.com/google/uuid"
)

type FetchAndSaveProductsUseCase struct {
	repo    domain.ProductRepository
	fetcher domain.MLProductFetcher
}

func NewFetchAndSaveProductsUseCase(repo domain.ProductRepository, fetcher domain.MLProductFetcher) *FetchAndSaveProductsUseCase {
	return &FetchAndSaveProductsUseCase{
		repo:    repo,
		fetcher: fetcher,
	}
}

func (uc *FetchAndSaveProductsUseCase) Execute(ctx context.Context, category string) error {
	// 1. Fetch products from Mercado Libre by category
	products, err := uc.fetcher.FetchByCategory(ctx, category)
	if err != nil {
		return fmt.Errorf("error fetching products from ML: %w", err)
	}

	for i := range products {
		product := &products[i]

		// 2. Fetch description
		desc, err := uc.fetcher.FetchItemDescription(ctx, product.MLID)
		if err != nil {
			fmt.Printf("warning: could not fetch description for %s: %v\n", product.MLID, err)
		} else {
			product.Descripcion = desc
		}

		// Generate ID if empty
		if product.ID == "" {
			product.ID = uuid.New().String()
		}

		// 3. Upsert product
		err = uc.repo.Upsert(ctx, product)
		if err != nil {
			fmt.Printf("error upserting product %s: %v\n", product.MLID, err)
			continue
		}

		// 4. Create and save price snapshot
		snapshot := &entities.PriceSnapshot{
			ID:        uuid.New().String(),
			ProductID: product.ID,
			Precio:    product.Precio,
			Moneda:    product.Moneda,
			FetchedAt: time.Now(),
		}

		err = uc.repo.SaveSnapshot(ctx, snapshot)
		if err != nil {
			fmt.Printf("error saving snapshot for product %s: %v\n", product.MLID, err)
		}
	}

	return nil
}
