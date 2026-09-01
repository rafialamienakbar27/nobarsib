/*
  Service worker NOBARSIB.
  Ditulis tangan, tanpa Workbox: yang dibutuhkan hanya tiga aturan, dan
  menambah 20 KB pustaka untuk itu justru melawan target LCP di §13.6.

  Tiga aturan itu:

    1. Aset statis Next (/_next/static/*)  -> cache dulu.
       Nama filenya sudah mengandung hash, jadi isinya tidak pernah berubah.

    2. Permintaan API (/v1/*)              -> jaringan saja, tanpa cache.
       Status "Dikonfirmasi" berubah sampai menit terakhir menjelang kickoff
       (§15.3). Menyajikan versi lama justru menghasilkan persis kesalahan yang
       paling ingin dihindari blueprint: data basi (§21).

    3. Navigasi halaman                    -> jaringan dulu, halaman luring
       sebagai cadangan kalau benar-benar tidak ada koneksi.
*/

// Dinaikkan ke v2 saat palet berganti ke biru #171B87: cache lama menyimpan
// CSS dan ikon warna sebelumnya, dan tanpa versi baru pengguna lama akan
// melihat tampilan campur — biru baru di halaman, ikon lama di layar utama.
const VERSI = "nobarsib-v2";
const CACHE_STATIS = `${VERSI}-statis`;
const CACHE_HALAMAN = `${VERSI}-halaman`;
const HALAMAN_LURING = "/luring";

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE_HALAMAN)
      .then((c) => c.addAll([HALAMAN_LURING]))
      // Gagal menyiapkan halaman luring tidak boleh menggagalkan pemasangan;
      // aplikasinya tetap berfungsi penuh selama ada koneksi.
      .catch(() => undefined)
      .then(() => self.skipWaiting()),
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((nama) =>
        Promise.all(
          nama.filter((n) => !n.startsWith(VERSI)).map((n) => caches.delete(n)),
        ),
      )
      .then(() => self.clients.claim()),
  );
});

self.addEventListener("fetch", (event) => {
  const { request } = event;

  // Hanya GET yang boleh disentuh. POST /track harus selalu lewat apa adanya.
  if (request.method !== "GET") return;

  const url = new URL(request.url);

  // Aturan 2 — jangan pernah menyimpan jawaban API.
  if (url.pathname.startsWith("/v1/")) return;

  // Lintas-domain (foto venue, dsb.) dibiarkan ditangani browser.
  if (url.origin !== self.location.origin) return;

  // Aturan 1 — aset ber-hash.
  if (url.pathname.startsWith("/_next/static/")) {
    event.respondWith(cacheDulu(request));
    return;
  }

  // Aturan 3 — navigasi halaman.
  if (request.mode === "navigate") {
    event.respondWith(jaringanDulu(request));
  }
});

async function cacheDulu(request) {
  const tersimpan = await caches.match(request);
  if (tersimpan) return tersimpan;

  const res = await fetch(request);
  if (res.ok) {
    const cache = await caches.open(CACHE_STATIS);
    cache.put(request, res.clone());
  }
  return res;
}

async function jaringanDulu(request) {
  try {
    const res = await fetch(request);
    if (res.ok) {
      const cache = await caches.open(CACHE_HALAMAN);
      cache.put(request, res.clone());
    }
    return res;
  } catch {
    return (
      (await caches.match(request)) ??
      (await caches.match(HALAMAN_LURING)) ??
      new Response("Sedang luring", {
        status: 503,
        headers: { "Content-Type": "text/plain; charset=utf-8" },
      })
    );
  }
}
