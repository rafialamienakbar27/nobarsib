package handler

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

type Public struct {
	nobar   *service.NobarService
	venues  *service.VenueService
	matches matchLister
	views   domain.EventViewRepository
	cache   *service.NobarCache
	log     *slog.Logger
}

type matchLister interface {
	Upcoming(ctx context.Context, p domain.UpcomingParams) ([]domain.Match, error)
	Season(ctx context.Context, teamSlug string, from, to time.Time) ([]domain.Match, error)
}

func NewPublic(
	nobar *service.NobarService,
	venues *service.VenueService,
	matches matchLister,
	views domain.EventViewRepository,
	cache *service.NobarCache,
	log *slog.Logger,
) *Public {
	return &Public{nobar: nobar, venues: venues, matches: matches, views: views, cache: cache, log: log}
}

// UpcomingMatches menangani GET /v1/matches/upcoming (§8.2).
func (h *Public) UpcomingMatches(c *fiber.Ctx) error {
	limit := queryInt(c, "limit", service.DefaultUpcoming)
	if limit <= 0 || limit > 50 {
		limit = service.DefaultUpcoming
	}

	matches, err := h.matches.Upcoming(c.UserContext(), domain.UpcomingParams{
		TeamSlug: c.Query("team_slug", service.DefaultTeamSlug),
		Limit:    limit,
	})
	if err != nil {
		return fail(c, err)
	}

	out := make([]matchDTO, 0, len(matches))
	for _, m := range matches {
		out = append(out, newMatchDTO(m))
	}
	return c.JSON(fiber.Map{"data": out})
}

// Season menangani GET /v1/matches — halaman /jadwal (§13.1).
func (h *Public) Season(c *fiber.Ctx) error {
	now := time.Now().In(service.WIB)
	from, to := now.AddDate(-1, 0, 0), now.AddDate(1, 0, 0)

	matches, err := h.matches.Season(c.UserContext(), c.Query("team_slug", service.DefaultTeamSlug), from, to)
	if err != nil {
		return fail(c, err)
	}

	out := make([]matchDTO, 0, len(matches))
	for _, m := range matches {
		out = append(out, newMatchDTO(m))
	}
	return c.JSON(fiber.Map{"data": out})
}

// NobarForMatch menangani GET /v1/matches/{match_id}/nobar — endpoint utama
// aplikasi (§8.2).
func (h *Public) NobarForMatch(c *fiber.Ctx) error {
	matchID, err := strconv.ParseInt(c.Params("match_id"), 10, 64)
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "match_id harus berupa angka")
	}

	params := domain.NobarSearchParams{
		MatchID:      matchID,
		Lat:          queryFloat(c, "lat", 0),
		Lng:          queryFloat(c, "lng", 0),
		Sort:         c.Query("sort"),
		RadiusKm:     queryFloat(c, "radius_km", 0),
		Facilities:   splitCSV(c.Query("facilities")),
		EntryType:    c.Query("entry_type"),
		OpenUntilEnd: queryBool(c, "open_until_end", true),
		Limit:        queryInt(c, "per_page", service.DefaultPerPage),
	}
	page := queryInt(c, "page", 1)
	if page < 1 {
		page = 1
	}

	params, err = service.NormalizeSearchParams(params)
	if err != nil {
		return fail(c, err)
	}
	params.Offset = (page - 1) * params.Limit

	// Cache dilewati untuk halaman selain pertama: hampir semua pengguna hanya
	// melihat halaman pertama, dan menyimpan setiap kombinasi halaman hanya
	// memenuhi Redis tanpa menambah kecepatan yang terasa.
	useCache := params.Offset == 0
	if useCache {
		if cached, ok := h.cache.Get(c.UserContext(), params); ok {
			return c.JSON(cached)
		}
	}

	result, err := h.nobar.Search(c.UserContext(), params)
	if err != nil {
		return fail(c, err)
	}

	data := make([]eventDTO, 0, len(result.Events))
	for _, e := range result.Events {
		data = append(data, newEventDTO(e))
	}

	body := fiber.Map{
		"data": data,
		"meta": fiber.Map{
			"total":    result.Total,
			"page":     page,
			"per_page": params.Limit,
			"match": fiber.Map{
				"id":         result.Match.ID,
				"kickoff_at": wib(result.Match.KickoffAt),
				"label":      result.Match.Label(),
			},
		},
	}
	if useCache {
		h.cache.Set(c.UserContext(), params, result.Match.KickoffAt, body)
	}
	return c.JSON(body)
}

// VenueDetail menangani GET /v1/venues/{slug} (§8.2).
func (h *Public) VenueDetail(c *fiber.Ctx) error {
	detail, err := h.venues.Detail(c.UserContext(), c.Params("slug"))
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(fiber.Map{"data": newVenueDetailDTO(detail)})
}

// SearchVenues menangani GET /v1/venues/search (§8.2).
func (h *Public) SearchVenues(c *fiber.Ctx) error {
	limit := queryInt(c, "per_page", service.DefaultVenueTop)
	page := queryInt(c, "page", 1)
	if page < 1 {
		page = 1
	}

	venues, total, err := h.venues.Search(c.UserContext(), domain.VenueSearchParams{
		Query:  c.Query("q"),
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return fail(c, err)
	}

	out := make([]venueCardDTO, 0, len(venues))
	for i := range venues {
		out = append(out, newVenueCardDTO(&venues[i], 0))
	}
	return c.JSON(fiber.Map{
		"data": out,
		"meta": meta{Total: total, Page: page, PerPage: limit},
	})
}

// Track menangani POST /v1/events/{id}/track (§8.2).
//
// Fire-and-forget: response 202 dikirim tanpa menunggu penulisan selesai.
// Statistik tidak boleh memperlambat pengguna yang sedang memilih tempat, dan
// kegagalan mencatatnya tidak boleh terlihat sebagai error.
func (h *Public) Track(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "id event tidak valid")
	}

	var body struct {
		Action     string `json:"action"`
		DeviceHash string `json:"device_hash"`
	}
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput, "body tidak bisa dibaca")
	}
	if !domain.IsValidAction(body.Action) {
		return sendError(c, fiber.StatusBadRequest, CodeInvalidInput,
			"action harus salah satu dari: "+strings.Join(domain.ValidActions, ", "))
	}

	view := domain.EventView{NobarEventID: id, DeviceHash: body.DeviceHash, Action: body.Action}
	go func() {
		// Context sendiri, karena context permintaan sudah dibatalkan begitu
		// response terkirim.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.views.Track(ctx, view); err != nil {
			h.log.Warn("gagal mencatat interaksi",
				slog.String("event_id", id.String()),
				slog.String("action", view.Action),
				slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusAccepted)
}

// Facilities menangani GET /v1/facilities — daftar chip filter di §13.2.
func (h *Public) Facilities(c *fiber.Ctx) error {
	list, err := h.venues.Facilities(c.UserContext())
	if err != nil {
		return fail(c, err)
	}
	type dto struct {
		Code     string `json:"code"`
		Label    string `json:"label"`
		Category string `json:"category,omitempty"`
	}
	out := make([]dto, 0, len(list))
	for _, f := range list {
		out = append(out, dto{Code: f.Code, Label: f.Label, Category: f.Category})
	}
	return c.JSON(fiber.Map{"data": out})
}

func queryInt(c *fiber.Ctx, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryFloat(c *fiber.Ctx, key string, def float64) float64 {
	v := c.Query(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func queryBool(c *fiber.Ctx, key string, def bool) bool {
	v := c.Query(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

// splitCSV membaca parameter seperti "area_anak,parkir_mobil" (§8.2).
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
