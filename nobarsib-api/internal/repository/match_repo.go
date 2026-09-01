package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

type MatchRepo struct{ db *pgxpool.Pool }

func NewMatchRepo(db *pgxpool.Pool) *MatchRepo { return &MatchRepo{db: db} }

// matchColumns menyertakan nobar_count: jumlah venue yang menayangkan laga itu.
// Dihitung sebagai subquery berkorelasi supaya laga tanpa nobar tetap muncul
// dengan angka 0 — §13.5 mensyaratkan keadaan "ada laga, belum ada venue"
// punya tampilannya sendiri, jadi datanya harus sampai ke frontend.
const matchColumns = `
    m.id, m.kickoff_at, COALESCE(m.venue_name,''), COALESCE(m.broadcast_tv,''),
    m.status, m.score_home, m.score_away,
    COALESCE(c.id,0), COALESCE(c.name,''), COALESCE(c.slug,''),
    ht.id, ht.name, COALESCE(ht.short_name,''), ht.slug, COALESCE(ht.logo_url,''),
    at.id, at.name, COALESCE(at.short_name,''), at.slug, COALESCE(at.logo_url,''),
    (SELECT count(*) FROM nobar_event ne
      WHERE ne.match_id = m.id AND ne.status IN ('published','confirmed'))`

const matchFrom = `
FROM match m
JOIN team ht ON ht.id = m.home_team_id
JOIN team at ON at.id = m.away_team_id
LEFT JOIN competition c ON c.id = m.competition_id`

func scanMatch(row pgx.Row) (*domain.Match, error) {
	var m domain.Match
	err := row.Scan(&m.ID, &m.KickoffAt, &m.VenueName, &m.BroadcastTV,
		&m.Status, &m.ScoreHome, &m.ScoreAway,
		&m.Competition.ID, &m.Competition.Name, &m.Competition.Slug,
		&m.HomeTeam.ID, &m.HomeTeam.Name, &m.HomeTeam.ShortName, &m.HomeTeam.Slug, &m.HomeTeam.LogoURL,
		&m.AwayTeam.ID, &m.AwayTeam.Name, &m.AwayTeam.ShortName, &m.AwayTeam.Slug, &m.AwayTeam.LogoURL,
		&m.NobarCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan match: %w", err)
	}
	return &m, nil
}

// Upcoming mengembalikan laga mendatang untuk satu tim (§8.2).
//
// Laga yang sedang berlangsung ikut ditampilkan: orang yang membuka aplikasi
// jam 19.30 saat laga sudah mulai tetap butuh tahu di mana bisa menonton.
// Batasnya kickoff + 3 jam, cukup untuk 90 menit plus jeda dan tambahan waktu.
func (r *MatchRepo) Upcoming(ctx context.Context, p domain.UpcomingParams) ([]domain.Match, error) {
	q := `SELECT ` + matchColumns + matchFrom + `
WHERE m.kickoff_at > now() - INTERVAL '3 hours'
  AND m.status <> 'finished'
  AND ($1 = '' OR ht.slug = $1 OR at.slug = $1)
ORDER BY m.kickoff_at
LIMIT $2`

	rows, err := r.db.Query(ctx, q, p.TeamSlug, p.Limit)
	if err != nil {
		return nil, fmt.Errorf("daftar laga mendatang: %w", err)
	}
	defer rows.Close()
	return collectMatches(rows)
}

// Season mengembalikan seluruh laga dalam rentang waktu, untuk halaman /jadwal.
func (r *MatchRepo) Season(ctx context.Context, teamSlug string, from, to time.Time) ([]domain.Match, error) {
	q := `SELECT ` + matchColumns + matchFrom + `
WHERE m.kickoff_at BETWEEN $2 AND $3
  AND ($1 = '' OR ht.slug = $1 OR at.slug = $1)
ORDER BY m.kickoff_at`

	rows, err := r.db.Query(ctx, q, teamSlug, from, to)
	if err != nil {
		return nil, fmt.Errorf("jadwal musim: %w", err)
	}
	defer rows.Close()
	return collectMatches(rows)
}

func collectMatches(rows pgx.Rows) ([]domain.Match, error) {
	var out []domain.Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *MatchRepo) GetByID(ctx context.Context, id int64) (*domain.Match, error) {
	return scanMatch(r.db.QueryRow(ctx, `SELECT `+matchColumns+matchFrom+` WHERE m.id = $1`, id))
}

func (r *MatchRepo) Create(ctx context.Context, m *domain.Match) error {
	const q = `
INSERT INTO match (competition_id, home_team_id, away_team_id, kickoff_at,
                   venue_name, broadcast_tv, status)
VALUES (NULLIF($1,0),$2,$3,$4,NULLIF($5,''),NULLIF($6,''),$7)
RETURNING id`

	err := r.db.QueryRow(ctx, q, m.Competition.ID, m.HomeTeam.ID, m.AwayTeam.ID,
		m.KickoffAt, m.VenueName, m.BroadcastTV, m.Status).Scan(&m.ID)
	if isCheckViolation(err) {
		return fmt.Errorf("%w: tim tuan rumah dan tamu tidak boleh sama", domain.ErrInvalidInput)
	}
	if err != nil {
		return fmt.Errorf("simpan match: %w", err)
	}
	return nil
}

func (r *MatchRepo) Update(ctx context.Context, m *domain.Match) error {
	const q = `
UPDATE match SET kickoff_at = $2, venue_name = NULLIF($3,''),
       broadcast_tv = NULLIF($4,''), status = $5, score_home = $6, score_away = $7
WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, m.ID, m.KickoffAt, m.VenueName, m.BroadcastTV,
		m.Status, m.ScoreHome, m.ScoreAway)
	if err != nil {
		return fmt.Errorf("ubah match: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// TeamBySlug dipakai saat memvalidasi parameter team_slug.
func (r *MatchRepo) TeamBySlug(ctx context.Context, slug string) (*domain.Team, error) {
	var t domain.Team
	err := r.db.QueryRow(ctx,
		`SELECT id, name, COALESCE(short_name,''), slug, COALESCE(logo_url,''), is_featured
		 FROM team WHERE slug = $1`, slug).
		Scan(&t.ID, &t.Name, &t.ShortName, &t.Slug, &t.LogoURL, &t.IsFeatured)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ambil tim: %w", err)
	}
	return &t, nil
}
