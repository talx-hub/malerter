package backup

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Add(model.Metric) error
	Get() []model.Metric
}

type File struct {
	lastBackup     time.Time
	storage        Storage
	producer       Producer
	restorer       Restorer
	backupInterval time.Duration
}

func New(config *server.Builder, storage Storage) (*File, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	p, err := NewProducer(config.FileStoragePath)
	if err != nil {
		return nil, err
	}
	r, err := NewRestorer(config.FileStoragePath)
	if err != nil {
		return nil, err
	}

	return &File{
		producer:       *p,
		restorer:       *r,
		backupInterval: config.StoreInterval,
		lastBackup:     time.Now().UTC(),
		storage:        storage,
	}, nil
}

func (b *File) Restore() {
	for {
		metric, err := b.restorer.ReadMetric()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Printf("unable to restore metric: %v", err)
		}
		err = b.storage.Add(*metric)
		if err != nil {
			log.Printf("unable to store metric, during backup restore: %v", err)
		}
	}
}

func (b *File) Backup() {
	metrics := b.storage.Get()
	for _, m := range metrics {
		if err := b.producer.WriteMetric(m); err != nil {
			log.Printf("unable to backup metric: %v\n", err)
		}
	}
}

func (b *File) Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		if now.Sub(b.lastBackup) >= b.backupInterval {
			b.Backup()
		}
		h.ServeHTTP(w, r)
	}
}

func (b *File) Close() error {
	if err := b.producer.Close(); err != nil {
		return err
	}
	if err := b.restorer.Close(); err != nil {
		return err
	}
	return nil
}
