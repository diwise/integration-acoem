package lwm2m

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/diwise/integration-acoem/domain"
	testutils "github.com/diwise/service-chassis/pkg/test/http"
	"github.com/diwise/service-chassis/pkg/test/http/expects"
	"github.com/diwise/service-chassis/pkg/test/http/response"
	"github.com/matryer/is"
)

var Expects = testutils.Expects
var Returns = testutils.Returns
var method = expects.RequestMethod

func TestSendingTemperaturePack(t *testing.T) {
	is := is.New(t)
	devices := []domain.DeviceData{}

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodPost),
		),
		Returns(
			response.Code(http.StatusCreated),
			response.Body([]byte("")),
		),
	)

	err := json.Unmarshal([]byte(deviceDataResponse), &devices)
	is.NoErr(err)

	err = CreateAndSendAsLWM2M(context.Background(), devices, 888100, "temperature", s.URL())
	is.NoErr(err)
}

func TestSendingAirQualityPack(t *testing.T) {
	is := is.New(t)
	devices := []domain.DeviceData{}

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodPost),
		),
		Returns(
			response.Code(http.StatusCreated),
			response.Body([]byte("")),
		),
	)

	err := json.Unmarshal([]byte(deviceDataResponse), &devices)
	is.NoErr(err)

	err = CreateAndSendAsLWM2M(context.Background(), devices, 888100, "NO2", s.URL())
	is.NoErr(err)
}

const deviceDataResponse string = `
[
  {
    "Channels": [
      {
        "Channel": 11,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 3.888,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 3.888,
          "ValidPercentage": 100
        },
        "SensorLabel": "NO2",
        "SensorName": "Nitrogen Dioxide",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Parts Per Billion"
      },
      {
        "Channel": 12,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 5.421,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 5.421,
          "ValidPercentage": 100
        },
        "SensorLabel": "NOx",
        "SensorName": "Nitrogen Oxides",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Parts Per Billion"
      },
	  {
        "Channel": 12,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 5.421,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 5.421,
          "ValidPercentage": 100
        },
        "SensorLabel": "",
        "SensorName": "Temperature",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Parts Per Billion"
      }
    ],
    "Location": {
      "Altitude": null,
      "Latitude": 62.388618,
      "Longitude": 17.308968
    },
    "Timestamp": {
      "Convention": "TimeBeginning",
      "Timestamp": "2023-08-27T22:08:00+00:00"
    }
  }
]
`
