/**
 * Identitas publik situs dan berkas aplikasi Android.
 *
 * Dipisah ke satu berkas karena tiga tempat harus selalu sepakat soal alamat
 * yang sama: halaman unduh, berkas assetlinks.json, dan URL yang dipanggang ke
 * dalam APK. Kalau ketiganya berbeda — walau hanya beda "www" — aplikasi
 * Android akan terbuka dengan bilah alamat peramban di atasnya, karena
 * verifikasi Digital Asset Links gagal diam-diam.
 */

/**
 * Alamat situs di produksi.
 *
 * Diambil dari env supaya pindah domain tidak perlu menyentuh kode; nilai
 * bawaannya adalah alamat yang dipakai APK saat ini. Mengubah ini WAJIB diikuti
 * membangun ulang APK (`make apk`), karena alamatnya ikut ditandatangani.
 */
export const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "https://nobarsib.vercel.app";

/**
 * Keterangan berkas APK yang sedang ditawarkan di /unduh.
 *
 * Nilai-nilai ini DITULIS ULANG OTOMATIS oleh nobarsib-android/build.sh setiap
 * kali APK dibangun — jangan disunting tangan, suntingannya akan tertimpa.
 * Ukuran dan versi sengaja disimpan sebagai data, bukan dibaca dari disk saat
 * permintaan masuk: di hosting tanpa server (Vercel) berkas di public/ dilayani
 * CDN dan tidak selalu ada di sistem berkas fungsi Node-nya.
 */
export const APK: {
  nama: string;
  versi: string;
  ukuran: string;
  dibangun: string;
  androidMinimal: string;
} = {
  /** Nama berkas di public/. */
  nama: "nobarsib.apk",
  /** Versi yang tampil ke pengguna. */
  versi: "1.0.0",
  /** Ukuran berkas, sudah diformat untuk dibaca manusia. */
  ukuran: "0,9 MB",
  /** Tanggal build, format ISO. Kosong berarti APK belum pernah dibangun. */
  dibangun: "2026-09-01",
  /** Android minimal yang didukung TWA. */
  androidMinimal: "Android 8.0",
};

/** True kalau APK-nya sudah ada dan boleh ditawarkan. */
export const apkTersedia = APK.dibangun !== "";

/** URL unduhan APK. */
export const apkUrl = `/${APK.nama}`;
