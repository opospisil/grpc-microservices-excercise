package main

import (
	"math"
)

type CalculatorService interface {
	CalculateDistance(pair CoordsPair) (float64, error)
}

type CalculatorServiceImpl struct{
}

func (cs *CalculatorServiceImpl) CalculateDistance(pair CoordsPair) (float64, error) {
  err := pair.Validate()
  if err != nil {
    return 0, err
  }
  return CalculateDistance(pair), nil
}

func NewCalculatorService() CalculatorService {
	return &CalculatorServiceImpl{}
}

func CalculateDistance(pair CoordsPair) float64 {
	return math.Sqrt(
		(pair.Current.Lat-pair.Previous.Lat)*(pair.Current.Lat-pair.Previous.Lat) +
			(pair.Current.Lon-pair.Previous.Lon)*(pair.Current.Lon-pair.Previous.Lon),
	)
}
