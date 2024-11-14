package backup

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/talx-hub/malerter/internal/model"
)

type Restorer struct {
	file   *os.File
	reader *bufio.Reader
}

func NewRestorer(filename string) (*Restorer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Restorer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (r *Restorer) ReadMetric() (*model.Metric, error) {
	data, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	metric := model.NewMetric()
	if err = json.Unmarshal(data, metric); err != nil {
		return nil, err
	}

	if err = metric.CheckValid(); err != nil {
		return nil, err
	}
	return metric, nil
}

func (r *Restorer) Close() error {
	if err := r.file.Close(); err != nil {
		return err
	}
	return nil
}
