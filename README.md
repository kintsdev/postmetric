# PostMetric

This application is a Prometheus exporter for PostgreSQL, designed to collect and expose a wide range of PostgreSQL metrics. It provides detailed insights into PostgreSQL performance, including connections, queries, table statistics, replication, and more.

## Features

The exporter collects and exposes the following metrics:

### Connection Metrics
- **Connection usage percentage**
- **Idle connections**
- **Long-running queries**

### Query Metrics
- **Active and waiting queries**

### Table Metrics
- **Live tuples**
- **Dead tuples**
- **Table size**
- **Index size**

### Write-Ahead Log (WAL) Metrics
- **WAL size in bytes**
- **WAL files count**

### Buffer Cache Metrics
- **Cache hits and reads**
- **Total buffer usage**

### Replication Metrics
- **Replication lag in bytes**

### Transaction Metrics
- **Commit rate**
- **Rollback transactions**

### Other Metrics
- **Table bloat**
- **Autovacuum workers**
- **Temporary table size**
- **Checkpoint write time**
- **Deadlocks**
- **Locks**

### User Metrics
- **User connections**
- **User queries**

### Database Metrics
- **Database transaction rate**

## Prerequisites

- A PostgreSQL instance with sufficient permissions to access `pg_stat` and related views.
- A Prometheus server to scrape and visualize the metrics.
- Go 1.16 or later (for building the application).

## Setup

Follow these steps to set up PostMetric:

1. **Clone the Repository**
   ```bash
   git clone git@github.com:kintsdev/postmetric.git
   cd postmetric
   ```

2. **Build the Application**
   ```bash
   go build -o postmetric
   ```

3. **Configure the PostgreSQL Connection**
   Set the PostgreSQL connection string as an environment variable:
   ```bash
   export POSTGRES_CONNECTION_STRING="postgresql://user:password@host:port/dbname?sslmode=disable"
   ```

4. **Run the Application**
   ```bash
   ./postmetric
   ```

## Prometheus Configuration

Add the following scrape configuration to your Prometheus configuration file:

```yaml
scrape_configs:
  - job_name: 'postmetric'
    static_configs:
      - targets: ['localhost:8080']
```

## Accessing Metrics

PostMetric exposes metrics at:

```
http://localhost:8080/metrics
```

## PostgreSQL Metrics Overview

Below is an overview of the metrics exposed by PostMetric:

### Connection Metrics
- **`postgres_connection_usage`**: Percentage of used connections out of the maximum connections.
- **`postgres_idle_connections`**: Number of idle connections.

### Query Metrics
- **`postgres_active_queries`**: Number of currently active queries.
- **`postgres_waiting_queries`**: Number of queries waiting for locks.
- **`postgres_long_running_queries`**: Queries running for more than 5 minutes.

### Table Statistics
- **`postgres_table_stats`**: Includes the following statistics:
  - Live tuples
  - Dead tuples
  - Total size
  - Table size
  - Index size
  - *(Labeled by `table_name` and `stat_type`)*

### Replication Metrics
- **`postgres_replication_lag`**: Replication lag in bytes.

### WAL Metrics
- **`postgres_wal_stats`**: Includes:
  - WAL size
  - WAL file count

### Cache Metrics
- **`postgres_buffer_cache`**: Includes:
  - Cache hits
  - Cache reads
  - Total usage

### Transaction Metrics
- **`postgres_commit_rate`**: Percentage of committed transactions.
- **`postgres_rollback_transactions`**: Number of rolled-back transactions.

### Other Metrics
- **`postgres_table_bloat_bytes`**: Estimated bloat size in bytes.
- **`postgres_autovacuum_workers`**: Number of active autovacuum workers.
- **`postgres_temp_table_size_bytes`**: Total size of temporary tables.
- **`postgres_checkpoint_write_time_ms`**: Average checkpoint write time in milliseconds.
- **`postgres_deadlocks`**: Number of deadlocks detected.
- **`postgres_locks`**: Number of active locks.

### User Metrics
- **`postgres_user_connections`**: Number of connections per user.
- **`postgres_user_queries`**: Number of queries executed per user.

### Database Metrics
- **`postgres_database_transaction_rate`**: Transaction rate per database.

---

PostMetric provides a comprehensive set of metrics to help you monitor and optimize PostgreSQL performance effectively. Happy monitoring!

## Package Structure

The codebase is organized into the following packages:

- **`main`**: The entry point of the application.
- **`metrics`**: Contains metric definitions and registration logic.
- **`collector`**: Contains the logic for collecting metrics from PostgreSQL.

## Adding New Metrics

To add new metrics, follow these steps:

1. **Define the Metric**: Add the new metric definition in `pkg/metrics/metrics.go`.
2. **Register the Metric**: Ensure the new metric is registered in the `RegisterMetrics` function in `pkg/metrics/metrics.go`.
3. **Collect the Metric**: Update the `CollectMetrics` function in `pkg/metrics/collector.go` to collect the new metric.

