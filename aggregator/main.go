package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/opospisil/grpc-microservices-excercise/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		distanceRepo    = NewInMemoryDistanceRepository()
		distanceService = NewInvoiceAggregator(distanceRepo)
    httpHandler     = NewAggHttpHandler(distanceService)
		httpListenAddr  = os.Getenv("AGG_HTTP_ADDR")
		grpcListenAddr  = os.Getenv("AGG_GRPC_ADDR")
	)
	distanceService = NewLogMiddleware(distanceService)
	distanceService = NewMetricsMiddleware(distanceService)
  httpHandler = NewHttpMetricsMiddleware(httpHandler)

	go makeGRPCTransport(grpcListenAddr, distanceService)
	makeHTTPTransport(httpListenAddr, httpHandler)
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

func makeHTTPTransport(listenerAddr string, ah AggHttpHandler) error {
	http.HandleFunc("/aggregate", ah.HandleAggregate().ServeHTTP)
	http.HandleFunc("/invoice", ah.HandleGetInvoice().ServeHTTP)
	http.Handle("/metrics", promhttp.Handler())
	logrus.Infof("Aggregator service HTTP listening on %s", listenerAddr)
	return http.ListenAndServe(listenerAddr, nil)
}
