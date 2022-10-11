package application

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwise/integration-acoem/domain"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
)

func TestThatGetStationsFailsIfResponseCodeIsNotOK(t *testing.T) {
	is, mockApp := testSetup(t, http.StatusNotFound, "")

	stn, err := mockApp.getStations()
	is.True(err != nil)
	is.True(stn == nil)
}

func TestThatGetStationsFailsIfReturnedStationDataIsIncorrect(t *testing.T) {
	is, mockApp := testSetup(t, http.StatusOK, stationsBadResponse)
	stn, err := mockApp.getStations()

	is.True(err != nil)
	is.True(stn == nil)
}

func TestGetSensorDataFailsOnEmptyStationData(t *testing.T) {
	is, mockApp := testSetup(t, http.StatusOK, "")

	_, err := mockApp.getSensorData(domain.Station{})
	is.True(err != nil)
}

func TestThatGetSensorDataFailsIfResponseCodeIsNotOK(t *testing.T) {
	is, mockApp := testSetup(t, http.StatusNotFound, "")

	stn := domain.Station{
		UniqueId:    123,
		StationName: "abc",
	}

	result, err := mockApp.getSensorData(stn)
	is.True(err != nil)
	is.True(result == nil)
}

func TestThatGetSensorDataReturnsAndMarshalsCorrectly(t *testing.T) {
	is, mockApp := testSetup(t, http.StatusOK, acoemResponse)
	stn := domain.Station{
		UniqueId:    123,
		StationName: "abc",
	}

	result, err := mockApp.getSensorData(stn)
	is.NoErr(err)

	_, err = json.MarshalIndent(result, "", "  ")
	is.NoErr(err)
}

func testSetup(t *testing.T, responseCode int, responseBody string) (*is.I, *integrationAcoem) {
	is := is.New(t)
	log := zerolog.Logger{}
	server := setupMockService(responseCode, responseBody)
	app := New(server.URL, "notarealID", "notarealkey", log, nil)
	mockApp := app.(*integrationAcoem)

	return is, mockApp
}

func setupMockService(responseCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/ld+json")
		w.WriteHeader(responseCode)
		w.Write([]byte(responseBody))
	}))
}

const stationsResponse string = `
[{"UniqueId":888100,"StationType":"Gen2 Logger","StationName":"SUNDSVALL GEN2","SerialNumber":1336,"Firmware":null,"Imsi":null,"Latitude":62.388618,"Longitude":17.308968,"Altitude":null,"CustomerId":"CSUN105032030469"},{"UniqueId":1098100,"StationType":"Mini Gateway","StationName":"SUNDSVALL BERGSGATAN","SerialNumber":null,"Firmware":null,"Imsi":"089462048008002994526","Latitude":62.386485,"Longitude":17.303442,"Altitude":null,"CustomerId":"CSUN105032030469"}]
`

const stationsBadResponse string = `
[{"UniqueId":888100,"StationType":"Gen2 Logger","StationName":"SUNDSVALL GEN2","SerialNumber":1336,"Firmware":null,"Imsi":null,"Latitude":62.388618,"Longitude":17.308968,"Altitude":null,"CustomerId":"CSUN105032030469"}{"UniqueId":1098100,"StationType":"Mini Gateway","StationName":"SUNDSVALL BERGSGATAN","SerialNumber":null,"Firmware":null,"Imsi":"089462048008002994526","Latitude":62.386485,"Longitude":17.303442,"Altitude":null,"CustomerId":"CSUN105032030469"}]
`

const acoemResponse string = `
[{"TBTimestamp":"2021-10-12T11:15:00+00:00","TETimestamp":"2021-10-12T11:16:00+00:00","Latitude":62.388618,"Longitude":17.308968,"Altitude":null,"Channels":[{"SensorName":"Air Pressure","SensorLabel":"AIRPRES","Channel":9,"PreScaled":1008,"Scaled":1008,"UnitName":"Pressure (mbar)","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Humidity","SensorLabel":"HUM","Channel":8,"PreScaled":74.32,"Scaled":74.32,"UnitName":"Percent","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Nitric Oxide","SensorLabel":"NO","Channel":10,"PreScaled":2.679,"Scaled":2.679,"UnitName":"Parts Per Billion","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Nitrogen Dioxide","SensorLabel":"NO2","Channel":11,"PreScaled":4.126,"Scaled":4.126,"UnitName":"Parts Per Billion","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Nitrogen Oxides","SensorLabel":"NOx","Channel":12,"PreScaled":6805,"Scaled":6.805,"UnitName":"Parts Per Billion","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Particle Count","SensorLabel":"PARTICLE_COUNT","Channel":6,"PreScaled":8.142,"Scaled":8.142,"UnitName":"Particles per cm3","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Particulate Matter (PM 1)","SensorLabel":"PM1","Channel":1,"PreScaled":0.284,"Scaled":0.284,"UnitName":"Micrograms Per Cubic Meter","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Particulate Matter (PM 10)","SensorLabel":"PM10","Channel":4,"PreScaled":4.747,"Scaled":4.747,"UnitName":"Micrograms Per Cubic Meter","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Particulate Matter (PM 2.5)","SensorLabel":"PM2.5","Channel":2,"PreScaled":0.893,"Scaled":0.893,"UnitName":"Micrograms Per Cubic Meter","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"PM 4","SensorLabel":"PM4","Channel":3,"PreScaled":1.478,"Scaled":1.478,"UnitName":"Micrograms Per Cubic Meter","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Temperature","SensorLabel":"TEMP","Channel":7,"PreScaled":9.273,"Scaled":9.273,"UnitName":"Celsius","Slope":1,"Offset":0,"Flags":["Valid"]},{"SensorName":"Total Suspended Particulate","SensorLabel":"TSP","Channel":5,"PreScaled":11.29,"Scaled":11.29,"UnitName":"Micrograms Per Cubic Meter","Slope":1,"Offset":0,"Flags":["Valid"]}]}]
`
