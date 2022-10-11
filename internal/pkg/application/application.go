package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/diwise/context-broker/pkg/datamodels/fiware"
	"github.com/diwise/context-broker/pkg/ngsild/client"
	"github.com/diwise/context-broker/pkg/ngsild/types/entities"
	. "github.com/diwise/context-broker/pkg/ngsild/types/entities/decorators"
	"github.com/diwise/context-broker/pkg/ngsild/types/properties"
	"github.com/diwise/integration-acoem/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type IntegrationAcoem interface {
	CreateAirQualityObserved(ctx context.Context) error
}

type integrationAcoem struct {
	baseUrl    string
	accountID  string
	accountKey string
	log        zerolog.Logger
	cb         client.ContextBrokerClient
}

func New(baseUrl, accountID, accountKey string, log zerolog.Logger, cb client.ContextBrokerClient) IntegrationAcoem {
	return &integrationAcoem{
		baseUrl:    baseUrl,
		accountID:  accountID,
		accountKey: accountKey,
		log:        log,
		cb:         cb,
	}
}

func (i *integrationAcoem) CreateAirQualityObserved(ctx context.Context) error {
	headers := map[string][]string{"Content-Type": {"application/ld+json"}}

	var err error
	if i.accountID == "" || i.accountKey == "" {
		log.Error().Err(err).Msg("account id and account key must not be empty")
	}

	stations, err := i.getStations()
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve stations")
		return err
	}

	for _, stn := range stations {
		sensors, err := i.getSensorData(stn)
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve sensor data")
			return err
		}

		id := fiware.AirQualityObservedIDPrefix + strconv.Itoa(stn.UniqueId)

		decorators := []entities.EntityDecoratorFunc{}

		for _, sensor := range sensors {

			observed, err := time.Parse(time.RFC3339, sensor.TBTimestamp)
			if err != nil {
				observed = time.Time{}
			}

			decorators = append(decorators,
				Location(sensor.Latitude, sensor.Longitude),
				DateTimeIfNotZero(properties.DateObserved, observed),
			)

			sensorReadings := createFragmentsFromSensorData(sensor.Channels)

			decorators = append(decorators, sensorReadings...)
		}

		entity, err := entities.New(id, fiware.AirQualityObservedTypeName, decorators...)
		if err != nil {
			log.Error().Err(err).Msg("failed to create new entity")
		}

		_, err = i.cb.CreateEntity(ctx, entity, headers)
		if err != nil {
			log.Error().Err(err).Msg("failed to post entity to context broker")
		}

	}

	return nil
}

func createFragmentsFromSensorData(sensors []domain.Channel) []entities.EntityDecoratorFunc {
	readings := []entities.EntityDecoratorFunc{}

	for _, sensor := range sensors {
		readings = append(readings,
			Number(strings.ToLower(sensor.SensorLabel), sensor.Scaled, properties.UnitCode(sensor.UnitName)),
		)
	}

	return readings
}

func (i *integrationAcoem) getSensorData(station domain.Station) ([]domain.StationData, error) {
	if station.UniqueId == 0 || station.StationName == "" {
		return nil, fmt.Errorf("cannot retrieve sensor data as no valid station ID has been provided")
	}
	stationData := []domain.StationData{}

	resp, err := http.Get(fmt.Sprintf("%s/3.5/GET/%s/%s/stationdata/latest/2/%d", i.baseUrl, i.accountID, i.accountKey, station.UniqueId))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve list of stations: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body as bytes: %s", err.Error())
	}

	err = json.Unmarshal(respBytes, &stationData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json data: %s", err.Error())
	}

	return stationData, nil
}

func (i *integrationAcoem) getStations() ([]domain.Station, error) {
	stations := []domain.Station{}

	response, err := http.Get(fmt.Sprintf("%s/3.5/GET/%s/%s/stations", i.baseUrl, i.accountID, i.accountKey))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve list of stations: %s", err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body as bytes: %s", err.Error())
	}

	err = json.Unmarshal(responseBytes, &stations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %s,\ndue to: %s", string(responseBytes), err.Error())
	}

	return stations, nil
}
