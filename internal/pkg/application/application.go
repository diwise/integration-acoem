package application

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

func Run(baseUrl, accountID, accountKey, interval string, log zerolog.Logger) error {
	if accountID == "" || accountKey == "" {
		log.Error().Msg("account id and account key must not be empty")
	}

	parsedInterval, err := strconv.Atoi(interval)
	if err != nil {
		log.Err(err).Msg("failed to parse string interval to integer")
		return err
	}

	for {
		stations, err := GetStations(baseUrl, accountID, accountKey)
		if err != nil {
			log.Err(err).Msg("failed to retrieve stations")
			return err
		}

		for _, stn := range stations {
			result, err := GetSensorData(baseUrl, accountID, accountKey, stn)
			if err != nil {
				log.Err(err).Msg("failed to retrieve sensor data")
				return err
			}

			stnBytes, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				log.Err(err).Msg("failed to marshal data into structs")
				return err
			}

			fmt.Printf("latest data from %s station: %s", stn.StationName, string(stnBytes))
		}

		time.Sleep(time.Duration(parsedInterval) * time.Second)
	}
}
