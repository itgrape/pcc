package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/app"
	"github.com/itgrape/pcc/server/internal/config"
	"github.com/itgrape/pcc/server/internal/domain"
	"github.com/itgrape/pcc/server/internal/infra"
	"github.com/itgrape/pcc/server/internal/model"
)

type rejectingAuthenticator struct{}

func (rejectingAuthenticator) Login(string, string) (string, error) {
	return "", errors.New("not implemented")
}
func (rejectingAuthenticator) Logout(string)      {}
func (rejectingAuthenticator) SecureCookie() bool { return true }

func (rejectingAuthenticator) Authenticate(string) (*domain.Identity, error) {
	return nil, errors.New("unauthorized")
}

type acceptingAuthenticator struct {
	rejectingAuthenticator
}

type runtimeConfigProviderFake struct {
	config *model.Config
}

func (f runtimeConfigProviderFake) Get(_ context.Context, accessKey, secret, namespace, configName string) (*model.Config, error) {
	if accessKey != "pcc_ak_service" || secret != "pcc_sk_service" || namespace != "production" || configName != "database" {
		return nil, domain.ErrUnauthenticated
	}
	return f.config, nil
}

func (f runtimeConfigProviderFake) List(context.Context, string, string, string) ([]*model.Config, error) {
	return []*model.Config{f.config}, nil
}

func (acceptingAuthenticator) Authenticate(token string) (*domain.Identity, error) {
	if token != "valid-token" {
		return nil, errors.New("unauthorized")
	}
	return &domain.Identity{Username: "alice"}, nil
}

func TestAdminRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := NewAdmin(AdminDeps{
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

	const expectedAdminRoutes = 26
	if tested != expectedAdminRoutes {
		t.Fatalf("tested %d administrative routes, want %d", tested, expectedAdminRoutes)
	}
}

func TestLoginIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := NewAdmin(AdminDeps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       rejectingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("login returned %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestAdminLoginSessionAndLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, err := infra.NewAdminAuth(config.AdminAuthConfig{Username: "admin", Password: "a-long-test-password", SecureCookie: true})
	if err != nil {
		t.Fatal(err)
	}
	router, err := NewAdmin(AdminDeps{Namespaces: new(app.NamespaceUsecase), Configs: new(app.ConfigUsecase), AccessKeys: new(app.AccessKeyUsecase), Auth: auth})
	if err != nil {
		t.Fatal(err)
	}

	login := func(origin, password string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"`+password+`"}`))
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		return out
	}
	if got := login("http://evil.example", "a-long-test-password").Code; got != http.StatusForbidden {
		t.Fatalf("cross-origin login = %d", got)
	}
	if got := login("http://example.com", "a-long-test-password").Code; got != http.StatusBadRequest {
		t.Fatalf("HTTP login with Secure cookies = %d", got)
	}
	if got := login("https://example.com", "wrong-password").Code; got != http.StatusUnauthorized {
		t.Fatalf("wrong password = %d", got)
	}
	result := login("https://example.com", "a-long-test-password")
	if result.Code != http.StatusOK {
		t.Fatalf("login = %d: %s", result.Code, result.Body.String())
	}
	cookies := result.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("unsafe session cookie: %+v", cookies)
	}
	userRequest := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, adminBasePath+"/user/info", nil)
		req.AddCookie(cookies[0])
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		return out
	}
	if got := userRequest().Code; got != http.StatusOK {
		t.Fatalf("authenticated user = %d", got)
	}
	logout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logout.Header.Set("Origin", "https://example.com")
	logout.AddCookie(cookies[0])
	out := httptest.NewRecorder()
	router.ServeHTTP(out, logout)
	if out.Code != http.StatusOK {
		t.Fatalf("logout = %d", out.Code)
	}
	if got := userRequest().Code; got != http.StatusUnauthorized {
		t.Fatalf("revoked session = %d", got)
	}
}

func TestHTTPLoginWhenSecureCookieIsExplicitlyDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, err := infra.NewAdminAuth(config.AdminAuthConfig{Username: "admin", Password: "a-long-test-password", SecureCookie: false})
	if err != nil {
		t.Fatal(err)
	}
	router, err := NewAdmin(AdminDeps{Namespaces: new(app.NamespaceUsecase), Configs: new(app.ConfigUsecase), AccessKeys: new(app.AccessKeyUsecase), Auth: auth})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"a-long-test-password"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://example.com")
	out := httptest.NewRecorder()
	router.ServeHTTP(out, req)
	if out.Code != http.StatusOK {
		t.Fatalf("HTTP login = %d: %s", out.Code, out.Body.String())
	}
	cookies := out.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected HTTP test cookie: %+v", cookies)
	}
}

func TestDocumentationRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := NewAdmin(AdminDeps{
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
	router, err := NewAdmin(AdminDeps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       acceptingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, adminBasePath+"/user/info", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-token"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("user info returned %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"name":"alice"`) {
		t.Fatalf("user info body does not contain authenticated user: %s", recorder.Body.String())
	}
}

func TestRuntimeConfigUsesAccessKeyCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := &model.Config{
		Base: model.Base{
			ID:        30,
			Version:   4,
			UpdatedAt: time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC),
		},
		ConfigName: "database",
		Format:     model.ConfigFormatJSON,
		Content:    `{"host":"db.internal"}`,
	}
	router, err := NewRuntime(RuntimeDeps{
		Configs: runtimeConfigProviderFake{config: config},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	path := runtimeBasePath + "/namespaces/production/configs/database"
	unauthorized := httptest.NewRequest(http.MethodGet, path, nil)
	unauthorizedRecorder := httptest.NewRecorder()
	router.ServeHTTP(unauthorizedRecorder, unauthorized)
	if unauthorizedRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("request without credentials returned %d, want %d", unauthorizedRecorder.Code, http.StatusUnauthorized)
	}

	authorized := httptest.NewRequest(http.MethodGet, path, nil)
	authorized.Header.Set(accessKeyHeader, "pcc_ak_service")
	authorized.Header.Set(accessSecretHeader, "pcc_sk_service")
	authorizedRecorder := httptest.NewRecorder()
	router.ServeHTTP(authorizedRecorder, authorized)
	if authorizedRecorder.Code != http.StatusOK {
		t.Fatalf("authorized request returned %d, want %d; body=%s", authorizedRecorder.Code, http.StatusOK, authorizedRecorder.Body.String())
	}
	if !strings.Contains(authorizedRecorder.Body.String(), `"content":"{\"host\":\"db.internal\"}"`) {
		t.Fatalf("authorized response does not contain config: %s", authorizedRecorder.Body.String())
	}
	if authorizedRecorder.Header().Get("ETag") != `"30-4"` {
		t.Fatalf("ETag = %q, want %q", authorizedRecorder.Header().Get("ETag"), `"30-4"`)
	}
}

func TestAdminAndRuntimeRoutesAreIsolated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	adminRouter, err := NewAdmin(AdminDeps{
		Namespaces: new(app.NamespaceUsecase),
		Configs:    new(app.ConfigUsecase),
		AccessKeys: new(app.AccessKeyUsecase),
		Auth:       rejectingAuthenticator{},
	})
	if err != nil {
		t.Fatalf("NewAdmin() error = %v", err)
	}
	runtimeRouter, err := NewRuntime(RuntimeDeps{
		Configs: new(app.RuntimeConfigUsecase),
	})
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}

	assertRouteNotFound(t, adminRouter, runtimeBasePath+"/namespaces/production/configs")
	assertRouteNotFound(t, runtimeRouter, adminBasePath+"/user/info")
	assertRouteNotFound(t, runtimeRouter, "/api/v1/auth/login")
}

func assertRouteNotFound(t *testing.T, router http.Handler, path string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("GET %s returned %d, want %d", path, recorder.Code, http.StatusNotFound)
	}
}
