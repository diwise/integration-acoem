package domain

type Device struct {
	UniqueId   int    `json:"uniqueID"`
	DeviceName string `json:"deviceName"`
}

type DeviceData struct {
	Timestamp struct {
		Convention string `json:"convention"`
		Timestamp  string `json:"timestamp"`
	} `json:"timestamp"`
	Location struct {
		Altitude  float64 `json:"altitude"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	} `json:"location"`
	Channels []Channel `json:"channels"`
}

type Channel struct {
	SensorName  string `json:"sensorName"`
	SensorLabel string `json:"sensorLabel"`
	Channel     int    `json:"channel"`
	PreScaled   struct {
		Reading float64 `json:"reading"`
	} `json:"preScaled"`
	Scaled struct {
		Reading float64 `json:"reading"`
	} `json:"scaled"`
	UnitName string   `json:"unitName"`
	Slope    int      `json:"slope"`
	Offset   int      `json:"offset"`
	Flags    []string `json:"flags"`
}

type Sensor struct {
	Active      bool   `json:"active"`
	SensorLabel string `json:"sensorLabel"`
}
