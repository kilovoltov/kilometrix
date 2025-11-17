package main

import (
	"fmt"
	"net/http"

	"github.com/kilovoltov/kilometrix/internal/handlers"
	"github.com/kilovoltov/kilometrix/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	parseFlags()
	if err := LoggerInitialize(logLevel); err != nil {
		panic(err)
	}
	memStor := repository.NewMemStorage()
	stor := handlers.NewStor(memStor)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", requestLogger(gzipMiddleware(stor.HandleMetricUpdate)))
	r.Post("/update", requestLogger(gzipMiddleware(stor.HandleMetricUpdateJSON)))
	r.Post("/update/", requestLogger(gzipMiddleware(stor.HandleMetricUpdateJSON)))
	r.Post("/value", requestLogger(gzipMiddleware(stor.HandleValueJSON)))
	r.Post("/value/", requestLogger(gzipMiddleware(stor.HandleValueJSON)))
	r.Get("/value/{metricType}/{metricName}", requestLogger(gzipMiddleware(stor.HandleMetricGet)))
	r.Get("/", requestLogger(gzipMiddleware(stor.HandleMain)))

	fmt.Printf("Server started at http://%s\n", addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
