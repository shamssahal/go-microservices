package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shamssahal/toll-calculator/types"
)

type HTTPClient struct {
	Endpoint string
}

func NewHTTPClient(endpoint string) *HTTPClient {
	return &HTTPClient{
		Endpoint: endpoint,
	}
}

func (c *HTTPClient) Aggregate(ctx context.Context, aggReq *types.AggregateRequest) error {
	endpoint := fmt.Sprintf("%s/aggregate", c.Endpoint)
	var b []byte
	b, err := json.Marshal(aggReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the service responded with a non 200 status code %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPClient) Invoice(ctx context.Context, id string) (*types.Invoice, error) {
	endpoint := fmt.Sprintf("%s/invoice?id=%s", c.Endpoint, id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the service responded with a non 200 status code %d", resp.StatusCode)
	}
	var invoiceData types.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&invoiceData); err != nil {
		return nil, err
	}
	return &invoiceData, nil
}
