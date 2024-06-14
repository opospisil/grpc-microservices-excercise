package model

type Invoice struct {
	OBUID    int     `json:"obuid"`
	Amount   float64 `json:"amount"`
	DateTime string  `json:"dateTime"`
	distance float64
}
