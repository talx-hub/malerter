package backup

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
)

type Restorer struct {
	reader *bufio.Reader
	file   *os.File
}

func NewRestorer(filename string) (*Restorer, error) {
	file, err := os.OpenFile(
		filename,
		os.O_RDONLY|os.O_CREATE,
		constants.PermissionFilePrivate)
	if err != nil {
		return nil, fmt.Errorf("unable to open backup: %w", err)
	}

	return &Restorer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (r *Restorer) ReadMetric() (*model.Metric, error) {
	data, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("unable to read line from backup: %w", err)
	}

	metric := model.NewMetric()
	if err = json.Unmarshal(data, metric); err != nil {
		return nil, fmt.Errorf("unable to unmarshal metric: %w", err)
	}

	if err = metric.CheckValid(); err != nil {
		return nil, fmt.Errorf("backed metric is invalid: %w", err)
	}
	return metric, nil
}

func (r *Restorer) Close() error {
	if err := r.file.Close(); err != nil {
		return fmt.Errorf("unable to close backup file: %w", err)
	}
	return nil
}
