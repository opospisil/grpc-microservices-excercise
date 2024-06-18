package repository

import (
	"fmt"

	"github.com/opospisil/grpc-microservices-excercise/model"
)

type DistanceRepository interface {
	Store(*model.Distance) error
  Get(int64) (float64, error)
}

type InMemoryDistanceRepository struct {
	data map[int64]float64
}

func NewInMemoryDistanceRepository() DistanceRepository {
	return &InMemoryDistanceRepository{
		data: make(map[int64]float64),
	}
}

func (idr *InMemoryDistanceRepository) Store(distance *model.Distance) error {
  idr.data[distance.OBUID] = distance.Value
	return nil
}

func (idr *InMemoryDistanceRepository) Get(obuid int64) (float64, error) {
  result, ok := idr.data[obuid]
  if !ok {
    return 0, fmt.Errorf("OBU ID %d not found", obuid)
    //return 76.54, nil // fix distance just to get some data to invoice
  }
  return result, nil
}
