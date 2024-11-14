package backup

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/repo"
)

type Backup struct {
	producer       Producer
	restorer       Restorer
	backupInterval time.Duration
	lastBackup     time.Time
	storage        repo.Repository
}

func New(config server.Builder, storage repo.Repository) (*Backup, error) {
	p, err := NewProducer(config.FileStoragePath)
	if err != nil {
		return nil, err
	}
	r, err := NewRestorer(config.FileStoragePath)
	if err != nil {
		return nil, err
	}

	return &Backup{
		producer:       *p,
		restorer:       *r,
		backupInterval: config.StoreInterval,
		lastBackup:     time.Now().UTC(),
		storage:        storage,
	}, nil
}

func (b *Backup) Restore() {
	for {
		metric, err := b.restorer.ReadMetric()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("unable to restore metric: %v", err)
		}
		err = b.storage.Store(*metric)
		if err != nil {
			log.Printf("unable to store metric, during backup restore: %v", err)
		}
	}
}

func (b *Backup) Backup() {
	metrics := b.storage.GetAll()
	for _, m := range metrics {
		if err := b.producer.WriteMetric(m); err != nil {
			log.Printf("unable to backup metric: %v\n", err)
		}
	}
}

func (b *Backup) Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UTC()
		if now.Sub(b.lastBackup) >= b.backupInterval {
			b.Backup()
		}
		h.ServeHTTP(w, r)
	}
}

func (b *Backup) Close() error {
	if err := b.producer.Close(); err != nil {
		return err
	}
	if err := b.restorer.Close(); err != nil {
		return err
	}
	return nil
}
