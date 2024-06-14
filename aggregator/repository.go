package main

import "github.com/opospisil/grpc-microservices-excercise/model"

type DistanceRepository interface {
	Store(model.Distance) error
}

type InMemoryDistanceRepository struct {
	data map[int]float64
}

func NewInMemoryDistanceRepository() DistanceRepository {
	return &InMemoryDistanceRepository{
		data: make(map[int]float64),
	}
}

func (idr *InMemoryDistanceRepository) Store(distance model.Distance) error {
  idr.data[distance.OBUID] = distance.Value
	return nil
}
