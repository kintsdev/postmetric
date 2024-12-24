package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	connectionUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_connection_usage",
		Help: "Percentage of used connections out of max_connections.",
	})

	idleConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_idle_connections",
		Help: "Number of idle connections.",
	})

	longRunningQueries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_long_running_queries",
		Help: "Number of queries running for more than 5 minutes.",
	})

	replicationLag = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_replication_lag",
		Help: "Replication lag in bytes.",
	})

	totalSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_total_size_bytes",
		Help: "Total database size in bytes.",
	})

	indexSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_index_size_bytes",
		Help: "Total index size in bytes.",
	})

	diskReads = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_disk_reads",
		Help: "Number of disk blocks read.",
	})

	cacheHits = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_cache_hits",
		Help: "Number of cache hits.",
	})

	commitRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_commit_rate",
		Help: "Transaction commit rate.",
	})

	rollbackTransactions = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_rollback_transactions",
		Help: "Number of rollback transactions.",
	})

	tableBloat = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_table_bloat_bytes",
		Help: "Bloat size in bytes for tables.",
	})

	autovacuumWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_autovacuum_workers",
		Help: "Number of active autovacuum workers.",
	})

	tempTableSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_temp_table_size_bytes",
		Help: "Total size of temporary tables in bytes.",
	})

	checkpointWriteTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_checkpoint_write_time_ms",
		Help: "Average checkpoint write time in milliseconds.",
	})

	deadlocks = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_deadlocks",
		Help: "Number of deadlocks detected.",
	})

	locks = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_locks",
		Help: "Number of active locks.",
	})

	activeQueries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_active_queries",
		Help: "Number of currently active queries.",
	})

	waitingQueries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_waiting_queries",
		Help: "Number of queries waiting for locks.",
	})

	tableStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_table_stats",
			Help: "Table statistics including row count, size, and dead tuples.",
		},
		[]string{"table_name", "stat_type"},
	)

	indexStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_index_stats",
			Help: "Index statistics including size and scans.",
		},
		[]string{"index_name", "table_name", "stat_type"},
	)

	vacuumProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_vacuum_progress",
			Help: "Progress of vacuum operations.",
		},
		[]string{"table_name", "phase"},
	)

	bufferCacheStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_buffer_cache",
			Help: "Buffer cache statistics.",
		},
		[]string{"type"},
	)

	walStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_wal_stats",
			Help: "WAL statistics including size and writes.",
		},
		[]string{"type"},
	)

	statementStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_statement_stats",
			Help: "Statement execution statistics.",
		},
		[]string{"type"},
	)

	bgwriterStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_bgwriter_stats",
			Help: "Background writer statistics.",
		},
		[]string{"type"},
	)

	tableIOStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_table_io_stats",
			Help: "Table I/O statistics.",
		},
		[]string{"table_name", "type"},
	)

	replicationSlotStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_replication_slot_stats",
			Help: "Replication slot statistics.",
		},
		[]string{"slot_name", "type"},
	)

	userConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_user_connections",
			Help: "Number of connections per user.",
		},
		[]string{"user"},
	)

	userQueries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_user_queries",
			Help: "Number of queries executed per user.",
		},
		[]string{"user"},
	)

	databaseTransactionRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_database_transaction_rate",
			Help: "Transaction rate per database.",
		},
		[]string{"database"},
	)

	port = TenaryString(os.Getenv("PORT") == "", "8080", os.Getenv("PORT"))
	host = TenaryString(os.Getenv("HOST") == "", "", os.Getenv("HOST"))
)

func TenaryString(value bool, v1 string, v2 string) string {
	if value {
		return v1
	}
	return v2
}

func init() {
	prometheus.MustRegister(connectionUsage)
	prometheus.MustRegister(idleConnections)
	prometheus.MustRegister(longRunningQueries)
	prometheus.MustRegister(replicationLag)
	prometheus.MustRegister(totalSize)
	prometheus.MustRegister(indexSize)
	prometheus.MustRegister(diskReads)
	prometheus.MustRegister(cacheHits)
	prometheus.MustRegister(commitRate)
	prometheus.MustRegister(rollbackTransactions)
	prometheus.MustRegister(tableBloat)
	prometheus.MustRegister(autovacuumWorkers)
	prometheus.MustRegister(tempTableSize)
	prometheus.MustRegister(checkpointWriteTime)
	prometheus.MustRegister(deadlocks)
	prometheus.MustRegister(locks)
	prometheus.MustRegister(activeQueries)
	prometheus.MustRegister(waitingQueries)
	prometheus.MustRegister(tableStats)
	prometheus.MustRegister(indexStats)
	prometheus.MustRegister(vacuumProgress)
	prometheus.MustRegister(bufferCacheStats)
	prometheus.MustRegister(walStats)
	prometheus.MustRegister(statementStats)
	prometheus.MustRegister(bgwriterStats)
	prometheus.MustRegister(tableIOStats)
	prometheus.MustRegister(replicationSlotStats)
	prometheus.MustRegister(userConnections)
	prometheus.MustRegister(userQueries)
	prometheus.MustRegister(databaseTransactionRate)
}

func collectMetrics(ctx context.Context, pool *pgxpool.Pool) {
	// Collect active and waiting queries

	var activeCount, waitingCount int
	err := pool.QueryRow(ctx, `
	SELECT 
		COUNT(*) FILTER (WHERE state = 'active'),
		COUNT(*) FILTER (WHERE wait_event_type IS NOT NULL)
	FROM pg_stat_activity
	WHERE backend_type = 'client backend'
`).Scan(&activeCount, &waitingCount)
	if err != nil {
		log.Printf("Error querying query stats: %v", err)
	} else {
		activeQueries.Set(float64(activeCount))
		waitingQueries.Set(float64(waitingCount))
	}

	// Collect table statistics
	rows, err := pool.Query(ctx, `
	SELECT 
		relname,
		n_live_tup,
		n_dead_tup,
		pg_total_relation_size(relid) as total_size,
		pg_table_size(relid) as table_size,
		pg_indexes_size(relid) as index_size
	FROM pg_stat_user_tables
`)
	if err != nil {
		log.Printf("Error querying table stats: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var tableName string
			var liveTuples, deadTuples, totalSize, tableSize, indexSize int64
			if err := rows.Scan(&tableName, &liveTuples, &deadTuples, &totalSize, &tableSize, &indexSize); err != nil {
				log.Printf("Error scanning table stats: %v", err)
				continue
			}
			tableStats.WithLabelValues(tableName, "live_tuples").Set(float64(liveTuples))
			tableStats.WithLabelValues(tableName, "dead_tuples").Set(float64(deadTuples))
			tableStats.WithLabelValues(tableName, "total_size_bytes").Set(float64(totalSize))
			tableStats.WithLabelValues(tableName, "table_size_bytes").Set(float64(tableSize))
			tableStats.WithLabelValues(tableName, "index_size_bytes").Set(float64(indexSize))
		}
	}

	// Collect WAL statistics
	var walBytes, walFiles int64
	err = pool.QueryRow(ctx, `
	SELECT 
		pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') as wal_bytes,
		count(*) as wal_files
	FROM pg_ls_waldir()
`).Scan(&walBytes, &walFiles)
	if err != nil {
		log.Printf("Error querying WAL stats: %v", err)
	} else {
		walStats.WithLabelValues("bytes").Set(float64(walBytes))
		walStats.WithLabelValues("files").Set(float64(walFiles))
	}

	// Collect buffer cache statistics
	var bufferHits, bufferReads, bufferEvictions int64
	err = pool.QueryRow(ctx, `
	SELECT 
		sum(heap_blks_hit) as buffer_hits,
		sum(heap_blks_read) as buffer_reads,
		sum(heap_blks_hit + heap_blks_read) as total_blocks
	FROM pg_statio_user_tables
`).Scan(&bufferHits, &bufferReads, &bufferEvictions)
	if err != nil {
		log.Printf("Error querying buffer cache stats: %v", err)
	} else {
		bufferCacheStats.WithLabelValues("hits").Set(float64(bufferHits))
		bufferCacheStats.WithLabelValues("reads").Set(float64(bufferReads))
		bufferCacheStats.WithLabelValues("total").Set(float64(bufferHits + bufferReads))
	}

	// Collect replication slot statistics
	rows, err = pool.Query(ctx, `
	SELECT 
		slot_name,
		pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) as lag_bytes,
		active
	FROM pg_replication_slots
`)
	if err != nil {
		log.Printf("Error querying replication slot stats: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var slotName string
			var lagBytes int64
			var active bool
			if err := rows.Scan(&slotName, &lagBytes, &active); err != nil {
				log.Printf("Error scanning replication slot stats: %v", err)
				continue
			}
			replicationSlotStats.WithLabelValues(slotName, "lag_bytes").Set(float64(lagBytes))
			if active {
				replicationSlotStats.WithLabelValues(slotName, "active").Set(1)
			} else {
				replicationSlotStats.WithLabelValues(slotName, "active").Set(0)
			}
		}
	}

	// Collect max_connections and connection usage
	var maxConnections int
	err = pool.QueryRow(ctx, "SELECT setting::int FROM pg_settings WHERE name = 'max_connections'").Scan(&maxConnections)
	if err != nil {
		log.Printf("Error querying max_connections: %v", err)
		return
	}

	var usedConnections int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity").Scan(&usedConnections)
	if err != nil {
		log.Printf("Error querying used connections: %v", err)
		return
	}

	connectionUsage.Set(float64(usedConnections) / float64(maxConnections) * 100)

	// Collect idle connections
	var idleConn int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'idle'").Scan(&idleConn)
	if err != nil {
		log.Printf("Error querying idle connections: %v", err)
		return
	}

	idleConnections.Set(float64(idleConn))

	// Collect long-running queries
	var longQueries int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '5 minutes'").Scan(&longQueries)
	if err != nil {
		log.Printf("Error querying long-running queries: %v", err)
		return
	}

	longRunningQueries.Set(float64(longQueries))

	// Collect replication lag
	var lagBytes int64
	err = pool.QueryRow(ctx, `
        SELECT CASE 
            WHEN pg_is_in_recovery() THEN 
                EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))::bigint
            ELSE 
                0 
        END as lag_seconds`).Scan(&lagBytes)
	if err != nil {
		log.Printf("Error querying replication lag: %v", err)
		return
	}
	replicationLag.Set(float64(lagBytes))

	// Collect total database size
	var sizeBytes int64
	err = pool.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&sizeBytes)
	if err != nil {
		log.Printf("Error querying total database size: %v", err)
		return
	}

	totalSize.Set(float64(sizeBytes))

	// Collect index size
	var idxSize int64
	err = pool.QueryRow(ctx, "SELECT sum(pg_indexes_size(indexrelid)) FROM pg_stat_user_indexes").Scan(&idxSize)
	if err != nil {
		log.Printf("Error querying index size: %v", err)
		return
	}

	indexSize.Set(float64(idxSize))

	// Collect disk reads and cache hits
	var diskRead, cacheHit int64
	err = pool.QueryRow(ctx, "SELECT sum(heap_blks_read), sum(heap_blks_hit) FROM pg_statio_user_tables").Scan(&diskRead, &cacheHit)
	if err != nil {
		log.Printf("Error querying disk reads and cache hits: %v", err)
		return
	}

	diskReads.Set(float64(diskRead))
	cacheHits.Set(float64(cacheHit))

	// Collect commit rate
	var commits, rollbacks int64
	err = pool.QueryRow(ctx, "SELECT sum(xact_commit), sum(xact_rollback) FROM pg_stat_database").Scan(&commits, &rollbacks)
	if err != nil {
		log.Printf("Error querying commit rate: %v", err)
		return
	}

	if (commits + rollbacks) > 0 {
		commitRate.Set(float64(commits) / float64(commits+rollbacks) * 100)
	}

	rollbackTransactions.Set(float64(rollbacks))

	// Collect table bloat
	var bloatSize int64
	err = pool.QueryRow(ctx, "SELECT sum(pg_total_relation_size(relid) - pg_relation_size(relid)) FROM pg_stat_user_tables").Scan(&bloatSize)
	if err != nil {
		log.Printf("Error querying table bloat: %v", err)
		return
	}

	tableBloat.Set(float64(bloatSize))

	// Collect autovacuum workers
	var autovacWorkers int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_progress_vacuum").Scan(&autovacWorkers)
	if err != nil {
		log.Printf("Error querying autovacuum workers: %v", err)
		return
	}

	autovacuumWorkers.Set(float64(autovacWorkers))

	var tempSize int64
	err = pool.QueryRow(ctx, "SELECT sum(temp_bytes) FROM pg_stat_database").Scan(&tempSize)
	if err != nil {
		log.Printf("Error querying temporary table size: %v", err)
		return
	}

	tempTableSize.Set(float64(tempSize))

	// Collect checkpoint write time
	var checkpointTime float64
	err = pool.QueryRow(ctx, "SELECT avg(checkpoint_write_time) FROM pg_stat_bgwriter").Scan(&checkpointTime)
	if err != nil {
		log.Printf("Error querying checkpoint write time: %v", err)
		return
	}

	checkpointWriteTime.Set(checkpointTime)

	// Collect deadlocks
	var totalDeadlocks int64
	err = pool.QueryRow(ctx, "SELECT sum(deadlocks) FROM pg_stat_database").Scan(&totalDeadlocks)
	if err != nil {
		log.Printf("Error querying deadlocks: %v", err)
		return
	}

	deadlocks.Set(float64(totalDeadlocks))

	// Collect locks
	var totalLocks int64
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_locks").Scan(&totalLocks)
	if err != nil {
		log.Printf("Error querying locks: %v", err)
		return
	}

	locks.Set(float64(totalLocks))

	// Collect user connections
	rows, err = pool.Query(ctx, `
	SELECT 
		usename,
		count(*)
	FROM pg_stat_activity
	WHERE backend_type = 'client backend'
	GROUP BY usename
`)
	if err != nil {
		log.Printf("Error querying user connections: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var user string
			var connections int
			if err := rows.Scan(&user, &connections); err != nil {
				log.Printf("Error scanning user connections: %v", err)
				continue
			}
			userConnections.WithLabelValues(user).Set(float64(connections))
		}
	}

	// Collect user queries
	rows, err = pool.Query(ctx, `
	SELECT 
		usename,
		count(*)
	FROM pg_stat_activity
	WHERE state = 'active'
	GROUP BY usename
`)
	if err != nil {
		log.Printf("Error querying user queries: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var user string
			var queries int
			if err := rows.Scan(&user, &queries); err != nil {
				log.Printf("Error scanning user queries: %v", err)
				continue
			}
			userQueries.WithLabelValues(user).Set(float64(queries))
		}
	}

	// Collect user transaction rate
	rows, err = pool.Query(ctx, `
	SELECT 
		datname,
		sum(xact_commit + xact_rollback) as transactions
	FROM pg_stat_database
	GROUP BY datname
`)
	if err != nil {
		log.Printf("Error querying database transaction rate: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var databasename sql.NullString
			var transactions int
			if err := rows.Scan(&databasename, &transactions); err != nil {
				log.Printf("Error scanning database transaction rate: %v", err)
				continue
			}
			if databasename.Valid {
				databaseTransactionRate.WithLabelValues(databasename.String).Set(float64(transactions))
			} else {
				databaseTransactionRate.WithLabelValues("unknown").Set(float64(transactions))
			}
		}
	}
}

func main() {
	// PostgreSQL connection string from environment variables
	pgConnStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	if pgConnStr == "" {
		log.Fatal("Environment variable POSTGRES_CONNECTION_STRING is required.")
	}

	// Create connection pool configuration
	config, err := pgxpool.ParseConfig(pgConnStr)
	if err != nil {
		log.Fatalf("Error parsing connection string: %v", err)
	}

	// Set pool configuration
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Create connection pool
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}
	defer pool.Close()

	// Periodic metric collection
	go func() {
		for {
			collectMetrics(context.Background(), pool)
			log.Println("Metrics collected.")
			time.Sleep(10 * time.Second)
		}
	}()

	// Expose metrics endpoint
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusMovedPermanently))
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Prometheus exporter started on " + net.JoinHostPort(host, port))

	log.Fatal(
		http.ListenAndServe(net.JoinHostPort(host, port), nil),
	)
}
