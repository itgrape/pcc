package pccclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetConfigSendsCredentialsAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runtime/namespaces/production/configs/database" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Header.Get(accessKeyHeader) != "pcc_ak_service" {
			t.Errorf("access key header = %q", r.Header.Get(accessKeyHeader))
		}
		if r.Header.Get(accessSecretHeader) != "pcc_sk_service" {
			t.Errorf("secret header = %q", r.Header.Get(accessSecretHeader))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"name":"database","format":1,"content":"{\"host\":\"db.internal\"}","version":3}}`))
	}))
	defer server.Close()

	client, err := New(server.URL, "pcc_ak_service", "pcc_sk_service", server.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	config, err := client.GetConfig(context.Background(), "production", "database")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if config.Name != "database" || config.Version != 3 {
		t.Fatalf("GetConfig() = %#v", config)
	}
}
