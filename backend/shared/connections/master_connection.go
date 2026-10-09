package connections

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	defaultDriver = "postgres"
	defaultDSN    = ""
)

// NewMasterConnection opens the primary database connection using env vars.
//
// Required env vars:
// - MASTER_DB_DSN
//
// Optional env vars:
// - MASTER_DB_DRIVER
// - MASTER_DB_MAX_OPEN_CONNS
// - MASTER_DB_MAX_IDLE_CONNS
// - MASTER_DB_CONN_MAX_LIFETIME
func NewMasterConnection() (*sqlx.DB, error) {
	driver := getenv("MASTER_DB_DRIVER", defaultDriver)
	dsn := os.Getenv("MASTER_DB_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	if dsn == "" {
		return nil, fmt.Errorf("MASTER_DB_DSN is required")
	}

	return openConfiguredConnection(connectionConfig{
		Driver:          driver,
		DSN:             dsn,
		MaxOpenConns:    getenvInt("MASTER_DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getenvInt("MASTER_DB_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: getenvDuration("MASTER_DB_CONN_MAX_LIFETIME", 5*time.Minute),
	})
}

// NewTRConnection opens the transaction database connection using TR_DB_DSN.
func NewTRConnection() (*sqlx.DB, error) {
	driver := getenv("TR_DB_DRIVER", getenv("MASTER_DB_DRIVER", defaultDriver))
	dsn := os.Getenv("TR_DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("TR_DB_DSN is required")
	}

	return openConfiguredConnection(connectionConfig{
		Driver:          driver,
		DSN:             dsn,
		MaxOpenConns:    getenvInt("TR_DB_MAX_OPEN_CONNS", getenvInt("MASTER_DB_MAX_OPEN_CONNS", 25)),
		MaxIdleConns:    getenvInt("TR_DB_MAX_IDLE_CONNS", getenvInt("MASTER_DB_MAX_IDLE_CONNS", 25)),
		ConnMaxLifetime: getenvDuration("TR_DB_CONN_MAX_LIFETIME", getenvDuration("MASTER_DB_CONN_MAX_LIFETIME", 5*time.Minute)),
	})
}

type connectionConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func openConfiguredConnection(config connectionConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return n
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return d
}
