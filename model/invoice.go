package model

type Invoice struct {
	OBUID    int64     `json:"obuid"`
	Amount   float64 `json:"amount"`
	DateTime string  `json:"dateTime"`
	distance float64
}
