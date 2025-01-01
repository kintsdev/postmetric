package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ConnectionUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_connection_usage",
		Help: "Percentage of used connections out of max_connections.",
	})

	IdleConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_idle_connections",
		Help: "Number of idle connections.",
	})

	LongRunningQueries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_long_running_queries",
		Help: "Number of queries running for more than 5 minutes.",
	})

	ReplicationLag = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_replication_lag",
		Help: "Replication lag in bytes.",
	})

	TotalSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_total_size_bytes",
		Help: "Total database size in bytes.",
	})

	IndexSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_index_size_bytes",
		Help: "Total index size in bytes.",
	})

	DiskReads = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_disk_reads",
		Help: "Number of disk blocks read.",
	})

	CacheHits = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_cache_hits",
		Help: "Number of cache hits.",
	})

	CommitRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_commit_rate",
		Help: "Transaction commit rate.",
	})

	RollbackTransactions = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_rollback_transactions",
		Help: "Number of rollback transactions.",
	})

	TableBloat = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_table_bloat_bytes",
		Help: "Bloat size in bytes for tables.",
	})

	AutovacuumWorkers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_autovacuum_workers",
		Help: "Number of active autovacuum workers.",
	})

	TempTableSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_temp_table_size_bytes",
		Help: "Total size of temporary tables in bytes.",
	})

	CheckpointWriteTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_checkpoint_write_time_ms",
		Help: "Average checkpoint write time in milliseconds.",
	})

	Deadlocks = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_deadlocks",
		Help: "Number of deadlocks detected.",
	})

	Locks = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_locks",
		Help: "Number of active locks.",
	})

	ActiveQueries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_active_queries",
		Help: "Number of currently active queries.",
	})

	WaitingQueries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "postgres_waiting_queries",
		Help: "Number of queries waiting for locks.",
	})

	TableStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_table_stats",
			Help: "Table statistics including row count, size, and dead tuples.",
		},
		[]string{"table_name", "stat_type"},
	)

	IndexStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_index_stats",
			Help: "Index statistics including size and scans.",
		},
		[]string{"index_name", "table_name", "stat_type"},
	)

	VacuumProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_vacuum_progress",
			Help: "Progress of vacuum operations.",
		},
		[]string{"table_name", "phase"},
	)

	BufferCacheStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_buffer_cache",
			Help: "Buffer cache statistics.",
		},
		[]string{"type"},
	)

	WalStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_wal_stats",
			Help: "WAL statistics including size and writes.",
		},
		[]string{"type"},
	)

	StatementStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_statement_stats",
			Help: "Statement execution statistics.",
		},
		[]string{"type"},
	)

	BgwriterStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_bgwriter_stats",
			Help: "Background writer statistics.",
		},
		[]string{"type"},
	)

	TableIOStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_table_io_stats",
			Help: "Table I/O statistics.",
		},
		[]string{"table_name", "type"},
	)

	ReplicationSlotStats = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_replication_slot_stats",
			Help: "Replication slot statistics.",
		},
		[]string{"slot_name", "type"},
	)

	UserConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_user_connections",
			Help: "Number of connections per user.",
		},
		[]string{"user"},
	)

	UserQueries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_user_queries",
			Help: "Number of queries executed per user.",
		},
		[]string{"user"},
	)

	DatabaseTransactionRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_database_transaction_rate",
			Help: "Transaction rate per database.",
		},
		[]string{"database"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(ConnectionUsage)
	prometheus.MustRegister(IdleConnections)
	prometheus.MustRegister(LongRunningQueries)
	prometheus.MustRegister(ReplicationLag)
	prometheus.MustRegister(TotalSize)
	prometheus.MustRegister(IndexSize)
	prometheus.MustRegister(DiskReads)
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CommitRate)
	prometheus.MustRegister(RollbackTransactions)
	prometheus.MustRegister(TableBloat)
	prometheus.MustRegister(AutovacuumWorkers)
	prometheus.MustRegister(TempTableSize)
	prometheus.MustRegister(CheckpointWriteTime)
	prometheus.MustRegister(Deadlocks)
	prometheus.MustRegister(Locks)
	prometheus.MustRegister(ActiveQueries)
	prometheus.MustRegister(WaitingQueries)
	prometheus.MustRegister(TableStats)
	prometheus.MustRegister(IndexStats)
	prometheus.MustRegister(VacuumProgress)
	prometheus.MustRegister(BufferCacheStats)
	prometheus.MustRegister(WalStats)
	prometheus.MustRegister(StatementStats)
	prometheus.MustRegister(BgwriterStats)
	prometheus.MustRegister(TableIOStats)
	prometheus.MustRegister(ReplicationSlotStats)
	prometheus.MustRegister(UserConnections)
	prometheus.MustRegister(UserQueries)
	prometheus.MustRegister(DatabaseTransactionRate)
}
