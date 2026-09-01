// Package domain berisi entity dan kontrak repository (blueprint §6.4).
// Paket ini tidak boleh bergantung pada Fiber, pgx, atau Redis — supaya logika
// bisnis bisa diuji tanpa menyalakan apa pun.
package domain

import "errors"

var (
	ErrNotFound     = errors.New("data tidak ditemukan")
	ErrConflict     = errors.New("data sudah ada")
	ErrInvalidInput = errors.New("input tidak valid")

	// ErrInvalidTransition dikembalikan kalau perpindahan status nobar_event
	// tidak sah menurut state machine §4.5.
	ErrInvalidTransition = errors.New("perpindahan status tidak diizinkan")
)
