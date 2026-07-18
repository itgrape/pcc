// Command server 启动 PCC Config Center HTTP 服务。
package main

import (
	"context"
	"errors"
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

	// 初始化 Casdoor JWT 验签
	infra.InitCasdoor(cfg.Casdoor.Certificate)

	// 事件总线
	bus := app.NewEventBus()

	// 用例层
	nsUsecase := app.NewNamespaceUsecase(store.Namespace, store.Config, store.AccessKeyNamespace)
	cfgUsecase := app.NewConfigUsecase(store, bus)
	akUsecase := app.NewAccessKeyUsecase(store)

	r := httpapi.New(httpapi.Deps{
		Namespaces: nsUsecase,
		Configs:    cfgUsecase,
		AccessKeys: akUsecase,
		Bus:        bus,
		Store:      store,
	})

	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-stop:
		slog.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
