package domain

import (
	"context"
	"time"
)

// Status pertandingan (§7.2).
const (
	MatchScheduled = "scheduled"
	MatchLive      = "live"
	MatchFinished  = "finished"
	MatchPostponed = "postponed"
)

type Team struct {
	ID         int
	Name       string
	ShortName  string
	Slug       string
	LogoURL    string
	IsFeatured bool
}

type Competition struct {
	ID   int
	Name string
	Slug string
}

type Match struct {
	ID          int64
	Competition Competition
	HomeTeam    Team
	AwayTeam    Team
	KickoffAt   time.Time
	VenueName   string
	BroadcastTV string
	Status      string
	ScoreHome   *int
	ScoreAway   *int

	// NobarCount diisi query daftar laga; jumlah venue yang menayangkan.
	NobarCount int
}

// Label mengembalikan "Persija vs Persib" untuk meta response (§8.2).
func (m Match) Label() string {
	home := m.HomeTeam.ShortName
	if home == "" {
		home = m.HomeTeam.Name
	}
	away := m.AwayTeam.ShortName
	if away == "" {
		away = m.AwayTeam.Name
	}
	return home + " vs " + away
}

// UpcomingParams dipakai GET /v1/matches/upcoming (§8.2).
type UpcomingParams struct {
	TeamSlug string
	Limit    int
}

type MatchRepository interface {
	Create(ctx context.Context, m *Match) error
	Update(ctx context.Context, m *Match) error
	GetByID(ctx context.Context, id int64) (*Match, error)
	Upcoming(ctx context.Context, p UpcomingParams) ([]Match, error)
	Season(ctx context.Context, teamSlug string, from, to time.Time) ([]Match, error)
}
