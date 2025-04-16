package backup

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/talx-hub/malerter/internal/constants"
	"github.com/talx-hub/malerter/internal/model"
)

type restorer struct {
	reader *bufio.Reader
	file   *os.File
}

func newRestorer(filename string) (*restorer, error) {
	file, err := os.OpenFile(
		filename,
		os.O_RDONLY|os.O_CREATE,
		constants.PermissionFilePrivate)
	if err != nil {
		return nil,
			fmt.Errorf("unable to open backup file %s: %w", filename, err)
	}

	return &restorer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (r *restorer) readMetric() (model.Metric, error) {
	data, err := r.reader.ReadBytes('\n')
	if err != nil {
		return model.Metric{}, fmt.Errorf("unable to read line from backup: %w", err)
	}

	metric := model.Metric{}
	if err = json.Unmarshal(data, &metric); err != nil {
		return metric, fmt.Errorf("unable to unmarshal metric: %w", err)
	}

	if err = metric.CheckValid(); err != nil {
		return metric, fmt.Errorf("metric is invalid: %w", err)
	}
	return metric, nil
}

func (r *restorer) read() ([]model.Metric, error) {
	metrics := make([]model.Metric, 0)
	errCount := 0
	for {
		metric, err := r.readMetric()
		if err == nil {
			metrics = append(metrics, metric)
			continue
		}
		if errors.Is(err, io.EOF) {
			break
		}
		errCount++
	}

	if errCount == 0 {
		return metrics, nil
	}
	return metrics, fmt.Errorf("failed to read %d metrics", errCount)
}

func (r *restorer) close() error {
	if err := r.file.Close(); err != nil {
		return fmt.Errorf("unable to close backup file: %w", err)
	}
	return nil
}
