package routers

import (
	"encoding/json"
	"net/http"

	"github.com/HectorOrantes-dev/visionpricebotrecolector/src/feature/bot/infraestructure/dependencies_bot"
)

type Router struct {
	container *dependencies_bot.Container
}

func NewRouter(container *dependencies_bot.Container) *Router {
	return &Router{
		container: container,
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", r.handleHealth)
	mux.HandleFunc("/sync", r.handleSync)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (r *Router) handleSync(w http.ResponseWriter, req *http.Request) {
	// Accept POST or GET (for easy manual browser invocation)
	if req.Method != http.MethodPost && req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category := req.URL.Query().Get("category")
	if category == "" {
		category = "materiales de construccion" // Default query/category
	}

	ctx := req.Context()
	err := r.container.FetchAndSaveProductsUseCase.Execute(ctx, category)
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
