package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/pkg/openhours"
)

type VenueRepo struct{ db *pgxpool.Pool }

func NewVenueRepo(db *pgxpool.Pool) *VenueRepo { return &VenueRepo{db: db} }

const venueColumns = `
    v.id, v.name, v.slug, v.address, COALESCE(v.district,''), v.city,
    ST_Y(v.location::geometry), ST_X(v.location::geometry),
    COALESCE(v.phone,''), COALESCE(v.whatsapp,''), COALESCE(v.instagram_handle,''),
    COALESCE(v.website,''), COALESCE(v.google_place_id,''),
    v.google_rating, v.google_rating_count,
    v.opening_hours, v.nobar_rating, v.nobar_rating_count, v.kondusif_score,
    v.kid_friendly_score, v.data_completeness, v.status, v.is_active,
    COALESCE(v.data_source,''), v.last_verified_at,
    v.created_at, v.updated_at`

func scanVenue(row pgx.Row) (*domain.Venue, error) {
	var (
		v   domain.Venue
		raw []byte
	)
	err := row.Scan(&v.ID, &v.Name, &v.Slug, &v.Address, &v.District, &v.City,
		&v.Lat, &v.Lng, &v.Phone, &v.WhatsApp, &v.InstagramHandle,
		&v.Website, &v.GooglePlaceID, &v.GoogleRating, &v.GoogleRatingCount,
		&raw, &v.NobarRating, &v.NobarRatingCount, &v.KondusifScore,
		&v.KidFriendlyScore, &v.DataCompleteness, &v.Status, &v.IsActive,
		&v.DataSource, &v.LastVerifiedAt,
		&v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan venue: %w", err)
	}

	week, err := openhours.UnmarshalWeek(raw)
	if err != nil {
		return nil, err
	}
	v.OpeningHours = week
	return &v, nil
}

func (r *VenueRepo) GetBySlug(ctx context.Context, slug string) (*domain.Venue, error) {
	v, err := scanVenue(r.db.QueryRow(ctx,
		`SELECT `+venueColumns+` FROM venue v WHERE v.slug = $1 AND v.is_active = TRUE`, slug))
	if err != nil {
		return nil, err
	}
	return r.loadRelations(ctx, v)
}

func (r *VenueRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Venue, error) {
	v, err := scanVenue(r.db.QueryRow(ctx,
		`SELECT `+venueColumns+` FROM venue v WHERE v.id = $1`, id))
	if err != nil {
		return nil, err
	}
	return r.loadRelations(ctx, v)
}

// loadRelations mengisi fasilitas dan foto untuk halaman detail (§13.4).
func (r *VenueRepo) loadRelations(ctx context.Context, v *domain.Venue) (*domain.Venue, error) {
	rows, err := r.db.Query(ctx, `
SELECT f.code FROM venue_facility vf
JOIN facility f ON f.id = vf.facility_id
WHERE vf.venue_id = $1 ORDER BY f.category, f.code`, v.ID)
	if err != nil {
		return nil, fmt.Errorf("ambil fasilitas venue: %w", err)
	}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan fasilitas: %w", err)
		}
		v.Facilities = append(v.Facilities, code)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	prows, err := r.db.Query(ctx, `
SELECT id, url, COALESCE(caption,''), is_primary, sort_order
FROM venue_photo WHERE venue_id = $1
ORDER BY is_primary DESC, sort_order, created_at`, v.ID)
	if err != nil {
		return nil, fmt.Errorf("ambil foto venue: %w", err)
	}
	defer prows.Close()
	for prows.Next() {
		var p domain.VenuePhoto
		if err := prows.Scan(&p.ID, &p.URL, &p.Caption, &p.IsPrimary, &p.SortOrder); err != nil {
			return nil, fmt.Errorf("scan foto: %w", err)
		}
		if v.PrimaryPhoto == "" {
			v.PrimaryPhoto = p.URL
		}
		v.Photos = append(v.Photos, p)
	}
	return v, prows.Err()
}

// Search memakai pg_trgm untuk pencarian nama yang toleran salah ketik (§8.2).
func (r *VenueRepo) Search(ctx context.Context, p domain.VenueSearchParams) ([]domain.Venue, int, error) {
	const q = `
SELECT v.id, v.name, v.slug, COALESCE(v.district,''), v.address,
       ST_Y(v.location::geometry), ST_X(v.location::geometry),
       v.google_rating, v.nobar_rating, v.nobar_rating_count,
       COALESCE((SELECT p.url FROM venue_photo p WHERE p.venue_id = v.id
                  ORDER BY p.is_primary DESC, p.sort_order LIMIT 1), ''),
       COUNT(*) OVER() AS total
FROM venue v
WHERE v.is_active = TRUE
  AND (v.name ILIKE '%' || $1 || '%' OR similarity(v.name, $1) > 0.2)
ORDER BY similarity(v.name, $1) DESC, v.nobar_rating_count DESC, v.name
LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, q, p.Query, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("cari venue: %w", err)
	}
	defer rows.Close()

	var (
		out   []domain.Venue
		total int
	)
	for rows.Next() {
		var v domain.Venue
		if err := rows.Scan(&v.ID, &v.Name, &v.Slug, &v.District, &v.Address,
			&v.Lat, &v.Lng, &v.GoogleRating, &v.NobarRating, &v.NobarRatingCount,
			&v.PrimaryPhoto, &total); err != nil {
			return nil, 0, fmt.Errorf("scan hasil pencarian: %w", err)
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func (r *VenueRepo) Create(ctx context.Context, v *domain.Venue) error {
	hours, err := marshalHours(v.OpeningHours)
	if err != nil {
		return err
	}

	const q = `
INSERT INTO venue (name, slug, address, district, city, location, phone, whatsapp,
                   instagram_handle, website, google_place_id, google_rating,
                   google_rating_count, opening_hours, status,
                   data_source, last_verified_at)
VALUES ($1,$2,$3,NULLIF($4,''),$5, ST_SetSRID(ST_MakePoint($7,$6),4326)::geography,
        NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),NULLIF($11,''),NULLIF($12,''),
        $13,$14,$15,$16,NULLIF($17,''),$18)
RETURNING id, created_at, updated_at`

	err = r.db.QueryRow(ctx, q, v.Name, v.Slug, v.Address, v.District, v.City,
		v.Lat, v.Lng, v.Phone, v.WhatsApp, v.InstagramHandle, v.Website,
		v.GooglePlaceID, v.GoogleRating, v.GoogleRatingCount, hours, v.Status,
		v.DataSource, v.LastVerifiedAt,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)

	if isUniqueViolation(err) {
		return fmt.Errorf("%w: slug atau google_place_id sudah dipakai", domain.ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("simpan venue: %w", err)
	}
	return nil
}

func (r *VenueRepo) Update(ctx context.Context, v *domain.Venue) error {
	hours, err := marshalHours(v.OpeningHours)
	if err != nil {
		return err
	}

	const q = `
UPDATE venue SET
    name = $2, address = $3, district = NULLIF($4,''), city = $5,
    location = ST_SetSRID(ST_MakePoint($7,$6),4326)::geography,
    phone = NULLIF($8,''), whatsapp = NULLIF($9,''),
    instagram_handle = NULLIF($10,''), website = NULLIF($11,''),
    google_rating = $12, google_rating_count = $13, opening_hours = $14,
    status = $15, is_active = $16,
    -- COALESCE, bukan penugasan langsung: pembaruan yang tidak menyebut
    -- asal-usulnya tidak boleh menghapus asal-usul yang sudah tercatat.
    data_source = COALESCE(NULLIF($17,''), data_source),
    last_verified_at = COALESCE($18, last_verified_at)
WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, v.ID, v.Name, v.Address, v.District, v.City,
		v.Lat, v.Lng, v.Phone, v.WhatsApp, v.InstagramHandle, v.Website,
		v.GoogleRating, v.GoogleRatingCount, hours, v.Status, v.IsActive,
		v.DataSource, v.LastVerifiedAt)
	if err != nil {
		return fmt.Errorf("ubah venue: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetFacilities mengganti seluruh fasilitas venue dalam satu transaksi.
// Diganti utuh, bukan ditambah, supaya mencabut fasilitas yang tidak lagi ada
// tidak butuh endpoint tersendiri.
func (r *VenueRepo) SetFacilities(ctx context.Context, venueID uuid.UUID, codes []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mulai transaksi fasilitas: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM venue_facility WHERE venue_id = $1`, venueID); err != nil {
		return fmt.Errorf("hapus fasilitas lama: %w", err)
	}
	if len(codes) > 0 {
		const q = `
INSERT INTO venue_facility (venue_id, facility_id)
SELECT $1, f.id FROM facility f WHERE f.code = ANY($2::text[])`
		if _, err := tx.Exec(ctx, q, venueID, codes); err != nil {
			return fmt.Errorf("pasang fasilitas: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (r *VenueRepo) AddPhoto(ctx context.Context, venueID uuid.UUID, p *domain.VenuePhoto) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mulai transaksi foto: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Index unik parsial di migrasi 000004 hanya mengizinkan satu foto utama
	// per venue, jadi yang lama harus diturunkan lebih dulu.
	if p.IsPrimary {
		if _, err := tx.Exec(ctx,
			`UPDATE venue_photo SET is_primary = FALSE WHERE venue_id = $1 AND is_primary`, venueID); err != nil {
			return fmt.Errorf("turunkan foto utama lama: %w", err)
		}
	}

	const q = `
INSERT INTO venue_photo (venue_id, url, caption, is_primary, sort_order)
VALUES ($1,$2,NULLIF($3,''),$4,$5) RETURNING id`
	if err := tx.QueryRow(ctx, q, venueID, p.URL, p.Caption, p.IsPrimary, p.SortOrder).Scan(&p.ID); err != nil {
		return fmt.Errorf("simpan foto: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *VenueRepo) DeletePhoto(ctx context.Context, photoID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM venue_photo WHERE id = $1`, photoID)
	if err != nil {
		return fmt.Errorf("hapus foto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *VenueRepo) ListFacilities(ctx context.Context) ([]domain.Facility, error) {
	rows, err := r.db.Query(ctx,
		`SELECT code, label, COALESCE(icon,''), COALESCE(category,'') FROM facility ORDER BY category, id`)
	if err != nil {
		return nil, fmt.Errorf("daftar fasilitas: %w", err)
	}
	defer rows.Close()

	var out []domain.Facility
	for rows.Next() {
		var f domain.Facility
		if err := rows.Scan(&f.Code, &f.Label, &f.Icon, &f.Category); err != nil {
			return nil, fmt.Errorf("scan fasilitas: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// RecalculateCompleteness memakai fungsi SQL dari migrasi 000007 (§9.3).
func (r *VenueRepo) RecalculateCompleteness(ctx context.Context, venueID uuid.UUID) (float64, error) {
	var score float64
	err := r.db.QueryRow(ctx, `
UPDATE venue SET data_completeness = venue_data_completeness($1)
WHERE id = $1 RETURNING data_completeness`, venueID).Scan(&score)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("hitung kelengkapan data: %w", err)
	}
	return score, nil
}

// marshalHours menormalisasi jam buka sebelum disimpan.
//
// Ini satu-satunya pintu masuk ke kolom opening_hours, sehingga close_minutes
// dijamin selalu terisi dan filter §9.4 tidak pernah bekerja di atas data
// mentah — inti dari PERBAIKAN #2.
func marshalHours(w openhours.Week) ([]byte, error) {
	if w == nil {
		return nil, nil
	}
	if err := w.Normalize(); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidInput, err)
	}
	b, err := json.Marshal(w)
	if err != nil {
		return nil, fmt.Errorf("tulis opening_hours: %w", err)
	}
	return b, nil
}
