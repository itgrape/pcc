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

	FormatJSON       = 1
	FormatYAML       = 2
	FormatText       = 3
	FormatProperties = 4
)

type Client struct {
	baseURL       string
	accessKey     string
	secret        string
	namespace     string
	defaultConfig string
	http          *http.Client
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

type clientOptions struct {
	namespace     string
	defaultConfig string
	httpClient    *http.Client
	timeout       time.Duration
	timeoutSet    bool
}

type Option func(*clientOptions) error

func WithNamespace(namespace string) Option {
	return func(options *clientOptions) error {
		if strings.TrimSpace(namespace) == "" {
			return ErrNamespaceRequired
		}
		options.namespace = namespace
		return nil
	}
}

func WithConfig(configName string) Option {
	return func(options *clientOptions) error {
		if strings.TrimSpace(configName) == "" {
			return ErrConfigRequired
		}
		options.defaultConfig = configName
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(options *clientOptions) error {
		if httpClient == nil {
			return fmt.Errorf("pccclient: HTTP client cannot be nil")
		}
		options.httpClient = httpClient
		return nil
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(options *clientOptions) error {
		if timeout <= 0 {
			return fmt.Errorf("pccclient: timeout must be positive")
		}
		options.timeout = timeout
		options.timeoutSet = true
		return nil
	}
}

func New(baseURL, accessKey, secret string, opts ...Option) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, ErrInvalidBaseURL
	}
	if accessKey == "" || secret == "" {
		return nil, ErrCredentialsRequired
	}

	options := clientOptions{timeout: 10 * time.Second}
	for _, option := range opts {
		if option == nil {
			continue
		}
		if err := option(&options); err != nil {
			return nil, err
		}
	}

	httpClient := options.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: options.timeout}
	} else if options.timeoutSet {
		cloned := *httpClient
		cloned.Timeout = options.timeout
		httpClient = &cloned
	}

	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		accessKey:     accessKey,
		secret:        secret,
		namespace:     options.namespace,
		defaultConfig: options.defaultConfig,
		http:          httpClient,
	}, nil
}

func (c *Client) GetConfig(ctx context.Context, namespace, configName string) (*Config, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, ErrNamespaceRequired
	}
	if strings.TrimSpace(configName) == "" {
		return nil, ErrConfigRequired
	}
	path := fmt.Sprintf(
		"/api/v1/runtime/namespaces/%s/configs/%s",
		url.PathEscape(namespace),
		url.PathEscape(configName),
	)
	var result response[Config]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, &APIError{Code: result.Code, Message: result.Message}
	}
	return &result.Data, nil
}

func (c *Client) ListConfigs(ctx context.Context, namespace string) ([]Config, error) {
	if strings.TrimSpace(namespace) == "" {
		return nil, ErrNamespaceRequired
	}
	path := fmt.Sprintf(
		"/api/v1/runtime/namespaces/%s/configs",
		url.PathEscape(namespace),
	)
	var result response[[]Config]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, &APIError{Code: result.Code, Message: result.Message}
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
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       strings.TrimSpace(string(body)),
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("pccclient: decode response: %w", err)
	}
	return nil
}
