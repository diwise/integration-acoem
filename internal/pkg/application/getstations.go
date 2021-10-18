package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/diwise/integration-acoem/domain"
)

func GetStations(baseUrl, accountID, accountKey string) ([]domain.Station, error) {
	stations := []domain.Station{}

	response, err := http.Get(fmt.Sprintf("%s/3.5/GET/%s/%s/stations", baseUrl, accountID, accountKey))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve list of stations: %s", err.Error())
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, response.StatusCode)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body as bytes: %s", err.Error())
	}

	err = json.Unmarshal(responseBytes, &stations)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %s,\ndue to: %s", string(responseBytes), err.Error())
	}

	return stations, nil
}
