package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Status nobar_event (§4.5).
//
// Ditulis huruf kecil, konsisten dengan DEFAULT kolom di §7.2 dan CHECK
// constraint di migrasi 000005. Diagram §4.5 menulisnya kapital — itu hanya
// gaya penulisan diagram, bukan nilai yang disimpan.
const (
	EventDraft         = "draft"
	EventPendingReview = "pending_review"
	EventPublished     = "published"
	EventConfirmed     = "confirmed"
	EventRejected      = "rejected"
	EventCancelled     = "cancelled"
	EventFinished      = "finished"
)

// Jenis biaya masuk (§7.2).
const (
	EntryFree     = "free"
	EntryMinOrder = "min_order"
	EntryTicket   = "ticket"
	EntryDonation = "donation"
)

// allowedTransitions adalah state machine §4.5:
//
//	DRAFT ──> PENDING_REVIEW ──> PUBLISHED ──> CONFIRMED ──> FINISHED
//	              │                  │             │
//	              └──> REJECTED      └──> CANCELLED└──> CANCELLED
//
// Status akhir (rejected, cancelled, finished) tidak punya tujuan lanjutan.
// Event yang sudah dibatalkan tidak bisa dihidupkan lagi — venue harus membuat
// pengumuman baru, supaya jejak pembatalannya tidak hilang.
var allowedTransitions = map[string][]string{
	EventDraft:         {EventPendingReview, EventCancelled},
	EventPendingReview: {EventPublished, EventRejected, EventCancelled},
	EventPublished:     {EventConfirmed, EventCancelled, EventFinished},
	EventConfirmed:     {EventCancelled, EventFinished},
	EventRejected:      {},
	EventCancelled:     {},
	EventFinished:      {},
}

// CanTransition menjawab apakah perpindahan status diizinkan.
func CanTransition(from, to string) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// ValidateTransition mengembalikan ErrInvalidTransition beserta keterangan
// status mana yang sebenarnya boleh dituju.
func ValidateTransition(from, to string) error {
	if from == to {
		return fmt.Errorf("%w: status sudah %q", ErrInvalidTransition, to)
	}
	if _, known := allowedTransitions[from]; !known {
		return fmt.Errorf("%w: status asal %q tidak dikenal", ErrInvalidTransition, from)
	}
	if CanTransition(from, to) {
		return nil
	}
	allowed := allowedTransitions[from]
	if len(allowed) == 0 {
		return fmt.Errorf("%w: %q adalah status akhir", ErrInvalidTransition, from)
	}
	return fmt.Errorf("%w: dari %q hanya boleh ke %v", ErrInvalidTransition, from, allowed)
}

// PublicStatuses adalah status yang tampil ke publik (§4.5).
var PublicStatuses = []string{EventPublished, EventConfirmed}

type NobarEvent struct {
	ID      uuid.UUID
	VenueID uuid.UUID
	MatchID int64

	DoorsOpenAt      *time.Time
	EntryType        string
	EntryAmount      int
	CapacityEstimate *int
	CrowdLevel       string
	Notes            string

	Status      string
	ConfirmedAt *time.Time
	IsPromoted  bool

	CreatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	// Diisi oleh query daftar; bukan kolom tabel.
	Venue      *Venue
	Match      *Match
	DistanceKm float64
	Score      float64
}

// IsConfirmed dipakai badge "Dikonfirmasi" (§13.3) dan komponen skor §9.2.
func (e NobarEvent) IsConfirmed() bool { return e.ConfirmedAt != nil }

// Validate memeriksa aturan yang juga dijaga CHECK constraint di migrasi 000005,
// supaya pesan kesalahannya bisa dibaca manusia sebelum menyentuh database.
func (e NobarEvent) Validate() error {
	switch e.EntryType {
	case EntryFree, EntryMinOrder, EntryTicket, EntryDonation:
	default:
		return fmt.Errorf("%w: entry_type %q tidak dikenal", ErrInvalidInput, e.EntryType)
	}
	if e.EntryAmount < 0 {
		return fmt.Errorf("%w: entry_amount tidak boleh negatif", ErrInvalidInput)
	}
	// Kartu venue tidak boleh menampilkan "Gratis" dan "Rp 25.000" sekaligus
	// (§3.1 no. 5: jangan janji yang tidak bisa ditepati).
	if e.EntryType == EntryFree && e.EntryAmount != 0 {
		return fmt.Errorf("%w: entry_type free harus entry_amount 0", ErrInvalidInput)
	}
	switch e.CrowdLevel {
	case "", "longgar", "ramai", "penuh":
	default:
		return fmt.Errorf("%w: crowd_level %q tidak dikenal", ErrInvalidInput, e.CrowdLevel)
	}
	return nil
}

// SortMode untuk GET /v1/matches/{id}/nobar (§8.2).
const (
	SortRecommended = "recommended"
	SortNearest     = "nearest"
	SortRating      = "rating"
)

// NobarSearchParams adalah seluruh query param endpoint utama aplikasi.
type NobarSearchParams struct {
	MatchID int64

	Lat float64
	Lng float64

	Sort       string
	RadiusKm   float64
	Facilities []string
	EntryType  string

	// OpenUntilEnd menyaring venue yang tutup sebelum laga selesai (§9.4).
	OpenUntilEnd bool

	// Dow dan RequiredMinutes diturunkan dari kickoff dalam zona WIB oleh
	// service, bukan dihitung di SQL — supaya urusan zona waktu hanya ada di
	// satu tempat.
	Dow             int
	RequiredMinutes int

	Limit  int
	Offset int
}

type NobarEventRepository interface {
	Create(ctx context.Context, e *NobarEvent) error
	Update(ctx context.Context, e *NobarEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*NobarEvent, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateStatus mengubah status beserta confirmed_at dalam satu perintah.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, confirmedAt *time.Time) error

	// SearchForMatch adalah query inti aplikasi (§9.1). Mengembalikan hasil
	// satu halaman beserta jumlah total sebelum paginasi.
	SearchForMatch(ctx context.Context, p NobarSearchParams) ([]NobarEvent, int, error)

	// HistoryForVenue dipakai blok "riwayat nobar" di halaman detail (§13.4).
	HistoryForVenue(ctx context.Context, venueID uuid.UUID, limit int) ([]NobarEvent, error)
}
