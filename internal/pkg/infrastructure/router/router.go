package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type Router interface {
	Start(port string) error
}

type routerStruct struct {
	router chi.Router
	log    zerolog.Logger
}

func SetupRouter(chiRouter chi.Router, log zerolog.Logger) *routerStruct {
	r := &routerStruct{
		router: chiRouter,
		log:    log,
	}

	chiRouter.Use(middleware.Logger)
	chiRouter.Get("/health", r.health)

	return r
}

func (r *routerStruct) Start(port string) error {
	r.log.Info().Str("port", port).Msg("starting to listen for connections")
	return http.ListenAndServe(fmt.Sprintf(":%s", port), r.router)
}

func (router *routerStruct) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
