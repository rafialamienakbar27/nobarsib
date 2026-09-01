#!/bin/bash
#
# Membangun ulang APK dan memasangnya ke situs.
#
# Menghasilkan tiga hal sekaligus, karena ketiganya harus selalu sejalan:
#   1. nobarsib-android/app-release-signed.apk   — hasil build
#   2. nobarsib-web/public/nobarsib.apk          — berkas yang diunduh pengguna
#   3. nobarsib-web/lib/site.ts                  — ukuran & tanggal di halaman /unduh
#
# Menyalin APK tanpa memperbarui site.ts akan membuat halaman unduh menyebut
# ukuran berkas yang salah — kecil, tapi persis jenis ketidakcocokan yang
# membuat orang ragu memasang berkas dari luar Play Store.
#
# Pemakaian:  ./build.sh [versi]
#   ./build.sh          -> naikkan versionCode, pakai versionName yang ada
#   ./build.sh 1.1.0    -> naikkan versionCode dan setel versionName ke 1.1.0
set -euo pipefail

cd "$(dirname "$0")"
AKAR="$(cd .. && pwd)"
WEB="$AKAR/nobarsib-web"

# --- toolchain --------------------------------------------------------------
# Dipasang oleh setup pertama ke folder home, bukan lewat Homebrew: keduanya
# tidak butuh sudo dan tidak bertabrakan dengan Java lain di mesin ini
# (mesin ini masih punya Java 8 bawaan, yang terlalu tua untuk Android).
export JAVA_HOME="${JAVA_HOME:-$HOME/.nobarsib-toolchain/jdk-17/Contents/Home}"
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
export ANDROID_SDK_ROOT="$ANDROID_HOME"

if [ ! -x "$JAVA_HOME/bin/java" ]; then
  echo "JDK 17 tidak ditemukan di $JAVA_HOME — jalankan setup-android.sh dulu." >&2
  exit 1
fi

# --- kata sandi keystore ----------------------------------------------------
# Dibaca dari signing.env supaya tidak pernah ikut tertulis di riwayat perintah.
if [ -f signing.env ]; then
  # shellcheck disable=SC1091
  . ./signing.env
fi
if [ -z "${BUBBLEWRAP_KEYSTORE_PASSWORD:-}" ]; then
  echo "BUBBLEWRAP_KEYSTORE_PASSWORD belum diisi — lihat signing.env.contoh" >&2
  exit 1
fi
export BUBBLEWRAP_KEYSTORE_PASSWORD
export BUBBLEWRAP_KEY_PASSWORD="${BUBBLEWRAP_KEY_PASSWORD:-$BUBBLEWRAP_KEYSTORE_PASSWORD}"

# --- versi ------------------------------------------------------------------
# versionCode WAJIB naik setiap rilis. Android menolak memasang APK dengan
# versionCode lebih rendah atau sama di atas yang sudah terpasang — gejalanya
# menyesatkan: "aplikasi tidak terpasang", tanpa alasan.
LAMA=$(grep -o '"appVersionCode": *[0-9]*' twa-manifest.json | grep -o '[0-9]*')
BARU=$((LAMA + 1))
sed -i '' "s/\"appVersionCode\": *$LAMA/\"appVersionCode\": $BARU/" twa-manifest.json

if [ $# -ge 1 ]; then
  sed -i '' "s/\"appVersionName\": *\"[^\"]*\"/\"appVersionName\": \"$1\"/" twa-manifest.json
fi
VERSI=$(grep -o '"appVersionName": *"[^"]*"' twa-manifest.json | sed 's/.*"appVersionName": *"//;s/"//')

echo "==> Membangun NOBARSIB $VERSI (versionCode $BARU)"

# Regenerasi proyek WAJIB, tidak bisa dilewati.
#
# versionCode yang dipakai saat kompilasi dibaca dari app/build.gradle, bukan
# dari twa-manifest.json. Menaikkan angka di manifest lalu langsung build hanya
# menghasilkan APK dengan versionCode lama — dan `bubblewrap build` akan berhenti
# menunggu jawaban di terminal karena checksum manifestnya tidak lagi cocok.
MANIFEST_URL=$(grep -o '"webManifestUrl": *"[^"]*"' twa-manifest.json | sed 's/.*: *"//;s/"//')
if ! curl -sfI --max-time 10 "$MANIFEST_URL" >/dev/null 2>&1; then
  echo >&2
  echo "Manifest PWA tidak bisa diambil dari $MANIFEST_URL" >&2
  echo >&2
  echo "Regenerasi proyek Android membaca ikon dan warna dari alamat itu." >&2
  case "$MANIFEST_URL" in
    http://localhost*)
      echo "Jalankan 'make web' di terminal lain, lalu ulangi." >&2 ;;
    *)
      echo "Pastikan situsnya sudah online, atau arahkan sementara" >&2
      echo "webManifestUrl dan iconUrl di twa-manifest.json ke localhost:3000." >&2 ;;
  esac
  # Kembalikan versionCode supaya percobaan berikutnya tidak melompat dua angka.
  sed -i '' "s/\"appVersionCode\": *$BARU/\"appVersionCode\": $LAMA/" twa-manifest.json
  exit 1
fi

bubblewrap update --skipVersionUpgrade

# --skipPwaValidation: pemeriksaan itu menjalankan Lighthouse terhadap situs
# produksi. Berguna sebelum masuk Play Store, tapi di sini ia hanya membuat
# build gagal setiap kali jaringan lambat.
bubblewrap build --skipPwaValidation

# --- pasang ke situs --------------------------------------------------------
cp app-release-signed.apk "$WEB/public/nobarsib.apk"

BYTES=$(stat -f%z "$WEB/public/nobarsib.apk")
UKURAN=$(python3 -c "print(f'{$BYTES/1048576:.1f}'.replace('.', ','))")
TANGGAL=$(date +%Y-%m-%d)

sed -i '' "s/  versi: \"[^\"]*\"/  versi: \"$VERSI\"/" "$WEB/lib/site.ts"
sed -i '' "s/  ukuran: \"[^\"]*\"/  ukuran: \"$UKURAN MB\"/" "$WEB/lib/site.ts"
sed -i '' "s/  dibangun: \"[^\"]*\"/  dibangun: \"$TANGGAL\"/" "$WEB/lib/site.ts"

echo "==> Selesai: $UKURAN MB → nobarsib-web/public/nobarsib.apk"
echo "    Halaman /unduh sudah menyebut versi $VERSI."
