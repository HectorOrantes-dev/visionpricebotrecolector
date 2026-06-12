package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/application"
)

type BotController struct {
	useCase *application.FetchAndSaveProductsUseCase
}

func NewBotController(useCase *application.FetchAndSaveProductsUseCase) *BotController {
	return &BotController{
		useCase: useCase,
	}
}

func (c *BotController) HandleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "running",
		"message":     "Vision Price Bot Recolector is active.",
		"endpoints":   []string{"/health", "/sync"},
		"description": "Trigger manual synchronization by visiting /sync?category=your_category",
	})
}

func (c *BotController) HandleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (c *BotController) HandleSync(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost && req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := req.URL.Query().Get("category")
	if category == "" {
		category = "materiales de construccion"
	}

	ctx := req.Context()
	err := c.useCase.Execute(ctx, category)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "sync triggered successfully for category: " + category,
	})
}
