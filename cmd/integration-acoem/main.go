package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
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

	ctx, logger, cleanup := o11y.Init(context.Background(), serviceName, serviceVersion, "json")
	defer cleanup()

	var outputType string

	flag.StringVar(&outputType, "output", OutputTypeFiware, "-output=<lwm2m or fiware>")
	testMode := flag.Bool("test", false, "Run in test mode with a local mock CIP server")
	flag.Parse()

	baseUrl := env.GetVariableOrDie(ctx, "ACOEM_BASEURL", "acoem base url")
	accountID := env.GetVariableOrDie(ctx, "ACOEM_ACCOUNT_ID", "acoem account ID")
	accountKey := env.GetVariableOrDie(ctx, "ACOEM_ACCOUNT_KEY", "acoem account key")
	cipUrl := env.GetVariableOrDefault(ctx, "CONTEXT_BROKER_URL", "")
	lwm2mUrl := env.GetVariableOrDefault(ctx, "LWM2M_ENDPOINT_URL", "")

	if outputType == OutputTypeFiware {
		if cipUrl == "" {
			logger.Error("no URL to context broker specified using env. var CONTEXT_BROKER_URL")
			os.Exit(1)
		}
	}

	if outputType == OutputTypeLwm2m {
		if lwm2mUrl == "" {
			logger.Error("no URL to lwm2m endpoint specified using env. var LWM2M_ENDPOINT_URL")
			os.Exit(1)
		}
	}

	if *testMode {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
				report := &struct {
					Type   string `json:"type"`
					Title  string `json:"title"`
					Detail string `json:"detail"`
				}{
					Type:   "https://uri.etsi.org/ngsi-ld/errors/ResourceNotFound",
					Title:  "Not Found",
					Detail: "The requested entity was not found",
				}
				w.Header().Set("Content-Type", "application/ld+json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(report)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("test server: failed to read body", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			var prettyJSON map[string]any
			err = json.Unmarshal(body, &prettyJSON)
			if err != nil {
				logger.Error("test server: failed to unmarshal body", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			prettyBody, err := json.MarshalIndent(prettyJSON, "", "  ")
			if err != nil {
				logger.Error("test server: failed to marshal body", "err", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			logger.Info("test server received data", "body", string(prettyBody))

			w.Header().Set("Location", r.URL.RequestURI())
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		cipUrl = server.URL
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
