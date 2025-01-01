package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kintsdev/postmetric/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
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
	metrics.RegisterMetrics()
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
			metrics.CollectMetrics(context.Background(), pool)
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
