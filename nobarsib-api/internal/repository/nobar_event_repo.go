package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/pkg/scoring"
)

type NobarEventRepo struct{ db *pgxpool.Pool }

func NewNobarEventRepo(db *pgxpool.Pool) *NobarEventRepo { return &NobarEventRepo{db: db} }

// searchColumns dipakai bersama oleh query daftar dan riwayat.
//
// facilities dan primary_photo diambil lewat subquery berkorelasi, bukan JOIN
// + GROUP BY, supaya baris venue tidak berlipat dan window function
// COUNT(*) OVER() tetap menghitung venue, bukan pasangan venue-fasilitas.
const searchColumns = `
    ne.id, ne.venue_id, ne.match_id, ne.doors_open_at, ne.entry_type,
    ne.entry_amount, ne.capacity_estimate, COALESCE(ne.crowd_level,''),
    COALESCE(ne.notes,''), ne.status, ne.confirmed_at, ne.is_promoted,
    v.name, v.slug, COALESCE(v.district,''), v.address,
    ST_Y(v.location::geometry), ST_X(v.location::geometry),
    COALESCE(v.whatsapp,''), COALESCE(v.instagram_handle,''),
    v.google_rating, v.nobar_rating, v.nobar_rating_count,
    v.kondusif_score, v.kid_friendly_score, v.data_completeness,
    (SELECT COALESCE(array_agg(f.code ORDER BY f.code), '{}')
       FROM venue_facility vf JOIN facility f ON f.id = vf.facility_id
      WHERE vf.venue_id = v.id) AS facilities,
    -- COALESCE harus di LUAR subquery: kalau venue belum punya foto sama
    -- sekali, subquery skalar mengembalikan NULL dan COALESCE di dalamnya
    -- tidak pernah dievaluasi.
    COALESCE((SELECT p.url FROM venue_photo p
               WHERE p.venue_id = v.id
               ORDER BY p.is_primary DESC, p.sort_order, p.created_at
               LIMIT 1), '') AS primary_photo`

// SearchForMatch adalah query inti aplikasi (§9.1).
//
// Perhatikan bahwa skor dihitung DI DALAM SQL dan dipakai langsung di ORDER BY
// — itulah PERBAIKAN #3. Kalau skor dihitung di Go setelah LIMIT/OFFSET seperti
// di blueprint, yang diurutkan hanya venue terdekat sejumlah satu halaman,
// bukan seluruh kandidat dalam radius.
func (r *NobarEventRepo) SearchForMatch(ctx context.Context, p domain.NobarSearchParams) ([]domain.NobarEvent, int, error) {
	orderBy, err := orderClause(p.Sort)
	if err != nil {
		return nil, 0, err
	}

	// $1 titik pengguna, $2 radius meter — dirujuk juga oleh SQLScoreExpr.
	args := []any{
		fmt.Sprintf("SRID=4326;POINT(%f %f)", p.Lng, p.Lat), // $1
		p.RadiusKm * 1000, // $2
		p.MatchID,         // $3
		p.OpenUntilEnd,    // $4
		p.Dow,             // $5
		p.RequiredMinutes, // $6
		p.EntryType,       // $7
		p.Facilities,      // $8
		p.Limit,           // $9
		p.Offset,          // $10
	}

	query := `
SELECT ` + searchColumns + `,
    ST_Distance(v.location, $1) / 1000.0 AS distance_km,
    (` + scoring.SQLScoreExpr + `) AS score,
    COUNT(*) OVER() AS total
FROM nobar_event ne
JOIN venue v ON v.id = ne.venue_id
WHERE ne.match_id = $3
  AND ne.status IN ('published', 'confirmed')
  AND v.is_active = TRUE
  AND ST_DWithin(v.location, $1, $2)
  -- PERBAIKAN #2: filter jam tutup yang benar untuk venue lewat tengah malam.
  AND ($4 = FALSE OR venue_open_until(v.opening_hours, $5, $6))
  AND ($7 = '' OR ne.entry_type = $7)
  -- Venue harus punya SEMUA fasilitas yang diminta, bukan salah satunya.
  AND (cardinality($8::text[]) = 0 OR (
        SELECT count(DISTINCT f.code)
          FROM venue_facility vf JOIN facility f ON f.id = vf.facility_id
         WHERE vf.venue_id = v.id AND f.code = ANY($8::text[])
      ) = cardinality($8::text[]))
` + orderBy + `
LIMIT $9 OFFSET $10`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("cari nobar event: %w", err)
	}
	defer rows.Close()

	var (
		events []domain.NobarEvent
		total  int
	)
	for rows.Next() {
		e, t, err := scanSearchRow(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("baca hasil nobar event: %w", err)
	}
	return events, total, nil
}

// orderClause memetakan mode sort ke ORDER BY.
//
// Setiap mode diakhiri kolom penentu yang unik (ne.id) supaya urutannya stabil
// antar halaman: tanpa itu, dua venue berskor sama bisa bertukar posisi antara
// halaman 1 dan 2 dan salah satunya tidak pernah terlihat.
func orderClause(sort string) (string, error) {
	switch sort {
	case domain.SortRecommended, "":
		return "ORDER BY score DESC, distance_km ASC, ne.id", nil
	case domain.SortNearest:
		return "ORDER BY distance_km ASC, ne.id", nil
	case domain.SortRating:
		// Rating nobar dipakai kalau ada, kalau tidak jatuh ke rating Google —
		// venue tanpa keduanya turun ke bawah, bukan naik ke atas.
		return `ORDER BY COALESCE(v.nobar_rating, v.google_rating, 0) DESC,
                 v.nobar_rating_count DESC, distance_km ASC, ne.id`, nil
	default:
		return "", fmt.Errorf("%w: sort %q tidak dikenal", domain.ErrInvalidInput, sort)
	}
}

func scanSearchRow(rows pgx.Rows) (domain.NobarEvent, int, error) {
	var (
		e     domain.NobarEvent
		v     domain.Venue
		total int
	)
	err := rows.Scan(
		&e.ID, &e.VenueID, &e.MatchID, &e.DoorsOpenAt, &e.EntryType,
		&e.EntryAmount, &e.CapacityEstimate, &e.CrowdLevel,
		&e.Notes, &e.Status, &e.ConfirmedAt, &e.IsPromoted,
		&v.Name, &v.Slug, &v.District, &v.Address,
		&v.Lat, &v.Lng,
		&v.WhatsApp, &v.InstagramHandle,
		&v.GoogleRating, &v.NobarRating, &v.NobarRatingCount,
		&v.KondusifScore, &v.KidFriendlyScore, &v.DataCompleteness,
		&v.Facilities, &v.PrimaryPhoto,
		&e.DistanceKm, &e.Score, &total,
	)
	if err != nil {
		return e, 0, fmt.Errorf("scan nobar event: %w", err)
	}
	v.ID = e.VenueID
	e.Venue = &v
	return e, total, nil
}

// HistoryForVenue mengisi blok "riwayat nobar sebelumnya" di halaman detail (§13.4).
func (r *NobarEventRepo) HistoryForVenue(ctx context.Context, venueID uuid.UUID, limit int) ([]domain.NobarEvent, error) {
	const q = `
SELECT ne.id, ne.match_id, ne.entry_type, ne.entry_amount, ne.status,
       ne.confirmed_at, m.kickoff_at,
       ht.name, COALESCE(ht.short_name,''), at.name, COALESCE(at.short_name,'')
FROM nobar_event ne
JOIN match m  ON m.id = ne.match_id
JOIN team ht  ON ht.id = m.home_team_id
JOIN team at  ON at.id = m.away_team_id
WHERE ne.venue_id = $1
  AND ne.status IN ('published','confirmed','finished')
  AND m.kickoff_at < now()
ORDER BY m.kickoff_at DESC
LIMIT $2`

	rows, err := r.db.Query(ctx, q, venueID, limit)
	if err != nil {
		return nil, fmt.Errorf("riwayat nobar: %w", err)
	}
	defer rows.Close()

	var out []domain.NobarEvent
	for rows.Next() {
		var (
			e domain.NobarEvent
			m domain.Match
		)
		if err := rows.Scan(&e.ID, &e.MatchID, &e.EntryType, &e.EntryAmount,
			&e.Status, &e.ConfirmedAt, &m.KickoffAt,
			&m.HomeTeam.Name, &m.HomeTeam.ShortName,
			&m.AwayTeam.Name, &m.AwayTeam.ShortName); err != nil {
			return nil, fmt.Errorf("scan riwayat: %w", err)
		}
		m.ID = e.MatchID
		e.Match = &m
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *NobarEventRepo) Create(ctx context.Context, e *domain.NobarEvent) error {
	const q = `
INSERT INTO nobar_event (venue_id, match_id, doors_open_at, entry_type, entry_amount,
                         capacity_estimate, crowd_level, notes, status, confirmed_at,
                         is_promoted, created_by)
VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),$9,$10,$11,$12)
RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, q, e.VenueID, e.MatchID, e.DoorsOpenAt, e.EntryType,
		e.EntryAmount, e.CapacityEstimate, e.CrowdLevel, e.Notes, e.Status,
		e.ConfirmedAt, e.IsPromoted, e.CreatedBy,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)

	if isUniqueViolation(err) {
		// UNIQUE (venue_id, match_id) — §7.3 mencegah duplikasi saat impor IG
		// dan submit venue bertabrakan.
		return fmt.Errorf("%w: venue ini sudah punya pengumuman untuk laga tersebut", domain.ErrConflict)
	}
	if isCheckViolation(err) {
		return fmt.Errorf("%w: melanggar aturan %s", domain.ErrInvalidInput, pgErrConstraint(err))
	}
	if err != nil {
		return fmt.Errorf("simpan nobar event: %w", err)
	}
	return nil
}

func (r *NobarEventRepo) Update(ctx context.Context, e *domain.NobarEvent) error {
	const q = `
UPDATE nobar_event SET
    doors_open_at = $2, entry_type = $3, entry_amount = $4,
    capacity_estimate = $5, crowd_level = NULLIF($6,''), notes = NULLIF($7,''),
    is_promoted = $8
WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, e.ID, e.DoorsOpenAt, e.EntryType, e.EntryAmount,
		e.CapacityEstimate, e.CrowdLevel, e.Notes, e.IsPromoted)
	if err != nil {
		return fmt.Errorf("ubah nobar event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NobarEventRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, confirmedAt *time.Time) error {
	const q = `UPDATE nobar_event SET status = $2, confirmed_at = $3 WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, status, confirmedAt)
	if err != nil {
		return fmt.Errorf("ubah status nobar event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NobarEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.NobarEvent, error) {
	const q = `
SELECT id, venue_id, match_id, doors_open_at, entry_type, entry_amount,
       capacity_estimate, COALESCE(crowd_level,''), COALESCE(notes,''),
       status, confirmed_at, is_promoted, created_by, created_at, updated_at
FROM nobar_event WHERE id = $1`

	var e domain.NobarEvent
	err := r.db.QueryRow(ctx, q, id).Scan(&e.ID, &e.VenueID, &e.MatchID, &e.DoorsOpenAt,
		&e.EntryType, &e.EntryAmount, &e.CapacityEstimate, &e.CrowdLevel, &e.Notes,
		&e.Status, &e.ConfirmedAt, &e.IsPromoted, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ambil nobar event: %w", err)
	}
	return &e, nil
}

func (r *NobarEventRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM nobar_event WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("hapus nobar event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListForMatch dipakai worker dan panel admin untuk semua status.
func (r *NobarEventRepo) ListForMatch(ctx context.Context, matchID int64, statuses []string) ([]domain.NobarEvent, error) {
	if len(statuses) == 0 {
		statuses = domain.PublicStatuses
	}
	const q = `
SELECT id, venue_id, match_id, doors_open_at, entry_type, entry_amount,
       capacity_estimate, COALESCE(crowd_level,''), COALESCE(notes,''),
       status, confirmed_at, is_promoted, created_by, created_at, updated_at
FROM nobar_event
WHERE match_id = $1 AND status = ANY($2::text[])
ORDER BY created_at`

	rows, err := r.db.Query(ctx, q, matchID, statuses)
	if err != nil {
		return nil, fmt.Errorf("daftar nobar event: %w", err)
	}
	defer rows.Close()

	var out []domain.NobarEvent
	for rows.Next() {
		var e domain.NobarEvent
		if err := rows.Scan(&e.ID, &e.VenueID, &e.MatchID, &e.DoorsOpenAt,
			&e.EntryType, &e.EntryAmount, &e.CapacityEstimate, &e.CrowdLevel, &e.Notes,
			&e.Status, &e.ConfirmedAt, &e.IsPromoted, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan nobar event: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListPending mengisi antrian tinjauan admin (§14.2).
//
// Urutannya berdasarkan kickoff, bukan waktu submit: laga yang paling dekat
// harus ditinjau lebih dulu, karena setelah kickoff lewat, meninjaunya sudah
// tidak ada gunanya sama sekali.
func (r *NobarEventRepo) ListPending(ctx context.Context, limit int) ([]domain.NobarEvent, error) {
	const q = `
SELECT ne.id, ne.venue_id, ne.match_id, ne.doors_open_at, ne.entry_type,
       ne.entry_amount, COALESCE(ne.notes,''), ne.status, ne.created_at,
       v.name, v.slug, COALESCE(v.district,''),
       m.kickoff_at, ht.name, COALESCE(ht.short_name,''),
       at.name, COALESCE(at.short_name,'')
FROM nobar_event ne
JOIN venue v ON v.id = ne.venue_id
JOIN match m ON m.id = ne.match_id
JOIN team ht ON ht.id = m.home_team_id
JOIN team at ON at.id = m.away_team_id
WHERE ne.status IN ('draft', 'pending_review')
  AND m.kickoff_at > now() - INTERVAL '3 hours'
ORDER BY m.kickoff_at, ne.created_at
LIMIT $1`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("antrian tinjauan: %w", err)
	}
	defer rows.Close()

	var out []domain.NobarEvent
	for rows.Next() {
		var (
			e domain.NobarEvent
			v domain.Venue
			m domain.Match
		)
		if err := rows.Scan(&e.ID, &e.VenueID, &e.MatchID, &e.DoorsOpenAt, &e.EntryType,
			&e.EntryAmount, &e.Notes, &e.Status, &e.CreatedAt,
			&v.Name, &v.Slug, &v.District,
			&m.KickoffAt, &m.HomeTeam.Name, &m.HomeTeam.ShortName,
			&m.AwayTeam.Name, &m.AwayTeam.ShortName); err != nil {
			return nil, fmt.Errorf("scan antrian: %w", err)
		}
		v.ID, m.ID = e.VenueID, e.MatchID
		e.Venue, e.Match = &v, &m
		out = append(out, e)
	}
	return out, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514"
}

// pgErrConstraint mengambil nama constraint untuk pesan kesalahan yang berguna.
func pgErrConstraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.TrimPrefix(pgErr.ConstraintName, "nobar_event_")
	}
	return ""
}
