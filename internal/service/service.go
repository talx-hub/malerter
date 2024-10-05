package service

type Service interface {
	DumpMetric(string) error
}
