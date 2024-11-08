package agent

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/talx-hub/malerter/internal/config"
)

const (
	HostDefault           = "localhost:8080"
	ReportIntervalDefault = 10
	PoolIntervalDefault   = 2
)

const (
	EnvHost           = "ADDRESS"
	EnvReportInterval = "REPORT_INTERVAL"
	EnvPollInterval   = "POLL_INTERVAL"
)

func NewDirector() *config.Director {
	return &config.Director{
		Builder: &Builder{},
	}
}

type Builder struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func (b *Builder) LoadFromFlags() config.Builder {
	flag.StringVar(&b.ServerAddress, "a", HostDefault, "alert-host address")

	var ri int64
	flag.Int64Var(&ri, "r", ReportIntervalDefault, "interval in seconds of sending metrics to alert server")

	var pi int64
	flag.Int64Var(&pi, "p", PoolIntervalDefault, "interval in seconds of polling and collecting metrics")
	flag.Parse()

	b.ReportInterval = time.Duration(ri) * time.Second
	b.PollInterval = time.Duration(pi) * time.Second
	return b
}

func (b *Builder) LoadFromEnv() config.Builder {
	if addr, found := os.LookupEnv(EnvHost); found {
		b.ServerAddress = addr
	}
	if ri, found := os.LookupEnv(EnvReportInterval); found {
		riInt, err := strconv.Atoi(ri)
		if err != nil {
			log.Fatal(err)
		}
		b.ReportInterval = time.Duration(riInt) * time.Second
	}
	if pi, found := os.LookupEnv(EnvPollInterval); found {
		piInt, err := strconv.Atoi(pi)
		if err != nil {
			log.Fatal(err)
		}
		b.PollInterval = time.Duration(piInt) * time.Second
	}
	return b
}

func (b *Builder) IsValid() (config.Builder, error) {
	if b.ReportInterval < 0 {
		return nil, fmt.Errorf("report interval must be positive")
	}
	if b.PollInterval < 0 {
		return nil, fmt.Errorf("poll interval must be positive")
	}
	return b, nil
}

func (b *Builder) Build() config.Config {
	// TODO: wtf?? почему вынужден возвращать значение?
	return *b
}
