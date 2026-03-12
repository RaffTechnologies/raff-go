// Package raff provides a Go client library for the Raff Cloud API.
//
// Usage:
//
//	client := raff.NewFromToken("raff_pub_xxx")
//	projects, _, err := client.Projects.List(ctx, nil)
package raff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Version is the client library version.
	Version = "0.1.0"

	defaultBaseURL   = "https://api.rafftechnologies.com"
	defaultUserAgent = "raff-go/" + Version
)

// Client manages communication with the Raff Cloud API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	projectID  string
	userAgent  string

	// Services
	Projects ProjectService
	VMs      VMService
}

// NewFromToken creates a new Raff API client with the given API key.
func NewFromToken(apiKey string) *Client {
	return New(&http.Client{Timeout: 30 * time.Second}, apiKey)
}

// New creates a new Raff API client with a custom HTTP client.
func New(httpClient *http.Client, apiKey string, opts ...ClientOpt) *Client {
	c := &Client{
		httpClient: httpClient,
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		userAgent:  defaultUserAgent,
	}

	c.Projects = &ProjectServiceOp{client: c}
	c.VMs = &VMServiceOp{client: c}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ClientOpt is a functional option for configuring the client.
type ClientOpt func(*Client)

// SetBaseURL sets the API base URL.
func SetBaseURL(baseURL string) ClientOpt {
	return func(c *Client) {
		baseURL = strings.TrimRight(baseURL, "/")
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "https://" + baseURL
		}
		c.baseURL = baseURL
	}
}

// SetUserAgent sets the User-Agent header.
func SetUserAgent(ua string) ClientOpt {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// SetProjectID sets the default project ID sent via X-Project-ID header.
func SetProjectID(id string) ClientOpt {
	return func(c *Client) {
		c.projectID = id
	}
}

// ListOptions specifies pagination parameters.
type ListOptions struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}

// Response wraps http.Response and adds API-specific fields.
type Response struct {
	*http.Response
	Total int `json:"total,omitempty"`
}

// apiResponse is the standard JSON envelope from the Raff API.
type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
	Total   int             `json:"total,omitempty"`
}

// ErrorResponse is returned when the API returns an error.
type ErrorResponse struct {
	StatusCode int
	Message    string
	Reason     string // Machine-readable reason (e.g., "banned", "no_payment_method")
	Body       string
}

func (e *ErrorResponse) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Body)
}

// NewRequest creates an API request.
func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	u := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	if c.projectID != "" {
		req.Header.Set("X-Project-ID", c.projectID)
	}

	return req, nil
}

// Do sends an API request and unmarshals the response data into v.
func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	response := &Response{Response: resp}

	if resp.StatusCode >= 400 {
		return response, checkResponse(resp.StatusCode, body)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return response, fmt.Errorf("parsing response: %w", err)
	}

	response.Total = apiResp.Total

	if v != nil && apiResp.Data != nil {
		if err := json.Unmarshal(apiResp.Data, v); err != nil {
			return response, fmt.Errorf("parsing response data: %w", err)
		}
	}

	return response, nil
}

func checkResponse(statusCode int, body []byte) error {
	var apiResp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	}

	errResp := &ErrorResponse{
		StatusCode: statusCode,
		Body:       string(body),
	}

	if json.Unmarshal(body, &apiResp) == nil {
		if apiResp.Message != "" {
			errResp.Message = apiResp.Message
		} else {
			errResp.Message = apiResp.Error
		}
		errResp.Reason = apiResp.Reason
	}

	return errResp
}

// addListOptions adds pagination query params to a path.
func addListOptions(path string, opts *ListOptions) string {
	if opts == nil {
		return path
	}

	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	if len(params) == 0 {
		return path
	}

	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + params.Encode()
}
