package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/pkg/openhours"
)

// Status venue (§7.2).
const (
	VenueUnclaimed = "unclaimed"
	VenueClaimed   = "claimed"
	VenueVerified  = "verified"
	VenueSuspended = "suspended"
)

type Venue struct {
	ID       uuid.UUID
	Name     string
	Slug     string
	Address  string
	District string
	City     string
	Lat      float64
	Lng      float64

	Phone           string
	WhatsApp        string
	InstagramHandle string
	GooglePlaceID   string

	GoogleRating      *float64
	GoogleRatingCount *int
	OpeningHours      openhours.Week

	NobarRating      *float64
	NobarRatingCount int
	KondusifScore    *float64
	KidFriendlyScore *float64
	DataCompleteness float64

	Status    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time

	// Diisi hanya oleh query yang memuatnya.
	Facilities   []string
	PrimaryPhoto string
	Photos       []VenuePhoto
}

// VisibleKondusif menerapkan §11.4: skor kondusif tidak ditampilkan sebelum
// ada 3 review. Sebelum itu "Belum ada penilaian" lebih jujur daripada angka
// dari satu orang yang kebetulan sedang kesal.
func (v Venue) VisibleKondusif() *float64 {
	if v.NobarRatingCount < MinReviewsForScore {
		return nil
	}
	return v.KondusifScore
}

// VisibleKidFriendly memakai ambang yang sama dengan VisibleKondusif.
func (v Venue) VisibleKidFriendly() *float64 {
	if v.NobarRatingCount < MinReviewsForScore {
		return nil
	}
	return v.KidFriendlyScore
}

// MinReviewsForScore adalah ambang tampil skor turunan review (§11.4).
const MinReviewsForScore = 3

type VenuePhoto struct {
	ID        uuid.UUID
	URL       string
	Caption   string
	IsPrimary bool
	SortOrder int
}

type Facility struct {
	Code     string
	Label    string
	Icon     string
	Category string
}

// VenueSearchParams dipakai GET /v1/venues/search (§8.2).
type VenueSearchParams struct {
	Query  string
	Limit  int
	Offset int
	Lat    *float64
	Lng    *float64
}

type VenueRepository interface {
	Create(ctx context.Context, v *Venue) error
	Update(ctx context.Context, v *Venue) error
	GetByID(ctx context.Context, id uuid.UUID) (*Venue, error)
	GetBySlug(ctx context.Context, slug string) (*Venue, error)
	Search(ctx context.Context, p VenueSearchParams) ([]Venue, int, error)

	SetFacilities(ctx context.Context, venueID uuid.UUID, codes []string) error
	AddPhoto(ctx context.Context, venueID uuid.UUID, p *VenuePhoto) error
	DeletePhoto(ctx context.Context, photoID uuid.UUID) error
	ListFacilities(ctx context.Context) ([]Facility, error)

	// RecalculateCompleteness menyegarkan venue.data_completeness (§9.3).
	// Dipanggil setelah setiap perubahan profil, foto, atau fasilitas supaya
	// skornya tidak pernah basi terhadap isinya.
	RecalculateCompleteness(ctx context.Context, venueID uuid.UUID) (float64, error)
}
