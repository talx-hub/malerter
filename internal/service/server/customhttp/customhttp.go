package customhttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/talx-hub/malerter/internal/api/handlers"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/service/server/router"
	"github.com/talx-hub/malerter/pkg/crypto"
)

type CustomHTTP struct {
	http.Server
}

func New(
	storage handlers.Storage,
	log *logger.ZeroLogger,
	decrypter *crypto.Decrypter,
	address, secret string,
	subnet *net.IPNet,
) *CustomHTTP {
	chiRouter := router.New(log, subnet, secret, decrypter)
	chiRouter.SetRouter(handlers.NewHTTPHandler(storage, log))

	return &CustomHTTP{
		Server: http.Server{
			Addr:    address,
			Handler: chiRouter.GetRouter(),
		},
	}
}

func (s *CustomHTTP) Start() error {
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error during HTTP server ListenAndServe: %w", err)
	}

	return nil
}

func (s *CustomHTTP) Stop(ctx context.Context) error {
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("error during HTTP server Shutdown: %w", err)
	}

	return nil
}
