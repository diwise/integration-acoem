package application

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

func TestThatGetDevicesgetDevicesFailsIfResponseCodeIsNotOK(t *testing.T) {
	is := is.New(t)

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodGet),
		),
		Returns(
			response.Code(http.StatusNotFound),
			response.Body([]byte("")),
		),
	)

	mockApp := newMockApp(t, s.URL())

	stn, err := mockApp.getDevices()
	is.True(err != nil)
	is.True(stn == nil)
}

func TestThatGetDevicesFailsIfReturnedStationDataIsIncorrect(t *testing.T) {
	is := is.New(t)

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodGet),
		),
		Returns(
			response.Code(http.StatusOK),
			response.Body([]byte(devicesBadResponse)),
		),
	)

	mockApp := newMockApp(t, s.URL())
	dev, err := mockApp.getDevices()

	is.True(err != nil)
	is.True(dev == nil)
}

func TestGetDeviceDataFailsOnEmptyStationData(t *testing.T) {
	is := is.New(t)

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodGet),
		),
		Returns(
			response.Code(http.StatusOK),
			response.Body([]byte("")),
		),
	)
	mockApp := newMockApp(t, s.URL())

	_, err := mockApp.getDeviceData(domain.Device{}, "s")
	is.True(err != nil)
}

func TestThatGetDeviceDataFailsIfResponseCodeIsNotOK(t *testing.T) {
	is := is.New(t)

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodGet),
		),
		Returns(
			response.Code(http.StatusNotFound),
			response.Body([]byte("")),
		),
	)
	mockApp := newMockApp(t, s.URL())

	dev := domain.Device{
		UniqueId:   123,
		DeviceName: "abc",
	}

	result, err := mockApp.getDeviceData(dev, "")
	is.True(err != nil)
	is.True(result == nil)
}

func TestThatGetSensorDataReturnsAndMarshalsCorrectly(t *testing.T) {
	is := is.New(t)

	s := testutils.NewMockServiceThat(
		Expects(
			is,
			method(http.MethodGet),
		),
		Returns(
			response.Code(http.StatusOK),
			response.Body([]byte(deviceDataResponse)),
		),
	)
	mockApp := newMockApp(t, s.URL())
	dev := domain.Device{
		UniqueId:   123,
		DeviceName: "abc",
	}

	result, err := mockApp.getDeviceData(dev, "$NO2+NOX")
	is.NoErr(err)

	_, err = json.MarshalIndent(result, "", "  ")
	is.NoErr(err)
}

func newMockApp(t *testing.T, serverURL string) *integrationAcoem {
	app := New(context.Background(), serverURL, "notreallyanaccesstoken", nil)
	mockApp := app.(*integrationAcoem)

	return mockApp
}

const devicesBadResponse string = `[
	{
	  "Altitude": null,
	  "Customer": "Sundsvall",
	  "DeviceName": "Sundsvall Gen2",
	  "DeviceType": "Gen2 Logger",
	  "Firmware": "1.138",
	  "Imsi": null,
	  "LastConnection": "2023-08-28T00:23:42+00:00",
	  "Latitude": 62.388618,
	  "Longitude": 17.308968,
	  "SerialNumber": 1336,
	  "UniqueId": 888100
	}
	{
	  "Altitude": null,
	  "Customer": "Sundsvall",
	  "DeviceName": "Sundsvall Bergsgatan",
	  "DeviceType": "Mini Gateway",
	  "Firmware": "1.00",
	  "Imsi": 89462048008003000000,
	  "LastConnection": "2023-08-28T00:23:14+00:00",
	  "Latitude": 62.386485,
	  "Longitude": 17.303442,
	  "SerialNumber": 105,
	  "UniqueId": 1098100
	}
  ]`

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
