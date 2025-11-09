package main

import (
	"flag"
	"os"
)

var (
	addr     string
	logLevel string
)

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "address of the server")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
}
