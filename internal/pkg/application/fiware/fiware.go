package fiware

import (
	"context"
	"errors"
	"strconv"

	fw "github.com/diwise/context-broker/pkg/datamodels/fiware"
	"github.com/diwise/context-broker/pkg/ngsild/client"
	ngsierrors "github.com/diwise/context-broker/pkg/ngsild/errors"
	"github.com/diwise/context-broker/pkg/ngsild/types"
	"github.com/diwise/context-broker/pkg/ngsild/types/entities"
	. "github.com/diwise/context-broker/pkg/ngsild/types/entities/decorators"
	"github.com/diwise/context-broker/pkg/ngsild/types/properties"
	"github.com/diwise/integration-acoem/domain"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("integration-acoem/fiware")

func CreateOrUpdateAirQualityObserved(ctx context.Context, cbClient client.ContextBrokerClient, sensors []domain.DeviceData, deviceName string, uniqueId int) error {
	var err error

	ctx, span := tracer.Start(ctx, "create-air-qualities")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	_, ctx, logger := o11y.AddTraceIDToLoggerAndStoreInContext(span, logging.GetFromContext(ctx), ctx)

	headers := map[string][]string{"Content-Type": {"application/ld+json"}}

	decorators := []entities.EntityDecoratorFunc{}

	decorators = append(decorators, entities.DefaultContext(), Text("areaServed", deviceName))

	for _, sensor := range sensors {
		decorators = append(decorators,
			Location(sensor.Location.Latitude, sensor.Location.Longitude),
			DateTime(properties.DateObserved, sensor.Timestamp.Timestamp),
		)

		sensorReadings := createFragmentsFromSensorData(sensor.Channels, sensor.Timestamp.Timestamp)

		decorators = append(decorators, sensorReadings...)
	}

	var fragment types.EntityFragment
	fragment, err = entities.NewFragment(decorators...)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create entity fragments")
	}

	entityID := fw.AirQualityObservedIDPrefix + strconv.Itoa(uniqueId)

	_, err = cbClient.MergeEntity(ctx, entityID, fragment, headers)
	if err != nil {
		if !errors.Is(err, ngsierrors.ErrNotFound) {
			logger.Error().Err(err).Msg("failed to merge entity")
		}

		var entity types.Entity
		entity, err = entities.New(entityID, fw.AirQualityObservedTypeName, decorators...)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create new entity")

		}

		_, err = cbClient.CreateEntity(ctx, entity, headers)
		if err != nil {
			logger.Error().Err(err).Msg("failed to post entity to context broker")

		}

		logger.Info().Msgf("created entity %s", entityID)
	} else {
		logger.Info().Msgf("updated entity %s", entityID)
	}

	return nil
}

func createFragmentsFromSensorData(sensors []domain.Channel, timestamp string) []entities.EntityDecoratorFunc {
	readings := []entities.EntityDecoratorFunc{}

	for _, sensor := range sensors {
		name, ok := sensorNames[sensor.SensorName]
		if ok {
			readings = append(readings, Number(
				name,
				sensor.Scaled.Reading,
				properties.UnitCode(unitCodes[sensor.UnitName]),
				properties.ObservedAt(timestamp),
			))
		}
	}

	return readings
}

var unitCodes map[string]string = map[string]string{
	"Micrograms Per Cubic Meter": "GQ",
	"Volts":                      "VLT",
	"Celsius":                    "CEL",
	"Percent":                    "P1",
	"Hectopascals":               "A97",
	"Parts Per Billion":          "61",
	"Pressure (mbar)":            "MBR",
}

var sensorNames map[string]string = map[string]string{
	"Humidity":                    "relativeHumidity",
	"Temperature":                 "temperature",
	"Air Pressure":                "atmosphericPressure",
	"Particulate Matter (PM 1)":   "PM1",
	"PM 4":                        "PM4",
	"Particulate Matter (PM 10)":  "PM10",
	"Particulate Matter (PM 2.5)": "PM25",
	"Total Suspended Particulate": "totalSuspendedParticulate",
	"Voltage":                     "voltage",
	"Nitric Oxide":                "NO",
	"Nitrogen Dioxide":            "NO2",
	"Nitrogen Oxides":             "NOx",
}