package lwm2m

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/diwise/integration-acoem/domain"
	"github.com/diwise/service-chassis/pkg/infrastructure/env"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"github.com/farshidtz/senml/v2"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var tlsSkipVerify bool

func init() {
	tlsSkipVerify = env.GetVariableOrDefault(zerolog.Logger{}, "TLS_SKIP_VERIFY", "0") == "1"
}

var tracer = otel.Tracer("integration-acoem/lwm2m")

const (
	AirQualityURN  string = "urn:oma:lwm2m:ext:3428"
	HumidityURN    string = "urn:oma:lwm2m:ext:3304"
	TemperatureURN string = "urn:oma:lwm2m:ext:3303"
)

func CreateAndSendAsLWM2M(ctx context.Context, sensors []domain.DeviceData, uniqueId int, deviceName, url string) error {
	logger := logging.GetFromContext(ctx)

	var errs []error

	/*
		Go through deviceData.Channels, create lwm2m packs for the properties we recognise:
			humidity
			temperature

		Create airquality for pollutants, where each pollutant is a pack on the record or whichever order it is.
	*/

	log := logger.With().Str("device_id", strconv.Itoa(uniqueId)).Logger()

	uniqueIdStr := strconv.Itoa(uniqueId)

	for _, s := range sensors {
		timestamp, err := time.Parse(time.RFC3339, s.Timestamp.Timestamp)
		if err != nil {
			errs = append(errs, err)
		}

		for _, c := range s.Channels {
			if c.SensorName == "temperature" {
				pack, err := temperature(ctx, uniqueIdStr, c.PreScaled.Reading, timestamp)
				if err != nil {
					log.Error().Err(err).Msg("unable to create lwm2m temperature object")
				}

				err = send(ctx, url, pack)
				if err != nil {
					log.Error().Err(err).Msg("unable to POST lwm2m temperature")
					errs = append(errs, err)
				}

				log.Info().Msgf("sending lwm2m pack for %s", timestamp)
			}
		}
	}

	return errors.Join(errs...)
}

func temperature(ctx context.Context, deviceID string, temp float64, date time.Time) (senml.Pack, error) {
	SensorValue := func(v float64, t time.Time) SenMLDecoratorFunc {
		return Value("5700", v, t, senml.UnitCelsius)
	}

	pack := NewSenMLPack(deviceID, TemperatureURN, date, SensorValue(temp, date))

	return pack, nil
}

func airquality(ctx context.Context, deviceID string, temp float64, date time.Time) (senml.Pack, error) {
	SensorValue := func(v float64, t time.Time) SenMLDecoratorFunc {
		return Value("17", v, t, "")
	}

	pack := NewSenMLPack(deviceID, AirQualityURN, date, SensorValue(temp, date))

	return pack, nil
}

func humidity(ctx context.Context, deviceID string, temp float64, date time.Time) (senml.Pack, error) {
	SensorValue := func(v float64, t time.Time) SenMLDecoratorFunc {
		return Value("5700", v, t, senml.UnitRelativeHumidity)
	}

	pack := NewSenMLPack(deviceID, HumidityURN, date, SensorValue(temp, date))

	return pack, nil
}

func send(ctx context.Context, url string, pack senml.Pack) error {
	var err error

	ctx, span := tracer.Start(ctx, "send-object")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	var httpClient http.Client

	if tlsSkipVerify {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient = http.Client{
			Transport: otelhttp.NewTransport(customTransport),
		}
	} else {
		httpClient = http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
	}

	b, err := json.Marshal(pack)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/senml+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusCreated {
		err = fmt.Errorf("unexpected response code %d", resp.StatusCode)
	}

	return err
}
