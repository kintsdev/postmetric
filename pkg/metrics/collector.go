package metrics

import (
	"context"
	"database/sql"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func CollectMetrics(ctx context.Context, pool *pgxpool.Pool) {
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
		ActiveQueries.Set(float64(activeCount))
		WaitingQueries.Set(float64(waitingCount))
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
			TableStats.WithLabelValues(tableName, "live_tuples").Set(float64(liveTuples))
			TableStats.WithLabelValues(tableName, "dead_tuples").Set(float64(deadTuples))
			TableStats.WithLabelValues(tableName, "total_size_bytes").Set(float64(totalSize))
			TableStats.WithLabelValues(tableName, "table_size_bytes").Set(float64(tableSize))
			TableStats.WithLabelValues(tableName, "index_size_bytes").Set(float64(indexSize))
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
		WalStats.WithLabelValues("bytes").Set(float64(walBytes))
		WalStats.WithLabelValues("files").Set(float64(walFiles))
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
		BufferCacheStats.WithLabelValues("hits").Set(float64(bufferHits))
		BufferCacheStats.WithLabelValues("reads").Set(float64(bufferReads))
		BufferCacheStats.WithLabelValues("total").Set(float64(bufferHits + bufferReads))
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
			ReplicationSlotStats.WithLabelValues(slotName, "lag_bytes").Set(float64(lagBytes))
			if active {
				ReplicationSlotStats.WithLabelValues(slotName, "active").Set(1)
			} else {
				ReplicationSlotStats.WithLabelValues(slotName, "active").Set(0)
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

	ConnectionUsage.Set(float64(usedConnections) / float64(maxConnections) * 100)

	// Collect idle connections
	var idleConn int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'idle'").Scan(&idleConn)
	if err != nil {
		log.Printf("Error querying idle connections: %v", err)
		return
	}

	IdleConnections.Set(float64(idleConn))

	// Collect long-running queries
	var longQueries int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '5 minutes'").Scan(&longQueries)
	if err != nil {
		log.Printf("Error querying long-running queries: %v", err)
		return
	}

	LongRunningQueries.Set(float64(longQueries))

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
	ReplicationLag.Set(float64(lagBytes))

	// Collect total database size
	var sizeBytes int64
	err = pool.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&sizeBytes)
	if err != nil {
		log.Printf("Error querying total database size: %v", err)
		return
	}

	TotalSize.Set(float64(sizeBytes))

	// Collect index size
	var idxSize int64
	err = pool.QueryRow(ctx, "SELECT sum(pg_indexes_size(indexrelid)) FROM pg_stat_user_indexes").Scan(&idxSize)
	if err != nil {
		log.Printf("Error querying index size: %v", err)
		return
	}

	IndexSize.Set(float64(idxSize))

	// Collect disk reads and cache hits
	var diskRead, cacheHit int64
	err = pool.QueryRow(ctx, "SELECT sum(heap_blks_read), sum(heap_blks_hit) FROM pg_statio_user_tables").Scan(&diskRead, &cacheHit)
	if err != nil {
		log.Printf("Error querying disk reads and cache hits: %v", err)
		return
	}

	DiskReads.Set(float64(diskRead))
	CacheHits.Set(float64(cacheHit))

	// Collect commit rate
	var commits, rollbacks int64
	err = pool.QueryRow(ctx, "SELECT sum(xact_commit), sum(xact_rollback) FROM pg_stat_database").Scan(&commits, &rollbacks)
	if err != nil {
		log.Printf("Error querying commit rate: %v", err)
		return
	}

	if (commits + rollbacks) > 0 {
		CommitRate.Set(float64(commits) / float64(commits+rollbacks) * 100)
	}

	RollbackTransactions.Set(float64(rollbacks))

	// Collect table bloat
	var bloatSize int64
	err = pool.QueryRow(ctx, "SELECT sum(pg_total_relation_size(relid) - pg_relation_size(relid)) FROM pg_stat_user_tables").Scan(&bloatSize)
	if err != nil {
		log.Printf("Error querying table bloat: %v", err)
		return
	}

	TableBloat.Set(float64(bloatSize))

	// Collect autovacuum workers
	var autovacWorkers int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_progress_vacuum").Scan(&autovacWorkers)
	if err != nil {
		log.Printf("Error querying autovacuum workers: %v", err)
		return
	}

	AutovacuumWorkers.Set(float64(autovacWorkers))

	var tempSize int64
	err = pool.QueryRow(ctx, "SELECT sum(temp_bytes) FROM pg_stat_database").Scan(&tempSize)
	if err != nil {
		log.Printf("Error querying temporary table size: %v", err)
		return
	}

	TempTableSize.Set(float64(tempSize))

	// Collect checkpoint write time
	var checkpointTime float64
	err = pool.QueryRow(ctx, "SELECT avg(checkpoint_write_time) FROM pg_stat_bgwriter").Scan(&checkpointTime)
	if err != nil {
		log.Printf("Error querying checkpoint write time: %v", err)
		return
	}

	CheckpointWriteTime.Set(checkpointTime)

	// Collect deadlocks
	var totalDeadlocks int64
	err = pool.QueryRow(ctx, "SELECT sum(deadlocks) FROM pg_stat_database").Scan(&totalDeadlocks)
	if err != nil {
		log.Printf("Error querying deadlocks: %v", err)
		return
	}

	Deadlocks.Set(float64(totalDeadlocks))

	// Collect locks
	var totalLocks int64
	err = pool.QueryRow(ctx, "SELECT count(*) FROM pg_locks").Scan(&totalLocks)
	if err != nil {
		log.Printf("Error querying locks: %v", err)
		return
	}

	Locks.Set(float64(totalLocks))

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
			UserConnections.WithLabelValues(user).Set(float64(connections))
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
			UserQueries.WithLabelValues(user).Set(float64(queries))
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
				DatabaseTransactionRate.WithLabelValues(databasename.String).Set(float64(transactions))
			} else {
				DatabaseTransactionRate.WithLabelValues("unknown").Set(float64(transactions))
			}
		}
	}
}
