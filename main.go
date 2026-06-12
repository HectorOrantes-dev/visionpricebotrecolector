package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/core"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/dependencies_bot"
	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/routers"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load env variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading, relying on system environment variables")
	}

	// Initialize core Supabase API configuration
	core.InitSupabase()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize dependencies container using Supabase API keys
	container := dependencies_bot.NewContainer(core.SupabaseURL, core.SupabaseKey)

	// Initialize background cron scheduler
	c := cron.New()

	cronExpr := os.Getenv("SYNC_CRON")
	if cronExpr == "" {
		cronExpr = "0 * * * *" // default: every hour
	}
	syncCategory := os.Getenv("SYNC_CATEGORY")
	if syncCategory == "" {
		syncCategory = "materiales de construccion"
	}

	_, err := c.AddFunc(cronExpr, func() {
		log.Printf("Starting scheduled sync for category: %s\n", syncCategory)
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := container.FetchAndSaveProductsUseCase.Execute(bgCtx, syncCategory); err != nil {
			log.Printf("Error during scheduled sync: %v\n", err)
		} else {
			log.Printf("Scheduled sync completed successfully for category: %s\n", syncCategory)
		}
	})
	if err != nil {
		log.Fatalf("Failed to schedule task: %v", err)
	}

	c.Start()
	defer c.Stop()
	log.Printf("Scheduler started. Sync task registered with cron: '%s'\n", cronExpr)

	// Setup HTTP Server
	mux := http.NewServeMux()
	router := routers.NewRouter(container)
	router.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP server starting on port %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	// Wait for interrupt signals
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}
	log.Println("HTTP server stopped")
}
