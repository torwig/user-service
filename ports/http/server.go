package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
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
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
