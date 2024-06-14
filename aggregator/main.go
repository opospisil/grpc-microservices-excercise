package main

import (
	"encoding/json"
	"net/http"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/sirupsen/logrus"
)

const listenAddr = "localhost:8080"

func main() {
	var (
		distanceRepo    = NewInMemoryDistanceRepository()
		distanceService = NewInvoiceAggregator(distanceRepo)
	)
	distanceService = NewLogMiddleware(distanceService)

	http.HandleFunc("/aggregate", handleAggregate(distanceService))
	logrus.Infof("Aggregator service listening on %s", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

func handleAggregate(svc AggregatorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var distance model.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := svc.AggregateDistance(distance); err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func writeJson(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
