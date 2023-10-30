package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	requestReadTimeout    = 5 * time.Second
	headerReadTimeout     = 2 * time.Second
	responseWriteTimeout  = 5 * time.Second
	connIdleTimeout       = 30 * time.Second
	serverShutdownTimeout = 5 * time.Second
)

type Config struct {
	BindAddress string `yaml:"bind_address"`
}

type Server struct {
	cfg    Config
	server *http.Server
}

func NewServer(config Config) *Server {
	return &Server{cfg: config}
}

func (s *Server) Run(handler http.Handler) error {
	listener, err := net.Listen("tcp", s.cfg.BindAddress)
	if err != nil {
		panic(fmt.Sprintf("failed to start listen on TCP socket: %s", err))
	}

	s.server = &http.Server{
		Handler:           handler,
		ReadTimeout:       requestReadTimeout,
		ReadHeaderTimeout: headerReadTimeout,
		WriteTimeout:      responseWriteTimeout,
		IdleTimeout:       connIdleTimeout,
	}

	err = s.server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shutdown HTTP server gracefully")
	}

	return nil
}
