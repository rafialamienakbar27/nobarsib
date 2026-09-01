// Package config memuat konfigurasi aplikasi dari environment variable.
//
// Checklist rilis §23 mensyaratkan "environment variable produksi tidak ada
// nilai default development". Karena itu semua rahasia wajib diisi eksplisit
// dan aplikasi menolak start kalau tidak — gagal saat boot jauh lebih murah
// daripada gagal saat malam pertandingan.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// minSecretLen adalah panjang minimum JWT_SECRET. 32 byte setara 256 bit,
// sepadan dengan HMAC-SHA256 yang dipakai untuk menandatangani token.
const minSecretLen = 32

type Config struct {
	AppEnv string
	Port   string

	DatabaseURL     string
	DatabaseMaxConn int32

	RedisURL string

	JWTSecret string

	// DeviceHashSalt dipakai untuk device_hash review (§11.5).
	// Kalau salt bocor, fingerprint browser bisa dihitung ulang oleh pihak lain,
	// jadi diperlakukan sama seperti JWT secret.
	DeviceHashSalt string

	// CORSOrigins adalah daftar origin yang boleh memanggil API dari browser.
	//
	// Dibutuhkan karena web dan API berjalan di port berbeda saat pengembangan
	// (3000 dan 8080), jadi setiap permintaan dari sisi klien — ganti urutan,
	// pasang filter, aktifkan lokasi — adalah permintaan lintas origin. Tanpa
	// ini browser memblokir responsnya dan seluruh kendali di halaman mati
	// tanpa pesan galat yang jelas.
	//
	// Di produksi web dan API biasanya berbagi domain lewat reverse proxy
	// sehingga daftar ini boleh kosong; yang tidak pernah boleh adalah "*",
	// dan itu ditolak di bawah.
	CORSOrigins []string

	ShutdownTimeout time.Duration
}

func (c Config) IsProduction() bool { return c.AppEnv == EnvProduction }

// Load membaca konfigurasi dan memvalidasinya. Semua masalah dikumpulkan lalu
// dilaporkan sekaligus, supaya tidak perlu start ulang berkali-kali hanya untuk
// menemukan variabel berikutnya yang kurang.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:          envOr("APP_ENV", EnvDevelopment),
		Port:            envOr("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		DatabaseMaxConn: int32(envIntOr("DATABASE_MAX_CONN", 10)),
		RedisURL:        os.Getenv("REDIS_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		DeviceHashSalt:  os.Getenv("DEVICE_HASH_SALT"),
		ShutdownTimeout: time.Duration(envIntOr("SHUTDOWN_TIMEOUT_SEC", 15)) * time.Second,
	}
	cfg.CORSOrigins = parseOrigins(os.Getenv("CORS_ORIGINS"))

	var problems []string

	switch cfg.AppEnv {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		problems = append(problems, fmt.Sprintf(
			"APP_ENV %q tidak dikenal (pilih: %s, %s, %s)",
			cfg.AppEnv, EnvDevelopment, EnvStaging, EnvProduction))
	}

	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL wajib diisi")
	}
	if cfg.RedisURL == "" {
		problems = append(problems, "REDIS_URL wajib diisi")
	}

	problems = append(problems, checkSecret("JWT_SECRET", cfg.JWTSecret, cfg.IsProduction())...)
	problems = append(problems, checkSecret("DEVICE_HASH_SALT", cfg.DeviceHashSalt, cfg.IsProduction())...)

	if cfg.IsProduction() && strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
		problems = append(problems, "DATABASE_URL memakai sslmode=disable di production")
	}

	problems = append(problems, checkOrigins(cfg.CORSOrigins)...)

	if len(problems) > 0 {
		return Config{}, fmt.Errorf("konfigurasi tidak valid:\n  - %s",
			strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}

// checkSecret menolak rahasia yang kosong, terlalu pendek, atau masih memakai
// nilai contoh. docker-compose.yml di blueprint §17.2 memakai literal
// "dev-secret-ganti-di-production" — persis nilai yang paling mungkin ikut
// terbawa ke server kalau tidak dicegat di sini.
func checkSecret(name, value string, isProd bool) []string {
	if value == "" {
		return []string{name + " wajib diisi"}
	}
	if !isProd {
		return nil
	}

	var problems []string
	if len(value) < minSecretLen {
		problems = append(problems, fmt.Sprintf(
			"%s minimal %d karakter di production (sekarang %d)",
			name, minSecretLen, len(value)))
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{"dev", "test", "ganti", "change", "example", "secret"} {
		if strings.Contains(lower, marker) {
			problems = append(problems, fmt.Sprintf(
				"%s masih mengandung %q — terlihat seperti nilai development",
				name, marker))
			break
		}
	}
	return problems
}

// parseOrigins memecah CORS_ORIGINS yang dipisah koma, membuang spasi dan
// garis miring di ujung — "http://localhost:3000/" dan "http://localhost:3000"
// adalah origin yang sama bagi browser, tapi tidak bagi pembanding string.
func parseOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if o != "" {
			out = append(out, o)
		}
	}
	return out
}

// checkOrigins menolak "*".
//
// Fiber memperlakukan "*" sebagai izin untuk semua situs. Untuk API yang
// sepenuhnya publik itu tidak berbahaya, tapi API ini juga melayani /v1/auth
// dan /v1/admin, dan membuka keduanya ke halaman mana pun adalah kesalahan
// yang tidak akan terlihat sampai ada yang memanfaatkannya.
func checkOrigins(origins []string) []string {
	var problems []string
	for _, o := range origins {
		if o == "*" {
			problems = append(problems, `CORS_ORIGINS tidak boleh "*" — tulis daftar origin yang diizinkan`)
			continue
		}
		if !strings.HasPrefix(o, "http://") && !strings.HasPrefix(o, "https://") {
			problems = append(problems, fmt.Sprintf(
				"CORS_ORIGINS %q harus lengkap dengan skema, contoh https://nobarsib.id", o))
		}
	}
	return problems
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
