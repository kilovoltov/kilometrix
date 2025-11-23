package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	addr         string
	logLevel     string
	storFilePath string
	storInterval int
)

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "address of the server")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.IntVar(&storInterval, "i", 300, "stor interval, sec")
	flag.StringVar(&storFilePath, "f", "/tmp/metrics.json", "filestorage path")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
	if envStorInterval := os.Getenv("STOR_INTERVAL"); envStorInterval != "" {
		interval, err := strconv.Atoi(envStorInterval)
		if err != nil {
			panic(err)
		}
		storInterval = interval
	}
	if envStorFilePath := os.Getenv("FILE_STORAGE_PATH"); envStorFilePath != "" {
		storFilePath = envStorFilePath
	}
}
