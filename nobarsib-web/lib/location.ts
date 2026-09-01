"use client";

export { BANDUNG } from "./geo";

/**
 * Hasil permintaan lokasi.
 *
 * "denied" dan "gagal" DIPISAH, dan pemisahan itu adalah inti berkas ini.
 * Keduanya sama-sama berarti "tidak ada koordinat", tapi artinya bagi pengguna
 * bertolak belakang: yang satu keputusan yang harus dihormati, yang satu lagi
 * kecelakaan teknis yang justru harus dicoba lagi.
 */
export type LocationState =
  | { status: "granted"; lat: number; lng: number }
  | { status: "denied" } // pengguna menolak izinnya
  | { status: "gagal" }; // timeout atau posisi tidak tersedia

/**
 * Keadaan izin menurut peramban.
 *
 * "unknown" bukan kegagalan: sebagian peramban (Safari, terutama di iOS) tidak
 * mengizinkan status geolokasi ditanyakan sebelum diminta. Di situ satu-satunya
 * cara mengetahuinya memang dengan mencoba.
 */
export type IzinLokasi = "granted" | "prompt" | "denied" | "unknown";

const KEY = "nobarsib:lokasi-ditolak";

/**
 * Pengguna yang pernah menolak tidak ditanya lagi setiap kali membuka aplikasi.
 *
 * Dengan pola pemakaian 15–20 malam setahun (§1), dialog izin yang muncul
 * berulang justru terasa seperti aplikasi yang tidak ingat apa-apa.
 */
export function pernahMenolak(): boolean {
  try {
    return localStorage.getItem(KEY) === "1";
  } catch {
    return false; // mode penyamaran atau penyimpanan diblokir
  }
}

export function catatPenolakan(): void {
  try {
    localStorage.setItem(KEY, "1");
  } catch {
    // Tidak apa-apa: paling banter pengguna ditanya lagi lain kali.
  }
}

export function lupakanPenolakan(): void {
  try {
    localStorage.removeItem(KEY);
  } catch {
    // idem
  }
}

/**
 * Membaca keadaan izin tanpa memicu dialog apa pun.
 *
 * Gunanya satu: membedakan "belum pernah ditanya" dari "sudah diblokir".
 * Tanpa ini, tombol "Aktifkan lokasi" pada perangkat yang izinnya sudah
 * diblokir permanen akan memanggil getCurrentPosition, ditolak seketika tanpa
 * dialog, dan dari sisi pengguna tombolnya tampak rusak — ditekan berkali-kali
 * pun tidak terjadi apa-apa.
 */
export async function statusIzin(): Promise<IzinLokasi> {
  try {
    if (typeof navigator === "undefined" || !navigator.permissions) return "unknown";
    const izin = await navigator.permissions.query({ name: "geolocation" });
    return izin.state;
  } catch {
    // Peramban yang tidak mendukung kueri geolokasi melempar di sini.
    return "unknown";
  }
}

/**
 * Memberi tahu saat pengguna mengubah izin di setelan peramban, selagi halaman
 * terbuka. Mengembalikan fungsi pembatalan.
 *
 * Ini yang membuat petunjuk "buka setelan lalu izinkan" berakhir dengan daftar
 * yang menyusun ulang dirinya sendiri, bukan dengan kalimat "sekarang muat
 * ulang halaman ini" — langkah yang paling sering membuat orang menyerah.
 */
export function pantauIzin(saatBerubah: (izin: IzinLokasi) => void): () => void {
  let status: PermissionStatus | null = null;
  const teruskan = () => status && saatBerubah(status.state);

  void (async () => {
    try {
      if (typeof navigator === "undefined" || !navigator.permissions) return;
      status = await navigator.permissions.query({ name: "geolocation" });
      status.addEventListener("change", teruskan);
    } catch {
      // Tidak didukung — tidak ada yang perlu dipantau.
    }
  })();

  return () => status?.removeEventListener("change", teruskan);
}

/**
 * Meminta lokasi pengguna.
 *
 * Selalu selesai — tidak pernah menolak promise — karena pemanggilnya harus
 * tetap menampilkan daftar venue apa pun hasilnya (§13.5).
 */
export function mintaLokasi(): Promise<LocationState> {
  return new Promise((resolve) => {
    if (typeof navigator === "undefined" || !navigator.geolocation) {
      resolve({ status: "denied" });
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) =>
        resolve({
          status: "granted",
          lat: pos.coords.latitude,
          lng: pos.coords.longitude,
        }),
      (err) => {
        // Hanya PERMISSION_DENIED yang boleh dicatat sebagai penolakan.
        //
        // Sebelumnya SEMUA kegagalan dicatat, dan akibatnya sunyi tapi
        // permanen: satu kali timeout 8 detik di dalam gedung bersinyal lemah
        // cukup untuk membuat aplikasi berhenti meminta lokasi SELAMANYA di
        // perangkat itu — padahal orangnya tidak pernah menolak apa pun. Dua
        // kode lainnya (POSITION_UNAVAILABLE, TIMEOUT) adalah kecelakaan
        // teknis yang wajar dicoba lagi lain kali.
        if (err.code === err.PERMISSION_DENIED) {
          catatPenolakan();
          resolve({ status: "denied" });
          return;
        }
        resolve({ status: "gagal" });
      },
      {
        // Ketelitian tinggi tidak sepadan di sini: yang dibutuhkan hanya
        // urutan jarak antar venue, dan GPS presisi menguras baterai serta
        // memperlambat tampilan pertama.
        enableHighAccuracy: false,
        timeout: 8000,
        maximumAge: 5 * 60 * 1000,
      },
    );
  });
}
