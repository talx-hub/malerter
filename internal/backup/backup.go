package backup

import (
	"context"
	"time"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
)

type Storage interface {
	Batch(context.Context, []model.Metric) error
	Get(context.Context) ([]model.Metric, error)
}

type Manager struct {
	storage        Storage
	log            *logger.ZeroLogger
	backupInterval time.Duration
	filename       string
	needRestore    bool
}

func New(
	config *server.Builder,
	storage Storage,
	log *logger.ZeroLogger,
) *Manager {
	if config == nil {
		log.Error().Msg("unable to load Backup service: config is nil")
		return nil
	}

	return &Manager{
		backupInterval: config.StoreInterval,
		storage:        storage,
		log:            log,
		filename:       config.FileStoragePath,
		needRestore:    config.Restore,
	}
}

func (b *Manager) Run(ctx context.Context) {
	if b.needRestore {
		b.restore(ctx)
	}

	ticker := time.NewTicker(b.backupInterval * time.Second)
	b.log.Info().Msg("start backup service")
	for {
		select {
		case <-ctx.Done():
			b.log.Info().Msg("shutdown backup service...")
			b.backup(ctx)
			return
		default:
			<-ticker.C
			b.backup(ctx)
		}
	}
}

func (b *Manager) restore(ctx context.Context) {
	b.log.Info().Msg("start restore metrics from backup...")
	r, err := newRestorer(b.filename)
	if err != nil {
		b.log.Error().Err(err).Msg("unable to open backup Restorer")
		return
	}
	defer func() {
		if err := r.close(); err != nil {
			b.log.Error().Err(err).Msg("close backup failed")
		}
	}()

	metrics, err := r.read()
	if err != nil {
		b.log.Error().Err(err).Msg("read backup failed")
		return
	}
	err = b.storage.Batch(ctx, metrics)
	if err != nil {
		b.log.Error().Err(err).Msg("write backup batch failed")
		return
	}
	b.log.Info().Msg("backup restore successful!")
}

func (b *Manager) backup(ctx context.Context) {
	b.log.Info().Msg("start metrics backup...")
	p, err := newProducer(b.filename)
	if err != nil {
		b.log.Error().Err(err).Msg("unable to open backup Producer")
		return
	}
	defer func() {
		if err := p.close(); err != nil {
			b.log.Error().Err(err).Msg("close backup failed")
		}
	}()

	metrics, err := b.storage.Get(ctx)
	if err != nil {
		b.log.Error().Err(err).Msg("get metrics from storage failed")
		return
	}

	if err = p.write(metrics); err != nil {
		b.log.Error().Err(err).Msg("write metrics to file failed")
		return
	}

	if err = p.flush(); err != nil {
		b.log.Error().Err(err).Msg("flush metrics to backup failed")
		return
	}
	b.log.Info().Msg("metrics backup successful!")
}
