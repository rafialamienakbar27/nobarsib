# Deploy NOBARSIB ke Vercel

Vercel hanya menjalankan `nobarsib-web` (Next.js). API Go, Postgres, dan Redis
tetap butuh host sendiri — lihat [bagian API](#api-go) di bawah.

---

## 1. Siapkan API dulu, bukan situsnya

Urutan ini penting. Situs yang online tanpa API menampilkan halaman "Tidak bisa
memuat data" di semua halaman, dan itu bukan keadaan yang enak untuk dites.

Kabar baiknya: **build-nya tidak akan gagal** kalau API belum siap. Sudah diuji
dengan API sengaja diarahkan ke port mati — `next build` tetap keluar dengan
kode 0. Semua halaman yang memanggil API (`/`, `/jadwal`, `/match/[id]`,
`/venue/[slug]`) dirender saat permintaan masuk, bukan saat build, karena
`cache: "no-store"` di [lib/api.ts](nobarsib-web/lib/api.ts#L45). Jadi deploy
boleh duluan; yang tidak boleh adalah menganggapnya selesai sebelum API hidup.

API harus bisa diakses lewat **HTTPS**, bukan HTTP. Alasannya bukan gaya-gayaan:
situs di Vercel dilayani lewat HTTPS, dan peramban **memblokir** permintaan HTTP
dari halaman HTTPS (mixed content). Gejalanya menyesatkan — halaman terbuka
normal, daftar venue kosong, tidak ada pesan error, dan penyebabnya hanya
terlihat di konsol peramban.

Jadi API di VPS perlu domain + sertifikat (Nginx + Let's Encrypt, sesuai §6.3
blueprint). Misalnya `https://api.nobarsib.id/v1`.

---

## 2. Deploy situsnya

Proyek ini **belum berupa repositori git**, jadi ada dua jalan.

### Jalan cepat — CLI (tidak perlu GitHub)

```bash
cd nobarsib-web
npx vercel            # deploy pratinjau, sekaligus membuat proyeknya
npx vercel --prod     # deploy produksi
```

Saat ditanya, jawab:

| Pertanyaan | Jawaban |
|---|---|
| Set up and deploy? | `y` |
| Which scope? | akun kamu |
| Link to existing project? | `n` |
| Project name? | `nobarsib` |
| In which directory is your code located? | `./` (kamu sudah di dalam `nobarsib-web`) |

Perintah dijalankan **dari dalam `nobarsib-web`**, jadi tidak perlu mengatur
Root Directory sama sekali.

### Jalan panjang — lewat GitHub

Kalau ingin setiap `git push` otomatis men-deploy:

```bash
cd /path/ke/nobarsib          # akar proyek, bukan nobarsib-web
git init && git add . && git commit -m "NOBARSIB"
gh repo create nobarsib --private --source=. --push
```

Lalu di dashboard Vercel: **Add New → Project → pilih repo**, dan yang paling
mudah terlewat:

> **Root Directory** harus diisi `nobarsib-web`.

Tanpa itu Vercel membangun dari akar repo, tidak menemukan `package.json`, dan
gagal. Di akar ada `nobarsib-api` (Go) dan `nobarsib-android` yang bukan bagian
dari build Next.js.

Sebelum `git init`, pastikan `nobarsib-android/.gitignore` sudah ada — di
dalamnya `android.keystore` dan `signing.env`, dua berkas yang tidak boleh
masuk repositori. (Sudah dibuat, tinggal dipastikan ikut terbaca.)

---

## 3. Environment variables

Isi di **Project Settings → Environment Variables**, untuk environment
Production (dan Preview kalau mau).

| Nama | Nilai | Kenapa |
|---|---|---|
| `API_URL` | `https://api.nobarsib.id/v1` | Dipakai server Next saat merender halaman |
| `NEXT_PUBLIC_API_URL` | `https://api.nobarsib.id/v1` | Dipakai peramban pengguna — **wajib HTTPS** |
| `NEXT_PUBLIC_SITE_URL` | `https://nobarsib.vercel.app` | Harus sama persis dengan alamat yang dipanggang ke APK |
| `NEXT_PUBLIC_ADMIN_WA` | nomor WA admin, mis. `628xxxxxxxxxx` | Tombol daftar di `/untuk-venue`; boleh dikosongkan |

Keduanya (`API_URL` dan `NEXT_PUBLIC_API_URL`) sengaja dipisah dan biasanya
diisi sama di Vercel. Pemisahannya berguna di Docker, di mana server memanggil
`http://api:8080/v1` sementara peramban tidak bisa — penjelasannya ada di
[lib/api.ts](nobarsib-web/lib/api.ts).

---

## 4. Apa isi `vercel.json`

Sudah disiapkan di `nobarsib-web/vercel.json`, tidak perlu disentuh. Tiga hal:

- **`regions: ["sin1"]`** — fungsi server dijalankan di Singapura, bukan bawaan
  Vercel di Amerika. Untuk pengguna di Bandung bedanya ratusan milidetik pada
  setiap permintaan, dan API-nya juga akan ada di Indonesia.
  *Kalau deploy ditolak dengan keluhan soal `regions`* (paket Hobby hanya boleh
  satu region dan pernah membatasinya lewat berkas): hapus baris itu, lalu setel
  dari **Project Settings → Functions → Function Region**.
- **Tipe berkas APK** — dipaksa ke `application/vnd.android.package-archive`
  dengan `Content-Disposition: attachment`, supaya Android mengenalinya sebagai
  pemasang aplikasi dan bukan berkas asing untuk dibuka entah dengan apa.
  `Cache-Control` sengaja nol: APK diganti setiap `make apk`, dan pengguna yang
  mengunduh versi lama dari cache CDN akan gagal memasang tanpa penjelasan.
- **`assetlinks.json`** — dipastikan dilayani sebagai `application/json`.
  Verifikasi Digital Asset Links menolak tipe konten lain, dan kalau gagal,
  aplikasi Android tetap terbuka tapi dengan bilah alamat peramban di atasnya.

## 5. Setelah deploy pertama — periksa tiga hal ini

```bash
SITUS=https://nobarsib.vercel.app   # ganti dengan alamat yang kamu dapat

# 1. Situsnya hidup dan datanya masuk
curl -s "$SITUS" | grep -c "tempat nobar"      # harus > 0

# 2. Berkas verifikasi aplikasi Android terlayani
curl -s "$SITUS/.well-known/assetlinks.json"   # harus JSON, bukan halaman 404

# 3. APK bisa diunduh dengan tipe yang benar
curl -sI "$SITUS/nobarsib.apk" | grep -i "content-type"
# harus: application/vnd.android.package-archive
```

---

## 6. Kalau alamatnya BUKAN `nobarsib.vercel.app`

Ini kemungkinan yang nyata: nama `nobarsib` bisa saja sudah dipakai orang lain,
dan Vercel memberi `nobarsib-xxxx.vercel.app`.

APK yang sudah dibangun menandatangani `nobarsib.vercel.app` ke dalam dirinya.
Kalau alamat aslinya berbeda, aplikasinya tetap terbuka tapi **dengan bilah
alamat peramban di atasnya**, karena verifikasi Digital Asset Links gagal.

Perbaikannya tiga langkah:

```bash
# 1. Ubah host di manifest TWA
#    nobarsib-android/twa-manifest.json → "host" dan "fullScopeUrl"

# 2. Ubah alamat situs
#    Vercel env NEXT_PUBLIC_SITE_URL, dan default di nobarsib-web/lib/site.ts

# 3. Bangun ulang, lalu deploy lagi
make apk
```

Sidik jari sertifikat di `nobarsib-web/public/.well-known/assetlinks.json`
**tidak berubah** — yang berubah hanya alamat yang menampungnya.

---

## API Go

Vercel tidak menjalankan ini. Yang dibutuhkan, sesuai §17.1 blueprint: VPS
Indonesia (Biznet/IDCloudHost) dengan Postgres+PostGIS, Redis, `cmd/api`,
`cmd/worker`, dan Nginx di depan untuk TLS.

`docker-compose.yml` di akar sudah menjalankan kelimanya untuk pengembangan.
Untuk produksi yang berubah: `POSTGRES_PASSWORD` yang sungguhan, `JWT_SECRET`
dan `DEVICE_HASH_SALT` yang bukan bawaan, serta Nginx + Let's Encrypt di depan
port 8080.

Satu setelan yang paling mudah terlewat dan gejalanya membingungkan:

```env
CORS_ORIGINS=https://nobarsib.vercel.app
```

Middleware CORS di [cmd/api/main.go](nobarsib-api/cmd/api/main.go#L86) **hanya
dipasang kalau daftar ini terisi**. Kalau dibiarkan kosong, API tetap sehat dan
`curl` tetap menjawab normal — tapi setiap permintaan dari peramban ditolak,
dan halaman hanya menampilkan "Gagal memuat daftar" tanpa petunjuk apa pun.
