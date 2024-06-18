package main

import (
	"net"
	"net/http"
	"os"

	"github.com/go-kit/log"

	"github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/http/controllers"
	"github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/http/handlers"
	"github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/repository"
	"github.com/opospisil/grpc-microservices-excercise/aggregator_gokit/service"
)

func main() {
	// Create a single logger, which we'll use and give to other components.


	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	repo := repository.NewInMemoryDistanceRepository()
	var (
		service     = service.NewAggregatorServiceImp(repo)
		endpoints   = controllers.NewSet(service, logger)
		httpHandler = handlers.NewHttpHandler(endpoints, logger)
	)
	// The HTTP listener mounts the Go kit HTTP handler we created.
	httpListener, err := net.Listen("tcp", ":4000")
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
		os.Exit(1)
	}
	logger.Log("transport", "HTTP", "addr", ":4000")
	err = http.Serve(httpListener, httpHandler)
	if err != nil {
		panic(err)
	}
}
