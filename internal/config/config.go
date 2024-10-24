package config

type Config interface{}

const (
	hostDefault = "localhost:8080"
)

const (
	reportIntervalDefault = 10
	poolIntervalDefault   = 2
)

const (
	envAddress        = "ADDRESS"
	envReportInterval = "REPORT_INTERVAL"
	envPollInterval   = "POLL_INTERVAL"
)
