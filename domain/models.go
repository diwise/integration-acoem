package domain

type Device struct {
	UniqueId   int    `json:"uniqueID"`
	DeviceName string `json:"deviceName"`
}

type DeviceData struct {
	TBTimestamp string    `json:"tbtimestamp"`
	TETimestamp string    `json:"tetimestamp"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Altitude    float64   `json:"altitude"`
	Channels    []Channel `json:"channels"`
}

type Channel struct {
	SensorName  string   `json:"sensorName"`
	SensorLabel string   `json:"sensorLabel"`
	Channel     int      `json:"channel"`
	PreScaled   float64  `json:"preScaled"`
	Scaled      float64  `json:"scaled"`
	UnitName    string   `json:"unitName"`
	Slope       int      `json:"slope"`
	Offset      int      `json:"offset"`
	Flags       []string `json:"flags"`
}
