package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/opospisil/grpc-microservices-excercise/aggregator/client"
	"github.com/opospisil/grpc-microservices-excercise/proto"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Fatal("Error loading .env file")
	}

	var (
		httpListenAddr     = os.Getenv("CALC_HTTP_ADDR")
		aggregatorEndpoint = fmt.Sprintf("http://%s", os.Getenv("AGG_HTTP_ADDR"))
	)

	httpClient := client.NewHttpClient(aggregatorEndpoint)
	invoiceSvc := NewInvoiceService(httpClient)

	http.HandleFunc("/invoice", invoiceSvc.HandleGetInvoiceApiFunc())
	logrus.Infof("Starting HTTP server on %s", httpListenAddr)
	logrus.Fatal(http.ListenAndServe(httpListenAddr, nil))
}

type InvoiceService struct {
	client client.AggClient
}

func (is *InvoiceService) handleGetInvoice(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("obuid")
	if id == "" {
		return fmt.Errorf("obuid is required")
	}
	intId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	inv, err := is.client.GetInvoice(r.Context(), &proto.GetInvoiceRequest{ObuID: int64(intId)})
	if err != nil {
		logrus.Errorf("Error getting invoice: %v", err.Error())
		return err
	}

	return writeJson(w, http.StatusOK, inv)
}

func (is *InvoiceService) HandleGetInvoiceApiFunc() http.HandlerFunc {
	return makeApiFunc(is.handleGetInvoice)
}

func NewInvoiceService(client client.AggClient) *InvoiceService {
	return &InvoiceService{client: client}
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeApiFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			logrus.Errorf("Error handling request: %+v", err)
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
}

func writeJson(rw http.ResponseWriter, status int, v any) error {
	rw.WriteHeader(status)
	rw.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(rw).Encode(v)
}
