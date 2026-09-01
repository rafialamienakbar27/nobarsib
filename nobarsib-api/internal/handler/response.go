package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

// Kode error yang dipakai klien untuk membedakan penanganan (§8.1).
const (
	CodeNotFound      = "NOT_FOUND"
	CodeInvalidInput  = "INVALID_INPUT"
	CodeConflict      = "CONFLICT"
	CodeInvalidState  = "INVALID_STATE"
	CodeInternalError = "INTERNAL_ERROR"
)

// fail menerjemahkan error domain menjadi response §8.1.
//
// Pesan error domain sengaja ditulis dalam bahasa Indonesia dan aman dibaca
// pengguna. Error tak terduga tidak pernah bocor isinya — hanya masuk log
// lewat ErrorHandler di cmd/api.
func fail(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return sendError(c, fiber.StatusNotFound, CodeNotFound, "Data tidak ditemukan")
	case errors.Is(err, domain.ErrInvalidInput):
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, cleanMessage(err))
	case errors.Is(err, domain.ErrConflict):
		return sendError(c, fiber.StatusConflict, CodeConflict, cleanMessage(err))
	case errors.Is(err, domain.ErrInvalidTransition):
		return sendError(c, fiber.StatusUnprocessableEntity, CodeInvalidState, cleanMessage(err))
	default:
		return err // ditangani ErrorHandler global, dicatat, tidak dibocorkan
	}
}

// cleanMessage membuang awalan sentinel supaya pesannya enak dibaca:
// "input tidak valid: sort \"jauh\" tidak dikenal" menjadi bagian setelah titik dua.
func cleanMessage(err error) string {
	msg := err.Error()
	if _, after, found := strings.Cut(msg, ": "); found {
		return capitalize(after)
	}
	return capitalize(msg)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func sendError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": message, "details": nil},
	})
}

// meta adalah blok paginasi §8.2.
type meta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func pageOf(offset, limit int) int {
	if limit <= 0 {
		return 1
	}
	return offset/limit + 1
}
