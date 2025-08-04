package server

import (
	"context"
	"fmt"
	"net"

	"github.com/talx-hub/malerter/internal/api/handlers"
	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/service/server/customgrpc"
	"github.com/talx-hub/malerter/internal/service/server/customhttp"
	"github.com/talx-hub/malerter/pkg/crypto"
)

type Server interface {
	Start() error
	Stop(context.Context) error
}

func Init(
	cfg *server.Builder,
	storage handlers.Storage,
	log *logger.ZeroLogger,
) Server {
	decrypter, err := initDecrypter(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("server init error")
		return nil
	}

	agentSubnet, err := parseTrustedSubnet(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("server init error")
		return nil
	}

	if cfg.UseGRPC {
		return customgrpc.New(
			storage, log, decrypter, cfg.RootAddress, cfg.Secret, agentSubnet)
	}

	return customhttp.New(
		storage, log, decrypter, cfg.RootAddress, cfg.Secret, agentSubnet)
}

func parseTrustedSubnet(cfg *server.Builder) (*net.IPNet, error) {
	if cfg.TrustedSubnet == "" {
		//nolint:nilnil // it's ok, to have *net.IPNet == nil
		return nil, nil
	}

	_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
	}
	return subnet, nil
}

func initDecrypter(cfg *server.Builder) (*crypto.Decrypter, error) {
	if cfg.CryptoKeyPath == constants.EmptyPath {
		//nolint:nilnil // it's ok, to have *decrypter == nil
		return nil, nil
	}

	decrypter, err := crypto.NewDecrypter(cfg.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init decrypter: %w", err)
	}
	return decrypter, nil
}
