// Package handler berisi HTTP handler (blueprint §6.4).
package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// checkTimeout menjaga /health tetap cepat merespons. UptimeRobot (§17.5)
// ping tiap 5 menit dan tidak sabar menunggu koneksi yang menggantung.
const checkTimeout = 2 * time.Second

type Health struct {
	db    *pgxpool.Pool
	redis *redis.Client
	env   string
}

func NewHealth(db *pgxpool.Pool, rdb *redis.Client, env string) *Health {
	return &Health{db: db, redis: rdb, env: env}
}

type dependencyStatus struct {
	Status string `json:"status"`          // "ok" | "down"
	Error  string `json:"error,omitempty"` // hanya di non-production
}

type healthResponse struct {
	Status       string                      `json:"status"`
	Env          string                      `json:"env"`
	Time         string                      `json:"time"`
	Dependencies map[string]dependencyStatus `json:"dependencies"`
}

// Check memverifikasi Postgres dan Redis benar-benar menjawab, bukan sekadar
// membalas 200 OK.
//
// Endpoint yang selalu 200 tidak berguna sebagai alarm: kalau database mati di
// malam pertandingan, monitoring harus tahu dari sini, bukan dari bobotoh yang
// mengeluh di Instagram.
func (h *Health) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), checkTimeout)
	defer cancel()

	deps := map[string]dependencyStatus{
		"postgres": h.checkErr(h.db.Ping(ctx)),
		"redis":    h.checkErr(h.redis.Ping(ctx).Err()),
	}

	status := fiber.StatusOK
	overall := "ok"
	for _, d := range deps {
		if d.Status != "ok" {
			status = fiber.StatusServiceUnavailable
			overall = "degraded"
			break
		}
	}

	return c.Status(status).JSON(healthResponse{
		Status:       overall,
		Env:          h.env,
		Time:         time.Now().Format(time.RFC3339),
		Dependencies: deps,
	})
}

// checkErr menyembunyikan detail error di production supaya /health yang
// terbuka ke publik tidak membocorkan host, port, atau nama database.
func (h *Health) checkErr(err error) dependencyStatus {
	if err == nil {
		return dependencyStatus{Status: "ok"}
	}
	d := dependencyStatus{Status: "down"}
	if h.env != "production" {
		d.Error = err.Error()
	}
	return d
}

// Live hanya menjawab apakah proses masih hidup, tanpa menyentuh dependensi.
// Dipakai orchestrator untuk memutuskan restart; jangan disamakan dengan Check.
func (h *Health) Live(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
