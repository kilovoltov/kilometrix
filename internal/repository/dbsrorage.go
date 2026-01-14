package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kilovoltov/kilometrix/internal/models"
)

type DBStorage struct {
	mem *MemStorage
	dsn string
	DB  *sql.DB
}

// NewDBStorage конструктор DBSrorage
func NewDBStorage(dsn string) *DBStorage {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	// defer db.Close()

	// Создаём приложение с зависимостями
	database := &DBStorage{
		mem: NewMemStorage(),
		dsn: dsn,
		DB:  db,
	}
	return database
}

func (db *DBStorage) InitStorage() error {
	// Создаём экземпляр migrate
	m, err := migrate.New(
		"file://./migrations", // путь к папке с миграциями
		db.dsn,
	)
	if err != nil {
		return err
	}

	// Выполняем все неприменённые миграции вверх
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) { // err != migrate.ErrNoChange
		return err
	}

	fmt.Println("Миграции успешно применены")
	m.Close()
	return db.CheckStorage()
}

func (db *DBStorage) CloseStorage() error {
	defer db.DB.Close()
	return nil
}

func (db *DBStorage) GetMetricNamesByType(t string) []string {
	var names []string
	rows, err := db.DB.Query("SELECT metric_name FROM storage.metrics WHERE metric_type = $1", t)
	if err != nil {
		fmt.Printf("ERROR GetGaugesNames from DB storage: %v", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			fmt.Printf("ERROR GetGaugesNames rows.Next(): %v", err)
		}

		names = append(names, name)
	}

	err = rows.Err()
	if err != nil {
		fmt.Printf("ERROR GetGaugesNames rows.Err(): %v", err)
	}

	return names
}

func (db *DBStorage) GetGaugesNames() []string {
	return db.GetMetricNamesByType("gauge")
}

func (db *DBStorage) GetCounterNames() []string {
	return db.GetMetricNamesByType("counter")
}

// AddGauge Добавление метрики типа Gauge в хранилище
func (db *DBStorage) AddGauge(name string, value float64) error {
	_, err := db.DB.Exec(`
	INSERT INTO storage.metrics (metric_name, metric_type, gauge_value, counter_value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (metric_name, metric_type)
	DO UPDATE SET
		gauge_value = EXCLUDED.gauge_value;`,
		name, "gauge", value, nil,
	)

	return err
}

// AddCounter Добавление метрики типа Counter в хранилище
func (db *DBStorage) AddCounter(name string, value int64) error {
	_, err := db.DB.Exec(`
	INSERT INTO storage.metrics (metric_name, metric_type, gauge_value, counter_value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (metric_name, metric_type)
	DO UPDATE SET
		counter_value = storage.metrics.counter_value + EXCLUDED.counter_value;`,
		name, "counter", nil, value,
	)

	return err
}

// GetGauge Получение значения метрики по имени
func (db *DBStorage) GetGauge(name string) (float64, error) {
	var v float64
	row := db.DB.QueryRow(
		"SELECT gauge_value FROM storage.metrics WHERE metric_name = $1 AND metric_type = 'gauge'", name)
	// порядок переменных должен соответствовать порядку колонок в запросе
	err := row.Scan(&v)
	if err != nil {
		fmt.Printf("ERROR GetGauge row.Scan(): %v", err)
	}
	return v, err
}

// GetCounter Получение значения метрики по имени
func (db *DBStorage) GetCounter(name string) (int64, error) {
	var v int64
	row := db.DB.QueryRow(
		"SELECT counter_value FROM storage.metrics WHERE metric_name = $1 AND metric_type = 'counter'", name)
	// порядок переменных должен соответствовать порядку колонок в запросе
	err := row.Scan(&v)
	if err != nil {
		fmt.Printf("ERROR GetCounter row.Scan(): %v", err)
	}
	return v, err
}

func (db *DBStorage) AddMetrics(metrics []models.Metrics) error {
	queryString := `
		INSERT INTO storage.metrics (metric_name, metric_type, gauge_value, counter_value)
		VALUES %s
		ON CONFLICT (metric_name, metric_type)
		DO UPDATE SET
			gauge_value = EXCLUDED.gauge_value,
			counter_value = storage.metrics.counter_value + EXCLUDED.counter_value;`

	valueStrings := make([]string, 0, len(metrics))

	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			valueStrings = append(
				valueStrings,
				fmt.Sprintf("('%s', '%s', %g, %s)", v.ID, v.MType, *v.Value, "NULL"),
			)
		case "counter":
			valueStrings = append(
				valueStrings,
				fmt.Sprintf("('%s', '%s', %s, %d)", v.ID, v.MType, "NULL", *v.Delta),
			)
		}
	}

	queryString = fmt.Sprintf(queryString, strings.Join(valueStrings, ","))

	_, err := db.DB.Exec(queryString)

	return err
}

func (db *DBStorage) Snapshot() []models.Metrics {
	snapshot := make([]models.Metrics, 0, 30)
	rows, err := db.DB.Query("SELECT metric_name, metric_type, gauge_value, counter_value FROM storage.metrics")
	if err != nil {
		fmt.Printf("ERROR dbSnapshot query: %v", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
		if err != nil {
			fmt.Printf("ERROR dbSnapshot rows.Next(): %v", err)
		}

		snapshot = append(snapshot, metric)
	}

	err = rows.Err()
	if err != nil {
		fmt.Printf("ERROR dbSnapshot rows.Err(): %v", err)
	}
	return snapshot
}

func (db *DBStorage) CheckStorage() error {
	err := db.DB.Ping()
	if err != nil {
		fmt.Printf("Database ping failed: %v", err)
	}
	return err
}
