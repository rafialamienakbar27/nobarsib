import type { Metadata } from "next";

import { EmptyState, TidakAdaLaga } from "@/components/EmptyState";
import { MatchCard } from "@/components/MatchCard";
import { getSeasonMatches, getUpcomingMatches } from "@/lib/api";
import type { Match } from "@/lib/types";

export const metadata: Metadata = {
  title: "Jadwal",
  description: "Jadwal laga Persib musim ini beserta jumlah tempat nobar per laga.",
};

/** Sama seperti daftar venue: ruang berlebih dipakai menambah kolom. */
const GRID_LAGA = "grid grid-cols-1 gap-2 md:grid-cols-2 xl:grid-cols-3";

export default async function HalamanJadwal() {
  let semua: Match[];
  let mendatang: Match[];
  try {
    // Pemisahan mendatang/lampau diserahkan ke API, bukan dihitung dari jam
    // perangkat ini.
    //
    // Dua alasan. Pertama, membaca jam saat render melanggar aturan kemurnian
    // React — hasilnya bisa berubah setiap kali komponen dirender ulang.
    // Kedua, dan lebih penting: "sekarang" yang sah adalah waktu server
    // database, satu-satunya jam yang juga dipakai saat memfilter laga dan
    // menghitung konfirmasi H-1. Jam browser bisa salah setel, dan laga yang
    // sedang berlangsung pun tetap harus muncul sebagai mendatang (§8.2).
    [semua, mendatang] = await Promise.all([getSeasonMatches(), getUpcomingMatches(50)]);
  } catch {
    return (
      <EmptyState
        judul="Tidak bisa memuat jadwal"
        pesan="Server sedang tidak bisa dihubungi. Coba muat ulang sebentar lagi."
      />
    );
  }

  if (semua.length === 0) return <TidakAdaLaga />;

  const idMendatang = new Set(mendatang.map((m) => m.id));
  // Laga yang sudah lewat diurut mundur: yang baru saja berlalu paling berguna,
  // karena di situlah orang mencari tempat yang pernah dipakai nobar.
  const lampau = semua.filter((m) => !idMendatang.has(m.id)).reverse();

  return (
    <div className="space-y-6 sm:space-y-8">
      <h1 className="text-2xl font-extrabold sm:text-3xl">Jadwal</h1>

      <section className="space-y-2">
        {/* Biru hanya di "Mendatang", abu-abu di "Sudah berlalu": beda warna
            di sini menyampaikan mana yang masih bisa ditindaklanjuti. */}
        <h2 className="text-sm font-bold uppercase tracking-wider text-brand-accent">
          Mendatang
        </h2>
        {mendatang.length > 0 ? (
          // Satu laga per baris di HP, tiga per baris di layar lebar. Jadwal
          // semusim yang dipaksa satu kolom butuh puluhan kali gulir untuk
          // sesuatu yang sebenarnya muat dalam satu layar.
          <div className={GRID_LAGA}>
            {mendatang.map((m) => (
              <MatchCard key={m.id} match={m} sebagaiTautan />
            ))}
          </div>
        ) : (
          <p className="rounded-xl border border-dashed border-border px-4 py-6 text-center text-sm text-text-faint">
            Belum ada laga terjadwal.
          </p>
        )}
      </section>

      {lampau.length > 0 && (
        <section className="space-y-2">
          <h2 className="text-sm font-semibold uppercase tracking-wider text-text-faint">
            Sudah berlalu
          </h2>
          <div className={GRID_LAGA}>
            {lampau.slice(0, 10).map((m) => (
              <MatchCard key={m.id} match={m} sebagaiTautan />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
