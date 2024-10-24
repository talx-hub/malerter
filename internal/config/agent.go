package config

import (
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"
    "time"
)

type Agent struct {
    ServerAddress  string
    ReportInterval time.Duration
    PollInterval   time.Duration
}

func (a *Agent) loadFromFlags() {
    flag.StringVar(&a.ServerAddress, "a", hostDefault,
        "alert-host address")

    var ri int64
    flag.Int64Var(&ri, "r", reportIntervalDefault,
        "interval in seconds of sending metrics to alert server")

    var pi int64
    flag.Int64Var(&pi, "p", poolIntervalDefault,
        "interval in seconds of polling and collecting metrics")
    flag.Parse()

    a.ReportInterval = time.Duration(ri) * time.Second
    a.PollInterval = time.Duration(pi) * time.Second
}

func (a *Agent) loadFromEnv() {
    if addr, found := os.LookupEnv(envAddress); found {
        a.ServerAddress = addr
    }
    if ri, found := os.LookupEnv(envReportInterval); found {
        riInt, err := strconv.Atoi(ri)
        if err != nil {
            log.Fatal(err)
        }
        a.ReportInterval = time.Duration(riInt) * time.Second
    }
    if pi, found := os.LookupEnv(envPollInterval); found {
        piInt, err := strconv.Atoi(pi)
        if err != nil {
            log.Fatal(err)
        }
        a.PollInterval = time.Duration(piInt) * time.Second
    }
}

func (a *Agent) isValid() (bool, error) {
    if a.ReportInterval < 0 {
        return false, fmt.Errorf("report interval must be positive")
    }
    if a.PollInterval < 0 {
        return false, fmt.Errorf("poll interval must be positive")
    }
    return true, nil
}
