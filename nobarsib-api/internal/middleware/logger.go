package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger mencatat setiap permintaan dalam format terstruktur.
//
// Sengaja tidak mencatat query string maupun body: UU PDP berlaku (§22) dan
// device_hash pengguna tidak perlu ikut tersimpan di log server.
func Logger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		attrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("took", time.Since(start)),
		}
		if id := RequestIDFrom(c); id != "" {
			attrs = append(attrs, slog.String("request_id", id))
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			log.Error("request gagal", attrs...)
			return err
		}

		if c.Response().StatusCode() >= fiber.StatusInternalServerError {
			log.Error("request error", attrs...)
		} else {
			log.Info("request", attrs...)
		}
		return nil
	}
}
