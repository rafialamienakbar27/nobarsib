// Package repository berisi implementasi penyimpanan data (blueprint §6.4).
// Di Fase 1 isinya baru koneksi; entity dan query menyusul di Fase 2.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgres membuka connection pool dan memastikan koneksinya benar-benar
// hidup sebelum aplikasi dinyatakan siap.
//
// MaxConns dibatasi karena §17.5 memperingatkan traffic melonjak tajam 2 jam
// sebelum kickoff lalu sepi lagi: pool yang tidak dibatasi akan menghabiskan
// slot koneksi Postgres justru di malam yang paling penting.
func NewPostgres(ctx context.Context, dsn string, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("buka pool postgres: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
