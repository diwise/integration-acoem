package application

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/diwise/integration-acoem/domain"
)

func GetSensorData(baseUrl, accountID, accountKey string, station domain.Station) ([]domain.StationData, error) {
	if station.UniqueId == 0 || station.StationName == "" {
		return nil, fmt.Errorf("cannot retrieve sensor data as no valid station ID has been provided")
	}
	stationData := []domain.StationData{}

	resp, err := http.Get(fmt.Sprintf("%s/3.5/GET/%s/%s/stationdata/latest/2/%d", baseUrl, accountID, accountKey, station.UniqueId))
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
