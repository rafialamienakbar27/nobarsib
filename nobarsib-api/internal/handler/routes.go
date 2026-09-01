package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/nobarsib/nobarsib-api/internal/middleware"
)

// RegisterPublicRoutes memasang seluruh endpoint publik §8.2 beserta batas
// lajunya §8.5.
//
// Batas dipasang per grup, bukan global: POST /track jauh lebih sering
// dipanggil daripada GET biasa (setiap kartu yang terlihat memicu satu), jadi
// menyamakan batasnya akan memblokir pengguna yang wajar hanya karena
// menggulir daftar.
func RegisterPublicRoutes(app *fiber.App, p *Public, rdb *redis.Client, log *slog.Logger) {
	v1 := app.Group("/v1")

	read := v1.Group("", middleware.RateLimit(rdb, log, middleware.LimitPublicGet))
	read.Get("/matches/upcoming", p.UpcomingMatches)
	read.Get("/matches", p.Season)
	read.Get("/matches/:match_id/nobar", p.NobarForMatch)
	read.Get("/venues/search", p.SearchVenues)
	read.Get("/venues/:slug", p.VenueDetail)
	read.Get("/facilities", p.Facilities)

	track := v1.Group("", middleware.RateLimit(rdb, log, middleware.LimitTrack))
	track.Post("/events/:id/track", p.Track)
}
