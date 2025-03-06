package backup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
)

type Producer struct {
	writer *bufio.Writer
	file   *os.File
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		constants.PermissionFilePrivate)
	if err != nil {
		return nil, fmt.Errorf("unable to open backup file %s: %w", filename, err)
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteMetric(metric model.Metric) error {
	jsonEncoder := json.NewEncoder(p.writer)
	if err := jsonEncoder.Encode(&metric); err != nil {
		return fmt.Errorf("unable to backup metric: %w", err)
	}

	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("unable to flush backup: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	if err := p.file.Close(); err != nil {
		return fmt.Errorf("unable to close backup file error: %w", p.file.Close())
	}
	return nil
}
