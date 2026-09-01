package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/middleware"
	"github.com/nobarsib/nobarsib-api/internal/pkg/openhours"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

type Admin struct {
	nobar  *service.NobarService
	venues *service.VenueService
	events adminEventLister
	match  adminMatchRepo
}

type adminEventLister interface {
	ListForMatch(ctx context.Context, matchID int64, statuses []string) ([]domain.NobarEvent, error)
	ListPending(ctx context.Context, limit int) ([]domain.NobarEvent, error)
}

type adminMatchRepo interface {
	Create(ctx context.Context, m *domain.Match) error
	Update(ctx context.Context, m *domain.Match) error
	GetByID(ctx context.Context, id int64) (*domain.Match, error)
	Season(ctx context.Context, teamSlug string, from, to time.Time) ([]domain.Match, error)
	TeamBySlug(ctx context.Context, slug string) (*domain.Team, error)
}

func NewAdmin(
	nobar *service.NobarService,
	venues *service.VenueService,
	events adminEventLister,
	match adminMatchRepo,
) *Admin {
	return &Admin{nobar: nobar, venues: venues, events: events, match: match}
}

// ---------------------------------------------------------------- antrian

// PendingEvents menangani GET /v1/admin/events?status=pending_review (§8.4).
//
// Ini layar yang paling sering dibuka (§14.2), jadi seluruh yang dibutuhkan
// untuk memutuskan dikirim sekaligus — nama venue, laga, jam, biaya, sumber —
// supaya panel tidak perlu memanggil endpoint lain sebelum menampilkan antrian.
func (h *Admin) PendingEvents(c *fiber.Ctx) error {
	limit := queryInt(c, "limit", 50)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	events, err := h.events.ListPending(c.UserContext(), limit)
	if err != nil {
		return fail(c, err)
	}

	out := make([]adminEventDTO, 0, len(events))
	for _, e := range events {
		out = append(out, newAdminEventDTO(e))
	}
	return c.JSON(fiber.Map{"data": out})
}

// Approve menangani POST /v1/admin/events/{id}/approve (§8.4).
func (h *Admin) Approve(c *fiber.Ctx) error {
	return h.transition(c, domain.EventPublished)
}

// Reject menangani POST /v1/admin/events/{id}/reject (§8.4).
func (h *Admin) Reject(c *fiber.Ctx) error {
	return h.transition(c, domain.EventRejected)
}

// Confirm memindahkan event ke status confirmed.
//
// Di Fase 5 ini dipicu venue lewat tautan WA satu ketukan (§15.3); sekarang
// admin bisa memakainya saat menerima konfirmasi lewat telepon atau chat.
func (h *Admin) Confirm(c *fiber.Ctx) error {
	return h.transition(c, domain.EventConfirmed)
}

func (h *Admin) Cancel(c *fiber.Ctx) error {
	return h.transition(c, domain.EventCancelled)
}

func (h *Admin) transition(c *fiber.Ctx, ke string) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "id event tidak valid")
	}
	e, err := h.nobar.Transition(c.UserContext(), id, ke)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"id": e.ID, "status": e.Status}})
}

// CreateEvent dipakai admin saat memasukkan info nobar secara manual —
// jalur utama sebelum portal venue ada (§4.4).
func (h *Admin) CreateEvent(c *fiber.Ctx) error {
	var req struct {
		VenueID     string `json:"venue_id"`
		MatchID     int64  `json:"match_id"`
		DoorsOpenAt string `json:"doors_open_at"`
		EntryType   string `json:"entry_type"`
		EntryAmount int    `json:"entry_amount"`
		CrowdLevel  string `json:"crowd_level"`
		Notes       string `json:"notes"`
		Status      string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	venueID, err := uuid.Parse(req.VenueID)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "venue_id tidak valid")
	}

	e := &domain.NobarEvent{
		VenueID:     venueID,
		MatchID:     req.MatchID,
		EntryType:   req.EntryType,
		EntryAmount: req.EntryAmount,
		CrowdLevel:  req.CrowdLevel,
		Notes:       req.Notes,
		Status:      req.Status,
	}
	if req.DoorsOpenAt != "" {
		t, err := time.Parse(time.RFC3339, req.DoorsOpenAt)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, CodeInvalidInput,
				"doors_open_at harus format ISO 8601, contoh 2026-09-05T18:00:00+07:00")
		}
		e.DoorsOpenAt = &t
	}
	if uid, ok := middleware.UserIDFrom(c); ok {
		e.CreatedBy = &uid
	}

	if err := h.nobar.Create(c.UserContext(), e); err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{"id": e.ID, "status": e.Status},
	})
}

// ------------------------------------------------------------------ venue

type venueRequest struct {
	Name            string                    `json:"name"`
	Slug            string                    `json:"slug"`
	Address         string                    `json:"address"`
	District        string                    `json:"district"`
	City            string                    `json:"city"`
	Lat             float64                   `json:"lat"`
	Lng             float64                   `json:"lng"`
	Phone           string                    `json:"phone"`
	WhatsApp        string                    `json:"whatsapp"`
	InstagramHandle string                    `json:"instagram_handle"`
	GooglePlaceID   string                    `json:"google_place_id"`
	GoogleRating    *float64                  `json:"google_rating"`
	GoogleCount     *int                      `json:"google_rating_count"`
	OpeningHours    map[string]*openhours.Day `json:"opening_hours"`
	Status          string                    `json:"status"`
	IsActive        *bool                     `json:"is_active"`
	Facilities      []string                  `json:"facilities"`
}

func (r venueRequest) toDomain() *domain.Venue {
	v := &domain.Venue{
		Name: r.Name, Slug: r.Slug, Address: r.Address, District: r.District,
		City: r.City, Lat: r.Lat, Lng: r.Lng, Phone: r.Phone, WhatsApp: r.WhatsApp,
		InstagramHandle: r.InstagramHandle, GooglePlaceID: r.GooglePlaceID,
		GoogleRating: r.GoogleRating, GoogleRatingCount: r.GoogleCount,
		Status: r.Status, IsActive: true,
	}
	if r.IsActive != nil {
		v.IsActive = *r.IsActive
	}
	if r.OpeningHours != nil {
		v.OpeningHours = openhours.Week(r.OpeningHours)
	}
	return v
}

// CreateVenue menangani POST /v1/admin/venues (§8.4).
func (h *Admin) CreateVenue(c *fiber.Ctx) error {
	var req venueRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	v := req.toDomain()
	if err := h.venues.Create(c.UserContext(), v, req.Facilities); err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{"id": v.ID, "slug": v.Slug, "data_completeness": v.DataCompleteness},
	})
}

// UpdateVenue menangani PATCH /v1/admin/venues/{id}.
func (h *Admin) UpdateVenue(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "id venue tidak valid")
	}

	var req venueRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	v := req.toDomain()
	v.ID = id
	if err := h.venues.Update(c.UserContext(), v, req.Facilities); err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{"id": v.ID, "data_completeness": v.DataCompleteness},
	})
}

// AddPhoto menangani POST /v1/admin/venues/{id}/photos.
//
// Menerima URL, bukan berkas: penyimpanan objek (§6.1) diunggah terpisah, dan
// menjadikan API ini perantara unggahan hanya menambah beban memori pada VPS
// kecil tanpa manfaat yang sepadan.
func (h *Admin) AddPhoto(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "id venue tidak valid")
	}

	var req struct {
		URL       string `json:"url"`
		Caption   string `json:"caption"`
		IsPrimary bool   `json:"is_primary"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	p := &domain.VenuePhoto{
		URL: req.URL, Caption: req.Caption,
		IsPrimary: req.IsPrimary, SortOrder: req.SortOrder,
	}
	if err := h.venues.AddPhoto(c.UserContext(), id, p); err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{"id": p.ID}})
}

// ------------------------------------------------------------------ jadwal

// CreateMatch menangani POST /v1/admin/matches.
//
// Tim dirujuk lewat slug, bukan id angka: slug tidak berubah saat data
// di-seed ulang, dan jauh lebih sulit salah ketik tanpa disadari.
func (h *Admin) CreateMatch(c *fiber.Ctx) error {
	var req struct {
		CompetitionID int    `json:"competition_id"`
		HomeTeamSlug  string `json:"home_team_slug"`
		AwayTeamSlug  string `json:"away_team_slug"`
		KickoffAt     string `json:"kickoff_at"`
		VenueName     string `json:"venue_name"`
		BroadcastTV   string `json:"broadcast_tv"`
	}
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	home, err := h.match.TeamBySlug(c.UserContext(), req.HomeTeamSlug)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput,
			"Tim tuan rumah tidak dikenal: "+req.HomeTeamSlug)
	}
	away, err := h.match.TeamBySlug(c.UserContext(), req.AwayTeamSlug)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput,
			"Tim tamu tidak dikenal: "+req.AwayTeamSlug)
	}

	kickoff, err := time.Parse(time.RFC3339, req.KickoffAt)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput,
			"kickoff_at harus format ISO 8601, contoh 2026-09-05T19:00:00+07:00")
	}

	m := &domain.Match{
		Competition: domain.Competition{ID: req.CompetitionID},
		HomeTeam:    *home,
		AwayTeam:    *away,
		KickoffAt:   kickoff,
		VenueName:   req.VenueName,
		BroadcastTV: req.BroadcastTV,
		Status:      domain.MatchScheduled,
	}
	if err := h.match.Create(c.UserContext(), m); err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{"id": m.ID}})
}

// UpdateMatch menangani PATCH /v1/admin/matches/{id}.
func (h *Admin) UpdateMatch(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "id laga harus angka")
	}

	m, err := h.match.GetByID(c.UserContext(), id)
	if err != nil {
		return fail(c, err)
	}

	var req struct {
		KickoffAt   *string `json:"kickoff_at"`
		VenueName   *string `json:"venue_name"`
		BroadcastTV *string `json:"broadcast_tv"`
		Status      *string `json:"status"`
		ScoreHome   *int    `json:"score_home"`
		ScoreAway   *int    `json:"score_away"`
	}
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "Body tidak bisa dibaca")
	}

	// Hanya field yang dikirim yang diubah; yang tidak disebut dibiarkan apa
	// adanya. Panel admin sering hanya mengubah satu hal (jam tayang bergeser).
	if req.KickoffAt != nil {
		t, err := time.Parse(time.RFC3339, *req.KickoffAt)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, CodeInvalidInput,
				"kickoff_at harus format ISO 8601")
		}
		m.KickoffAt = t
	}
	if req.VenueName != nil {
		m.VenueName = *req.VenueName
	}
	if req.BroadcastTV != nil {
		m.BroadcastTV = *req.BroadcastTV
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	m.ScoreHome, m.ScoreAway = req.ScoreHome, req.ScoreAway

	if err := h.match.Update(c.UserContext(), m); err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"data": fiber.Map{"id": m.ID, "status": m.Status}})
}

// ---------------------------------------------------------------- ringkasan

// Dashboard menangani GET /v1/admin/summary — ringkasan §14.1.
func (h *Admin) Dashboard(c *fiber.Ctx) error {
	ctx := c.UserContext()

	pending, err := h.events.ListPending(ctx, 200)
	if err != nil {
		return fail(c, err)
	}

	sekarang := time.Now().In(service.WIB)
	laga, err := h.match.Season(ctx, service.DefaultTeamSlug, sekarang, sekarang.AddDate(0, 1, 0))
	if err != nil {
		return fail(c, err)
	}

	var berikutnya *matchDTO
	if len(laga) > 0 {
		d := newMatchDTO(laga[0])
		berikutnya = &d
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"pending_events":   len(pending),
			"upcoming_matches": len(laga),
			"next_match":       berikutnya,
		},
	})
}

type adminEventDTO struct {
	ID          string `json:"id"`
	VenueName   string `json:"venue_name"`
	VenueSlug   string `json:"venue_slug"`
	District    string `json:"district"`
	MatchLabel  string `json:"match_label"`
	KickoffAt   string `json:"kickoff_at"`
	DoorsOpenAt string `json:"doors_open_at,omitempty"`
	EntryType   string `json:"entry_type"`
	EntryAmount int    `json:"entry_amount"`
	Notes       string `json:"notes,omitempty"`
	Status      string `json:"status"`
	Source      string `json:"source,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func newAdminEventDTO(e domain.NobarEvent) adminEventDTO {
	d := adminEventDTO{
		ID: e.ID.String(), EntryType: e.EntryType, EntryAmount: e.EntryAmount,
		Notes: e.Notes, Status: e.Status, CreatedAt: wib(e.CreatedAt),
	}
	if e.Venue != nil {
		d.VenueName, d.VenueSlug, d.District = e.Venue.Name, e.Venue.Slug, e.Venue.District
	}
	if e.Match != nil {
		d.MatchLabel, d.KickoffAt = e.Match.Label(), wib(e.Match.KickoffAt)
	}
	if e.DoorsOpenAt != nil {
		d.DoorsOpenAt = wib(*e.DoorsOpenAt)
	}
	return d
}
