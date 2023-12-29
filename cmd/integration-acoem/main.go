package main

import (
	"context"
	"flag"
	"os"

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

	baseUrl := env.GetVariableOrDie(ctx, "ACOEM_BASEURL", "acoem base url")
	accountID := env.GetVariableOrDie(ctx, "ACOEM_ACCOUNT_ID", "acoem account ID")
	accountKey := env.GetVariableOrDie(ctx, "ACOEM_ACCOUNT_KEY", "acoem account key")
	cipUrl := env.GetVariableOrDefault(ctx, "CONTEXT_BROKER_URL", "")
	lwm2mUrl := env.GetVariableOrDefault(ctx, "LWM2M_ENDPOINT_URL", "")

	if outputType == OutputTypeLwm2m {
		if lwm2mUrl == "" {
			logger.Error("no URL to lwm2m endpoint specified using env. var LWM2M_ENDPOINT_URL")
			os.Exit(1)
		}
	}

	if outputType == OutputTypeFiware {
		if cipUrl == "" {
			logger.Error("no URL to context broker specified using env. var CONTEXT_BROKER_URL")
			os.Exit(1)
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
		logger.Error("no output type selected")
		os.Exit(1)
	}

	a := application.New(baseUrl, accountID, accountKey)

	devices, err := a.GetDevices(ctx)
	if err != nil {
		logger.Error("failed to retrieve devices", "err", err.Error())
	}

	contextBroker := client.NewContextBrokerClient(cipUrl)

	for _, d := range devices {
		sensorLabels, err := a.GetSensorLabels(ctx, d.UniqueId)
		if err != nil {
			logger.Error("failed to retrieve sensor labels for device", "device_id", d.UniqueId, "err", err.Error())
		}

		logger.Info("retrieving data", "sensor_labels", sensorLabels, "device_id", d.UniqueId)

		sensors, err := a.GetDeviceData(ctx, d.UniqueId, sensorLabels)
		if err != nil {
			logger.Error("failed to retrieve sensor data", "err", err.Error())
		}

		if outputType == OutputTypeFiware {
			fiware.CreateOrUpdateAirQualityObserved(ctx, contextBroker, sensors, d.DeviceName, d.UniqueId)
		}

		if outputType == OutputTypeLwm2m {
			lwm2m.CreateAndSendAsLWM2M(ctx, sensors, d.UniqueId, lwm2mUrl, lwm2m.Send)
		}
	}

}
