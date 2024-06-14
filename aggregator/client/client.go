package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opospisil/grpc-microservices-excercise/model"
)

type AggClient struct {
	Endpoint string
}

func NewClient(endpoint string) *AggClient {
	return &AggClient{Endpoint: endpoint}
}

func (c *AggClient) AggregateInvoice(distance model.Distance) error {
	b, err := json.Marshal(distance)
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
