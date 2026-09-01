package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/middleware"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

// RegisterAuthRoutes memasang endpoint autentikasi (§8.3).
func RegisterAuthRoutes(app *fiber.App, h *Auth, rdb *redis.Client, log *slog.Logger) {
	// Rate limit ketat: 10 percobaan per 15 menit per IP (§8.5). Ini satu-satunya
	// endpoint yang menerima kata sandi.
	auth := app.Group("/v1/auth", middleware.RateLimit(rdb, log, middleware.LimitAuth))
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
}

// RegisterAdminRoutes memasang endpoint panel admin (§8.4).
//
// Seluruh grup dilindungi RequireAuth lalu RequireRole("admin"), dipasang di
// satu tempat supaya tidak ada rute yang bisa lolos karena lupa ditandai.
func RegisterAdminRoutes(app *fiber.App, h *Admin, auth *service.AuthService) {
	admin := app.Group("/v1/admin",
		middleware.RequireAuth(auth),
		middleware.RequireRole(domain.RoleAdmin),
	)

	admin.Get("/summary", h.Dashboard)

	admin.Get("/events", h.PendingEvents)
	admin.Post("/events", h.CreateEvent)
	admin.Post("/events/:id/approve", h.Approve)
	admin.Post("/events/:id/reject", h.Reject)
	admin.Post("/events/:id/confirm", h.Confirm)
	admin.Post("/events/:id/cancel", h.Cancel)

	admin.Post("/venues", h.CreateVenue)
	admin.Patch("/venues/:id", h.UpdateVenue)
	admin.Post("/venues/:id/photos", h.AddPhoto)

	admin.Post("/matches", h.CreateMatch)
	admin.Patch("/matches/:id", h.UpdateMatch)
}
