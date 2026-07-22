// Package pccclient provides a small HTTP client for service-to-service
// configuration reads from PCC.
package pccclient

import (
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
	accessKeyHeader    = "X-PCC-Access-Key"
	accessSecretHeader = "X-PCC-Secret"
)

type Client struct {
	baseURL   string
	accessKey string
	secret    string
	http      *http.Client
}

type Config struct {
	Name        string `json:"name"`
	Format      int    `json:"format"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Version     uint   `json:"version"`
	UpdatedAt   string `json:"updated_at"`
}

type response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func New(baseURL, accessKey, secret string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("pccclient: invalid base URL")
	}
	if accessKey == "" || secret == "" {
		return nil, fmt.Errorf("pccclient: access key and secret are required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		accessKey: accessKey,
		secret:    secret,
		http:      httpClient,
	}, nil
}

func (c *Client) GetConfig(ctx context.Context, namespace, configName string) (*Config, error) {
	path := fmt.Sprintf(
		"/api/v1/runtime/namespaces/%s/configs/%s",
		url.PathEscape(namespace),
		url.PathEscape(configName),
	)
	var result response[Config]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

func (c *Client) ListConfigs(ctx context.Context, namespace string) ([]Config, error) {
	path := fmt.Sprintf(
		"/api/v1/runtime/namespaces/%s/configs",
		url.PathEscape(namespace),
	)
	var result response[[]Config]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *Client) get(ctx context.Context, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("pccclient: build request: %w", err)
	}
	req.Header.Set(accessKeyHeader, c.accessKey)
	req.Header.Set(accessSecretHeader, c.secret)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("pccclient: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("pccclient: server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("pccclient: decode response: %w", err)
	}
	return nil
}
