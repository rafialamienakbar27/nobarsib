# NOBARSIB untuk Android

APK yang diunduh pengguna dari halaman `/unduh`.

## Apa isinya

Ini **bukan** aplikasi terpisah. Isinya Trusted Web Activity (TWA): pembungkus
setebal ±950 KB yang membuka `https://nobarsib.vercel.app` layar penuh, tanpa
bilah alamat, memakai mesin Chrome yang sudah ada di HP.

Konsekuensinya, dan semuanya disengaja:

- **Isi aplikasi selalu ikut situs.** Perbaikan di situs langsung sampai ke
  semua pengguna tanpa mereka memasang apa pun. APK hanya perlu dibangun ulang
  kalau ikon, nama, warna, atau alamat situsnya berubah.
- **Tetap butuh internet.** Sama seperti situsnya (lihat komentar di
  `nobarsib-web/public/sw.js` soal kenapa data nobar tidak disimpan luring).
- **Ukurannya kecil** karena tidak membawa mesin peramban sendiri.

## Membangun ulang

```bash
./build.sh          # naikkan versionCode, versi tetap
./build.sh 1.1.0    # sekaligus setel versionName
```

Skrip ini membangun APK, menyalinnya ke `nobarsib-web/public/nobarsib.apk`,
dan memperbarui ukuran serta versi di `nobarsib-web/lib/site.ts` supaya halaman
`/unduh` tidak menyebut angka yang basi.

Kata sandi keystore dibaca dari `signing.env` (tidak masuk git).

### Prasyarat

`bubblewrap` (npm global), JDK 17 di `~/.nobarsib-toolchain/jdk-17`, dan Android
SDK di `~/Library/Android/sdk` dengan `build-tools;36.1.0`. Semuanya sudah
terpasang di mesin ini. Untuk mesin baru, `bubblewrap doctor` akan menunjukkan
apa yang kurang.

Catatan: Bubblewrap mencari `sdkmanager` di `$ANDROID_HOME/bin`, bukan di
`cmdline-tools/latest/bin` seperti tata letak SDK sekarang — folder `bin/` dan
`lib/` di akar SDK adalah salinan untuk memenuhi itu.

## Yang WAJIB dijaga

`android.keystore` + kata sandinya. Keduanya menentukan identitas aplikasi:

- **Hilang** → kamu tidak bisa lagi menerbitkan pembaruan. Semua pengguna harus
  mencopot aplikasi lalu memasang ulang dari nol.
- **Bocor** → orang lain bisa membuat APK yang dipercaya penuh oleh HP yang
  sudah memasang aplikasi ini.

Salin keduanya ke tempat aman di luar mesin ini (password manager, atau drive
terenkripsi). Ini satu-satunya berkas di proyek ini yang tidak bisa dibuat ulang.

## Kalau ganti domain

Alamat situs ikut ditandatangani ke dalam APK, jadi tiga tempat harus diubah
bersamaan — kalau salah satu tertinggal, aplikasi tetap terbuka tapi dengan
bilah alamat peramban di atasnya, karena verifikasi Digital Asset Links gagal:

1. `twa-manifest.json` → `host` dan `fullScopeUrl`
2. `nobarsib-web/lib/site.ts` → `SITE_URL` (atau env `NEXT_PUBLIC_SITE_URL`)
3. Pastikan `https://<domain>/.well-known/assetlinks.json` bisa diakses publik
   dan berisi sidik jari di bawah

Lalu `bubblewrap update && ./build.sh`.

### Sidik jari sertifikat

```
55:2A:76:2F:A5:19:24:27:85:E2:F2:66:38:E2:9E:C4:4E:B5:7F:76:1C:73:67:20:38:B0:2D:D4:98:7E:80:27
```

Sudah tertulis di `nobarsib-web/public/.well-known/assetlinks.json`. Nilainya
ikut berubah kalau keystore diganti — ambil ulang dengan:

```bash
keytool -list -v -keystore android.keystore -alias nobarsib
```

## Kenapa `webManifestUrl` menunjuk ke localhost

`webManifestUrl` dan `iconUrl` di `twa-manifest.json` adalah **bahan build**:
dibaca sekali saat proyek Android dibuat, untuk mengambil ikon dan warna. Yang
menentukan perilaku aplikasi saat dijalankan adalah `host`, `startUrl`, dan
`fullScopeUrl` — dan ketiganya sudah menunjuk ke produksi.

Keduanya dibiarkan menunjuk ke `http://localhost:3000` supaya `bubblewrap
update` bisa dijalankan kapan saja dengan `npm run dev` menyala, termasuk
sebelum situsnya pernah di-deploy. Setelah produksi hidup, keduanya boleh
diarahkan ke domain sungguhan — tapi tidak ada yang rusak kalau tidak.

## Play Store

Berkas `app-release-bundle.aab` yang ikut dihasilkan adalah format yang diminta
Play Store, kalau nanti diperlukan. Belum diperlukan sekarang: distribusi lewat
tautan langsung tidak menunggu review, dan Play Store tidak menambah satu pun
kemampuan pada aplikasi ini.
