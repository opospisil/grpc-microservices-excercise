package main

import (
	"errors"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/sirupsen/logrus"
)

type CoordsPair struct {
	Current, Previous *model.OBUData
}

type DataCache interface {
	Add(*model.OBUData) error
	PopPairs() ([]CoordsPair, error)
}

type NonPersistentCache struct {
	cache map[int][]*model.OBUData
}

func NewDataCache() DataCache {
	return &NonPersistentCache{
		cache: make(map[int][]*model.OBUData),
	}
}

func (dc *NonPersistentCache) Add(data *model.OBUData) error {
	dc.cache[data.OBUID] = append(dc.cache[data.OBUID], data)
	return nil
}

func (dc *NonPersistentCache) PopPairs() ([]CoordsPair, error) {
	var pairs []CoordsPair
	for k, data := range dc.cache {
		l := len(data)
		logrus.Infof("OBUID: %d, data: %d", k, l)
		if l < 2 {
			continue
		}
		pair := CoordsPair{data[l-1], data[l-2]}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (cp *CoordsPair) Validate() error {
	if cp.Current.OBUID != cp.Previous.OBUID {
		return errors.New("OBUIDs do not match")
	}
	return nil
}
