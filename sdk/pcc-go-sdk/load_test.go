package pccclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadAndLoadField(t *testing.T) {
	server := newConfigServer(t, Config{
		Name:    "application",
		Format:  FormatJSON,
		Content: `{"server":{"host":"api.internal","port":8080,"timeout":"5s"},"features":[{"enabled":true}]}`,
	})
	defer server.Close()

	client, err := New(
		server.URL,
		"pcc_ak_service",
		"pcc_sk_service",
		WithNamespace("production"),
		WithConfig("application"),
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var application struct {
		Server struct {
			Host    string        `json:"host"`
			Port    int           `json:"port"`
			Timeout time.Duration `json:"timeout"`
		} `json:"server"`
	}
	if err := client.Load(&application); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if application.Server.Host != "api.internal" ||
		application.Server.Port != 8080 ||
		application.Server.Timeout != 5*time.Second {
		t.Fatalf("Load() = %#v", application)
	}

	var timeout time.Duration
	if err := client.LoadField("server.timeout", &timeout); err != nil {
		t.Fatalf("LoadField(timeout) error = %v", err)
	}
	if timeout != 5*time.Second {
		t.Fatalf("timeout = %s, want 5s", timeout)
	}

	var enabled bool
	if err := client.LoadField("features[0].enabled", &enabled); err != nil {
		t.Fatalf("LoadField(indexed) error = %v", err)
	}
	if !enabled {
		t.Fatal("enabled = false, want true")
	}
}

func TestDecodeSupportedFormats(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "yaml",
			config: Config{
				Name:    "application",
				Format:  FormatYAML,
				Content: "server:\n  host: api.internal\n  port: 8080\n",
			},
		},
		{
			name: "properties",
			config: Config{
				Name:    "application",
				Format:  FormatProperties,
				Content: "server.host=api.internal\nserver.port=8080\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var target struct {
				Server struct {
					Host string `json:"host" yaml:"host"`
					Port int    `json:"port" yaml:"port"`
				} `json:"server" yaml:"server"`
			}
			if err := decodeConfig(&test.config, &target); err != nil {
				t.Fatalf("decodeConfig() error = %v", err)
			}
			if target.Server.Host != "api.internal" || target.Server.Port != 8080 {
				t.Fatalf("decodeConfig() = %#v", target)
			}
		})
	}
}

func TestLoadValidationAndMissingField(t *testing.T) {
	client, err := New(
		"https://pcc.example.com",
		"pcc_ak_service",
		"pcc_sk_service",
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := client.Load(&struct{}{}); !errors.Is(err, ErrNamespaceRequired) {
		t.Fatalf("Load() error = %v, want namespace required", err)
	}

	config := &Config{Name: "application", Format: FormatJSON, Content: `{"server":{}}`}
	document, err := decodeDocument(config)
	if err != nil {
		t.Fatalf("decodeDocument() error = %v", err)
	}
	if _, err := lookupPath(document, "server.port"); !errors.Is(err, ErrFieldNotFound) {
		t.Fatalf("lookupPath() error = %v, want field not found", err)
	}
}

func TestLoadConfigWithExplicitNamespace(t *testing.T) {
	server := newConfigServer(t, Config{
		Name:    "application",
		Format:  FormatText,
		Content: "release-2026-07",
	})
	defer server.Close()

	client, err := New(
		server.URL,
		"pcc_ak_service",
		"pcc_sk_service",
		WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	var release string
	if err := client.LoadConfig(context.Background(), "production", "application", &release); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if release != "release-2026-07" {
		t.Fatalf("release = %q", release)
	}
}

func newConfigServer(t *testing.T, config Config) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(accessKeyHeader) == "" || r.Header.Get(accessSecretHeader) == "" {
			t.Error("request is missing credentials")
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response[Config]{
			Code:    0,
			Message: "success",
			Data:    config,
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
}
