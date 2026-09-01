// Command api menjalankan HTTP server NOBARSIB.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/nobarsib/nobarsib-api/internal/config"
	"github.com/nobarsib/nobarsib-api/internal/handler"
	"github.com/nobarsib/nobarsib-api/internal/middleware"
	"github.com/nobarsib/nobarsib-api/internal/repository"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

// corsPreflightTTL adalah lama browser boleh menyimpan hasil preflight.
// Dua belas jam: cukup panjang supaya OPTIONS tidak mendahului setiap POST,
// masih cukup pendek supaya perubahan daftar origin berlaku dalam sehari.
const corsPreflightTTL = 12 * time.Hour

func main() {
	// .env hanya untuk kenyamanan development. Di server, variabel datang dari
	// environment sungguhan, jadi ketiadaan file ini bukan error.
	_ = godotenv.Load()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if err := run(log); err != nil {
		log.Error("api berhenti", slog.String("error", err.Error()))
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

	app := fiber.New(fiber.Config{
		AppName:               "nobarsib-api",
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler(log),
	})

	app.Use(fiberrecover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(log))

	// CORS hanya dipasang kalau ada origin yang didaftarkan.
	//
	// Saat pengembangan web ada di :3000 dan API di :8080 — dua origin berbeda,
	// jadi tanpa ini browser membuang setiap respons dari sisi klien dan tab
	// urutan, chip filter, serta tombol lokasi berhenti bekerja.
	//
	// AllowCredentials sengaja dibiarkan mati: autentikasi memakai header
	// Authorization, bukan cookie. Menyalakannya berarti mengizinkan situs lain
	// mengirim kredensial pengguna tanpa keuntungan apa pun bagi kita.
	if len(cfg.CORSOrigins) > 0 {
		app.Use(cors.New(cors.Config{
			AllowOrigins: strings.Join(cfg.CORSOrigins, ","),
			AllowMethods: strings.Join([]string{
				fiber.MethodGet, fiber.MethodPost, fiber.MethodPatch,
				fiber.MethodDelete, fiber.MethodOptions,
			}, ","),
			AllowHeaders: "Content-Type,Authorization",
			MaxAge:       int(corsPreflightTTL.Seconds()),
		}))
		log.Info("cors aktif", slog.Any("origins", cfg.CORSOrigins))
	}

	health := handler.NewHealth(db, rdb, cfg.AppEnv)
	app.Get("/health", health.Check)
	app.Get("/live", health.Live)

	venueRepo := repository.NewVenueRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	eventRepo := repository.NewNobarEventRepo(db)
	reviewRepo := repository.NewReviewRepo(db)
	viewRepo := repository.NewEventViewRepo(db)

	cache := service.NewNobarCache(rdb, log)
	nobarSvc := service.NewNobarService(eventRepo, matchRepo, cache)
	venueSvc := service.NewVenueService(venueRepo, reviewRepo, eventRepo)

	public := handler.NewPublic(nobarSvc, venueSvc, matchRepo, viewRepo, cache, log)
	handler.RegisterPublicRoutes(app, public, rdb, log)

	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, rdb, cfg.JWTSecret)
	handler.RegisterAuthRoutes(app, handler.NewAuth(authSvc), rdb, log)

	adminHandler := handler.NewAdmin(nobarSvc, venueSvc, eventRepo, matchRepo)
	handler.RegisterAdminRoutes(app, adminHandler, authSvc)

	errc := make(chan error, 1)
	go func() {
		log.Info("api mendengarkan",
			slog.String("port", cfg.Port),
			slog.String("env", cfg.AppEnv))
		if err := app.Listen(":" + cfg.Port); err != nil {
			errc <- err
		}
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		log.Info("sinyal berhenti diterima, menutup koneksi")
		return app.ShutdownWithTimeout(cfg.ShutdownTimeout)
	}
}

// errorHandler menyeragamkan bentuk error sesuai §8.1:
//
//	{"error": {"code": "...", "message": "...", "details": null}}
func errorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "Terjadi kesalahan pada server"

		var fe *fiber.Error
		if errors.As(err, &fe) {
			status = fe.Code
			message = fe.Message
			if status == fiber.StatusNotFound {
				code = "NOT_FOUND"
			} else if status < fiber.StatusInternalServerError {
				code = "BAD_REQUEST"
			}
		}

		if status >= fiber.StatusInternalServerError {
			// Pesan asli hanya masuk log, tidak dikirim ke klien.
			log.Error("unhandled error",
				slog.String("error", err.Error()),
				slog.String("request_id", middleware.RequestIDFrom(c)))
		}

		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": message,
				"details": nil,
			},
		})
	}
}
