package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mizbancloud/cli/pkg/config"
)

type Client struct {
	httpClient *http.Client
	config     *config.Config
}

type Response struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config.GetConfig(),
	}
}

func (c *Client) request(method, endpoint string, body interface{}) (*Response, error) {
	url := c.config.BaseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized: please login again using 'mizban login'")
	}

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited: please wait and try again")
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	return &response, nil
}

func (c *Client) Get(endpoint string) (*Response, error) {
	return c.request(http.MethodGet, endpoint, nil)
}

func (c *Client) Post(endpoint string, body interface{}) (*Response, error) {
	return c.request(http.MethodPost, endpoint, body)
}

func (c *Client) Put(endpoint string, body interface{}) (*Response, error) {
	return c.request(http.MethodPut, endpoint, body)
}

func (c *Client) Delete(endpoint string) (*Response, error) {
	return c.request(http.MethodDelete, endpoint, nil)
}

func ParseData[T any](resp *Response) (T, error) {
	var result T
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return result, fmt.Errorf("error parsing data: %w", err)
	}
	return result, nil
}
