package main

import (
    "fmt"
    "net/http"
    "github.com/kilovoltov/kilometrix/internal/repository"
    "github.com/kilovoltov/kilometrix/internal/handlers"

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
    r.Post("/update/{metricType}/{metricName}/{metricValue}", requestLogger(stor.HandleMetricUpdate))
	r.Post("/update", stor.HandleMetricUpdateJson)
	r.Post("/value", stor.HandleValueJson)
    r.Get("/value/{metricType}/{metricName}", requestLogger(stor.HandleMetricGet))
    r.Get("/", requestLogger(stor.HandleMain))

    fmt.Printf("Server started at http://%s\n", addr)
    err := http.ListenAndServe(addr, r)
    if err != nil {
        panic(err)
    }
}
