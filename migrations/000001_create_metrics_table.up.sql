-- migrations/000001_create_metrics_table.up.sql
-- Создание cхемы
CREATE SCHEMA IF NOT EXISTS storage;

-- Создание типа metric_kind
CREATE TYPE storage.metric_kind AS ENUM ('gauge', 'counter');

-- Создание таблицы метрик
CREATE TABLE IF NOT EXISTS storage.metrics(
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL,
    metric_type storage.metric_kind NOT NULL,
    gauge_value DOUBLE PRECISION,
    counter_value BIGINT,
    CONSTRAINT chk_metric_type_value CHECK (
        (metric_type = 'gauge' AND gauge_value IS NOT NULL AND counter_value IS NULL)
        OR
        (metric_type = 'counter' AND counter_value IS NOT NULL AND gauge_value IS NULL)
    ),
    CONSTRAINT uk_metric_name_type UNIQUE (metric_name, metric_type)
);

-- Создание индекса
CREATE INDEX IF NOT EXISTS idx_metrics_name ON storage.metrics(metric_name);