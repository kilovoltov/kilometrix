package main

import (
	// "context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
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
	if err := LoggerInitialize(logLevel); err != nil {
		fmt.Println(err)
	}
	defer fmt.Println("STOPe!")

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

	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    // defer cancel()
    // if dberr = db.PingContext(ctx); dberr != nil {
    //     panic(dberr)
    // }

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
	r.Get("/ping", app.pingDatabaseHandler)

	fmt.Printf("Server started at http://%s\nParameters: %v\n, filepath: %s, interval: %d\n", addr, os.Args, storFilePath, storInterval)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v", err)
	}
}
