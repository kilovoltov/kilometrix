package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

var (
	addr           string
	pollInterval   int64
	reportInterval int
	secretKey      string
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	SecretKey      string `env:"KEY"`
}

// параметры запуска: адрес сервера, интервалы сбора и отправки метрик
func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "address of the server")
	flag.Int64Var(&pollInterval, "p", 2, "poll interval, sec")
	flag.IntVar(&reportInterval, "r", 10, "report interval, sec")
	flag.StringVar(&secretKey, "k", "secretstring", "secretkey for hash")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// если есть переменные окружения, то они перезаписывают значения флагов
	if cfg.Address != "" {
		addr = cfg.Address
	}
	if cfg.PollInterval != 0 {
		pollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval != 0 {
		reportInterval = cfg.ReportInterval
	}
	if cfg.SecretKey != "" {
		secretKey = cfg.SecretKey
	}
}
