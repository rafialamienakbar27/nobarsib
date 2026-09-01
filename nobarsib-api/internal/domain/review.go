package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Aksi yang dicatat POST /v1/events/{id}/track (§8.2).
const (
	ActionViewCard   = "view_card"
	ActionOpenDetail = "open_detail"
	ActionOpenMaps   = "open_maps"
	ActionClickWA    = "click_wa"
)

// ValidActions dipakai handler untuk menolak aksi yang tidak dikenal sebelum
// menyentuh database.
var ValidActions = []string{ActionViewCard, ActionOpenDetail, ActionOpenMaps, ActionClickWA}

func IsValidAction(a string) bool {
	for _, v := range ValidActions {
		if v == a {
			return true
		}
	}
	return false
}

type Review struct {
	ID           uuid.UUID
	NobarEventID *uuid.UUID
	VenueID      uuid.UUID

	RatingOverall  int
	RatingKondusif *int
	IsKidFriendly  *bool
	CrowdActual    string
	Comment        string

	DeviceHash string
	IPHash     string
	IsHidden   bool
	CreatedAt  time.Time
}

// ReviewRepository di Fase 2 baru dipakai untuk membaca. Pengiriman review
// (POST /v1/reviews) beserta anti-spamnya masuk di Fase 5 bersama §11.5.
type ReviewRepository interface {
	ListByVenue(ctx context.Context, venueID uuid.UUID, limit int) ([]Review, error)
}

type EventView struct {
	NobarEventID uuid.UUID
	DeviceHash   string
	Action       string
}

type EventViewRepository interface {
	// Track mencatat interaksi. Sengaja fire-and-forget di sisi handler (§8.2):
	// kegagalan mencatat statistik tidak boleh merusak pengalaman pengguna.
	Track(ctx context.Context, v EventView) error
}
