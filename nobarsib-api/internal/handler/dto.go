package handler

import (
	"time"

	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

// Bentuk response mengikuti contoh §8.2 persis, termasuk nama field snake_case
// dan timestamp ISO 8601 berzona +07:00 (§8.1).

type teamDTO struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	LogoURL string `json:"logo_url,omitempty"`
}

type matchDTO struct {
	ID          int64   `json:"id"`
	Competition string  `json:"competition,omitempty"`
	HomeTeam    teamDTO `json:"home_team"`
	AwayTeam    teamDTO `json:"away_team"`
	KickoffAt   string  `json:"kickoff_at"`
	VenueName   string  `json:"venue_name,omitempty"`
	BroadcastTV string  `json:"broadcast_tv,omitempty"`
	Status      string  `json:"status"`
	NobarCount  int     `json:"nobar_count"`
}

func newMatchDTO(m domain.Match) matchDTO {
	return matchDTO{
		ID:          m.ID,
		Competition: m.Competition.Name,
		HomeTeam:    teamDTO{Name: m.HomeTeam.Name, Slug: m.HomeTeam.Slug, LogoURL: m.HomeTeam.LogoURL},
		AwayTeam:    teamDTO{Name: m.AwayTeam.Name, Slug: m.AwayTeam.Slug, LogoURL: m.AwayTeam.LogoURL},
		KickoffAt:   wib(m.KickoffAt),
		VenueName:   m.VenueName,
		BroadcastTV: m.BroadcastTV,
		Status:      m.Status,
		NobarCount:  m.NobarCount,
	}
}

type venueCardDTO struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	District         string    `json:"district,omitempty"`
	DistanceKm       float64   `json:"distance_km"`
	PrimaryPhoto     string    `json:"primary_photo,omitempty"`
	GoogleRating     *float64  `json:"google_rating"`
	NobarRating      *float64  `json:"nobar_rating"`
	NobarRatingCount int       `json:"nobar_rating_count"`
	KondusifScore    *float64  `json:"kondusif_score"`
	KidFriendlyScore *float64  `json:"kid_friendly_score"`
	Facilities       []string  `json:"facilities"`
}

func newVenueCardDTO(v *domain.Venue, distanceKm float64) venueCardDTO {
	facilities := v.Facilities
	if facilities == nil {
		facilities = []string{}
	}
	return venueCardDTO{
		ID:           v.ID,
		Name:         v.Name,
		Slug:         v.Slug,
		District:     v.District,
		DistanceKm:   round1(distanceKm),
		PrimaryPhoto: v.PrimaryPhoto,
		GoogleRating: v.GoogleRating,
		NobarRating:  v.NobarRating,

		NobarRatingCount: v.NobarRatingCount,
		// §11.4: skor kondusif dan ramah anak tidak ditampilkan sebelum ada
		// 3 review. Frontend menampilkan "Belum ada penilaian" saat null.
		KondusifScore:    v.VisibleKondusif(),
		KidFriendlyScore: v.VisibleKidFriendly(),
		Facilities:       facilities,
	}
}

type eventDTO struct {
	EventID     uuid.UUID    `json:"event_id"`
	Venue       venueCardDTO `json:"venue"`
	DoorsOpenAt string       `json:"doors_open_at,omitempty"`
	EntryType   string       `json:"entry_type"`
	EntryAmount int          `json:"entry_amount"`
	CrowdLevel  string       `json:"crowd_level,omitempty"`
	IsConfirmed bool         `json:"is_confirmed"`
	ConfirmedAt string       `json:"confirmed_at,omitempty"`
	IsPromoted  bool         `json:"is_promoted"`
	Notes       string       `json:"notes,omitempty"`
}

func newEventDTO(e domain.NobarEvent) eventDTO {
	d := eventDTO{
		EventID:     e.ID,
		EntryType:   e.EntryType,
		EntryAmount: e.EntryAmount,
		CrowdLevel:  e.CrowdLevel,
		IsConfirmed: e.IsConfirmed(),
		IsPromoted:  e.IsPromoted,
		Notes:       e.Notes,
	}
	if e.Venue != nil {
		d.Venue = newVenueCardDTO(e.Venue, e.DistanceKm)
	}
	if e.DoorsOpenAt != nil {
		d.DoorsOpenAt = wib(*e.DoorsOpenAt)
	}
	if e.ConfirmedAt != nil {
		d.ConfirmedAt = wib(*e.ConfirmedAt)
	}
	return d
}

type reviewDTO struct {
	RatingOverall  int    `json:"rating_overall"`
	RatingKondusif *int   `json:"rating_kondusif"`
	IsKidFriendly  *bool  `json:"is_kid_friendly"`
	CrowdActual    string `json:"crowd_actual,omitempty"`
	Comment        string `json:"comment,omitempty"`
	CreatedAt      string `json:"created_at"`
}

type historyDTO struct {
	MatchLabel  string `json:"match_label"`
	KickoffAt   string `json:"kickoff_at"`
	EntryType   string `json:"entry_type"`
	EntryAmount int    `json:"entry_amount"`
	IsConfirmed bool   `json:"is_confirmed"`
}

type venueDetailDTO struct {
	ID               uuid.UUID             `json:"id"`
	Name             string                `json:"name"`
	Slug             string                `json:"slug"`
	Address          string                `json:"address"`
	District         string                `json:"district,omitempty"`
	City             string                `json:"city"`
	Lat              float64               `json:"lat"`
	Lng              float64               `json:"lng"`
	Phone            string                `json:"phone,omitempty"`
	WhatsApp         string                `json:"whatsapp,omitempty"`
	InstagramHandle  string                `json:"instagram_handle,omitempty"`
	Website          string                `json:"website,omitempty"`
	GoogleRating     *float64              `json:"google_rating"`
	NobarRating      *float64              `json:"nobar_rating"`
	NobarRatingCount int                   `json:"nobar_rating_count"`
	KondusifScore    *float64              `json:"kondusif_score"`
	KidFriendlyScore *float64              `json:"kid_friendly_score"`
	DataCompleteness float64               `json:"data_completeness"`
	Status           string                `json:"status"`
	// Asal-usul data profil. Keduanya omitempty: venue yang tidak diketahui
	// asalnya lebih baik tidak menyebut apa-apa daripada mengirim string kosong
	// yang di sisi klien gampang terbaca sebagai "sudah diverifikasi".
	DataSource     string `json:"data_source,omitempty"`
	LastVerifiedAt string `json:"last_verified_at,omitempty"`
	Facilities       []string              `json:"facilities"`
	Photos           []photoDTO            `json:"photos"`
	OpeningHours     map[string]openDayDTO `json:"opening_hours"`
	Reviews          []reviewDTO           `json:"recent_reviews"`
	History          []historyDTO          `json:"nobar_history"`
}

type photoDTO struct {
	URL       string `json:"url"`
	Caption   string `json:"caption,omitempty"`
	IsPrimary bool   `json:"is_primary"`
}

type openDayDTO struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

func newVenueDetailDTO(d *service.VenueDetail) venueDetailDTO {
	v := d.Venue

	facilities := v.Facilities
	if facilities == nil {
		facilities = []string{}
	}

	photos := make([]photoDTO, 0, len(v.Photos))
	for _, p := range v.Photos {
		photos = append(photos, photoDTO{URL: p.URL, Caption: p.Caption, IsPrimary: p.IsPrimary})
	}

	// close_minutes tidak ikut dikirim: itu bentuk internal untuk memfilter
	// (lihat internal/pkg/openhours), bukan sesuatu yang perlu dibaca frontend.
	hours := make(map[string]openDayDTO, len(v.OpeningHours))
	for dow, day := range v.OpeningHours {
		if day == nil {
			continue // hari tutup: sengaja tidak muncul di response
		}
		hours[dow] = openDayDTO{Open: day.Open, Close: day.Close}
	}

	reviews := make([]reviewDTO, 0, len(d.Reviews))
	for _, r := range d.Reviews {
		reviews = append(reviews, reviewDTO{
			RatingOverall:  r.RatingOverall,
			RatingKondusif: r.RatingKondusif,
			IsKidFriendly:  r.IsKidFriendly,
			CrowdActual:    r.CrowdActual,
			Comment:        r.Comment,
			CreatedAt:      wib(r.CreatedAt),
		})
	}

	history := make([]historyDTO, 0, len(d.History))
	for _, h := range d.History {
		item := historyDTO{
			EntryType:   h.EntryType,
			EntryAmount: h.EntryAmount,
			IsConfirmed: h.IsConfirmed(),
		}
		if h.Match != nil {
			item.MatchLabel = h.Match.Label()
			item.KickoffAt = wib(h.Match.KickoffAt)
		}
		history = append(history, item)
	}

	// Tanggal saja, tanpa jam dan tanpa zona waktu — sama seperti kolomnya di
	// basis data. Mengirimnya sebagai RFC3339 akan menambahkan "T00:00:00Z"
	// yang bukan cuma berisik, tapi keliru: itu tengah malam UTC, yaitu pukul
	// tujuh pagi WIB di hari yang sama, dan pembaca yang teliti akan menyimpulkan
	// jam verifikasi yang tidak pernah kami catat.
	var diverifikasi string
	if v.LastVerifiedAt != nil {
		diverifikasi = v.LastVerifiedAt.Format("2006-01-02")
	}

	return venueDetailDTO{
		ID: v.ID, Name: v.Name, Slug: v.Slug, Address: v.Address,
		District: v.District, City: v.City, Lat: v.Lat, Lng: v.Lng,
		Phone: v.Phone, WhatsApp: v.WhatsApp, InstagramHandle: v.InstagramHandle,
		Website:      v.Website,
		GoogleRating: v.GoogleRating, NobarRating: v.NobarRating,
		NobarRatingCount: v.NobarRatingCount,
		KondusifScore:    v.VisibleKondusif(),
		KidFriendlyScore: v.VisibleKidFriendly(),
		DataCompleteness: v.DataCompleteness,
		Status:           v.Status,
		DataSource:       v.DataSource,
		LastVerifiedAt:   diverifikasi,
		Facilities:       facilities,
		Photos:           photos,
		OpeningHours:     hours,
		Reviews:          reviews,
		History:          history,
	}
}

// wib memformat waktu sebagai ISO 8601 berzona +07:00 (§8.1).
func wib(t time.Time) string {
	return t.In(service.WIB).Format(time.RFC3339)
}

// round1 membulatkan jarak ke satu desimal: "2,4 km" seperti di §13.2.
// Ketelitian lebih dari itu tidak berarti apa-apa bagi orang yang sedang
// memilih tempat nonton.
func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
