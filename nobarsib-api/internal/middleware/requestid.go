// Package middleware berisi lapisan lintas-permintaan: request ID, logging,
// dan nanti auth serta rate limit (blueprint §6.4).
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// HeaderRequestID dikirim balik ke klien supaya keluhan pengguna
	// ("kok error jam 8 tadi") bisa langsung ditelusuri di log.
	HeaderRequestID = "X-Request-ID"

	// CtxRequestID adalah kunci penyimpanan di fiber.Ctx.Locals.
	CtxRequestID = "request_id"
)

// RequestID memakai ulang header dari reverse proxy kalau ada (Nginx di §6.3),
// supaya satu permintaan punya satu ID dari tepi sampai database.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals(CtxRequestID, id)
		c.Set(HeaderRequestID, id)
		return c.Next()
	}
}

// RequestIDFrom mengambil request ID dari context; mengembalikan string kosong
// kalau middleware RequestID belum terpasang.
func RequestIDFrom(c *fiber.Ctx) string {
	if id, ok := c.Locals(CtxRequestID).(string); ok {
		return id
	}
	return ""
}
