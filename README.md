# NOBARSIB

Direktori tempat nonton bareng Persib di Bandung.

| Dokumen | Isi |
|---|---|
| [blueprint-nobarsib.md](blueprint-nobarsib.md) | Desain lengkap — **apa** yang dibangun dan **kenapa** |
| [rencana-pengerjaan-nobarsib.md](rencana-pengerjaan-nobarsib.md) | Rencana kerja per fase — **kapan** dan **urutan apa** |
| [DEPLOY.md](DEPLOY.md) | Menaikkan situs ke Vercel dan API ke VPS |
| [nobarsib-android/README.md](nobarsib-android/README.md) | Aplikasi Android (APK) dan cara membangunnya ulang |

Status: **Fase 3 selesai** — API publik dan PWA jalan. Berikutnya Fase 4: panel admin dan pengisian data venue sungguhan.

## Endpoint

| Endpoint | Isi |
|---|---|
| `GET /v1/matches/upcoming` | Laga mendatang + jumlah venue nobar |
| `GET /v1/matches` | Seluruh jadwal, untuk halaman `/jadwal` |
| `GET /v1/matches/{id}/nobar` | **Endpoint utama** — venue penayang, terurut |
| `GET /v1/venues/{slug}` | Detail venue, fasilitas, jam buka, review |
| `GET /v1/venues/search?q=` | Pencarian nama, toleran salah ketik |
| `GET /v1/facilities` | Daftar fasilitas untuk chip filter |
| `POST /v1/events/{id}/track` | Catat interaksi, fire-and-forget |
| `GET /health`, `/live` | Monitoring |

Parameter endpoint utama: `lat`, `lng`, `sort` (`recommended`/`nearest`/`rating`),
`radius_km`, `facilities`, `entry_type`, `open_until_end`, `page`, `per_page`.

## Struktur

```
nobarsib/
├── nobarsib-api/          Backend Go + Fiber
│   ├── cmd/api            HTTP server
│   ├── cmd/worker         Tugas terjadwal (aktif di Fase 5)
│   ├── internal/
│   │   ├── domain/        Entity + kontrak repository, state machine §4.5
│   │   ├── repository/    Implementasi Postgres
│   │   ├── service/       Logika bisnis + cache
│   │   ├── handler/       HTTP handler & DTO
│   │   └── pkg/           openhours (jam tutup), scoring (skor rekomendasi)
│   ├── migrations/        7 migrasi golang-migrate
│   ├── seed/              Data referensi (competition, team)
│   └── testdata/          Fixture pengembangan
├── nobarsib-web/          Frontend Next.js 16 (App Router) + Tailwind 4
│   ├── app/               Halaman: /, /match/[id], /venue/[slug], /jadwal, …
│   ├── components/        Kartu venue, filter, galeri, tombol aksi
│   ├── lib/               Klien API, format Indonesia, lokasi, device hash
│   ├── proxy.ts           Header keamanan (bukan next.config — lihat catatan)
│   └── public/sw.js       Service worker PWA
├── docker-compose.yml     Postgres+PostGIS, Redis, api, worker, web
└── Makefile               Perintah pengembangan
```

## Mulai

Butuh Go 1.26+, Docker, dan runtime container yang menyala
(`colima start` kalau memakai colima).

```bash
cp nobarsib-api/.env.example nobarsib-api/.env
make infra          # nyalakan Postgres + Redis
make migrate-up     # buat skema
make seed           # isi competition & team
make fixture        # data uji pengembangan (opsional)
make run            # jalankan API di host  -> :8080
make web            # jalankan frontend     -> :3000
```

Atau seluruh stack sekaligus lewat Docker:

```bash
make up             # db, redis, api, worker, web
```

Cek:

```bash
curl -s localhost:8080/health
# {"status":"ok","dependencies":{"postgres":{"status":"ok"},"redis":{"status":"ok"}}}

curl -s "localhost:8080/v1/matches/upcoming"
```

`make` tanpa argumen menampilkan seluruh target yang tersedia.

## Catatan lingkungan

**Postgres container dipublikasikan di port host 5434**, bukan 5432. Mesin
pengembangan sering sudah menjalankan Postgres lain (Homebrew, Postgres.app),
dan karena instalasi itu mengikat `127.0.0.1` secara spesifik, ia selalu menang
atas binding `0.0.0.0` milik Docker saat host menghubungi `localhost`.

Gejalanya menyesatkan: `psql` dari host tampak berhasil tapi berisi database
proyek lain, atau menolak dengan `password authentication failed` padahal
`docker compose exec db psql` baik-baik saja. Kalau ragu, pastikan dengan:

```bash
psql "$DATABASE_URL" -tAc "select current_setting('data_directory')"
# /var/lib/postgresql/data  -> container (benar)
# /opt/homebrew/var/...     -> Postgres host (salah)
```

Ubah lewat `POSTGRES_HOST_PORT` kalau port 5432 di mesinmu memang bebas.

## Catatan frontend

**`output: "standalone"` sengaja tidak dipakai.** Server minimal yang
dihasilkannya tidak menerapkan `proxy.ts` maupun `headers()` dari
`next.config.ts` — sudah diverifikasi pada build yang sama: `next start`
mengirim header keamanan, `node server.js` standalone tidak mengirim apa pun,
dan gagalnya tanpa error sama sekali.

Konsekuensinya image web membengkak dari sekitar 150 MB ke 793 MB. Itu dipilih
sadar: portal venue di Fase 5 akan memakai proxy untuk melindungi rute, dan
perlindungan yang diam-diam tidak jalan di produksi jauh lebih mahal daripada
ruang disk.

**Service worker tidak menyimpan jawaban API.** Hanya aset ber-hash yang
di-cache. Status "Dikonfirmasi" berubah sampai menit terakhir menjelang kickoff;
menyajikan daftar lama akan menghasilkan persis kesalahan yang paling ingin
dihindari — data basi.
