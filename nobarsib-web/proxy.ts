import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

/**
 * Header keamanan dan aturan cache untuk service worker.
 *
 * Dipasang di sini, bukan lewat `headers()` di next.config.ts, karena server
 * standalone (`output: "standalone"`, dipakai oleh Dockerfile) TIDAK menerapkan
 * header dari konfigurasi itu — sudah diverifikasi: `next start` mengirimnya,
 * `node server.js` dari keluaran standalone tidak, padahal keduanya dibangun
 * dari build yang sama dan header-nya ada di .next/routes-manifest.json.
 *
 * Proxy berjalan di dalam server standalone, jadi aturannya ikut ke produksi.
 *
 * Catatan: Next.js 16 mengganti nama "middleware" menjadi "proxy"; fungsinya
 * sama.
 */
export function proxy(request: NextRequest) {
  const res = NextResponse.next();

  res.headers.set("X-Content-Type-Options", "nosniff");
  res.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
  // Aplikasi ini hanya butuh lokasi. Kamera, mikrofon, dan pembayaran
  // dimatikan supaya tidak bisa diminta skrip pihak ketiga.
  res.headers.set(
    "Permissions-Policy",
    "camera=(), microphone=(), payment=(), geolocation=(self)",
  );

  // Service worker tidak boleh di-cache. Kalau versinya membeku di browser,
  // pembaruan aplikasi tidak pernah sampai ke pengguna — dan aturan cache di
  // dalamnya ikut membeku bersamanya.
  if (request.nextUrl.pathname === "/sw.js") {
    res.headers.set("Cache-Control", "no-cache, no-store, must-revalidate");
  }

  return res;
}

export const config = {
  // Aset ber-hash di /_next/static/ dilewati: isinya tidak pernah berubah dan
  // tidak perlu melewati proxy pada setiap permintaan.
  matcher: ["/((?!_next/static|_next/image).*)"],
};
