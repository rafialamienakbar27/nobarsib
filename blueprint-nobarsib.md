# NOBARSIB — Blueprint Aplikasi Pencari Lokasi Nobar Persib

> Dokumen perencanaan lengkap: dari proses bisnis, arsitektur, database, backend,
> frontend, sampai peluncuran.
> Nama "NOBARSIB" adalah nama kerja sementara — ganti sesukanya.
>
> Versi dokumen: 1.0
> Disusun: Agustus 2026

---

## Daftar Isi

1. [Ringkasan Eksekutif](#1-ringkasan-eksekutif)
2. [Masalah yang Dipecahkan](#2-masalah-yang-dipecahkan)
3. [Batasan Scope & Prinsip Desain](#3-batasan-scope--prinsip-desain)
4. [Proses Bisnis](#4-proses-bisnis)
5. [Model Monetisasi](#5-model-monetisasi)
6. [Arsitektur Teknis](#6-arsitektur-teknis)
7. [Skema Database](#7-skema-database)
8. [Desain API](#8-desain-api)
9. [Algoritma Pencarian & Rekomendasi](#9-algoritma-pencarian--rekomendasi)
10. [Fitur Ekstraksi Poster Instagram](#10-fitur-ekstraksi-poster-instagram)
11. [Sistem Review & Skor Kondusif](#11-sistem-review--skor-kondusif)
12. [Notifikasi](#12-notifikasi)
13. [Frontend — Struktur & Halaman](#13-frontend--struktur--halaman)
14. [Panel Admin](#14-panel-admin)
15. [Portal Venue](#15-portal-venue)
16. [Seeding Data Awal](#16-seeding-data-awal)
17. [Deployment & Infrastruktur](#17-deployment--infrastruktur)
18. [Estimasi Biaya Operasional](#18-estimasi-biaya-operasional)
19. [Roadmap Bertahap](#19-roadmap-bertahap)
20. [Metrik Keberhasilan](#20-metrik-keberhasilan)
21. [Risiko & Mitigasi](#21-risiko--mitigasi)
22. [Aspek Legal](#22-aspek-legal)
23. [Checklist Sebelum Rilis](#23-checklist-sebelum-rilis)

---

## 1. Ringkasan Eksekutif

**Apa ini:** Direktori tempat nobar Persib di Bandung. Pengguna membuka aplikasi
sebelum pertandingan, melihat venue mana saja yang menayangkan laga malam itu,
lalu memilih berdasarkan jarak, rating, fasilitas, dan seberapa kondusif
suasananya.

**Siapa penggunanya:**
- **Sisi permintaan** — bobotoh yang tidak ke stadion dan tidak mau nonton sendirian di rumah.
- **Sisi penawaran** — cafe, resto, dan warkop di Bandung yang mengadakan nobar.

**Nilai pembeda:** Informasi yang tidak pernah ada di poster Instagram —
apakah suasananya kondusif, aman dibawa anak, suaranya terdengar jelas,
parkirnya cukup. Data ini hanya bisa datang dari penonton yang sudah pernah
datang, dan itulah yang membuat aplikasi ini sulit ditiru setelah 2–3 musim.

**Kenyataan yang harus diterima sejak awal:** Persib main sekitar 34 laga liga
per musim. Kira-kira setengahnya laga tandang yang layak dinobarkan. Artinya
aplikasi ini punya alasan dibuka sekitar **15–20 malam per tahun**. Retensi
alaminya rendah. Semua keputusan desain di dokumen ini mempertimbangkan hal itu.

---

## 2. Masalah yang Dipecahkan

### 2.1 Dari sisi penonton

| Pertanyaan | Kondisi sekarang | Setelah ada aplikasi |
|---|---|---|
| Di mana ada nobar malam ini? | Scroll Instagram, tanya grup WA | Satu layar, terurut jarak |
| Tempatnya proper atau tidak? | Tidak tahu sampai datang | Rating khusus konteks nobar |
| Aman bawa anak? | Tebak-tebakan | Tag ramah anak + skor kondusif |
| Bakal ramai atau sepi? | Tidak tahu | Indikator perkiraan keramaian |
| Berapa biayanya? | Sering tidak tertulis | Field biaya masuk / minimum order |
| Tutup jam berapa? | Sering luput | Filter otomatis vs jam kickoff |

### 2.2 Dari sisi venue

- Sulit memperkirakan jumlah pengunjung → salah hitung stok dan staf.
- Promosi hanya menjangkau follower Instagram sendiri.
- Tidak punya data historis: laga mana yang ramai, mana yang sepi.

### 2.3 Masalah yang **sengaja tidak** dipecahkan di versi awal

- Reservasi kursi realtime (terlalu kompleks, bergantung disiplin operasional venue)
- Pembayaran / DP online
- Streaming pertandingan (jelas ilegal)
- Cabang olahraga lain

---

## 3. Batasan Scope & Prinsip Desain

### 3.1 Lima prinsip yang mengunci keputusan

1. **Berpusat pada pertandingan, bukan pada venue.**
   Pengguna membuka aplikasi dengan pertanyaan "Sabtu nanti nonton di mana",
   bukan "cafe apa saja yang ada di Bandung". Jadwal laga adalah kerangka utama;
   venue menempel di bawahnya.

2. **Jangan hardcode Persib.**
   Tim dan kompetisi disimpan sebagai data, bukan logika program. Tampilan dan
   pemasaran khusus Persib, tapi menambah Timnas nanti cukup satu baris insert.

3. **Data manual dulu, otomasi belakangan.**
   Mengetik 30 venue butuh satu jam. Membangun scraper butuh berminggu-minggu
   dan hasilnya lebih buruk.

4. **Satu kecamatan lengkap > seluruh kota kosong.**
   Marketplace mati kalau kedua sisi sepi. Mulai dari area sekitar rumah.

5. **Jangan janji yang tidak bisa ditepati.**
   Kalau tidak yakin ramai atau tidak, tampilkan "Perkiraan ramai", bukan angka
   pasti. Sekali pengguna merasa dibohongi, mereka tidak kembali.

### 3.2 Definisi MVP

MVP dianggap selesai kalau seorang bobotoh bisa:

1. Buka aplikasi, langsung lihat laga Persib terdekat.
2. Lihat daftar venue yang menayangkan laga itu, terurut jarak.
3. Filter: ramah anak, indoor/outdoor, gratis/bayar, masih buka saat laga selesai.
4. Buka detail venue → alamat, fasilitas, tombol ke Google Maps, kontak WA.
5. Setelah laga, isi review 3 pertanyaan.

Yang **tidak** masuk MVP: login pengguna, reservasi, pembayaran, chat, notifikasi push.

---

## 4. Proses Bisnis

### 4.1 Aktor

| Aktor | Peran | Akses |
|---|---|---|
| Pengunjung (bobotoh) | Cari venue, baca info, kirim review | Publik, tanpa login |
| Pemilik venue | Daftarkan tempat, umumkan nobar per laga | Portal venue (login) |
| Admin (kamu) | Verifikasi venue, moderasi, kelola jadwal | Panel admin |

### 4.2 Alur utama — Pengunjung mencari nobar

```
[Buka aplikasi]
       |
       v
[Sistem ambil laga Persib berikutnya (H-7 s/d H+0)]
       |
       v
[Minta izin lokasi]  --tolak-->  [Fallback: pusat kota Bandung]
       |
       v
[Query venue yang punya nobar_event untuk laga tsb]
       |
       v
[Filter otomatis: jam_tutup > jam_kickoff + 2 jam]
       |
       v
[Urutkan: Rekomendasi / Terdekat / Rating tertinggi]
       |
       v
[Tampilkan kartu venue]
       |
       v
[Ketuk kartu] --> [Halaman detail] --> [Buka Maps / WA venue]
       |
       v
[H+1 setelah laga: muncul prompt review]
```

### 4.3 Alur — Venue mengumumkan nobar

```
[Venue login ke portal]
       |
       v
[Lihat daftar laga Persib mendatang]
       |
       v
[Klik "Saya nayangin laga ini"]
       |
       v
[Isi: jam buka pintu, biaya, kapasitas kira-kira, catatan]
       |
       v
[Status: PENDING] --> [Admin verifikasi] --> [Status: PUBLISHED]
       |
       v
[H-1: sistem kirim WA/email konfirmasi ulang]
       |
       +--[Venue konfirmasi]--> tetap tampil, badge "Dikonfirmasi hari ini"
       |
       +--[Tidak konfirmasi 12 jam sebelum kickoff]--> turun ke bawah,
                                                       badge "Belum dikonfirmasi"
```

**Ini mekanisme paling penting di seluruh sistem.** FANZO di Inggris gagal
persis di titik ini — venue mengisi data lalu tidak diperbarui, pengguna datang
dan pertandingannya tidak ada. Konfirmasi H-1 adalah pembeda antara direktori
yang dipercaya dan direktori yang ditinggalkan.

### 4.4 Alur — Admin menambah venue baru (manual)

```
[Admin cari venue di Google Places / IG]
       |
       v
[Input: nama, alamat, koordinat, IG handle, telepon]
       |
       v
[Isi checklist fasilitas: layar, indoor/outdoor, parkir, musala, area anak]
       |
       v
[Simpan status: UNCLAIMED]  (venue belum punya akun)
       |
       v
[Venue bisa klaim nanti lewat verifikasi nomor telepon]
```

### 4.5 State machine — `nobar_event`

```
DRAFT ──> PENDING_REVIEW ──> PUBLISHED ──> CONFIRMED ──> FINISHED
                │                │             │
                └──> REJECTED    └──> CANCELLED└──> CANCELLED
```

| Status | Arti | Tampil ke publik? |
|---|---|---|
| `DRAFT` | Hasil ekstraksi poster, belum ditinjau | Tidak |
| `PENDING_REVIEW` | Venue submit, menunggu admin | Tidak |
| `PUBLISHED` | Sudah diverifikasi | Ya |
| `CONFIRMED` | Venue konfirmasi ulang di H-1 | Ya, dengan badge |
| `CANCELLED` | Dibatalkan | Tidak |
| `FINISHED` | Laga sudah lewat | Arsip, jadi basis review |

---

## 5. Model Monetisasi

Bagian ini ditulis apa adanya karena inilah titik terlemah dari ide berbasis
fanbase.

### 5.1 Kenapa iklan bukan jawaban

Dengan 15–20 malam aktif per tahun, traffic tahunan aplikasi ini kemungkinan
setara satu blog kecil. Pendapatan iklan display di angka itu tidak menutup
biaya server, apalagi waktu pengembangan.

### 5.2 Sumber pendapatan yang realistis

| Model | Cara kerja | Estimasi | Kapan dimulai |
|---|---|---|---|
| **Langganan venue** | Rp 100–300rb/bln: muncul di urutan atas, banner, statistik pengunjung | Butuh 20+ venue bayar baru berarti | Setelah traffic terbukti |
| **Promosi per event** | Rp 25–50rb sekali tayang untuk laga besar (derby, final) | Lebih mudah dijual daripada langganan | Musim kedua |
| **Fee komunitas nobar** | Penyelenggara nobar berbayar (komunitas bobotoh) pasang info + jual tiket | Komisi 5–10% | Kalau fitur tiket dibangun |
| **Sponsor lokal** | Brand kopi/rokok/apparel lokal sponsori satu musim | Paling besar tapi butuh angka traffic dulu | Musim kedua |

### 5.3 Saran jujur soal urutan

Jangan bangun sistem pembayaran di tahun pertama. Bangun audiens dulu.
Kalau setelah satu musim ada 2.000 pengguna aktif di malam laga, kamu punya
angka untuk dijual ke venue dan sponsor. Kalau tidak sampai, monetisasi apapun
tidak akan jalan dan lebih baik tahu itu sebelum menulis kode pembayaran.

### 5.4 Jalur alternatif yang lebih cepat menghasilkan

Kalau tujuan utamanya uang, catat bahwa aplikasi ini lebih tepat diperlakukan
sebagai **portofolio dan alat belajar backend**, sementara pendapatan dikejar
lewat produk B2B (sistem manajemen rental, sistem operasional jasa kirim) yang
pelanggannya sudah punya anggaran. Dua-duanya bisa jalan paralel.

---

## 6. Arsitektur Teknis

### 6.1 Stack yang direkomendasikan

| Lapisan | Pilihan | Alasan |
|---|---|---|
| Backend | **Go (Gin/Fiber)** atau **NestJS** | Sesuai arah karier backend; Go lebih menantang dan hemat resource |
| Database | **PostgreSQL 16 + PostGIS** | Query geospasial native, atomic update, JSON support |
| Cache | **Redis** | Cache daftar venue per laga, rate limit |
| Storage | **Cloudflare R2** atau **S3** | Foto venue & poster, tanpa egress fee (R2) |
| Frontend | **Next.js (App Router) + Tailwind** | PWA dulu, native belakangan |
| Peta | **Google Maps Embed / Leaflet + OSM** | Leaflet gratis, Google lebih akurat di Indonesia |
| Notifikasi | **WhatsApp Business API** atau **Fonnte** | WA jauh lebih efektif dari email di Indonesia |
| Auth | **JWT + refresh token** | Hanya untuk venue & admin |
| Deploy | **VPS Indonesia** (Biznet/IDCloudHost) | Latensi rendah, harga bersahabat |

### 6.2 Kenapa PWA dulu, bukan native

- Tidak ada friksi install — bobotoh cukup buka link dari Instagram.
- Update instan, tidak menunggu review Play Store.
- Satu codebase.
- Dengan pola pemakaian 15–20 malam setahun, orang **tidak akan** menginstal
  aplikasi native. Ini bukan asumsi, ini konsekuensi dari frekuensi pemakaian.

Native baru masuk akal kalau notifikasi push jadi fitur inti dan retensi sudah
terbukti.

### 6.3 Diagram arsitektur

```
┌─────────────────────────────────────────────────────┐
│                    PENGGUNA                          │
│  PWA (Next.js)   Portal Venue   Panel Admin          │
└────────────┬────────────┬───────────────┬───────────┘
             │            │               │
             └────────────┴───────────────┘
                          │ HTTPS / REST
                          v
             ┌────────────────────────────┐
             │      API Gateway (Nginx)    │
             │   rate limit, TLS, gzip     │
             └────────────┬───────────────┘
                          v
             ┌────────────────────────────┐
             │      Backend (Go/Nest)      │
             │  ┌──────────────────────┐   │
             │  │ Handler / Controller │   │
             │  ├──────────────────────┤   │
             │  │ Service (bisnis)     │   │
             │  ├──────────────────────┤   │
             │  │ Repository (data)    │   │
             │  └──────────────────────┘   │
             └───┬──────────┬──────────┬───┘
                 │          │          │
                 v          v          v
        ┌──────────┐ ┌─────────┐ ┌──────────────┐
        │ Postgres │ │  Redis  │ │  Object      │
        │ +PostGIS │ │  cache  │ │  Storage     │
        └──────────┘ └─────────┘ └──────────────┘
                 ^
                 │
        ┌────────┴─────────────────────────┐
        │      Worker / Cron                │
        │  • sync jadwal Persib             │
        │  • reminder konfirmasi H-1        │
        │  • prompt review H+1              │
        │  • hitung ulang skor kondusif     │
        │  • ekstraksi poster IG (queue)    │
        └───────────────────────────────────┘
```

### 6.4 Struktur folder backend (contoh Go)

```
nobarsib-api/
├── cmd/
│   ├── api/main.go
│   └── worker/main.go
├── internal/
│   ├── config/
│   ├── domain/           # entity + interface repository
│   │   ├── venue.go
│   │   ├── match.go
│   │   ├── nobar_event.go
│   │   └── review.go
│   ├── repository/       # implementasi Postgres
│   ├── service/          # logika bisnis
│   ├── handler/          # HTTP handler
│   ├── middleware/       # auth, ratelimit, logging
│   └── pkg/
│       ├── geo/          # helper haversine, bounding box
│       ├── scoring/      # algoritma rekomendasi
│       └── notif/        # WA / email
├── migrations/           # goose / golang-migrate
├── docs/
│   └── openapi.yaml
├── docker-compose.yml
└── Makefile
```

---

## 7. Skema Database

### 7.1 Diagram relasi

```
   team                competition
     │                      │
     └──────┬───────────────┘
            v
          match ──────────────┐
                              │
   venue ──── venue_facility  │
     │                        │
     ├──── venue_photo        │
     │                        │
     └────────> nobar_event <─┘
                    │
                    ├──> review
                    └──> event_source
```

### 7.2 DDL lengkap (PostgreSQL)

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;   -- untuk pencarian nama venue

-- =====================================================
-- REFERENSI PERTANDINGAN
-- =====================================================

CREATE TABLE competition (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,      -- 'Liga 1', 'Piala Presiden'
    slug        VARCHAR(60) UNIQUE NOT NULL,
    country     VARCHAR(50) DEFAULT 'Indonesia',
    is_active   BOOLEAN DEFAULT TRUE
);

CREATE TABLE team (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,      -- 'Persib Bandung'
    short_name  VARCHAR(30),                -- 'Persib'
    slug        VARCHAR(60) UNIQUE NOT NULL,
    logo_url    TEXT,
    is_featured BOOLEAN DEFAULT FALSE       -- Persib = TRUE
);

CREATE TABLE match (
    id              BIGSERIAL PRIMARY KEY,
    competition_id  INT REFERENCES competition(id),
    home_team_id    INT NOT NULL REFERENCES team(id),
    away_team_id    INT NOT NULL REFERENCES team(id),
    kickoff_at      TIMESTAMPTZ NOT NULL,
    venue_name      VARCHAR(150),           -- stadion
    broadcast_tv    VARCHAR(100),           -- 'Indosiar', 'Vidio'
    status          VARCHAR(20) DEFAULT 'scheduled',
                    -- scheduled | live | finished | postponed
    score_home      SMALLINT,
    score_away      SMALLINT,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_match_kickoff ON match(kickoff_at);
CREATE INDEX idx_match_teams   ON match(home_team_id, away_team_id);

-- =====================================================
-- VENUE
-- =====================================================

CREATE TABLE venue (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name                VARCHAR(150) NOT NULL,
    slug                VARCHAR(180) UNIQUE NOT NULL,
    address             TEXT NOT NULL,
    district            VARCHAR(80),        -- kecamatan
    city                VARCHAR(80) DEFAULT 'Kota Bandung',
    location            GEOGRAPHY(POINT, 4326) NOT NULL,
    phone               VARCHAR(30),
    whatsapp            VARCHAR(30),
    instagram_handle    VARCHAR(60),
    google_place_id     VARCHAR(120) UNIQUE,
    google_rating       NUMERIC(2,1),
    google_rating_count INT,

    -- jam operasional per hari (0=Minggu ... 6=Sabtu)
    -- format: {"0": {"open":"08:00","close":"23:00"}, ...}
    opening_hours       JSONB,

    -- skor internal
    nobar_rating        NUMERIC(3,2),       -- rata-rata review nobar
    nobar_rating_count  INT DEFAULT 0,
    kondusif_score      NUMERIC(3,2),       -- 1.00 - 5.00
    kid_friendly_score  NUMERIC(3,2),
    data_completeness   NUMERIC(3,2) DEFAULT 0,  -- 0-1, untuk ranking

    status              VARCHAR(20) DEFAULT 'unclaimed',
                        -- unclaimed | claimed | verified | suspended
    owner_user_id       UUID REFERENCES app_user(id),
    is_active           BOOLEAN DEFAULT TRUE,

    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_venue_location ON venue USING GIST(location);
CREATE INDEX idx_venue_district ON venue(district);
CREATE INDEX idx_venue_name_trgm ON venue USING GIN(name gin_trgm_ops);

-- Fasilitas dibuat tabel terpisah supaya bisa nambah tanpa migrasi
CREATE TABLE facility (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(40) UNIQUE NOT NULL,
    label       VARCHAR(80) NOT NULL,
    icon        VARCHAR(40),
    category    VARCHAR(40)      -- 'nonton' | 'kenyamanan' | 'akses'
);

INSERT INTO facility (code, label, category) VALUES
('layar_besar',    'Layar besar / proyektor', 'nonton'),
('multi_layar',    'Lebih dari 1 layar',      'nonton'),
('sound_system',   'Sound system memadai',    'nonton'),
('indoor',         'Area indoor',             'kenyamanan'),
('outdoor',        'Area outdoor',            'kenyamanan'),
('ac',             'Ber-AC',                  'kenyamanan'),
('area_anak',      'Ramah anak',              'kenyamanan'),
('musala',         'Musala',                  'kenyamanan'),
('toilet_bersih',  'Toilet bersih',           'kenyamanan'),
('parkir_mobil',   'Parkir mobil',            'akses'),
('parkir_motor',   'Parkir motor',            'akses'),
('non_smoking',    'Ada area non-smoking',    'kenyamanan'),
('wifi',           'WiFi',                    'kenyamanan');

CREATE TABLE venue_facility (
    venue_id    UUID REFERENCES venue(id) ON DELETE CASCADE,
    facility_id INT  REFERENCES facility(id),
    note        VARCHAR(200),
    PRIMARY KEY (venue_id, facility_id)
);

CREATE TABLE venue_photo (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_id    UUID REFERENCES venue(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    caption     VARCHAR(200),
    is_primary  BOOLEAN DEFAULT FALSE,
    sort_order  SMALLINT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- =====================================================
-- NOBAR EVENT (inti aplikasi)
-- =====================================================

CREATE TABLE nobar_event (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_id            UUID NOT NULL REFERENCES venue(id) ON DELETE CASCADE,
    match_id            BIGINT NOT NULL REFERENCES match(id) ON DELETE CASCADE,

    doors_open_at       TIMESTAMPTZ,        -- jam buka pintu
    entry_type          VARCHAR(30) NOT NULL DEFAULT 'free',
                        -- free | min_order | ticket | donation
    entry_amount        INT DEFAULT 0,      -- rupiah
    capacity_estimate   INT,                -- perkiraan kursi tersedia
    crowd_level         VARCHAR(20),        -- longgar | ramai | penuh
                                            -- diisi manual venue atau dihitung
    notes               TEXT,               -- 'bawa jersey dapat diskon'

    status              VARCHAR(20) NOT NULL DEFAULT 'draft',
    confirmed_at        TIMESTAMPTZ,        -- kapan venue konfirmasi H-1
    is_promoted         BOOLEAN DEFAULT FALSE,  -- venue berbayar

    created_by          UUID REFERENCES app_user(id),
    created_at          TIMESTAMPTZ DEFAULT now(),
    updated_at          TIMESTAMPTZ DEFAULT now(),

    UNIQUE (venue_id, match_id)
);

CREATE INDEX idx_nobar_match  ON nobar_event(match_id, status);
CREATE INDEX idx_nobar_venue  ON nobar_event(venue_id);
CREATE INDEX idx_nobar_status ON nobar_event(status, confirmed_at);

-- Jejak asal data (dari IG, input manual, submit venue)
CREATE TABLE event_source (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nobar_event_id  UUID REFERENCES nobar_event(id) ON DELETE CASCADE,
    source_type     VARCHAR(30) NOT NULL,   -- instagram | manual | venue_portal
    source_url      TEXT,
    raw_caption     TEXT,
    poster_image_url TEXT,
    extracted_json  JSONB,                  -- hasil mentah ekstraksi AI
    extracted_at    TIMESTAMPTZ,
    reviewed_by     UUID REFERENCES app_user(id),
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- =====================================================
-- REVIEW
-- =====================================================

CREATE TABLE review (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nobar_event_id  UUID REFERENCES nobar_event(id) ON DELETE SET NULL,
    venue_id        UUID NOT NULL REFERENCES venue(id) ON DELETE CASCADE,

    -- 3 pertanyaan inti
    rating_overall  SMALLINT NOT NULL CHECK (rating_overall BETWEEN 1 AND 5),
    rating_kondusif SMALLINT CHECK (rating_kondusif BETWEEN 1 AND 5),
    is_kid_friendly BOOLEAN,

    -- opsional
    crowd_actual    VARCHAR(20),    -- sepi | pas | penuh | overload
    comment         VARCHAR(500),

    -- anti-spam tanpa login
    device_hash     VARCHAR(64) NOT NULL,
    ip_hash         VARCHAR(64),

    is_hidden       BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT now(),

    UNIQUE (nobar_event_id, device_hash)
);

CREATE INDEX idx_review_venue ON review(venue_id, created_at DESC);

-- =====================================================
-- USER (hanya venue & admin)
-- =====================================================

CREATE TABLE app_user (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(150) UNIQUE,
    phone           VARCHAR(30) UNIQUE,
    password_hash   TEXT,
    full_name       VARCHAR(120),
    role            VARCHAR(20) NOT NULL DEFAULT 'venue_owner',
                    -- admin | venue_owner
    is_active       BOOLEAN DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- =====================================================
-- ANALITIK RINGAN
-- =====================================================

CREATE TABLE event_view (
    id              BIGSERIAL PRIMARY KEY,
    nobar_event_id  UUID REFERENCES nobar_event(id) ON DELETE CASCADE,
    device_hash     VARCHAR(64),
    action          VARCHAR(30),    -- view_card | open_detail | open_maps | click_wa
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_view_event ON event_view(nobar_event_id, action, created_at);
```

### 7.3 Catatan desain penting

**Kenapa `opening_hours` pakai JSONB, bukan tabel terpisah?**
Jam buka jarang di-query terpisah dan selalu diambil utuh bersama venue.
JSONB menghindari 7 baris per venue dan join tambahan.

**Kenapa fasilitas pakai tabel, bukan kolom boolean?**
Karena daftarnya akan tumbuh. Menambah "ada photobooth" cukup satu insert,
bukan migrasi `ALTER TABLE`.

**Kenapa `device_hash` bukan `user_id` di review?**
Karena MVP tanpa login. Hash dari fingerprint browser + salt cukup untuk
mencegah spam ringan. Kalau nanti ada login, tambah kolom `user_id` nullable.

**Kenapa `UNIQUE (venue_id, match_id)` di nobar_event?**
Satu venue hanya boleh punya satu event per pertandingan. Mencegah duplikasi
saat impor dari IG dan submit venue bertabrakan.

---

## 8. Desain API

### 8.1 Konvensi

- Base URL: `https://api.nobarsib.id/v1`
- Format: JSON, `snake_case`
- Auth: `Authorization: Bearer <jwt>` (hanya endpoint venue/admin)
- Semua timestamp ISO 8601 dengan timezone (`+07:00`)
- Error format konsisten:

```json
{
  "error": {
    "code": "VENUE_NOT_FOUND",
    "message": "Venue tidak ditemukan",
    "details": null
  }
}
```

### 8.2 Endpoint publik

#### `GET /v1/matches/upcoming`
Daftar laga Persib mendatang.

Query params:
| Param | Tipe | Default | Keterangan |
|---|---|---|---|
| `team_slug` | string | `persib-bandung` | Tim yang diikuti |
| `limit` | int | 5 | Jumlah laga |

Response:
```json
{
  "data": [
    {
      "id": 1042,
      "competition": "Liga 1",
      "home_team": { "name": "Persija Jakarta", "logo_url": "..." },
      "away_team": { "name": "Persib Bandung", "logo_url": "..." },
      "kickoff_at": "2026-09-05T19:00:00+07:00",
      "venue_name": "Stadion Utama GBK",
      "broadcast_tv": "Indosiar",
      "nobar_count": 14
    }
  ]
}
```

#### `GET /v1/matches/{match_id}/nobar`
**Endpoint utama aplikasi.** Daftar venue yang menayangkan laga tertentu.

Query params:
| Param | Tipe | Default | Keterangan |
|---|---|---|---|
| `lat`, `lng` | float | - | Lokasi pengguna |
| `sort` | enum | `recommended` | `recommended` \| `nearest` \| `rating` |
| `radius_km` | float | 15 | Radius pencarian |
| `facilities` | csv | - | `area_anak,parkir_mobil` |
| `entry_type` | enum | - | `free` \| `min_order` \| `ticket` |
| `open_until_end` | bool | true | Hanya venue yang buka sampai laga selesai |
| `page`, `per_page` | int | 1, 20 | Paginasi |

Response:
```json
{
  "data": [
    {
      "event_id": "8f3c...",
      "venue": {
        "id": "a91b...",
        "name": "Sekawan Kopi & Space",
        "slug": "sekawan-kopi-space",
        "district": "Antapani",
        "distance_km": 2.4,
        "primary_photo": "https://...",
        "google_rating": 4.8,
        "nobar_rating": 4.6,
        "nobar_rating_count": 23,
        "kondusif_score": 4.2,
        "kid_friendly_score": 3.8,
        "facilities": ["layar_besar", "outdoor", "parkir_motor", "musala"]
      },
      "doors_open_at": "2026-09-05T18:00:00+07:00",
      "entry_type": "min_order",
      "entry_amount": 25000,
      "crowd_level": "ramai",
      "is_confirmed": true,
      "confirmed_at": "2026-09-04T20:11:00+07:00",
      "is_promoted": false,
      "notes": "Datang sebelum 18.30 biar kebagian tempat depan"
    }
  ],
  "meta": {
    "total": 14,
    "page": 1,
    "per_page": 20,
    "match": {
      "id": 1042,
      "kickoff_at": "2026-09-05T19:00:00+07:00",
      "label": "Persija vs Persib"
    }
  }
}
```

#### `GET /v1/venues/{slug}`
Detail venue lengkap: fasilitas, foto, jam buka, review terbaru, riwayat nobar.

#### `GET /v1/venues/search`
Pencarian venue berdasarkan nama (pakai `pg_trgm`).

#### `POST /v1/reviews`
Kirim review pasca-laga. Tanpa login, pakai `device_hash`.

Request:
```json
{
  "nobar_event_id": "8f3c...",
  "rating_overall": 5,
  "rating_kondusif": 4,
  "is_kid_friendly": true,
  "crowd_actual": "penuh",
  "comment": "Layar jelas, suara kedengeran sampai belakang",
  "device_hash": "sha256:..."
}
```

Validasi:
- Event harus berstatus `FINISHED`
- Hanya boleh dikirim dalam 7 hari setelah kickoff
- Satu `device_hash` satu review per event

#### `POST /v1/events/{id}/track`
Catat interaksi (`open_detail`, `open_maps`, `click_wa`). Fire-and-forget,
tanpa response body. Ini sumber data statistik yang nanti dijual ke venue.

### 8.3 Endpoint venue (butuh auth)

```
POST   /v1/auth/login
POST   /v1/auth/refresh
POST   /v1/venues/claim              # klaim venue unclaimed via OTP WA
GET    /v1/me/venue                  # data venue milik user
PATCH  /v1/me/venue                  # update profil, fasilitas, jam buka
POST   /v1/me/venue/photos
GET    /v1/me/events                 # daftar nobar event milik venue
POST   /v1/me/events                 # umumkan nobar untuk suatu laga
PATCH  /v1/me/events/{id}
POST   /v1/me/events/{id}/confirm    # konfirmasi H-1
DELETE /v1/me/events/{id}
GET    /v1/me/stats                  # statistik view, klik maps, klik WA
```

### 8.4 Endpoint admin

```
GET    /v1/admin/events?status=pending_review
POST   /v1/admin/events/{id}/approve
POST   /v1/admin/events/{id}/reject
POST   /v1/admin/venues              # tambah venue manual
POST   /v1/admin/import/instagram    # submit link IG untuk diekstrak
GET    /v1/admin/import/drafts       # daftar hasil ekstraksi menunggu review
POST   /v1/admin/matches/sync        # tarik ulang jadwal
GET    /v1/admin/reviews?flagged=1
POST   /v1/admin/reviews/{id}/hide
```

### 8.5 Rate limiting

| Endpoint | Limit |
|---|---|
| `GET` publik | 120 req/menit per IP |
| `POST /reviews` | 5 req/jam per IP |
| `POST /track` | 300 req/menit per IP |
| Auth | 10 percobaan login / 15 menit |

---

## 9. Algoritma Pencarian & Rekomendasi

### 9.1 Query geospasial

```sql
-- Cari nobar event untuk suatu laga, dalam radius tertentu
SELECT
    ne.id AS event_id,
    v.id AS venue_id,
    v.name,
    v.slug,
    v.district,
    ST_Distance(v.location, ST_MakePoint($2, $1)::geography) / 1000 AS distance_km,
    v.google_rating,
    v.nobar_rating,
    v.nobar_rating_count,
    v.kondusif_score,
    v.data_completeness,
    ne.doors_open_at,
    ne.entry_type,
    ne.entry_amount,
    ne.crowd_level,
    ne.is_promoted,
    (ne.confirmed_at IS NOT NULL) AS is_confirmed
FROM nobar_event ne
JOIN venue v ON v.id = ne.venue_id
WHERE ne.match_id = $3
  AND ne.status IN ('published', 'confirmed')
  AND v.is_active = TRUE
  AND ST_DWithin(v.location, ST_MakePoint($2, $1)::geography, $4)  -- meter
ORDER BY distance_km ASC
LIMIT $5 OFFSET $6;
```

`ST_DWithin` memakai index GIST, jadi jauh lebih cepat daripada menghitung
`ST_Distance` untuk semua baris lalu memfilter.

### 9.2 Skor rekomendasi

Jangan bikin rumit. Rumus sederhana yang bisa dijelaskan ke venue:

```
skor_akhir = (0.35 × skor_jarak)
           + (0.25 × skor_rating)
           + (0.20 × skor_kondusif)
           + (0.10 × skor_kelengkapan)
           + (0.10 × bonus_konfirmasi)
           + boost_promoted
```

Di mana:

```
skor_jarak       = max(0, 1 - (jarak_km / radius_km))
skor_rating      = (nobar_rating × w) + (google_rating × (1-w))
                   dengan w = min(1, nobar_rating_count / 10)
                   → venue baru pakai rating Google, lama pakai rating nobar
skor_kondusif    = kondusif_score / 5
skor_kelengkapan = data_completeness          -- 0..1
bonus_konfirmasi = 1 kalau confirmed_at != NULL, else 0
boost_promoted   = 0.15 kalau is_promoted     -- jangan lebih, nanti hasilnya sampah
```

Implementasi di Go:

```go
func CalculateScore(v VenueScoreInput, radiusKm float64) float64 {
    distScore := math.Max(0, 1-(v.DistanceKm/radiusKm))

    w := math.Min(1, float64(v.NobarRatingCount)/10.0)
    ratingScore := (v.NobarRating*w + v.GoogleRating*(1-w)) / 5.0

    kondusifScore := v.KondusifScore / 5.0

    confirmBonus := 0.0
    if v.IsConfirmed {
        confirmBonus = 1.0
    }

    score := 0.35*distScore +
        0.25*ratingScore +
        0.20*kondusifScore +
        0.10*v.DataCompleteness +
        0.10*confirmBonus

    if v.IsPromoted {
        score += 0.15
    }
    return score
}
```

### 9.3 Menghitung `data_completeness`

```
kelengkapan = jumlah_field_terisi / total_field_penting
```

Field penting: koordinat, telepon/WA, minimal 1 foto, minimal 3 fasilitas,
jam buka lengkap, IG handle. Skor ini mendorong venue melengkapi profilnya
tanpa perlu dipaksa — karena berpengaruh langsung ke urutan tampil.

### 9.4 Filter jam tutup

Ini filter yang sering dilupakan tapi paling sering menyelamatkan pengguna:

```sql
-- Venue harus masih buka minimal 2 jam setelah kickoff
-- (90 menit pertandingan + 30 menit bubar)
AND (
    (v.opening_hours -> EXTRACT(DOW FROM $kickoff)::text ->> 'close')::time
    > ($kickoff + INTERVAL '2 hours')::time
    OR (v.opening_hours -> EXTRACT(DOW FROM $kickoff)::text ->> 'close') = '24:00'
)
```

Catatan: banyak cafe tutup 22.00–23.00 sementara Liga 1 sering kickoff 19.00
atau lebih malam. Tanpa filter ini, daftar akan penuh venue yang percuma.

### 9.5 Estimasi keramaian tanpa reservasi

Karena tidak ada sistem reservasi di MVP, keramaian diperkirakan dari sinyal
tidak langsung:

```
sinyal = (jumlah open_detail event ini / rata-rata event lain di laga sama)
```

Ditampilkan sebagai label kualitatif, **bukan angka**:

| Rasio | Label |
|---|---|
| < 0.7 | Masih longgar |
| 0.7 – 1.5 | Mulai ramai |
| > 1.5 | Kemungkinan penuh |

Selalu beri disclaimer kecil: "perkiraan berdasarkan minat pengguna aplikasi".
Jangan pernah tampilkan "sisa 4 kursi" kalau datanya tidak nyata.

---

## 10. Fitur Ekstraksi Poster Instagram

### 10.1 Kenapa tidak scraping

- Instagram Graph API resmi hanya bisa membaca akun yang kamu kelola sendiri.
- Endpoint **Business Discovery** bisa membaca akun Business/Creator lain, tapi:
  - butuh App Review Meta (berminggu-minggu) + verifikasi bisnis
  - akun target harus sudah Business/Creator — banyak cafe masih personal
  - kuota per akun per minggu dibatasi, rate limit 200 panggilan/jam
  - **Stories tidak bisa diambil**, padahal info nobar paling sering di situ
- Scraping melanggar ToS dan bisa mematikan aplikasi kapan saja.

### 10.2 Alur yang dipakai

```
[Admin/venue tempel link postingan IG]
              |
              v
[Simpan ke tabel event_source, status: pending_extraction]
              |
              v
[Worker ambil gambar poster]
              |
              v
[Kirim gambar ke model vision + prompt JSON ketat]
              |
              v
[Simpan hasil ke extracted_json]
              |
              v
[Buat nobar_event status DRAFT]
              |
              v
[Admin review 10 detik di panel: koreksi → publish]
```

### 10.3 Prompt ekstraksi

```
Kamu adalah pengekstrak data dari poster acara nonton bareng sepak bola.
Baca gambar poster berikut dan keluarkan HANYA JSON valid, tanpa penjelasan,
tanpa markdown fence.

Skema:
{
  "venue_name": string | null,
  "address_hint": string | null,
  "match_label": string | null,     // "Persib vs Persija"
  "match_date": string | null,      // YYYY-MM-DD
  "doors_open": string | null,      // HH:MM
  "kickoff": string | null,         // HH:MM
  "entry_type": "free" | "min_order" | "ticket" | "donation" | null,
  "entry_amount": number | null,    // rupiah, 0 kalau gratis
  "extras": string[],               // "doorprize", "live music", "bawa jersey diskon"
  "contact": string | null,
  "confidence": number              // 0.0 - 1.0, seberapa yakin kamu
}

Aturan:
- Kalau suatu informasi tidak terbaca jelas, isi null. JANGAN menebak.
- Tahun tidak tertulis di poster → gunakan tahun berjalan.
- Kalau poster bukan tentang nobar sepak bola, kembalikan {"error":"bukan_nobar"}.
```

### 10.4 Aturan wajib

- Hasil ekstraksi **tidak pernah** langsung publish. Selalu lewat review manusia.
- Kalau `confidence < 0.6`, tandai merah di panel admin.
- Simpan `raw_caption` dan `source_url` untuk penelusuran kalau ada komplain.
- Cocokkan `venue_name` hasil ekstraksi ke tabel venue dengan `pg_trgm`
  similarity; kalau > 0.6 tawarkan auto-link, kalau tidak minta admin pilih manual.

### 10.5 Kapan dibangun

**Jangan di awal.** Untuk 30 venue pertama, ketik manual lebih cepat dan lebih
akurat. Bangun fitur ini setelah volume submit melebihi kemampuan input manual —
kira-kira saat sudah 50+ event per laga.

---

## 11. Sistem Review & Skor Kondusif

### 11.1 Kenapa review generik tidak cukup

Rating Google mengukur kualitas kopi dan keramahan pelayan. Untuk nobar,
yang penting sama sekali berbeda: suara kedengaran tidak, layar kelihatan dari
pojok tidak, ramainya kondusif atau rusuh, aman bawa anak atau tidak.

### 11.2 Tiga pertanyaan (jangan lebih)

Muncul otomatis H+1 setelah laga, hanya untuk device yang membuka detail event
tersebut:

1. **Seberapa enak nonton di sini?** (1–5 bintang)
2. **Suasananya kondusif?** (1–5: 1 = rusuh/berisik, 5 = ramai tapi tertib)
3. **Aman kalau bawa anak?** (Ya / Tidak / Tidak tahu)

Opsional: pilih keramaian (sepi / pas / penuh / overload) + komentar 1 kalimat.

Tiga pertanyaan adalah batas atas. Setiap pertanyaan tambahan memotong tingkat
pengisian secara signifikan.

### 11.3 Perhitungan skor

```sql
-- Dijalankan worker setiap malam
UPDATE venue v SET
    nobar_rating = sub.avg_overall,
    nobar_rating_count = sub.cnt,
    kondusif_score = sub.avg_kondusif,
    kid_friendly_score = sub.kid_ratio * 5
FROM (
    SELECT
        venue_id,
        AVG(rating_overall)::numeric(3,2) AS avg_overall,
        AVG(rating_kondusif)::numeric(3,2) AS avg_kondusif,
        COUNT(*) AS cnt,
        AVG(CASE WHEN is_kid_friendly THEN 1.0
                 WHEN is_kid_friendly IS FALSE THEN 0.0
                 ELSE NULL END)::numeric(3,2) AS kid_ratio
    FROM review
    WHERE is_hidden = FALSE
      AND created_at > now() - INTERVAL '18 months'   -- data lama dibuang
    GROUP BY venue_id
) sub
WHERE v.id = sub.venue_id;
```

**Kenapa dibatasi 18 bulan:** venue berubah. Manajemen ganti, layar diganti,
kebijakan merokok berubah. Review 3 tahun lalu menyesatkan.

### 11.4 Ambang tampil

Jangan tampilkan skor kondusif sebelum ada minimal **3 review**. Sebelum itu,
tampilkan "Belum ada penilaian" — lebih jujur dan tidak merusak reputasi venue
karena satu orang yang kebetulan sedang kesal.

### 11.5 Anti-spam tanpa login

- `device_hash` = SHA256(fingerprint browser + salt server)
- Satu device satu review per event (constraint UNIQUE)
- Review hanya bisa dikirim kalau device tersebut pernah `open_detail` event itu
- Rate limit 5 review/jam per IP
- Flag otomatis kalau 3+ review dengan rating ekstrem dari subnet IP yang sama

---

## 12. Notifikasi

### 12.1 Kanal

WhatsApp jauh lebih efektif daripada email di Indonesia. Untuk venue, pakai
WA Business API atau layanan pihak ketiga (Fonnte, Wablas). Untuk pengguna,
mulai dari Web Push (gratis, sudah cukup untuk PWA).

### 12.2 Jadwal notifikasi

| Penerima | Kapan | Isi | Kanal |
|---|---|---|---|
| Venue | H-3 | "Persib main Sabtu. Mau umumkan nobar?" | WA |
| Venue | H-1 pagi | "Konfirmasi nobar besok, tekan link ini" | WA |
| Venue | H-1 malam (belum konfirmasi) | Pengingat terakhir | WA |
| Pengguna | H-1 sore | "Besok Persib main, ada 14 tempat nobar dekat kamu" | Web Push |
| Pengguna | H+1 pagi | "Gimana nobar semalam? Kasih penilaian" | Web Push |
| Admin | Tiap ada submit baru | Notifikasi review antrian | Email/WA |

### 12.3 Implementasi cron

```
0  9  * * *   worker:check-upcoming-matches   # cek laga H-3 dan H-1
0  19 * * *   worker:remind-unconfirmed       # venue belum konfirmasi
0  8  * * *   worker:review-prompt            # H+1 setelah laga
30 2  * * *   worker:recalculate-scores       # hitung ulang rating
0  3  * * 1   worker:sync-match-schedule      # tarik jadwal mingguan
```

---

## 13. Frontend — Struktur & Halaman

### 13.1 Peta halaman

```
/                          Beranda — laga terdekat + daftar venue
/match/[id]                Daftar nobar untuk satu laga
/venue/[slug]              Detail venue
/jadwal                    Semua laga Persib musim ini
/tentang                   Tentang aplikasi
/untuk-venue               Landing page ajakan venue bergabung

/venue-portal/login        Login venue
/venue-portal/dashboard    Dashboard venue
/venue-portal/profil       Edit profil & fasilitas
/venue-portal/events       Kelola pengumuman nobar
/venue-portal/statistik    Statistik kunjungan

/admin/*                   Panel admin
```

### 13.2 Beranda — struktur

```
┌────────────────────────────────────┐
│  NOBARSIB          [ikon lokasi]   │  header
├────────────────────────────────────┤
│  ┌──────────────────────────────┐  │
│  │  LAGA BERIKUTNYA             │  │  kartu match
│  │  Persija  vs  PERSIB         │  │
│  │  Sabtu, 5 Sep · 19:00 WIB    │  │
│  │  Indosiar · 14 tempat nobar  │  │
│  └──────────────────────────────┘  │
├────────────────────────────────────┤
│ [Rekomendasi][Terdekat][Rating]    │  tab sort
│ [Ramah anak][Gratis][Indoor][...]  │  chip filter
├────────────────────────────────────┤
│  ┌──────────────────────────────┐  │
│  │ [foto]  Sekawan Kopi & Space │  │  kartu venue
│  │         Antapani · 2,4 km    │  │
│  │         ★4,6 (23) · Kondusif │  │
│  │         Min order 25rb       │  │
│  │         ✓ Dikonfirmasi       │  │
│  └──────────────────────────────┘  │
│  ┌──────────────────────────────┐  │
│  │ ...                          │  │
└────────────────────────────────────┘
```

### 13.3 Aturan UI

- **Kartu venue maksimal 5 baris informasi.** Lebih dari itu tidak terbaca di HP.
- **Badge "Dikonfirmasi"** harus menonjol — itu janji kepercayaan aplikasi.
- **Jarak selalu tampil**, karena itu alasan utama orang memilih.
- **Jangan tampilkan foto besar** di daftar; thumbnail cukup, foto besar di detail.
- **Tombol utama di detail: "Buka di Maps"** dan **"Chat WA"**. Dua ini yang
  benar-benar dipakai orang.

### 13.4 Halaman detail venue

Urutan blok dari atas:

1. Galeri foto (swipe)
2. Nama, kecamatan, jarak
3. Skor: rating nobar, kondusif, ramah anak
4. Info nobar malam ini (jam buka pintu, biaya, catatan)
5. Tombol aksi: Maps, WA, Instagram
6. Fasilitas (grid ikon)
7. Jam buka per hari
8. Review terbaru (5 teratas)
9. Riwayat nobar sebelumnya di venue ini

### 13.5 Penanganan state kosong

Ini sering dilupakan tapi penting untuk aplikasi baru yang datanya masih tipis:

| Kondisi | Tampilan |
|---|---|
| Belum ada laga mendatang | "Persib belum ada jadwal. Cek jadwal lengkap →" |
| Ada laga, belum ada venue | "Belum ada info nobar untuk laga ini. Punya info? Kirim ke kami" |
| Radius kosong | "Tidak ada nobar dalam 15 km. Perluas radius?" |
| Lokasi ditolak | Fallback ke pusat kota + tombol "Aktifkan lokasi" |

### 13.6 Performa

- Target LCP < 2,5 detik di 4G
- Gambar pakai `next/image` + format WebP
- Daftar venue di-cache Redis 5 menit (kecuali sudah lewat H-1, cache 1 menit)
- Lazy load peta — jangan muat SDK peta di beranda

---

## 14. Panel Admin

### 14.1 Halaman yang dibutuhkan

| Halaman | Fungsi |
|---|---|
| Dashboard | Ringkasan: laga terdekat, event pending, review baru |
| Antrian Event | Approve/reject submit venue & hasil ekstraksi IG |
| Kelola Venue | CRUD venue, isi fasilitas, upload foto |
| Import IG | Tempel link → lihat hasil ekstraksi → koreksi → publish |
| Kelola Jadwal | Tambah/edit laga, sinkronisasi manual |
| Moderasi Review | Sembunyikan review spam/kasar |
| Statistik | Traffic per laga, venue terpopuler, konversi klik Maps |

### 14.2 Prioritas: antrian event

Ini layar yang paling sering kamu buka. Desainnya harus cepat:

```
┌─────────────────────────────────────────────┐
│  ANTRIAN — 7 menunggu review                │
├─────────────────────────────────────────────┤
│ [poster]  Sekawan Kopi & Space              │
│           Persija vs Persib · 5 Sep 19:00   │
│           Buka 18:00 · Min order 25rb       │
│           Sumber: IG @sekawankopi           │
│           Confidence: 0.82                  │
│                                             │
│   [Setujui]  [Edit]  [Tolak]                │
├─────────────────────────────────────────────┤
```

Target: satu event selesai ditinjau dalam 10 detik.

---

## 15. Portal Venue

### 15.1 Prinsip

Pemilik cafe bukan orang teknis dan sedang sibuk. Setiap langkah tambahan
memotong tingkat pemakaian. Aturan: **umumkan nobar harus selesai dalam
3 ketukan.**

### 15.2 Alur ideal

```
[Buka link dari WA]  →  [Lihat laga mendatang]  →  [Tekan "Saya nayangin"]
                                                          |
                                              [Isi 3 field: jam buka,
                                               biaya, catatan]  →  [Kirim]
```

Jangan minta login ulang tiap kali. Pakai magic link dari WA yang berlaku
7 hari.

### 15.3 Konfirmasi H-1 harus satu ketukan

Kirim WA berisi link unik:

```
Halo Sekawan Kopi 👋
Besok Persib vs Persija, kickoff 19:00.
Masih jadi nobar? Tekan: https://nobarsib.id/c/x7Kq2m
(link berlaku 24 jam)
```

Link langsung mengubah status jadi `CONFIRMED` tanpa halaman perantara.
Semakin sedikit friksi, semakin tinggi tingkat konfirmasi, semakin dipercaya
aplikasimu.

### 15.4 Statistik untuk venue

Ini yang nanti membuat mereka mau bayar:

- Berapa orang melihat pengumuman nobar mereka
- Berapa yang membuka detail
- Berapa yang menekan "Buka di Maps" (indikator niat datang)
- Perbandingan dengan rata-rata venue lain di laga yang sama
- Riwayat: laga mana yang paling menarik minat

---

## 16. Seeding Data Awal

### 16.1 Strategi

Jangan sebar ke seluruh Bandung. Ambil satu area dulu — misalnya Sukajadi dan
sekitarnya — kumpulkan 15 venue dengan data lengkap. Satu kecamatan yang penuh
jauh lebih berguna daripada seluruh kota yang kosong.

### 16.2 Kandidat hasil penelusuran Google Places

Sinyal kuat (memang tempat nonton):

| Venue | Area | Catatan |
|---|---|---|
| Tempat Nobar Bandung | Mandalajati | Buka 24 jam, review sebut layar sangat besar |
| Kedai NOBAR PSM 46 | Kiaracondong | Tutup 23:00 — cek untuk laga malam |
| Ludo Sports Kitchen | Batununggal (Kiara Artha) | Sports bar, sarana nobar + live music, buka s/d 03:00 |
| RJ's Sports Bar & Grill | Bandung Wetan | Bar — tidak untuk anak |
| Sekawan Kopi & Space | Antapani | Sering bikin event olahraga, s/d 24:00 |
| Kurito Coffee & Eatery | Cinambo | Sound system bagus, ada private room |
| Barrack Billiard & Cafe | Lengkong | s/d 02:00, keluhan AC saat ramai |
| Grow Billiard and Cafe | Ujung Berung | s/d 02:00 |

Kandidat ramah anak:

| Venue | Area | Catatan |
|---|---|---|
| 150 Coffee Garden | Cicaheum | Halaman rumput luas, banyak review sebut kid friendly |
| Amfiteater Coffee | Sarijadi | Rating 4,9 dari 2.103 ulasan |
| Tigre Coffee and Eatery | Taman Kopo Indah | Family friendly, parkir sulit saat ramai |
| Kala Cemara | Parongpong | Ada area outdoor movie + fasilitas main anak |

Buka sampai larut (cocok laga malam):

| Venue | Area | Tutup |
|---|---|---|
| Jabarano Coffee Angklung 4.0 | Dago | 04:00 |
| Rooftop Coffee & Eatery | Ujung Berung | 03:00 |
| Bober Cafe | Cihapit | 24 jam |
| MMB Cafe | Batununggal | 02:00 (Senin tutup) |

### 16.3 Yang harus kamu verifikasi sendiri

Google Places tidak punya data ini — dan justru inilah nilai tambah aplikasimu:

- [ ] Benar menayangkan **Liga 1** (bukan cuma Liga Inggris/Piala Dunia)
- [ ] Jumlah dan ukuran layar
- [ ] Sound: apakah komentator terdengar atau tertutup musik
- [ ] Suasana: kondusif atau rusuh
- [ ] Kebijakan merokok
- [ ] Apakah ada area terpisah untuk keluarga
- [ ] Minimum order khusus saat nobar
- [ ] Apakah mereka tutup lebih malam khusus malam pertandingan

### 16.4 Template survei lapangan

Bawa ini saat datang ke venue:

```
Nama venue      : ______________________
Tanggal survei  : ______________________
Kontak (nama/WA): ______________________

[ ] Nayangin Liga 1 rutin?      Ya / Kadang / Tidak
[ ] Jumlah layar                ___ buah, ukuran ___ inci/proyektor
[ ] Sound terdengar jelas?      Ya / Sebagian / Tidak
[ ] Kapasitas saat nobar        ± ___ orang
[ ] Biaya masuk                 Gratis / Min order Rp ___ / Tiket Rp ___
[ ] Jam buka pintu nobar        ___:___
[ ] Tutup malam pertandingan    ___:___
[ ] Area non-smoking            Ada / Tidak
[ ] Area/kursi anak             Ada / Tidak
[ ] Musala                      Ada / Tidak
[ ] Parkir mobil                Muat ___ / Tidak ada
[ ] Kondusif menurut pemilik    Rusuh / Ramai tertib / Santai

Pertanyaan kunci ke pemilik:
"Bapak/Ibu repot nggak memperkirakan berapa orang yang bakal datang
 pas malam pertandingan?"
→ Jawaban ya = calon pelanggan pertama.
```

---

## 17. Deployment & Infrastruktur

### 17.1 Lingkungan

| Env | Fungsi | Spesifikasi |
|---|---|---|
| Local | Pengembangan | Docker Compose |
| Staging | Uji sebelum rilis | VPS kecil 1 vCPU / 2 GB |
| Production | Live | VPS 2 vCPU / 4 GB |

### 17.2 docker-compose.yml (development)

```yaml
version: "3.9"

services:
  db:
    image: postgis/postgis:16-3.4
    environment:
      POSTGRES_DB: nobarsib
      POSTGRES_USER: nobarsib
      POSTGRES_PASSWORD: devpassword
    ports: ["5432:5432"]
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U nobarsib"]
      interval: 10s

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    command: redis-server --appendonly yes
    volumes:
      - redisdata:/data

  api:
    build: ./nobarsib-api
    depends_on:
      db: { condition: service_healthy }
    environment:
      DATABASE_URL: postgres://nobarsib:devpassword@db:5432/nobarsib?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev-secret-ganti-di-production
    ports: ["8080:8080"]

  worker:
    build: ./nobarsib-api
    command: ["/app/worker"]
    depends_on: [db, redis]
    environment:
      DATABASE_URL: postgres://nobarsib:devpassword@db:5432/nobarsib?sslmode=disable
      REDIS_URL: redis://redis:6379

  web:
    build: ./nobarsib-web
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080/v1
    ports: ["3000:3000"]

volumes:
  pgdata:
  redisdata:
```

### 17.3 Pipeline CI/CD (GitHub Actions)

```yaml
name: deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgis/postgis:16-3.4
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready --health-interval 10s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go test ./... -race -cover

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy ke VPS
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd /opt/nobarsib
            git pull origin main
            docker compose build
            docker compose run --rm api ./migrate up
            docker compose up -d
            docker image prune -f
```

### 17.4 Backup

```bash
#!/bin/bash
# /opt/nobarsib/scripts/backup.sh — cron harian 02:00
DATE=$(date +%Y%m%d)
docker compose exec -T db pg_dump -U nobarsib nobarsib \
  | gzip > /backup/nobarsib_$DATE.sql.gz

# simpan 14 hari terakhir
find /backup -name "nobarsib_*.sql.gz" -mtime +14 -delete

# sinkron ke object storage
rclone copy /backup/nobarsib_$DATE.sql.gz r2:nobarsib-backup/
```

### 17.5 Monitoring minimal

- **Uptime**: UptimeRobot (gratis) ping `/health` tiap 5 menit
- **Error**: Sentry (free tier cukup)
- **Log**: `docker compose logs` + logrotate, atau Grafana Loki kalau mau rapi
- **Alert**: webhook ke Telegram/WA kalau API down atau error rate naik

Yang paling penting dipantau: **malam pertandingan**. Traffic akan melonjak
tajam 2 jam sebelum kickoff lalu sepi lagi. Pastikan cache siap dan koneksi
database tidak habis.

---

## 18. Estimasi Biaya Operasional

| Komponen | Bulanan | Catatan |
|---|---|---|
| VPS 2 vCPU / 4 GB (Indonesia) | Rp 150.000 – 250.000 | Biznet Gio / IDCloudHost |
| Domain `.id` | ± Rp 20.000 | Rp 250.000/tahun |
| Object storage (foto) | Rp 0 – 30.000 | R2 free tier 10 GB |
| WhatsApp API (Fonnte) | Rp 50.000 – 100.000 | Untuk notifikasi venue |
| Google Maps API | Rp 0 | Kredit gratis $200/bln biasanya cukup |
| Model AI (ekstraksi poster) | Rp 0 – 50.000 | Hanya kalau fitur ini aktif |
| Sentry / monitoring | Rp 0 | Free tier |
| **Total** | **± Rp 220.000 – 450.000** | |

Titik impas: sekitar **2–3 venue berlangganan** sudah menutup biaya server.
Itu target yang masuk akal untuk musim pertama.

---

## 19. Roadmap Bertahap

### Fase 0 — Validasi (2 minggu, tanpa kode)

- [ ] Datangi 8–10 venue kandidat saat malam pertandingan
- [ ] Isi template survei lapangan
- [ ] Tanya pemilik: repot tidak memperkirakan jumlah pengunjung?
- [ ] Catat berapa yang tertarik ikut kalau ada aplikasinya
- [ ] **Gerbang lanjut:** minimal 5 venue bilang tertarik

Kalau tidak lolos gerbang ini, hentikan. Lebih baik kehilangan 2 minggu
daripada 3 bulan.

### Fase 1 — MVP (4–6 minggu)

Backend:
- [ ] Setup project, Docker, migrasi database
- [ ] CRUD venue + fasilitas
- [ ] CRUD match (input manual dulu)
- [ ] CRUD nobar_event
- [ ] Endpoint `/matches/upcoming` dan `/matches/{id}/nobar`
- [ ] Query geospasial + skor rekomendasi
- [ ] Panel admin sederhana

Frontend:
- [ ] Beranda dengan kartu laga + daftar venue
- [ ] Filter & sort
- [ ] Halaman detail venue
- [ ] Deteksi lokasi + fallback
- [ ] PWA manifest + service worker

Data:
- [ ] Isi manual 20–30 venue satu area
- [ ] Input jadwal Persib satu putaran

**Definisi selesai:** kamu sendiri bisa memakainya untuk memutuskan nonton
di mana pada laga berikutnya.

### Fase 2 — Kepercayaan (musim berjalan)

- [ ] Sistem review 3 pertanyaan
- [ ] Perhitungan skor kondusif & ramah anak
- [ ] Portal venue + magic link WA
- [ ] Konfirmasi H-1 satu ketukan
- [ ] Badge "Dikonfirmasi"
- [ ] Web Push notification
- [ ] Statistik dasar untuk venue

### Fase 3 — Skala (musim kedua)

- [ ] Ekstraksi poster Instagram
- [ ] Perluas ke seluruh Bandung Raya
- [ ] Tambah Timnas Indonesia (ekspansi paling alami)
- [ ] Langganan venue + sistem promoted
- [ ] Statistik lanjutan

### Fase 4 — Opsional

- [ ] Reservasi meja (baru masuk akal kalau venue sudah disiplin pakai portal)
- [ ] Aplikasi native
- [ ] Kota lain (Surabaya untuk Persebaya, Malang untuk Arema)

---

## 20. Metrik Keberhasilan

### 20.1 Metrik utama per laga

| Metrik | Target musim 1 | Cara ukur |
|---|---|---|
| Pengguna unik malam laga | 500 → 2.000 | Device hash unik |
| Venue tayang per laga | 15 → 40 | Count nobar_event published |
| Tingkat konfirmasi venue H-1 | > 70% | confirmed / published |
| Klik "Buka di Maps" | > 30% dari pembuka detail | event_view |
| Review terkumpul per laga | > 20 | Count review |
| Venue dengan data lengkap | > 60% | data_completeness > 0.8 |

### 20.2 Metrik yang menentukan kelanjutan

Setelah satu musim penuh, tanya:

1. Apakah pengguna kembali di laga berikutnya? (retensi antar-laga)
2. Apakah ada venue yang minta pasang lebih menonjol? (sinyal willingness to pay)
3. Apakah ada yang membagikan link aplikasi tanpa diminta?

Kalau ketiganya "tidak", produk ini tetap portofolio yang bagus tapi bukan
bisnis. Itu bukan kegagalan — itu informasi.

---

## 21. Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| **Data basi** — venue tidak jadi nobar tapi masih tampil | Fatal. Ini yang membunuh kepercayaan FANZO | Konfirmasi H-1 wajib; tanpa konfirmasi turunkan peringkat + badge peringatan |
| **Ayam-telur marketplace** | Kedua sisi sepi, aplikasi mati | Isi manual dulu, fokus satu kecamatan |
| **Frekuensi pakai rendah** (15–20 malam/tahun) | Aplikasi dilupakan | PWA bukan native; notifikasi H-1; tambah Timnas |
| **Instagram menutup akses** | Fitur impor mati | Jangan jadikan pondasi; manual selalu jadi fallback |
| **Venue tidak mau bayar** | Tidak ada pendapatan | Jangan bangun sistem bayar sebelum traffic terbukti |
| **Persib pindah jam tayang / jadwal kacau** | Data salah | Sinkronisasi jadwal + verifikasi manual sebelum H-3 |
| **Muncul pesaing besar** (Goers, Vidio) | Kehilangan pasar | Keunggulan kamu: data kualitatif (kondusif, ramah anak) yang butuh waktu 2–3 musim dikumpulkan |
| **Kamu kehabisan waktu** | Proyek mangkrak | Batasi MVP; jangan bangun reservasi/pembayaran di awal |

---

## 22. Aspek Legal

Beberapa hal yang perlu diperhatikan:

- **Jangan menyediakan streaming.** Aplikasi hanya menampilkan informasi lokasi
  nobar. Menautkan ke sumber ilegal bisa menyeret tanggung jawab hukum.
- **Hak siar.** Venue yang menayangkan siaran secara komersial idealnya punya
  izin dari pemegang hak siar. Ini tanggung jawab venue, bukan kamu, tapi
  cantumkan disclaimer bahwa aplikasi tidak memverifikasi izin siar venue.
- **Logo dan merek Persib.** Jangan pakai logo klub tanpa izin, terutama kalau
  aplikasi menghasilkan uang. Pilih nama dan identitas visual sendiri.
- **Data pribadi.** UU PDP berlaku. Simpan seminimal mungkin: untuk pengguna
  cukup device hash, tidak perlu nama atau email. Buat halaman kebijakan privasi.
- **Foto venue.** Kalau memakai foto dari Google Places, ikuti syarat atribusi
  Google. Lebih aman minta foto langsung dari venue.
- **Konten review.** Sediakan mekanisme lapor dan hapus untuk menghindari
  masalah pencemaran nama baik.

---

## 23. Checklist Sebelum Rilis

### Teknis
- [ ] Semua endpoint punya test minimal happy path
- [ ] Rate limiting aktif di semua endpoint publik
- [ ] HTTPS dengan sertifikat valid
- [ ] Backup otomatis berjalan dan **sudah diuji restore**
- [ ] Monitoring uptime aktif
- [ ] Error tracking terpasang
- [ ] Environment variable produksi tidak ada nilai default development
- [ ] Query lambat sudah di-index (jalankan `EXPLAIN ANALYZE` pada query utama)
- [ ] Uji beban ringan: simulasi 500 request bersamaan

### Konten
- [ ] Minimal 20 venue dengan data lengkap
- [ ] Jadwal Persib satu putaran sudah masuk
- [ ] Semua venue punya minimal 1 foto
- [ ] Halaman kebijakan privasi & syarat penggunaan
- [ ] Halaman "untuk venue" berisi ajakan bergabung

### UX
- [ ] Diuji di HP Android kelas menengah, bukan cuma di laptop
- [ ] Diuji di jaringan lambat (throttle 3G di DevTools)
- [ ] Semua state kosong punya tampilan yang jelas
- [ ] Tombol Maps dan WA berfungsi di iOS dan Android
- [ ] PWA bisa di-"Add to Home Screen"

### Operasional
- [ ] Nomor WA khusus untuk komunikasi venue
- [ ] Template pesan WA H-3 dan H-1 sudah disiapkan
- [ ] Kamu tahu apa yang akan dilakukan kalau ada venue komplain
- [ ] Rencana peluncuran: posting di grup bobotoh mana, kapan

---

## Penutup

Satu hal yang perlu diingat sepanjang mengerjakan ini: bagian tersulit dari
proyek ini bukan kodenya. Query geospasial, skema database, dan API di dokumen
ini bisa selesai dalam beberapa minggu.

Yang sulit adalah **menjaga data tetap benar** setiap malam pertandingan, dan
**meyakinkan venue** untuk repot sedikit mengisi informasi. FANZO punya dana,
tim, dan 12 tahun pengalaman, dan masih gagal di titik itu.

Jadi kalau ada satu bagian yang harus dikerjakan paling serius, itu adalah
mekanisme konfirmasi H-1 dan kemudahan portal venue. Sisanya hanya pendukung.

Mulai dari Fase 0. Datangi venue dulu sebelum menulis baris kode pertama.
