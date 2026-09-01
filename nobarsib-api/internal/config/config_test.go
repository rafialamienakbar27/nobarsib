package config

import (
	"strings"
	"testing"
)

// Test ini menjaga satu baris checklist rilis §23: "environment variable
// produksi tidak ada nilai default development". Nilai contoh di blueprint
// §17.2 adalah "dev-secret-ganti-di-production" — persis bentuk yang paling
// mungkin ikut terbawa ke server.
func TestLoadTolakRahasiaDevelopmentDiProduction(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          EnvProduction,
		"DATABASE_URL":     "postgres://u:p@db:5432/nobarsib",
		"REDIS_URL":        "redis://redis:6379/0",
		"JWT_SECRET":       "dev-secret-ganti-di-production",
		"DEVICE_HASH_SALT": "8Wq2mZx7Lp0RtYv3NbGh5Kc1Ee9Dd4Ff",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() menerima JWT_SECRET development di production, seharusnya ditolak")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("pesan error tidak menyebut JWT_SECRET: %v", err)
	}
}

func TestLoadTolakSslmodeDisableDiProduction(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          EnvProduction,
		"DATABASE_URL":     "postgres://u:p@db:5432/nobarsib?sslmode=disable",
		"REDIS_URL":        "redis://redis:6379/0",
		"JWT_SECRET":       "8Wq2mZx7Lp0RtYv3NbGh5Kc1Ee9Dd4Ff",
		"DEVICE_HASH_SALT": "3Aa6Bb9Cc2Dd5Ee8Ff1Gg4Hh7Ii0Jj3K",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() menerima sslmode=disable di production, seharusnya ditolak")
	}
	if !strings.Contains(err.Error(), "sslmode=disable") {
		t.Fatalf("pesan error tidak menyebut sslmode: %v", err)
	}
}

// Aturan panjang dan kata terlarang hanya berlaku di production; memaksakannya
// di development hanya menyulitkan tanpa menambah keamanan.
func TestLoadTerimaRahasiaPendekDiDevelopment(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          EnvDevelopment,
		"DATABASE_URL":     "postgres://u:p@localhost:5432/nobarsib?sslmode=disable",
		"REDIS_URL":        "redis://localhost:6379/0",
		"JWT_SECRET":       "dev-secret",
		"DEVICE_HASH_SALT": "dev-salt",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() gagal di development: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port default = %q, mau %q", cfg.Port, "8080")
	}
	if cfg.DatabaseMaxConn != 10 {
		t.Errorf("DatabaseMaxConn default = %d, mau 10", cfg.DatabaseMaxConn)
	}
}

// Semua masalah dilaporkan sekaligus supaya tidak perlu start ulang berkali-kali
// hanya untuk menemukan variabel berikutnya yang kurang.
func TestLoadLaporkanSemuaMasalahSekaligus(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV": EnvDevelopment,
	})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() berhasil tanpa DATABASE_URL, REDIS_URL, dan rahasia")
	}
	for _, want := range []string{"DATABASE_URL", "REDIS_URL", "JWT_SECRET", "DEVICE_HASH_SALT"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("pesan error tidak menyebut %s: %v", want, err)
		}
	}
}

func TestLoadTolakAppEnvTidakDikenal(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          "prod",
		"DATABASE_URL":     "postgres://u:p@localhost:5432/nobarsib",
		"REDIS_URL":        "redis://localhost:6379/0",
		"JWT_SECRET":       "dev-secret",
		"DEVICE_HASH_SALT": "dev-salt",
	})

	if _, err := Load(); err == nil {
		t.Fatal(`Load() menerima APP_ENV="prod", seharusnya hanya production yang sah`)
	}
}

// setEnv mengosongkan seluruh variabel yang dibaca Load, lalu memasang nilai
// yang diminta. t.Setenv mengembalikan nilai semula secara otomatis setelah test.
func setEnv(t *testing.T, values map[string]string) {
	t.Helper()
	all := []string{
		"APP_ENV", "PORT", "DATABASE_URL", "DATABASE_MAX_CONN",
		"REDIS_URL", "JWT_SECRET", "DEVICE_HASH_SALT", "SHUTDOWN_TIMEOUT_SEC",
	}
	for _, k := range all {
		t.Setenv(k, "")
	}
	for k, v := range values {
		t.Setenv(k, v)
	}
}

// ---------------------------------------------------------------------- CORS

// Tanpa CORS, web di :3000 tidak bisa memanggil API di :8080 dari sisi browser:
// responsnya sampai tapi dibuang, dan tab urutan, chip filter, serta tombol
// lokasi mati tanpa pesan galat. Test ini menjaga penguraian daftarnya.
func TestParseOriginsBersihkanSpasiDanGarisMiring(t *testing.T) {
	got := parseOrigins(" http://localhost:3000/ , https://nobarsib.id ,, ")
	want := []string{"http://localhost:3000", "https://nobarsib.id"}

	if len(got) != len(want) {
		t.Fatalf("parseOrigins() = %v, mau %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseOrigins()[%d] = %q, mau %q", i, got[i], want[i])
		}
	}
}

func TestParseOriginsKosongJadiNil(t *testing.T) {
	if got := parseOrigins("   "); len(got) != 0 {
		t.Fatalf("parseOrigins(spasi) = %v, mau kosong", got)
	}
}

// "*" berarti situs mana pun boleh memanggil API — termasuk /v1/auth dan
// /v1/admin. Ditolak di konfigurasi supaya tidak pernah sampai ke server.
func TestLoadTolakCORSBintang(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          EnvDevelopment,
		"DATABASE_URL":     "postgres://u:p@db:5432/nobarsib",
		"REDIS_URL":        "redis://redis:6379/0",
		"JWT_SECRET":       "rahasia-dev",
		"DEVICE_HASH_SALT": "garam-dev",
		"CORS_ORIGINS":     "http://localhost:3000,*",
	})

	_, err := Load()
	if err == nil {
		t.Fatal(`Load() menerima CORS_ORIGINS "*", seharusnya ditolak`)
	}
	if !strings.Contains(err.Error(), "CORS_ORIGINS") {
		t.Fatalf("pesan error tidak menyebut CORS_ORIGINS: %v", err)
	}
}

// Origin tanpa skema tidak akan pernah cocok dengan header Origin browser,
// jadi lebih baik gagal saat boot daripada diam-diam tidak berfungsi.
func TestLoadTolakOriginTanpaSkema(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          EnvDevelopment,
		"DATABASE_URL":     "postgres://u:p@db:5432/nobarsib",
		"REDIS_URL":        "redis://redis:6379/0",
		"JWT_SECRET":       "rahasia-dev",
		"DEVICE_HASH_SALT": "garam-dev",
		"CORS_ORIGINS":     "nobarsib.id",
	})

	if _, err := Load(); err == nil {
		t.Fatal("Load() menerima origin tanpa skema, seharusnya ditolak")
	}
}

func TestLoadTerimaDaftarOriginYangSah(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          EnvDevelopment,
		"DATABASE_URL":     "postgres://u:p@db:5432/nobarsib",
		"REDIS_URL":        "redis://redis:6379/0",
		"JWT_SECRET":       "rahasia-dev",
		"DEVICE_HASH_SALT": "garam-dev",
		"CORS_ORIGINS":     "http://localhost:3000,https://nobarsib.id",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() = %v, mau nil", err)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Fatalf("CORSOrigins = %v, mau 2 entri", cfg.CORSOrigins)
	}
}
