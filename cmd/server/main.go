package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kilovoltov/kilometrix/internal/handlers"
	"github.com/kilovoltov/kilometrix/internal/repository"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// App — структура для хранения зависимостей (например, БД)
type App struct {
	DB *sql.DB
}

func (a *App) pingDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.DB.Ping(); err != nil {
		fmt.Printf("Database ping failed: %v", err)
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
}

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

	db, dberr := sql.Open("pgx", dsn)
	if dberr != nil {
		panic(dberr)
	}
	defer db.Close()

	// Проверка соединения при старте
	if err := db.Ping(); err != nil {
		fmt.Printf("Failed to ping database on startup: %v", err)
	}
	// Создаём приложение с зависимостями
	app := &App{DB: db}

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
	r.Get("/ping", requestLogger(app.pingDatabaseHandler))

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

	memStor.SaveToFile()
}
