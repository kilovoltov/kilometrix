package main

import (
    "flag"
    "fmt"
    "net/http"
    "os"
    "github.com/kilovoltov/kilometrix/internal/repository"
    "github.com/kilovoltov/kilometrix/internal/handlers"

    "github.com/go-chi/chi/v5"
)

func main() {
    // переменная для адреса сервера со значением по умолчанию
    var addr string
    flag.StringVar(&addr, "a", "localhost:8080", "address of the server")

    flag.Parse()
    
    // если есть переменная окружения, то она перезаписывает значение флага
    if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
        addr = envRunAddr
    }

    memStor := repository.NewMemStorage()
    stor := handlers.NewStor(memStor)

    r := chi.NewRouter()
    r.Post("/update/{metricType}/{metricName}/{metricValue}", stor.HandleMetricUpdate)
    r.Get("/value/{metricType}/{metricName}", stor.HandleMetricGet)
    r.Get("/", stor.HandleMain)

    fmt.Printf("Server started at http://%s\n", addr)
    err := http.ListenAndServe(addr, r)
    if err != nil {
        panic(err)
    }
}
