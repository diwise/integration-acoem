package lwm2m

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/diwise/integration-acoem/domain"
	"github.com/diwise/senml"
	"github.com/matryer/is"
)

func TestCreateAndSendAsLWM2M(t *testing.T) {
	is := is.New(t)

	var deviceData []domain.DeviceData
	json.Unmarshal([]byte(devicedataJson), &deviceData)

	packs := make([]senml.Pack, 0)

	err := CreateAndSendAsLWM2M(context.Background(), deviceData, 11111, "/url", func(ctx context.Context, s string, p senml.Pack) error {
		packs = append(packs, p)
		return nil
	})

	is.NoErr(err)
	is.Equal(3, len(packs))

	clone := packs[0].Clone()
	rec, ok := clone.GetRecord(senml.FindByName("0"))
	is.True(ok)

	is.Equal("urn:oma:lwm2m:ext:3304", rec.StringValue)
	is.Equal("11111/3304/0", rec.Name)
}

const devicedataJson string = `
[
  {
    "Channels": [
      {
        "Channel": 8,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 90.68,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 90.68,
          "ValidPercentage": 100
        },
        "SensorLabel": "HUM",
        "SensorName": "Humidity",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Percent"
      },
      {
        "Channel": 1,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 7.901,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 7.901,
          "ValidPercentage": 100
        },
        "SensorLabel": "PM1",
        "SensorName": "Particulate Matter (PM 1)",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Micrograms Per Cubic Meter"
      },
      {
        "Channel": 4,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 20.548,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 20.548,
          "ValidPercentage": 100
        },
        "SensorLabel": "PM10",
        "SensorName": "Particulate Matter (PM 10)",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Micrograms Per Cubic Meter"
      },
      {
        "Channel": 2,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 10.71,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 10.71,
          "ValidPercentage": 100
        },
        "SensorLabel": "PM2.5",
        "SensorName": "Particulate Matter (PM 2.5)",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Micrograms Per Cubic Meter"
      },
      {
        "Channel": 7,
        "DataRate": 60,
        "Offset": 0,
        "PreScaled": {
          "Flags": null,
          "Reading": 16.99,
          "ValidPercentage": 100
        },
        "RedactedPercentage": 0,
        "Scaled": {
          "Flags": null,
          "Reading": 16.99,
          "ValidPercentage": 100
        },
        "SensorLabel": "TEMP",
        "SensorName": "Temperature",
        "Slope": 1,
        "UniqueId": 888100,
        "UnitName": "Celsius"
      }
    ],
    "Location": {
      "Altitude": null,
      "Latitude": 62.388618,
      "Longitude": 17.308968
    },
    "Timestamp": {
      "Convention": "TimeBeginning",
      "Timestamp": "2023-09-22T15:01:00+00:00"
    }
  }
]
`
