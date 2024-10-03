package repo

import "github.com/alant1t/metricscoll/internal/customerror"

type Repository interface {
	Store(metric string) error
	Get(metric string) (string, error)
}

type MemRepository struct {
	data []string
}

func (r *MemRepository) Store(metric string) error {
	r.data = append(r.data, metric)
	return nil
}

func (r *MemRepository) Get(metric string) (string, error) {
	for _, e := range r.data {
		if e == metric {
			return e, nil
		}
	}
	// почему тут нужно брать адрес???
	return "", &customerror.NotFoundError{Metric: metric}
}
