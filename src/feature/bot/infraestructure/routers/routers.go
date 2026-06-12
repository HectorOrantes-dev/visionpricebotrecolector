package routers

import (
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
	mux.HandleFunc("/", r.container.BotController.HandleRoot)
	mux.HandleFunc("/health", r.container.BotController.HandleHealth)
	mux.HandleFunc("/sync", r.container.BotController.HandleSync)
}
