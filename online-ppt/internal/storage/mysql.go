package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 25
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 10 * time.Minute
	defaultConnectTimeout  = 5 * time.Second
)

// MySQLFactory exposes helpers to open pooled database connections.
type MySQLFactory struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Open initializes the SQL client and verifies connectivity with a bounded ping.
func (f MySQLFactory) Open(ctx context.Context) (*sql.DB, error) {
	if f.DSN == "" {
		return nil, fmt.Errorf("mysql factory requires DSN")
	}

	db, err := sql.Open("mysql", f.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	maxIdle := f.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = defaultMaxIdleConns
	}
	db.SetMaxIdleConns(maxIdle)

	maxOpen := f.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = defaultMaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	life := f.ConnMaxLifetime
	if life <= 0 {
		life = defaultConnMaxLifetime
	}
	db.SetConnMaxLifetime(life)

	idle := f.ConnMaxIdleTime
	if idle <= 0 {
		idle = defaultConnMaxIdleTime
	}
	db.SetConnMaxIdleTime(idle)

	pingCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}
