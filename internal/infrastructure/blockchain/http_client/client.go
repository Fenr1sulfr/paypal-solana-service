package http_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"payment.Diploma.service/internal/config"
)

type Client struct {
	baseURL    string
	privateKey string
	httpClient *http.Client
}

// type TSClientConfig struct {
// 	URL        string
// 	PrivateKey string
// 	TimeoutSec int
// }

func NewClient(cfg config.Config) *Client {

	return &Client{
		baseURL:    cfg.TSClient.URL,
		privateKey: cfg.TSClient.PrivateKey,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TSClient.TimeoutSec) * time.Second,
		},
	}
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) ProxyInitStorage(req *InitializeStorageRequest) (*InitializeStorageResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Request body: %s\n", string(body)) // Debugging line
	resp, err := c.httpClient.Post(c.baseURL+"/initialize", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %s, body: %s", resp.Status, string(respBody))
	}

	var result InitializeStorageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ProxyAddPayment(req *AddPaymentRequest) (*AddPaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Request body: %s\n", string(body)) // Debugging line
	resp, err := c.httpClient.Post(c.baseURL+"/add-payment", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %s, body: %s", resp.Status, string(respBody))
	}
	var result AddPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
