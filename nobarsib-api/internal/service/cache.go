package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

// TTL cache daftar venue (§13.6).
//
// Menjelang laga, informasi berubah paling cepat: venue mengonfirmasi lewat
// link H-1 (§15.3) dan badge "Dikonfirmasi" harus segera terlihat. Cache lima
// menit di jam-jam itu berarti pengguna melihat status yang sudah basi, persis
// hal yang paling merusak kepercayaan menurut §21.
const (
	CacheTTLNormal  = 5 * time.Minute
	CacheTTLNearKO  = 1 * time.Minute
	NearKickoffFrom = 24 * time.Hour // mulai H-1
)

type NobarCache struct {
	rdb *redis.Client
	log *slog.Logger
}

func NewNobarCache(rdb *redis.Client, log *slog.Logger) *NobarCache {
	return &NobarCache{rdb: rdb, log: log}
}

// key membentuk kunci dari seluruh parameter yang memengaruhi hasil.
//
// Koordinat dibulatkan ke 3 desimal (±100 meter) supaya pengguna di satu
// kawasan berbagi entri cache yang sama. Tanpa pembulatan, setiap pengguna
// punya kunci sendiri dan cache tidak pernah kena.
func (c *NobarCache) key(p domain.NobarSearchParams) string {
	raw := fmt.Sprintf("%d|%.3f|%.3f|%s|%.1f|%s|%s|%t|%d",
		p.MatchID, p.Lat, p.Lng, p.Sort, p.RadiusKm,
		strings.Join(p.Facilities, ","), p.EntryType, p.OpenUntilEnd, p.Limit)
	sum := sha256.Sum256([]byte(raw))
	return "nobar:" + hex.EncodeToString(sum[:16])
}

func (c *NobarCache) Get(ctx context.Context, p domain.NobarSearchParams) (map[string]any, bool) {
	if c.rdb == nil {
		return nil, false
	}
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	raw, err := c.rdb.Get(ctx, c.key(p)).Bytes()
	if err != nil {
		return nil, false // termasuk redis.Nil; cache meleset bukan error
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, false
	}
	return body, true
}

// Set menyimpan hasil dengan TTL yang menyesuaikan kedekatan kickoff.
func (c *NobarCache) Set(ctx context.Context, p domain.NobarSearchParams, kickoff time.Time, body any) {
	if c.rdb == nil {
		return
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return
	}

	ttl := CacheTTLNormal
	if time.Until(kickoff) < NearKickoffFrom {
		ttl = CacheTTLNearKO
	}

	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	key := c.key(p)
	setKey := matchSetKey(p.MatchID)

	// Kunci dicatat di set penunjuk per laga supaya InvalidateMatch bisa
	// membuangnya tanpa memindai seluruh Redis. Set-nya diberi TTL lebih
	// panjang dari entri cache agar tidak menumpuk kalau tidak pernah dibatalkan.
	pipe := c.rdb.TxPipeline()
	pipe.Set(ctx, key, raw, ttl)
	pipe.SAdd(ctx, setKey, key)
	pipe.Expire(ctx, setKey, ttl+time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		c.log.Warn("gagal menyimpan cache", slog.String("error", err.Error()))
	}
}

func matchSetKey(matchID int64) string {
	return fmt.Sprintf("nobar:match:%d", matchID)
}

// InvalidateMatch membuang seluruh cache untuk satu laga.
//
// Dipanggil setelah status event berubah — terutama konfirmasi H-1 — supaya
// badge "Dikonfirmasi" muncul tanpa menunggu TTL habis.
//
// Kunci cache di-hash sehingga tidak bisa dipilih dengan pola. Karena itu
// setiap kunci dicatat di set penunjuk per laga (lihat Set), dan pembatalannya
// cukup membaca set itu. Alternatifnya KEYS/SCAN atas seluruh Redis — KEYS
// memblokir server, dan keduanya terasa persis di malam pertandingan saat
// trafik memuncak (§17.5).
func (c *NobarCache) InvalidateMatch(ctx context.Context, matchID int64) {
	if c.rdb == nil {
		return
	}
	setKey := matchSetKey(matchID)
	keys, err := c.rdb.SMembers(ctx, setKey).Result()
	if err != nil {
		c.log.Warn("gagal membaca penunjuk cache", slog.String("error", err.Error()))
		return
	}
	if len(keys) > 0 {
		c.rdb.Del(ctx, keys...)
	}
	c.rdb.Del(ctx, setKey)
}
