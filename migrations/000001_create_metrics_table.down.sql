-- migrations/000001_create_metrics_table.down.sql
DROP TABLE IF EXISTS storage.metrics;
DROP TYPE IF EXISTS storage.metric_kind;