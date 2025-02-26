package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/diwise/integration-acoem/domain"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type IntegrationAcoem interface {
	GetDevices(ctx context.Context) ([]domain.Device, error)
	GetDeviceData(ctx context.Context, uniqueId int, sensorLabels string) ([]domain.DeviceData, error)
	GetSensorLabels(ctx context.Context, deviceID int) (string, error)
}

type integrationAcoem struct {
	baseUrl     string
	accessToken string
}

var tracer = otel.Tracer("integration-acoem/app")

func New(baseUrl, accountID, accountKey string) IntegrationAcoem {
	accessToken := fmt.Sprintf(
		"Basic %s",
		base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", accountID, accountKey)),
		),
	)

	return &integrationAcoem{
		baseUrl:     baseUrl,
		accessToken: accessToken,
	}
}

func (i *integrationAcoem) GetSensorLabels(ctx context.Context, deviceID int) (string, error) {
	var err error

	ctx, span := tracer.Start(ctx, "get-sensor-labels")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/devices/setup/%d", i.baseUrl, deviceID), nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %s", err.Error())
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)

	var resp *http.Response
	resp, err = httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request failed: %s", err.Error())
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed, expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		return "", err
	}

	sensors := []domain.Sensor{}

	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %s", err.Error())
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &sensors)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal response body: %s", err.Error())
		return "", err
	}

	sensorLabels := []string{}

	for _, s := range sensors {
		if s.Active && s.Type == "data" {
			sensorLabels = append(sensorLabels, s.SensorLabel)
		}
	}

	labels := strings.Join(sensorLabels, "+")

	return labels, nil
}

func (i *integrationAcoem) GetDeviceData(ctx context.Context, uniqueId int, sensorLabels string) ([]domain.DeviceData, error) {
	var err error

	ctx, span := tracer.Start(ctx, "get-device-data")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	if uniqueId == 0 || sensorLabels == "" {
		err = fmt.Errorf("cannot retrieve sensor data as either uniqueId or sensor labels are empty")
		return nil, err
	}
	deviceData := []domain.DeviceData{}

	numberOfRecords := 1 //The number of records you want to retrieve
	average := "AVG300"  //'AVG' or 'AVERAGE' followed by the average period in seconds. Valid seconds are: 0, 300, 600, 900, 1200, 1800, 3600, 7200, 10800, 14400, 21600, 28800, 43200, 86400
	type_ := "data"      //This can be 'data', 'diagnostic' or 'datadiagnostic'

	devicedataUrl := fmt.Sprintf("%s/devicedata/%d/latest/%d/%s/%s/%s", i.baseUrl, uniqueId, numberOfRecords, average, type_, sensorLabels)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, devicedataUrl, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %s", err.Error())
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)
	//req.Header.Add("TimeConvention", "TimeBeginning")
	req.Header.Add("TimeConvention", "TimeEnding") //The time convention for the request

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to retrieve sensor data: %s", err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		return nil, err
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body as bytes: %s", err.Error())
		return nil, err
	}

	err = json.Unmarshal(respBytes, &deviceData)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal json data: %s", err.Error())
		return nil, err
	}

	return deviceData, nil
}

func (i *integrationAcoem) GetDevices(ctx context.Context) ([]domain.Device, error) {
	var err error

	ctx, span := tracer.Start(ctx, "get-devices")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	devices := []domain.Device{}

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/devices", i.baseUrl), nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %s", err.Error())
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", i.accessToken)

	var response *http.Response
	response, err = httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to retrieve list of devices: %s", err.Error())
		return nil, err
	}

	defer response.Body.Close()

	var responseBytes []byte
	responseBytes, err = io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body as bytes: %s", err.Error())
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed, expected status code %d, got %d", http.StatusOK, response.StatusCode)
		return nil, err
	}

	err = json.Unmarshal(responseBytes, &devices)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal response: %s,\ndue to: %s", string(responseBytes), err.Error())
		return nil, err
	}

	return devices, nil
}
