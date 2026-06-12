package dependencies_bot

import (
	"database/sql"
	"os"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/application"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/adapters"
)

type Container struct {
	Repo                        *adapters.SupabaseRepositoryAdapter
	Fetcher                     *adapters.MLProductFetcherAdapter
	FetchAndSaveProductsUseCase *application.FetchAndSaveProductsUseCase
}

func NewContainer(db *sql.DB) *Container {
	siteID := os.Getenv("ML_SITE_ID")
	if siteID == "" {
		siteID = "MLM" // Default to Mexico
	}

	// Instantiate adapters using core.DB
	repo := adapters.NewSupabaseRepositoryAdapter(db)
	fetcher := adapters.NewMLProductFetcherAdapter(siteID)

	// Instantiate UseCase
	useCase := application.NewFetchAndSaveProductsUseCase(repo, fetcher)

	return &Container{
		Repo:                        repo,
		Fetcher:                     fetcher,
		FetchAndSaveProductsUseCase: useCase,
	}
}
