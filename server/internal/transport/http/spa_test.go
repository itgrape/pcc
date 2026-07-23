package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterSPAServesAssetsAndClientRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	if err := os.Mkdir(assets, 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<main>PCC</main>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assets, "app.js"), []byte("window.PCC=true"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	router := gin.New()
	if err := registerSPAFrom(router, root); err != nil {
		t.Fatalf("registerSPAFrom() error = %v", err)
	}

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   "<main>PCC</main>",
			wantCache:  "no-cache",
		},
		{
			path:       "/namespaces/123/configs",
			wantStatus: http.StatusOK,
			wantBody:   "<main>PCC</main>",
			wantCache:  "no-cache",
		},
		{
			path:       "/assets/app.js",
			wantStatus: http.StatusOK,
			wantBody:   "window.PCC=true",
			wantCache:  "public, max-age=31536000, immutable",
		},
		{
			path:       "/api/v1/missing",
			wantStatus: http.StatusNotFound,
			wantBody:   `"message":"not found"`,
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if !strings.Contains(recorder.Body.String(), test.wantBody) {
				t.Fatalf("body = %q, want to contain %q", recorder.Body.String(), test.wantBody)
			}
			if test.wantCache != "" && recorder.Header().Get("Cache-Control") != test.wantCache {
				t.Fatalf("Cache-Control = %q, want %q", recorder.Header().Get("Cache-Control"), test.wantCache)
			}
		})
	}
}

func TestRegisterSPARejectsMissingIndex(t *testing.T) {
	router := gin.New()
	if err := registerSPAFrom(router, t.TempDir()); err == nil {
		t.Fatal("registerSPAFrom() error = nil, want missing index error")
	}
}
