// Command importer memasukkan data venue dari berkas JSON.
//
// Panel admin bisa memasukkan venue satu per satu, tapi untuk 20–30 venue hasil
// survei lapangan (§16), mengetik sekali di satu berkas jauh lebih cepat
// daripada mengisi tiga puluh formulir — dan berkasnya bisa diperiksa ulang,
// disimpan, serta dijalankan lagi setelah dikoreksi.
//
// Idempoten berdasarkan slug: menjalankan ulang berkas yang sama akan
// MEMPERBARUI venue yang sudah ada, bukan membuat duplikat. Jadi alurnya
// menjadi: ketik, impor, lihat hasilnya, koreksi berkas, impor lagi.
//
//	make import FILE=nobarsib-api/testdata/venues.contoh.json
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/nobarsib/nobarsib-api/internal/config"
	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/pkg/openhours"
	"github.com/nobarsib/nobarsib-api/internal/repository"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

type venueInput struct {
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
	Facilities      []string                  `json:"facilities"`
	Photos          []photoInput              `json:"photos"`
	Status          string                    `json:"status"`
	Catatan         string                    `json:"_catatan"` // untuk dirimu sendiri, diabaikan
}

type photoInput struct {
	URL       string `json:"url"`
	Caption   string `json:"caption"`
	IsPrimary bool   `json:"is_primary"`
}

func main() {
	_ = godotenv.Load()

	berkas := flag.String("file", "", "berkas JSON berisi daftar venue")
	kering := flag.Bool("dry-run", false, "periksa berkas tanpa menulis ke database")
	flag.Parse()

	if *berkas == "" {
		fmt.Fprintln(os.Stderr, "pakai: importer -file venues.json [-dry-run]")
		os.Exit(2)
	}

	if err := jalan(*berkas, *kering); err != nil {
		fmt.Fprintln(os.Stderr, "gagal:", err)
		os.Exit(1)
	}
}

func jalan(berkas string, kering bool) error {
	raw, err := os.ReadFile(berkas)
	if err != nil {
		return fmt.Errorf("baca berkas: %w", err)
	}

	var daftar []venueInput
	dec := json.NewDecoder(bytes.NewReader(raw))
	// Field yang salah ketik akan ditolak, bukan diabaikan diam-diam. Salah
	// ketik "instagram_handel" yang lolos berarti data itu hilang tanpa jejak.
	dec.DisallowUnknownFields()
	if err := dec.Decode(&daftar); err != nil {
		return fmt.Errorf("berkas bukan JSON yang sah: %w", err)
	}

	// Seluruh berkas divalidasi lebih dulu, sebelum satu baris pun ditulis.
	// Impor yang berhenti di tengah meninggalkan keadaan setengah jadi yang
	// harus dibereskan manual.
	var masalah []string
	for i, v := range daftar {
		for _, m := range periksa(v) {
			masalah = append(masalah, fmt.Sprintf("venue #%d (%s): %s", i+1, v.Name, m))
		}
	}
	if len(masalah) > 0 {
		for _, m := range masalah {
			fmt.Fprintln(os.Stderr, "  ✗", m)
		}
		return fmt.Errorf("%d masalah ditemukan, tidak ada yang ditulis", len(masalah))
	}
	fmt.Printf("%d venue lolos pemeriksaan\n", len(daftar))

	if kering {
		fmt.Println("dry-run: tidak ada yang ditulis ke database")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL, 4)
	if err != nil {
		return err
	}
	defer db.Close()

	venues := repository.NewVenueRepo(db)
	svc := service.NewVenueService(venues, repository.NewReviewRepo(db), repository.NewNobarEventRepo(db))

	var dibuat, diperbarui int
	for _, in := range daftar {
		v := in.toDomain()
		lama, err := venues.GetBySlug(ctx, v.Slug)

		switch {
		case err == nil:
			v.ID = lama.ID
			if err := svc.Update(ctx, v, in.Facilities); err != nil {
				return fmt.Errorf("perbarui %s: %w", v.Slug, err)
			}
			diperbarui++
		case errors.Is(err, domain.ErrNotFound):
			if err := svc.Create(ctx, v, in.Facilities); err != nil {
				return fmt.Errorf("buat %s: %w", v.Slug, err)
			}
			dibuat++
		default:
			return fmt.Errorf("cek %s: %w", v.Slug, err)
		}

		// Foto dicocokkan berdasarkan URL, bukan sekadar ditambahkan.
		//
		// Tanpa ini, menjalankan ulang berkas yang sama akan menumpuk foto yang
		// persis sama berkali-kali — dan seluruh gagasan importer ini adalah
		// "ketik, impor, koreksi, impor lagi".
		sudahAda := map[string]bool{}
		if kini, err := venues.GetBySlug(ctx, v.Slug); err == nil {
			for _, f := range kini.Photos {
				sudahAda[f.URL] = true
			}
		}
		for _, p := range in.Photos {
			if sudahAda[p.URL] {
				continue
			}
			foto := &domain.VenuePhoto{URL: p.URL, Caption: p.Caption, IsPrimary: p.IsPrimary}
			if err := svc.AddPhoto(ctx, v.ID, foto); err != nil {
				return fmt.Errorf("foto %s: %w", v.Slug, err)
			}
		}

		skor, err := venues.RecalculateCompleteness(ctx, v.ID)
		if err != nil {
			return err
		}
		tanda := "  "
		// §20.1 menargetkan >60% venue punya data_completeness di atas 0,8.
		if skor < 0.8 {
			tanda = "! "
		}
		fmt.Printf("%s%-34s kelengkapan %.2f\n", tanda, v.Slug, skor)
	}

	fmt.Printf("\nselesai: %d dibuat, %d diperbarui\n", dibuat, diperbarui)
	fmt.Println("venue bertanda ! belum mencapai kelengkapan 0,80 (§20.1)")
	return nil
}

// periksa mengumpulkan seluruh masalah satu venue sekaligus, bukan berhenti di
// yang pertama — supaya sekali jalan cukup untuk memperbaiki semuanya.
func periksa(v venueInput) []string {
	var m []string

	if v.Name == "" {
		m = append(m, "name wajib diisi")
	}
	if v.Address == "" {
		m = append(m, "address wajib diisi")
	}
	if v.Lat == 0 || v.Lng == 0 {
		m = append(m, "lat/lng wajib diisi (ambil dari Google Maps: klik kanan titiknya)")
	}
	if v.Lat < -90 || v.Lat > 90 || v.Lng < -180 || v.Lng > 180 {
		m = append(m, "lat/lng di luar rentang yang masuk akal")
	}
	// Bandung ada di sekitar -6,8..-7,0 dan 107,5..107,8. Lat dan lng yang
	// tertukar adalah kesalahan paling sering saat menyalin dari Maps, dan
	// hasilnya venue muncul di tengah Samudra Hindia tanpa ada yang sadar.
	if v.Lat > 0 || v.Lng < 100 {
		m = append(m, "lat/lng sepertinya tertukar (lat harus negatif, lng sekitar 107)")
	}

	if v.OpeningHours != nil {
		w := openhours.Week(v.OpeningHours)
		if err := w.Normalize(); err != nil {
			m = append(m, "opening_hours: "+err.Error())
		}
		if len(v.OpeningHours) != 7 {
			m = append(m, fmt.Sprintf(
				"opening_hours baru %d dari 7 hari — hari yang tutup pun harus ditulis (\"1\": null)",
				len(v.OpeningHours)))
		}
	}

	primer := 0
	for _, p := range v.Photos {
		if p.URL == "" {
			m = append(m, "ada foto tanpa url")
		}
		if p.IsPrimary {
			primer++
		}
	}
	if primer > 1 {
		m = append(m, "hanya boleh satu foto is_primary")
	}
	return m
}

func (in venueInput) toDomain() *domain.Venue {
	v := &domain.Venue{
		Name: in.Name, Slug: in.Slug, Address: in.Address, District: in.District,
		City: in.City, Lat: in.Lat, Lng: in.Lng, Phone: in.Phone,
		WhatsApp: in.WhatsApp, InstagramHandle: in.InstagramHandle,
		GooglePlaceID: in.GooglePlaceID, GoogleRating: in.GoogleRating,
		GoogleRatingCount: in.GoogleCount, Status: in.Status, IsActive: true,
	}
	if v.Slug == "" {
		v.Slug = service.Slugify(v.Name)
	}
	if v.City == "" {
		v.City = "Kota Bandung"
	}
	if in.OpeningHours != nil {
		v.OpeningHours = openhours.Week(in.OpeningHours)
	}
	return v
}
