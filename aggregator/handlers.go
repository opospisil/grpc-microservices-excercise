package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/sirupsen/logrus"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func (af apiFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := af(w, r); err != nil {
		logrus.Errorf("Error handling request: %+v", err)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

type ApiError struct {
	Code int
	Err  error
}

func (e ApiError) Error() string {
	return e.Err.Error()
}

type AggHttpHandler interface {
	HandleAggregate() apiFunc
	HandleGetInvoice() apiFunc
}

type AggHttpHandlerImpl struct {
	svc AggregatorService
}

func NewAggHttpHandler(svc AggregatorService) AggHttpHandler {
	return &AggHttpHandlerImpl{svc: svc}
}

func (ah *AggHttpHandlerImpl) HandleAggregate() apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var distance model.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			return ApiError{
				Code: http.StatusBadRequest,
				Err:  err,
			}
		}
		if err := ah.svc.AggregateDistance(&distance); err != nil {
			return ApiError{
				Code: http.StatusInternalServerError,
				Err:  err,
			}
		}
		w.WriteHeader(http.StatusCreated)
		return nil
	}
}

func (ah *AggHttpHandlerImpl) HandleGetInvoice() apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		obuid := r.URL.Query().Get("obuid")
		if obuid == "" {
			return ApiError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("OBU ID is required"),
			}
		}
		obuidInt, err := strconv.ParseInt(obuid, 10, 64)
		if err != nil {
			return ApiError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("invalid OBU ID format %s", obuid),
			}
		}

		invoice, err := ah.svc.GetInvoice(obuidInt)
		if err != nil {
			return ApiError{
				Code: http.StatusInternalServerError,
				Err:  err,
			}
		}
		return writeJson(w, http.StatusOK, invoice)
	}
}

func writeJson(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}

