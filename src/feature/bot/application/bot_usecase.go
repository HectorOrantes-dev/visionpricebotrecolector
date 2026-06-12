package application

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("[UseCase] Starting sync process for category/query: '%s'...\n", category)

	// 1. Fetch products from Mercado Libre by category
	products, err := uc.fetcher.FetchByCategory(ctx, category)
	if err != nil {
		return fmt.Errorf("error fetching products from ML: %w", err)
	}

	log.Printf("[UseCase] Successfully retrieved %d products from Mercado Libre.\n", len(products))

	for i := range products {
		product := &products[i]
		log.Printf("[UseCase] Processing product [%d/%d]: MLID=%s, Nombre='%s', Precio=%.2f %s\n", i+1, len(products), product.MLID, product.Nombre, product.Precio, product.Moneda)

		// 2. Fetch description
		log.Printf("[UseCase] Fetching description for product %s...\n", product.MLID)
		desc, err := uc.fetcher.FetchItemDescription(ctx, product.MLID)
		if err != nil {
			log.Printf("[UseCase] Warning: could not fetch description for %s: %v\n", product.MLID, err)
		} else {
			product.Descripcion = desc
			log.Printf("[UseCase] Description fetched successfully (%d chars) for product %s.\n", len(desc), product.MLID)
		}

		// Generate ID if empty
		if product.ID == "" {
			product.ID = uuid.New().String()
		}

		// 3. Upsert product
		log.Printf("[UseCase] Saving/Upserting product %s in Supabase (ID=%s)...\n", product.MLID, product.ID)
		err = uc.repo.Upsert(ctx, product)
		if err != nil {
			log.Printf("[UseCase] Error upserting product %s: %v\n", product.MLID, err)
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

		log.Printf("[UseCase] Creating price snapshot (ID=%s) for product %s...\n", snapshot.ID, product.MLID)
		err = uc.repo.SaveSnapshot(ctx, snapshot)
		if err != nil {
			log.Printf("[UseCase] Error saving price snapshot for product %s: %v\n", product.MLID, err)
		} else {
			log.Printf("[UseCase] Successfully saved product %s and its price snapshot.\n", product.MLID)
		}
	}

	log.Printf("[UseCase] Sync process finished successfully for category: '%s'.\n", category)
	return nil
}
