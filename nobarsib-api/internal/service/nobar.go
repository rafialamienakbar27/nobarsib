// Package service berisi logika bisnis (blueprint §6.4).
//
// Aturan yang tidak boleh dilanggar siapa pun — state machine §4.5, ambang
// tampil skor §11.4, zona waktu kickoff — tinggal di sini, bukan di handler.
// Handler boleh berganti (REST, gRPC, worker), aturannya tidak.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

// WIB adalah zona waktu seluruh aplikasi. Semua timestamp API memakai +07:00
// (§8.1), dan hari operasional venue dihitung menurut zona ini — bukan menurut
// UTC, yang akan menggeser laga malam ke hari berikutnya.
var WIB = time.FixedZone("WIB", 7*60*60)

// Default pencarian (§8.2).
const (
	DefaultRadiusKm = 15.0
	MaxRadiusKm     = 50.0
	DefaultPerPage  = 20
	MaxPerPage      = 50
	DefaultTeamSlug = "persib-bandung"
	DefaultUpcoming = 5
)

// MatchWrapUpMinutes adalah 90 menit pertandingan + 30 menit bubar (§9.4).
const MatchWrapUpMinutes = 120

// Pusat kota Bandung, dipakai kalau pengguna menolak izin lokasi (§4.2).
const (
	BandungLat = -6.9175
	BandungLng = 107.6191
)

type NobarService struct {
	events  repoEvents
	matches repoMatches
	cache   cacheInvalidator
}

// cacheInvalidator dipenuhi *NobarCache. Dibuat sebagai antarmuka supaya
// service tetap bisa diuji tanpa Redis.
type cacheInvalidator interface {
	InvalidateMatch(ctx context.Context, matchID int64)
}

// Antarmuka sempit: service hanya menyebut yang benar-benar dipakainya,
// sehingga test bisa memalsukan sedikit method saja.
type repoEvents interface {
	SearchForMatch(ctx context.Context, p domain.NobarSearchParams) ([]domain.NobarEvent, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.NobarEvent, error)
	Create(ctx context.Context, e *domain.NobarEvent) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, confirmedAt *time.Time) error
}

type repoMatches interface {
	GetByID(ctx context.Context, id int64) (*domain.Match, error)
}

func NewNobarService(events repoEvents, matches repoMatches, cache cacheInvalidator) *NobarService {
	return &NobarService{events: events, matches: matches, cache: cache}
}

// SearchResult adalah hasil endpoint utama beserta konteks laganya.
type SearchResult struct {
	Events []domain.NobarEvent
	Total  int
	Match  *domain.Match
}

// Search menjalankan pencarian venue untuk satu laga (§8.2).
//
// Perhitungan hari dan menit untuk filter jam tutup dilakukan di sini, dalam
// WIB, lalu diteruskan ke SQL sebagai angka. Urusan zona waktu hanya ada di
// satu tempat, dan fungsi SQL-nya tetap bisa diuji dengan angka biasa.
func (s *NobarService) Search(ctx context.Context, p domain.NobarSearchParams) (*SearchResult, error) {
	match, err := s.matches.GetByID(ctx, p.MatchID)
	if err != nil {
		return nil, err
	}

	kickoff := match.KickoffAt.In(WIB)
	p.Dow = int(kickoff.Weekday()) // time.Sunday == 0, sama dengan EXTRACT(DOW)
	p.RequiredMinutes = kickoff.Hour()*60 + kickoff.Minute() + MatchWrapUpMinutes

	events, total, err := s.events.SearchForMatch(ctx, p)
	if err != nil {
		return nil, err
	}
	return &SearchResult{Events: events, Total: total, Match: match}, nil
}

// NormalizeSearchParams menerapkan nilai default dan batas atas.
//
// Batas dipasang di service, bukan di handler, supaya worker dan panel admin
// yang memanggil jalur yang sama tidak bisa melewatinya.
func NormalizeSearchParams(p domain.NobarSearchParams) (domain.NobarSearchParams, error) {
	if p.Lat == 0 && p.Lng == 0 {
		p.Lat, p.Lng = BandungLat, BandungLng
	}
	if p.Lat < -90 || p.Lat > 90 || p.Lng < -180 || p.Lng > 180 {
		return p, fmt.Errorf("%w: koordinat di luar rentang", domain.ErrInvalidInput)
	}

	if p.RadiusKm <= 0 {
		p.RadiusKm = DefaultRadiusKm
	}
	if p.RadiusKm > MaxRadiusKm {
		p.RadiusKm = MaxRadiusKm
	}

	switch p.Sort {
	case "", domain.SortRecommended, domain.SortNearest, domain.SortRating:
	default:
		return p, fmt.Errorf("%w: sort %q tidak dikenal", domain.ErrInvalidInput, p.Sort)
	}
	if p.Sort == "" {
		p.Sort = domain.SortRecommended
	}

	switch p.EntryType {
	case "", domain.EntryFree, domain.EntryMinOrder, domain.EntryTicket, domain.EntryDonation:
	default:
		return p, fmt.Errorf("%w: entry_type %q tidak dikenal", domain.ErrInvalidInput, p.EntryType)
	}

	if p.Limit <= 0 {
		p.Limit = DefaultPerPage
	}
	if p.Limit > MaxPerPage {
		p.Limit = MaxPerPage
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Facilities == nil {
		p.Facilities = []string{}
	}
	return p, nil
}

// Transition memindahkan status event sesuai state machine §4.5.
//
// Inilah satu-satunya jalan mengubah status. Repository hanya menulis apa yang
// diperintahkan; keputusan boleh atau tidaknya diambil di sini.
func (s *NobarService) Transition(ctx context.Context, id uuid.UUID, to string) (*domain.NobarEvent, error) {
	e, err := s.events.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := domain.ValidateTransition(e.Status, to); err != nil {
		return nil, err
	}

	confirmedAt := e.ConfirmedAt
	if to == domain.EventConfirmed {
		// CHECK constraint di migrasi 000005 menolak status confirmed tanpa
		// timestamp; badge "Dikonfirmasi" (§13.3) harus selalu bisa ditelusuri
		// ke satu waktu.
		now := time.Now().In(WIB)
		confirmedAt = &now
	}

	if err := s.events.UpdateStatus(ctx, id, to, confirmedAt); err != nil {
		return nil, err
	}
	e.Status = to
	e.ConfirmedAt = confirmedAt

	// Daftar venue untuk laga ini sudah berubah. Tanpa pembatalan cache, badge
	// "Dikonfirmasi" baru terlihat setelah TTL habis — dan justru di jam-jam
	// menjelang kickoff itulah statusnya paling menentukan (§15.3).
	if s.cache != nil {
		s.cache.InvalidateMatch(ctx, e.MatchID)
	}
	return e, nil
}

// Create menyimpan pengumuman nobar baru setelah divalidasi.
func (s *NobarService) Create(ctx context.Context, e *domain.NobarEvent) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if e.Status == "" {
		e.Status = domain.EventDraft
	}
	if _, known := map[string]bool{
		domain.EventDraft: true, domain.EventPendingReview: true, domain.EventPublished: true,
	}[e.Status]; !known {
		return fmt.Errorf("%w: event baru tidak boleh langsung berstatus %q",
			domain.ErrInvalidInput, e.Status)
	}
	return s.events.Create(ctx, e)
}
