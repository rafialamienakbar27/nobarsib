package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

type ReviewRepo struct{ db *pgxpool.Pool }

func NewReviewRepo(db *pgxpool.Pool) *ReviewRepo { return &ReviewRepo{db: db} }

// ListByVenue mengambil review terbaru untuk halaman detail (§13.4 blok 8).
//
// Review tersembunyi tidak ikut, dan batas 18 bulan mengikuti §11.3: venue
// berubah — manajemen ganti, layar diganti, kebijakan merokok berubah — dan
// review tiga tahun lalu menyesatkan.
func (r *ReviewRepo) ListByVenue(ctx context.Context, venueID uuid.UUID, limit int) ([]domain.Review, error) {
	const q = `
SELECT id, nobar_event_id, venue_id, rating_overall, rating_kondusif,
       is_kid_friendly, COALESCE(crowd_actual,''), COALESCE(comment,''), created_at
FROM review
WHERE venue_id = $1
  AND is_hidden = FALSE
  AND created_at > now() - INTERVAL '18 months'
ORDER BY created_at DESC
LIMIT $2`

	rows, err := r.db.Query(ctx, q, venueID, limit)
	if err != nil {
		return nil, fmt.Errorf("daftar review: %w", err)
	}
	defer rows.Close()

	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.NobarEventID, &rv.VenueID, &rv.RatingOverall,
			&rv.RatingKondusif, &rv.IsKidFriendly, &rv.CrowdActual, &rv.Comment,
			&rv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

type EventViewRepo struct{ db *pgxpool.Pool }

func NewEventViewRepo(db *pgxpool.Pool) *EventViewRepo { return &EventViewRepo{db: db} }

// Track mencatat interaksi pengguna (§8.2).
//
// Ini sumber data statistik yang nanti dijual ke venue (§15.4) dan bahan
// estimasi keramaian (§9.5), jadi dipasang sejak sekarang meski belum ada
// yang membacanya.
func (r *EventViewRepo) Track(ctx context.Context, v domain.EventView) error {
	const q = `INSERT INTO event_view (nobar_event_id, device_hash, action) VALUES ($1, NULLIF($2,''), $3)`
	if _, err := r.db.Exec(ctx, q, v.NobarEventID, v.DeviceHash, v.Action); err != nil {
		return fmt.Errorf("catat interaksi: %w", err)
	}
	return nil
}
