package main

import (
	"context"

	"github.com/opospisil/grpc-microservices-excercise/model"
	"github.com/opospisil/grpc-microservices-excercise/proto"
)

type GRPCServer struct {
	proto.UnimplementedDistanceAggregatorServer
	svc AggregatorService
}

func NewGRPCServer(svc AggregatorService) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) AggregateDistance(ctx context.Context, in *proto.AggregateDistanceRequest) (*proto.None, error) {
	dist := &model.Distance{
		Value:     in.Value,
		OBUID:     int64(in.ObuID),
		Timestamp: in.Timestamp,
	}

	return &proto.None{}, s.svc.AggregateDistance(dist)
}
