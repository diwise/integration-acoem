package main

import (
	"context"

	"github.com/diwise/context-broker/pkg/ngsild/client"
	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/env"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"

	"github.com/diwise/integration-acoem/internal/pkg/application"
)

const serviceName string = "integration-acoem"

func main() {
	serviceVersion := buildinfo.SourceVersion()

	ctx, logger, cleanup := o11y.Init(context.Background(), serviceName, serviceVersion)
	defer cleanup()

	baseUrl := env.GetVariableOrDie(logger, "ACOEM_BASEURL", "acoem base url")
	accessToken := env.GetVariableOrDie(logger, "ACOEM_ACCESS_TOKEN", "acoem access token")
	contextBrokerUrl := env.GetVariableOrDie(logger, "CONTEXT_BROKER_URL", "context broker url")

	contextBroker := client.NewContextBrokerClient(contextBrokerUrl)

	a := application.New(ctx, baseUrl, accessToken, contextBroker)

	a.CreateAirQualityObserved(ctx)
}
