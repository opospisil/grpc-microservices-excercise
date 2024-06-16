package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/opospisil/grpc-microservices-excercise/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	httpListenAddr = "localhost:8080"
	grpcListenAddr = "localhost:8081"
)

func main() {
	var (
		distanceRepo    = NewInMemoryDistanceRepository()
		distanceService = NewInvoiceAggregator(distanceRepo)
	)
	distanceService = NewLogMiddleware(distanceService)

	go makeGRPCTransport(grpcListenAddr, distanceService)

	// aggClient, err := client.NewGRPCClient(grpcListenAddr)
	// if err != nil {
	//   logrus.Fatalf("Error creating gRPC client: %v", err)
	// }

	// aggClient.AggregateDistance(context.Background(), &proto.AggregateDistanceRequest{Value: 100, ObuID: 1, Timestamp: 0})

	makeHTTPTransport(httpListenAddr, distanceService)
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
		obuidInt, err := strconv.ParseInt(obuid, 10, 64)
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

func makeGRPCTransport(listenerAddr string, svc AggregatorService) error {
	// Create a listener on the specified address
	ln, err := net.Listen("tcp", listenerAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	// Create a new gRPC server
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	// Register the DistanceAggregator service with the gRPC server
	proto.RegisterDistanceAggregatorServer(grpcServer, NewGRPCServer(svc))
	logrus.Infof("Aggregator service gRPC listening on %s", listenerAddr)
	return grpcServer.Serve(ln)
}

func makeHTTPTransport(listenerAddr string, svc AggregatorService) error {
	http.HandleFunc("/aggregate", handleAggregate(svc))
	http.HandleFunc("/invoice", handleGetInvoice(svc))
	logrus.Infof("Aggregator service HTTP listening on %s", listenerAddr)
	return http.ListenAndServe(listenerAddr, nil)
}
