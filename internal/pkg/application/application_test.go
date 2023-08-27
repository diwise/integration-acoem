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

	_, err := mockApp.getDeviceData(domain.Device{})
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

	result, err := mockApp.getDeviceData(dev)
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
			response.Body([]byte(acoemResponse)),
		),
	)
	mockApp := newMockApp(t, s.URL())
	dev := domain.Device{
		UniqueId:   123,
		DeviceName: "abc",
	}

	result, err := mockApp.getDeviceData(dev)
	is.NoErr(err)

	_, err = json.MarshalIndent(result, "", "  ")
	is.NoErr(err)
}

func newMockApp(t *testing.T, serverURL string) *integrationAcoem {
	app := New(context.Background(), serverURL, "notreallyanaccesstoken", nil)
	mockApp := app.(*integrationAcoem)

	return mockApp
}

const devicesBadResponse string = `
[{"UniqueId":888100,"StationType":"Gen2 Logger","StationName":"SUNDSVALL GEN2","SerialNumber":1336,"Firmware":null,"Imsi":null,"Latitude":62.388618,"Longitude":17.308968,"Altitude":null,"CustomerId":"CSUN105032030469"}{"UniqueId":1098100,"StationType":"Mini Gateway","StationName":"SUNDSVALL BERGSGATAN","SerialNumber":null,"Firmware":null,"Imsi":"089462048008002994526","Latitude":62.386485,"Longitude":17.303442,"Altitude":null,"CustomerId":"CSUN105032030469"}]
`

const acoemResponse string = `
[
  {
    "Active": true,
    "Channel": 9,
    "Rate": 60,
    "SensorLabel": "AIRPRES",
    "SensorName": "Air Pressure",
    "Type": "data",
    "Unit": "mbar",
    "UnitName": "Pressure (mbar)"
  },
  {
    "Active": true,
    "Channel": 8,
    "Rate": 60,
    "SensorLabel": "HUM",
    "SensorName": "Humidity",
    "Type": "data",
    "Unit": "%",
    "UnitName": "Percent"
  },
  {
    "Active": true,
    "Channel": 10,
    "Rate": 60,
    "SensorLabel": "NO",
    "SensorName": "Nitric Oxide",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": false,
    "Channel": 30,
    "Rate": 900,
    "SensorLabel": "NO",
    "SensorName": "Nitric Oxide",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": true,
    "Channel": 11,
    "Rate": 60,
    "SensorLabel": "NO2",
    "SensorName": "Nitrogen Dioxide",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": false,
    "Channel": 12,
    "Rate": 60,
    "SensorLabel": "NO2",
    "SensorName": "Nitrogen Dioxide",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": false,
    "Channel": 13,
    "Rate": 60,
    "SensorLabel": "NO2",
    "SensorName": "Nitrogen Dioxide",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": false,
    "Channel": 31,
    "Rate": 900,
    "SensorLabel": "NO2",
    "SensorName": "Nitrogen Dioxide",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": false,
    "Channel": 11,
    "Rate": 60,
    "SensorLabel": "NOx",
    "SensorName": "Nitrogen Oxides",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": true,
    "Channel": 12,
    "Rate": 60,
    "SensorLabel": "NOx",
    "SensorName": "Nitrogen Oxides",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": false,
    "Channel": 32,
    "Rate": 900,
    "SensorLabel": "NOx",
    "SensorName": "Nitrogen Oxides",
    "Type": "data",
    "Unit": "ppb",
    "UnitName": "Parts Per Billion"
  },
  {
    "Active": true,
    "Channel": 6,
    "Rate": 60,
    "SensorLabel": "PARTICLE_COUNT",
    "SensorName": "Particle Count",
    "Type": "data",
    "Unit": "P/cm3",
    "UnitName": "Particles per cm3"
  },
  {
    "Active": true,
    "Channel": 1,
    "Rate": 60,
    "SensorLabel": "PM1",
    "SensorName": "Particulate Matter (PM 1)",
    "Type": "data",
    "Unit": "ug/m3",
    "UnitName": "Micrograms Per Cubic Meter"
  },
  {
    "Active": true,
    "Channel": 4,
    "Rate": 60,
    "SensorLabel": "PM10",
    "SensorName": "Particulate Matter (PM 10)",
    "Type": "data",
    "Unit": "ug/m3",
    "UnitName": "Micrograms Per Cubic Meter"
  },
  {
    "Active": true,
    "Channel": 2,
    "Rate": 60,
    "SensorLabel": "PM2.5",
    "SensorName": "Particulate Matter (PM 2.5)",
    "Type": "data",
    "Unit": "ug/m3",
    "UnitName": "Micrograms Per Cubic Meter"
  },
  {
    "Active": true,
    "Channel": 3,
    "Rate": 60,
    "SensorLabel": "PM4",
    "SensorName": "PM 4",
    "Type": "data",
    "Unit": "ug/m3",
    "UnitName": "Micrograms Per Cubic Meter"
  },
  {
    "Active": true,
    "Channel": 7,
    "Rate": 60,
    "SensorLabel": "TEMP",
    "SensorName": "Temperature",
    "Type": "data",
    "Unit": "C",
    "UnitName": "Celsius"
  },
  {
    "Active": true,
    "Channel": 5,
    "Rate": 60,
    "SensorLabel": "TSP",
    "SensorName": "Total Suspended Particulate",
    "Type": "data",
    "Unit": "ug/m3",
    "UnitName": "Micrograms Per Cubic Meter"
  }
]
`
