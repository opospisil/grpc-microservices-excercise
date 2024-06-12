package model

type OBUData struct {
	OBUID int     `json:"obuid"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
}
