package application

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/diwise/context-broker/pkg/ngsild/client"
	"github.com/diwise/integration-acoem/domain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type IntegrationAcoem interface {
	CreateAirQualityObserved() error
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

func (i *integrationAcoem) CreateAirQualityObserved() error {
	stations, err := i.getData()
	if err != nil {
		return err
	}

	//this function will map station data to fiware air quality observeds

	bytes, _ := json.Marshal(stations)

	fmt.Printf("stations %s", string(bytes))

	return nil
}

func (i *integrationAcoem) getData() ([]domain.Station, error) {
	if i.accountID == "" || i.accountKey == "" {
		log.Error().Msg("account id and account key must not be empty")
	}

	stations, err := i.getStations()
	if err != nil {
		log.Err(err).Msg("failed to retrieve stations")
		return nil, err
	}

	for _, stn := range stations {
		result, err := i.getSensorData(stn)
		if err != nil {
			log.Err(err).Msg("failed to retrieve sensor data")
			return nil, err
		}

		stn.StationData = append(stn.StationData, result...)
	}

	return stations, nil
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
