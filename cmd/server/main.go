package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kilovoltov/kilometrix/internal/handlers"
	"github.com/kilovoltov/kilometrix/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	parseFlags()
	if err := LoggerInitialize(logLevel); err != nil {
		fmt.Println(err)
	}
	defer fmt.Println("STOPe!")

	// memStor := repository.NewMemStorage()
	memStor := repository.NewFileStorage(storFilePath, time.Duration(storInterval)*time.Second)

	if restore {
		fmt.Println("Loaded from file")
		memStor.LoadFromFile()
	}

	stor := handlers.NewStor(memStor)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", requestLogger(gzipMiddleware(stor.HandleMetricUpdate)))
	r.Post("/update", gzipMiddleware(stor.HandleMetricUpdateJSON))
	r.Post("/update/", requestLogger(gzipMiddleware(stor.HandleMetricUpdateJSON)))
	r.Post("/value", requestLogger(gzipMiddleware(stor.HandleValueJSON)))
	r.Post("/value/", requestLogger(gzipMiddleware(stor.HandleValueJSON)))
	r.Get("/value/{metricType}/{metricName}", requestLogger(gzipMiddleware(stor.HandleMetricGet)))
	r.Get("/", requestLogger(gzipMiddleware(stor.HandleMain)))

	fmt.Printf("Server started at http://%s\nParameters: %v\n, filepath: %s, interval: %d\n", addr, os.Args, storFilePath, storInterval)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v", err)
	}
}
