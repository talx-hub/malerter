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
	data, err := json.Marshal(&metric)
	if err != nil {
		return fmt.Errorf("unable marshal metric: %w", err)
	}

	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("unable to write backup: %w", err)
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("unable to write string: %w", err)
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
