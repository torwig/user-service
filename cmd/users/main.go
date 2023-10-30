package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/torwig/user-service/adapters/user"
	"github.com/torwig/user-service/config"
	"github.com/torwig/user-service/log"
	"github.com/torwig/user-service/ports/http"
	"github.com/torwig/user-service/ports/http/jwt"
	"github.com/torwig/user-service/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	Version        = "Undefined"
	configFilePath = "config.yaml"
)

func main() {
	flag.StringVar(&configFilePath, "c", configFilePath, "path to config file")
	flag.Parse()

	cfg, err := config.ParseFromFile(configFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse configuration: %s", err))
	}

	logger := log.NewZapLogger(cfg.Log)
	defer func(logger *zap.SugaredLogger) {
		_ = logger.Sync()
	}(logger)

	logger.Infof("starting service, version=%s", Version)

	repo, err := user.NewPostgresRepository(cfg.Repository)
	if err != nil {
		panic(fmt.Sprintf("failed to create repository: %s", err))
	}

	svc := service.New(repo)
	authenticator := jwt.NewAuthenticator(cfg.JWT)
	handler := http.NewHandler(svc, authenticator, logger)
	srv := http.NewServer(cfg.HTTP)

	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	errGroup, errCtx := errgroup.WithContext(signalCtx)

	errGroup.Go(func() error {
		return srv.Run(handler.Router())
	})

	errGroup.Go(func() error {
		<-errCtx.Done()

		shutdownCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		return srv.Shutdown(shutdownCtx)
	})

	if err := errGroup.Wait(); err != nil && errors.Is(err, context.Canceled) {
		logger.Errorf("service error: %s", err)
	}

	logger.Infof("service is about to exit")
}
