package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

func TestNormalizeSearchParamsDefault(t *testing.T) {
	p, err := NormalizeSearchParams(domain.NobarSearchParams{MatchID: 1})
	if err != nil {
		t.Fatalf("NormalizeSearchParams() error: %v", err)
	}

	// §4.2: lokasi ditolak -> jatuh ke pusat kota Bandung.
	if p.Lat != BandungLat || p.Lng != BandungLng {
		t.Errorf("fallback lokasi = (%v, %v), mau pusat Bandung", p.Lat, p.Lng)
	}
	if p.RadiusKm != DefaultRadiusKm {
		t.Errorf("radius default = %v, mau %v", p.RadiusKm, DefaultRadiusKm)
	}
	if p.Sort != domain.SortRecommended {
		t.Errorf("sort default = %q, mau %q", p.Sort, domain.SortRecommended)
	}
	if p.Limit != DefaultPerPage {
		t.Errorf("per_page default = %d, mau %d", p.Limit, DefaultPerPage)
	}
	if p.Facilities == nil {
		t.Error("Facilities nil; SQL cardinality($8) butuh slice kosong, bukan nil")
	}
}

// Batas atas dipasang di service supaya pemanggil mana pun tidak bisa
// meminta radius 5000 km atau 10.000 baris sekaligus.
func TestNormalizeSearchParamsBatasAtas(t *testing.T) {
	p, err := NormalizeSearchParams(domain.NobarSearchParams{
		MatchID: 1, RadiusKm: 9999, Limit: 9999, Lat: -6.9, Lng: 107.6,
	})
	if err != nil {
		t.Fatalf("NormalizeSearchParams() error: %v", err)
	}
	if p.RadiusKm != MaxRadiusKm {
		t.Errorf("radius = %v, mau dibatasi ke %v", p.RadiusKm, MaxRadiusKm)
	}
	if p.Limit != MaxPerPage {
		t.Errorf("per_page = %d, mau dibatasi ke %d", p.Limit, MaxPerPage)
	}
}

func TestNormalizeSearchParamsTolakNilaiNgawur(t *testing.T) {
	cases := []struct {
		nama string
		in   domain.NobarSearchParams
	}{
		{"sort tidak dikenal", domain.NobarSearchParams{Sort: "jauh"}},
		{"entry_type tidak dikenal", domain.NobarSearchParams{EntryType: "gratis"}},
		{"lintang di luar rentang", domain.NobarSearchParams{Lat: 95, Lng: 107}},
		{"bujur di luar rentang", domain.NobarSearchParams{Lat: -6.9, Lng: 200}},
	}
	for _, c := range cases {
		if _, err := NormalizeSearchParams(c.in); !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("%s: error = %v, mau ErrInvalidInput", c.nama, err)
		}
	}
}

// Hari dan menit untuk filter jam tutup dihitung dalam WIB, bukan UTC.
// Kickoff Sabtu 19:00 WIB adalah Sabtu 12:00 UTC — kalau dihitung dalam UTC,
// harinya masih benar, tapi menitnya menjadi 720 dan seluruh filter jam tutup
// bergeser tujuh jam.
func TestSearchMenghitungDowDanMenitDalamWIB(t *testing.T) {
	kickoff := time.Date(2026, 9, 5, 19, 0, 0, 0, WIB) // Sabtu
	events := &fakeEvents{}
	svc := NewNobarService(events, &fakeMatches{kickoff: kickoff}, nil)

	if _, err := svc.Search(context.Background(), domain.NobarSearchParams{MatchID: 1}); err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if events.got.Dow != 6 {
		t.Errorf("Dow = %d, mau 6 (Sabtu)", events.got.Dow)
	}
	want := 19*60 + MatchWrapUpMinutes // 1260
	if events.got.RequiredMinutes != want {
		t.Errorf("RequiredMinutes = %d, mau %d", events.got.RequiredMinutes, want)
	}
}

// Kickoff yang disimpan dalam UTC harus tetap menghasilkan hari WIB yang benar.
// 2026-09-05T17:00:00Z adalah Minggu 00:00 WIB — harinya 0, bukan 6.
func TestSearchKickoffUTCDikonversiKeWIB(t *testing.T) {
	kickoff := time.Date(2026, 9, 5, 17, 0, 0, 0, time.UTC)
	events := &fakeEvents{}
	svc := NewNobarService(events, &fakeMatches{kickoff: kickoff}, nil)

	if _, err := svc.Search(context.Background(), domain.NobarSearchParams{MatchID: 1}); err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if events.got.Dow != 0 {
		t.Errorf("Dow = %d, mau 0 (Minggu dini hari WIB)", events.got.Dow)
	}
	if events.got.RequiredMinutes != MatchWrapUpMinutes {
		t.Errorf("RequiredMinutes = %d, mau %d", events.got.RequiredMinutes, MatchWrapUpMinutes)
	}
}

func TestTransitionMengisiConfirmedAt(t *testing.T) {
	id := uuid.New()
	events := &fakeEvents{event: &domain.NobarEvent{
		ID: id, MatchID: 42, Status: domain.EventPublished,
	}}
	cache := &fakeCache{}
	svc := NewNobarService(events, &fakeMatches{}, cache)

	e, err := svc.Transition(context.Background(), id, domain.EventConfirmed)
	if err != nil {
		t.Fatalf("Transition() error: %v", err)
	}
	// CHECK constraint migrasi 000005 menolak confirmed tanpa timestamp.
	if e.ConfirmedAt == nil {
		t.Error("ConfirmedAt masih nil setelah konfirmasi")
	}
	if cache.invalidated != 42 {
		t.Errorf("cache laga %d dibatalkan, mau 42", cache.invalidated)
	}
}

func TestTransitionMenolakLompatanTidakSah(t *testing.T) {
	id := uuid.New()
	events := &fakeEvents{event: &domain.NobarEvent{ID: id, Status: domain.EventDraft}}
	svc := NewNobarService(events, &fakeMatches{}, &fakeCache{})

	if _, err := svc.Transition(context.Background(), id, domain.EventConfirmed); !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("error = %v, mau ErrInvalidTransition", err)
	}
	if events.statusUpdated {
		t.Error("status tetap ditulis ke database meski transisi ditolak")
	}
}

func TestCreateMenolakStatusAwalYangTidakSah(t *testing.T) {
	svc := NewNobarService(&fakeEvents{}, &fakeMatches{}, nil)
	e := &domain.NobarEvent{EntryType: domain.EntryFree, Status: domain.EventConfirmed}

	if err := svc.Create(context.Background(), e); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("error = %v, mau ErrInvalidInput", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Sekawan Kopi & Space":    "sekawan-kopi-space",
		"150 Coffee Garden":       "150-coffee-garden",
		"RJ's Sports Bar & Grill": "rj-s-sports-bar-grill",
		"  Spasi Berlebih  ":      "spasi-berlebih",
		"Jabarano Coffee 4.0":     "jabarano-coffee-4-0",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, mau %q", in, got, want)
		}
	}
}

// ------------------------------------------------------------------ palsu

type fakeEvents struct {
	got           domain.NobarSearchParams
	event         *domain.NobarEvent
	statusUpdated bool
}

func (f *fakeEvents) SearchForMatch(_ context.Context, p domain.NobarSearchParams) ([]domain.NobarEvent, int, error) {
	f.got = p
	return nil, 0, nil
}

func (f *fakeEvents) GetByID(context.Context, uuid.UUID) (*domain.NobarEvent, error) {
	if f.event == nil {
		return nil, domain.ErrNotFound
	}
	return f.event, nil
}

func (f *fakeEvents) Create(context.Context, *domain.NobarEvent) error { return nil }

func (f *fakeEvents) UpdateStatus(context.Context, uuid.UUID, string, *time.Time) error {
	f.statusUpdated = true
	return nil
}

type fakeMatches struct{ kickoff time.Time }

func (f *fakeMatches) GetByID(context.Context, int64) (*domain.Match, error) {
	return &domain.Match{ID: 1, KickoffAt: f.kickoff}, nil
}

type fakeCache struct{ invalidated int64 }

func (f *fakeCache) InvalidateMatch(_ context.Context, matchID int64) { f.invalidated = matchID }
