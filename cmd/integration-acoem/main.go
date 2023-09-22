package main

import (
	"context"
	"flag"

	"github.com/diwise/context-broker/pkg/ngsild/client"
	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/env"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"

	"github.com/diwise/integration-acoem/internal/pkg/application"
	"github.com/diwise/integration-acoem/internal/pkg/application/fiware"
	"github.com/diwise/integration-acoem/internal/pkg/application/lwm2m"
)

const (
	serviceName      string = "integration-acoem"
	OutputTypeLwm2m  string = "lwm2m"
	OutputTypeFiware string = "fiware"
)

func main() {
	serviceVersion := buildinfo.SourceVersion()

	ctx, logger, cleanup := o11y.Init(context.Background(), serviceName, serviceVersion)
	defer cleanup()

	var outputType string

	flag.StringVar(&outputType, "output", "", "-output=<lwm2m or fiware>")
	flag.Parse()

	baseUrl := env.GetVariableOrDie(logger, "ACOEM_BASEURL", "acoem base url")
	accountID := env.GetVariableOrDie(logger, "ACOEM_ACCOUNT_ID", "acoem account ID")
	accountKey := env.GetVariableOrDie(logger, "ACOEM_ACCOUNT_KEY", "acoem account key")
	cipUrl := env.GetVariableOrDie(logger, "CONTEXT_BROKER_URL", "context broker url")
	lwm2mUrl := env.GetVariableOrDefault(logger, "LWM2M_ENDPOINT_URL", "")

	if outputType == OutputTypeLwm2m {
		if lwm2mUrl == "" {
			logger.Fatal().Msg("no URL to lwm2m endpoint specified using env. var LWM2M_ENDPOINT_URL")
		}
	}

	if outputType == OutputTypeFiware {
		if cipUrl == "" {
			logger.Fatal().Msg("no URL to context broker specified using env. var CONTEXT_BROKER_URL")
		}
	}

	if outputType == "" {
		if lwm2mUrl != "" {
			outputType = OutputTypeLwm2m
		} else if cipUrl != "" {
			outputType = OutputTypeFiware
		}
	}

	if outputType == "" {
		logger.Fatal().Msg("no output type selected")
	}

	contextBroker := client.NewContextBrokerClient(cipUrl)

	a := application.New(baseUrl, accountID, accountKey, contextBroker)

	devices, err := a.GetDevices(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve devices")
	}

	for _, d := range devices {
		sensorLabels, err := a.GetSensorLabels(ctx, d.UniqueId)
		if err != nil {
			logger.Error().Err(err).Msgf("failed to retrieve sensor labels for device %d", d.UniqueId)
		}

		logger.Info().Msgf("retrieving data for %s from %d", sensorLabels, d.UniqueId)

		sensors, err := a.GetDeviceData(ctx, d.UniqueId, sensorLabels)
		if err != nil {
			logger.Error().Err(err).Msg("failed to retrieve sensor data")
		}

		if outputType == OutputTypeFiware {
			fiware.CreateOrUpdateAirQualityObserved(ctx, contextBroker, sensors, d.DeviceName, d.UniqueId)
		}

		if outputType == OutputTypeLwm2m {
			lwm2m.CreateAndSendAsLWM2M(ctx, sensors, d.UniqueId, lwm2mUrl, lwm2m.Send)
		}
	}

}
