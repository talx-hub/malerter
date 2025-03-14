package backup

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(context.Context, model.Metric) error
	Get(context.Context) ([]model.Metric, error)
}

type File struct {
	lastBackup     time.Time
	storage        Storage
	producer       Producer
	restorer       Restorer
	log            *logger.ZeroLogger
	backupInterval time.Duration
}

func New(
	config *server.Builder,
	storage Storage,
	log *logger.ZeroLogger,
) *File {
	if config == nil {
		log.Error().Msg("unable to load Backup service: config is nil")
		return nil
	}
	p, err := NewProducer(config.FileStoragePath)
	if err != nil {
		log.Error().Err(err).Msg("unable to create backup Producer")
		return nil
	}
	r, err := NewRestorer(config.FileStoragePath)
	if err != nil {
		log.Error().Err(err).Msg("unable to create backup Restorer")
		return nil
	}

	return &File{
		producer:       *p,
		restorer:       *r,
		backupInterval: config.StoreInterval,
		lastBackup:     time.Now().UTC(),
		storage:        storage,
		log:            log,
	}
}

func (b *File) Restore() {
	for {
		metric, err := b.restorer.ReadMetric()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			b.log.Error().Err(err).
				Msg("unable to restore metric")
		}
		err = b.storage.Add(context.TODO(), *metric)
		if err != nil {
			b.log.Error().Err(err).
				Msg("unable to store metric, during backup restore")
		}
	}
}

func (b *File) Backup() {
	metrics, _ := b.storage.Get(context.TODO())
	for _, m := range metrics {
		if err := b.producer.WriteMetric(m); err != nil {
			b.log.Error().Err(err).Msg("unable to backup metric")
		}
	}
}

func (b *File) Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		now := time.Now().UTC()
		if now.Sub(b.lastBackup) >= b.backupInterval {
			b.Backup()
		}
	}
}

func (b *File) Close() {
	if err := b.producer.Close(); err != nil {
		b.log.Error().Err(err).
			Msg("unable to properly close Backup service")
	}
	if err := b.restorer.Close(); err != nil {
		b.log.Error().Err(err).
			Msg("unable to properly close Backup service")
	}
}
