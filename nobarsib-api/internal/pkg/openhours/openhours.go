// Package openhours menangani jam operasional venue.
//
// Paket ini ada khusus untuk PERBAIKAN #2 (rencana-pengerjaan §Fase 2).
//
// Blueprint §9.4 memfilter jam tutup dengan `close::time > (kickoff + 2 jam)::time`.
// Untuk venue yang tutup lewat tengah malam perbandingan itu menjadi
// `02:00 > 21:00` = false, sehingga venue justru terbuang — padahal Jabarano
// (tutup 04:00), Rooftop Coffee (03:00), Ludo (03:00), Barrack (02:00) dan
// Grow (02:00) adalah kandidat terbaik di §16.2. Ditambah `'24:00'::time`
// melempar error di Postgres, dan cabang OR di blueprint tidak menyelamatkannya
// karena cast tetap dievaluasi.
//
// Solusinya: jam tutup dinormalisasi menjadi MENIT SEJAK TENGAH MALAM HARI BUKA,
// dan boleh melebihi 1440 kalau tutupnya keesokan hari.
//
//	tutup 23:00 -> 1380
//	tutup 24:00 -> 1440
//	tutup 02:00 -> 1560
//
// Perbandingan lalu dilakukan dalam bilangan bulat, bukan tipe `time`.
package openhours

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// MinutesPerDay dipakai sebagai ambang "tutupnya keesokan hari".
const MinutesPerDay = 24 * 60

// Day adalah jam operasional satu hari. Nil berarti venue tutup di hari itu.
type Day struct {
	Open  string `json:"open"`  // "08:00"
	Close string `json:"close"` // "02:00" (boleh lewat tengah malam), "24:00"

	// CloseMinutes adalah bentuk ternormalisasi dari Close, satu-satunya field
	// yang dipakai saat memfilter. Diisi otomatis oleh Normalize.
	CloseMinutes int `json:"close_minutes"`
}

// Week memetakan hari (0 = Minggu ... 6 = Sabtu) ke jam operasionalnya.
// Kunci yang tidak ada berarti "tidak diketahui"; kunci dengan nilai nil
// berarti "tutup di hari itu". Dua hal itu sengaja dibedakan: venue baru yang
// datanya belum lengkap tidak boleh diperlakukan sama dengan venue yang memang
// tutup (§13.5 — data tipis adalah kondisi normal di awal).
type Week map[string]*Day

// ParseClock membaca "HH:MM" menjadi menit sejak tengah malam.
//
// "24:00" diterima dan bernilai 1440. Postgres menolak nilai itu sebagai `time`,
// dan justru "24:00" yang dipakai blueprint §7.2 sebagai penanda buka 24 jam.
func ParseClock(s string) (int, error) {
	s = strings.TrimSpace(s)
	h, m, ok := strings.Cut(s, ":")
	if !ok {
		return 0, fmt.Errorf("format jam %q tidak dikenal, harus HH:MM", s)
	}
	hh, err := strconv.Atoi(h)
	if err != nil {
		return 0, fmt.Errorf("jam %q tidak valid: %w", s, err)
	}
	mm, err := strconv.Atoi(m)
	if err != nil {
		return 0, fmt.Errorf("menit %q tidak valid: %w", s, err)
	}
	if hh < 0 || hh > 24 || mm < 0 || mm > 59 {
		return 0, fmt.Errorf("jam %q di luar rentang 00:00–24:00", s)
	}
	if hh == 24 && mm != 0 {
		return 0, fmt.Errorf("jam %q di luar rentang 00:00–24:00", s)
	}
	return hh*60 + mm, nil
}

// Normalize mengisi CloseMinutes untuk seluruh hari.
//
// Jam tutup yang lebih kecil atau sama dengan jam buka dianggap keesokan hari
// dan ditambah 1440: venue buka 17:00 tutup 02:00 menghasilkan 1560, bukan 120.
// Inilah satu-satunya tempat aturan itu ditulis; sisanya cukup membaca
// CloseMinutes.
func (w Week) Normalize() error {
	for dow, day := range w {
		if err := validDow(dow); err != nil {
			return err
		}
		if day == nil {
			continue
		}

		openMin, err := ParseClock(day.Open)
		if err != nil {
			return fmt.Errorf("hari %s, jam buka: %w", dow, err)
		}
		closeMin, err := ParseClock(day.Close)
		if err != nil {
			return fmt.Errorf("hari %s, jam tutup: %w", dow, err)
		}
		if closeMin <= openMin {
			closeMin += MinutesPerDay
		}
		day.CloseMinutes = closeMin
	}
	return nil
}

// OpenUntil menjawab apakah venue masih buka melewati menit ke-required pada
// hari dow, dengan required dihitung dari tengah malam hari yang sama.
//
// Perbandingannya tegas (>), mengikuti §9.4. Venue yang tutup persis di menit
// itu berarti mengusir penonton tepat saat peluit akhir dan masa bubar habis —
// secara teknis "masih buka", tapi bukan tempat yang layak direkomendasikan.
//
// Mengembalikan (hasil, diketahui). Ketika diketahui == false, pemanggil tidak
// boleh menyimpulkan apa pun: hari itu tidak tercatat, bukan berarti tutup.
func (w Week) OpenUntil(dow int, required int) (open bool, known bool) {
	day, ok := w[strconv.Itoa(dow)]
	if !ok {
		return false, false // hari tidak tercatat — tidak diketahui
	}
	if day == nil {
		return false, true // tercatat tutup
	}
	if day.CloseMinutes == 0 {
		return false, false // belum dinormalisasi — jangan menebak
	}
	return day.CloseMinutes > required, true
}

func validDow(dow string) error {
	n, err := strconv.Atoi(dow)
	if err != nil || n < 0 || n > 6 {
		return fmt.Errorf("kunci hari %q tidak valid, harus \"0\"–\"6\"", dow)
	}
	return nil
}

// UnmarshalWeek membaca JSONB dari kolom venue.opening_hours.
func UnmarshalWeek(raw []byte) (Week, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var w Week
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("baca opening_hours: %w", err)
	}
	return w, nil
}
