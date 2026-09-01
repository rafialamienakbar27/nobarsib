// Package scoring menghitung skor rekomendasi venue (blueprint §9.2).
//
// Rumusnya sengaja tetap sederhana supaya bisa dijelaskan ke pemilik venue
// yang bertanya "kenapa tempat saya di bawah". Dua hal berbeda dari blueprint,
// keduanya disengaja — lihat BoostPromoted dan SQLScoreExpr di bawah.
package scoring

import "math"

// Bobot rumus §9.2. Kelimanya berjumlah 1,0.
const (
	WeightDistance     = 0.35
	WeightRating       = 0.25
	WeightKondusif     = 0.20
	WeightCompleteness = 0.10
	WeightConfirmed    = 0.10
)

// BoostPromoted adalah PERBAIKAN #4 (rencana-pengerjaan §Fase 2).
//
// Blueprint §9.2 memberi promoted 0,15 sementara bonus konfirmasi hanya 0,10,
// sehingga venue berbayar yang BELUM konfirmasi bisa mengalahkan venue yang
// SUDAH konfirmasi. Itu bertentangan dengan blueprint sendiri: §4.3 menyebut
// konfirmasi H-1 "mekanisme paling penting di seluruh sistem" dan §21 menyebut
// data basi sebagai risiko fatal yang membunuh FANZO.
//
// Karena itu dua hal diubah:
//  1. nilainya turun ke 0,05 — §9.2 sendiri sudah memperingatkan
//     "jangan lebih, nanti hasilnya sampah";
//  2. boost hanya berlaku kalau venue sudah konfirmasi (lihat CalculateScore).
//
// Venue bayar tetap diuntungkan, tapi harus tetap konfirmasi untuk
// mendapatkannya — insentifnya searah dengan mekanisme yang paling dijaga.
const BoostPromoted = 0.05

// MinReviewsForKondusif mengikuti §11.4: skor kondusif tidak ditampilkan
// sebelum ada 3 review. Ambang yang sama dipakai saat memberi peringkat —
// kalau angkanya belum layak ditampilkan, ia juga belum layak menggeser urutan.
const MinReviewsForKondusif = 3

// ratingBlendAt adalah jumlah review saat rating nobar sepenuhnya menggantikan
// rating Google (§9.2: w = min(1, nobar_rating_count / 10)).
const ratingBlendAt = 10.0

// Input adalah nilai yang dibutuhkan rumus untuk satu venue.
type Input struct {
	DistanceKm       float64
	RadiusKm         float64
	NobarRating      float64 // 0 kalau belum ada
	NobarRatingCount int
	GoogleRating     float64 // 0 kalau tidak diketahui
	KondusifScore    float64 // 0 kalau belum ada
	DataCompleteness float64 // 0..1
	IsConfirmed      bool
	IsPromoted       bool
}

// CalculateScore mengembalikan skor akhir. Nilainya 0..1 ditambah boost.
func CalculateScore(in Input) float64 {
	score := WeightDistance*DistanceScore(in.DistanceKm, in.RadiusKm) +
		WeightRating*RatingScore(in.NobarRating, in.NobarRatingCount, in.GoogleRating) +
		WeightKondusif*KondusifScore(in.KondusifScore, in.NobarRatingCount) +
		WeightCompleteness*clamp01(in.DataCompleteness)

	if in.IsConfirmed {
		score += WeightConfirmed
		// Boost berbayar menempel pada konfirmasi, bukan berdiri sendiri.
		if in.IsPromoted {
			score += BoostPromoted
		}
	}
	return score
}

// DistanceScore bernilai 1 di titik pengguna dan turun linear sampai 0 di
// batas radius.
func DistanceScore(distanceKm, radiusKm float64) float64 {
	if radiusKm <= 0 {
		return 0
	}
	return math.Max(0, 1-(distanceKm/radiusKm))
}

// RatingScore mencampur rating nobar dan rating Google.
//
// Venue baru dinilai dengan rating Google; setelah 10 review nobar, rating
// Google tidak lagi berpengaruh. Rating Google mengukur kualitas kopi dan
// keramahan pelayan (§11.1) — berguna sebagai tebakan awal, bukan sebagai
// jawaban akhir untuk pertanyaan "enak tidak nonton di sini".
func RatingScore(nobarRating float64, nobarCount int, googleRating float64) float64 {
	w := math.Min(1, float64(nobarCount)/ratingBlendAt)
	return clamp01((nobarRating*w + googleRating*(1-w)) / 5.0)
}

// KondusifScore mengembalikan 0 selama review belum mencapai ambang §11.4.
func KondusifScore(kondusif float64, reviewCount int) float64 {
	if reviewCount < MinReviewsForKondusif {
		return 0
	}
	return clamp01(kondusif / 5.0)
}

func clamp01(v float64) float64 {
	return math.Min(1, math.Max(0, v))
}

// SQLScoreExpr adalah PERBAIKAN #3 (rencana-pengerjaan §Fase 2).
//
// Blueprint §9.1 mengurutkan `ORDER BY distance_km LIMIT/OFFSET`, lalu
// menjalankan CalculateScore di Go terhadap hasilnya. Akibatnya yang diurutkan
// berdasarkan skor hanya 20 venue terdekat, bukan seluruh kandidat dalam radius:
// venue ber-rating tinggi dan sudah dikonfirmasi di km ke-8 tidak akan pernah
// muncul kalau ada 20 venue lain yang lebih dekat — padahal mode sortnya
// bernama "recommended" dan jarak hanya berbobot 0,35.
//
// Karena itu rumus yang sama ditulis ulang sebagai ekspresi SQL supaya
// LIMIT/OFFSET bekerja pada urutan yang benar. CalculateScore tetap
// dipertahankan sebagai rujukan yang bisa diuji, dan sebuah test memastikan
// kedua jalur menghasilkan angka yang sama.
//
// Placeholder yang dipakai:
//
//	$1 = titik lokasi pengguna (geography)
//	$2 = radius dalam meter
//
// Kolom yang dirujuk berasal dari alias `v` (venue) dan `ne` (nobar_event).
const SQLScoreExpr = `
    ` + sqlWeightDistance + ` * GREATEST(0, 1 - (ST_Distance(v.location, $1) / $2))
  + ` + sqlWeightRating + ` * LEAST(1, GREATEST(0, (
        COALESCE(v.nobar_rating, 0) * LEAST(1, COALESCE(v.nobar_rating_count, 0) / 10.0)
      + COALESCE(v.google_rating, 0) * (1 - LEAST(1, COALESCE(v.nobar_rating_count, 0) / 10.0))
    ) / 5.0))
  + ` + sqlWeightKondusif + ` * CASE
        WHEN COALESCE(v.nobar_rating_count, 0) >= ` + sqlMinReviews + `
        THEN LEAST(1, GREATEST(0, COALESCE(v.kondusif_score, 0) / 5.0))
        ELSE 0
    END
  + ` + sqlWeightCompleteness + ` * LEAST(1, GREATEST(0, COALESCE(v.data_completeness, 0)))
  + CASE
        WHEN ne.confirmed_at IS NOT NULL
        THEN ` + sqlWeightConfirmed + ` + CASE WHEN ne.is_promoted THEN ` + sqlBoostPromoted + ` ELSE 0 END
        ELSE 0
    END`

// Konstanta di atas ditulis sebagai literal SQL agar rumus Go dan rumus SQL
// tidak bisa berpisah diam-diam; TestSQLDanGoSepakat menjaga keduanya sama.
const (
	sqlWeightDistance     = "0.35"
	sqlWeightRating       = "0.25"
	sqlWeightKondusif     = "0.20"
	sqlWeightCompleteness = "0.10"
	sqlWeightConfirmed    = "0.10"
	sqlBoostPromoted      = "0.05"
	sqlMinReviews         = "3"
)
