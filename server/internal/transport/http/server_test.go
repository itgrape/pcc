package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/app"
	"github.com/itgrape/pcc/server/internal/domain"
)

type rejectingAuthenticator struct{}

func (rejectingAuthenticator) ExchangeCode(string, string) (string, error) {
	return "", errors.New("not implemented")
}

func (rejectingAuthenticator) Authenticate(string) (*domain.Identity, error) {
	return nil, errors.New("unauthorized")
}

type acceptingAuthenticator struct {
	rejectingAuthenticator
}

func (acceptingAuthenticator) Authenticate(authHeader string) (*domain.Identity, error) {
	if authHeader != "Bearer valid-token" {
		return nil, errors.New("unauthorized")
	}
	return &domain.Identity{Username: "alice"}, nil
}

func TestAdminRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := New(Deps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       rejectingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tested := 0
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, adminBasePath) {
			continue
		}
		tested++

		path := strings.NewReplacer(
			":namespace_id", "1",
			":config_id", "1",
			":history_id", "1",
			":key_id", "1",
		).Replace(route.Path)
		req := httptest.NewRequest(route.Method, path, nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s returned %d, want %d", route.Method, route.Path, recorder.Code, http.StatusUnauthorized)
		}
	}

	const expectedAdminRoutes = 25
	if tested != expectedAdminRoutes {
		t.Fatalf("tested %d administrative routes, want %d", tested, expectedAdminRoutes)
	}
}

func TestOAuthCallbackIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := New(Deps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       rejectingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/callback", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("callback returned %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestDocumentationRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := New(Deps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       rejectingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, path := range []string{"/openapi.json", "/docs", "/schemas/Response"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s returned %d, want %d", path, recorder.Code, http.StatusUnauthorized)
		}
	}
}

func TestAuthenticatedUserIsInjectedIntoRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := New(Deps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       acceptingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, adminBasePath+"/user/info", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("user info returned %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"alice"`) {
		t.Fatalf("user info body does not contain authenticated user: %s", recorder.Body.String())
	}
}
