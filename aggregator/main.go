package main

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	http.HandleFunc("/invoice", handleGetInvoice(distanceService))

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
		if err := svc.AggregateDistance(&distance); err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func handleGetInvoice(svc AggregatorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		obuid := r.URL.Query().Get("obuid")
		if obuid == "" {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "OBU ID is required"})
			return
		}
		obuidInt, err := strconv.Atoi(obuid)
		if err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "OBU ID must be an integer"})
			return
		}

		invoice, err := svc.GetInvoice(obuidInt)
		if err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, http.StatusOK, invoice)
	}
}

func writeJson(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
