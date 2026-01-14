package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kilovoltov/kilometrix/internal/handlers"
	"github.com/kilovoltov/kilometrix/internal/repository"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	parseFlags()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt, // syscall.SIGINT
		syscall.SIGTERM,
	)
	defer stop()

	if err := LoggerInitialize(logLevel); err != nil {
		fmt.Println(err)
	}
	defer Log.Info("Saving data")

	var mStor repository.MetricStorage

	if dsn == "" {
		if storFilePath == "" {
			mStor = repository.NewMemStorage()
		} else {
			mStor = repository.NewFileStorage(storFilePath, time.Duration(storInterval)*time.Second, restore)
		}
	} else {
		mStor = repository.NewDBStorage(dsn)
	}
	if err := mStor.InitStorage(); err != nil {
		Log.Fatal("Init Storage Error", zap.Error(err))
	}
	stor := handlers.NewStor(mStor)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", requestLogger(gzipMiddleware(stor.HandleMetricUpdate)))
	r.Post("/update", gzipMiddleware(stor.HandleMetricUpdateJSON))
	r.Post("/update/", requestLogger(gzipMiddleware(stor.HandleMetricUpdateJSON)))
	r.Post("/updates/", requestLogger(gzipMiddleware(stor.HandleMetricsUpdateJSON)))
	r.Post("/value", requestLogger(gzipMiddleware(stor.HandleValueJSON)))
	r.Post("/value/", requestLogger(gzipMiddleware(stor.HandleValueJSON)))
	r.Get("/value/{metricType}/{metricName}", requestLogger(gzipMiddleware(stor.HandleMetricGet)))
	r.Get("/", requestLogger(gzipMiddleware(stor.HandleMain)))
	r.Get("/ping", requestLogger(stor.HandleCheckStorage))

	// Запуск сервера в фоне (graceful sutdown)
	go func() {
		if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка запуска сервера: %v", err)
		}
	}()

	fmt.Printf("Server started at http://%s\nParameters: %v\n, filepath: %s, interval: %d\n", addr, os.Args, storFilePath, storInterval)
	fmt.Println("Press Ctrl+C to interrupt.")

	<-ctx.Done()

	fmt.Println("Server was interrupted")

	mStor.CloseStorage()
}
