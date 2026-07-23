// Command server 启动 PCC Config Center HTTP 服务。
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/itgrape/pcc/server/internal/app"
	"github.com/itgrape/pcc/server/internal/config"
	"github.com/itgrape/pcc/server/internal/infra"
	"github.com/itgrape/pcc/server/internal/repository"
	httpapi "github.com/itgrape/pcc/server/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	gin.SetMode(gin.ReleaseMode)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := infra.OpenDB(cfg.DB, cfg.Snowflake.NodeID)
	if err != nil {
		return err
	}
	store := repository.New(db)

	auth, err := infra.NewCasdoor(cfg.Casdoor)
	if err != nil {
		return err
	}

	// 事件总线
	bus := app.NewEventBus()

	// 用例层
	nsUsecase := app.NewNamespaceUsecase(store.Namespace, store.Config, store.AccessKeyNamespace)
	cfgUsecase := app.NewConfigUsecase(store, bus)
	akUsecase := app.NewAccessKeyUsecase(store)
	runtimeConfigUsecase := app.NewRuntimeConfigUsecase(
		store.AccessKey,
		store.Namespace,
		store.AccessKeyNamespace,
		store.Config,
	)

	adminRouter, err := httpapi.NewAdmin(httpapi.AdminDeps{
		Namespaces: nsUsecase,
		Configs:    cfgUsecase,
		AccessKeys: akUsecase,
		Auth:       auth,
	})
	if err != nil {
		return err
	}
	runtimeRouter, err := httpapi.NewRuntime(httpapi.RuntimeDeps{
		Configs: runtimeConfigUsecase,
	})
	if err != nil {
		return err
	}

	servers := map[string]*http.Server{
		"admin":   newHTTPServer(cfg.Server.AdminAddr, adminRouter),
		"runtime": newHTTPServer(cfg.Server.RuntimeAddr, runtimeRouter),
	}
	errCh := make(chan error, len(servers))
	for name, srv := range servers {
		go serve(name, srv, errCh)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	var runErr error
	select {
	case runErr = <-errCh:
	case <-stop:
		slog.Info("shutting down")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var shutdownErr error
	for name, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown %s server: %w", name, err))
		}
	}
	return errors.Join(runErr, shutdownErr)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}

func serve(name string, srv *http.Server, errCh chan<- error) {
	slog.Info("http server listening", "name", name, "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- fmt.Errorf("%s server: %w", name, err)
	}
}
