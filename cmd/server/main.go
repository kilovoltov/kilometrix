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

	logger, err := NewLogger(logLevel)
	if err != nil {
		panic(err)
	}

	logger.Info("Logger initialized successfully")
	defer logger.Sync()

	var mStor repository.MetricStorage

	if dsn == "" {
		if storFilePath == "" {
			mStor = repository.NewMemStorage(logger)
		} else {
			mStor = repository.NewFileStorage(storFilePath, time.Duration(storInterval)*time.Second, restore, logger)
		}
	} else {
		mStor = repository.NewDBStorage(dsn, logger)
		defer mStor.CloseStorage()
	}
	if err := mStor.InitStorage(); err != nil {
		logger.Fatal("Init Storage Error", zap.Error(err))
	}
	stor := handlers.NewStor(mStor)
	r := chi.NewRouter()
	r.Use(ErrorLoggerMiddleware(logger))
	r.Use(RequestLoggerMiddleware(logger))
	r.Use(gzipMiddleware())
	r.Use(CheckHash(logger, secretKey))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", stor.HandleMetricUpdate)
	r.Post("/update", stor.HandleMetricUpdateJSON)
	r.Post("/update/", stor.HandleMetricUpdateJSON)
	r.Post("/updates/", stor.HandleMetricsUpdateJSON)
	r.Post("/value", stor.HandleValueJSON)
	r.Post("/value/", stor.HandleValueJSON)
	r.Get("/value/{metricType}/{metricName}", stor.HandleMetricGet)
	r.Get("/", stor.HandleMain)
	r.Get("/ping", stor.HandleCheckStorage)

	// Запуск сервера в фоне (graceful sutdown)
	go func() {
		if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка запуска сервера: %v", err)
		}
	}()

	fmt.Printf("Server started at http://%s\nParameters: %v\n, filepath: %s, interval: %d\n", addr, os.Args, storFilePath, storInterval)
	fmt.Println("Press Ctrl+C to interrupt.")

	<-ctx.Done()

	logger.Info("Server was interrupted")
	logger.Info("Saving data")

	mStor.CloseStorage()
}
