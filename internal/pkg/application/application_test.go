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

func TestThatGetDevicesFailsIfResponseCodeIsNotOK(t *testing.T) {
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

	stn, err := mockApp.GetDevices(context.Background())
	is.True(err != nil)
	is.True(stn == nil)
}

func TestThatGetDevicesFailsIfReturnedDeviceDataIsIncorrect(t *testing.T) {
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
	dev, err := mockApp.GetDevices(context.Background())

	is.True(err != nil)
	is.True(dev == nil)
}

func TestGetDeviceDataFailsOnEmptyDeviceData(t *testing.T) {
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

	_, err := mockApp.GetDeviceData(context.Background(), 123, "s")
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

	result, err := mockApp.GetDeviceData(context.Background(), dev.UniqueId, "")
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

	result, err := mockApp.GetDeviceData(context.Background(), dev.UniqueId, "NO2+NOX")
	is.NoErr(err)

	data, err := json.Marshal(result)
	is.NoErr(err)

	expectation := `[{"timestamp":{"convention":"TimeBeginning","timestamp":"2023-08-27T22:08:00+00:00"},"location":{"altitude":0,"longitude":17.308968,"latitude":62.388618},"channels":[{"sensorName":"Nitrogen Dioxide","sensorLabel":"NO2","channel":11,"preScaled":{"reading":3.888},"scaled":{"reading":3.888},"unitName":"Parts Per Billion","slope":1,"offset":0,"flags":null},{"sensorName":"Nitrogen Oxides","sensorLabel":"NOx","channel":12,"preScaled":{"reading":5.421},"scaled":{"reading":5.421},"unitName":"Parts Per Billion","slope":1,"offset":0,"flags":null}]}]`
	is.Equal(expectation, string(data))
}

func newMockApp(t *testing.T, serverURL string) *integrationAcoem {
	app := New(serverURL, "user", "pass", nil)
	mockApp := app.(*integrationAcoem)

	return mockApp
}

const devicesBadResponse string = `[
	{
	  "Altitude": null,
	  "Customer": "Sundsvall",
	  "DeviceName": "/////",
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
