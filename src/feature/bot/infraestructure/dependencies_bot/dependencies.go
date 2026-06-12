package dependencies_bot

import (
	"context"
	"fmt"
	"os"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/application"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/adapters"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Pool                        *pgxpool.Pool
	Repo                        *adapters.SupabaseRepositoryAdapter
	Fetcher                     *adapters.MLProductFetcherAdapter
	FetchAndSaveProductsUseCase *application.FetchAndSaveProductsUseCase
}

func NewContainer(ctx context.Context) (*Container, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	siteID := os.Getenv("ML_SITE_ID")
	if siteID == "" {
		siteID = "MLM" // Default to Mexico
	}

	// Connect to database using pgxpool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Ping database to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Instantiate adapters
	repo := adapters.NewSupabaseRepositoryAdapter(pool)
	fetcher := adapters.NewMLProductFetcherAdapter(siteID)

	// Instantiate UseCase
	useCase := application.NewFetchAndSaveProductsUseCase(repo, fetcher)

	return &Container{
		Pool:                        pool,
		Repo:                        repo,
		Fetcher:                     fetcher,
		FetchAndSaveProductsUseCase: useCase,
	}, nil
}

func (c *Container) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}
