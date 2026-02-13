package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kilovoltov/kilometrix/internal/models"
	"github.com/kilovoltov/kilometrix/internal/utils"
	"go.uber.org/zap"
)

type DBStorage struct {
	mem *MemStorage
	dsn string
	DB  *sql.DB
	logger *zap.Logger
}

// NewDBStorage конструктор DBSrorage
func NewDBStorage(dsn string, logger *zap.Logger) *DBStorage {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	// Создаём приложение с зависимостями
	database := &DBStorage{
		mem: NewMemStorage(logger),
		dsn: dsn,
		DB:  db,
	}
	return database
}

func (db *DBStorage) InitStorage() error {
	fmt.Println("Started with database storage")
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
	_, err := db.execWithConnectionRetry(`
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
	_, err := db.execWithConnectionRetry(`
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
			counter_value = CASE
				WHEN EXCLUDED.counter_value IS NOT NULL
				THEN storage.metrics.counter_value + EXCLUDED.counter_value
				ELSE storage.metrics.counter_value
			END;`

	metrics = utils.RemoveDuplicates(metrics)
	valueStrings := make([]string, 0, len(metrics))
	args := make([]any, 0, len(metrics)*3) // 3 параметра для каждой метрики, 4-е - NULL

	for i, v := range metrics {
		switch v.MType {
		case "gauge":
			if v.Value == nil {
				return fmt.Errorf("gauge value is nil for metric %s", v.ID)
			}
			valueStrings = append(
				valueStrings,
				fmt.Sprintf("($%d, $%d, $%d, NULL)", i*3+1, i*3+2, i*3+3),
			)
			args = append(args, v.ID, v.MType, *v.Value)
		case "counter":
			if v.Delta == nil {
				return fmt.Errorf("counter delta is nil for metric %s", v.ID)
			}
			valueStrings = append(
				valueStrings,
				fmt.Sprintf("($%d, $%d, NULL, $%d)", i*3+1, i*3+2, i*3+3),
			)
			args = append(args, v.ID, v.MType, *v.Delta)
		default:
			return fmt.Errorf("unknown metric type: %s", v.MType)
		}
	}

	queryString = fmt.Sprintf(queryString, strings.Join(valueStrings, ","))

	_, err := db.execWithConnectionRetry(queryString, args...)
	return err
}

// execWithConnectionRetry выполняет Exec с ретраями при ошибках подключения (SQLSTATE Class 08)
func (db *DBStorage) execWithConnectionRetry(queryString string, args ...any) (sql.Result, error) {
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second}
	for attempt := 0; ; attempt++ {
		result, err := db.DB.Exec(queryString, args...)
		if err == nil {
			return result, nil // успех
		} else {
			fmt.Printf("%s\n", err)
		}

		// Пытаемся извлечь *pgconn.PgError через errors.As
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) { //  || pgerrcode.UndefinedTable == pgErr.Code
			// Это ошибка подключения — пробуем повторить, если остались попытки
			if attempt < len(backoff) {
				time.Sleep(backoff[attempt])
				fmt.Printf("Next attempt")
				continue
			}
		} else {
			return result, err
		}

		// Либо это не ошибка подключения, либо исчерпаны попытки
		return result, err
	}
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
		fmt.Printf("Database ping failed: %v\n", err)
	} else {
		db.mem.logger.Info("Database ping is OK")
	}
	return err
}
