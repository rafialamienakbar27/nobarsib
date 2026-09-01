# NOBARSIB — Rencana Pengerjaan

> Dokumen kerja. Dibuka tiap hari, dicoret tiap task selesai.
> Pendamping `blueprint-nobarsib.md` — bukan pengganti.
>
> Versi: 1.0 · Disusun: Agustus 2026

---

## Cara pakai dokumen ini

| Dokumen | Menjawab | Kapan dibuka |
|---|---|---|
| `blueprint-nobarsib.md` | **Apa** yang dibangun dan **kenapa** | Saat ragu soal keputusan desain |
| `rencana-pengerjaan-nobarsib.md` (ini) | **Kapan** dan **urutan apa** dikerjakan | Tiap hari, saat memilih task berikutnya |

Rujukan `§x.y` di dokumen ini menunjuk ke bagian bernomor di blueprint. Isinya
tidak disalin ulang supaya tidak ada dua versi yang berbeda saat blueprint direvisi.

### Aturan main

1. **Satu fase tidak dimulai sebelum gerbang fase sebelumnya lolos.** Gerbang
   ditulis sebagai pertanyaan ya/tidak, bukan penilaian rasa.
2. **Task perbaikan (🔧) wajib dikerjakan di fasenya.** Ini bukan
   "nanti kalau sempat" — keempatnya adalah bug yang sudah ada di blueprint dan
   akan muncul persis di titik yang paling merepotkan kalau ditunda.
3. **Kalau sebuah fase molor lebih dari 50% estimasi, potong scope-nya**, jangan
   perpanjang jadwal. Risiko terbesar proyek ini menurut §21 adalah "kamu
   kehabisan waktu", dan itu terjadi karena scope melar, bukan karena kurang jam.

### Keputusan yang sudah dikunci

Jangan buka ulang tiap minggu. Kalau mau berubah, catat alasannya di sini
beserta tanggalnya.

| Lapisan | Keputusan | Catatan |
|---|---|---|
| Backend | **Go + Fiber** | Sesuai arah karier backend; hemat resource di VPS 2 vCPU |
| Database | **PostgreSQL 16 + PostGIS** | Query geospasial native |
| Migrasi | **golang-migrate** | Satu file per perubahan, `up`/`down` lengkap |
| Cache | **Redis 7** | Daftar venue per laga + rate limit |
| Frontend | **Next.js App Router + Tailwind**, PWA | Native ditunda ke Fase 6+ |
| Peta | **Leaflet + OSM** untuk tampilan, link keluar ke Google Maps | Hemat kuota API |
| Storage | **Cloudflare R2** | Free tier 10 GB cukup untuk foto venue |
| Notifikasi venue | **WhatsApp (Fonnte)** | Email tidak dibaca pemilik cafe |
| Notifikasi pengguna | **Web Push** | Gratis, cukup untuk PWA |
| Deploy | **VPS Indonesia** + Docker Compose | Biznet Gio / IDCloudHost |

---

## Peta fase

```
Fase 0  Validasi lapangan          2 minggu    nol baris kode
   │    gerbang: ≥5 venue tertarik
   v
Fase 1  Fondasi backend            2 minggu
   │    gerbang: docker up + migrate bersih
   v
Fase 2  Domain inti & API publik   2 minggu    ← 3 dari 4 task perbaikan di sini
   │    gerbang: endpoint utama urut benar
   v
Fase 3  Frontend PWA               2 minggu
   │    gerbang: kamu sendiri pakai untuk memutuskan nonton di mana
   v
Fase 4  Panel admin & seeding      1 minggu
   │    gerbang: checklist rilis §23 lolos
   v
Fase 5  Kepercayaan                sepanjang musim   ← fase terpenting
   │    gerbang: konfirmasi H-1 > 70%
   v
Fase 6  Skala                      musim kedua, kondisional
```

Total sampai rilis: **±9 minggu** setelah Fase 0 lolos.

---

## Fase 0 — Validasi Lapangan

**Durasi:** 2 minggu · **Kode yang ditulis:** nol

Blueprint menutup dengan satu kalimat: *"Datangi venue dulu sebelum menulis
baris kode pertama."* Fase ini adalah kalimat itu.

### Task

- [ ] Susun daftar 10 venue kandidat dari §16.2 — ambil campuran ketiga kategori
      (sinyal kuat, ramah anak, buka sampai larut), jangan semua dari satu kategori
- [ ] Cek jadwal Persib, tandai 2–3 malam pertandingan dalam 2 minggu ke depan
- [ ] Kunjungi 8–10 venue **saat malam pertandingan**, bukan siang hari.
      Suasana nobar tidak bisa dinilai dari kafe yang sedang sepi
- [ ] Isi template survei lapangan §16.4 untuk tiap venue — semua 12 baris,
      termasuk yang terasa sepele seperti musala dan area non-smoking
- [ ] Verifikasi delapan hal di §16.3 yang tidak ada di Google Places.
      Terutama: benar menayangkan **Liga 1**, bukan cuma Liga Inggris
- [ ] Ajukan pertanyaan kunci ke tiap pemilik:
      *"Bapak/Ibu repot nggak memperkirakan berapa orang yang bakal datang pas
      malam pertandingan?"*
- [ ] Catat siapa yang menjawab "ya" — mereka calon pelanggan pertama
- [ ] Minta nomor WA pemilik/manajer, jelaskan akan dihubungi lagi

### Output

- 8–10 lembar survei terisi lengkap
- Daftar kontak WA pemilik venue
- Catatan siapa yang tertarik dan siapa yang tidak, beserta alasannya

### 🚪 Gerbang lanjut

> **Apakah minimal 5 venue menyatakan tertarik ikut kalau aplikasinya ada?**

**Tidak** → hentikan proyek. Ini bukan kegagalan, ini informasi yang dibeli
dengan 2 minggu alih-alih 3 bulan. Kalau tetap ingin lanjut, lanjutkan sebagai
proyek belajar backend murni dan turunkan ekspektasi bisnisnya ke nol —
jangan setengah-setengah.

---

## Fase 1 — Fondasi Backend  ✅ SELESAI

**Durasi:** 2 minggu · **Diselesaikan:** 30 Agustus 2026

Tidak ada logika bisnis di fase ini. Tujuannya satu: lingkungan yang bisa
dijalankan satu perintah dan database yang skemanya benar.

### Task

- [x] Inisialisasi repo `nobarsib-api/` dengan struktur folder §6.4:
      `cmd/api`, `cmd/worker`, `internal/{config,domain,repository,service,handler,middleware,pkg}`,
      `migrations/`, `docs/`
- [x] `go mod init`, pasang Fiber v2, pgx v5, golang-migrate, godotenv
- [x] `docker-compose.yml` sesuai §17.2 — `postgis/postgis:16-3.4` + `redis:7-alpine`,
      dengan healthcheck di service `db`
- [x] `Makefile`: target `up`, `down`, `migrate-up`, `migrate-down`, `test`, `lint`
- [x] Config loader dari environment variable, **tanpa nilai default production**
      (checklist rilis §23 mensyaratkan ini — lebih mudah dibenarkan sekarang)
- [x] Endpoint `/health` — cek koneksi Postgres dan Redis, bukan sekadar `200 OK`.
      UptimeRobot akan ping ini tiap 5 menit (§17.5)
- [x] Structured logging + request ID middleware

### Migrasi database

DDL §7.2 dipecah jadi beberapa file migrasi, bukan satu file raksasa:

| File | Isi |
|---|---|
| `000001_extensions` | `postgis`, `uuid-ossp`, `pg_trgm` |
| `000002_app_user` | `app_user` — **harus paling awal**, lihat 🔧#1 |
| `000003_match_reference` | `competition`, `team`, `match` + index |
| `000004_venue` | `venue`, `facility`, `venue_facility`, `venue_photo` + index GIST & GIN |
| `000005_nobar_event` | `nobar_event`, `event_source` + index |
| `000006_review` | `review`, `event_view` + index |

#### 🔧 Task perbaikan #1 — urutan DDL

Blueprint §7.2 membuat tabel `venue` (baris 475) dengan
`owner_user_id UUID REFERENCES app_user(id)`, padahal `app_user` baru dibuat di
baris 608. Masalah yang sama di `nobar_event.created_by` (baris 549).
DDL itu **gagal dijalankan apa adanya**.

Solusi yang dipakai: buat `app_user` di migrasi `000002`, sebelum semua tabel
yang mereferensinya.

Alternatif kalau urutan mau dipertahankan persis seperti blueprint: buat semua
tabel tanpa FK dulu, lalu satu migrasi terakhir berisi
`ALTER TABLE ... ADD CONSTRAINT`. Lebih berisik, tidak ada untungnya di sini.

### Seed data referensi

- [x] `competition` — Liga 1 (`is_active = TRUE`)
- [x] `team` — Persib Bandung (`is_featured = TRUE`) + 17 klub Liga 1 lain.
      Ikuti §3.1 prinsip 2: Persib disimpan sebagai **data**, tidak di-hardcode
      di logika program. Menambah Timnas nanti harus cukup satu `INSERT`.
      Ada di `seed/001_reference.sql`, idempoten (`ON CONFLICT DO NOTHING`),
      dipisah dari migrasi karena roster berubah tiap musim
- [x] `facility` — 13 baris dari §7.2, ikut migrasi `000004` karena kodenya
      dirujuk langsung oleh filter API (§8.2 param `facilities`)

### ✅ Definisi selesai — terverifikasi

- [x] `make infra` menyalakan db + redis, `make run` menjalankan API tanpa error
- [x] `make migrate-up` bersih dari nol; siklus `up → down -all → up` diuji ulang
      dari volume kosong dan bersih di ketiga langkah
- [x] `curl localhost:8080/health` → `200` dengan status kedua dependensi;
      saat Redis dimatikan → `503 degraded`, `/live` tetap `200`
- [x] Seed masuk: `competition=2, team=18, facility=13, persib_featured=1`;
      dijalankan dua kali, yang kedua `INSERT 0 0`
- [x] `go build`, `go vet`, `go test -race` bersih (config 89,7% coverage)
- [x] `EXPLAIN` pada `ST_DWithin` menunjukkan `Index Scan using idx_venue_location`

### Catatan pelaksanaan — tiga hal yang berbeda dari rencana

1. **`docker-compose.yml` dan `Makefile` diletakkan di root proyek**, bukan di
   dalam `nobarsib-api/`. Blueprint §6.4 menaruhnya di dalam `nobarsib-api/`
   tapi §17.2 membangun `./nobarsib-web` dari file yang sama — dua hal itu tidak
   bisa benar bersamaan. Compose mengorkestrasi dua aplikasi, jadi tempatnya di root.

2. **Port host Postgres 5434, bukan 5432.** Mesin ini menjalankan Homebrew
   `postgresql@16` di port 5433, dan instalasi itu mengikat `127.0.0.1` secara
   spesifik sehingga selalu menang atas binding `0.0.0.0` milik Docker.
   Gejalanya menyesatkan — `psql` dari host tampak berhasil tapi berisi database
   proyek lain. Cara memastikan ada di `README.md`.

3. **Migrasi `000001` tidak men-drop ekstensi saat `down`.** Image
   `postgis/postgis` sudah memasang `postgis_tiger_geocoder` yang bergantung
   pada `postgis`, sehingga `DROP EXTENSION` gagal dan menandai database dirty.
   `CASCADE` akan ikut menghapus ekstensi milik image. Untuk database benar-benar
   kosong, buang volumenya: `make reset`.

Selain itu ditambahkan di luar rencana, karena murah dan mencegah bug yang mahal:
`CHECK` constraint untuk semua kolom berenum (status, entry_type, crowd_level,
role, action) — termasuk `entry_type='free'` wajib `entry_amount=0` dan
`status='confirmed'` wajib punya `confirmed_at`; trigger `updated_at` yang
benar-benar memperbarui kolomnya; index parsial untuk worker penghitung skor
(§11.3) dan untuk pengecekan `open_detail` sebelum review (§11.5); serta
`start_period` + healthcheck TCP di compose, karena `pg_isready` lewat unix
socket lolos terlalu cepat saat volume masih kosong dan menyebabkan `EOF`.

---

## Fase 2 — Domain Inti & Endpoint Publik  ✅ SELESAI

**Durasi:** 2 minggu · **Diselesaikan:** 30 Agustus 2026

Fase paling padat secara teknis, dan tempat tiga dari empat task perbaikan
berada. Kerjakan perbaikannya bersamaan dengan fiturnya, jangan ditunda —
memperbaiki query pencarian setelah frontend menempel jauh lebih mahal.

### Domain & repository

- [x] Entity + interface repository di `internal/domain/`:
      `venue.go`, `match.go`, `nobar_event.go`, `review.go` (§6.4)
- [x] Implementasi Postgres di `internal/repository/`
- [x] CRUD venue + `venue_facility` + `venue_photo`
- [x] CRUD match — input manual dulu. §3.1 prinsip 3: *"Mengetik 30 venue butuh
      satu jam. Membangun scraper butuh berminggu-minggu dan hasilnya lebih buruk."*
      Sinkronisasi otomatis jadwal ditunda ke Fase 5
- [x] CRUD `nobar_event`
- [x] **State machine §4.5 divalidasi di layer service, bukan di handler.**
      Transisi sah: `draft → pending_review → published → confirmed → finished`,
      dengan `rejected` dari `pending_review` dan `cancelled` dari `published`/`confirmed`.
      Transisi lain ditolak dengan error, bukan diam-diam disimpan.
      Pakai huruf kecil konsisten dengan default kolom di §7.2 — blueprint
      menulisnya kapital di diagram §4.5, itu hanya gaya penulisan diagram

### Endpoint publik

- [x] `GET /v1/matches/upcoming` (§8.2) — termasuk `nobar_count` per laga
- [x] `GET /v1/matches/{match_id}/nobar` — **endpoint utama aplikasi** (§8.2).
      Semua query param: `lat`, `lng`, `sort`, `radius_km`, `facilities`,
      `entry_type`, `open_until_end`, `page`, `per_page`
- [x] `GET /v1/venues/{slug}` — detail lengkap (§8.2)
- [x] `GET /v1/venues/search` — pencarian nama pakai `pg_trgm`
- [x] `POST /v1/events/{id}/track` — fire-and-forget, tanpa response body.
      Ini sumber data statistik yang nanti dijual ke venue (§15.4), jadi
      pasang sekarang meski belum ada yang membacanya
- [x] Format error konsisten §8.1 — `{ error: { code, message, details } }`
- [x] Rate limiting §8.5 sebagai middleware Redis
- [x] Cache Redis daftar venue per laga, TTL 5 menit (1 menit setelah H-1) (§13.6)

### Query geospasial & skor

- [x] Query §9.1 dengan `ST_DWithin` — memakai index GIST, jauh lebih cepat
      daripada `ST_Distance` untuk semua baris lalu difilter
- [x] Fungsi skor rekomendasi §9.2 di `internal/pkg/scoring/`
- [x] Perhitungan `data_completeness` §9.3

#### 🔧 Task perbaikan #2 — filter jam tutup lewat tengah malam

§9.4 memfilter dengan `close::time > (kickoff + 2 jam)::time`. Untuk venue yang
tutup lewat tengah malam, perbandingan itu jadi `02:00 > 21:00` = **false** —
venue terbuang.

Yang terbuang justru kandidat terbaik di §16.2: Jabarano (tutup 04:00),
Rooftop Coffee (03:00), Ludo Sports Kitchen (03:00), Barrack Billiard (02:00),
Grow Billiard (02:00). Padahal §9.4 sendiri menulis filter ini ada supaya
"daftar tidak penuh venue yang percuma" — hasilnya justru kebalikannya.

Masalah kedua: `'24:00'::time` melempar error di Postgres. Blueprint mencoba
menanganinya dengan cabang `OR`, tapi cast di cabang pertama tetap dievaluasi
dan tetap error.

Solusi:
- Normalisasi `opening_hours` saat ditulis: simpan `close_minutes` sebagai menit
  sejak tengah malam, dengan nilai > 1440 untuk yang tutup keesokan harinya
  (02:00 → 1560). `24:00` jadi 1440, bukan string bermasalah
- Bandingkan dalam menit, bukan tipe `time`
- Test wajib: venue tutup 02:00 dengan kickoff 19:00 **harus lolos**;
  venue tutup 21:00 dengan kickoff 19:00 **harus terbuang**

#### 🔧 Task perbaikan #3 — skor dihitung setelah paginasi

Query §9.1 memakai `ORDER BY distance_km ASC LIMIT $5 OFFSET $6`, lalu
`CalculateScore` (§9.2) dijalankan di Go terhadap hasilnya. Artinya yang
diurutkan berdasarkan skor hanya 20 venue terdekat, bukan seluruh kandidat
dalam radius.

Akibatnya venue dengan rating tinggi dan sudah dikonfirmasi di km ke-8 tidak
akan pernah muncul kalau ada 20 venue lain yang lebih dekat — padahal mode
sort-nya bernama `recommended`, dan §9.2 memberi jarak bobot hanya 0,35.

Solusi: hitung skor **di SQL** sebagai expression di `ORDER BY`, sehingga
`LIMIT/OFFSET` bekerja pada urutan yang benar. Fungsi Go tetap dipertahankan
untuk unit test rumusnya.

Alternatif kalau rumusnya terlalu ramai di SQL: ambil seluruh kandidat dalam
radius (tanpa `LIMIT`), skor dan paginasi di memori. Aman selama jumlah venue
per laga di bawah ~500 — realistis untuk beberapa musim ke depan, tapi catat
sebagai utang teknis.

#### 🔧 Task perbaikan #4 — bobot promoted mengalahkan konfirmasi

§9.2 memberi `boost_promoted = 0.15` sementara `bonus_konfirmasi` hanya
berbobot `0.10`. Venue berbayar yang **belum** konfirmasi bisa naik di atas
venue yang **sudah** konfirmasi.

Ini bertentangan langsung dengan blueprint sendiri. §4.3 menyebut konfirmasi
H-1 sebagai "mekanisme paling penting di seluruh sistem", §21 menyebut data
basi sebagai risiko fatal yang membunuh FANZO, dan penutup blueprint menyebut
mekanisme ini bagian yang "harus dikerjakan paling serius". Lalu rumus
peringkatnya memberi uang bobot lebih besar.

Solusi: **boost promoted hanya berlaku kalau `confirmed_at IS NOT NULL`.**
Venue bayar tetap dapat keuntungan, tapi harus tetap konfirmasi untuk
mendapatkannya — insentif yang searah, bukan berlawanan.

Sekalian turunkan nilainya ke 0,05. §9.2 sendiri sudah memperingatkan
"jangan lebih, nanti hasilnya sampah".

### ✅ Definisi selesai — terverifikasi

Diuji dengan `make fixture` (27 venue, 27 event, laga Sabtu 19:00 WIB) yang
sengaja disusun agar ketiga perbaikan bisa dibuktikan, bukan sekadar diklaim.

- [x] `GET /v1/matches/{id}/nobar` benar untuk ketiga mode sort:
      `recommended` mendahulukan venue terkonfirmasi, `nearest` murni jarak,
      `rating` mendahulukan rating tertinggi
- [x] `EXPLAIN ANALYZE` bersih: `Index Scan using idx_nobar_match`, lalu join
      venue lewat primary key. Index GIST `idx_venue_location` terpakai saat
      query berangkat dari sisi venue
- [x] 🔧#2 terbukti: venue tutup 02:00–04:00 lolos, venue tutup 21:00 tersaring
      (total 26 dengan filter, 27 tanpa filter). Fungsi SQL `venue_open_until`
      diuji 8 kasus dan cocok persis dengan `openhours.Week.OpenUntil` di Go
- [x] 🔧#3 terbukti: Jabarano (8 km) menempati peringkat 3 dari 26, di atas
      22 venue medioker yang lebih dekat. Dengan cara blueprint ia berada di
      urutan ~23 menurut jarak dan tidak pernah muncul di halaman pertama
- [x] 🔧#4 terbukti: Ludo (2,9 km, berbayar, belum konfirmasi) di peringkat 4,
      di bawah ketiga venue terkonfirmasi
- [x] §11.4 terbukti: `kondusif_score` dan `kid_friendly_score` bernilai `null`
      selama review < 3
- [x] 15 dari 15 pemeriksaan endpoint lolos, termasuk 404 dan 400
- [x] `go build`, `go vet`, `go test -race` bersih di 5 paket

### Catatan pelaksanaan — yang berbeda dari rencana

1. **Ditambah migrasi `000007_query_functions`** berisi dua fungsi SQL:
   `venue_open_until()` (cermin SQL dari `internal/pkg/openhours`) dan
   `venue_data_completeness()` (§9.3). Fungsi 🔧#2 harus hidup di SQL karena
   filternya wajib berjalan sebelum `LIMIT` — alasan yang sama dengan 🔧#3.
   Test `TestSQLDanGoSepakat` menjaga bobot skor di Go dan di SQL tidak
   berpisah diam-diam.

2. **`POST /v1/reviews` tidak dibuat di fase ini.** Endpointnya milik Fase 5
   bersama anti-spam §11.5; di Fase 2 review hanya dibaca untuk halaman detail.
   Ditambahkan di luar rencana: `GET /v1/matches` (halaman /jadwal) dan
   `GET /v1/facilities` (chip filter §13.2) — keduanya dibutuhkan Fase 3.

3. **Dua bug ditemukan saat pengujian, bukan saat menulis:**
   `COALESCE` diletakkan di dalam subquery foto, sehingga venue tanpa foto
   membuat seluruh endpoint gagal `500`; dan perbandingan jam tutup memakai
   `>=` sementara §9.4 memakai perbandingan tegas — venue yang tutup persis di
   jam bubar seharusnya tidak direkomendasikan. Keduanya sudah diperbaiki.

4. **Cache dibatalkan saat status berubah**, bukan hanya menunggu TTL. Setiap
   kunci dicatat di set penunjuk per laga, sehingga konfirmasi H-1 langsung
   terlihat tanpa memindai Redis dengan `KEYS`.

---

## Fase 3 — Frontend PWA  ✅ SELESAI

**Durasi:** 2 minggu · **Diselesaikan:** 30 Agustus 2026

### Task

- [x] Inisialisasi `nobarsib-web/` — Next.js App Router + Tailwind + TypeScript
- [x] `/` — beranda: kartu laga berikutnya + daftar venue (struktur §13.2)
- [x] `/match/[id]` — daftar nobar untuk satu laga
- [x] `/venue/[slug]` — detail venue, sembilan blok berurutan sesuai §13.4
- [x] `/jadwal` — semua laga Persib musim ini
- [x] `/tentang` dan `/untuk-venue` — halaman statis, tapi `/untuk-venue`
      adalah pintu masuk akuisisi venue, jangan diperlakukan sebagai pelengkap
- [x] Deteksi lokasi + fallback pusat kota Bandung saat izin ditolak (§4.2)
- [x] Tab sort (Rekomendasi / Terdekat / Rating) + chip filter
- [x] Empat state kosong §13.5 — masing-masing dengan pesan dan aksi lanjutannya
      sendiri, bukan satu pesan generik. Ini paling sering muncul di awal saat
      data masih tipis, jadi justru yang paling banyak dilihat pengguna pertama
- [x] PWA manifest + service worker + ikon

### Aturan UI yang tidak boleh dilanggar (§13.3)

- Kartu venue **maksimal 5 baris informasi** — lebih dari itu tidak terbaca di HP
- Badge "Dikonfirmasi" harus menonjol — itu janji kepercayaan aplikasi
- Jarak **selalu** tampil, karena itu alasan utama orang memilih
- Thumbnail di daftar, foto besar hanya di halaman detail
- Tombol utama di detail: **"Buka di Maps"** dan **"Chat WA"**

### Performa (§13.6)

- [x] Target LCP < 2,5 detik di 4G
- [x] `next/image` + WebP
- [x] Lazy load peta — jangan muat SDK peta di beranda
- [x] Uji dengan throttle 3G di DevTools, bukan cuma di WiFi kantor

### ✅ Definisi selesai

Definisi dari blueprint §19 Fase 1, dan sengaja tidak diperlunak:

> **Kamu sendiri bisa memakainya untuk memutuskan nonton di mana pada laga berikutnya.**

Kalau di laga berikutnya kamu tetap membuka Instagram untuk memutuskan, fase ini
belum selesai — apa pun yang dikatakan checklist di atas.

**Status:** aplikasinya sudah bisa dipakai, tapi definisi di atas belum bisa
dijawab sampai ada data venue sungguhan (Fase 4). Yang bisa diverifikasi
sekarang adalah bahwa mekanismenya benar:

- [x] 19 dari 19 pemeriksaan lolos: 6 halaman, 2 kondisi 404, manifest, service
      worker, 3 ikon, halaman luring, header keamanan, dan backend tetap sehat
- [x] Seluruh stack jalan lewat `make up` — db, redis, api, worker, web
- [x] Server merender daftar venue lengkap sebelum JavaScript dimuat, sehingga
      halaman tetap berguna kalau izin lokasi ditolak dan LCP tidak menunggu
      dialog izin
- [x] Urutan hasil perbaikan #3 terbawa sampai ke HTML: Jabarano (8 km) di
      peringkat 3, di atas 22 venue yang lebih dekat
- [x] Jarak tampil di 20 dari 20 kartu; kartu tetap 5 baris — catatan venue dan
      daftar fasilitas sengaja tidak ikut
- [x] `tsc --noEmit` dan `eslint` bersih

### Catatan pelaksanaan — empat hal yang perlu diketahui

1. **`output: "standalone"` dilepas, dan ini keputusan penting.** Server minimal
   yang dihasilkannya tidak menerapkan `proxy.ts` maupun `headers()` dari
   `next.config.ts` — diverifikasi pada build yang sama: `next start` mengirim
   header keamanan, `node server.js` standalone tidak mengirim apa pun, tanpa
   error apa pun. Fase 5 akan memakai proxy untuk melindungi rute portal venue;
   perlindungan yang diam-diam tidak jalan di produksi jauh lebih mahal daripada
   image yang membengkak dari ±150 MB ke 793 MB.

2. **`service worker` sengaja tidak menyimpan jawaban API.** Hanya aset
   ber-hash yang di-cache; navigasi memakai jaringan lebih dulu dengan halaman
   luring sebagai cadangan. Menyajikan daftar venue lama justru menghasilkan
   persis kesalahan yang paling dihindari blueprint — data basi (§21).

3. **Foto masih memakai `<img>`, belum `next/image`.** `next/image` mewajibkan
   hostname sumber terdaftar di `images.remotePatterns`, sementara host
   penyimpanan foto (R2, §6.1) baru disiapkan di Fase 4. Mendaftarkan `**`
   akan mengubah pengoptimal gambar jadi proksi terbuka. **Ini utang yang harus
   dibayar di Fase 4**, sudah ditambahkan sebagai task di sana.

4. **Pemisahan laga mendatang/lampau di `/jadwal` memakai jam server, bukan jam
   perangkat.** Selain karena membaca jam saat render melanggar aturan kemurnian
   React, "sekarang" yang sah adalah jam database — jam yang sama yang dipakai
   memfilter laga dan menghitung konfirmasi H-1.

Ditambahkan di luar rencana karena murah dan jelas dibutuhkan: halaman `/luring`,
`not-found`, `error`, header keamanan lewat proxy, dan `make fixture` untuk
mengisi data uji.

---

## Fase 4 — Panel Admin & Seeding Data  ⏳ SEBAGIAN

**Durasi:** 1 minggu

### Panel admin

Bangun tiga halaman dulu, sisanya menyusul saat memang dibutuhkan:

- [x] **Antrian event** (§14.2) — layar yang paling sering dibuka.
      Target: satu event selesai ditinjau dalam **10 detik**. Poster, nama venue,
      laga, jam, sumber, lalu tiga tombol: Setujui / Edit / Tolak
- [x] **Kelola venue** — form tambah + pencarian; untuk banyak venue pakai importer
- [x] **Kelola jadwal** — tambah laga + daftar tersimpan
- [x] Auth admin: JWT + refresh token dengan rotasi (§6.1). Rate limit login
      10 percobaan per 15 menit (§8.5)
- [x] `make admin-create` — CLI membuat akun admin pertama
- [x] `make import` / `make import-check` — importer venue dari JSON, idempoten

Halaman Import IG, Moderasi Review, dan Statistik ditunda — masing-masing baru
berguna setelah fiturnya ada (Fase 5 dan 6).

### Seeding data

- [ ] Input manual **20–30 venue satu kecamatan**, bukan tersebar se-Bandung.
      §3.1 prinsip 4: *"Satu kecamatan lengkap > seluruh kota kosong."*
      Pakai hasil survei Fase 0 sebagai basis — datanya sudah terverifikasi langsung
- [ ] Tiap venue: koordinat, kontak WA, minimal 1 foto, minimal 3 fasilitas,
      jam buka lengkap 7 hari, IG handle → target `data_completeness > 0.8`
- [ ] Input jadwal Persib satu putaran penuh
- [ ] Buat `nobar_event` untuk laga terdekat berdasarkan konfirmasi venue
- [ ] Siapkan bucket Cloudflare R2 untuk foto venue (§6.1)
- [ ] **Utang dari Fase 3:** pindahkan foto venue dari `<img>` ke `next/image`.
      Baru bisa dikerjakan setelah bucket di atas ada, karena `next/image`
      mewajibkan hostname sumber terdaftar di `images.remotePatterns`. Titik yang
      perlu diubah ditandai komentar `eslint-disable @next/next/no-img-element`
      di `nobarsib-web/components/`

### ⏸ Yang belum bisa dikerjakan — menunggu kamu

Tiga hal ini bukan soal kode, jadi tidak bisa diselesaikan tanpamu:

1. **Data venue sungguhan.** Alatnya sudah siap: `make import` menerima berkas
   JSON, memvalidasi seluruhnya sebelum menulis apa pun (koordinat tertukar,
   jam tutup mustahil, hari yang belum lengkap, foto ganda), dan aman dijalankan
   berulang. Templatnya di `nobarsib-api/testdata/venues.contoh.json`.
   Isinya harus datang dari survei lapangan Fase 0 — §16.3 menegaskan justru
   data itulah yang tidak ada di Google Places dan menjadi nilai tambah aplikasi.

2. **Bucket Cloudflare R2** untuk foto venue — butuh akunmu.

3. **Utang Fase 3: `<img>` → `next/image`** — terhalang oleh nomor 2, karena
   hostname sumber wajib terdaftar di `images.remotePatterns`.

Sampai ketiganya beres, database masih berisi fixture pengembangan
(`make fixture`), bukan data yang bisa dipakai orang sungguhan.

### 🎨 Palet biru #171B87 — dikerjakan setelah Fase 4

Warna utama diganti ke **#171B87** atas permintaan, karena itu biru yang
diasosiasikan orang dengan Persib. Yang ikut berubah:

| Berkas | Perubahan |
| --- | --- |
| `app/globals.css` | Palet penuh terang & gelap, netral dicondongkan ke rona yang sama, kelas `.bidang-brand` dan `.judul-bagian` |
| `app/layout.tsx` | Header jadi bilah biru penuh, `themeColor` satu warna untuk kedua mode |
| `components/MatchCard.tsx` | Dipecah dua: kepala beranda berbidang biru, baris `/jadwal` tetap kartu terang |
| `Controls`, `VenueCard`, `EmptyState`, `VenueActions` | Tab & chip aktif diisi biru, jarak diberi warna, tombol utama seragam |
| `scripts/generate-icons.mjs`, `app/manifest.ts` | Ikon PWA dibuat ulang di #171B87 |
| `public/sw.js` | `VERSI` naik ke `nobarsib-v2` supaya cache lama tidak menyisakan ikon warna lama |

**Dua token biru, sengaja dipisah** — `--brand` untuk BIDANG (selalu gelap,
teks putih di atasnya) dan `--brand-accent` untuk TEKS di atas latar halaman
(ikut terang di mode gelap). Menyatukan keduanya adalah jebakan yang pasti
menggigit di mode gelap: satu token yang diterangkan demi teks akan mengubah
tombol jadi biru pucat bertulisan putih.

**Soal §22.** Blueprint melarang memakai logo dan identitas visual klub. Warna
tunggal tidak bisa dimonopoli siapa pun, dan aplikasi tetap memakai nama serta
lambangnya sendiri. Batasnya: jangan pernah menambahkan perisai, harimau, atau
tulisan "PERSIB" sebagai identitas visual.

Seluruh pasangan warna diperiksa terhadap WCAG AA — **22 pasangan, semuanya
lolos**, termasuk teks putih transparan di atas bidang biru pada kedua mode.

### 🐛 CORS — ditemukan saat menguji tampilan di browser

Halaman dirender di browser sungguhan (Chrome headless, viewport 390px) dan
muncul banner *"Gagal memuat daftar"*. Penyebabnya bukan tampilan: **API tidak
punya middleware CORS sama sekali**.

Akibatnya nyata dan tidak terlihat dari pengujian `curl`: web di `:3000` dan API
di `:8080` adalah dua origin berbeda, jadi setiap permintaan dari sisi klien —
tab urutan, chip filter, tombol "Aktifkan lokasi", perluasan radius, dan seluruh
panel admin — responsnya sampai lalu dibuang browser. Halaman hanya bekerja
selama isinya datang dari render server.

Perbaikannya:

- `CORS_ORIGINS` di `internal/config/config.go`, daftar dipisah koma, spasi dan
  garis miring di ujung dibersihkan
- Nilai `*` **ditolak saat boot** — API ini juga melayani `/v1/auth` dan
  `/v1/admin`; membukanya ke halaman mana pun adalah kesalahan yang baru
  terlihat setelah dimanfaatkan
- Origin tanpa skema ditolak, karena tidak akan pernah cocok dengan header
  `Origin` dan hanya akan gagal diam-diam
- `AllowCredentials` sengaja mati: autentikasi memakai header `Authorization`,
  bukan cookie
- 5 test baru di `internal/config`, dan `CORS_ORIGINS` ditambahkan ke `.env`,
  `.env.example`, serta `docker-compose.yml`

### 🐛 Jam buka: "tidak tercatat" ≠ "tutup"

Halaman detail menampilkan `Tutup` untuk setiap hari yang tidak ada di data,
sehingga venue yang jam bukanya baru terisi sebagian terbaca seolah tutup enam
hari seminggu. Sisi server sudah membedakan keduanya (§9.4,
`openhours.OpenUntil` mengembalikan `known`); UI-nya belum. Sekarang hari yang
tidak ada di data tertulis **"Belum tercatat"**.

### 🚪 Gerbang rilis

Checklist §23 dijalankan penuh — teknis, konten, UX, operasional. Tiga yang
paling sering dilewati dan paling mahal kalau dilewati:

- [ ] **Backup otomatis berjalan dan sudah diuji restore.** Backup yang belum
      pernah di-restore bukan backup
- [ ] **Diuji di HP Android kelas menengah**, bukan cuma di laptop
- [ ] **Environment variable produksi tidak ada nilai default development** —
      terutama `JWT_SECRET`, yang di `docker-compose.yml` §17.2 nilainya
      literal `dev-secret-ganti-di-production`

---

## Fase 5 — Kepercayaan

**Durasi:** sepanjang musim berjalan

Penutup blueprint menyebut ini bagian yang harus dikerjakan paling serius:

> *"Kalau ada satu bagian yang harus dikerjakan paling serius, itu adalah
> mekanisme konfirmasi H-1 dan kemudahan portal venue. Sisanya hanya pendukung."*

Perlakukan sesuai kalimat itu. Fase 1–4 hanya membuat aplikasi ini **ada**;
fase ini yang membuatnya **dipercaya**.

### Portal venue

- [ ] Login via **magic link WA berlaku 7 hari** (§15.2). Jangan minta login
      ulang tiap kali — pemilik cafe bukan orang teknis dan sedang sibuk
- [ ] Dashboard: daftar laga mendatang + tombol "Saya nayangin laga ini"
- [ ] **Umumkan nobar selesai dalam 3 ketukan** (§15.1). Tiga field saja:
      jam buka pintu, biaya, catatan. Setiap field tambahan memotong pemakaian
- [ ] Edit profil venue, fasilitas, jam buka, foto
- [ ] `POST /v1/venues/claim` — klaim venue `unclaimed` via OTP WA

### Konfirmasi H-1 — inti seluruh sistem

- [ ] Endpoint `/c/{token}` yang **langsung** mengubah status ke `confirmed`,
      tanpa halaman perantara, tanpa login (§15.3). Token sekali pakai, berlaku 24 jam
- [ ] Template pesan WA H-1 sesuai contoh §15.3
- [ ] Badge "Dikonfirmasi hari ini" di kartu venue
- [ ] Venue yang belum konfirmasi 12 jam sebelum kickoff: turun peringkat +
      badge "Belum dikonfirmasi" (§4.3). Jangan disembunyikan — informasinya
      tetap berguna, statusnya saja yang jujur

### Worker & cron (§12.3)

- [ ] `check-upcoming-matches` — 09:00 harian, cek laga H-3 dan H-1
- [ ] `remind-unconfirmed` — 19:00 harian, venue yang belum konfirmasi
- [ ] `review-prompt` — 08:00 harian, H+1 setelah laga
- [ ] `recalculate-scores` — 02:30 harian, query §11.3
- [ ] `sync-match-schedule` — Senin 03:00, tarik jadwal mingguan
- [ ] Notifikasi Web Push pengguna: H-1 sore dan H+1 pagi (§12.2)

### Sistem review

- [ ] `POST /v1/reviews` dengan validasi §8.2: event harus `finished`, maksimal
      7 hari setelah kickoff, satu `device_hash` satu review per event
- [ ] Form **tiga pertanyaan, tidak lebih** (§11.2). Blueprint menyebut tiga
      sebagai batas atas — setiap pertanyaan tambahan memotong pengisian
      secara signifikan
- [ ] Anti-spam §11.5: `device_hash` = SHA256(fingerprint + salt server),
      review hanya bisa dikirim kalau device pernah `open_detail` event itu,
      rate limit 5 review/jam per IP, flag otomatis untuk rating ekstrem dari
      subnet sama
- [ ] Perhitungan skor §11.3, dengan batas 18 bulan — venue berubah, review
      lama menyesatkan
- [ ] **Ambang tampil: minimal 3 review** (§11.4). Sebelum itu tampilkan
      "Belum ada penilaian", bukan skor dari satu orang yang kebetulan sedang kesal
- [ ] Halaman moderasi review di panel admin + mekanisme lapor (§22)

### Statistik venue (§15.4)

Ini yang nanti membuat venue mau bayar, jadi bangun sebelum ada yang diminta bayar:

- [ ] Jumlah yang melihat pengumuman, membuka detail, menekan "Buka di Maps"
- [ ] Perbandingan dengan rata-rata venue lain di laga yang sama
- [ ] Riwayat: laga mana yang paling menarik minat

### 🚪 Gerbang lanjut

> **Apakah tingkat konfirmasi venue H-1 di atas 70%?** (§20.1)

**Tidak** → jangan lanjut ke Fase 6. Masalahnya ada di friksi portal venue atau
di cara komunikasi WA, dan menambah fitur baru di atas fondasi yang tidak
dipercaya hanya memperbesar kerugian. Perbaiki dulu.

---

## Fase 6 — Skala

**Musim kedua · kondisional**

Fase ini **hanya dikerjakan kalau metrik §20.2 terpenuhi.** Setelah satu musim
penuh, jawab tiga pertanyaan:

1. Apakah pengguna kembali di laga berikutnya?
2. Apakah ada venue yang minta pasang lebih menonjol?
3. Apakah ada yang membagikan link aplikasi tanpa diminta?

Kalau ketiganya "tidak", berhenti di Fase 5. Blueprint sudah menyiapkan
jawabannya: *"produk ini tetap portofolio yang bagus tapi bukan bisnis.
Itu bukan kegagalan — itu informasi."*

Kalau lolos:

- [ ] Ekstraksi poster Instagram (§10) — blueprint sendiri bilang "jangan di awal",
      baru relevan saat volume submit melebihi kemampuan input manual,
      kira-kira **50+ event per laga**. Hasil ekstraksi **tidak pernah** langsung
      publish, selalu lewat review manusia (§10.4)
- [ ] Perluas cakupan ke Bandung Raya
- [ ] Tambah Timnas Indonesia — ekspansi paling alami, dan menaikkan frekuensi
      pakai dari 15–20 malam/tahun yang jadi kelemahan struktural produk ini
- [ ] Langganan venue + sistem promoted (dengan syarat konfirmasi dari 🔧#4)
- [ ] Statistik lanjutan

---

## Tabel gerbang keputusan

| Setelah | Pertanyaan | Kalau tidak lolos |
|---|---|---|
| Fase 0 | ≥5 venue menyatakan tertarik? | Hentikan, atau lanjut sebagai proyek belajar dengan ekspektasi bisnis nol |
| Fase 1 | `docker compose up` + `migrate up` bersih? | Jangan lanjut — masalah fondasi berlipat di fase berikutnya |
| Fase 2 | Endpoint utama urut benar, index terpakai? | Perbaiki sekarang; setelah frontend menempel jauh lebih mahal |
| Fase 3 | Kamu sendiri pakai untuk memutuskan nonton di mana? | Cari tahu kenapa kamu masih buka Instagram, benahi itu |
| Fase 4 | Checklist rilis §23 lolos penuh? | Tunda rilis. Kesan pertama tidak bisa diulang |
| Fase 5 | Konfirmasi H-1 > 70%? | Benahi friksi portal venue, jangan tambah fitur |
| Fase 5 | Tiga pertanyaan §20.2 terjawab "ya"? | Berhenti di Fase 5, jangan bangun Fase 6 |

---

## Yang sengaja tidak dibangun

Ditulis eksplisit supaya tidak diam-diam masuk scope di tengah jalan.
Sumber: §2.3 dan §3.2.

| Tidak dibangun | Alasan | Kapan boleh ditinjau ulang |
|---|---|---|
| Reservasi kursi realtime | Terlalu kompleks, bergantung disiplin operasional venue | Fase 6+, hanya kalau venue sudah disiplin pakai portal |
| Pembayaran / DP online | Bangun audiens dulu, bukan sistem bayar (§5.3) | Setelah ada 2.000 pengguna aktif malam laga |
| Streaming pertandingan | Ilegal (§22) | Tidak pernah |
| Login pengguna | MVP pakai `device_hash`; login menambah friksi tanpa nilai jelas | Kalau review antar-device jadi masalah nyata |
| Cabang olahraga lain | Fokus dulu, Timnas lebih dekat | Setelah Timnas berhasil |
| Aplikasi native | 15–20 malam/tahun, orang tidak akan menginstal (§6.2) | Kalau push notification jadi fitur inti dan retensi terbukti |

---

## Catatan penutup

Estimasi 9 minggu di dokumen ini adalah waktu **menulis kode**. Yang tidak
masuk hitungan, dan justru yang menentukan hasilnya:

- Menjaga data tetap benar setiap malam pertandingan
- Meyakinkan venue untuk repot sedikit mengisi informasi

Blueprint menutup dengan pengingat bahwa FANZO punya dana, tim, dan 12 tahun
pengalaman, dan tetap gagal persis di dua hal itu. Query geospasial dan skema
database di Fase 1–2 akan selesai sesuai jadwal. Fase 5 tidak punya tanggal
selesai karena memang tidak ada.

**Mulai dari Fase 0.**
