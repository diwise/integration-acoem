package domain

// many of these properties may be unnecessary, so far we only really need the UniqueID to make the request for the latest data.
type Station struct {
	UniqueId    int    `json:"uniqueID"`
	StationName string `json:"stationName"`
}

type StationData struct {
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
