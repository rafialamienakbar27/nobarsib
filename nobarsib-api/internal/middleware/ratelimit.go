package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig menerapkan batas §8.5.
type RateLimitConfig struct {
	Name   string        // pembeda antar aturan dalam kunci Redis
	Limit  int           // jumlah permintaan yang diizinkan
	Window time.Duration // dalam rentang waktu ini
}

// Batas dari §8.5.
var (
	LimitPublicGet = RateLimitConfig{Name: "get", Limit: 120, Window: time.Minute}
	LimitTrack     = RateLimitConfig{Name: "track", Limit: 300, Window: time.Minute}
	LimitReview    = RateLimitConfig{Name: "review", Limit: 5, Window: time.Hour}
	LimitAuth      = RateLimitConfig{Name: "auth", Limit: 10, Window: 15 * time.Minute}
)

// RateLimit membatasi permintaan per IP memakai penghitung Redis.
//
// Kalau Redis sedang mati, permintaan DILOLOSKAN, bukan ditolak. Rate limit
// adalah perlindungan, bukan fitur: menolak seluruh trafik karena cache mati
// akan mematikan aplikasi tepat di malam pertandingan — persis saat §17.5
// memperingatkan trafik melonjak.
func RateLimit(rdb *redis.Client, log *slog.Logger, cfg RateLimitConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("rl:%s:%s", cfg.Name, clientIP(c))

		ctx, cancel := context.WithTimeout(c.UserContext(), 200*time.Millisecond)
		defer cancel()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			log.Warn("rate limit dilewati, redis tidak menjawab",
				slog.String("error", err.Error()))
			return c.Next()
		}
		// Kedaluwarsa dipasang hanya saat penghitung baru dibuat, supaya
		// jendelanya tetap dan tidak ikut memanjang setiap permintaan.
		if count == 1 {
			rdb.Expire(ctx, key, cfg.Window)
		}

		remaining := cfg.Limit - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprint(cfg.Limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprint(remaining))

		if int(count) > cfg.Limit {
			c.Set("Retry-After", fmt.Sprint(int(cfg.Window.Seconds())))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "RATE_LIMITED",
					"message": "Terlalu banyak permintaan, coba lagi sebentar lagi",
					"details": nil,
				},
			})
		}
		return c.Next()
	}
}

// clientIP mengambil IP asli di belakang Nginx (§6.3).
func clientIP(c *fiber.Ctx) string {
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return c.IP()
}
