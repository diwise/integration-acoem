package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/diwise/context-broker/pkg/datamodels/fiware"
	"github.com/diwise/context-broker/pkg/ngsild/client"
	ngsierrors "github.com/diwise/context-broker/pkg/ngsild/errors"
	"github.com/diwise/context-broker/pkg/ngsild/types"
	"github.com/diwise/context-broker/pkg/ngsild/types/entities"
	. "github.com/diwise/context-broker/pkg/ngsild/types/entities/decorators"
	"github.com/diwise/context-broker/pkg/ngsild/types/properties"
	"github.com/diwise/integration-acoem/domain"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type IntegrationAcoem interface {
	CreateAirQualityObserved(ctx context.Context) error
}

type integrationAcoem struct {
	baseUrl     string
	accessToken string
	cb          client.ContextBrokerClient
}

var tracer = otel.Tracer("integration-acoem/app")

func New(ctx context.Context, baseUrl, accountID, accountKey string, cb client.ContextBrokerClient) IntegrationAcoem {
	accessToken := fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", accountID, accountKey)),
		),
	)

	return &integrationAcoem{
		baseUrl:     baseUrl,
		accessToken: accessToken,
		cb:          cb,
	}
}

func (i *integrationAcoem) CreateAirQualityObserved(ctx context.Context) error {
	var err error

	ctx, span := tracer.Start(ctx, "create-air-qualities")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	_, ctx, logger := o11y.AddTraceIDToLoggerAndStoreInContext(span, logging.GetFromContext(ctx), ctx)

	headers := map[string][]string{"Content-Type": {"application/ld+json"}}

	var devices []domain.Device

	devices, err = i.getDevices(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve devices")
		return err
	}

	logger.Info().Msgf("retrieved %d devices from acoem", len(devices))

	var sensorLabels string
	var sensors []domain.DeviceData

	for _, dev := range devices {

		sensorLabels, err = i.getSensorLabels(ctx, dev.UniqueId)
		if err != nil {
			logger.Error().Err(err).Msgf("failed to retrieve sensor labels for device %d", dev.UniqueId)
			return err
		}

		logger.Info().Msgf("retrieving data for %s from %d", sensorLabels, dev.UniqueId)

		sensors, err = i.getDeviceData(ctx, dev, sensorLabels)
		if err != nil {
			logger.Error().Err(err).Msg("failed to retrieve sensor data")
			return err
		}

		decorators := []entities.EntityDecoratorFunc{}

		decorators = append(decorators, entities.DefaultContext(), Text("areaServed", dev.DeviceName))

		for _, sensor := range sensors {
			decorators = append(decorators,
				Location(sensor.Location.Latitude, sensor.Location.Longitude),
				DateTime(properties.DateObserved, sensor.Timestamp.Timestamp),
			)

			sensorReadings := createFragmentsFromSensorData(sensor.Channels, sensor.Timestamp.Timestamp)

			decorators = append(decorators, sensorReadings...)
		}

		var fragment types.EntityFragment
		fragment, err = entities.NewFragment(decorators...)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create entity fragments")
		}

		entityID := fiware.AirQualityObservedIDPrefix + strconv.Itoa(dev.UniqueId)

		_, err = i.cb.MergeEntity(ctx, entityID, fragment, headers)
		if err != nil {
			if !errors.Is(err, ngsierrors.ErrNotFound) {
				logger.Error().Err(err).Msg("failed to merge entity")
				continue
			}

			var entity types.Entity
			entity, err = entities.New(entityID, fiware.AirQualityObservedTypeName, decorators...)
			if err != nil {
				logger.Error().Err(err).Msg("failed to create new entity")
				continue
			}

			_, err = i.cb.CreateEntity(ctx, entity, headers)
			if err != nil {
				logger.Error().Err(err).Msg("failed to post entity to context broker")
				continue
			}

			logger.Info().Msgf("created entity %s", entityID)
		} else {
			logger.Info().Msgf("updated entity %s", entityID)
		}
	}

	return nil
}

func (i *integrationAcoem) getSensorLabels(ctx context.Context, deviceID int) (string, error) {
	var err error

	ctx, span := tracer.Start(ctx, "get-sensor-labels")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/devices/setup/%d", i.baseUrl, deviceID), nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %s", err.Error())
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)

	var resp *http.Response
	resp, err = httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request failed: %s", err.Error())
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed, expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		return "", err
	}

	sensors := []domain.Sensor{}

	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %s", err.Error())
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &sensors)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal response body: %s", err.Error())
		return "", err
	}

	sensorLabels := []string{}

	for _, s := range sensors {
		if s.Active && s.Type == "data" {
			sensorLabels = append(sensorLabels, s.SensorLabel)
		}
	}

	labels := strings.Join(sensorLabels, "+")

	return labels, nil
}

var unitCodes map[string]string = map[string]string{
	"Micrograms Per Cubic Meter": "GQ",
	"Volts":                      "VLT",
	"Celsius":                    "CEL",
	"Percent":                    "P1",
	"Hectopascals":               "A97",
	"Parts Per Billion":          "61",
	"Pressure (mbar)":            "MBR",
}

var sensorNames map[string]string = map[string]string{
	"Humidity":                    "relativeHumidity",
	"Temperature":                 "temperature",
	"Air Pressure":                "atmosphericPressure",
	"Particulate Matter (PM 1)":   "PM1",
	"PM 4":                        "PM4",
	"Particulate Matter (PM 10)":  "PM10",
	"Particulate Matter (PM 2.5)": "PM25",
	"Total Suspended Particulate": "totalSuspendedParticulate",
	"Voltage":                     "voltage",
	"Nitric Oxide":                "NO",
	"Nitrogen Dioxide":            "NO2",
	"Nitrogen Oxides":             "NOx",
}

func createFragmentsFromSensorData(sensors []domain.Channel, timestamp string) []entities.EntityDecoratorFunc {
	readings := []entities.EntityDecoratorFunc{}

	for _, sensor := range sensors {
		name, ok := sensorNames[sensor.SensorName]
		if ok {
			readings = append(readings, Number(
				name,
				sensor.Scaled.Reading,
				properties.UnitCode(unitCodes[sensor.UnitName]),
				properties.ObservedAt(timestamp),
			))
		}
	}

	return readings
}

func (i *integrationAcoem) getDeviceData(ctx context.Context, device domain.Device, sensorLabels string) ([]domain.DeviceData, error) {
	var err error

	ctx, span := tracer.Start(ctx, "get-device-data")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	if device.UniqueId == 0 || device.DeviceName == "" || sensorLabels == "" {
		err = fmt.Errorf("cannot retrieve sensor data as either station ID, device name, or sensor labels are empty")
		return nil, err
	}
	deviceData := []domain.DeviceData{}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/devicedata/%d/latest/1/300/data/%s", i.baseUrl, device.UniqueId, sensorLabels), nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %s", err.Error())
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)
	req.Header.Add("TimeConvention", "TimeBeginning")

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to retrieve sensor data: %s", err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		return nil, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body as bytes: %s", err.Error())
		return nil, err
	}

	err = json.Unmarshal(respBytes, &deviceData)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal json data: %s", err.Error())
		return nil, err
	}

	return deviceData, nil
}

func (i *integrationAcoem) getDevices(ctx context.Context) ([]domain.Device, error) {
	var err error

	ctx, span := tracer.Start(ctx, "get-devices")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	devices := []domain.Device{}

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/devices", i.baseUrl), nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %s", err.Error())
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)

	var response *http.Response
	response, err = httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to retrieve list of devices: %s", err.Error())
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, response.StatusCode)
		return nil, err
	}

	var responseBytes []byte
	responseBytes, err = io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body as bytes: %s", err.Error())
		return nil, err
	}

	err = json.Unmarshal(responseBytes, &devices)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal response: %s,\ndue to: %s", string(responseBytes), err.Error())
		return nil, err
	}

	return devices, nil
}
