package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opospisil/grpc-microservices-excercise/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AggClient interface {
	AggregateDistance(context.Context, *proto.AggregateDistanceRequest) error
}

type HttpAggClient struct {
	Endpoint string
}

func NewHttpClient(endpoint string) AggClient {
	return &HttpAggClient{Endpoint: endpoint}
}

func (c *HttpAggClient) AggregateDistance(ctx context.Context, aggReq *proto.AggregateDistanceRequest) error {
	b, err := json.Marshal(aggReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.Endpoint+"/aggregate", bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d\nresponse: %+v", resp.StatusCode, resp)
	}
	return nil
}

type GrpcAggClient struct {
	Endpoint string
	client   proto.DistanceAggregatorClient
}

func NewGRPCClient(endpoint string) (*GrpcAggClient, error) {
	conn, err := grpc.NewClient(endpoint, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}...)
	if err != nil {
		return nil, err
	}

	c := proto.NewDistanceAggregatorClient(conn)

	return &GrpcAggClient{
		Endpoint: endpoint,
		client:   c,
	}, nil
}

func (c *GrpcAggClient) AggregateDistance(ctx context.Context, aggReq *proto.AggregateDistanceRequest) error {
	_, err := c.client.AggregateDistance(ctx, aggReq)
	return err
}
