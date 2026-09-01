package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

// Batas pencarian venue (§8.2).
const (
	MinSearchQuery  = 2
	DefaultVenueTop = 20
	RecentReviews   = 5 // §13.4 blok 8: "review terbaru (5 teratas)"
	HistoryLimit    = 5 // §13.4 blok 9: riwayat nobar sebelumnya
)

type repoVenues interface {
	GetBySlug(ctx context.Context, slug string) (*domain.Venue, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Venue, error)
	Search(ctx context.Context, p domain.VenueSearchParams) ([]domain.Venue, int, error)
	Create(ctx context.Context, v *domain.Venue) error
	Update(ctx context.Context, v *domain.Venue) error
	SetFacilities(ctx context.Context, venueID uuid.UUID, codes []string) error
	AddPhoto(ctx context.Context, venueID uuid.UUID, p *domain.VenuePhoto) error
	DeletePhoto(ctx context.Context, photoID uuid.UUID) error
	ListFacilities(ctx context.Context) ([]domain.Facility, error)
	RecalculateCompleteness(ctx context.Context, venueID uuid.UUID) (float64, error)
}

type repoReviews interface {
	ListByVenue(ctx context.Context, venueID uuid.UUID, limit int) ([]domain.Review, error)
}

type repoHistory interface {
	HistoryForVenue(ctx context.Context, venueID uuid.UUID, limit int) ([]domain.NobarEvent, error)
}

type VenueService struct {
	venues  repoVenues
	reviews repoReviews
	history repoHistory
}

func NewVenueService(v repoVenues, r repoReviews, h repoHistory) *VenueService {
	return &VenueService{venues: v, reviews: r, history: h}
}

// VenueDetail adalah isi halaman detail venue (§13.4).
type VenueDetail struct {
	Venue   *domain.Venue
	Reviews []domain.Review
	History []domain.NobarEvent
}

func (s *VenueService) Detail(ctx context.Context, slug string) (*VenueDetail, error) {
	v, err := s.venues.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	reviews, err := s.reviews.ListByVenue(ctx, v.ID, RecentReviews)
	if err != nil {
		return nil, err
	}
	history, err := s.history.HistoryForVenue(ctx, v.ID, HistoryLimit)
	if err != nil {
		return nil, err
	}
	return &VenueDetail{Venue: v, Reviews: reviews, History: history}, nil
}

func (s *VenueService) Search(ctx context.Context, p domain.VenueSearchParams) ([]domain.Venue, int, error) {
	p.Query = strings.TrimSpace(p.Query)
	if len(p.Query) < MinSearchQuery {
		return nil, 0, fmt.Errorf("%w: kata kunci minimal %d huruf", domain.ErrInvalidInput, MinSearchQuery)
	}
	if p.Limit <= 0 {
		p.Limit = DefaultVenueTop
	}
	if p.Limit > MaxPerPage {
		p.Limit = MaxPerPage
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return s.venues.Search(ctx, p)
}

// Create menyimpan venue baru dan langsung menghitung kelengkapan datanya.
func (s *VenueService) Create(ctx context.Context, v *domain.Venue, facilities []string) error {
	if err := validateVenue(v); err != nil {
		return err
	}
	if v.Slug == "" {
		v.Slug = Slugify(v.Name)
	}
	if v.City == "" {
		v.City = "Kota Bandung"
	}
	if v.Status == "" {
		v.Status = domain.VenueUnclaimed
	}

	if err := s.venues.Create(ctx, v); err != nil {
		return err
	}
	if len(facilities) > 0 {
		if err := s.venues.SetFacilities(ctx, v.ID, facilities); err != nil {
			return err
		}
	}
	score, err := s.venues.RecalculateCompleteness(ctx, v.ID)
	if err != nil {
		return err
	}
	v.DataCompleteness = score
	return nil
}

// Update mengubah profil venue. Kelengkapan data dihitung ulang setiap kali,
// supaya skor peringkat tidak pernah basi terhadap isinya (§9.3).
func (s *VenueService) Update(ctx context.Context, v *domain.Venue, facilities []string) error {
	if err := validateVenue(v); err != nil {
		return err
	}
	if err := s.venues.Update(ctx, v); err != nil {
		return err
	}
	if facilities != nil {
		if err := s.venues.SetFacilities(ctx, v.ID, facilities); err != nil {
			return err
		}
	}
	score, err := s.venues.RecalculateCompleteness(ctx, v.ID)
	if err != nil {
		return err
	}
	v.DataCompleteness = score
	return nil
}

// Facilities mengembalikan daftar fasilitas untuk chip filter di beranda (§13.2).
func (s *VenueService) Facilities(ctx context.Context) ([]domain.Facility, error) {
	return s.venues.ListFacilities(ctx)
}

func (s *VenueService) AddPhoto(ctx context.Context, venueID uuid.UUID, p *domain.VenuePhoto) error {
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("%w: url foto wajib diisi", domain.ErrInvalidInput)
	}
	if err := s.venues.AddPhoto(ctx, venueID, p); err != nil {
		return err
	}
	_, err := s.venues.RecalculateCompleteness(ctx, venueID)
	return err
}

func (s *VenueService) DeletePhoto(ctx context.Context, venueID, photoID uuid.UUID) error {
	if err := s.venues.DeletePhoto(ctx, photoID); err != nil {
		return err
	}
	_, err := s.venues.RecalculateCompleteness(ctx, venueID)
	return err
}

func validateVenue(v *domain.Venue) error {
	if strings.TrimSpace(v.Name) == "" {
		return fmt.Errorf("%w: nama venue wajib diisi", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(v.Address) == "" {
		return fmt.Errorf("%w: alamat wajib diisi", domain.ErrInvalidInput)
	}
	// Koordinat wajib: seluruh aplikasi berdiri di atas urutan jarak (§13.3),
	// dan venue tanpa titik tidak bisa diurutkan sama sekali.
	if v.Lat < -90 || v.Lat > 90 || v.Lng < -180 || v.Lng > 180 {
		return fmt.Errorf("%w: koordinat di luar rentang", domain.ErrInvalidInput)
	}
	if v.Lat == 0 && v.Lng == 0 {
		return fmt.Errorf("%w: koordinat venue wajib diisi", domain.ErrInvalidInput)
	}
	switch v.Status {
	case "", domain.VenueUnclaimed, domain.VenueClaimed, domain.VenueVerified, domain.VenueSuspended:
	default:
		return fmt.Errorf("%w: status venue %q tidak dikenal", domain.ErrInvalidInput, v.Status)
	}
	return nil
}

// Slugify membuat slug URL dari nama venue.
//
// Hanya huruf, angka, dan tanda hubung yang dipertahankan; "Sekawan Kopi & Space"
// menjadi "sekawan-kopi-space" seperti contoh di §8.2.
func Slugify(s string) string {
	var b strings.Builder
	lastDash := true // cegah tanda hubung di awal
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
