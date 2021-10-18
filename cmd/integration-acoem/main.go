package main

import (
	"os"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"

	"github.com/diwise/integration-acoem/internal/pkg/application"
	router "github.com/diwise/integration-acoem/internal/pkg/infrastructure/router"
)

func main() {
	baseUrl := os.Getenv("ACOEM_BASEURL")
	accountID := os.Getenv("ACOEM_ACCOUNT_ID")
	accountKey := os.Getenv("ACOEM_ACCOUNT_KEY")
	interval := os.Getenv("INTEGRATION_ACOEM_INTERVAL")
	if interval == "" {
		interval = "60"
	}
	port := os.Getenv("INTEGRATION_ACOEM_PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	router := router.SetupRouter(r, log.Logger)

	go application.Run(baseUrl, accountID, accountKey, interval, log.Logger)

	router.Start(port)

}
