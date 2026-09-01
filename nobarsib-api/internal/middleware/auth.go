package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/service"
)

const (
	CtxUserID = "user_id"
	CtxRole   = "role"
)

// RequireAuth menolak permintaan tanpa access token yang sah.
func RequireAuth(auth *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := bearerToken(c)
		if !ok {
			return tolak(c, "Akses ditolak, silakan masuk lebih dulu")
		}

		claims, err := auth.Verify(token)
		if err != nil {
			return tolak(c, "Sesi tidak berlaku atau sudah kedaluwarsa")
		}

		id, err := uuid.Parse(claims.Subject)
		if err != nil {
			return tolak(c, "Sesi tidak berlaku")
		}

		c.Locals(CtxUserID, id)
		c.Locals(CtxRole, claims.Role)
		return c.Next()
	}
}

// RequireRole dipasang setelah RequireAuth.
//
// Dipisah dari RequireAuth supaya portal venue di Fase 5 bisa memakai
// RequireAuth yang sama tanpa ikut memaksa peran admin.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if r, _ := c.Locals(CtxRole).(string); r != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Akun ini tidak punya akses ke halaman tersebut",
					"details": nil,
				},
			})
		}
		return c.Next()
	}
}

// UserIDFrom mengambil id pengguna yang sudah terautentikasi.
func UserIDFrom(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(CtxUserID).(uuid.UUID)
	return id, ok
}

func bearerToken(c *fiber.Ctx) (string, bool) {
	h := c.Get("Authorization")
	if h == "" {
		return "", false
	}
	skema, token, ok := strings.Cut(h, " ")
	if !ok || !strings.EqualFold(skema, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

func tolak(c *fiber.Ctx, pesan string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": fiber.Map{"code": "UNAUTHORIZED", "message": pesan, "details": nil},
	})
}
