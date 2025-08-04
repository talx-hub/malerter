package server

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Add(ctx context.Context, metric model.Metric) error {
	args := m.Called(ctx, metric)
	//nolint:wrapcheck // it's tests
	return args.Error(0)
}

func (m *mockStorage) Batch(_ context.Context, _ []model.Metric) error {
	return nil
}

func (m *mockStorage) Find(_ context.Context, _ string) (model.Metric, error) {
	return model.Metric{}, nil
}

func (m *mockStorage) Get(_ context.Context) ([]model.Metric, error) {
	return nil, nil
}

func (m *mockStorage) Ping(_ context.Context) error {
	return nil
}

func Test_parseTrustedSubnet_valid(t *testing.T) {
	cfg := &server.Builder{TrustedSubnet: "10.0.0.0/8"}
	subnet, err := parseTrustedSubnet(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	_, cidrNet, _ := net.ParseCIDR("10.0.0.0/8")
	assert.Equal(t, cidrNet.String(), subnet.String())
}

func Test_parseTrustedSubnet_empty(t *testing.T) {
	cfg := &server.Builder{TrustedSubnet: ""}
	subnet, err := parseTrustedSubnet(cfg)
	assert.NoError(t, err)
	assert.Nil(t, subnet)
}

func Test_parseTrustedSubnet_invalid(t *testing.T) {
	cfg := &server.Builder{TrustedSubnet: "invalid-subnet"}
	subnet, err := parseTrustedSubnet(cfg)
	assert.Error(t, err)
	assert.Nil(t, subnet)
}

func Test_initDecrypter_emptyPath(t *testing.T) {
	cfg := &server.Builder{CryptoKeyPath: constants.EmptyPath}
	decrypter, err := initDecrypter(cfg)
	assert.NoError(t, err)
	assert.Nil(t, decrypter)
}

func Test_initDecrypter_invalidPath(t *testing.T) {
	cfg := &server.Builder{CryptoKeyPath: "/non/existing/path.key"}
	decrypter, err := initDecrypter(cfg)
	assert.Error(t, err)
	assert.Nil(t, decrypter)
}

func Test_Init_returnsHTTP(t *testing.T) {
	cfg := &server.Builder{
		UseGRPC:       false,
		CryptoKeyPath: constants.EmptyPath,
		TrustedSubnet: "",
		RootAddress:   ":8080",
		Secret:        "",
	}

	log, _ := logger.New("debug")
	storage := new(mockStorage)

	s := Init(cfg, storage, log)
	assert.NotNil(t, s)
	assert.Implements(t, (*Server)(nil), s)
}

func Test_Init_returnsGRPC(t *testing.T) {
	cfg := &server.Builder{
		UseGRPC:       true,
		CryptoKeyPath: constants.EmptyPath,
		TrustedSubnet: "",
		RootAddress:   ":50051",
		Secret:        "",
	}

	log, _ := logger.New("debug")
	storage := new(mockStorage)

	s := Init(cfg, storage, log)
	assert.NotNil(t, s)
	assert.Implements(t, (*Server)(nil), s)
}
