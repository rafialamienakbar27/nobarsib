package openhours

import "testing"

func TestParseClock(t *testing.T) {
	cases := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"00:00", 0, false},
		{"08:30", 510, false},
		{"23:00", 1380, false},
		// Postgres menolak '24:00'::time, dan justru nilai inilah yang dipakai
		// blueprint §7.2 sebagai penanda buka 24 jam.
		{"24:00", 1440, false},
		{"02:00", 120, false},
		{"24:01", 0, true},
		{"25:00", 0, true},
		{"08:60", 0, true},
		{"2300", 0, true},
		{"", 0, true},
	}
	for _, c := range cases {
		got, err := ParseClock(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseClock(%q) = %d, mau error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseClock(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseClock(%q) = %d, mau %d", c.in, got, c.want)
		}
	}
}

// Inti PERBAIKAN #2: jam tutup yang melewati tengah malam harus jadi >1440,
// bukan berputar balik ke angka kecil.
func TestNormalizeMelewatiTengahMalam(t *testing.T) {
	w := Week{
		"6": {Open: "10:00", Close: "23:00"}, // tutup hari yang sama
		"5": {Open: "17:00", Close: "02:00"}, // Barrack, Grow
		"4": {Open: "10:00", Close: "04:00"}, // Jabarano
		"3": {Open: "08:00", Close: "24:00"}, // buka sampai tengah malam
		"2": nil,                             // tutup
	}
	if err := w.Normalize(); err != nil {
		t.Fatalf("Normalize() error: %v", err)
	}

	want := map[string]int{"6": 1380, "5": 1560, "4": 1680, "3": 1440}
	for dow, exp := range want {
		if got := w[dow].CloseMinutes; got != exp {
			t.Errorf("hari %s CloseMinutes = %d, mau %d", dow, got, exp)
		}
	}
	if w["2"] != nil {
		t.Errorf("hari 2 seharusnya tetap nil (tutup)")
	}
}

// Kasus yang membuat filter blueprint membuang venue terbaiknya.
// Kickoff 19:00 (menit 1140) + 2 jam = menit 1260.
func TestOpenUntilKickoffMalam(t *testing.T) {
	const required = 19*60 + 120 // 1260

	cases := []struct {
		nama      string
		day       *Day
		wantOpen  bool
		wantKnown bool
	}{
		// Justru inilah yang terbuang oleh `close::time > (kickoff+2h)::time`:
		// perbandingannya menjadi `02:00 > 21:00` = false.
		{"Barrack Billiard, tutup 02:00", &Day{Open: "17:00", Close: "02:00"}, true, true},
		{"Jabarano, tutup 04:00", &Day{Open: "10:00", Close: "04:00"}, true, true},
		{"Bober, buka 24 jam", &Day{Open: "00:00", Close: "24:00"}, true, true},
		{"Sekawan, tutup 24:00", &Day{Open: "10:00", Close: "24:00"}, true, true},
		// Yang memang harus terbuang.
		{"Kedai PSM 46, tutup 23:00", &Day{Open: "10:00", Close: "23:00"}, true, true},
		{"cafe tutup 21:00", &Day{Open: "08:00", Close: "21:00"}, false, true},
		{"cafe tutup 20:00", &Day{Open: "08:00", Close: "20:00"}, false, true},
		{"tutup di hari itu", nil, false, true},
	}

	for _, c := range cases {
		w := Week{"6": c.day}
		if err := w.Normalize(); err != nil {
			t.Fatalf("%s: Normalize() error: %v", c.nama, err)
		}
		open, known := w.OpenUntil(6, required)
		if open != c.wantOpen || known != c.wantKnown {
			t.Errorf("%s: OpenUntil = (%v, %v), mau (%v, %v)",
				c.nama, open, known, c.wantOpen, c.wantKnown)
		}
	}
}

// Hari yang tidak tercatat berbeda dari hari yang tercatat tutup. Venue baru
// dengan data belum lengkap tidak boleh disembunyikan oleh filter — peringkatnya
// sudah dihukum lewat data_completeness (§9.3).
func TestOpenUntilHariTidakTercatat(t *testing.T) {
	w := Week{"6": {Open: "10:00", Close: "23:00"}}
	if err := w.Normalize(); err != nil {
		t.Fatalf("Normalize() error: %v", err)
	}

	if _, known := w.OpenUntil(3, 1260); known {
		t.Error("hari yang tidak tercatat seharusnya dilaporkan sebagai tidak diketahui")
	}
	if _, known := w.OpenUntil(6, 1260); !known {
		t.Error("hari yang tercatat seharusnya dilaporkan sebagai diketahui")
	}
}

func TestNormalizeTolakKunciHariTidakValid(t *testing.T) {
	for _, dow := range []string{"7", "-1", "senin", ""} {
		w := Week{dow: {Open: "08:00", Close: "22:00"}}
		if err := w.Normalize(); err == nil {
			t.Errorf("Normalize() menerima kunci hari %q", dow)
		}
	}
}

func TestUnmarshalWeek(t *testing.T) {
	raw := []byte(`{"0":{"open":"08:00","close":"23:00"},"1":null}`)
	w, err := UnmarshalWeek(raw)
	if err != nil {
		t.Fatalf("UnmarshalWeek() error: %v", err)
	}
	if err := w.Normalize(); err != nil {
		t.Fatalf("Normalize() error: %v", err)
	}
	if w["0"].CloseMinutes != 1380 {
		t.Errorf("hari 0 CloseMinutes = %d, mau 1380", w["0"].CloseMinutes)
	}
	if day, ok := w["1"]; !ok || day != nil {
		t.Errorf("hari 1 seharusnya tercatat dan nil, dapat ok=%v day=%v", ok, day)
	}

	if w, err := UnmarshalWeek(nil); err != nil || w != nil {
		t.Errorf("UnmarshalWeek(nil) = (%v, %v), mau (nil, nil)", w, err)
	}
}
