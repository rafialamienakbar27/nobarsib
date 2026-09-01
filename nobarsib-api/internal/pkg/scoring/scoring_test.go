package scoring

import (
	"math"
	"strconv"
	"testing"
)

const eps = 1e-9

// Inti PERBAIKAN #4: venue berbayar yang belum konfirmasi tidak boleh
// mengalahkan venue yang sudah konfirmasi.
//
// Dengan bobot blueprint (promoted 0,15 > konfirmasi 0,10) test ini gagal,
// dan itulah bug yang diperbaiki.
func TestPromotedTidakMengalahkanKonfirmasi(t *testing.T) {
	base := Input{DistanceKm: 3, RadiusKm: 15, GoogleRating: 4.5, DataCompleteness: 0.8}

	promotedBelumKonfirmasi := base
	promotedBelumKonfirmasi.IsPromoted = true

	sudahKonfirmasi := base
	sudahKonfirmasi.IsConfirmed = true

	a := CalculateScore(promotedBelumKonfirmasi)
	b := CalculateScore(sudahKonfirmasi)
	if a >= b {
		t.Errorf("venue promoted belum konfirmasi (%.4f) >= venue terkonfirmasi (%.4f); "+
			"konfirmasi H-1 harus lebih berat daripada bayar", a, b)
	}
}

func TestBoostPromotedButuhKonfirmasi(t *testing.T) {
	base := Input{DistanceKm: 3, RadiusKm: 15, GoogleRating: 4.5}

	belum := base
	belum.IsPromoted = true
	if got, want := CalculateScore(belum), CalculateScore(base); math.Abs(got-want) > eps {
		t.Errorf("promoted tanpa konfirmasi mengubah skor: %.4f vs %.4f", got, want)
	}

	konfirmasi := base
	konfirmasi.IsConfirmed = true
	konfirmasiPromoted := konfirmasi
	konfirmasiPromoted.IsPromoted = true

	diff := CalculateScore(konfirmasiPromoted) - CalculateScore(konfirmasi)
	if math.Abs(diff-BoostPromoted) > eps {
		t.Errorf("boost promoted = %.4f, mau %.4f", diff, BoostPromoted)
	}
}

func TestDistanceScore(t *testing.T) {
	cases := []struct{ dist, radius, want float64 }{
		{0, 15, 1},
		{15, 15, 0},
		{7.5, 15, 0.5},
		{20, 15, 0}, // di luar radius tidak boleh negatif
		{5, 0, 0},   // radius nol tidak boleh membagi nol
	}
	for _, c := range cases {
		if got := DistanceScore(c.dist, c.radius); math.Abs(got-c.want) > eps {
			t.Errorf("DistanceScore(%.1f, %.1f) = %.4f, mau %.4f", c.dist, c.radius, got, c.want)
		}
	}
}

// Venue baru dinilai dengan rating Google; setelah 10 review nobar, rating
// Google tidak lagi berpengaruh (§9.2).
func TestRatingScoreBeralihDariGoogleKeNobar(t *testing.T) {
	cases := []struct {
		nobar float64
		count int
		goog  float64
		want  float64
	}{
		{0, 0, 4.5, 0.9},  // belum ada review nobar -> murni Google
		{5, 10, 3.0, 1.0}, // 10 review -> murni nobar
		{5, 20, 3.0, 1.0}, // lebih dari 10 tidak melebihi 1
		{4, 5, 2.0, 0.6},  // setengah-setengah: (4*0.5 + 2*0.5)/5
		{0, 0, 0, 0},      // venue tanpa rating apa pun
	}
	for _, c := range cases {
		got := RatingScore(c.nobar, c.count, c.goog)
		if math.Abs(got-c.want) > eps {
			t.Errorf("RatingScore(%.1f, %d, %.1f) = %.4f, mau %.4f",
				c.nobar, c.count, c.goog, got, c.want)
		}
	}
}

// §11.4: skor kondusif tidak dipakai sebelum ada 3 review.
func TestKondusifScoreButuhTigaReview(t *testing.T) {
	for count := 0; count < MinReviewsForKondusif; count++ {
		if got := KondusifScore(5, count); got != 0 {
			t.Errorf("KondusifScore(5, %d) = %.4f, mau 0", count, got)
		}
	}
	if got := KondusifScore(5, 3); math.Abs(got-1) > eps {
		t.Errorf("KondusifScore(5, 3) = %.4f, mau 1", got)
	}
}

func TestCalculateScoreRentangWajar(t *testing.T) {
	max := CalculateScore(Input{
		DistanceKm: 0, RadiusKm: 15,
		NobarRating: 5, NobarRatingCount: 20, GoogleRating: 5,
		KondusifScore: 5, DataCompleteness: 1,
		IsConfirmed: true, IsPromoted: true,
	})
	want := WeightDistance + WeightRating + WeightKondusif +
		WeightCompleteness + WeightConfirmed + BoostPromoted
	if math.Abs(max-want) > eps {
		t.Errorf("skor maksimum = %.4f, mau %.4f", max, want)
	}

	min := CalculateScore(Input{DistanceKm: 99, RadiusKm: 15})
	if min != 0 {
		t.Errorf("skor minimum = %.4f, mau 0", min)
	}
}

// Bobot yang dipakai rumus Go dan yang ditanam di SQLScoreExpr harus sama.
// Tanpa test ini, mengubah salah satu saja akan membuat urutan hasil query
// berbeda dari skor yang dihitung ulang di Go — dan itu jenis bug yang tidak
// menimbulkan error, hanya urutan yang diam-diam salah.
func TestSQLDanGoSepakat(t *testing.T) {
	pairs := []struct {
		nama string
		sql  string
		go_  float64
	}{
		{"WeightDistance", sqlWeightDistance, WeightDistance},
		{"WeightRating", sqlWeightRating, WeightRating},
		{"WeightKondusif", sqlWeightKondusif, WeightKondusif},
		{"WeightCompleteness", sqlWeightCompleteness, WeightCompleteness},
		{"WeightConfirmed", sqlWeightConfirmed, WeightConfirmed},
		{"BoostPromoted", sqlBoostPromoted, BoostPromoted},
	}
	for _, p := range pairs {
		v, err := strconv.ParseFloat(p.sql, 64)
		if err != nil {
			t.Errorf("%s: literal SQL %q bukan angka: %v", p.nama, p.sql, err)
			continue
		}
		if math.Abs(v-p.go_) > eps {
			t.Errorf("%s: SQL %s != Go %v", p.nama, p.sql, p.go_)
		}
	}

	n, err := strconv.Atoi(sqlMinReviews)
	if err != nil || n != MinReviewsForKondusif {
		t.Errorf("sqlMinReviews = %q, mau %d", sqlMinReviews, MinReviewsForKondusif)
	}
}

func TestBobotBerjumlahSatu(t *testing.T) {
	total := WeightDistance + WeightRating + WeightKondusif +
		WeightCompleteness + WeightConfirmed
	if math.Abs(total-1.0) > eps {
		t.Errorf("jumlah bobot = %.4f, mau 1.0", total)
	}
}
