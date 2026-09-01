package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/nobarsib/nobarsib-api/internal/service"
)

type Auth struct{ auth *service.AuthService }

func NewAuth(a *service.AuthService) *Auth { return &Auth{auth: a} }

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	User         userView `json:"user"`
}

type userView struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// Login menangani POST /v1/auth/login (§8.3).
func (h *Auth) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	pair, err := h.auth.Login(c.UserContext(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrKredensialSalah) {
			// Pesannya sengaja tidak membedakan "email tidak ada" dari "kata
			// sandi salah": membedakannya sama saja memberi tahu penyerang
			// email mana yang terdaftar.
			return sendError(c, fiber.StatusUnauthorized, "UNAUTHORIZED",
				"Email atau kata sandi salah")
		}
		return fail(c, err)
	}
	return c.JSON(newTokenResponse(pair))
}

// Refresh menangani POST /v1/auth/refresh (§8.3).
func (h *Auth) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "refresh_token wajib diisi")
	}

	pair, err := h.auth.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return sendError(c, fiber.StatusUnauthorized, "UNAUTHORIZED",
			"Sesi sudah berakhir, silakan masuk lagi")
	}
	return c.JSON(newTokenResponse(pair))
}

// Logout mencabut refresh token.
func (h *Auth) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		_ = h.auth.Logout(c.UserContext(), req.RefreshToken)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func newTokenResponse(p *service.TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
		ExpiresIn:    p.ExpiresIn,
		User: userView{
			ID:       p.User.ID.String(),
			Email:    p.User.Email,
			FullName: p.User.FullName,
			Role:     p.User.Role,
		},
	}
}
