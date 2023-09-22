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
	"strings"
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

func CreateAndSendAsLWM2M(ctx context.Context, sensors []domain.DeviceData, uniqueId int, url string, sender SenderFunc) error {
	logger := logging.GetFromContext(ctx)

	var errs []error

	uniqueIdStr := strconv.Itoa(uniqueId)
	log := logger.With().Str("device_id", uniqueIdStr).Logger()

	for _, s := range sensors {
		timestamp, err := time.Parse(time.RFC3339, s.Timestamp.Timestamp)
		if err != nil {
			errs = append(errs, err)
			log.Error().Err(err).Msg("could not parse timestamp")
			continue
		}

		packs := make(map[string]senml.Pack)

		for _, c := range s.Channels {
			if strings.EqualFold("Temperature", c.SensorName) {
				if _, ok := packs[TemperatureURN]; !ok {
					packs[TemperatureURN] = newPack(TemperatureURN, "5700", uniqueIdStr, c.PreScaled.Reading, senml.UnitCelsius, timestamp, timestamp)
				}
			}
			if strings.EqualFold("Humidity", c.SensorName) {
				if _, ok := packs[HumidityURN]; !ok {
					packs[HumidityURN] = newPack(HumidityURN, "5700", uniqueIdStr, c.PreScaled.Reading, senml.UnitRelativeHumidity, timestamp, timestamp)
				}
			}
			if strings.EqualFold("Particulate Matter (PM 10)", c.SensorName) {
				if _, ok := packs[AirQualityURN]; !ok {
					packs[AirQualityURN] = newPack(AirQualityURN, "1", uniqueIdStr, c.PreScaled.Reading, "ug/m3", timestamp, timestamp)
				} else {
					packs[AirQualityURN] = append(packs[AirQualityURN], newRec("1", c.PreScaled.Reading, "ug/m3", timestamp))
				}
			}
			if strings.EqualFold("Particulate Matter (PM 2.5)", c.SensorName) {
				if _, ok := packs[AirQualityURN]; !ok {
					packs[AirQualityURN] = newPack(AirQualityURN, "3", uniqueIdStr, c.PreScaled.Reading, "ug/m3", timestamp, timestamp)
				} else {
					packs[AirQualityURN] = append(packs[AirQualityURN], newRec("3", c.PreScaled.Reading, "ug/m3", timestamp))
				}
			}
			if strings.EqualFold("Particulate Matter (PM 1)", c.SensorName) {
				if _, ok := packs[AirQualityURN]; !ok {
					packs[AirQualityURN] = newPack(AirQualityURN, "5", uniqueIdStr, c.PreScaled.Reading, "ug/m3", timestamp, timestamp)
				} else {
					packs[AirQualityURN] = append(packs[AirQualityURN], newRec("5", c.PreScaled.Reading, "ug/m3", timestamp))
				}
			}
			if strings.EqualFold("Nitrogen Dioxide", c.SensorName) {
				if _, ok := packs[AirQualityURN]; !ok {
					packs[AirQualityURN] = newPack(AirQualityURN, "15", uniqueIdStr, c.PreScaled.Reading, "ppm", timestamp, timestamp)
				} else {
					packs[AirQualityURN] = append(packs[AirQualityURN], newRec("15", c.PreScaled.Reading, "ppm", timestamp))
				}
			}
			if strings.EqualFold("Nitric Oxide", c.SensorName) {
				if _, ok := packs[AirQualityURN]; !ok {
					packs[AirQualityURN] = newPack(AirQualityURN, "19", uniqueIdStr, c.PreScaled.Reading, "ppm", timestamp, timestamp)
				} else {
					packs[AirQualityURN] = append(packs[AirQualityURN], newRec("19", c.PreScaled.Reading, "ppm", timestamp))
				}
			}
		}

		for _, p := range packs {
			err := sender(ctx, url, p)
			if err != nil {
				log.Error().Err(err).Msg("could not send pack")
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func newPack(baseName, name, id string, v float64, u string, bt, t time.Time) senml.Pack {
	p := senml.Pack{
		senml.Record{
			BaseName:    baseName,
			BaseTime:    float64(bt.Unix()),
			Name:        "0",
			StringValue: id,
		},
		newRec(name, v, u, t),
	}
	return p
}

func newRec(name string, v float64, u string, t time.Time) senml.Record {
	return senml.Record{
		Name:  name,
		Value: &v,
		Time:  float64(t.Unix()),
		Unit:  u,
	}
}

type SenderFunc = func(context.Context, string, senml.Pack) error

func Send(ctx context.Context, url string, pack senml.Pack) error {
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
