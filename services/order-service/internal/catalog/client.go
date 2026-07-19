package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrProductNotFound = errors.New("product not found")

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Client is an HTTP client for the Product Catalog service.
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// GetProduct fetches a product by ID, forwarding the request ID for
// cross-service log correlation.
func (c *Client) GetProduct(ctx context.Context, id, requestID string) (Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/products/"+id, nil)
	if err != nil {
		return Product{}, err
	}
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Product{}, fmt.Errorf("product catalog unreachable: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var p Product
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return Product{}, fmt.Errorf("decode catalog response: %w", err)
		}
		return p, nil
	case http.StatusNotFound:
		return Product{}, ErrProductNotFound
	default:
		return Product{}, fmt.Errorf("product catalog returned status %d", resp.StatusCode)
	}
}
