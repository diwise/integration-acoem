package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/diwise/context-broker/pkg/datamodels/fiware"
	"github.com/diwise/context-broker/pkg/ngsild/client"
	ngsierrors "github.com/diwise/context-broker/pkg/ngsild/errors"
	"github.com/diwise/context-broker/pkg/ngsild/types/entities"
	. "github.com/diwise/context-broker/pkg/ngsild/types/entities/decorators"
	"github.com/diwise/context-broker/pkg/ngsild/types/properties"
	"github.com/diwise/integration-acoem/domain"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
)

type IntegrationAcoem interface {
	CreateAirQualityObserved(ctx context.Context) error
}

type integrationAcoem struct {
	baseUrl     string
	accessToken string
	cb          client.ContextBrokerClient
}

func New(ctx context.Context, baseUrl, accessToken string, cb client.ContextBrokerClient) IntegrationAcoem {
	return &integrationAcoem{
		baseUrl:     baseUrl,
		accessToken: accessToken,
		cb:          cb,
	}
}

func (i *integrationAcoem) CreateAirQualityObserved(ctx context.Context) error {
	logger := logging.GetFromContext(ctx)

	headers := map[string][]string{"Content-Type": {"application/ld+json"}}

	devices, err := i.getDevices()
	if err != nil {
		logger.Error().Err(err).Msg("failed to retrieve devices")
		return err
	}

	for _, dev := range devices {
		sensors, err := i.getDeviceData(dev)
		if err != nil {
			logger.Error().Err(err).Msg("failed to retrieve sensor data")
			return err
		}

		decorators := []entities.EntityDecoratorFunc{}

		decorators = append(decorators, entities.DefaultContext(), Text("areaServed", dev.DeviceName))

		for _, sensor := range sensors {
			decorators = append(decorators,
				Location(sensor.Latitude, sensor.Longitude),
				DateTime(properties.DateObserved, sensor.TBTimestamp),
			)

			sensorReadings := createFragmentsFromSensorData(sensor.Channels, sensor.TBTimestamp)

			decorators = append(decorators, sensorReadings...)
		}

		fragment, err := entities.NewFragment(decorators...)
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

			entity, err := entities.New(entityID, fiware.AirQualityObservedTypeName, decorators...)
			if err != nil {
				logger.Error().Err(err).Msg("failed to create new entity")
				continue
			}

			_, err = i.cb.CreateEntity(ctx, entity, headers)
			if err != nil {
				logger.Error().Err(err).Msg("failed to post entity to context broker")
				continue
			}
		}
	}

	return nil
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
				sensor.Scaled,
				properties.UnitCode(unitCodes[sensor.UnitName]),
				properties.ObservedAt(timestamp),
			))
		}
	}

	return readings
}

func (i *integrationAcoem) getDeviceData(device domain.Device) ([]domain.DeviceData, error) {
	if device.UniqueId == 0 || device.DeviceName == "" {
		return nil, fmt.Errorf("cannot retrieve sensor data as no valid station ID has been provided")
	}
	deviceData := []domain.DeviceData{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/devices/setup/%d", i.baseUrl, device.UniqueId), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err.Error())
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sensor data: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body as bytes: %s", err.Error())
	}

	err = json.Unmarshal(respBytes, &deviceData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json data: %s", err.Error())
	}

	return deviceData, nil
}

func (i *integrationAcoem) getDevices() ([]domain.Device, error) {
	devices := []domain.Device{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/devices", i.baseUrl), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err.Error())
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)

	client := http.DefaultClient

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve list of devices: %s", err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body as bytes: %s", err.Error())
	}

	err = json.Unmarshal(responseBytes, &devices)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %s,\ndue to: %s", string(responseBytes), err.Error())
	}

	return devices, nil
}
