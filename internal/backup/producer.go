package backup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
)

type producer struct {
	writer *bufio.Writer
	file   *os.File
}

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		constants.PermissionFilePrivate)
	if err != nil {
		return nil,
			fmt.Errorf("unable to open backup file %s: %w", filename, err)
	}

	return &producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *producer) writeMetric(metric model.Metric) error {
	jsonEncoder := json.NewEncoder(p.writer)
	if err := jsonEncoder.Encode(&metric); err != nil {
		return fmt.Errorf("unable to backup metric: %w", err)
	}

	return nil
}

func (p *producer) write(metrics []model.Metric) error {
	errCount := 0
	for _, m := range metrics {
		if err := p.writeMetric(m); err != nil {
			errCount += 1
		}
	}
	if errCount == 0 {
		return nil
	}
	return fmt.Errorf("failed to write %d metrics", errCount)
}

func (p *producer) flush() error {
	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("unable to flush backup: %w", err)
	}
	return nil
}

func (p *producer) close() error {
	if err := p.file.Close(); err != nil {
		return fmt.Errorf("unable to close backup file: %w", err)
	}
	return nil
}
