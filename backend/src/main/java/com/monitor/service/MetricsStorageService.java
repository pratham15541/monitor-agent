package com.monitor.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
public class MetricsStorageService {

    private static final Logger logger = LoggerFactory.getLogger(MetricsStorageService.class);

    private final JdbcTemplate jdbcTemplate;
    private final int metricRetentionDays;
    private final int metricDetailRetentionDays;
    private volatile boolean timescaleEnabled = false;

    public MetricsStorageService(
            JdbcTemplate jdbcTemplate,
            @Value("${app.metrics.retentionDays:30}") int metricRetentionDays,
            @Value("${app.metrics.detailRetentionDays:7}") int metricDetailRetentionDays) {
        this.jdbcTemplate = jdbcTemplate;
        this.metricRetentionDays = Math.max(metricRetentionDays, 1);
        this.metricDetailRetentionDays = Math.max(metricDetailRetentionDays, 1);
    }

    @EventListener(ApplicationReadyEvent.class)
    public void initializeStorage() {
        try {
            jdbcTemplate.execute("CREATE EXTENSION IF NOT EXISTS timescaledb");
            normalizeTimestampColumns();
            prepareMetricTablesForHypertables();
            jdbcTemplate.execute("SELECT create_hypertable('metric', 'created_at', if_not_exists => TRUE, migrate_data => TRUE)");
            jdbcTemplate.execute(
                    "SELECT create_hypertable('metric_detail', 'created_at', if_not_exists => TRUE, migrate_data => TRUE)");
            createMetricIndexes();
            jdbcTemplate.execute(
                    "SELECT add_retention_policy('metric', INTERVAL '" + metricRetentionDays
                            + " days', if_not_exists => TRUE)");
            jdbcTemplate.execute(
                    "SELECT add_retention_policy('metric_detail', INTERVAL '" + metricDetailRetentionDays
                            + " days', if_not_exists => TRUE)");
            timescaleEnabled = true;
            logger.info("TimescaleDB hypertables and retention policies enabled");
        } catch (Exception ex) {
            timescaleEnabled = false;
            logger.warn("TimescaleDB not available, falling back to scheduled deletes", ex);
        }
    }

    private void normalizeTimestampColumns() {
        alterTimestampColumn("company", "created_at");
        alterTimestampColumn("device", "created_at");
        alterTimestampColumn("device", "last_seen_at");
        alterTimestampColumn("metric", "created_at");
        alterTimestampColumn("metric_detail", "created_at");
    }

    private void alterTimestampColumn(String tableName, String columnName) {
        jdbcTemplate.execute("""
                DO $$
                BEGIN
                    IF EXISTS (
                        SELECT 1
                        FROM information_schema.columns
                        WHERE table_schema = 'public'
                          AND table_name = '%s'
                          AND column_name = '%s'
                          AND data_type = 'timestamp without time zone'
                    ) THEN
                        ALTER TABLE %s
                        ALTER COLUMN %s TYPE TIMESTAMPTZ
                        USING %s AT TIME ZONE 'UTC';
                    END IF;
                END $$;
                """.formatted(tableName, columnName, tableName, columnName, columnName));
    }

    private void prepareMetricTablesForHypertables() {
        dropPrimaryKeyIfExists("metric");
        dropPrimaryKeyIfExists("metric_detail");
    }

    private void dropPrimaryKeyIfExists(String tableName) {
        jdbcTemplate.execute("""
                DO $$
                DECLARE
                    constraint_name text;
                BEGIN
                    SELECT conname INTO constraint_name
                    FROM pg_constraint
                    WHERE conrelid = '%s'::regclass
                      AND contype = 'p';

                    IF constraint_name IS NOT NULL THEN
                        EXECUTE format('ALTER TABLE %%I DROP CONSTRAINT %%I', '%s', constraint_name);
                    END IF;
                END $$;
                """.formatted(tableName, tableName));
    }

    private void createMetricIndexes() {
        jdbcTemplate.execute("CREATE INDEX IF NOT EXISTS idx_metric_device_created_at ON metric (device_id, created_at DESC)");
        jdbcTemplate.execute(
                "CREATE INDEX IF NOT EXISTS idx_metric_detail_device_created_at ON metric_detail (device_id, created_at DESC)");
    }

    @Scheduled(cron = "0 30 2 * * *")
    public void enforceRetentionFallback() {
        if (timescaleEnabled) {
            return;
        }

        int metricDeleted = jdbcTemplate.update(
                "DELETE FROM metric WHERE created_at < now() - (? || ' days')::interval",
                metricRetentionDays);
        int detailDeleted = jdbcTemplate.update(
                "DELETE FROM metric_detail WHERE created_at < now() - (? || ' days')::interval",
                metricDetailRetentionDays);

        if (metricDeleted > 0 || detailDeleted > 0) {
            logger.info("Retention cleanup removed {} metric rows and {} metric_detail rows", metricDeleted,
                    detailDeleted);
        }
    }
}
