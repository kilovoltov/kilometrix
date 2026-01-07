package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	addr         string
	logLevel     string
	storFilePath string
	storInterval int
	restore      bool
)

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "address of the server")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.IntVar(&storInterval, "i", 15, "stor interval, sec")
	flag.StringVar(&storFilePath, "f", "./metrics.json", "filestorage path")
	flag.BoolVar(&restore, "r", false, "restore from file")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
	if envStorInterval := os.Getenv("STORE_INTERVAL"); envStorInterval != "" {
		interval, err := strconv.Atoi(envStorInterval)
		if err != nil {
			fmt.Printf("Ошбика преобразования STORE_INTERVAL: %v", err)
		}
		storInterval = interval
	}
	if envStorFilePath := os.Getenv("FILE_STORAGE_PATH"); envStorFilePath != "" {
		storFilePath = envStorFilePath
	}
	if _, ok := os.LookupEnv("RESTORE"); ok {
		restore = true
	}
}
