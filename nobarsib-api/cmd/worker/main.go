// Command worker menjalankan tugas terjadwal NOBARSIB.
//
// Di Fase 1 worker hanya membuktikan koneksi dan siklus hidupnya benar.
// Lima jadwal cron di §12.3 dipasang di Fase 5:
//
//	09:00 harian  check-upcoming-matches   cek laga H-3 dan H-1
//	19:00 harian  remind-unconfirmed       venue yang belum konfirmasi
//	08:00 harian  review-prompt            H+1 setelah laga
//	02:30 harian  recalculate-scores       hitung ulang rating §11.3
//	Senin 03:00   sync-match-schedule      tarik jadwal mingguan
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/nobarsib/nobarsib-api/internal/config"
	"github.com/nobarsib/nobarsib-api/internal/repository"
)

// heartbeat sengaja panjang: worker ini belum punya pekerjaan, dan log yang
// ramai hanya menyulitkan saat mencari masalah sungguhan.
const heartbeat = 5 * time.Minute

func main() {
	_ = godotenv.Load()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(log); err != nil {
		log.Error("worker berhenti", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL, cfg.DatabaseMaxConn)
	if err != nil {
		return err
	}
	defer db.Close()

	rdb, err := repository.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer func() { _ = rdb.Close() }()

	log.Info("worker siap", slog.String("env", cfg.AppEnv))

	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("sinyal berhenti diterima, worker keluar")
			return nil
		case <-ticker.C:
			log.Info("worker hidup, belum ada tugas terjadwal (Fase 5)")
		}
	}
}
