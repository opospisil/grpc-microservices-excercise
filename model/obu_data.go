package model

type Distance struct {
	Value     float64 `json:"value"`
	OBUID     int     `json:"obuid"`
	Timestamp int64   `json:"timestamp"`
}

type OBUData struct {
	OBUID int     `json:"obuid"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
}
