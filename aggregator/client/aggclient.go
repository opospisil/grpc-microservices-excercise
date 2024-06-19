package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/opospisil/grpc-microservices-excercise/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AggClient interface {
	AggregateDistance(context.Context, *proto.AggregateDistanceRequest) error
	GetInvoice(context.Context, *proto.GetInvoiceRequest) (*proto.InvoiceResponse, error)
}

type HttpAggClient struct {
	Endpoint string
}

func NewHttpClient(endpoint string) AggClient {
	return &HttpAggClient{Endpoint: endpoint}
}

func (c *HttpAggClient) GetInvoice(ctx context.Context, invRq *proto.GetInvoiceRequest) (*proto.InvoiceResponse, error) {
	resp, err := http.Get(c.Endpoint + "/invoice?obuid=" + strconv.Itoa(int(invRq.ObuID)))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d\nresponse: %+v", resp.StatusCode, resp)
	}

	var invoice proto.InvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invoice); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &invoice, nil
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

func (c *GrpcAggClient) GetInvoice(ctx context.Context, invRq *proto.GetInvoiceRequest) (*proto.InvoiceResponse, error) {
	return c.client.GetInvoice(ctx, invRq)
}
