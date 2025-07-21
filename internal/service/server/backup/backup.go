package backup

import (
	"context"
	"time"

	"github.com/talx-hub/malerter/internal/config/server"
	"github.com/talx-hub/malerter/internal/logger"
	"github.com/talx-hub/malerter/internal/model"
	"github.com/talx-hub/malerter/pkg/queue"
)

type Storage interface {
	Batch(context.Context, []model.Metric) error
	Get(context.Context) ([]model.Metric, error)
}

type Manager struct {
	log            *logger.ZeroLogger
	buffer         *queue.Queue[model.Metric]
	storage        Storage
	filename       string
	backupInterval time.Duration
	needRestore    bool
}

func New(
	config *server.Builder,
	buffer *queue.Queue[model.Metric],
	storage Storage,
	log *logger.ZeroLogger,
) *Manager {
	if log == nil {
		return nil
	}
	if config == nil {
		log.Error().Msg("backup service: config is nil")
		return nil
	}
	if buffer == nil {
		log.Error().Msg("backup service: buffer is nil")
		return nil
	}
	if storage == nil {
		log.Error().Msg("backup service: storage is nil")
		return nil
	}

	return &Manager{
		log:            log,
		buffer:         buffer,
		storage:        storage,
		filename:       config.FileStoragePath,
		backupInterval: config.StoreInterval,
		needRestore:    config.Restore,
	}
}

func (b *Manager) Run(ctx context.Context) {
	if b.needRestore {
		b.restore(ctx)
	}

	var ticker *time.Ticker
	if b.backupInterval != 0 {
		ticker = time.NewTicker(b.backupInterval)
	}

	b.log.Info().Msg("START backup SERVICE")
	for {
		// можно ли в default делать только b.backup()
		// а отдельным case как-то проверять b.backupInterval != 0
		// если условие выполняется, то блокируюсь до срабатывания ticker
		// и затем fallthrough в default??
		select {
		case <-ctx.Done():
			b.log.Info().Msg("SHUTDOWN backup SERVICE...")
			b.backup()
			return
		default:
			if b.backupInterval != 0 {
				<-ticker.C
			}
			b.backup()
		}
	}
}

func (b *Manager) restore(ctx context.Context) {
	b.log.Info().Msg("start RESTORE metrics from backup...")
	b.buffer.Close()
	defer b.buffer.Open()

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
	b.log.Info().Msg("backup RESTORE successful!")
}

func (b *Manager) backup() {
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

	var metrics []model.Metric
	for b.buffer.Len() > 0 {
		metrics = append(metrics, b.buffer.Pop())
	}
	if len(metrics) == 0 {
		b.log.Info().Msg("no metrics to backup")
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
