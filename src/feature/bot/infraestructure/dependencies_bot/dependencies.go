package dependencies_bot

import (
	"database/sql"
	"os"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/application"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/adapters"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/controllers"
)

type Container struct {
	Repo                        *adapters.SupabaseRepositoryAdapter
	Fetcher                     *adapters.MLProductFetcherAdapter
	FetchAndSaveProductsUseCase *application.FetchAndSaveProductsUseCase
	BotController               *controllers.BotController
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

	// Instantiate Controller
	botController := controllers.NewBotController(useCase)

	return &Container{
		Repo:                        repo,
		Fetcher:                     fetcher,
		FetchAndSaveProductsUseCase: useCase,
		BotController:               botController,
	}
}
